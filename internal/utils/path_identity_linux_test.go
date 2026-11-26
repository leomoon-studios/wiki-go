package utils

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDocumentIdentityLinuxCaseSensitive(t *testing.T) {
	root, documents := documentIdentityRoot(t)
	lower := filepath.Join(root, documents, "finance")
	upper := filepath.Join(root, documents, "Finance")
	if err := os.MkdirAll(filepath.Join(upper, "q3"), 0o755); err != nil {
		t.Fatal(err)
	}
	lowerInfo, err := os.Stat(lower)
	if err != nil {
		t.Fatal(err)
	}
	upperInfo, err := os.Stat(upper)
	if err != nil {
		t.Fatal(err)
	}
	if os.SameFile(lowerInfo, upperInfo) {
		t.Skip("filesystem does not support distinct case-sensitive directories")
	}
	for _, logical := range []string{"/finance/q3", "/Finance/q3"} {
		for _, resolver := range []func(string, string, string) (ResolvedDocumentPath, error){ResolveExistingDocumentPath, ResolveDocumentDestination} {
			got, err := resolver(root, documents, logical)
			if err != nil {
				t.Fatal(err)
			}
			assertDocumentIdentity(t, got, root, documents, logical)
		}
	}
	if got, err := ResolveExistingDocumentPath(root, documents, "/FINANCE/q3"); !errors.Is(err, os.ErrNotExist) || got != (ResolvedDocumentPath{}) {
		t.Fatalf("different case = %+v, %v; want not-exist error", got, err)
	}
	got, err := ResolveDocumentDestination(root, documents, "/FINANCE/NewChild")
	if err != nil {
		t.Fatal(err)
	}
	assertDocumentIdentity(t, got, root, documents, "/FINANCE/NewChild")
	// Windows-only filename restrictions must not leak into Linux resolution.
	for _, name := range []string{"finance.", "finance ", "file:stream", "NUL", "literal%"} {
		if err := os.Mkdir(filepath.Join(root, documents, name), 0o755); err != nil {
			t.Fatal(err)
		}
		got, err := ResolveExistingDocumentPath(root, documents, "/"+name)
		if err != nil {
			t.Fatal(err)
		}
		assertDocumentIdentity(t, got, root, documents, "/"+name)
	}
}

func TestDocumentIdentityLinuxPermissionFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses filesystem permission checks")
	}
	root, documents := documentIdentityRoot(t)
	parent := filepath.Join(root, documents, "finance")
	if err := os.Chmod(parent, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(parent, 0o755); err != nil {
			t.Error(err)
		}
	})
	for _, resolver := range []func(string, string, string) (ResolvedDocumentPath, error){ResolveExistingDocumentPath, ResolveDocumentDestination} {
		if got, err := resolver(root, documents, "/finance/q3"); !errors.Is(err, os.ErrPermission) || got != (ResolvedDocumentPath{}) {
			t.Fatalf("inaccessible path = %+v, %v; want no result and permission error", got, err)
		}
	}
}
