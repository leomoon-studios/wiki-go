package goldext

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWikiLinkPreprocessor(t *testing.T) {
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

// writePage creates data/documents/<rel>/document.md under dir.
func writePage(t *testing.T, dir, rel string) {
	t.Helper()
	full := filepath.Join(dir, documentsRoot, filepath.FromSlash(rel), "document.md")
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
	want := "[ghost](/ghost)" // root-level path -> LinkPreprocessor will red-link it
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
