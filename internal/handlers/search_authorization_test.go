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

func TestSearchAuthorization(t *testing.T) {
	cfg := installUserSessionTestConfig(t, nil)
	cfg.Wiki.DocumentsDir = "documents"
	root := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir)
	writeChapterLinksDocument(t, filepath.Join(root, "hr", "payroll", "document.md"), "# SECRET_TITLE\nsalary secret phrase SECRET_EXCERPT")
	writeChapterLinksDocument(t, filepath.Join(root, "public", "document.md"), "# PUBLIC_TITLE\nsalary secret phrase PUBLIC_EXCERPT")
	writeChapterLinksDocument(t, filepath.Join(root, "public", "attachment.md"), "# ATTACHMENT_CANARY\nsalary secret phrase")
	writeChapterLinksDocument(t, filepath.Join(root, "document.md-notes", "document.md"), "# DIRECTORY_CANARY\nsalary secret phrase")
	viewer := sessionCookieForUser(t, cfg, "viewer", config.RoleViewer)
	member := sessionCookieForUserWithGroups(t, cfg, "member", config.RoleViewer, []string{"hr"})
	admin := sessionCookieForUser(t, cfg, "admin", config.RoleAdmin)

	for _, tc := range []struct {
		name    string
		rules   []config.AccessRule
		private bool
		cookie  *http.Cookie
		allowed bool
	}{
		{"public exact anonymous", []config.AccessRule{{Pattern: "/hr/payroll", Access: "restricted", Groups: []string{"hr"}}}, false, nil, false},
		{"public single anonymous", []config.AccessRule{{Pattern: "/hr/*", Access: "restricted", Groups: []string{"hr"}}}, false, nil, false},
		{"public recursive anonymous", []config.AccessRule{{Pattern: "/hr/**", Access: "restricted", Groups: []string{"hr"}}}, false, nil, false},
		{"private viewer", []config.AccessRule{{Pattern: "/hr/payroll", Access: "restricted", Groups: []string{"hr"}}}, true, viewer, false},
		{"private member", []config.AccessRule{{Pattern: "/hr/payroll", Access: "restricted", Groups: []string{"hr"}}}, true, member, true},
		{"private admin", []config.AccessRule{{Pattern: "/hr/payroll", Access: "restricted", Groups: []string{"hr"}}}, true, admin, true},
		{"deny before allow", []config.AccessRule{{Pattern: "/hr/payroll", Access: "restricted", Groups: []string{"hr"}}, {Pattern: "/hr/**", Access: "public"}}, false, nil, false},
		{"allow before deny", []config.AccessRule{{Pattern: "/hr/payroll", Access: "public"}, {Pattern: "/hr/**", Access: "restricted", Groups: []string{"hr"}}}, false, nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg.AccessRules, cfg.Wiki.Private = tc.rules, tc.private
			pageRequest := httptest.NewRequest(http.MethodGet, "/hr/payroll", nil)
			if tc.cookie != nil {
				pageRequest.AddCookie(tc.cookie)
			}
			pageResponse := httptest.NewRecorder()
			PageHandler(pageResponse, pageRequest, cfg)
			if tc.allowed && pageResponse.Code == http.StatusForbidden || !tc.allowed && pageResponse.Code != http.StatusForbidden && pageResponse.Code != http.StatusFound {
				t.Fatalf("direct page status = %d, allowed=%t", pageResponse.Code, tc.allowed)
			}
			for _, query := range []string{"salary", `"secret phrase"`, ""} {
				body, _ := json.Marshal(SearchRequest{Query: query})
				request := httptest.NewRequest(http.MethodPost, "/api/search", strings.NewReader(string(body)))
				if tc.cookie != nil {
					request.AddCookie(tc.cookie)
				}
				response := httptest.NewRecorder()
				SearchHandler(response, request, cfg)
				if response.Code != http.StatusOK {
					t.Fatalf("status = %d", response.Code)
				}
				var results []SearchResult
				if err := json.Unmarshal(response.Body.Bytes(), &results); err != nil {
					t.Fatal(err)
				}
				foundSecret, foundPublic, foundDirectory, foundAttachment := false, false, false, false
				for _, result := range results {
					if result.Path == "/hr/payroll" && result.Title == "SECRET_TITLE" && strings.Contains(strings.ToLower(result.Excerpt), "secret_excerpt") {
						foundSecret = true
					}
					if result.Path == "/public" && result.Title == "PUBLIC_TITLE" {
						foundPublic = true
					}
					if result.Path == "/document.md-notes" && result.Title == "DIRECTORY_CANARY" {
						foundDirectory = true
					}
					if strings.Contains(result.Title, "ATTACHMENT_CANARY") {
						foundAttachment = true
					}
				}
				if !tc.allowed && (strings.Contains(response.Body.String(), "SECRET_TITLE") || strings.Contains(strings.ToLower(response.Body.String()), "secret_excerpt") || strings.Contains(response.Body.String(), "/hr/payroll")) {
					t.Fatalf("restricted search disclosure: %s", response.Body.String())
				}
				if foundAttachment {
					t.Fatal("Markdown attachment was indexed")
				}
				if foundSecret != tc.allowed || !foundPublic || !foundDirectory {
					t.Fatalf("query %q: secret=%t public=%t directory=%t; results=%s", query, foundSecret, foundPublic, foundDirectory, response.Body.String())
				}
			}
		})
	}
}
