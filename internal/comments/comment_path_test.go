package comments

import (
	"path/filepath"
	"testing"
)

func TestResolveCommentDirectoryStaysWithinCommentsRoot(t *testing.T) {
	commentsRoot := filepath.Join(t.TempDir(), "comments")
	tests := []struct {
		name        string
		document    string
		wantLogical string
		wantPath    string
	}{
		{
			name:        "nested document",
			document:    "/finance/reports",
			wantLogical: "/finance/reports",
			wantPath:    filepath.Join(commentsRoot, "finance", "reports"),
		},
		{
			name:        "backslash separators",
			document:    `finance\reports`,
			wantLogical: "/finance/reports",
			wantPath:    filepath.Join(commentsRoot, "finance", "reports"),
		},
		{
			name:        "homepage",
			document:    "/",
			wantLogical: "/",
			wantPath:    commentsRoot,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolved, err := resolveCommentDirectory(commentsRoot, test.document)
			if err != nil {
				t.Fatal(err)
			}
			if resolved.logicalPath != test.wantLogical {
				t.Errorf("logicalPath = %q, want %q", resolved.logicalPath, test.wantLogical)
			}
			if resolved.directoryPath != test.wantPath {
				t.Errorf("directoryPath = %q, want %q", resolved.directoryPath, test.wantPath)
			}
			if resolved.filePath != "" || resolved.commentID != "" {
				t.Fatalf("directory resolution unexpectedly returned a file: %+v", resolved)
			}
		})
	}
}

func TestResolveCommentDirectoryRejectsUnsafePaths(t *testing.T) {
	commentsRoot := filepath.Join(t.TempDir(), "comments")
	unsafePaths := []string{
		"",
		"../outside",
		"finance/../../outside",
		`finance\..\..\outside`,
		"finance/%2e%2e/outside",
		"finance/%5coutside",
		"//server/share",
		`C:\Windows\Temp`,
		"bad\x00path",
	}
	for _, unsafePath := range unsafePaths {
		t.Run(unsafePath, func(t *testing.T) {
			if resolved, err := resolveCommentDirectory(commentsRoot, unsafePath); err == nil {
				t.Fatalf("resolveCommentDirectory(%q) = %+v, want error", unsafePath, resolved)
			}
		})
	}
}

func TestResolveCommentFileRequiresContainedPathAndValidID(t *testing.T) {
	commentsRoot := filepath.Join(t.TempDir(), "comments")
	const validID = "20260921123456_editor-name.md"
	resolved, err := resolveCommentFile(commentsRoot, "/guides/start", validID)
	if err != nil {
		t.Fatal(err)
	}
	wantFile := filepath.Join(commentsRoot, "guides", "start", validID)
	if resolved.filePath != wantFile {
		t.Fatalf("filePath = %q, want %q", resolved.filePath, wantFile)
	}
	if resolved.commentID != validID {
		t.Fatalf("commentID = %q, want %q", resolved.commentID, validID)
	}

	unsafeDocuments := []string{"../tmp", `..\tmp`, "%2e%2e/tmp", "/x/../tmp"}
	for _, document := range unsafeDocuments {
		if result, err := resolveCommentFile(commentsRoot, document, validID); err == nil {
			t.Errorf("resolveCommentFile(%q, valid ID) = %+v, want error", document, result)
		}
	}
	invalidIDs := []string{
		"",
		"2026092112345_editor.md",
		"2026092112345x_editor.md",
		"20260921123456_.md",
		"20260921123456_..md",
		"20260921123456_editor.txt",
		"../20260921123456_editor.md",
		`..\20260921123456_editor.md`,
		"20260921123456_editor\nforged.md",
		"20260921123456_editor\rforged.md",
		"20260921123456_editor\u2028forged.md",
	}
	for _, commentID := range invalidIDs {
		if result, err := resolveCommentFile(commentsRoot, "/guides/start", commentID); err == nil {
			t.Errorf("resolveCommentFile(valid path, %q) = %+v, want error", commentID, result)
		}
	}
}

func TestIsValidCommentIDMatchesGeneratedFilenameShape(t *testing.T) {
	for _, commentID := range []string{
		"20260921123456_admin.md",
		"20260921123456_editor-name.md",
		"20260921123456_first_last.md",
		"20260921123456_user@example.com.md",
	} {
		if !isValidCommentID(commentID) {
			t.Errorf("isValidCommentID(%q) = false, want true", commentID)
		}
	}
	for _, commentID := range []string{
		"999_admin.md",
		"20260921123456_.md",
		"20260921123456_../admin.md",
		"20260921123456_admin.md/extra",
		"20260921123456_user name.md",
		"20260921123456_user%name.md",
		"20260921123456_admin\n.md",
	} {
		if isValidCommentID(commentID) {
			t.Errorf("isValidCommentID(%q) = true, want false", commentID)
		}
	}
}
