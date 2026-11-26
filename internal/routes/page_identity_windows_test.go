package routes

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestWindowsPageIdentityRoutesAliases(t *testing.T) {
	cfg, secret, public, _ := pageIdentityFixture(t)
	users := pageIdentityUsers(t, cfg)
	cases := []struct {
		name, target string
		valid        bool
	}{
		{"mixed case", "/Finance/Q3", true},
		{"backslash", `/FINANCE\q3`, true},
		{"encoded backslash", "/FINANCE%5cq3", true},
		{"trailing dot", "/finance./q3", false},
		{"trailing space", "/finance%20/q3", false},
		{"ADS", "/finance/q3:stream", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, private := range []bool{false, true} {
				cfg.Wiki.Private = private
				for _, user := range users {
					assertPageIdentityResponse(t, "/finance/q3", user.cookie, user.member, secret)
					assertPageIdentityResponse(t, tc.target, user.cookie, tc.valid && user.member, secret)
					assertPageIdentityResponse(t, "/PUBLIC", user.cookie, !private || user.cookie != nil, public)
				}
			}
		})
	}
	t.Run("8.3 alias", func(t *testing.T) {
		short := pageIdentityShortDirectory(t, filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, "finance-quarterly-2026"))
		for _, private := range []bool{false, true} {
			cfg.Wiki.Private = private
			for _, user := range users {
				assertPageIdentityResponse(t, "/finance-quarterly-2026", user.cookie, user.member, secret)
				assertPageIdentityResponse(t, "/"+short, user.cookie, user.member, secret)
			}
		}
	})
}

func pageIdentityShortDirectory(t *testing.T, directory string) string {
	t.Helper()
	input, err := syscall.UTF16PtrFromString(directory)
	if err != nil {
		t.Fatal(err)
	}
	n, err := syscall.GetShortPathName(input, nil, 0)
	if err != nil || n == 0 {
		t.Fatalf("GetShortPathName size=%d, err=%v", n, err)
	}
	buffer := make([]uint16, n)
	n, err = syscall.GetShortPathName(input, &buffer[0], uint32(len(buffer)))
	if err != nil || n == 0 || n >= uint32(len(buffer)) {
		t.Fatalf("GetShortPathName length=%d, err=%v", n, err)
	}
	short := filepath.Base(syscall.UTF16ToString(buffer[:n]))
	if strings.EqualFold(short, filepath.Base(directory)) {
		t.Skip("fixture has no 8.3 alias; short-name creation may be disabled")
	}
	canonicalInfo, err := os.Stat(directory)
	if err != nil {
		t.Fatal(err)
	}
	aliasInfo, err := os.Stat(filepath.Join(filepath.Dir(directory), short))
	if err != nil || !os.SameFile(canonicalInfo, aliasInfo) {
		t.Fatalf("short alias does not identify fixture: %v", err)
	}
	t.Logf("discovered 8.3 alias: %s -> %s", filepath.Base(directory), short)
	return short
}
