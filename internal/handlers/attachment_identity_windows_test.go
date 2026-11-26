package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"wiki-go/internal/config"
)

func TestWindowsAttachmentIdentityOperations(t *testing.T) {
	for _, tc := range []struct {
		name, storage, physical string
		restricted              bool
	}{
		{"mixed case", "Finance/Q3", "knowledge/finance/q3", true},
		{"backslash", `FINANCE\q3`, "knowledge/finance/q3", true},
		{"public case", "PUBLIC", "knowledge/public", false},
		{"homepage case", "PAGES/HOME", "pages/home", true},
		{"homepage nested", `PAGES\HOME\ASSETS`, "pages/home/assets", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, marker := attachmentIdentityFixture(t)
			testAttachmentIdentityOperations(t, cfg, tc.storage, filepath.Join(cfg.Wiki.RootDir, filepath.FromSlash(tc.physical)), "ATT.TXT", "att.txt", marker, tc.restricted)
		})
	}
	t.Run("8.3 directory and file", func(t *testing.T) {
		cfg, marker := attachmentIdentityFixture(t)
		directory := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, "finance-quarterly-2026")
		filename := "quarterly-attachment-2026.txt"
		if err := os.WriteFile(filepath.Join(directory, filename), []byte(marker), 0o644); err != nil {
			t.Fatal(err)
		}
		shortDirectory := financeShortDirectory(t, directory)
		shortFile := financeShortDirectory(t, filepath.Join(directory, filename))
		testAttachmentIdentityOperations(t, cfg, shortDirectory, directory, shortFile, filename, marker, true)
	})
}

func TestWindowsAttachmentIdentityInvalidAliases(t *testing.T) {
	for _, storage := range []string{"finance./q3", "finance /q3", "finance/q3.", "finance/q3 ", "PAGES./HOME", "PAGES/HOME.", "pages/other", "finance/q3:stream"} {
		t.Run(storage, func(t *testing.T) {
			cfg, marker := attachmentIdentityFixture(t)
			admin := sessionCookieForUser(t, cfg, "admin", config.RoleAdmin)
			for _, operation := range []string{"read", "list", "upload", "rename", "delete"} {
				assertAttachmentIdentityDenied(t, attachmentIdentityRequest(t, cfg, operation, storage, "att.txt", admin), marker)
			}
			filename := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, "finance", "q3", "att.txt")
			contents, err := os.ReadFile(filename)
			if err != nil || string(contents) != marker {
				t.Fatalf("invalid alias changed original: %v", err)
			}
		})
	}
	for _, filename := range []string{"att.txt.", "att.txt ", "att.txt:stream", "att.txt::$DATA", "NUL", "CON"} {
		t.Run(filename, func(t *testing.T) {
			cfg, marker := attachmentIdentityFixture(t)
			admin := sessionCookieForUser(t, cfg, "admin", config.RoleAdmin)
			for _, operation := range []string{"read", "rename", "delete"} {
				assertAttachmentIdentityDenied(t, attachmentIdentityRequest(t, cfg, operation, "finance/q3", filename, admin), marker)
			}
			payload, err := json.Marshal(map[string]string{"currentPath": "finance/q3/att.txt", "newName": filename})
			if err != nil {
				t.Fatal(err)
			}
			request := httptest.NewRequest(http.MethodPost, "/api/files/rename", strings.NewReader(string(payload)))
			request.AddCookie(admin)
			response := httptest.NewRecorder()
			RenameFileHandler(response, request, cfg)
			assertAttachmentIdentityDenied(t, response, marker)
		})
	}
}

func TestWindowsAttachmentIdentitySVG(t *testing.T) {
	cfg, _ := attachmentIdentityFixture(t)
	member := sessionCookieForUserWithGroups(t, cfg, "member", config.RoleEditor, []string{"finance-team"})
	directory := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, "finance", "q3")
	for name, contents := range map[string]string{"safe.svg": safeSVGFixture, "unsafe.svg": `<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	response := attachmentIdentityRequest(t, cfg, "read", "FiNaNcE/Q3", "SAFE.SVG", member)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "<svg") || !strings.Contains(response.Header().Get("Content-Security-Policy"), "sandbox") || response.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("SVG alias lost sanitization or isolation headers: %d", response.Code)
	}
	response = attachmentIdentityRequest(t, cfg, "read", "FINANCE/Q3", "UNSAFE.SVG", member)
	if response.Code != http.StatusNotFound || strings.Contains(response.Body.String(), "<script") {
		t.Fatal("unsafe stored SVG served through case alias")
	}
}

func TestWindowsAttachmentIdentityJunctions(t *testing.T) {
	cfg, marker := attachmentIdentityFixture(t)
	root := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir)
	outside := filepath.Join(cfg.Wiki.RootDir, "outside")
	if err := os.Mkdir(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outside, "att.txt"), []byte(marker), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, link := range []struct{ name, target string }{
		{"finance-alias", filepath.Join(root, "finance", "q3")}, {"escape", outside},
	} {
		junction := filepath.Join(root, link.name)
		if output, err := exec.Command("cmd", "/c", "mklink", "/J", junction, link.target).CombinedOutput(); err != nil {
			t.Fatalf("create junction: %v: %s", err, output)
		}
		t.Cleanup(func() {
			if err := os.Remove(junction); err != nil {
				t.Error(err)
			}
		})
	}
	testAttachmentIdentityOperations(t, cfg, "finance-alias", filepath.Join(root, "finance", "q3"), "att.txt", "att.txt", marker, true)
	admin := sessionCookieForUser(t, cfg, "admin", config.RoleAdmin)
	for _, operation := range []string{"read", "list", "upload", "rename", "delete"} {
		assertAttachmentIdentityDenied(t, attachmentIdentityRequest(t, cfg, operation, "escape", "att.txt", admin), marker)
	}
	contents, err := os.ReadFile(filepath.Join(outside, "att.txt"))
	if err != nil || string(contents) != marker {
		t.Fatalf("junction escaped the storage root: %v", err)
	}
	// Homepage aliases must select homepage storage before authorization; no
	// similarly named unrestricted documents directory may substitute for it.
	cfg.Wiki.Private = false
	response := attachmentIdentityRequest(t, cfg, "read", "PaGeS/HoMe/AsSeTs", "ATT.TXT", nil)
	if response.Code == http.StatusOK || strings.Contains(response.Body.String(), marker) {
		t.Fatal("homepage case alias bypassed the root rule")
	}
}
