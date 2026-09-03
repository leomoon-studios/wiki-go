package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
	if response := login("198.51.100.2"); response.Code != http.StatusTooManyRequests {
		t.Fatalf("second login status = %d, want 429; body: %s", response.Code, response.Body.String())
	}
}
