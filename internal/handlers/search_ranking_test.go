package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"wiki-go/internal/config"
)

func TestSearchTitleRanking(t *testing.T) {
	cfg := &config.Config{}
	cfg.Wiki.RootDir = t.TempDir()
	cfg.Wiki.DocumentsDir = "documents"
	root := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir)
	for path, content := range map[string]string{
		"a-body":          "# Overview\nalpha beta new page in the body",
		"a-no-title":      "Untitled alpha beta new page in the body",
		"b-title":         "# Alpha Beta New Page SECRET_TITLE\nmatching document",
		"c-title":         "# ALPHA BETA NEW PAGE\nmatching document",
		"d-partial":       "# Alpha Overview\nbeta new page in the body",
		"e-empty-title":   "# \nalpha beta new page in the body",
		"f-excluded":      "# Alpha Beta New Page\nforbidden in the body",
		"g-unicode":       "# CAFÉ Zürich\nmatching document",
		"z-real-untitled": "# Untitled\nmatching document",
	} {
		writeChapterLinksDocument(t, filepath.Join(root, path, "document.md"), content)
	}

	for _, tc := range []struct {
		name, query string
		paths       []string
	}{
		{"multiple words with excluded body", "alpha beta NOT forbidden", []string{"/b-title", "/c-title", "/a-body", "/a-no-title", "/d-partial", "/e-empty-title"}},
		{"quoted phrase with excluded body", `"new page" NOT forbidden`, []string{"/b-title", "/c-title", "/a-body", "/a-no-title", "/d-partial", "/e-empty-title"}},
		{"phrase and words must all match title", `"new page" alpha beta NOT forbidden`, []string{"/b-title", "/c-title", "/a-body", "/a-no-title", "/d-partial", "/e-empty-title"}},
		{"positive title and body", "alpha beta", []string{"/b-title", "/c-title", "/f-excluded", "/a-body", "/a-no-title", "/d-partial", "/e-empty-title"}},
		{"synthetic title has no bonus", "Untitled", []string{"/z-real-untitled", "/a-no-title"}},
		{"unicode case folding", "café", []string{"/g-unicode"}},
		{"negative only keeps path order", "NOT forbidden", []string{"/a-body", "/a-no-title", "/b-title", "/c-title", "/d-partial", "/e-empty-title", "/g-unicode", "/z-real-untitled"}},
		{"empty terms keep path order", "", []string{"/a-body", "/a-no-title", "/b-title", "/c-title", "/d-partial", "/e-empty-title", "/f-excluded", "/g-unicode", "/z-real-untitled"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			results := performSearch(tc.query, nil, cfg)
			paths := make([]string, len(results))
			seen := make(map[string]bool)
			for i, result := range results {
				if seen[result.Path] {
					t.Fatalf("duplicate result at %q", result.Path)
				}
				seen[result.Path] = true
				paths[i] = result.Path
				if result.Title == "" || result.Excerpt == "" {
					t.Fatalf("result missing title or excerpt: %+v", result)
				}
			}
			if !reflect.DeepEqual(paths, tc.paths) {
				t.Fatalf("paths = %v, want %v", paths, tc.paths)
			}
		})
	}

	cfg.AccessRules = []config.AccessRule{{Pattern: "/b-title", Access: "private"}}
	results := performSearch("alpha beta", nil, cfg)
	if len(results) == 0 || results[0].Path != "/c-title" {
		t.Fatalf("restricted title affected ranking: %+v", results)
	}
	for _, result := range results {
		if result.Path == "/b-title" || strings.Contains(result.Title, "SECRET_TITLE") {
			t.Fatalf("restricted title leaked: %+v", result)
		}
	}

	request := httptest.NewRequest(http.MethodGet, "/api/search?q=alpha+beta", nil)
	response := httptest.NewRecorder()
	SearchHandler(response, request, cfg)
	if response.Code != http.StatusOK || strings.Contains(response.Body.String(), `"path":"/b-title"`) {
		t.Fatalf("restricted API title leaked: %d %s", response.Code, response.Body.String())
	}
	var apiResults []SearchResult
	if err := json.Unmarshal(response.Body.Bytes(), &apiResults); err != nil || len(apiResults) == 0 || apiResults[0].Path != "/c-title" {
		t.Fatalf("ranked API results = %+v, error = %v", apiResults, err)
	}

	pageResponse := httptest.NewRecorder()
	SearchPageHandler(pageResponse, httptest.NewRequest(http.MethodGet, "/api/search/page?q=alpha+beta", nil), cfg)
	if pageResponse.Code != http.StatusOK || strings.Contains(pageResponse.Body.String(), "SECRET_TITLE") || strings.Contains(pageResponse.Body.String(), `href="/b-title"`) {
		t.Fatalf("restricted HTML title leaked: %d %s", pageResponse.Code, pageResponse.Body.String())
	}
}
