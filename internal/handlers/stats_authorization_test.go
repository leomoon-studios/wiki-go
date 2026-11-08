package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"wiki-go/internal/config"
	"wiki-go/internal/goldext"
)

func TestStatsRespectDocumentAccessAcrossRenderPaths(t *testing.T) {
	testConfig, _ := newChapterLinksHandlerTestConfig(t)
	testConfig.AccessRules = []config.AccessRule{{Pattern: "/finance/**", Access: "restricted", Groups: []string{"finance"}}}
	previousRoot := goldext.DocumentsRoot()
	goldext.ConfigureContentPaths(testConfig.Wiki.RootDir, testConfig.Wiki.DocumentsDir)
	t.Cleanup(func() { goldext.ConfigureContentPaths(filepath.Dir(previousRoot), filepath.Base(previousRoot)) })
	goldext.SetWikiTimezone("UTC")
	t.Cleanup(func() { goldext.SetWikiTimezone("") })

	const secret = "RESTRICTED_FINANCE_TITLE_CANARY"
	const public = "PUBLIC_DOCUMENT_CANARY"
	const markdown = "# Dashboard\n\n:::stats recent=100:::\n\n:::stats count=*:::\n\n:::stats count=finance:::\n"
	writeChapterLinksDocument(t, filepath.Join(testConfig.Wiki.RootDir, "documents", "finance", "secret", "document.md"), "# "+secret)
	writeChapterLinksDocument(t, filepath.Join(testConfig.Wiki.RootDir, "documents", "public", "document.md"), "# "+public)
	writeChapterLinksDocument(t, filepath.Join(testConfig.Wiki.RootDir, "documents", "dashboard", "document.md"), markdown)
	writeChapterLinksDocument(t, filepath.Join(testConfig.Wiki.RootDir, "pages", "home", "document.md"), markdown)
	secretPath := filepath.Join(testConfig.Wiki.RootDir, "documents", "finance", "secret", "document.md")
	secretTime := time.Date(2020, 2, 3, 4, 5, 0, 0, time.UTC)
	if err := os.Chtimes(secretPath, secretTime, secretTime); err != nil {
		t.Fatal(err)
	}

	viewer := sessionCookieForUser(t, testConfig, "viewer", config.RoleViewer)
	member := sessionCookieForUserWithGroups(t, testConfig, "finance", config.RoleViewer, []string{"finance"})
	admin := sessionCookieForUser(t, testConfig, "admin", config.RoleAdmin)
	deniedRequest := httptest.NewRequest(http.MethodGet, "/finance/secret", nil)
	deniedRequest.AddCookie(viewer)
	deniedResponse := httptest.NewRecorder()
	PageHandler(deniedResponse, deniedRequest, testConfig)
	if deniedResponse.Code != http.StatusForbidden {
		t.Fatalf("restricted page status = %d, want 403", deniedResponse.Code)
	}
	const stats = ":::stats recent=100:::\n\n:::stats count=*:::\n\n:::stats count=finance:::"
	for _, test := range []struct {
		name, method, path, body string
		cookie                   *http.Cookie
		allowed                  bool
	}{
		{"preview viewer", "POST", "/api/render-markdown", stats, viewer, false},
		{"preview source lines", "POST", "/api/render-markdown?source_lines=1", stats, viewer, false},
		{"preview kanban", "POST", "/api/render-markdown", "---\nlayout: kanban\n---\n\n" + stats, viewer, false},
		{"preview member", "POST", "/api/render-markdown", stats, member, true},
		{"preview admin", "POST", "/api/render-markdown", stats, admin, true},
		{"anonymous home", "GET", "/", "", nil, false},
		{"anonymous page", "GET", "/dashboard", "", nil, false},
		{"viewer home", "GET", "/", "", viewer, false},
		{"viewer page", "GET", "/dashboard", "", viewer, false},
		{"member page", "GET", "/dashboard", "", member, true},
		{"admin page", "GET", "/dashboard", "", admin, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			if test.cookie != nil {
				request.AddCookie(test.cookie)
			}
			response := httptest.NewRecorder()
			if test.method == "POST" {
				RenderMarkdownHandler(response, request)
			} else {
				PageHandler(response, request, testConfig)
			}
			body := response.Body.String()
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d: %s", response.Code, body)
			}
			for _, value := range []string{secret, "/finance/secret", "2020-02-03 04:05"} {
				if strings.Contains(body, value) != test.allowed {
					t.Errorf("presence of %q differs from allowed=%t", value, test.allowed)
				}
			}
			if !strings.Contains(body, public) {
				t.Error("public document missing from recent stats")
			}
			wantGlobal, wantFolder := "2", "0"
			if test.allowed {
				wantGlobal, wantFolder = "3", "1"
			}
			for _, count := range []string{wantGlobal, wantFolder} {
				if !strings.Contains(body, `class="count-number">`+count+`</div>`) {
					t.Errorf("missing count %s in %s", count, body)
				}
			}
		})
	}
}
