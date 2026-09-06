package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"wiki-go/internal/config"
)

const chapterLinksIntegrationMarkdown = `# Safe <img src=x onerror="alert(1)">

## Child heading

> [!IMPORTANT]
> Trusted alert content.

:::stats count=*:::
`

func TestChapterLinksRenderThroughPageAndHomeHandlers(t *testing.T) {
	testConfig, _ := newChapterLinksHandlerTestConfig(t)
	writeChapterLinksDocument(t, filepath.Join(testConfig.Wiki.RootDir, "documents", "guide", "document.md"), chapterLinksIntegrationMarkdown)
	writeChapterLinksDocument(t, filepath.Join(testConfig.Wiki.RootDir, "pages", "home", "document.md"), chapterLinksIntegrationMarkdown)

	for _, test := range []struct {
		name string
		path string
	}{
		{name: "document page", path: "/guide"},
		{name: "home page", path: "/"},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()
			PageHandler(response, request, testConfig)

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body: %s", response.Code, response.Body.String())
			}
			assertRenderedPageHasSafeChapterLinks(t, response.Body.String())
		})
	}
}

func TestChapterLinksStayOutOfEditAndStructuredLayouts(t *testing.T) {
	testConfig, editorCookie := newChapterLinksHandlerTestConfig(t)
	documents := map[string]string{
		"editable": chapterLinksIntegrationMarkdown,
		"kanban": `---
layout: kanban
---
# Kanban heading

#### Board
##### Column
- [ ] Task
`,
		"links": `---
layout: links
---
# Links heading

## Category
- [Example](https://example.com)
`,
	}
	for name, markdown := range documents {
		writeChapterLinksDocument(t, filepath.Join(testConfig.Wiki.RootDir, "documents", name, "document.md"), markdown)
	}

	tests := []struct {
		name       string
		path       string
		auth       bool
		expectPage string
	}{
		{name: "editor", path: "/editable?mode=edit", auth: true, expectPage: `class="editor-container"`},
		{name: "kanban", path: "/kanban", expectPage: `/static/js/kanban-core.js`},
		{name: "links", path: "/links", expectPage: `/static/js/links.js`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			if test.auth {
				request.AddCookie(editorCookie)
			}
			response := httptest.NewRecorder()
			PageHandler(response, request, testConfig)

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body: %s", response.Code, response.Body.String())
			}
			rendered := response.Body.String()
			if !strings.Contains(rendered, test.expectPage) {
				t.Errorf("%s response omitted %q", test.name, test.expectPage)
			}
			if strings.Contains(rendered, `<aside class="chapter-links"`) {
				t.Fatalf("%s response rendered the chapter-links panel", test.name)
			}
		})
	}
}

func TestHistoryPreviewUsesSafeMarkdownRenderer(t *testing.T) {
	testConfig, editorCookie := newChapterLinksHandlerTestConfig(t)
	writeChapterLinksDocument(t, filepath.Join(testConfig.Wiki.RootDir, "documents", "history", "document.md"), chapterLinksIntegrationMarkdown)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/render-markdown?path=history",
		strings.NewReader(chapterLinksIntegrationMarkdown),
	)
	request.AddCookie(editorCookie)
	response := httptest.NewRecorder()
	RenderMarkdownHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", response.Code, response.Body.String())
	}
	rendered := response.Body.String()
	for _, forbidden := range []string{"<img src=x", `<script`, ` onerror="`, ` onload="`, ` onclick="`} {
		if strings.Contains(strings.ToLower(rendered), forbidden) {
			t.Fatalf("history preview returned active markup %q: %s", forbidden, rendered)
		}
	}
	for _, expected := range []string{
		`Safe &lt;img src=x onerror=&quot;alert(1)&quot;&gt;`,
		`class="markdown-alert markdown-alert-important"`,
		`class="wiki-stats doc-count"`,
	} {
		if !strings.Contains(rendered, expected) {
			t.Errorf("history preview omitted trusted renderer output %q: %s", expected, rendered)
		}
	}
}

func newChapterLinksHandlerTestConfig(t *testing.T) (*config.Config, *http.Cookie) {
	t.Helper()
	testConfig := installUserSessionTestConfig(t, nil)
	testConfig.Wiki.DocumentsDir = "documents"
	testConfig.Wiki.Title = "Test Wiki"
	testConfig.Wiki.Owner = "wiki.example"
	testConfig.Wiki.Timezone = "UTC"
	testConfig.Wiki.Language = "en"
	testConfig.Wiki.DisableComments = true
	testConfig.Wiki.HideAttachments = true
	testConfig.Server.AllowInsecureCookies = true
	for _, directory := range []string{
		filepath.Join(testConfig.Wiki.RootDir, "documents"),
		filepath.Join(testConfig.Wiki.RootDir, "pages", "home"),
	} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return testConfig, sessionCookieForUser(t, testConfig, "editor", config.RoleEditor)
}

func writeChapterLinksDocument(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertRenderedPageHasSafeChapterLinks(t *testing.T, rendered string) {
	t.Helper()
	if !strings.Contains(rendered, `<aside class="chapter-links"`) {
		t.Fatal("rendered page omitted chapter links")
	}
	for _, forbidden := range []string{"<img src=x", `<svg onload`, ` onerror="`, ` onload="`, ` onclick="`} {
		if strings.Contains(strings.ToLower(rendered), forbidden) {
			t.Fatalf("rendered page contains active heading markup %q", forbidden)
		}
	}
	for _, expected := range []string{
		`Safe &lt;img src=x onerror=&#34;alert(1)&#34;&gt;`,
		`class="markdown-alert markdown-alert-important"`,
		`class="wiki-stats doc-count"`,
	} {
		if !strings.Contains(rendered, expected) {
			t.Errorf("rendered page omitted %q", expected)
		}
	}

	match := regexp.MustCompile(`<h1 id="([^"]+)"`).FindStringSubmatch(rendered)
	if len(match) != 2 {
		t.Fatalf("rendered page omitted the trusted heading ID: %s", rendered)
	}
	if strings.Count(rendered, `href="#`+match[1]+`"`) != 2 {
		t.Fatalf("heading ID %q does not match its permalink and chapter link: %s", match[1], rendered)
	}
}
