package handlers

import (
	"os"
	"path/filepath"
	"testing"

	"wiki-go/internal/config"
)

func TestSearchDocumentPath(t *testing.T) {
	root := filepath.Join(t.TempDir(), "custom-documents")
	tests := []struct {
		name, file, want string
	}{
		{"nested", filepath.Join(root, "hr", "payroll", "document.md"), "/hr/payroll"},
		{"document name in directory", filepath.Join(root, "document.md-notes", "document.md"), "/document.md-notes"},
		{"attachment", filepath.Join(root, "hr", "notes.md"), ""},
		{"root document", filepath.Join(root, "document.md"), ""},
		{"outside root", filepath.Join(filepath.Dir(root), "outside", "document.md"), ""},
		{"encoded segment", filepath.Join(root, "%2e%2e", "document.md"), ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := searchDocumentPath(root, test.file)
			if test.want == "" {
				if err == nil {
					t.Fatalf("searchDocumentPath() = %q, want error", got)
				}
				return
			}
			if err != nil || got != test.want {
				t.Fatalf("searchDocumentPath() = %q, %v; want %q", got, err, test.want)
			}
		})
	}
}

func TestSearchIndexesWikiDocumentsOnly(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "custom-documents", "document.md-notes")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	for name, contents := range map[string]string{"document.md": "# Page\nneedle", "attachment.md": "# Attachment\nneedle"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0644); err != nil {
			t.Fatal(err)
		}
	}
	cfg := &config.Config{}
	cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir = root, "custom-documents"
	results := performSearch("needle", nil, cfg)
	if len(results) != 1 || results[0].Path != "/document.md-notes" || results[0].Title != "Page" {
		t.Fatalf("search results = %+v, want only canonical wiki document", results)
	}
}
