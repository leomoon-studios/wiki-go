package routes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPageIdentityRoutesLinuxCaseSensitive(t *testing.T) {
	cfg, secret, _, _ := pageIdentityFixture(t)
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
	const marker = "DISTINCT_CASE_SENSITIVE_PAGE"
	if err := os.WriteFile(filepath.Join(upper, "document.md"), []byte("# "+marker), 0o644); err != nil {
		t.Fatal(err)
	}
	assertPageIdentityResponse(t, "/finance/q3", nil, false, secret)
	assertPageIdentityResponse(t, "/Finance/q3", nil, true, marker)
	if strings.Contains(pageIdentityResponse("/Finance/q3", nil).Body.String(), secret) {
		t.Fatal("case-sensitive public page disclosed lowercase restricted content")
	}
}
