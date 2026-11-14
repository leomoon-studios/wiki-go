package handlers

import (
	"bytes"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
	"wiki-go/internal/config"
)

func TestOpenSearchDescription(t *testing.T) {
	cfg := &config.Config{}
	cfg.Server.PublicURL = "https://wiki.example:8443/"
	cfg.Wiki.Title = `Wiki & <Example> "test"`
	req := httptest.NewRequest(http.MethodGet, "/opensearch.xml", nil)
	req.Host = "evil.example"
	req.Header.Set("X-Forwarded-Host", "evil.example")
	req.Header.Set("X-Forwarded-Proto", "http")
	w := httptest.NewRecorder()
	OpenSearchHandler(w, req, cfg)
	if w.Code != http.StatusOK || !strings.HasPrefix(w.Header().Get("Content-Type"), "application/opensearchdescription+xml") {
		t.Fatalf("descriptor response = %d %v", w.Code, w.Header())
	}
	var desc openSearchDescription
	if err := xml.Unmarshal(w.Body.Bytes(), &desc); err != nil {
		t.Fatal(err)
	}
	if desc.XMLName.Space != "http://a9.com/-/spec/opensearch/1.1/" || desc.ShortName != "Wiki & <Example>" || desc.Description != "Search "+cfg.Wiki.Title || desc.InputEncoding != "UTF-8" {
		t.Fatalf("description = %+v", desc)
	}
	if desc.URL.Type != "text/html" || desc.URL.Template != "https://wiki.example:8443/?search={searchTerms}" || strings.Contains(w.Body.String(), "evil.example") {
		t.Fatalf("provider URL = %+v", desc.URL)
	}
	if !strings.Contains(w.Body.String(), "&lt;Example&gt;") {
		t.Fatalf("site title was not XML escaped: %s", w.Body.String())
	}
	providerURL := strings.Replace(desc.URL.Template, "{searchTerms}", url.QueryEscape(`café + "test"`), 1)
	parsed, err := url.Parse(providerURL)
	if err != nil || parsed.Query().Get("search") != `café + "test"` {
		t.Fatalf("provider request = %q, %v", providerURL, err)
	}
}

func TestOpenSearchRequiresValidPublicOrigin(t *testing.T) {
	for _, origin := range []string{"//example.com", "ftp://example.com", "https://user@example.com", "https://example.com/path", "https://example.com?x=1", "https://example.com/#fragment", "https://example.com:99999"} {
		cfg := &config.Config{}
		cfg.Server.PublicURL = origin
		w := httptest.NewRecorder()
		OpenSearchHandler(w, httptest.NewRequest(http.MethodGet, "/opensearch.xml", nil), cfg)
		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("origin %q: status = %d", origin, w.Code)
		}
	}
	w := httptest.NewRecorder()
	OpenSearchHandler(w, httptest.NewRequest(http.MethodPost, "/opensearch.xml", nil), &config.Config{})
	if w.Code != http.StatusMethodNotAllowed || w.Header().Get("Allow") != "GET" {
		t.Fatalf("POST = %d, Allow = %q", w.Code, w.Header().Get("Allow"))
	}
}

func TestOpenSearchAutomaticOrigin(t *testing.T) {
	for _, tc := range []struct {
		name, remote, host, forwardedHost, proto, want string
		trusted                                        []string
	}{
		{"direct", "203.0.113.1:1234", "wiki.example:8080", "", "", "http://wiki.example:8080", nil},
		{"untrusted forwarding", "203.0.113.1:1234", "wiki.example", "evil.example", "https", "http://wiki.example", nil},
		{"trusted proxy", "127.0.0.1:1234", "localhost:8080", "wiki.example:8443", "https", "https://wiki.example:8443", []string{"127.0.0.1"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{}
			cfg.Server.TrustedProxies = tc.trusted
			r := httptest.NewRequest(http.MethodGet, "/opensearch.xml", nil)
			r.RemoteAddr, r.Host = tc.remote, tc.host
			r.Header.Set("X-Forwarded-Host", tc.forwardedHost)
			r.Header.Set("X-Forwarded-Proto", tc.proto)
			w := httptest.NewRecorder()
			OpenSearchHandler(w, r, cfg)
			var desc openSearchDescription
			if w.Code != http.StatusOK || xml.Unmarshal(w.Body.Bytes(), &desc) != nil || desc.URL.Template != tc.want+"/?search={searchTerms}" {
				t.Fatalf("descriptor = %d %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestOpenSearchPublicURLSurvivesConfigSave(t *testing.T) {
	cfg := &config.Config{}
	cfg.Server.PublicURL = "https://wiki.example:8443"
	var saved bytes.Buffer
	if err := config.SaveConfig(cfg, &saved); err != nil {
		t.Fatal(err)
	}
	var loaded config.Config
	if err := yaml.Unmarshal(saved.Bytes(), &loaded); err != nil {
		t.Fatal(err)
	}
	if loaded.Server.PublicURL != cfg.Server.PublicURL {
		t.Fatalf("public URL = %q", loaded.Server.PublicURL)
	}
}
