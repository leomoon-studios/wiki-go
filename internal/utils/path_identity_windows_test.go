package utils

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestWindowsDocumentIdentityMixedCase(t *testing.T) {
	root, documents := documentIdentityRoot(t)
	for _, resolver := range []func(string, string, string) (ResolvedDocumentPath, error){ResolveExistingDocumentPath, ResolveDocumentDestination} {
		for _, input := range []string{"/Finance/Q3", `/FINANCE\q3`, "/FiNaNcE/Q3/REPORT.TXT"} {
			got, err := resolver(root, strings.ToUpper(documents), input)
			if err != nil {
				t.Fatal(err)
			}
			want := "/finance/q3"
			if strings.HasSuffix(input, "REPORT.TXT") {
				want += "/report.txt"
			}
			assertDocumentIdentity(t, got, root, documents, want)
		}
	}
	got, err := ResolveDocumentDestination(root, documents, "/FINANCE/NewChild/SubDir")
	if err != nil {
		t.Fatal(err)
	}
	assertDocumentIdentity(t, got, root, documents, "/finance/NewChild/SubDir")
}

func TestWindowsDocumentIdentityShortName(t *testing.T) {
	root, documents := documentIdentityRoot(t)
	longName := "finance-quarterly-2026"
	directory := filepath.Join(root, documents, longName)
	if err := os.Mkdir(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	short := documentIdentityShortName(t, directory)
	longInfo, err := os.Stat(directory)
	if err != nil {
		t.Fatal(err)
	}
	aliasInfo, err := os.Stat(filepath.Join(root, documents, short))
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(longInfo, aliasInfo) {
		t.Fatal("discovered short name does not identify fixture directory")
	}
	for _, resolver := range []func(string, string, string) (ResolvedDocumentPath, error){ResolveExistingDocumentPath, ResolveDocumentDestination} {
		got, err := resolver(root, documents, "/"+short)
		if err != nil {
			t.Fatal(err)
		}
		assertDocumentIdentity(t, got, root, documents, "/"+longName)
	}
	got, err := ResolveDocumentDestination(root, documents, "/"+short+"/NewChild/SubDir")
	if err != nil {
		t.Fatal(err)
	}
	assertDocumentIdentity(t, got, root, documents, "/"+longName+"/NewChild/SubDir")
	t.Run("file alias", func(t *testing.T) {
		filename := "annual-attachment-2026.txt"
		if err := os.WriteFile(filepath.Join(directory, filename), []byte("attachment"), 0o644); err != nil {
			t.Fatal(err)
		}
		shortFile := documentIdentityShortName(t, filepath.Join(directory, filename))
		for _, resolver := range []func(string, string, string) (ResolvedDocumentPath, error){ResolveExistingDocumentPath, ResolveDocumentDestination} {
			got, err := resolver(root, documents, "/"+short+"/"+shortFile)
			if err != nil {
				t.Fatal(err)
			}
			assertDocumentIdentity(t, got, root, documents, "/"+longName+"/"+filename)
		}
	})
	t.Run("root alias", func(t *testing.T) {
		shortRoot := documentIdentityShortName(t, filepath.Join(root, documents))
		got, err := ResolveExistingDocumentPath(root, shortRoot, "/finance/q3")
		if err != nil {
			t.Fatal(err)
		}
		assertDocumentIdentity(t, got, root, documents, "/finance/q3")
	})
}

func documentIdentityShortName(t *testing.T, directory string) string {
	t.Helper()
	input, err := syscall.UTF16PtrFromString(directory)
	if err != nil {
		t.Fatal(err)
	}
	n, err := syscall.GetShortPathName(input, nil, 0)
	if err != nil || n == 0 {
		t.Fatalf("GetShortPathName size = %d, %v", n, err)
	}
	buffer := make([]uint16, n)
	n, err = syscall.GetShortPathName(input, &buffer[0], uint32(len(buffer)))
	if err != nil || n == 0 || n >= uint32(len(buffer)) {
		t.Fatalf("GetShortPathName = %d, %v", n, err)
	}
	short := filepath.Base(syscall.UTF16ToString(buffer[:n]))
	if strings.EqualFold(short, filepath.Base(directory)) {
		t.Skip("fixture directory has no 8.3 alias; short-name creation may be disabled")
	}
	t.Logf("discovered 8.3 alias: %s -> %s", filepath.Base(directory), short)
	return short
}

func TestWindowsDocumentIdentityRejectsAmbiguousNames(t *testing.T) {
	root, documents := documentIdentityRoot(t)
	for _, input := range []string{
		"/finance./q3", "/finance /q3", "/finance/q3.", "/finance/q3 ",
		"/finance./new/child", "/finance /new/child", "/finance/new./child", "/finance/new /child",
		"/finance/q3/report.txt:secret", "/finance/q3/report.txt::$DATA", "/finance:secret/new",
		"/NUL", "/CON", "/COM1", "/LPT1", "/CONIN$", "/CONOUT$", "/COM¹",
		"/finance/*", "/finance/?", "/finance/<name>", "/finance/na|me", "/finance/na\"me", "/finance/na\x01me",
		"/finance/invalid\xff", `//?/C:/Windows`, `//./C:/Windows`, `/??/C:/Windows`,
	} {
		t.Run(input, func(t *testing.T) {
			for _, resolver := range []func(string, string, string) (ResolvedDocumentPath, error){ResolveExistingDocumentPath, ResolveDocumentDestination} {
				if got, err := resolver(root, documents, input); err == nil || got != (ResolvedDocumentPath{}) {
					t.Fatalf("ambiguous Windows path = %+v, %v; want no result and error", got, err)
				}
				// An invalid path must be rejected before attempting to resolve
				// even the selected root, which deliberately does not exist here.
				if _, err := resolver(root, "missing-root", input); err == nil || errors.Is(err, os.ErrNotExist) {
					t.Fatalf("invalid path was not rejected before lookup: %v", err)
				}
			}
		})
	}
}

func TestWindowsDocumentIdentityJunctions(t *testing.T) {
	root, documents := documentIdentityRoot(t)
	documentsRoot := filepath.Join(root, documents)
	outside := filepath.Join(root, documents+"-outside")
	if err := os.Mkdir(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	// Junctions do not require the symlink privilege. Use isolated directories
	// without changing machine-wide filesystem settings.
	for _, link := range []struct{ name, target string }{
		{"alias", filepath.Join(documentsRoot, "finance")}, {"escape", outside},
	} {
		junction := filepath.Join(documentsRoot, link.name)
		output, err := exec.Command("cmd", "/c", "mklink", "/J", junction, link.target).CombinedOutput()
		if err != nil {
			t.Fatalf("create junction: %v: %s", err, output)
		}
		junctionInfo, err := os.Stat(junction)
		if err != nil {
			t.Fatal(err)
		}
		targetInfo, err := os.Stat(link.target)
		if err != nil || !os.SameFile(junctionInfo, targetInfo) {
			t.Fatalf("junction does not identify its target: %v", err)
		}
		// Remove the junction itself before TempDir cleanup, never its target.
		t.Cleanup(func() {
			if err := os.Remove(junction); err != nil {
				t.Error(err)
			}
		})
	}
	for _, resolver := range []func(string, string, string) (ResolvedDocumentPath, error){ResolveExistingDocumentPath, ResolveDocumentDestination} {
		got, err := resolver(root, documents, "/alias")
		if err != nil {
			t.Fatal(err)
		}
		assertDocumentIdentity(t, got, root, documents, "/finance")
		got, err = resolver(root, documents, "/alias/q3/report.txt")
		if err != nil {
			t.Fatal(err)
		}
		assertDocumentIdentity(t, got, root, documents, "/finance/q3/report.txt")
		for _, input := range []string{"/escape", "/escape/new/child"} {
			if got, err := resolver(root, documents, input); err == nil || got != (ResolvedDocumentPath{}) {
				t.Fatalf("junction escape = %+v, %v; want no result and error", got, err)
			}
		}
	}
	got, err := ResolveDocumentDestination(root, documents, "/alias/new/child")
	if err != nil {
		t.Fatal(err)
	}
	assertDocumentIdentity(t, got, root, documents, "/finance/new/child")
	got, err = ResolveExistingDocumentPath(root, filepath.Join(documents, "alias"), "/q3/report.txt")
	if err != nil {
		t.Fatal(err)
	}
	assertDocumentIdentity(t, got, root, filepath.Join(documents, "finance"), "/q3/report.txt")
}

func TestWindowsDocumentIdentityLongPath(t *testing.T) {
	root, documents := documentIdentityRoot(t)
	logical := "/finance/" + strings.Repeat("long-directory-name/", 16) + "report.txt"
	filename := filepath.Join(root, documents, filepath.FromSlash(logical))
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filename, []byte("long path fixture"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := ResolveExistingDocumentPath(root, documents, strings.ToUpper(logical))
	if err != nil {
		t.Fatal(err)
	}
	assertDocumentIdentity(t, got, root, documents, logical)
}
