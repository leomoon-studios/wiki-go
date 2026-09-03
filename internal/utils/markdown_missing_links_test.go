package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"wiki-go/internal/goldext"
)

func TestMissingInternalDocumentLinksRenderAsNotFound(t *testing.T) {
	root := t.TempDir()
	goldext.ConfigureContentPaths(root, "documents")
	t.Cleanup(func() { goldext.ConfigureContentPaths("data", "documents") })

	existing := filepath.Join(root, "documents", "existing")
	if err := os.MkdirAll(existing, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(existing, "document.md"), []byte("# Existing"), 0o644); err != nil {
		t.Fatal(err)
	}

	rendered := string(RenderMarkdownWithPath(
		"[Existing](/existing#section) [Missing](/missing#section) [[ghost]] [External](https://example.com) [Anchor](#local)",
		"current",
	))
	for _, expected := range []string{
		`<a href="/existing#section">Existing</a>`,
		`<span class="notfound"><a href="/missing#section">Missing</a></span>`,
		`<span class="notfound"><a href="/ghost">ghost</a></span>`,
		`<a href="https://example.com">External</a>`,
		`<a href="#local">Anchor</a>`,
	} {
		if !strings.Contains(rendered, expected) {
			t.Errorf("rendered Markdown omitted %q:\n%s", expected, rendered)
		}
	}
	if strings.Contains(rendered, `<span class="notfound"><a href="/existing`) ||
		strings.Contains(rendered, `<span class="notfound"><a href="https://example.com`) ||
		strings.Contains(rendered, `<span class="notfound"><a href="#local`) {
		t.Fatalf("valid or external link was marked missing: %s", rendered)
	}
}

func TestMissingLocalAttachmentLinksRenderAsNotFound(t *testing.T) {
	root := t.TempDir()
	goldext.ConfigureContentPaths(root, "documents")
	t.Cleanup(func() { goldext.ConfigureContentPaths("data", "documents") })

	documentDir := filepath.Join(root, "documents", "current")
	if err := os.MkdirAll(documentDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(documentDir, "manual.pdf"), []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	rendered := string(RenderMarkdownWithPath(
		"[Manual](manual.pdf) [Missing attachment](missing.pdf)",
		"current",
	))
	if strings.Contains(rendered, `<span class="notfound"><a href="/api/files/current/manual.pdf`) {
		t.Fatalf("existing attachment was marked missing: %s", rendered)
	}
	if !strings.Contains(rendered, `<span class="notfound"><a href="/api/files/current/missing.pdf">Missing attachment</a></span>`) {
		t.Fatalf("missing attachment was not marked missing: %s", rendered)
	}
}
