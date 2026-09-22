package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"wiki-go/internal/ban"
	"wiki-go/internal/config"
)

func TestClientIPTrustBoundary(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		trustedProxies []string
		forwardedFor   string
		realIP         string
		want           string
	}{
		{
			name:         "untrusted peer cannot spoof forwarding headers",
			remoteAddr:   "203.0.113.10:1234",
			forwardedFor: "198.51.100.20",
			realIP:       "192.0.2.30",
			want:         "203.0.113.10",
		},
		{
			name:           "trusted exact proxy may supply real IP",
			remoteAddr:     "127.0.0.1:4321",
			trustedProxies: []string{"127.0.0.1"},
			realIP:         "198.51.100.20",
			want:           "198.51.100.20",
		},
		{
			name:           "trusted proxy chain is removed right to left",
			remoteAddr:     "10.0.0.2:4321",
			trustedProxies: []string{"10.0.0.0/8", "192.0.2.0/24"},
			forwardedFor:   "198.51.100.20, 192.0.2.10",
			want:           "198.51.100.20",
		},
		{
			name:           "leftmost spoof does not replace nearest untrusted address",
			remoteAddr:     "10.0.0.2:4321",
			trustedProxies: []string{"10.0.0.0/8"},
			forwardedFor:   "203.0.113.66, 198.51.100.20",
			want:           "198.51.100.20",
		},
		{
			name:           "malformed forwarded chain fails closed",
			remoteAddr:     "10.0.0.2:4321",
			trustedProxies: []string{"10.0.0.0/8"},
			forwardedFor:   "198.51.100.20, not-an-ip",
			want:           "10.0.0.2",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/login", nil)
			request.RemoteAddr = test.remoteAddr
			request.Header.Set("X-Forwarded-For", test.forwardedFor)
			request.Header.Set("X-Real-IP", test.realIP)
			testConfig := &config.Config{}
			testConfig.Server.TrustedProxies = test.trustedProxies
			if got := clientIP(request, testConfig); got != test.want {
				t.Fatalf("clientIP() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestLoginThrottlingIgnoresSpoofedForwardingHeaders(t *testing.T) {
	previousConfig := cfg
	t.Cleanup(func() {
		cfg = previousConfig
		loginBan = nil
		ban.UpdatePolicy(5, 180, 60, 86400)
	})

	testConfig := &config.Config{}
	testConfig.Wiki.RootDir = t.TempDir()
	testConfig.Security.LoginBan.Enabled = true
	testConfig.Security.LoginBan.MaxFailures = 2
	testConfig.Security.LoginBan.WindowSeconds = 180
	testConfig.Security.LoginBan.InitialBanSeconds = 60
	testConfig.Security.LoginBan.MaxBanSeconds = 60
	cfg = testConfig
	InitLoginBan(testConfig)

	login := func(spoofedIP string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"username":"nobody","password":"wrong"}`))
		request.RemoteAddr = "203.0.113.10:1234"
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-Forwarded-For", spoofedIP)
		response := httptest.NewRecorder()
		LoginHandler(response, request)
		return response
	}

	if response := login("198.51.100.1"); response.Code != http.StatusUnauthorized {
		t.Fatalf("first login status = %d, want 401; body: %s", response.Code, response.Body.String())
	}
	waitForLoginBanPersistence(t, testConfig, true)
	banStatePath := filepath.Join(testConfig.Wiki.RootDir, "temp", "login_ban.json")
	if err := os.Remove(banStatePath); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if response := login("198.51.100.2"); response.Code != http.StatusTooManyRequests {
		t.Fatalf("second login status = %d, want 429; body: %s", response.Code, response.Body.String())
	}
	waitForLoginBanPersistence(t, testConfig, true)
}

func TestLoginHandlerNeutralizesUsernameLogInjection(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantUser string
	}{
		{
			name:     "advisory multiple newline payload",
			username: "pwn\n[FAKE LOG LINE INJECTED BY ATTACKER]\nadmin",
			wantUser: `pwn\n[FAKE LOG LINE INJECTED BY ATTACKER]\nadmin`,
		},
		{name: "carriage return", username: "pwn\rforged", wantUser: `pwn\rforged`},
		{name: "CRLF", username: "pwn\r\nforged", wantUser: `pwn\r\nforged`},
		{name: "multiple newlines", username: "pwn\n\n\nforged", wantUser: `pwn\n\n\nforged`},
	}
	for _, rateLimitingEnabled := range []bool{false, true} {
		rateName := "rate limiting disabled"
		if rateLimitingEnabled {
			rateName = "rate limiting enabled"
		}
		t.Run(rateName, func(t *testing.T) {
			for _, test := range tests {
				t.Run(test.name, func(t *testing.T) {
					testConfig := installUserSessionTestConfig(t, nil)
					configureLoginBanForLogTest(t, testConfig, rateLimitingEnabled)
					output := captureHandlerLogs(t)

					response := performLoginRequest(t, test.username, "wrong", "203.0.113.10:1234", "")
					if response.Code != http.StatusUnauthorized {
						t.Fatalf("status = %d, want 401; body: %s", response.Code, response.Body.String())
					}
					assertLoginResponse(t, response, false, "Invalid credentials")

					want := "[WARN] Login failed for user " + test.wantUser + " from 203.0.113.10\n"
					if output.String() != want {
						t.Fatalf("log output = %q, want %q", output.String(), want)
					}
					if strings.Count(output.String(), "\n") != 1 {
						t.Fatalf("physical log record count = %d, want 1", strings.Count(output.String(), "\n"))
					}
					waitForLoginBanPersistence(t, testConfig, rateLimitingEnabled)
				})
			}
		})
	}
}

func TestLoginHandlerKeepsOrdinaryLogsReadable(t *testing.T) {
	for _, rateLimitingEnabled := range []bool{false, true} {
		rateName := "rate limiting disabled"
		if rateLimitingEnabled {
			rateName = "rate limiting enabled"
		}
		t.Run(rateName, func(t *testing.T) {
			testConfig := installUserSessionTestConfig(t, []config.User{
				userWithPassword(t, "alice", "correct-password", config.RoleEditor),
			})
			testConfig.Server.TrustedProxies = []string{"127.0.0.1"}
			configureLoginBanForLogTest(t, testConfig, rateLimitingEnabled)
			output := captureHandlerLogs(t)

			failed := performLoginRequest(t, "plainuser", "wrong", "127.0.0.1:4321", "198.51.100.20")
			if failed.Code != http.StatusUnauthorized {
				t.Fatalf("failed login status = %d, want 401; body: %s", failed.Code, failed.Body.String())
			}
			assertLoginResponse(t, failed, false, "Invalid credentials")
			const wantFailed = "[WARN] Login failed for user plainuser from 198.51.100.20\n"
			if output.String() != wantFailed {
				t.Fatalf("failed login output = %q, want %q", output.String(), wantFailed)
			}
			waitForLoginBanPersistence(t, testConfig, rateLimitingEnabled)

			output.Reset()
			if rateLimitingEnabled {
				if err := os.Remove(filepath.Join(testConfig.Wiki.RootDir, "temp", "login_ban.json")); err != nil && !os.IsNotExist(err) {
					t.Fatal(err)
				}
			}
			successful := performLoginRequest(t, "alice", "correct-password", "127.0.0.1:4321", "198.51.100.20")
			if successful.Code != http.StatusOK {
				t.Fatalf("successful login status = %d, want 200; body: %s", successful.Code, successful.Body.String())
			}
			assertLoginResponse(t, successful, true, "Login successful")
			const wantSuccess = "[INFO] User alice logged in from 198.51.100.20\n"
			if output.String() != wantSuccess {
				t.Fatalf("successful login output = %q, want %q", output.String(), wantSuccess)
			}
			waitForLoginBanPersistence(t, testConfig, rateLimitingEnabled)
		})
	}
}

func configureLoginBanForLogTest(t *testing.T, testConfig *config.Config, enabled bool) {
	t.Helper()
	testConfig.Security.LoginBan.Enabled = enabled
	if !enabled {
		loginBan = nil
		return
	}
	testConfig.Security.LoginBan.MaxFailures = 100
	testConfig.Security.LoginBan.WindowSeconds = 180
	testConfig.Security.LoginBan.InitialBanSeconds = 60
	testConfig.Security.LoginBan.MaxBanSeconds = 60
	InitLoginBan(testConfig)
	t.Cleanup(func() { ban.UpdatePolicy(5, 180, 60, 86400) })
}

func captureHandlerLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var output bytes.Buffer
	originalWriter := log.Writer()
	originalFlags := log.Flags()
	originalPrefix := log.Prefix()
	log.SetOutput(&output)
	log.SetFlags(0)
	log.SetPrefix("")
	t.Cleanup(func() {
		log.SetOutput(originalWriter)
		log.SetFlags(originalFlags)
		log.SetPrefix(originalPrefix)
	})
	return &output
}

func performLoginRequest(t *testing.T, username, password, remoteAddr, forwardedFor string) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(LoginRequest{Username: username, Password: password})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(body))
	request.RemoteAddr = remoteAddr
	request.Header.Set("Content-Type", "application/json")
	if forwardedFor != "" {
		request.Header.Set("X-Forwarded-For", forwardedFor)
	}
	response := httptest.NewRecorder()
	LoginHandler(response, request)
	return response
}

func assertLoginResponse(t *testing.T, response *httptest.ResponseRecorder, wantSuccess bool, wantMessage string) {
	t.Helper()
	var result struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Success != wantSuccess || result.Message != wantMessage {
		t.Fatalf("response = %+v, want success=%v message=%q", result, wantSuccess, wantMessage)
	}
}

func waitForLoginBanPersistence(t *testing.T, testConfig *config.Config, enabled bool) {
	t.Helper()
	if !enabled {
		return
	}
	path := filepath.Join(testConfig.Wiki.RootDir, "temp", "login_ban.json")
	deadline := time.Now().Add(time.Second)
	for {
		if _, err := os.Stat(path); err == nil {
			return
		} else if !os.IsNotExist(err) {
			t.Fatal(err)
		}
		if time.Now().After(deadline) {
			t.Fatalf("login ban state was not persisted at %q", path)
		}
		time.Sleep(time.Millisecond)
	}
}
