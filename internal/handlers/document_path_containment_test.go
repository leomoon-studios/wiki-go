package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"wiki-go/internal/config"
)

func TestDocumentHandlersRejectBackslashTraversalWithoutExternalAccess(t *testing.T) {
	testConfig := installUserSessionTestConfig(t, nil)
	testConfig.Wiki.DocumentsDir = "documents"
	testConfig.Wiki.MaxVersions = 10
	editorCookie := sessionCookieForUser(t, testConfig, "editor", config.RoleEditor)

	externalDir := filepath.Join(testConfig.Wiki.RootDir, "outside")
	externalDocument := filepath.Join(externalDir, "document.md")
	if err := os.MkdirAll(externalDir, 0o755); err != nil {
		t.Fatal(err)
	}
	const documentCanary = "EXTERNAL-DOCUMENT-CANARY"
	if err := os.WriteFile(externalDocument, []byte(documentCanary), 0o644); err != nil {
		t.Fatal(err)
	}

	versionsOutsideDir := filepath.Join(testConfig.Wiki.RootDir, "versions", "outside")
	if err := os.MkdirAll(versionsOutsideDir, 0o755); err != nil {
		t.Fatal(err)
	}
	const versionTimestamp = "20260920235959"
	const versionCanary = "EXTERNAL-VERSION-CANARY"
	if err := os.WriteFile(filepath.Join(versionsOutsideDir, versionTimestamp+".md"), []byte(versionCanary), 0o644); err != nil {
		t.Fatal(err)
	}

	safeDir := filepath.Join(testConfig.Wiki.RootDir, "documents", "safe")
	if err := os.MkdirAll(safeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(safeDir, "document.md"), []byte("# Safe"), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		request *http.Request
		handler http.HandlerFunc
	}{
		{
			name:    "source",
			request: httptest.NewRequest(http.MethodGet, "/api/source/..%5coutside", nil),
			handler: SourceHandler,
		},
		{
			name:    "save",
			request: httptest.NewRequest(http.MethodPost, "/api/save/..%5coutside", strings.NewReader("# Attacker overwrite")),
			handler: SaveHandler,
		},
		{
			name:    "move source",
			request: httptest.NewRequest(http.MethodPost, "/api/document/move", strings.NewReader(`{"sourcePath":"..\\outside","targetPath":"safe"}`)),
			handler: func(w http.ResponseWriter, r *http.Request) { MoveDocumentHandler(w, r, testConfig) },
		},
		{
			name:    "move target",
			request: httptest.NewRequest(http.MethodPost, "/api/document/move", strings.NewReader(`{"sourcePath":"safe","targetPath":"..\\outside"}`)),
			handler: func(w http.ResponseWriter, r *http.Request) { MoveDocumentHandler(w, r, testConfig) },
		},
		{
			name:    "version list",
			request: httptest.NewRequest(http.MethodGet, "/api/versions/..%5coutside", nil),
			handler: func(w http.ResponseWriter, r *http.Request) { VersionsHandler(w, r, testConfig) },
		},
		{
			name:    "version read",
			request: httptest.NewRequest(http.MethodGet, "/api/versions/..%5coutside/"+versionTimestamp, nil),
			handler: func(w http.ResponseWriter, r *http.Request) { VersionsHandler(w, r, testConfig) },
		},
		{
			name:    "version restore",
			request: httptest.NewRequest(http.MethodPost, "/api/versions/..%5coutside/"+versionTimestamp+"/restore", nil),
			handler: func(w http.ResponseWriter, r *http.Request) { VersionsHandler(w, r, testConfig) },
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.request.AddCookie(editorCookie)
			response := httptest.NewRecorder()
			test.handler(response, test.request)
			if response.Code < 400 || response.Code >= 500 {
				t.Fatalf("status = %d, want client error; body: %s", response.Code, response.Body.String())
			}
			if strings.Contains(response.Body.String(), documentCanary) || strings.Contains(response.Body.String(), versionCanary) {
				t.Fatal("response disclosed an external canary")
			}
		})
	}

	documentContent, err := os.ReadFile(externalDocument)
	if err != nil {
		t.Fatal(err)
	}
	if string(documentContent) != documentCanary {
		t.Fatalf("external document changed to %q", documentContent)
	}
	versionContent, err := os.ReadFile(filepath.Join(versionsOutsideDir, versionTimestamp+".md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(versionContent) != versionCanary {
		t.Fatalf("external version changed to %q", versionContent)
	}
	if _, err := os.Stat(filepath.Join(safeDir, "document.md")); err != nil {
		t.Fatalf("safe source document moved or removed: %v", err)
	}
}

func TestResolveVersionDocumentPathsPreservesDocumentsAndHomepageRoots(t *testing.T) {
	testConfig := &config.Config{}
	testConfig.Wiki.RootDir = t.TempDir()
	testConfig.Wiki.DocumentsDir = "documents"

	document, err := resolveVersionDocumentPaths(testConfig, "documents/finance/reports")
	if err != nil {
		t.Fatal(err)
	}
	if document.logicalPath != "/finance/reports" {
		t.Fatalf("logicalPath = %q, want /finance/reports", document.logicalPath)
	}
	if document.documentDir != filepath.Join(testConfig.Wiki.RootDir, "documents", "finance", "reports") {
		t.Fatalf("documentDir = %q", document.documentDir)
	}
	if document.versionsDir != filepath.Join(testConfig.Wiki.RootDir, "versions", "documents", "finance", "reports") {
		t.Fatalf("versionsDir = %q", document.versionsDir)
	}

	homepage, err := resolveVersionDocumentPaths(testConfig, "pages/home")
	if err != nil {
		t.Fatal(err)
	}
	if homepage.logicalPath != "/" {
		t.Fatalf("homepage logicalPath = %q, want /", homepage.logicalPath)
	}
	if homepage.documentDir != filepath.Join(testConfig.Wiki.RootDir, "pages", "home") {
		t.Fatalf("homepage documentDir = %q", homepage.documentDir)
	}
	if homepage.versionsDir != filepath.Join(testConfig.Wiki.RootDir, "versions", "pages", "home") {
		t.Fatalf("homepage versionsDir = %q", homepage.versionsDir)
	}
}

func TestEditorCanReadSaveMoveAndRestoreContainedDocument(t *testing.T) {
	testConfig := installUserSessionTestConfig(t, nil)
	testConfig.Wiki.DocumentsDir = "documents"
	testConfig.Wiki.MaxVersions = 10
	editorCookie := sessionCookieForUser(t, testConfig, "editor", config.RoleEditor)
	originalDir := filepath.Join(testConfig.Wiki.RootDir, "documents", "alpha")
	if err := os.MkdirAll(originalDir, 0o755); err != nil {
		t.Fatal(err)
	}
	const originalContent = "# Original document"
	if err := os.WriteFile(filepath.Join(originalDir, "document.md"), []byte(originalContent), 0o644); err != nil {
		t.Fatal(err)
	}

	sourceRequest := httptest.NewRequest(http.MethodGet, "/api/source/alpha", nil)
	sourceRequest.AddCookie(editorCookie)
	sourceResponse := httptest.NewRecorder()
	SourceHandler(sourceResponse, sourceRequest)
	if sourceResponse.Code != http.StatusOK || sourceResponse.Body.String() != originalContent {
		t.Fatalf("source status = %d, body = %q", sourceResponse.Code, sourceResponse.Body.String())
	}

	const updatedContent = "# Updated document"
	saveRequest := httptest.NewRequest(http.MethodPost, "/api/save/alpha", strings.NewReader(updatedContent))
	saveRequest.AddCookie(editorCookie)
	saveResponse := httptest.NewRecorder()
	SaveHandler(saveResponse, saveRequest)
	if saveResponse.Code != http.StatusOK {
		t.Fatalf("save status = %d, want 200; body: %s", saveResponse.Code, saveResponse.Body.String())
	}

	versionDir := filepath.Join(testConfig.Wiki.RootDir, "versions", "documents", "alpha")
	versions, err := os.ReadDir(versionDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 1 {
		t.Fatalf("saved versions = %d, want 1", len(versions))
	}
	timestamp := strings.TrimSuffix(versions[0].Name(), ".md")

	moveRequest := httptest.NewRequest(http.MethodPost, "/api/document/move", strings.NewReader(`{
		"sourcePath":"alpha",
		"targetPath":"archive",
		"newSlug":"renamed"
	}`))
	moveRequest.AddCookie(editorCookie)
	moveResponse := httptest.NewRecorder()
	MoveDocumentHandler(moveResponse, moveRequest, testConfig)
	if moveResponse.Code != http.StatusOK {
		t.Fatalf("move status = %d, want 200; body: %s", moveResponse.Code, moveResponse.Body.String())
	}

	movedDocument := filepath.Join(testConfig.Wiki.RootDir, "documents", "archive", "renamed", "document.md")
	movedContent, err := os.ReadFile(movedDocument)
	if err != nil {
		t.Fatal(err)
	}
	if string(movedContent) != updatedContent {
		t.Fatalf("moved content = %q, want %q", movedContent, updatedContent)
	}

	restoreRequest := httptest.NewRequest(http.MethodPost, "/api/versions/archive/renamed/"+timestamp+"/restore", nil)
	restoreRequest.AddCookie(editorCookie)
	restoreResponse := httptest.NewRecorder()
	VersionsHandler(restoreResponse, restoreRequest, testConfig)
	if restoreResponse.Code != http.StatusOK {
		t.Fatalf("restore status = %d, want 200; body: %s", restoreResponse.Code, restoreResponse.Body.String())
	}
	restoredContent, err := os.ReadFile(movedDocument)
	if err != nil {
		t.Fatal(err)
	}
	if string(restoredContent) != originalContent {
		t.Fatalf("restored content = %q, want %q", restoredContent, originalContent)
	}
}
