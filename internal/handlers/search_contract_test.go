package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"wiki-go/internal/config"
)

func TestSearchRequestContract(t *testing.T) {
	cfg := installUserSessionTestConfig(t, nil)
	cfg.Wiki.DocumentsDir = "documents"
	root := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir)
	writeChapterLinksDocument(t, filepath.Join(root, "public", "document.md"), "# Visible\nunique-needle")
	writeChapterLinksDocument(t, filepath.Join(root, "private", "document.md"), "# Hidden\nunique-needle")
	writeChapterLinksDocument(t, filepath.Join(root, "encoded", "document.md"), "# Literal%2B\nURL decoding control")
	cfg.AccessRules = []config.AccessRule{{Pattern: "/private", Access: "private"}}
	viewer := sessionCookieForUser(t, cfg, "viewer", config.RoleViewer)

	request := func(method, target, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, target, strings.NewReader(body))
		if cookie != nil {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		SearchHandler(w, r, cfg)
		if got := w.Header().Get("Cache-Control"); got != "private, no-store" {
			t.Fatalf("%s %s: Cache-Control = %q", method, target, got)
		}
		return w
	}

	get := request(http.MethodGet, "/api/search?q=unique-needle", "", nil)
	post := request(http.MethodPost, "/api/search", `{"query":"unique-needle"}`, nil)
	if get.Code != http.StatusOK || post.Code != http.StatusOK || get.Body.String() != post.Body.String() {
		t.Fatalf("GET/POST mismatch: GET %d %s POST %d %s", get.Code, get.Body.String(), post.Code, post.Body.String())
	}
	var results []SearchResult
	if err := json.Unmarshal(get.Body.Bytes(), &results); err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Path != "/public" {
		t.Fatalf("anonymous results = %+v", results)
	}
	encoded := request(http.MethodGet, "/api/search?q=Literal%252B", "", nil)
	if encoded.Code != http.StatusOK || !strings.Contains(encoded.Body.String(), `"path":"/encoded"`) {
		t.Fatalf("query was decoded more than once: %d %s", encoded.Code, encoded.Body.String())
	}
	if query, err := searchQueryFromURL("q=C%2B%2B"); err != nil || query != "C++" {
		t.Fatalf("literal plus decode = %q, %v", query, err)
	}
	if private := request(http.MethodGet, "/api/search?q=unique-needle", "", viewer); private.Code != http.StatusOK || !strings.Contains(private.Body.String(), `"path":"/private"`) {
		t.Fatalf("authenticated search = %d %s", private.Code, private.Body.String())
	}

	for _, target := range []string{"/api/search", "/api/search?q=", "/api/search?q=+"} {
		response := request(http.MethodGet, target, "", nil)
		if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
			t.Fatalf("blank GET %s = %d %s", target, response.Code, response.Body.String())
		}
	}
	if response := request(http.MethodPost, "/api/search", `{"query":" "}`, nil); response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
		t.Fatalf("blank POST = %d %s", response.Code, response.Body.String())
	}
	for _, tc := range []struct {
		method, target, body string
		status               int
	}{
		{http.MethodGet, "/api/search?q=one&q=two", "", http.StatusBadRequest},
		{http.MethodPost, "/api/search", "{", http.StatusBadRequest},
		{http.MethodPut, "/api/search", "", http.StatusMethodNotAllowed},
	} {
		response := request(tc.method, tc.target, tc.body, nil)
		if response.Code != tc.status {
			t.Fatalf("%s %s = %d, want %d", tc.method, tc.target, response.Code, tc.status)
		}
		if tc.status == http.StatusMethodNotAllowed && response.Header().Get("Allow") != "GET, POST" {
			t.Fatalf("Allow = %q", response.Header().Get("Allow"))
		}
	}

	cfg.Wiki.Private = true
	cfg.AccessRules = []config.AccessRule{{Pattern: "/public", Access: "public"}}
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		target, body := "/api/search?q=unique-needle", ""
		if method == http.MethodPost {
			target, body = "/api/search", `{"query":"unique-needle"}`
		}
		response := request(method, target, body, nil)
		if response.Code != http.StatusUnauthorized || strings.Contains(response.Body.String(), "Visible") {
			t.Fatalf("private anonymous %s = %d %s", method, response.Code, response.Body.String())
		}
	}
	if response := request(http.MethodGet, "/api/search?q=unique-needle", "", viewer); response.Code != http.StatusOK {
		t.Fatalf("private authenticated = %d %s", response.Code, response.Body.String())
	}
}

func TestSearchPageContract(t *testing.T) {
	cfg := installUserSessionTestConfig(t, nil)
	cfg.Wiki.DocumentsDir = "documents"
	root := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir)
	writeChapterLinksDocument(t, filepath.Join(root, "public", "document.md"), "# Visible\nneedle")
	writeChapterLinksDocument(t, filepath.Join(root, "secret", "document.md"), "# SECRET_TITLE\nneedle SECRET_EXCERPT")
	cfg.AccessRules = []config.AccessRule{{Pattern: "/secret", Access: "private"}}
	viewer := sessionCookieForUser(t, cfg, "viewer", config.RoleViewer)

	request := func(method, target string, cookie *http.Cookie) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, target, nil)
		if cookie != nil {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		SearchPageHandler(w, r, cfg)
		if got := w.Header().Get("Cache-Control"); got != "private, no-store" {
			t.Fatalf("Cache-Control = %q", got)
		}
		return w
	}

	for _, target := range []string{"/api/search/page", "/api/search/page?q="} {
		response := request(http.MethodGet, target, nil)
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `name="q"`) || strings.Contains(response.Body.String(), "SECRET_TITLE") {
			t.Fatalf("empty search %s = %d %s", target, response.Code, response.Body.String())
		}
	}
	response := request(http.MethodGet, "/api/search/page?q=needle", nil)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `href="/public"`) || strings.Contains(response.Body.String(), "SECRET_TITLE") || strings.Contains(response.Body.String(), "SECRET_EXCERPT") {
		t.Fatalf("anonymous HTML search = %d %s", response.Code, response.Body.String())
	}
	response = request(http.MethodGet, "/api/search/page?q=needle", viewer)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "SECRET_TITLE") {
		t.Fatalf("authenticated HTML search = %d %s", response.Code, response.Body.String())
	}
	response = request(http.MethodGet, "/api/search/page?q=%3Cscript%3E", nil)
	if response.Code != http.StatusOK || strings.Contains(response.Body.String(), `<script>`) || !strings.Contains(response.Body.String(), `value="&lt;script&gt;"`) {
		t.Fatalf("unsafe HTML query = %d %s", response.Code, response.Body.String())
	}
	for _, tc := range []struct {
		method, target string
		status         int
	}{
		{http.MethodGet, "/api/search/page?q=one&q=two", http.StatusBadRequest},
		{http.MethodPost, "/api/search/page", http.StatusMethodNotAllowed},
	} {
		response := request(tc.method, tc.target, nil)
		if response.Code != tc.status {
			t.Fatalf("%s %s = %d, want %d", tc.method, tc.target, response.Code, tc.status)
		}
	}
	cfg.Wiki.Private = true
	response = request(http.MethodGet, "/api/search/page?q=needle%20two", nil)
	if response.Code != http.StatusSeeOther || response.Header().Get("Location") != "/login?redirect=%2Fapi%2Fsearch%2Fpage%3Fq%3Dneedle%2520two" || strings.Contains(response.Body.String(), "SECRET_TITLE") {
		t.Fatalf("private search redirect = %d %s, body=%s", response.Code, response.Header().Get("Location"), response.Body.String())
	}
	response = request(http.MethodGet, "/api/search/page?q=needle", viewer)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "SECRET_TITLE") {
		t.Fatalf("private authenticated page = %d %s", response.Code, response.Body.String())
	}
}

func TestSearchQueryFromURLRejectsMalformedEscape(t *testing.T) {
	if _, err := searchQueryFromURL("q=%ZZ"); err == nil {
		t.Fatal("malformed URL query was accepted")
	}
}
