package utils

import (
	"path/filepath"
	"testing"
)

func TestCanonicalRequestPath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "root", input: "/", want: "/"},
		{name: "nested document", input: "/guides/getting-started/", want: "/guides/getting-started"},
		{name: "repeated separators", input: "/guides//install", want: "/guides/install"},
		{name: "backslash separator", input: `/guides\install`, want: "/guides/install"},
		{name: "forward traversal", input: "/../../outside", wantErr: true},
		{name: "backslash traversal", input: `/..\..\outside`, wantErr: true},
		{name: "in-wiki traversal", input: "/x/../finance", wantErr: true},
		{name: "double encoded traversal", input: "/x/%2e%2e/finance", wantErr: true},
		{name: "double encoded separator", input: "/x%2ffinance", wantErr: true},
		{name: "literal percent", input: "/coverage-100%", want: "/coverage-100%"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := CanonicalRequestPath(test.input)
			if test.wantErr {
				if err == nil {
					t.Fatalf("CanonicalRequestPath(%q) = %q, want error", test.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("CanonicalRequestPath(%q): %v", test.input, err)
			}
			if got != test.want {
				t.Fatalf("CanonicalRequestPath(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestResolveDocumentPathStaysWithinDocumentsRoot(t *testing.T) {
	root := t.TempDir()
	documentsRoot := filepath.Join(root, "documents")

	got, err := ResolveDocumentPath(root, "documents", "/finance/reports")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(documentsRoot, "finance", "reports")
	if got != want {
		t.Fatalf("ResolveDocumentPath() = %q, want %q", got, want)
	}

	for _, unsafePath := range []string{"/../outside", `/..\outside`, "/%2e%2e/outside"} {
		if got, err := ResolveDocumentPath(root, "documents", unsafePath); err == nil {
			t.Errorf("ResolveDocumentPath(%q) = %q, want error", unsafePath, got)
		}
	}
}

func TestResolveRelativeDocumentPath(t *testing.T) {
	root := t.TempDir()
	documentsRoot := filepath.Join(root, "documents")
	tests := []struct {
		name        string
		input       string
		wantLogical string
		wantPath    string
	}{
		{
			name:        "nested document",
			input:       "guides/getting-started",
			wantLogical: "/guides/getting-started",
			wantPath:    filepath.Join(documentsRoot, "guides", "getting-started"),
		},
		{
			name:        "platform separator input",
			input:       `guides\windows-install`,
			wantLogical: "/guides/windows-install",
			wantPath:    filepath.Join(documentsRoot, "guides", "windows-install"),
		},
		{
			name:        "repeated separator input",
			input:       "guides//install",
			wantLogical: "/guides/install",
			wantPath:    filepath.Join(documentsRoot, "guides", "install"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolved, err := ResolveRelativeDocumentPath(root, "documents", test.input)
			if err != nil {
				t.Fatal(err)
			}
			if resolved.LogicalPath != test.wantLogical {
				t.Errorf("LogicalPath = %q, want %q", resolved.LogicalPath, test.wantLogical)
			}
			if resolved.FilesystemPath != test.wantPath {
				t.Errorf("FilesystemPath = %q, want %q", resolved.FilesystemPath, test.wantPath)
			}
		})
	}
}

func TestResolveRelativeDocumentPathRejectsUnsafeInput(t *testing.T) {
	root := t.TempDir()
	unsafePaths := []string{
		"",
		".",
		"../outside",
		"guides/../../outside",
		`guides\..\..\outside`,
		"guides/%5c../../outside",
		"guides/%2e%2e/outside",
		"/etc/passwd",
		`\etc\passwd`,
		`C:\Windows\system32`,
	}

	for _, unsafePath := range unsafePaths {
		t.Run(unsafePath, func(t *testing.T) {
			if resolved, err := ResolveRelativeDocumentPath(root, "documents", unsafePath); err == nil {
				t.Fatalf("ResolveRelativeDocumentPath(%q) = %+v, want error", unsafePath, resolved)
			}
		})
	}
}
