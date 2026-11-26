package auth

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"wiki-go/internal/config"
)

// These tests assert the secure behavior and intentionally fail on the unfixed build.
func TestWindowsFinanceAliasesDenied(t *testing.T) {
	cfg := financeAccessConfig(t)
	root := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir)
	cases := []struct{ name, canonical, alias string }{
		{"case", "finance/q3", "Finance/q3"},
		{"trailing dot", "finance/q3", "finance./q3"},
		{"attachment separator", "finance/q3", `finance\q3`},
		{"short name", "finance-quarterly-2026", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			alias := tc.alias
			if tc.name == "short name" {
				alias = windowsShortDirectory(t, filepath.Join(root, tc.canonical))
			}
			canonicalInfo, err := os.Stat(filepath.Join(root, filepath.FromSlash(tc.canonical)))
			if err != nil {
				t.Fatal(err)
			}
			aliasInfo, err := os.Stat(filepath.Join(root, filepath.FromSlash(alias)))
			if os.IsNotExist(err) {
				t.Skipf("filesystem does not resolve %q to the fixture", alias)
			}
			if err != nil {
				t.Fatal(err)
			}
			if !os.SameFile(canonicalInfo, aliasInfo) {
				t.Skipf("filesystem treats %q as a different directory", alias)
			}
			for _, private := range []bool{false, true} {
				cfg.Wiki.Private = private
				var session *Session
				if private {
					session = &Session{Role: config.RoleViewer}
				}
				if CanAccessDocument("/"+tc.canonical, session, cfg) {
					t.Fatal("canonical restricted path was allowed")
				}
				if CanAccessDocument("/"+alias, session, cfg) {
					t.Errorf("private=%v: alias %q bypassed restriction on %q", private, alias, tc.canonical)
				}
			}
		})
	}
}

func windowsShortDirectory(t *testing.T, directory string) string {
	t.Helper()
	input, err := syscall.UTF16PtrFromString(directory)
	if err != nil {
		t.Fatal(err)
	}
	n, err := syscall.GetShortPathName(input, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	if n == 0 {
		t.Fatal("GetShortPathName returned an empty buffer size")
	}
	buffer := make([]uint16, n)
	n, err = syscall.GetShortPathName(input, &buffer[0], uint32(len(buffer)))
	if err != nil {
		t.Fatal(err)
	}
	if n == 0 || n >= uint32(len(buffer)) {
		t.Fatal("unexpected GetShortPathName result")
	}
	short := filepath.Base(syscall.UTF16ToString(buffer[:n]))
	if strings.EqualFold(short, filepath.Base(directory)) {
		t.Skip("fixture directory has no 8.3 alias; short-name creation may be disabled")
	}
	t.Logf("discovered 8.3 alias: %s -> %s", filepath.Base(directory), short)
	return short
}
