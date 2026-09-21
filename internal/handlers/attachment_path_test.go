package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"wiki-go/internal/config"
)

func TestResolveAttachmentPath(t *testing.T) {
	root := t.TempDir()
	testConfig := &config.Config{}
	testConfig.Wiki.RootDir = root
	testConfig.Wiki.DocumentsDir = "documents"

	tests := []struct {
		name        string
		input       string
		wantPath    string
		wantRoot    string
		wantStorage string
	}{
		{
			name:        "document attachment",
			input:       "finance/reports/quarterly.report_v2-final.pdf",
			wantPath:    filepath.Join(root, "documents", "finance", "reports", "quarterly.report_v2-final.pdf"),
			wantRoot:    filepath.Join(root, "documents"),
			wantStorage: "finance/reports/quarterly.report_v2-final.pdf",
		},
		{
			name:        "homepage attachment",
			input:       "pages/home/logo.dark_v2-final.png",
			wantPath:    filepath.Join(root, "pages", "home", "logo.dark_v2-final.png"),
			wantRoot:    filepath.Join(root, "pages", "home"),
			wantStorage: "pages/home/logo.dark_v2-final.png",
		},
		{
			name:        "backslash separators",
			input:       `finance\reports\summary.pdf`,
			wantPath:    filepath.Join(root, "documents", "finance", "reports", "summary.pdf"),
			wantRoot:    filepath.Join(root, "documents"),
			wantStorage: "finance/reports/summary.pdf",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolved, err := resolveAttachmentPath(testConfig, test.input)
			if err != nil {
				t.Fatal(err)
			}
			if resolved.filesystemPath != test.wantPath {
				t.Errorf("filesystemPath = %q, want %q", resolved.filesystemPath, test.wantPath)
			}
			if resolved.rootPath != test.wantRoot {
				t.Errorf("rootPath = %q, want %q", resolved.rootPath, test.wantRoot)
			}
			if resolved.storagePath != test.wantStorage {
				t.Errorf("storagePath = %q, want %q", resolved.storagePath, test.wantStorage)
			}
		})
	}
}

func TestResolveAttachmentPathRejectsEscapingPaths(t *testing.T) {
	testConfig := &config.Config{}
	testConfig.Wiki.RootDir = t.TempDir()
	testConfig.Wiki.DocumentsDir = "documents"

	unsafePaths := []string{
		"",
		"../config.yaml",
		"finance/../../config.yaml",
		`finance\..\..\config.yaml`,
		"finance/%2e%2e/config.yaml",
		"finance/%5c../config.yaml",
		"/etc/passwd",
		`C:\Windows\system.ini`,
		"pages/home/../config.yaml",
		"pages/other/attachment.txt",
	}
	for _, unsafePath := range unsafePaths {
		t.Run(unsafePath, func(t *testing.T) {
			if resolved, err := resolveAttachmentPath(testConfig, unsafePath); err == nil {
				t.Fatalf("resolveAttachmentPath(%q) = %+v, want error", unsafePath, resolved)
			}
		})
	}
}

func TestResolveAttachmentRenameDestinationStaysInSourceDirectory(t *testing.T) {
	testConfig := &config.Config{}
	testConfig.Wiki.RootDir = t.TempDir()
	testConfig.Wiki.DocumentsDir = "documents"
	source, err := resolveAttachmentPath(testConfig, "finance/reports/original.pdf")
	if err != nil {
		t.Fatal(err)
	}

	destination, err := resolveAttachmentRenameDestination(source, "quarterly_report-final.pdf")
	if err != nil {
		t.Fatal(err)
	}
	wantPath := filepath.Join(testConfig.Wiki.RootDir, "documents", "finance", "reports", "quarterly_report-final.pdf")
	if destination.filesystemPath != wantPath {
		t.Fatalf("filesystemPath = %q, want %q", destination.filesystemPath, wantPath)
	}
	if destination.rootPath != source.rootPath {
		t.Fatalf("rootPath = %q, want source root %q", destination.rootPath, source.rootPath)
	}
	if destination.storagePath != "finance/reports/quarterly_report-final.pdf" {
		t.Fatalf("storagePath = %q, want same document directory", destination.storagePath)
	}

	for _, unsafeName := range []string{"../config.yaml", `..\config.yaml`, "nested/report.pdf"} {
		if resolved, err := resolveAttachmentRenameDestination(source, unsafeName); err == nil {
			t.Errorf("resolveAttachmentRenameDestination(%q) = %+v, want error", unsafeName, resolved)
		}
	}
}

func TestValidateAttachmentFilename(t *testing.T) {
	for _, filename := range []string{
		"report.pdf",
		"quarterly.report_v2-final.pdf",
		"archive_2026-09-20.tar.gz",
	} {
		t.Run("valid "+filename, func(t *testing.T) {
			if err := validateAttachmentFilename(filename); err != nil {
				t.Fatalf("validateAttachmentFilename(%q): %v", filename, err)
			}
		})
	}

	for _, filename := range []string{
		"",
		".",
		"..",
		"../config.yaml",
		`..\config.yaml`,
		"nested/report.pdf",
		`nested\report.pdf`,
		"report..pdf",
		"report name.pdf",
	} {
		t.Run("invalid "+filename, func(t *testing.T) {
			if err := validateAttachmentFilename(filename); err == nil {
				t.Fatalf("validateAttachmentFilename(%q) succeeded, want error", filename)
			}
		})
	}
}

func TestRenameFileHandlerRejectsNestedNewName(t *testing.T) {
	testConfig := installUserSessionTestConfig(t, nil)
	testConfig.Wiki.DocumentsDir = "documents"
	editorCookie := sessionCookieForUser(t, testConfig, "editor", config.RoleEditor)

	request := httptest.NewRequest(http.MethodPost, "/api/files/rename", strings.NewReader(`{
		"currentPath":"finance/report.pdf",
		"newName":"../../config.yaml"
	}`))
	request.AddCookie(editorCookie)
	response := httptest.NewRecorder()
	RenameFileHandler(response, request, testConfig)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body: %s", response.Code, response.Body.String())
	}
}

func TestRenameFileHandlerAcceptsValidFilename(t *testing.T) {
	for _, newName := range []string{
		"quarterly.report.pdf",
		"quarterly_report.pdf",
		"quarterly-report.pdf",
	} {
		t.Run(newName, func(t *testing.T) {
			testConfig := installUserSessionTestConfig(t, nil)
			testConfig.Wiki.DocumentsDir = "documents"
			editorCookie := sessionCookieForUser(t, testConfig, "editor", config.RoleEditor)
			financeDir := filepath.Join(testConfig.Wiki.RootDir, "documents", "finance")
			if err := os.MkdirAll(financeDir, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(financeDir, "original.pdf"), []byte("attachment"), 0o644); err != nil {
				t.Fatal(err)
			}

			request := httptest.NewRequest(http.MethodPost, "/api/files/rename", strings.NewReader(`{
				"currentPath":"finance/original.pdf",
				"newName":"`+newName+`"
			}`))
			request.AddCookie(editorCookie)
			response := httptest.NewRecorder()
			RenameFileHandler(response, request, testConfig)

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body: %s", response.Code, response.Body.String())
			}
			if _, err := os.Stat(filepath.Join(financeDir, newName)); err != nil {
				t.Fatalf("renamed attachment missing: %v", err)
			}
			if _, err := os.Stat(filepath.Join(financeDir, "original.pdf")); !os.IsNotExist(err) {
				t.Fatalf("original attachment still exists: %v", err)
			}
		})
	}
}

func TestRenameFileHandlerRenamesHomepageAttachmentWithinHomepageRoot(t *testing.T) {
	testConfig := installUserSessionTestConfig(t, nil)
	testConfig.Wiki.DocumentsDir = "documents"
	editorCookie := sessionCookieForUser(t, testConfig, "editor", config.RoleEditor)
	homeDir := filepath.Join(testConfig.Wiki.RootDir, "pages", "home")
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(homeDir, "original.png"), []byte("homepage attachment"), 0o644); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/files/rename", strings.NewReader(`{
		"currentPath":"pages/home/original.png",
		"newName":"renamed.png"
	}`))
	request.AddCookie(editorCookie)
	response := httptest.NewRecorder()
	RenameFileHandler(response, request, testConfig)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", response.Code, response.Body.String())
	}
	if _, err := os.Stat(filepath.Join(homeDir, "renamed.png")); err != nil {
		t.Fatalf("renamed homepage attachment missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(homeDir, "original.png")); !os.IsNotExist(err) {
		t.Fatalf("original homepage attachment still exists: %v", err)
	}
}

func TestRenameFileHandlerAuthorizesBeforeCheckingSourceExistence(t *testing.T) {
	testConfig := installUserSessionTestConfig(t, nil)
	testConfig.Wiki.DocumentsDir = "documents"
	testConfig.AccessRules = []config.AccessRule{{
		Pattern: "/finance/**",
		Access:  "restricted",
		Groups:  []string{"finance"},
	}}
	editorCookie := sessionCookieForUser(t, testConfig, "editor", config.RoleEditor)
	financeDir := filepath.Join(testConfig.Wiki.RootDir, "documents", "finance")
	if err := os.MkdirAll(financeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(financeDir, "existing.pdf"), []byte("restricted"), 0o644); err != nil {
		t.Fatal(err)
	}

	for _, sourceName := range []string{"existing.pdf", "missing.pdf"} {
		t.Run(sourceName, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/files/rename", strings.NewReader(`{
				"currentPath":"finance/`+sourceName+`",
				"newName":"renamed.pdf"
			}`))
			request.AddCookie(editorCookie)
			response := httptest.NewRecorder()
			RenameFileHandler(response, request, testConfig)

			if response.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want 403; body: %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestRenameFileHandlerRejectsEscapingSourcePath(t *testing.T) {
	testConfig := installUserSessionTestConfig(t, nil)
	testConfig.Wiki.DocumentsDir = "documents"
	editorCookie := sessionCookieForUser(t, testConfig, "editor", config.RoleEditor)

	request := httptest.NewRequest(http.MethodPost, "/api/files/rename", strings.NewReader(`{
		"currentPath":"../config.yaml",
		"newName":"renamed.yaml"
	}`))
	request.AddCookie(editorCookie)
	response := httptest.NewRecorder()
	RenameFileHandler(response, request, testConfig)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body: %s", response.Code, response.Body.String())
	}
}

func TestRenameFileHandlerReturnsConflictForContainedDestination(t *testing.T) {
	testConfig := installUserSessionTestConfig(t, nil)
	testConfig.Wiki.DocumentsDir = "documents"
	editorCookie := sessionCookieForUser(t, testConfig, "editor", config.RoleEditor)
	documentDir := filepath.Join(testConfig.Wiki.RootDir, "documents", "public")
	if err := os.MkdirAll(documentDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{
		"source.pdf":      "source",
		"destination.pdf": "destination",
	} {
		if err := os.WriteFile(filepath.Join(documentDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	request := httptest.NewRequest(http.MethodPost, "/api/files/rename", strings.NewReader(`{
		"currentPath":"public/source.pdf",
		"newName":"destination.pdf"
	}`))
	request.AddCookie(editorCookie)
	response := httptest.NewRecorder()
	RenameFileHandler(response, request, testConfig)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body: %s", response.Code, response.Body.String())
	}
}

func TestRenameFileHandlerRejectsReporterTraversalWithoutChangingFiles(t *testing.T) {
	testConfig := installUserSessionTestConfig(t, nil)
	testConfig.Wiki.DocumentsDir = "documents"
	editorCookie := sessionCookieForUser(t, testConfig, "editor", config.RoleEditor)
	configPath := filepath.Join(testConfig.Wiki.RootDir, "config.yaml")
	const configContents = "users:\n  - username: admin\n"
	if err := os.WriteFile(configPath, []byte(configContents), 0o600); err != nil {
		t.Fatal(err)
	}
	documentDir := filepath.Join(testConfig.Wiki.RootDir, "documents", "foo")
	if err := os.MkdirAll(documentDir, 0o755); err != nil {
		t.Fatal(err)
	}
	payloadPath := filepath.Join(documentDir, "payload.txt")
	const payloadContents = "controlled attachment"
	if err := os.WriteFile(payloadPath, []byte(payloadContents), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		body string
	}{
		{
			name: "source escape targeting config",
			body: `{"currentPath":"../config.yaml","newName":"config-stolen.txt"}`,
		},
		{
			name: "destination escape from attachment",
			body: `{"currentPath":"foo/payload.txt","newName":"../../pwned.txt"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/files/rename", strings.NewReader(test.body))
			request.AddCookie(editorCookie)
			response := httptest.NewRecorder()
			RenameFileHandler(response, request, testConfig)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; body: %s", response.Code, response.Body.String())
			}
		})
	}

	for path, want := range map[string]string{
		configPath:  configContents,
		payloadPath: payloadContents,
	} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("protected source %s is missing: %v", path, err)
		}
		if string(content) != want {
			t.Fatalf("protected source %s changed to %q", path, content)
		}
	}
	for _, unexpectedPath := range []string{
		filepath.Join(testConfig.Wiki.RootDir, "config-stolen.txt"),
		filepath.Join(testConfig.Wiki.RootDir, "pwned.txt"),
		filepath.Join(testConfig.Wiki.RootDir, "documents", "config-stolen.txt"),
	} {
		if _, err := os.Stat(unexpectedPath); !os.IsNotExist(err) {
			t.Fatalf("escaping destination exists at %s: %v", unexpectedPath, err)
		}
	}
}

func TestUploadAndDeleteUseAttachmentStorageBoundary(t *testing.T) {
	testConfig := installUserSessionTestConfig(t, nil)
	testConfig.Wiki.DocumentsDir = "documents"
	testConfig.Wiki.MaxUploadSize = 10
	testConfig.Wiki.DisableFileUploadChecking = true
	editorCookie := sessionCookieForUser(t, testConfig, "editor", config.RoleEditor)
	outsideHomepageRoot := filepath.Join(testConfig.Wiki.RootDir, "pages", "other")
	if err := os.MkdirAll(outsideHomepageRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	protectedPath := filepath.Join(outsideHomepageRoot, "protected.txt")
	const protectedContents = "must remain unchanged"
	if err := os.WriteFile(protectedPath, []byte(protectedContents), 0o644); err != nil {
		t.Fatal(err)
	}

	uploadRequest := attachmentUploadRequest(t, "pages/other", "uploaded.txt", "blocked upload")
	uploadRequest.AddCookie(editorCookie)
	uploadResponse := httptest.NewRecorder()
	UploadFileHandler(uploadResponse, uploadRequest, testConfig)
	if uploadResponse.Code != http.StatusBadRequest {
		t.Fatalf("upload status = %d, want 400; body: %s", uploadResponse.Code, uploadResponse.Body.String())
	}
	if _, err := os.Stat(filepath.Join(outsideHomepageRoot, "uploaded.txt")); !os.IsNotExist(err) {
		t.Fatalf("upload escaped the attachment roots: %v", err)
	}

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/files/delete/pages/other/protected.txt", nil)
	deleteRequest.AddCookie(editorCookie)
	deleteResponse := httptest.NewRecorder()
	DeleteFileHandler(deleteResponse, deleteRequest, testConfig)
	if deleteResponse.Code != http.StatusBadRequest {
		t.Fatalf("delete status = %d, want 400; body: %s", deleteResponse.Code, deleteResponse.Body.String())
	}
	content, err := os.ReadFile(protectedPath)
	if err != nil {
		t.Fatalf("protected file was removed: %v", err)
	}
	if string(content) != protectedContents {
		t.Fatalf("protected file changed to %q", content)
	}
}

func attachmentUploadRequest(t *testing.T, documentPath, filename, contents string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("docPath", documentPath); err != nil {
		t.Fatal(err)
	}
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte(contents)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/files/upload", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}
