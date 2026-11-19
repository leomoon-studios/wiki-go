package routes

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"wiki-go/internal/auth"
	"wiki-go/internal/basepath"
	"wiki-go/internal/config"
)

func newBasePathTestConfig(t *testing.T) *config.Config {
	t.Helper()
	previous := basepath.Get()
	basepath.Set("/wiki")
	t.Cleanup(func() { basepath.Set(previous) })
	cfg := newRouteSecurityTestConfig(t)
	cfg.Server.BasePath = "/wiki"
	return cfg
}

func serveBasePath(target string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, target, nil)
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	recorder := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(recorder, request)
	return recorder
}

func TestBasePathAutomaticRedirects(t *testing.T) {
	for _, base := range []string{"", "/wiki", "/docs/wiki-go", "/sitemap"} {
		t.Run("base="+base, func(t *testing.T) {
			previous := basepath.Get()
			basepath.Set(base)
			t.Cleanup(func() { basepath.Set(previous) })
			newRouteSecurityTestConfig(t)

			for _, path := range []string{"/sitemap", "/static", "/api/versions", "/%73itemap"} {
				for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodPost} {
					t.Run(method+path, func(t *testing.T) {
						const query = "?q=one%20two&next=%2Fguide"
						target := base + path + query
						request := httptest.NewRequest(method, target, nil)
						response := httptest.NewRecorder()
						http.DefaultServeMux.ServeHTTP(response, request)
						if response.Code != http.StatusTemporaryRedirect {
							t.Fatalf("%s: status = %d, want %d", target, response.Code, http.StatusTemporaryRedirect)
						}
						want := base + path + "/" + query
						if got := response.Header().Get("Location"); got != want {
							t.Fatalf("%s: Location = %q, want %q", target, got, want)
						}
					})
				}
			}

			response := serveBasePath(base + "/sitemap/")
			if response.Code != http.StatusOK {
				t.Fatalf("sitemap after redirect: status = %d; body: %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestBasePathServesPagesAssetsAndLinks(t *testing.T) {
	cfg := newBasePathTestConfig(t)
	doc := filepath.Join(cfg.Wiki.RootDir, "documents", "guide", "document.md")
	if err := os.MkdirAll(filepath.Dir(doc), 0o755); err != nil {
		t.Fatal(err)
	}
	markdown := "# BASE_PATH_CANARY\n\n[Install](/guide/install) [Docs](https://example.com/docs) ![Logo](logo.png)\n\n```\n<a href=\"/guide/code\">\n```\n"
	if err := os.WriteFile(doc, []byte(markdown), 0o644); err != nil {
		t.Fatal(err)
	}

	page := serveBasePath("/wiki/guide")
	body := page.Body.String()
	if page.Code != http.StatusOK || !strings.Contains(body, "BASE_PATH_CANARY") {
		t.Fatalf("page under base path = %d", page.Code)
	}
	for _, expected := range []string{
		`<meta name="wiki-base-path" content="/wiki">`,
		`<script src="/wiki/static/js/base-path.js?=`,
		`href="/wiki/static/css/theme.css?=`,
		`href="/wiki/opensearch.xml"`,
		`href="/wiki/manifest.json"`,
		`href="/wiki/guide/install"`,
		`href="https://example.com/docs"`,
		`src="/wiki/api/files/guide/logo.png"`,
		`&lt;a href=&#34;/guide/code&#34;&gt;`,
	} {
		if !strings.Contains(body, expected) && !strings.Contains(body, strings.ReplaceAll(expected, "&#34;", "&quot;")) {
			t.Errorf("page is missing %q", expected)
		}
	}
	if match := regexp.MustCompile(`\s(?:href|src|action)="/(?:[^w/"]|w[^i])[^"]*"`).FindString(body); match != "" {
		t.Errorf("page contains a URL outside the base path: %s", match)
	}
	home := serveBasePath("/wiki/")
	if link := home.Header().Get("Link"); home.Code != http.StatusOK || !strings.Contains(link, "</wiki/api/data/emojis>") {
		t.Errorf("home = %d, preload link = %q", home.Code, link)
	}

	for _, target := range []string{"/wiki/static/js/base-path.js", "/wiki/api/data/emojis"} {
		if response := serveBasePath(target); response.Code != http.StatusOK {
			t.Errorf("%s = %d", target, response.Code)
		}
	}
	if response := serveBasePath("/guide"); response.Code != http.StatusNotFound {
		t.Errorf("unprefixed page = %d, want 404", response.Code)
	}
	if response := serveBasePath("/"); response.Code != http.StatusFound || response.Header().Get("Location") != "/wiki/" {
		t.Errorf("root = %d %q", response.Code, response.Header().Get("Location"))
	}
}

func TestBasePathRedirectsCookiesAndPublicURLs(t *testing.T) {
	cfg := newBasePathTestConfig(t)
	cfg.Server.PublicURL = "https://example.com/wiki/"
	cfg.Wiki.Private = true

	redirect := serveBasePath("/wiki/guide?mode=edit")
	if redirect.Code != http.StatusSeeOther || redirect.Header().Get("Location") != "/wiki/login?redirect=%2Fguide%3Fmode%3Dedit" {
		t.Fatalf("private redirect = %d %q", redirect.Code, redirect.Header().Get("Location"))
	}
	login := serveBasePath("/wiki/login")
	if login.Code != http.StatusOK || !strings.Contains(login.Body.String(), `<script src="/wiki/static/js/login.js?=`) {
		t.Fatalf("login page = %d", login.Code)
	}

	session := httptest.NewRecorder()
	if err := auth.CreateSession(session, "admin", config.RoleAdmin, nil, false, cfg); err != nil {
		t.Fatal(err)
	}
	var token *http.Cookie
	for _, cookie := range session.Result().Cookies() {
		if cookie.Path != "/wiki" {
			t.Errorf("cookie %s path = %q, want /wiki", cookie.Name, cookie.Path)
		}
		if cookie.Name == "session_token" {
			token = cookie
		}
	}
	if response := serveBasePath("/wiki/login", token); response.Code != http.StatusSeeOther || response.Header().Get("Location") != "/wiki/" {
		t.Fatalf("logged-in login redirect = %d %q", response.Code, response.Header().Get("Location"))
	}

	descriptor := serveBasePath("/wiki/opensearch.xml", token)
	if !strings.Contains(descriptor.Body.String(), "https://example.com/wiki/?search={searchTerms}") {
		t.Fatalf("OpenSearch template = %s", descriptor.Body.String())
	}
	sitemap := serveBasePath("/wiki/sitemap.xml", token)
	if !strings.Contains(sitemap.Body.String(), "<loc>https://example.com/wiki/</loc>") {
		t.Fatalf("sitemap = %d %s", sitemap.Code, sitemap.Body.String())
	}
	manifest := serveBasePath("/wiki/manifest.json")
	for _, expected := range []string{`"start_url":"/wiki/"`, `"scope":"/wiki/"`, `"src":"/wiki/static/favicon.png"`} {
		if !strings.Contains(manifest.Body.String(), expected) {
			t.Errorf("manifest is missing %s: %s", expected, manifest.Body.String())
		}
	}
}

// Templates and scripts must route wiki URLs through the base path helpers;
// a literal root path would break wikis served from a subdirectory.
func TestTemplatesAndScriptsDoNotHardcodeRootURLs(t *testing.T) {
	templateURL := regexp.MustCompile(`\b(?:href|src|action|data-[a-z-]+-url)="/`)
	scriptURL := regexp.MustCompile("\\b(?:fetch|loadScript)\\(['\"`]/|\\.href = ['\"`]/")
	check := func(pattern *regexp.Regexp, dir, suffix string) {
		entries, err := os.ReadDir(filepath.Join("..", "resources", dir))
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if !strings.HasSuffix(entry.Name(), suffix) {
				continue
			}
			content, err := os.ReadFile(filepath.Join("..", "resources", dir, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			for _, line := range strings.Split(string(content), "\n") {
				if pattern.MatchString(line) {
					t.Errorf("%s/%s hardcodes a root URL: %s", dir, entry.Name(), strings.TrimSpace(line))
				}
			}
		}
	}
	check(templateURL, "templates", ".html")
	check(scriptURL, filepath.Join("static", "js"), ".js")
}
