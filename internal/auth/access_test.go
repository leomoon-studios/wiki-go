package auth

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"wiki-go/internal/config"
)

func financeAccessConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg := &config.Config{}
	cfg.Wiki.RootDir = t.TempDir()
	cfg.Wiki.DocumentsDir = "documents"
	cfg.AccessRules = []config.AccessRule{
		{Pattern: "/finance/**", Access: "restricted", Groups: []string{"finance-team"}},
		{Pattern: "/finance-quarterly-2026/**", Access: "restricted", Groups: []string{"finance-team"}},
	}
	for _, name := range []string{"finance/q3", "finance-quarterly-2026", "public"} {
		if err := os.MkdirAll(filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, filepath.FromSlash(name)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return cfg
}

func TestDocumentAccessMissingAndInvalidPaths(t *testing.T) {
	cfg := financeAccessConfig(t)
	for _, session := range []*Session{nil, {Role: config.RoleViewer}, {Role: config.RoleEditor}, {Role: config.RoleViewer, Groups: []string{"finance-team"}}, {Role: config.RoleAdmin}} {
		for _, path := range []string{"/missing", "/public/missing", "/finance/../public", `/finance\..\public`, "/%2fpublic", "/public\x00", "//server/share", `C:\Windows`} {
			if CanAccessDocument(path, session, cfg) {
				t.Errorf("CanAccessDocument(%q, %+v) allowed an invalid or missing resource", path, session)
			}
		}
		if CanAccessDocument("/public", session, nil) {
			t.Fatal("access allowed without configuration")
		}
		if !CanCreateDocument("/public/new/child", session, cfg) {
			t.Fatal("explicit public destination policy was denied")
		}
		want := session != nil && (session.Role == config.RoleAdmin || len(session.Groups) > 0)
		if got := CanCreateDocument("/finance/new/child", session, cfg); got != want {
			t.Errorf("restricted destination access = %v, want %v for %+v", got, want, session)
		}
		if CanCreateDocument("/finance/../public", session, cfg) {
			t.Fatal("invalid creation destination was allowed")
		}
	}
	if err := os.RemoveAll(filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir)); err != nil {
		t.Fatal(err)
	}
	if CanCreateDocument("/new", &Session{Role: config.RoleAdmin}, cfg) {
		t.Fatal("missing selected root was treated as a creation destination")
	}
}

func TestDocumentAccessRuleSemantics(t *testing.T) {
	cfg := financeAccessConfig(t)
	restricted := config.AccessRule{Pattern: "/finance/**", Access: "restricted", Groups: []string{"finance-team"}}
	public := config.AccessRule{Pattern: "/finance/q?", Access: "public"}
	for _, tc := range []struct {
		name  string
		rules []config.AccessRule
		want  bool
	}{
		{"first matching public rule", []config.AccessRule{public, restricted}, true},
		{"first matching restricted rule", []config.AccessRule{restricted, public}, false},
		{"single wildcard", []config.AccessRule{{Pattern: "/fin*/q?", Access: "restricted"}}, false},
		{"single wildcard does not cross slash", []config.AccessRule{{Pattern: "/fin*", Access: "restricted"}}, true},
		{"glob characters stay literal", []config.AccessRule{{Pattern: "/finance/q[3]", Access: "restricted"}}, true},
		{"backslash pattern separator", []config.AccessRule{{Pattern: `\finance\**`, Access: "restricted"}}, false},
		{"platform case matching", []config.AccessRule{{Pattern: "/FINANCE/**", Access: "restricted"}}, runtime.GOOS != "windows"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg.AccessRules = tc.rules
			if got := CanAccessDocument("/finance/q3", nil, cfg); got != tc.want {
				t.Fatalf("access = %v, want %v", got, tc.want)
			}
		})
	}
	cfg.AccessRules = []config.AccessRule{restricted}
	for _, path := range []string{"/finance", `/finance\q3`, "/finance//q3/"} {
		if CanAccessDocument(path, nil, cfg) {
			t.Errorf("normalized restricted path %q was allowed", path)
		}
	}
	if CanAccessDocument("/finance/q3", &Session{Role: config.RoleViewer, Groups: []string{"FINANCE-TEAM"}}, cfg) {
		t.Fatal("group names became case insensitive")
	}
}

func TestDocumentAccessHomepageIdentity(t *testing.T) {
	cfg := financeAccessConfig(t)
	// The built-in homepage is valid even before its default file is created.
	resolved, allowed, err := ResolveDocumentAccess("/", nil, cfg)
	if err != nil || !allowed || resolved.LogicalPath != "/" || resolved.FilesystemPath != filepath.Join(cfg.Wiki.RootDir, "pages", "home") {
		t.Fatalf("default homepage = %+v, %v, %v", resolved, allowed, err)
	}
	for _, dir := range []string{filepath.Join("pages", "home"), filepath.Join(cfg.Wiki.DocumentsDir, "pages", "home")} {
		if err := os.MkdirAll(filepath.Join(cfg.Wiki.RootDir, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	cfg.AccessRules = []config.AccessRule{{Pattern: "/**", Access: "restricted", Groups: []string{"home-team"}}}
	if CanAccessDocument("/", nil, cfg) || !CanAccessDocument("/public", nil, cfg) || !CanAccessDocument("/pages/home", nil, cfg) {
		t.Fatal("root-only recursive rule changed the document namespace")
	}
	for _, session := range []*Session{{Role: config.RoleAdmin}, {Role: config.RoleViewer, Groups: []string{"home-team"}}} {
		if !CanAccessDocument("/", session, cfg) {
			t.Fatal("authorized homepage access was denied")
		}
	}
}

func TestFinanceAccessControls(t *testing.T) {
	for _, private := range []bool{false, true} {
		name := "public"
		if private {
			name = "private"
		}
		t.Run(name, func(t *testing.T) {
			cfg := financeAccessConfig(t)
			cfg.Wiki.Private = private
			for _, user := range []struct {
				name       string
				session    *Session
				restricted bool
			}{
				{"anonymous", nil, false},
				{"outsider", &Session{Role: config.RoleViewer}, false},
				{"member", &Session{Role: config.RoleViewer, Groups: []string{"finance-team"}}, true},
				{"administrator", &Session{Role: config.RoleAdmin}, true},
			} {
				t.Run(user.name, func(t *testing.T) {
					for _, path := range []string{"/finance", "/finance/q3", "/finance-quarterly-2026"} {
						if got := CanAccessDocument(path, user.session, cfg); got != user.restricted {
							t.Errorf("CanAccessDocument(%q) = %v, want %v", path, got, user.restricted)
						}
					}
					if got, want := CanAccessDocument("/public", user.session, cfg), !private || user.session != nil; got != want {
						t.Errorf("public document access = %v, want %v", got, want)
					}
				})
			}
		})
	}
}
