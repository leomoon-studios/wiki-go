package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestDocumentAPILinuxCaseSensitive(t *testing.T) {
	cfg, marker := documentAPIFixture(t, "finance/q3")
	lower := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, "finance", "q3")
	upper := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, "Finance", "q3")
	writeAPIIdentityFile(t, filepath.Join(upper, "document.md"), "UPPERCASE_PUBLIC")
	lowInfo, err := os.Stat(lower)
	if err != nil {
		t.Fatal(err)
	}
	upInfo, err := os.Stat(upper)
	if err != nil {
		t.Fatal(err)
	}
	if os.SameFile(lowInfo, upInfo) {
		t.Skip("filesystem does not support distinct case-sensitive directories")
	}
	editor := sessionCookieForUser(t, cfg, "editor", "editor")
	response := apiIdentityRequest(t, cfg, "save", "Finance/q3", editor)
	if response.Code != http.StatusOK {
		t.Fatalf("distinct uppercase save = %d %s", response.Code, response.Body.String())
	}
	if contents, err := os.ReadFile(filepath.Join(lower, "document.md")); err != nil || string(contents) != marker {
		t.Fatalf("uppercase save changed lowercase document: %v", err)
	}
	if response := apiIdentityRequest(t, cfg, "source", "finance/q3", editor); response.Code != http.StatusForbidden {
		t.Fatalf("lowercase source = %d", response.Code)
	}
	if response := apiIdentityRequest(t, cfg, "source", "Finance/q3", editor); response.Code != http.StatusOK || response.Body.String() != "# UPDATED" {
		t.Fatalf("uppercase source = %d", response.Code)
	}
}
