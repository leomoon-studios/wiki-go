package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"wiki-go/internal/auth"
	"wiki-go/internal/config"
)

const safeSVGFixture = `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64" width="64" height="64" role="img" aria-label="Safe icon">
  <title>Safe &amp; static</title>
  <defs><linearGradient id="paint"><stop offset="0" stop-color="#f00"/><stop offset="1" stop-color="#00f"/></linearGradient></defs>
  <g fill="url(#paint)" stroke="#111"><circle cx="32" cy="32" r="30"/><path d="M16 32 L28 44 L50 18" fill="none"/></g>
</svg>`

func TestSanitizeSVGAllowsMinimalStaticSVG(t *testing.T) {
	got, err := sanitizeSVG([]byte(safeSVGFixture))
	if err != nil {
		t.Fatal(err)
	}
	output := string(got)
	for _, want := range []string{
		`<svg xmlns="http://www.w3.org/2000/svg"`,
		`<linearGradient id="paint">`,
		`fill="url(#paint)"`,
		`<circle cx="32" cy="32" r="30"></circle>`,
		`Safe &amp; static`,
	} {
		if !strings.Contains(output, want) {
			t.Errorf("sanitized safe SVG omitted %q: %s", want, output)
		}
	}
}

func TestSanitizeSVGRejectsActiveAndAmbiguousMarkup(t *testing.T) {
	tests := []struct {
		name    string
		payload string
	}{
		{"event attribute", `<svg xmlns="http://www.w3.org/2000/svg"><circle onmouseover="alert(1)"/></svg>`},
		{"mixed case event", `<svg xmlns="http://www.w3.org/2000/svg"><circle oNlOaD="alert(1)"/></svg>`},
		{"script", `<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`},
		{"namespaced script", `<s:svg xmlns:s="http://www.w3.org/2000/svg"><s:script>alert(1)</s:script></s:svg>`},
		{"animation", `<svg xmlns="http://www.w3.org/2000/svg"><animate attributeName="href" values="x;javascript:alert(1)"/></svg>`},
		{"CSS element", `<svg xmlns="http://www.w3.org/2000/svg"><style>@import url(https://evil.example/x.css)</style></svg>`},
		{"CSS attribute", `<svg xmlns="http://www.w3.org/2000/svg"><rect style="fill:url(https://evil.example/x.svg)"/></svg>`},
		{"external paint", `<svg xmlns="http://www.w3.org/2000/svg"><rect fill="url(https://evil.example/x.svg#paint)"/></svg>`},
		{"data paint", `<svg xmlns="http://www.w3.org/2000/svg"><rect fill="url(data:image/svg+xml;base64,AAAA)"/></svg>`},
		{"escaped CSS URL", `<svg xmlns="http://www.w3.org/2000/svg"><rect fill="\75rl(https://evil.example/x.svg)"/></svg>`},
		{"external href", `<svg xmlns="http://www.w3.org/2000/svg"><text href="https://evil.example/">click</text></svg>`},
		{"xlink href", `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"><text xlink:href="javascript:alert(1)">click</text></svg>`},
		{"embedded HTML", `<svg xmlns="http://www.w3.org/2000/svg"><foreignObject><div xmlns="http://www.w3.org/1999/xhtml">HTML</div></foreignObject></svg>`},
		{"embedded image", `<svg xmlns="http://www.w3.org/2000/svg"><image href="data:image/svg+xml;base64,AAAA"/></svg>`},
		{"wrong root namespace", `<svg xmlns="http://www.w3.org/1999/xhtml"><circle/></svg>`},
		{"extra namespace", `<svg xmlns="http://www.w3.org/2000/svg" xmlns:evil="https://evil.example/ns"><circle/></svg>`},
		{"doctype entity", `<!DOCTYPE svg [<!ENTITY xxe SYSTEM "file:///etc/passwd">]><svg xmlns="http://www.w3.org/2000/svg"><text>&xxe;</text></svg>`},
		{"stylesheet instruction", `<?xml-stylesheet href="https://evil.example/x.css"?><svg xmlns="http://www.w3.org/2000/svg"></svg>`},
		{"malformed XML", `<svg xmlns="http://www.w3.org/2000/svg"><circle></svg>`},
		{"multiple roots", `<svg xmlns="http://www.w3.org/2000/svg"></svg><svg xmlns="http://www.w3.org/2000/svg"></svg>`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if output, err := sanitizeSVG([]byte(test.payload)); err == nil {
				t.Fatalf("active SVG was accepted: %s", output)
			}
		})
	}
}

func TestSVGUploadIsSanitizedWhenMIMECheckingIsDisabled(t *testing.T) {
	cfg, documentDir := svgTestConfig(t)
	cfg.Wiki.DisableFileUploadChecking = true
	cookie := editorSessionCookie(t, cfg)

	tests := []struct {
		name       string
		filename   string
		content    string
		wantStatus int
		wantSaved  bool
	}{
		{"safe", "safe.svg", safeSVGFixture, http.StatusOK, true},
		{"event", "event.svg", `<svg xmlns="http://www.w3.org/2000/svg"><circle onmouseover="alert(1)"/></svg>`, http.StatusBadRequest, false},
		{"malformed", "malformed.svg", `<svg><circle></svg>`, http.StatusBadRequest, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := newSVGUploadRequest(t, test.filename, test.content)
			request.AddCookie(cookie)
			response := httptest.NewRecorder()
			UploadFileHandler(response, request, cfg)
			if response.Code != test.wantStatus {
				t.Fatalf("upload status = %d, want %d; body: %s", response.Code, test.wantStatus, response.Body.String())
			}

			savedPath := filepath.Join(documentDir, test.filename)
			_, err := os.Stat(savedPath)
			if test.wantSaved && err != nil {
				t.Fatalf("safe SVG was not saved: %v", err)
			}
			if !test.wantSaved && !os.IsNotExist(err) {
				t.Fatalf("rejected SVG exists at %s", savedPath)
			}
			if test.wantSaved {
				saved, err := os.ReadFile(savedPath)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := sanitizeSVG(saved); err != nil {
					t.Fatalf("stored SVG is not canonical safe SVG: %v", err)
				}
			}
		})
	}
}

func TestServeSVGRevalidatesContentAndSetsIsolationHeaders(t *testing.T) {
	cfg, documentDir := svgTestConfig(t)
	cfg.Wiki.DisableFileUploadChecking = true

	if err := os.WriteFile(filepath.Join(documentDir, "safe.svg"), []byte(safeSVGFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(documentDir, "unsafe.svg"), []byte(`<svg xmlns="http://www.w3.org/2000/svg"><circle onload="alert(1)"/></svg>`), 0o644); err != nil {
		t.Fatal(err)
	}

	safeResponse := httptest.NewRecorder()
	ServeFileHandler(safeResponse, httptest.NewRequest(http.MethodGet, "/api/files/security/safe.svg", nil), cfg)
	if safeResponse.Code != http.StatusOK {
		t.Fatalf("safe SVG status = %d; body: %s", safeResponse.Code, safeResponse.Body.String())
	}
	if got := safeResponse.Header().Get("Content-Type"); got != "image/svg+xml" {
		t.Errorf("Content-Type = %q", got)
	}
	if got := safeResponse.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q", got)
	}
	if got := safeResponse.Header().Get("Content-Disposition"); !strings.HasPrefix(got, "attachment;") {
		t.Errorf("Content-Disposition = %q", got)
	}
	csp := safeResponse.Header().Get("Content-Security-Policy")
	for _, directive := range []string{"default-src 'none'", "script-src 'none'", "style-src 'none'", "sandbox"} {
		if !strings.Contains(csp, directive) {
			t.Errorf("CSP omitted %q: %s", directive, csp)
		}
	}

	unsafeResponse := httptest.NewRecorder()
	ServeFileHandler(unsafeResponse, httptest.NewRequest(http.MethodGet, "/api/files/security/unsafe.svg", nil), cfg)
	if unsafeResponse.Code != http.StatusNotFound {
		t.Fatalf("unsafe stored SVG status = %d, want 404; body: %s", unsafeResponse.Code, unsafeResponse.Body.String())
	}
	if got := unsafeResponse.Header().Get("Content-Type"); strings.HasPrefix(got, "image/svg+xml") {
		t.Fatalf("rejected SVG error was mislabeled as renderable SVG: %q", got)
	}
}

func svgTestConfig(t *testing.T) (*config.Config, string) {
	t.Helper()
	root := t.TempDir()
	documentDir := filepath.Join(root, "documents", "security")
	if err := os.MkdirAll(documentDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{}
	cfg.Wiki.RootDir = root
	cfg.Wiki.DocumentsDir = "documents"
	cfg.Wiki.MaxUploadSize = 1
	return cfg, documentDir
}

func editorSessionCookie(t *testing.T, cfg *config.Config) *http.Cookie {
	t.Helper()
	if err := auth.InitSessionStore(filepath.Join(t.TempDir(), "sessions.json")); err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	if err := auth.CreateSession(response, "svg-editor", config.RoleEditor, nil, false, cfg); err != nil {
		t.Fatal(err)
	}
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == "session_token" {
			return cookie
		}
	}
	t.Fatal("session cookie was not created")
	return nil
}

func newSVGUploadRequest(t *testing.T, filename, content string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("docPath", "security"); err != nil {
		t.Fatal(err)
	}
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/files/upload", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}
