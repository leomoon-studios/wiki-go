package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"wiki-go/internal/auth"
	"wiki-go/internal/config"
	"wiki-go/internal/crypto"
)

func TestPasswordHandlerRevokesAllSessionsAndReturnsLogoutInstructions(t *testing.T) {
	testConfig := installUserSessionTestConfig(t, []config.User{
		userWithPassword(t, "alice", "old-password", config.RoleEditor),
	})
	aliceOne := sessionCookieForUser(t, testConfig, "alice", config.RoleEditor)
	aliceTwo := sessionCookieForUser(t, testConfig, "alice", config.RoleEditor)

	request := httptest.NewRequest(http.MethodPost, "/api/password", strings.NewReader(`{
		"current_password":"old-password",
		"new_password":"new-password"
	}`))
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(aliceOne)
	response := httptest.NewRecorder()
	PasswordHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", response.Code, response.Body.String())
	}
	var result struct {
		Success   bool   `json:"success"`
		LoggedOut bool   `json:"logged_out"`
		Redirect  string `json:"redirect"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if !result.Success || !result.LoggedOut || result.Redirect != "/login" {
		t.Fatalf("unexpected response: %+v", result)
	}
	assertSessionCookieExpired(t, response, "session_token")
	assertSessionCookieExpired(t, response, "session_user")
	assertSessionInvalid(t, aliceOne)
	assertSessionInvalid(t, aliceTwo)
	if !crypto.CheckPasswordHash("new-password", testConfig.Users[0].Password) {
		t.Fatal("new password was not saved")
	}

	loginRequest := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{
		"username":"alice",
		"password":"new-password",
		"keepLoggedIn":false
	}`))
	loginRequest.RemoteAddr = "203.0.113.10:1234"
	loginResponse := httptest.NewRecorder()
	LoginHandler(loginResponse, loginRequest)
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("login after password change status = %d, want 200; body: %s", loginResponse.Code, loginResponse.Body.String())
	}
	var replacementSession *http.Cookie
	for _, cookie := range loginResponse.Result().Cookies() {
		if cookie.Name == "session_token" {
			replacementSession = cookie
			break
		}
	}
	if replacementSession == nil || sessionFromCookie(replacementSession) == nil {
		t.Fatal("login after password change did not create a replacement session")
	}
}

func TestAdministratorAccountChangesRevokeTargetSessions(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{
			name:   "role demotion",
			method: http.MethodPut,
			path:   "/api/users",
			body:   `{"username":"target","role":"viewer"}`,
		},
		{
			name:   "administrator password reset",
			method: http.MethodPut,
			path:   "/api/users",
			body:   `{"username":"target","new_password":"replacement","role":"editor"}`,
		},
		{
			name:   "group access change",
			method: http.MethodPut,
			path:   "/api/users",
			body:   `{"username":"target","role":"editor","groups":["new-group"]}`,
		},
		{
			name:   "user deletion",
			method: http.MethodDelete,
			path:   "/api/users?username=target",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testConfig := installUserSessionTestConfig(t, []config.User{
				userWithPassword(t, "admin", "admin-password", config.RoleAdmin),
				userWithPassword(t, "target", "target-password", config.RoleEditor),
			})
			adminCookie := sessionCookieForUser(t, testConfig, "admin", config.RoleAdmin)
			targetCookie := sessionCookieForUser(t, testConfig, "target", config.RoleEditor)

			request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(adminCookie)
			response := httptest.NewRecorder()
			switch test.method {
			case http.MethodPut:
				UpdateUserHandler(response, request)
			case http.MethodDelete:
				DeleteUserHandler(response, request)
			}

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body: %s", response.Code, response.Body.String())
			}
			assertSessionInvalid(t, targetCookie)
			if session := sessionFromCookie(adminCookie); session == nil || session.Username != "admin" {
				t.Fatal("administrator session was unexpectedly revoked")
			}
		})
	}
}

func installUserSessionTestConfig(t *testing.T, users []config.User) *config.Config {
	t.Helper()
	previousConfig := cfg
	previousPath := config.ConfigFilePath
	previousLoginBan := loginBan
	t.Cleanup(func() {
		cfg = previousConfig
		config.ConfigFilePath = previousPath
		loginBan = previousLoginBan
	})

	root := t.TempDir()
	testConfig := &config.Config{}
	testConfig.Server.AllowInsecureCookies = true
	testConfig.Wiki.RootDir = root
	testConfig.Security.PasswordStrength = 4
	testConfig.Users = users
	cfg = testConfig
	loginBan = nil
	config.ConfigFilePath = filepath.Join(root, "config.yaml")
	if err := auth.InitSessionStore(filepath.Join(root, "sessions.json")); err != nil {
		t.Fatal(err)
	}
	return testConfig
}

func userWithPassword(t *testing.T, username, password, role string) config.User {
	t.Helper()
	hash, err := crypto.HashPassword(password, 4)
	if err != nil {
		t.Fatal(err)
	}
	return config.User{Username: username, Password: hash, Role: role}
}

func sessionCookieForUser(t *testing.T, testConfig *config.Config, username, role string) *http.Cookie {
	t.Helper()
	response := httptest.NewRecorder()
	if err := auth.CreateSession(response, username, role, nil, false, testConfig); err != nil {
		t.Fatal(err)
	}
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == "session_token" {
			return cookie
		}
	}
	t.Fatal("session token cookie was not created")
	return nil
}

func sessionFromCookie(cookie *http.Cookie) *auth.Session {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(cookie)
	return auth.GetSession(request)
}

func assertSessionInvalid(t *testing.T, cookie *http.Cookie) {
	t.Helper()
	if session := sessionFromCookie(cookie); session != nil {
		t.Fatalf("session for %q remained active", session.Username)
	}
}

func assertSessionCookieExpired(t *testing.T, response *httptest.ResponseRecorder, name string) {
	t.Helper()
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == name {
			if cookie.MaxAge >= 0 {
				t.Fatalf("%s MaxAge = %d, want a negative value", name, cookie.MaxAge)
			}
			return
		}
	}
	t.Fatalf("response did not expire %s", name)
}
