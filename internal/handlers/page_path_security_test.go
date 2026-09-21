package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"wiki-go/internal/auth"
	"wiki-go/internal/config"
)

func TestPageHandlerDoesNotDiscloseDoubleEncodedTraversalTarget(t *testing.T) {
	testConfig := installUserSessionTestConfig(t, nil)
	testConfig.Wiki.DocumentsDir = "documents"
	canary := "EXTERNAL-DOCUMENT-CANARY"
	externalDocument := filepath.Join(testConfig.Wiki.RootDir, "outside", "document.md")
	if err := os.MkdirAll(filepath.Dir(externalDocument), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(externalDocument, []byte(canary), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		cookie *http.Cookie
	}{
		{name: "anonymous"},
		{name: "viewer", cookie: sessionCookieForUser(t, testConfig, "viewer", config.RoleViewer)},
		{name: "editor", cookie: sessionCookieForUser(t, testConfig, "editor", config.RoleEditor)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/%252e%252e/outside", nil)
			if test.cookie != nil {
				request.AddCookie(test.cookie)
			}
			response := httptest.NewRecorder()

			PageHandler(response, request, testConfig)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; body: %s", response.Code, response.Body.String())
			}
			if strings.Contains(response.Body.String(), canary) {
				t.Fatal("response disclosed the external document canary")
			}
		})
	}
}

func TestRestrictedDocumentAccessAllowsFinanceGroupAndAdministrators(t *testing.T) {
	testConfig := installUserSessionTestConfig(t, nil)
	testConfig.Wiki.DocumentsDir = "documents"
	testConfig.Wiki.Title = "Security Test Wiki"
	testConfig.Wiki.Owner = "wiki.example"
	testConfig.Wiki.Timezone = "UTC"
	testConfig.Wiki.Language = "en"
	testConfig.Wiki.DisableComments = true
	testConfig.Wiki.HideAttachments = true
	testConfig.AccessRules = []config.AccessRule{{
		Pattern: "/finance/**",
		Access:  "restricted",
		Groups:  []string{"finance"},
	}}
	financeDocument := filepath.Join(testConfig.Wiki.RootDir, "documents", "finance", "document.md")
	if err := os.MkdirAll(filepath.Dir(financeDocument), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(financeDocument, []byte("# Authorized finance content"), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		cookie *http.Cookie
	}{
		{
			name:   "finance group editor",
			cookie: sessionCookieForUserWithGroups(t, testConfig, "finance-editor", config.RoleEditor, []string{"finance"}),
		},
		{
			name:   "administrator",
			cookie: sessionCookieForUser(t, testConfig, "admin", config.RoleAdmin),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pageRequest := httptest.NewRequest(http.MethodGet, "/finance", nil)
			pageRequest.AddCookie(test.cookie)
			pageResponse := httptest.NewRecorder()
			PageHandler(pageResponse, pageRequest, testConfig)
			if pageResponse.Code != http.StatusOK {
				t.Fatalf("page status = %d, want 200; body: %s", pageResponse.Code, pageResponse.Body.String())
			}
			if !strings.Contains(pageResponse.Body.String(), "Authorized finance content") {
				t.Fatal("page response omitted authorized finance content")
			}

			sourceRequest := httptest.NewRequest(http.MethodGet, "/api/source/finance", nil)
			sourceRequest.AddCookie(test.cookie)
			sourceResponse := httptest.NewRecorder()
			SourceHandler(sourceResponse, sourceRequest)
			if sourceResponse.Code != http.StatusOK {
				t.Fatalf("source status = %d, want 200; body: %s", sourceResponse.Code, sourceResponse.Body.String())
			}
			if sourceResponse.Body.String() != "# Authorized finance content" {
				t.Fatalf("source body = %q, want authorized finance content", sourceResponse.Body.String())
			}
		})
	}
}

func sessionCookieForUserWithGroups(t *testing.T, testConfig *config.Config, username, role string, groups []string) *http.Cookie {
	t.Helper()
	response := httptest.NewRecorder()
	if err := auth.CreateSession(response, username, role, groups, false, testConfig); err != nil {
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
