package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"wiki-go/internal/comments"
	"wiki-go/internal/config"
)

func TestCommentHandlersSupportDocumentsAndHomepage(t *testing.T) {
	testConfig := installUserSessionTestConfig(t, nil)
	testConfig.Wiki.DocumentsDir = "documents"
	testConfig.Wiki.Title = "Comment Test Wiki"
	testConfig.Wiki.Owner = "wiki.example"
	testConfig.Wiki.Timezone = "UTC"
	testConfig.Wiki.Language = "en"
	testConfig.Wiki.HideAttachments = true
	editorCookie := sessionCookieForUser(t, testConfig, "editor", config.RoleEditor)
	adminCookie := sessionCookieForUser(t, testConfig, "admin", config.RoleAdmin)

	writeCommentTestDocument(t, filepath.Join(testConfig.Wiki.RootDir, "documents", "guides", "start", "document.md"))
	writeCommentTestDocument(t, filepath.Join(testConfig.Wiki.RootDir, "pages", "home", "document.md"))

	tests := []struct {
		name        string
		apiPath     string
		storagePath string
	}{
		{name: "nested document", apiPath: "guides/start", storagePath: filepath.Join("guides", "start")},
		{name: "homepage", apiPath: "pages/home", storagePath: ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			addRequest := httptest.NewRequest(http.MethodPost, "/api/comments/add/"+test.apiPath, strings.NewReader(`{"content":"Rendered **comment**"}`))
			addRequest.AddCookie(editorCookie)
			addResponse := httptest.NewRecorder()
			AddCommentHandler(addResponse, addRequest)
			if addResponse.Code != http.StatusOK {
				t.Fatalf("add status = %d, want 200; body: %s", addResponse.Code, addResponse.Body.String())
			}

			listRequest := httptest.NewRequest(http.MethodGet, "/api/comments/"+test.apiPath, nil)
			listResponse := httptest.NewRecorder()
			GetCommentsHandler(listResponse, listRequest)
			if listResponse.Code != http.StatusOK {
				t.Fatalf("list status = %d, want 200; body: %s", listResponse.Code, listResponse.Body.String())
			}
			var result struct {
				Success  bool               `json:"success"`
				Comments []comments.Comment `json:"comments"`
			}
			if err := json.Unmarshal(listResponse.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if !result.Success || len(result.Comments) != 1 {
				t.Fatalf("list response = %#v, want one comment", result)
			}
			if !strings.Contains(string(result.Comments[0].RenderedHTML), "<strong>comment</strong>") {
				t.Fatalf("rendered comment = %q, want strong markup", result.Comments[0].RenderedHTML)
			}
			if test.apiPath == "guides/start" {
				pageRequest := httptest.NewRequest(http.MethodGet, "/guides/start", nil)
				pageResponse := httptest.NewRecorder()
				PageHandler(pageResponse, pageRequest, testConfig)
				if pageResponse.Code != http.StatusOK {
					t.Fatalf("page status = %d, want 200; body: %s", pageResponse.Code, pageResponse.Body.String())
				}
				if !strings.Contains(pageResponse.Body.String(), "<strong>comment</strong>") {
					t.Fatal("page response omitted the rendered comment")
				}
			}
			storedFile := filepath.Join(testConfig.Wiki.RootDir, "comments", test.storagePath, result.Comments[0].ID)
			if _, err := os.Stat(storedFile); err != nil {
				t.Fatalf("comment was not stored at %q: %v", storedFile, err)
			}

			deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/comments/delete/"+test.apiPath+"/"+result.Comments[0].ID, nil)
			deleteRequest.AddCookie(adminCookie)
			deleteResponse := httptest.NewRecorder()
			DeleteCommentHandler(deleteResponse, deleteRequest)
			if deleteResponse.Code != http.StatusOK {
				t.Fatalf("delete status = %d, want 200; body: %s", deleteResponse.Code, deleteResponse.Body.String())
			}
			if _, err := os.Stat(storedFile); !os.IsNotExist(err) {
				t.Fatalf("deleted comment still exists at %q", storedFile)
			}
		})
	}
}

func TestCommentHandlersDenyRestrictedDocumentsBeforeFilesystemAccess(t *testing.T) {
	testConfig := installUserSessionTestConfig(t, nil)
	testConfig.Wiki.DocumentsDir = "documents"
	testConfig.AccessRules = []config.AccessRule{{
		Pattern: "/finance/**",
		Access:  "restricted",
		Groups:  []string{"finance"},
	}}
	editorCookie := sessionCookieForUser(t, testConfig, "editor", config.RoleEditor)

	addRequest := httptest.NewRequest(http.MethodPost, "/api/comments/add/finance/budget", strings.NewReader(`{"content":"blocked"}`))
	addRequest.AddCookie(editorCookie)
	addResponse := httptest.NewRecorder()
	AddCommentHandler(addResponse, addRequest)
	if addResponse.Code != http.StatusForbidden {
		t.Fatalf("restricted add status = %d, want 403; body: %s", addResponse.Code, addResponse.Body.String())
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/api/comments/finance/budget", nil)
	listResponse := httptest.NewRecorder()
	GetCommentsHandler(listResponse, listRequest)
	if listResponse.Code != http.StatusUnauthorized {
		t.Fatalf("restricted anonymous list status = %d, want 401; body: %s", listResponse.Code, listResponse.Body.String())
	}

	listRequest = httptest.NewRequest(http.MethodGet, "/api/comments/finance/budget", nil)
	listRequest.AddCookie(editorCookie)
	listResponse = httptest.NewRecorder()
	GetCommentsHandler(listResponse, listRequest)
	if listResponse.Code != http.StatusForbidden {
		t.Fatalf("restricted editor list status = %d, want 403; body: %s", listResponse.Code, listResponse.Body.String())
	}

	if _, err := os.Stat(filepath.Join(testConfig.Wiki.RootDir, "comments", "finance")); !os.IsNotExist(err) {
		t.Fatalf("restricted comment storage was accessed or created: %v", err)
	}
}

func TestCommentHandlersRejectInvalidDocumentPaths(t *testing.T) {
	testConfig := installUserSessionTestConfig(t, nil)
	testConfig.Wiki.DocumentsDir = "documents"
	editorCookie := sessionCookieForUser(t, testConfig, "editor", config.RoleEditor)

	tests := []struct {
		name    string
		method  string
		target  string
		body    string
		cookie  *http.Cookie
		handler http.HandlerFunc
	}{
		{name: "add", method: http.MethodPost, target: "/api/comments/add/%252e%252e/outside", body: `{"content":"blocked"}`, cookie: editorCookie, handler: AddCommentHandler},
		{name: "list", method: http.MethodGet, target: "/api/comments/%252e%252e/outside", handler: GetCommentsHandler},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.target, strings.NewReader(test.body))
			if test.cookie != nil {
				request.AddCookie(test.cookie)
			}
			response := httptest.NewRecorder()
			test.handler(response, request)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; body: %s", response.Code, response.Body.String())
			}
		})
	}
}

func writeCommentTestDocument(t *testing.T, filename string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filename, []byte("# Comment test document"), 0o644); err != nil {
		t.Fatal(err)
	}
}
