package auth

import (
	"os"
	"path/filepath"
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
