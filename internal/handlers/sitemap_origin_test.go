package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"wiki-go/internal/config"
)

func TestSitemapUsesConfiguredPublicOrigin(t *testing.T) {
	cfg := installUserSessionTestConfig(t, nil)
	cfg.Wiki.DocumentsDir = "documents"
	if err := os.MkdirAll(filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg.Server.PublicURL = "https://wiki.example:8443/"
	r := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	r.Host = "attacker.example"
	r.Header.Set("X-Forwarded-Proto", "http")
	r.Header.Set("X-Forwarded-Host", "attacker.example")
	w := httptest.NewRecorder()
	SitemapHandler(w, r, cfg)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "<loc>https://wiki.example:8443</loc>") || strings.Contains(w.Body.String(), "attacker.example") {
		t.Fatalf("sitemap = %d %s", w.Code, w.Body.String())
	}
}

func TestSitemapOriginFallbackAndInvalidConfig(t *testing.T) {
	cfg := &config.Config{}
	r := httptest.NewRequest(http.MethodGet, "http://wiki.example/sitemap.xml", nil)
	r.Header.Set("X-Forwarded-Proto", "https")
	r.RemoteAddr = "127.0.0.1:1234"
	cfg.Server.TrustedProxies = []string{"127.0.0.1"}
	got, err := getBaseURL(r, cfg)
	if err != nil || got != "https://wiki.example" {
		t.Fatalf("fallback = %q, %v", got, err)
	}
	cfg.Server.PublicURL = "https://wiki.example/invalid-path"
	w := httptest.NewRecorder()
	SitemapHandler(w, r, cfg)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("invalid configured origin status = %d", w.Code)
	}
}
