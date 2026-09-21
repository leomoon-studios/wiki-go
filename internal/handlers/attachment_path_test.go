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
