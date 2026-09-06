package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
	for _, faviconOrigin := range []string{"https://www.google.com", "https://*.gstatic.com"} {
		if !strings.Contains(imageDirective, faviconOrigin) {
			t.Errorf("image policy blocks link favicons from %s: %q", faviconOrigin, imageDirective)
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

func cspDirective(policy, name string) string {
	for _, directive := range strings.Split(policy, ";") {
		directive = strings.TrimSpace(directive)
		if directive == name || strings.HasPrefix(directive, name+" ") {
			return directive
		}
	}
	return ""
}
