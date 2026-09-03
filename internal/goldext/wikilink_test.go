package goldext

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWikiLinkPreprocessor(t *testing.T) {
	// Run inside an empty temp dir so bare-name resolution always starts from a
	// known-empty slug index, regardless of where the test binary is invoked.
	dir := t.TempDir()
	t.Chdir(dir)
	slugCache.mu.Lock()
	slugCache.index = nil
	slugCache.modTime = time.Time{}
	slugCache.mu.Unlock()

	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "bare nested target uses last segment as label",
			in:   "see [[02-entities/mass-spec]] now",
			want: "see [mass-spec](/02-entities/mass-spec) now",
		},
		{
			name: "explicit label",
			in:   "[[02-entities/mass-spec|Mass Spectrometer 3]]",
			want: "[Mass Spectrometer 3](/02-entities/mass-spec)",
		},
		{
			name: "anchor on a page",
			in:   "[[notes#results]]",
			want: "[notes](/notes#results)",
		},
		{
			name: "in-page anchor only",
			in:   "[[#summary]]",
			want: "[summary](#summary)",
		},
		{
			name: "already-absolute target is preserved",
			in:   "[[/03-lab-tasks|Tasks]]",
			want: "[Tasks](/03-lab-tasks)",
		},
		{
			name: "embed form is left untouched",
			in:   "![[diagram.png]]",
			want: "![[diagram.png]]",
		},
		{
			name: "code span is not processed",
			in:   "use `[[literal]]` here",
			want: "use `[[literal]]` here",
		},
		{
			name: "fenced code block is not processed",
			in:   "```\n[[literal]]\n```",
			want: "```\n[[literal]]\n```",
		},
		{
			name: "two links on one line",
			in:   "[[a]] and [[b|Bee]]",
			want: "[a](/a) and [Bee](/b)",
		},
		{
			name: "whitespace around target and label is trimmed",
			in:   "[[ a/b | Label ]]",
			want: "[Label](/a/b)",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := WikiLinkPreprocessor(tc.in, "")
			if got != tc.want {
				t.Errorf("WikiLinkPreprocessor(%q)\n  got:  %q\n  want: %q", tc.in, got, tc.want)
			}
		})
	}
}

// writePage creates documents/<rel>/document.md under dir.
func writePage(t *testing.T, dir, rel string) {
	t.Helper()
	full := filepath.Join(dir, DocumentsRoot(), filepath.FromSlash(rel), "document.md")
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte("# "+rel), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestWikiLinkPreprocessor_ResolvesByName(t *testing.T) {
	dir := t.TempDir()
	writePage(t, dir, "02-entities/mass-spec")
	t.Chdir(dir)

	got := WikiLinkPreprocessor("see [[mass-spec]] now", "")
	want := "see [mass-spec](/02-entities/mass-spec) now"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWikiLinkPreprocessor_BareNameNotFoundFallsBackToRoot(t *testing.T) {
	dir := t.TempDir() // empty tree: no matching page
	t.Chdir(dir)

	got := WikiLinkPreprocessor("[[ghost]]", "")
	want := "[ghost](/ghost)" // the AST transformer marks this target during rendering
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWikiLinkPreprocessor_DuplicateNameShortestPathWins(t *testing.T) {
	dir := t.TempDir()
	writePage(t, dir, "a/b/mass-spec") // deeper
	writePage(t, dir, "02-entities/mass-spec")
	t.Chdir(dir)

	got := WikiLinkPreprocessor("[[mass-spec]]", "")
	want := "[mass-spec](/02-entities/mass-spec)" // fewer segments wins
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWikiLinkPreprocessorNoWikiLinks(t *testing.T) {
	input := "This is a normal paragraph with **bold** and [a link](https://example.com)."
	got := WikiLinkPreprocessor(input, "")
	if got != input {
		t.Errorf("expected input returned unchanged, got %q", got)
	}
}

func TestSlugCacheRevalidatesOnMtimeChange(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	// Reset the package-level cache so prior test runs don't interfere.
	slugCache.mu.Lock()
	slugCache.index = nil
	slugCache.modTime = time.Time{}
	slugCache.mu.Unlock()

	// First render: empty tree, bare name should fall back to root.
	got := WikiLinkPreprocessor("[[new-page]]", "")
	if got != "[new-page](/new-page)" {
		t.Fatalf("before add: got %q, want %q", got, "[new-page](/new-page)")
	}

	// Add the page so it now exists in the tree.
	writePage(t, dir, "section/new-page")

	// Bump the documents-root mtime so the cache detects a change.
	newMtime := time.Now().Add(2 * time.Second)
	docsDir := filepath.Join(dir, DocumentsRoot())
	if err := os.Chtimes(docsDir, newMtime, newMtime); err != nil {
		t.Fatal(err)
	}

	// Second render: cache should rebuild and resolve the new page.
	got = WikiLinkPreprocessor("[[new-page]]", "")
	want := "[new-page](/section/new-page)"
	if got != want {
		t.Errorf("after add: got %q, want %q", got, want)
	}
}

func TestWikiLinkPreprocessor_UsesConfiguredAbsoluteRoot(t *testing.T) {
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

	got := WikiLinkPreprocessor("[[nucilandia]]", "")
	if got != "[nucilandia](/nucilandia)" {
		t.Errorf("got %q, want %q", got, "[nucilandia](/nucilandia)")
	}
}
