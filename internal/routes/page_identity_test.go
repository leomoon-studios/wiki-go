package routes

import (
	"crypto/rand"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"

	"wiki-go/internal/auth"
	"wiki-go/internal/config"
	"wiki-go/internal/goldext"
	"wiki-go/internal/handlers"
)

func pageIdentityFixture(t *testing.T) (*config.Config, string, string, string) {
	t.Helper()
	cfg := newRouteSecurityTestConfig(t)
	cfg.Wiki.DocumentsDir = "knowledge"
	cfg.AccessRules = append(cfg.AccessRules, config.AccessRule{Pattern: "/finance-quarterly-2026/**", Access: "restricted", Groups: []string{"finance"}})
	secret, public, home := "SECRET"+rand.Text(), "PUBLIC"+rand.Text(), "HOME"+rand.Text()
	for relative, content := range map[string]string{
		"knowledge/finance/q3/document.md":             "# " + secret,
		"knowledge/finance-quarterly-2026/document.md": "# " + secret,
		"knowledge/public/document.md":                 "# " + public + "\n\n:::stats recent=100:::",
		"knowledge/home/document.md":                   "# Documents home\n\nWRONG_DOCUMENTS_HOME",
		"pages/home/document.md":                       "# " + home,
	} {
		filename := filepath.Join(cfg.Wiki.RootDir, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	previousRoot := goldext.DocumentsRoot()
	goldext.ConfigureContentPaths(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir)
	t.Cleanup(func() { goldext.ConfigureContentPaths(filepath.Dir(previousRoot), filepath.Base(previousRoot)) })
	return cfg, secret, public, home
}

func TestPageIdentityPreparedRequest(t *testing.T) {
	cfg, secret, public, _ := pageIdentityFixture(t)
	entry := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, "entry")
	if err := os.Symlink(filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, "public"), entry); err != nil {
		if errors.Is(err, os.ErrPermission) || errors.Is(err, errors.ErrUnsupported) || runtime.GOOS == "windows" && errors.Is(err, syscall.Errno(1314)) {
			t.Skipf("symlink creation unavailable: %v", err)
		}
		t.Fatal(err)
	}
	request, _, allowed := auth.PrepareDocumentRequest(httptest.NewRequest(http.MethodGet, "/entry", nil), cfg)
	if !allowed {
		t.Fatal("public entry denied")
	}
	if err := os.Remove(entry); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, "finance", "q3"), entry); err != nil {
		t.Fatal(err)
	}
	// The handler must read the public identity authorized by the route, rather
	// than reconstructing a filesystem path from the changed entry spelling.
	response := httptest.NewRecorder()
	handlers.PageHandler(response, request, cfg)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), public) || strings.Contains(response.Body.String(), secret) {
		t.Fatalf("prepared page used a different identity: status=%d", response.Code)
	}
	// A fresh registered-route request resolves the changed alias and denies it.
	assertPageIdentityResponse(t, "/entry", nil, false, secret)
	// Reusing the context for a different URL must not reuse its access decision.
	changed := request.Clone(request.Context())
	changed.URL.Path = "/finance/q3"
	response = httptest.NewRecorder()
	handlers.PageHandler(response, changed, cfg)
	if response.Code == http.StatusOK || strings.Contains(response.Body.String(), secret) {
		t.Fatal("prepared identity was reused for a different request path")
	}
}

func pageIdentityUsers(t *testing.T, cfg *config.Config) []struct {
	name   string
	cookie *http.Cookie
	member bool
} {
	t.Helper()
	return []struct {
		name   string
		cookie *http.Cookie
		member bool
	}{
		{"anonymous", nil, false},
		{"outsider", routeSessionCookie(t, cfg, "outsider", config.RoleViewer, nil), false},
		{"member", routeSessionCookie(t, cfg, "member", config.RoleViewer, []string{"finance"}), true},
		{"administrator", routeSessionCookie(t, cfg, "admin", config.RoleAdmin, nil), true},
	}
}

func pageIdentityResponse(target string, cookie *http.Cookie) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, target, nil)
	if cookie != nil {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(response, request)
	return response
}

func assertPageIdentityResponse(t *testing.T, target string, cookie *http.Cookie, allowed bool, marker string) {
	t.Helper()
	response := pageIdentityResponse(target, cookie)
	present := strings.Contains(response.Body.String(), marker)
	t.Logf("target=%q status=%d marker_present=%v", target, response.Code, present)
	if allowed {
		if response.Code != http.StatusOK || !present {
			t.Fatalf("authorized page status=%d marker_present=%v", response.Code, present)
		}
		return
	}
	if present {
		t.Error("denied page disclosed its content marker")
	}
	switch response.Code {
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusFound, http.StatusSeeOther:
	default:
		t.Errorf("denied page status=%d", response.Code)
	}
}

func TestPageIdentityRoutesCanonical(t *testing.T) {
	cfg, secret, public, home := pageIdentityFixture(t)
	users := pageIdentityUsers(t, cfg)
	for _, private := range []bool{false, true} {
		cfg.Wiki.Private = private
		for _, user := range users {
			t.Run(user.name+map[bool]string{false: "/public", true: "/private"}[private], func(t *testing.T) {
				for _, target := range []string{"/finance/q3", "/finance-quarterly-2026"} {
					assertPageIdentityResponse(t, target, user.cookie, user.member, secret)
				}
				publicAllowed := !private || user.cookie != nil
				assertPageIdentityResponse(t, "/public", user.cookie, publicAllowed, public)
				assertPageIdentityResponse(t, "/", user.cookie, publicAllowed, home)
				if publicAllowed && !user.member {
					for _, target := range []string{"/public", "/missing"} {
						response := pageIdentityResponse(target, user.cookie)
						if strings.Contains(response.Body.String(), secret) {
							t.Fatalf("%s disclosed a restricted title through navigation or rendered stats", target)
						}
					}
				}
				if publicAllowed && strings.Contains(pageIdentityResponse("/", user.cookie).Body.String(), "WRONG_DOCUMENTS_HOME") {
					t.Fatal("homepage read from the documents root")
				}
			})
		}
	}
	cfg.Wiki.Private = false
	cfg.AccessRules = append([]config.AccessRule{{Pattern: "/**", Access: "restricted", Groups: []string{"finance"}}}, cfg.AccessRules...)
	assertPageIdentityResponse(t, "/", nil, false, home)
	assertPageIdentityResponse(t, "/public", nil, true, public)
	assertPageIdentityResponse(t, "/", users[2].cookie, true, home)
}

func TestPageIdentityRoutesCreationDestinations(t *testing.T) {
	cfg, _, _, _ := pageIdentityFixture(t)
	member := routeSessionCookie(t, cfg, "editor", config.RoleEditor, []string{"finance"})
	for _, target := range []string{"/public/new", "/finance/new"} {
		request := httptest.NewRequest(http.MethodPost, "/api/save"+target, strings.NewReader("# Created document"))
		request.AddCookie(member)
		response := httptest.NewRecorder()
		http.DefaultServeMux.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("save destination %s status=%d: %s", target, response.Code, response.Body.String())
		}
		assertPageIdentityResponse(t, target, member, true, "Created document")
	}
	if response := pageIdentityResponse("/missing", nil); response.Code != http.StatusNotFound {
		t.Fatalf("missing public document status=%d, want 404", response.Code)
	}
}
