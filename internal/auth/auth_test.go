package auth

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"wiki-go/internal/config"
)

func TestRevokeUserSessionsRemovesAllMatchingSessionsAndPersists(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "sessions.json")
	if err := InitSessionStore(storePath); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	cfg.Server.AllowInsecureCookies = true
	aliceOne := createTestSession(t, cfg, "alice")
	aliceTwo := createTestSession(t, cfg, "alice")
	bob := createTestSession(t, cfg, "bob")

	revoked, err := RevokeUserSessions("alice")
	if err != nil {
		t.Fatal(err)
	}
	if revoked != 2 {
		t.Fatalf("revoked = %d, want 2", revoked)
	}
	if sessionForCookie(aliceOne) != nil || sessionForCookie(aliceTwo) != nil {
		t.Fatal("an alice session remained active")
	}
	if session := sessionForCookie(bob); session == nil || session.Username != "bob" {
		t.Fatal("unrelated bob session was revoked")
	}

	persisted, err := NewSessionStore(storePath).LoadSessions()
	if err != nil {
		t.Fatal(err)
	}
	if len(persisted) != 1 {
		t.Fatalf("persisted session count = %d, want 1", len(persisted))
	}
	for _, session := range persisted {
		if session.Username != "bob" {
			t.Fatalf("persisted unexpected session for %q", session.Username)
		}
	}
}

func createTestSession(t *testing.T, cfg *config.Config, username string) *http.Cookie {
	t.Helper()
	recorder := httptest.NewRecorder()
	if err := CreateSession(recorder, username, config.RoleEditor, nil, false, cfg); err != nil {
		t.Fatal(err)
	}
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == "session_token" {
			return cookie
		}
	}
	t.Fatal("session token cookie was not created")
	return nil
}

func sessionForCookie(cookie *http.Cookie) *Session {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(cookie)
	return GetSession(request)
}

func TestCheckAccessCanonicalizesBeforeRuleMatching(t *testing.T) {
	cfg := &config.Config{}
	cfg.AccessRules = []config.AccessRule{{
		Pattern: "/finance/**",
		Access:  "restricted",
		Groups:  []string{"finance"},
	}}

	tests := []struct {
		name    string
		path    string
		private bool
		want    bool
	}{
		{name: "normal public path", path: "/public", want: true},
		{name: "normal private path", path: "/internal", private: true, want: false},
		{name: "restricted path", path: "/finance/report", want: false},
		{name: "double encoded traversal", path: "/x/%252e%252e/finance", want: false},
		{name: "in-wiki traversal", path: "/x/../finance", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg.Wiki.Private = test.private
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			_, got := CheckAccess(request, cfg)
			if got != test.want {
				t.Fatalf("CheckAccess(%q) = %t, want %t", test.path, got, test.want)
			}
		})
	}
}
