package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAttachmentIdentityLinuxCaseSensitive(t *testing.T) {
	cfg, marker := attachmentIdentityFixture(t)
	lower := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, "finance", "q3")
	upper := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, "Finance", "q3")
	if err := os.MkdirAll(upper, 0o755); err != nil {
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
	if err := os.WriteFile(filepath.Join(upper, "att.txt"), []byte(marker), 0o644); err != nil {
		t.Fatal(err)
	}
	testAttachmentIdentityOperations(t, cfg, "Finance/q3", upper, "att.txt", "att.txt", marker, false)
	contents, err := os.ReadFile(filepath.Join(lower, "att.txt"))
	if err != nil || string(contents) != marker {
		t.Fatalf("distinct lowercase attachment changed: %v", err)
	}
}

func TestAttachmentIdentityLinuxPermissionFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses filesystem permission checks")
	}
	cfg, marker := attachmentIdentityFixture(t)
	parent := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, "public")
	if err := os.Chmod(parent, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(parent, 0o755); err != nil {
			t.Error(err)
		}
	})
	for _, operation := range []string{"read", "list"} {
		response := attachmentIdentityRequest(t, cfg, operation, "public", "att.txt", nil)
		if response.Code < 400 || strings.Contains(response.Body.String(), marker) {
			t.Fatalf("permission failure status=%d disclosed restricted contents", response.Code)
		}
	}
}
