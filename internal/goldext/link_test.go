package goldext

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLinkPreprocessor_UsesConfiguredAbsoluteRoot(t *testing.T) {
	rootDir := t.TempDir()
	ConfigureContentPaths(rootDir, "documents")
	t.Cleanup(func() { ConfigureContentPaths("data", "documents") })

	pagePath := filepath.Join(rootDir, "documents", "nucilandia", "document.md")
	if err := os.MkdirAll(filepath.Dir(pagePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pagePath, []byte("# Nucilandia"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := LinkPreprocessor("[Nucilandia](/nucilandia#history)", "")
	if strings.Contains(got, `class="notfound"`) {
		t.Errorf("existing absolute-root link marked notfound: %q", got)
	}
	if got != "[Nucilandia](/nucilandia#history)" {
		t.Errorf("got %q, want anchored link unchanged", got)
	}
}

func TestLinkPreprocessor_MissingLinkRemainsMarkdown(t *testing.T) {
	rootDir := t.TempDir()
	ConfigureContentPaths(rootDir, "documents")
	t.Cleanup(func() { ConfigureContentPaths("data", "documents") })

	got := LinkPreprocessor("[Missing](/missing#section)", "")
	if got != "[Missing](/missing#section)" {
		t.Errorf("missing absolute link changed unexpectedly: %q", got)
	}
	if strings.Contains(got, "<") {
		t.Errorf("link preprocessor emitted HTML: %q", got)
	}
}
