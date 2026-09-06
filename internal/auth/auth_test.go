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
