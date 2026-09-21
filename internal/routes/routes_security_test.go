package routes

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

func TestCSPMiddlewareEnforcesSameOriginScripts(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	recorder := httptest.NewRecorder()

	CSPMiddleware(next).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/docs/security", nil))

	policy := recorder.Header().Get("Content-Security-Policy")
	for _, directive := range []string{
		"default-src 'self'",
		"script-src 'self'",
		"script-src-attr 'none'",
		"object-src 'none'",
	} {
		if !strings.Contains(policy, directive) {
			t.Errorf("enforcing CSP is missing %q: %q", directive, policy)
		}
	}
	imageDirective := cspDirective(policy, "img-src")
	for _, source := range []string{"'self'", "data:", "http:", "https:"} {
		if !strings.Contains(imageDirective, source) {
			t.Errorf("image policy blocks supported image source %s: %q", source, imageDirective)
		}
	}
	for _, forbidden := range []string{"'unsafe-inline'", "'unsafe-eval'"} {
		scriptDirective := cspDirective(policy, "script-src")
		if strings.Contains(scriptDirective, forbidden) {
			t.Errorf("script policy permits %s: %q", forbidden, scriptDirective)
		}
	}
	if reportOnly := recorder.Header().Get("Content-Security-Policy-Report-Only"); reportOnly != "" {
		t.Errorf("report-only CSP is still present: %q", reportOnly)
	}
}

func TestPageRouteUsesOneCanonicalPathForAccessAndFilesystemLookup(t *testing.T) {
	testConfig := newRouteSecurityTestConfig(t)
	financeDocument := filepath.Join(testConfig.Wiki.RootDir, "documents", "finance", "document.md")
	if err := os.MkdirAll(filepath.Dir(financeDocument), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(financeDocument, []byte("# Route finance canary"), 0o644); err != nil {
		t.Fatal(err)
	}
	financeCookie := routeSessionCookie(t, testConfig, "finance-editor", config.RoleEditor, []string{"finance"})

	request := httptest.NewRequest(http.MethodGet, "/%66inance", nil)
	request.AddCookie(financeCookie)
	response := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "Route finance canary") {
		t.Fatal("route authorization and filesystem lookup did not resolve the same finance path")
	}
}

func TestPageRouteRejectsResidualEncodedTraversalBeforeFilesystemLookup(t *testing.T) {
	testConfig := newRouteSecurityTestConfig(t)
	canary := "ROUTE-EXTERNAL-CANARY"
	externalDocument := filepath.Join(testConfig.Wiki.RootDir, "outside", "document.md")
	if err := os.MkdirAll(filepath.Dir(externalDocument), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(externalDocument, []byte(canary), 0o644); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodGet, "/%252e%252e/outside", nil)
	response := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(response, request)

	if response.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303; body: %s", response.Code, response.Body.String())
	}
	if strings.Contains(response.Body.String(), canary) {
		t.Fatal("route disclosed the external document canary")
	}
}

func newRouteSecurityTestConfig(t *testing.T) *config.Config {
	t.Helper()
	root := t.TempDir()
	testConfig := &config.Config{}
	testConfig.Server.AllowInsecureCookies = true
	testConfig.Wiki.RootDir = root
	testConfig.Wiki.DocumentsDir = "documents"
	testConfig.Wiki.Title = "Route Security Test Wiki"
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
	if err := os.MkdirAll(filepath.Join(root, "documents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := auth.InitSessionStore(filepath.Join(root, "sessions.json")); err != nil {
		t.Fatal(err)
	}

	previousMux := http.DefaultServeMux
	http.DefaultServeMux = http.NewServeMux()
	t.Cleanup(func() { http.DefaultServeMux = previousMux })
	SetupRoutes(testConfig)
	return testConfig
}

func routeSessionCookie(t *testing.T, testConfig *config.Config, username, role string, groups []string) *http.Cookie {
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

func cspDirective(policy, name string) string {
	for _, directive := range strings.Split(policy, ";") {
		directive = strings.TrimSpace(directive)
		if directive == name || strings.HasPrefix(directive, name+" ") {
			return directive
		}
	}
	return ""
}
