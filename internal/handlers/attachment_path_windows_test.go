package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"wiki-go/internal/config"
)

func TestWindowsFinancePageAliasesDenied(t *testing.T) {
	testWindowsFinanceDisclosureAliases(t, false)
}

func TestWindowsFinanceAttachmentAliasesDenied(t *testing.T) {
	testWindowsFinanceDisclosureAliases(t, true)
}

func testWindowsFinanceDisclosureAliases(t *testing.T, attachment bool) {
	t.Helper()
	testConfig, documentMarker, attachmentMarker := financeDisclosureFixture(t)
	outsider := sessionCookieForUser(t, testConfig, "outsider", config.RoleViewer)
	root := filepath.Join(testConfig.Wiki.RootDir, testConfig.Wiki.DocumentsDir)
	for _, aliasCase := range []struct{ name, canonical, alias string }{
		{"case", "finance/q3", "Finance/q3"},
		{"trailing dot", "finance/q3", "finance./q3"},
		{"short name", "finance-quarterly-2026", ""},
		{"canonical nested", "finance/q3", "finance/q3"},
	} {
		if !attachment && aliasCase.name == "canonical nested" {
			continue
		}
		t.Run(aliasCase.name, func(t *testing.T) {
			alias := aliasCase.alias
			if aliasCase.name == "short name" {
				alias = financeShortDirectory(t, filepath.Join(root, aliasCase.canonical))
			}
			canonicalDir := filepath.Join(root, filepath.FromSlash(aliasCase.canonical))
			canonicalInfo, err := os.Stat(canonicalDir)
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
			marker := documentMarker
			canonicalTarget, target := "/"+aliasCase.canonical, "/"+alias
			if attachment {
				marker = attachmentMarker
				canonicalTarget = "/api/files/" + aliasCase.canonical + "/att.txt"
				target = "/api/files/" + alias + "/att.txt"
			}
			for _, private := range []bool{false, true} {
				name := "public"
				if private {
					name = "private"
				}
				t.Run(name, func(t *testing.T) {
					testConfig.Wiki.Private = private
					// Private outsiders have a real session. The anonymous private control
					// also checks that an alias cannot create access without authentication.
					users := []struct {
						name   string
						cookie *http.Cookie
					}{{"anonymous", nil}}
					if private {
						users = append(users, struct {
							name   string
							cookie *http.Cookie
						}{"outsider", outsider})
					}
					for _, user := range users {
						t.Run(user.name, func(t *testing.T) {
							t.Run("canonical", func(t *testing.T) {
								assertFinanceRefused(t, financeDisclosureRequest(testConfig, canonicalTarget, user.cookie), marker)
							})
							t.Run("alias", func(t *testing.T) {
								t.Logf("requested alias %q for %q", target, canonicalTarget)
								assertFinanceRefused(t, financeDisclosureRequest(testConfig, target, user.cookie), marker)
							})
						})
					}
				})
			}
		})
	}
}

// GetShortPathName returns the volume's actual alias. No 8.3 names or volume
// settings are guessed or changed by the test.
func financeShortDirectory(t *testing.T, directory string) string {
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
