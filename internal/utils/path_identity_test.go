package utils

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
)

func documentIdentityRoot(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	documents := "custom-documents"
	if err := os.MkdirAll(filepath.Join(root, documents, "finance", "q3"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, documents, "finance", "q3", "report.txt"), []byte("fixture"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root, documents
}

func assertDocumentIdentity(t *testing.T, got ResolvedDocumentPath, root, documents, logical string) {
	t.Helper()
	canonicalRoot, err := filepath.EvalSymlinks(filepath.Join(root, documents))
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(canonicalRoot, filepath.FromSlash(logical))
	if got.LogicalPath != logical || got.FilesystemPath != want {
		t.Fatalf("resolved = %+v, want logical %q and filesystem %q", got, logical, want)
	}
}

func TestDocumentIdentityExistingAndDestination(t *testing.T) {
	root, documents := documentIdentityRoot(t)
	for _, logical := range []string{"/", "/finance", "/finance/q3", "/finance/q3/report.txt"} {
		for _, resolver := range []struct {
			name string
			run  func(string, string, string) (ResolvedDocumentPath, error)
		}{{"existing", ResolveExistingDocumentPath}, {"destination", ResolveDocumentDestination}} {
			t.Run(resolver.name+logical, func(t *testing.T) {
				got, err := resolver.run(root, documents, logical)
				if err != nil {
					t.Fatal(err)
				}
				assertDocumentIdentity(t, got, root, documents, logical)
			})
		}
	}
	for _, input := range []string{`finance\q3`, "/finance//./q3/"} {
		got, err := ResolveExistingDocumentPath(root, documents, input)
		if err != nil {
			t.Fatal(err)
		}
		assertDocumentIdentity(t, got, root, documents, "/finance/q3")
	}
	for _, logical := range []string{"/finance/new", "/finance/new/sub/report.txt"} {
		t.Run("missing"+logical, func(t *testing.T) {
			if got, err := ResolveExistingDocumentPath(root, documents, logical); !errors.Is(err, os.ErrNotExist) || got != (ResolvedDocumentPath{}) {
				t.Fatalf("existing missing path = %+v, %v; want no result and not-exist error", got, err)
			}
			got, err := ResolveDocumentDestination(root, documents, logical)
			if err != nil {
				t.Fatal(err)
			}
			assertDocumentIdentity(t, got, root, documents, logical)
			if _, err := os.Stat(got.FilesystemPath); !os.IsNotExist(err) {
				t.Fatalf("destination resolution created a resource: %v", err)
			}
		})
	}
}

func TestDocumentIdentityRejectsUnsafePaths(t *testing.T) {
	root, documents := documentIdentityRoot(t)
	for _, input := range []string{
		"/../outside", "/finance/../outside", `/finance\..\outside`,
		"/finance/%2e%2e/outside", "/finance%2fq3", "/finance\x00/q3",
		`C:\Windows`, `C:relative`, `/C:/Windows`, `\\server\share`,
		`//server/share`, `\\?\C:\Windows`, `\\.\NUL`,
	} {
		t.Run(input, func(t *testing.T) {
			for _, resolver := range []func(string, string, string) (ResolvedDocumentPath, error){ResolveExistingDocumentPath, ResolveDocumentDestination} {
				if got, err := resolver(root, documents, input); err == nil || got != (ResolvedDocumentPath{}) {
					t.Fatalf("unsafe path = %+v, %v; want no result and error", got, err)
				}
			}
		})
	}
}

func TestDocumentIdentityResolutionFailures(t *testing.T) {
	root, documents := documentIdentityRoot(t)
	for _, resolver := range []func(string, string, string) (ResolvedDocumentPath, error){ResolveExistingDocumentPath, ResolveDocumentDestination} {
		for _, tc := range []struct{ documents, logical string }{
			{"missing-root", "/new/child"},
			{filepath.Join(documents, "finance", "q3", "report.txt"), "/"},
			{documents, "/finance/q3/report.txt/child"},
		} {
			if got, err := resolver(root, tc.documents, tc.logical); err == nil || got != (ResolvedDocumentPath{}) {
				t.Errorf("resolve(%q, %q) = %+v, %v; want no result and error", tc.documents, tc.logical, got, err)
			}
		}
	}
}

func TestDocumentIdentityLinks(t *testing.T) {
	root, documents := documentIdentityRoot(t)
	documentsRoot := filepath.Join(root, documents)
	outside := filepath.Join(root, documents+"-outside")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(documentsRoot, "finance"), filepath.Join(documentsRoot, "alias")); err != nil {
		if errors.Is(err, os.ErrPermission) || errors.Is(err, errors.ErrUnsupported) || runtime.GOOS == "windows" && errors.Is(err, syscall.Errno(1314)) { // ERROR_PRIVILEGE_NOT_HELD
			t.Skipf("symlink creation unavailable: %v", err)
		}
		t.Fatal(err)
	}
	for _, link := range []struct{ name, target string }{
		{"escape", outside}, {"dangling", filepath.Join(root, "missing")}, {"loop", filepath.Join(documentsRoot, "loop")},
	} {
		if err := os.Symlink(link.target, filepath.Join(documentsRoot, link.name)); err != nil {
			t.Fatal(err)
		}
	}
	for _, resolver := range []func(string, string, string) (ResolvedDocumentPath, error){ResolveExistingDocumentPath, ResolveDocumentDestination} {
		got, err := resolver(root, documents, "/alias/q3/report.txt")
		if err != nil {
			t.Fatal(err)
		}
		assertDocumentIdentity(t, got, root, documents, "/finance/q3/report.txt")
		for _, input := range []string{"/escape", "/escape/new/child", "/dangling", "/dangling/new/child", "/loop", "/loop/new"} {
			if got, err := resolver(root, documents, input); err == nil || got != (ResolvedDocumentPath{}) {
				t.Errorf("resolve(%q) = %+v, %v; want no result and error", input, got, err)
			}
		}
	}
	got, err := ResolveDocumentDestination(root, documents, "/alias/new/child")
	if err != nil {
		t.Fatal(err)
	}
	assertDocumentIdentity(t, got, root, documents, "/finance/new/child")
	// The selected root itself may be a deployment symlink. Containment is
	// measured against its resolved identity, rather than its original spelling.
	if err := os.Symlink(documentsRoot, filepath.Join(root, "root-alias")); err != nil {
		t.Fatal(err)
	}
	got, err = ResolveExistingDocumentPath(root, "root-alias", "/finance/q3")
	if err != nil {
		t.Fatal(err)
	}
	assertDocumentIdentity(t, got, root, documents, "/finance/q3")
}
