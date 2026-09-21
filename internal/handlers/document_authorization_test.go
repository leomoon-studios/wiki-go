package handlers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"wiki-go/internal/config"
)

func TestRestrictedDocumentEditorAPIsDenyUnauthorizedEditor(t *testing.T) {
	testConfig := installUserSessionTestConfig(t, nil)
	testConfig.Wiki.DocumentsDir = "documents"
	testConfig.Wiki.MaxUploadSize = 10
	testConfig.Wiki.DisableFileUploadChecking = true
	testConfig.AccessRules = []config.AccessRule{{
		Pattern: "/finance/**",
		Access:  "restricted",
		Groups:  []string{"finance"},
	}}
	editorCookie := sessionCookieForUser(t, testConfig, "editor", config.RoleEditor)

	financeDir := filepath.Join(testConfig.Wiki.RootDir, "documents", "finance")
	publicDir := filepath.Join(testConfig.Wiki.RootDir, "documents", "public")
	for directory, content := range map[string]string{
		financeDir: "# Finance secret",
		publicDir:  "# Public document",
	} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, "document.md"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(financeDir, "note.txt"), []byte("secret attachment"), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		request *http.Request
		handler http.HandlerFunc
	}{
		{name: "source", request: httptest.NewRequest(http.MethodGet, "/api/source/finance", nil), handler: SourceHandler},
		{name: "save", request: httptest.NewRequest(http.MethodPost, "/api/save/finance", strings.NewReader("# overwritten")), handler: SaveHandler},
		{name: "create", request: httptest.NewRequest(http.MethodPost, "/api/document/create", strings.NewReader(`{"title":"Budget","path":"finance/budget","type":"document"}`)), handler: CreateDocumentHandler},
		{name: "delete", request: httptest.NewRequest(http.MethodDelete, "/api/document/finance", nil), handler: DeleteDocumentHandler},
		{name: "move", request: httptest.NewRequest(http.MethodPost, "/api/document/move", strings.NewReader(`{"sourcePath":"finance","targetPath":"public"}`)), handler: func(w http.ResponseWriter, r *http.Request) { MoveDocumentHandler(w, r, testConfig) }},
		{name: "versions", request: httptest.NewRequest(http.MethodGet, "/api/versions/finance", nil), handler: func(w http.ResponseWriter, r *http.Request) { VersionsHandler(w, r, testConfig) }},
		{name: "delete attachment", request: httptest.NewRequest(http.MethodDelete, "/api/files/delete/finance/note.txt", nil), handler: func(w http.ResponseWriter, r *http.Request) { DeleteFileHandler(w, r, testConfig) }},
		{name: "rename attachment", request: httptest.NewRequest(http.MethodPost, "/api/files/rename", strings.NewReader(`{"currentPath":"finance/note.txt","newName":"renamed.txt"}`)), handler: func(w http.ResponseWriter, r *http.Request) { RenameFileHandler(w, r, testConfig) }},
		{name: "upload attachment", request: restrictedUploadRequest(t), handler: func(w http.ResponseWriter, r *http.Request) { UploadFileHandler(w, r, testConfig) }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.request.AddCookie(editorCookie)
			response := httptest.NewRecorder()
			test.handler(response, test.request)
			if response.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want 403; body: %s", response.Code, response.Body.String())
			}
		})
	}

	content, err := os.ReadFile(filepath.Join(financeDir, "document.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "# Finance secret" {
		t.Fatalf("restricted document changed to %q", content)
	}
	if _, err := os.Stat(filepath.Join(financeDir, "note.txt")); err != nil {
		t.Fatalf("restricted attachment changed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(financeDir, "budget")); !os.IsNotExist(err) {
		t.Fatalf("restricted document was created: %v", err)
	}
}

func TestListDocumentsFiltersRestrictedDocumentsForEditor(t *testing.T) {
	testConfig := installUserSessionTestConfig(t, nil)
	testConfig.Wiki.DocumentsDir = "documents"
	testConfig.AccessRules = []config.AccessRule{{Pattern: "/finance/**", Access: "restricted", Groups: []string{"finance"}}}
	editorCookie := sessionCookieForUser(t, testConfig, "editor", config.RoleEditor)

	for name := range map[string]struct{}{"finance": {}, "public": {}} {
		directory := filepath.Join(testConfig.Wiki.RootDir, "documents", name)
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, "document.md"), []byte("# "+name), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	request := httptest.NewRequest(http.MethodGet, "/api/documents/list", nil)
	request.AddCookie(editorCookie)
	response := httptest.NewRecorder()
	ListDocumentsHandler(response, request, testConfig)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", response.Code, response.Body.String())
	}
	var result DocumentsResponse
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Documents) != 1 || result.Documents[0].Path != "/public" {
		t.Fatalf("documents = %#v, want only /public", result.Documents)
	}
}

func restrictedUploadRequest(t *testing.T) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("docPath", "finance"); err != nil {
		t.Fatal(err)
	}
	part, err := writer.CreateFormFile("file", "upload.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("blocked upload")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/files/upload", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}
