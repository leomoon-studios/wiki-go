package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"wiki-go/internal/config"
)

func TestWindowsDocumentAPIAliases(t *testing.T) {
	for _, alias := range []string{"Finance/Q3", `FINANCE\q3`, "finance-quarterly-2026"} {
		t.Run(alias, func(t *testing.T) {
			canonical := "finance/q3"
			if alias == "finance-quarterly-2026" {
				canonical = alias
			}
			cfg, marker := documentAPIFixture(t, canonical)
			if canonical == "finance-quarterly-2026" {
				alias = financeShortDirectory(t, filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, canonical))
			}
			testDocumentAPIAliasDenial(t, cfg, alias, marker)
			testDocumentAPIAliasAuthorized(t, cfg, alias, canonical, marker)
		})
	}
}

func TestWindowsDocumentAPICaseRename(t *testing.T) {
	cfg, marker := documentAPIFixture(t, "finance/q3")
	member := sessionCookieForUserWithGroups(t, cfg, "member", config.RoleEditor, []string{"finance-team"})
	payload, _ := json.Marshal(MoveRequest{SourcePath: "finance/q3", NewSlug: "Q3"})
	request := httptest.NewRequest(http.MethodPost, "/api/document/move", strings.NewReader(string(payload)))
	request.AddCookie(member)
	response := httptest.NewRecorder()
	MoveDocumentHandler(response, request, cfg)
	if response.Code != http.StatusOK {
		t.Fatalf("case-only rename = %d %s", response.Code, response.Body.String())
	}
	if contents, err := os.ReadFile(filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, "finance", "Q3", "document.md")); err != nil || string(contents) != marker {
		t.Fatalf("case-only rename lost content: %v", err)
	}
}

func TestWindowsDocumentAPIInvalidAliases(t *testing.T) {
	cfg, _ := documentAPIFixture(t, "finance/q3")
	admin := sessionCookieForUser(t, cfg, "admin", config.RoleAdmin)
	before := apiIdentitySnapshot(t, cfg)
	for _, alias := range []string{"finance./q3", "finance /q3", "finance/q3.", "finance/q3 ", "finance/q3:stream", "PAGES./HOME", `C:\Windows`, `\\server\share`, "finance/%2fq3"} {
		for _, operation := range []string{"source", "save", "create", "delete", "move source", "move target", "versions", "version read", "version restore", "version delete", "comments", "comment add", "comment delete", "link add", "link edit", "link delete"} {
			response := apiIdentityRequest(t, cfg, operation, alias, admin)
			if response.Code < 400 || response.Code >= 500 {
				t.Fatalf("%s via %q = %d", operation, alias, response.Code)
			}
		}
	}
	assertAPIIdentitySnapshot(t, cfg, before)
}

func TestWindowsDocumentAPIHomepageAliases(t *testing.T) {
	cfg, _ := documentAPIFixture(t, "finance/q3")
	member := sessionCookieForUserWithGroups(t, cfg, "member", config.RoleEditor, []string{"finance-team"})
	// Storage-style APIs accept the homepage alias and share the canonical
	// history/comments namespace; source and save keep their logical / contract.
	writeAPIIdentityFile(t, filepath.Join(cfg.Wiki.RootDir, "versions", "pages", "home", identityVersion+".md"), "HOME_VERSION")
	writeAPIIdentityFile(t, filepath.Join(cfg.Wiki.RootDir, "comments", identityComment), "HOME_COMMENT")
	for _, operation := range []string{"versions", "version read", "comments", "comment add", "version restore"} {
		if response := apiIdentityRequest(t, cfg, operation, "PaGeS/HoMe", member); response.Code != http.StatusOK {
			t.Fatalf("homepage %s = %d %s", operation, response.Code, response.Body.String())
		}
	}
	if contents, err := os.ReadFile(filepath.Join(cfg.Wiki.RootDir, "pages", "home", "document.md")); err != nil || string(contents) != "HOME_VERSION" {
		t.Fatalf("homepage restore changed its namespace: %v", err)
	}
}
