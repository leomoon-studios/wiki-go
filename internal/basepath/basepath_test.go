package basepath

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func useBasePath(t *testing.T, path string) {
	t.Helper()
	previous := Get()
	Set(path)
	t.Cleanup(func() { Set(previous) })
}

func TestNormalize(t *testing.T) {
	for raw, want := range map[string]string{
		"":              "",
		"/":             "",
		"  ":            "",
		"wiki":          "/wiki",
		"/wiki":         "/wiki",
		"/wiki/":        "/wiki",
		"/docs/wiki-go": "/docs/wiki-go",
		"/a.b_c~d":      "/a.b_c~d",
	} {
		got, err := Normalize(raw)
		if err != nil || got != want {
			t.Errorf("Normalize(%q) = %q, %v; want %q", raw, got, err, want)
		}
	}
	for _, raw := range []string{
		"/wiki//docs", "/../wiki", "/wiki/.", "/wi ki", "/wiki?x=1", "/wiki#top",
		"/wi%20ki", `/wiki\docs`, "/wiki\"", "/wiki<", "/wïki", "https://example.com/wiki",
	} {
		if got, err := Normalize(raw); err == nil {
			t.Errorf("Normalize(%q) = %q, want error", raw, got)
		}
	}
}

func TestURL(t *testing.T) {
	useBasePath(t, "")
	if got := URL("/static/app.js"); got != "/static/app.js" {
		t.Fatalf("root URL = %q", got)
	}
	if got := CookiePath(); got != "/" {
		t.Fatalf("root cookie path = %q", got)
	}

	useBasePath(t, "/wiki")
	for input, want := range map[string]string{
		"/":                     "/wiki/",
		"/static/app.js":        "/wiki/static/app.js",
		"/login?redirect=%2Fa":  "/wiki/login?redirect=%2Fa",
		"//evil.example/x":      "//evil.example/x",
		"https://example.com/x": "https://example.com/x",
		"#section":              "#section",
		"image.png":             "image.png",
		"":                      "",
	} {
		if got := URL(input); got != want {
			t.Errorf("URL(%q) = %q, want %q", input, got, want)
		}
	}
	if got := CookiePath(); got != "/wiki" {
		t.Fatalf("cookie path = %q", got)
	}
}

func TestMiddleware(t *testing.T) {
	var seenPath, seenRawPath string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath, seenRawPath = r.URL.Path, r.URL.RawPath
		w.WriteHeader(http.StatusNoContent)
	})

	useBasePath(t, "")
	recorder := httptest.NewRecorder()
	Middleware(next).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/docs/page", nil))
	if recorder.Code != http.StatusNoContent || seenPath != "/docs/page" {
		t.Fatalf("root mode: status=%d path=%q", recorder.Code, seenPath)
	}

	useBasePath(t, "/wiki")
	for _, tc := range []struct {
		target   string
		status   int
		path     string
		rawPath  string
		location string
	}{
		{target: "/wiki/", status: http.StatusNoContent, path: "/"},
		{target: "/wiki/docs/page?mode=edit", status: http.StatusNoContent, path: "/docs/page"},
		{target: "/wiki/api/files/a%2Fb.png", status: http.StatusNoContent, path: "/api/files/a/b.png", rawPath: "/api/files/a%2Fb.png"},
		{target: "/wiki/wiki/page", status: http.StatusNoContent, path: "/wiki/page"},
		{target: "/wiki", status: http.StatusFound, location: "/wiki/"},
		{target: "/?search=x", status: http.StatusFound, location: "/wiki/?search=x"},
		{target: "/wikipedia", status: http.StatusNotFound},
		{target: "/static/app.js", status: http.StatusNotFound},
	} {
		seenPath, seenRawPath = "", ""
		recorder := httptest.NewRecorder()
		Middleware(next).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.target, nil))
		if recorder.Code != tc.status {
			t.Errorf("%s: status = %d, want %d", tc.target, recorder.Code, tc.status)
			continue
		}
		if seenPath != tc.path || seenRawPath != tc.rawPath {
			t.Errorf("%s: routed path = %q raw %q, want %q raw %q", tc.target, seenPath, seenRawPath, tc.path, tc.rawPath)
		}
		if location := recorder.Header().Get("Location"); location != tc.location {
			t.Errorf("%s: Location = %q, want %q", tc.target, location, tc.location)
		}
	}
}

func TestRewriteHTML(t *testing.T) {
	input := `<p><a href="/docs/guide">Guide</a> <a href="/">Home</a> <img src="/api/files/doc/a.png" alt="a"> ` +
		`<a href="//evil.example/x">x</a> <a href="https://example.com/">e</a> <a href="#top">t</a> <a href="relative">r</a> ` +
		`<video><source src="/api/files/doc/v.mp4"></video> ` +
		`<code>href=&quot;/docs/escaped&quot;</code> <span data-path="/docs/data">d</span></p>`

	useBasePath(t, "")
	if got := string(RewriteHTML([]byte(input))); got != input {
		t.Fatalf("root mode changed HTML: %s", got)
	}

	useBasePath(t, "/wiki")
	want := `<p><a href="/wiki/docs/guide">Guide</a> <a href="/wiki/">Home</a> <img src="/wiki/api/files/doc/a.png" alt="a"> ` +
		`<a href="//evil.example/x">x</a> <a href="https://example.com/">e</a> <a href="#top">t</a> <a href="relative">r</a> ` +
		`<video><source src="/wiki/api/files/doc/v.mp4"></video> ` +
		`<code>href=&quot;/docs/escaped&quot;</code> <span data-path="/docs/data">d</span></p>`
	if got := string(RewriteHTML([]byte(input))); got != want {
		t.Fatalf("RewriteHTML:\n got %s\nwant %s", got, want)
	}
}
