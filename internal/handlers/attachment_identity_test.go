package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"

	"wiki-go/internal/config"
)

func attachmentIdentityFixture(t *testing.T) (*config.Config, string) {
	t.Helper()
	cfg, _, marker := financeDisclosureFixture(t)
	if err := os.Rename(filepath.Join(cfg.Wiki.RootDir, "documents"), filepath.Join(cfg.Wiki.RootDir, "knowledge")); err != nil {
		t.Fatal(err)
	}
	cfg.Wiki.DocumentsDir = "knowledge"
	cfg.Wiki.MaxUploadSize = 10
	cfg.Wiki.DisableFileUploadChecking = true
	cfg.AccessRules = append(cfg.AccessRules, config.AccessRule{Pattern: "/**", Access: "restricted", Groups: []string{"finance-team"}})
	for _, directory := range []string{"pages/home", "pages/home/assets", "knowledge/public", "knowledge/public/assets"} {
		filename := filepath.Join(cfg.Wiki.RootDir, filepath.FromSlash(directory), "att.txt")
		if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filename, []byte(marker), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return cfg, marker
}

func attachmentIdentityRequest(t *testing.T, cfg *config.Config, operation, storage, filename string, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	var request *http.Request
	var handler func(http.ResponseWriter, *http.Request, *config.Config)
	file := path.Join(storage, filename)
	escaped := (&url.URL{Path: file}).EscapedPath()
	switch operation {
	case "read":
		request = httptest.NewRequest(http.MethodGet, "/api/files/"+escaped, nil)
		handler = ServeFileHandler
	case "list":
		request = httptest.NewRequest(http.MethodGet, "/api/files/list/"+(&url.URL{Path: storage}).EscapedPath(), nil)
		handler = ListFilesHandler
	case "upload":
		request = attachmentUploadRequest(t, storage, filename, "UPLOADED_ATTACHMENT")
		handler = UploadFileHandler
	case "rename":
		payload, err := json.Marshal(map[string]string{"currentPath": file, "newName": "renamed.txt"})
		if err != nil {
			t.Fatal(err)
		}
		request = httptest.NewRequest(http.MethodPost, "/api/files/rename", strings.NewReader(string(payload)))
		handler = RenameFileHandler
	case "delete":
		request = httptest.NewRequest(http.MethodDelete, "/api/files/delete/"+escaped, nil)
		handler = DeleteFileHandler
	default:
		t.Fatalf("unknown attachment operation %q", operation)
	}
	if cookie != nil {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	handler(response, request, cfg)
	return response
}

func assertAttachmentIdentityDenied(t *testing.T, response *httptest.ResponseRecorder, marker string) {
	t.Helper()
	if response.Code < 400 || response.Code >= 500 || strings.Contains(response.Body.String(), marker) {
		t.Fatalf("denial status=%d marker_present=%v: %s", response.Code, strings.Contains(response.Body.String(), marker), response.Body.String())
	}
}

// Exercise both denial without filesystem changes and a complete authorized
// upload/read/rename/delete cycle through the same storage spelling.
func testAttachmentIdentityOperations(t *testing.T, cfg *config.Config, storage, canonicalDir, filename, actualFilename, marker string, restricted bool) {
	t.Helper()
	users := []struct {
		name   string
		cookie *http.Cookie
		member bool
		edit   bool
	}{
		{"anonymous", nil, false, false},
		{"outsider viewer", sessionCookieForUser(t, cfg, "viewer", config.RoleViewer), false, false},
		{"outsider editor", sessionCookieForUser(t, cfg, "editor", config.RoleEditor), false, true},
		{"member", sessionCookieForUserWithGroups(t, cfg, "member", config.RoleEditor, []string{"finance-team"}), true, true},
		{"administrator", sessionCookieForUser(t, cfg, "admin", config.RoleAdmin), true, true},
	}
	for _, private := range []bool{false, true} {
		cfg.Wiki.Private = private
		for _, user := range users {
			t.Run(user.name+map[bool]string{false: "/public", true: "/private"}[private], func(t *testing.T) {
				allowed := (!private || user.cookie != nil) && (!restricted || user.member)
				for _, operation := range []string{"read", "list"} {
					response := attachmentIdentityRequest(t, cfg, operation, storage, filename, user.cookie)
					if !allowed {
						assertAttachmentIdentityDenied(t, response, marker)
						if operation == "list" && strings.Contains(response.Body.String(), actualFilename) {
							t.Fatal("denied listing disclosed attachment metadata")
						}
					} else if response.Code != http.StatusOK || operation == "read" && response.Body.String() != marker || operation == "list" && !strings.Contains(response.Body.String(), actualFilename) {
						t.Fatalf("authorized %s status=%d: %s", operation, response.Code, response.Body.String())
					}
				}
				if !allowed || !user.edit {
					for _, operation := range []string{"upload", "rename", "delete"} {
						response := attachmentIdentityRequest(t, cfg, operation, storage, filename, user.cookie)
						assertAttachmentIdentityDenied(t, response, marker)
						contents, err := os.ReadFile(filepath.Join(canonicalDir, actualFilename))
						if err != nil || string(contents) != marker {
							t.Fatalf("denied %s changed original attachment: %v", operation, err)
						}
						if _, err := os.Stat(filepath.Join(canonicalDir, "renamed.txt")); !os.IsNotExist(err) {
							t.Fatalf("denied %s created a rename destination: %v", operation, err)
						}
					}
					return
				}
				// Exercise the requested spelling itself for every mutation,
				// including an actual Windows short filename when supplied.
				for _, operation := range []string{"upload", "rename", "delete"} {
					name := filename
					if operation == "delete" {
						name = "renamed.txt"
					}
					response := attachmentIdentityRequest(t, cfg, operation, storage, name, user.cookie)
					if response.Code != http.StatusOK {
						t.Fatalf("authorized alias %s status=%d: %s", operation, response.Code, response.Body.String())
					}
				}
				if err := os.WriteFile(filepath.Join(canonicalDir, actualFilename), []byte(marker), 0o644); err != nil {
					t.Fatal(err)
				}
				for _, operation := range []string{"upload", "read", "rename", "delete"} {
					name := "uploaded.txt"
					if operation == "delete" {
						name = "renamed.txt"
					}
					response := attachmentIdentityRequest(t, cfg, operation, storage, name, user.cookie)
					if response.Code != http.StatusOK {
						t.Fatalf("authorized %s status=%d: %s", operation, response.Code, response.Body.String())
					}
					if operation == "read" && response.Body.String() != "UPLOADED_ATTACHMENT" {
						t.Fatal("uploaded attachment was not read back")
					}
				}
				for _, name := range []string{"uploaded.txt", "renamed.txt"} {
					if _, err := os.Stat(filepath.Join(canonicalDir, name)); !os.IsNotExist(err) {
						t.Fatalf("authorized cycle did not remove %s: %v", name, err)
					}
				}
			})
		}
	}
}

func TestAttachmentIdentityOperations(t *testing.T) {
	for _, tc := range []struct {
		storage, physical string
		restricted        bool
	}{
		{"finance/q3", "knowledge/finance/q3", true},
		{"public", "knowledge/public", false},
		{"pages/home", "pages/home", true},
		{"pages/home/assets", "pages/home/assets", true},
	} {
		t.Run(tc.storage, func(t *testing.T) {
			cfg, marker := attachmentIdentityFixture(t)
			testAttachmentIdentityOperations(t, cfg, tc.storage, filepath.Join(cfg.Wiki.RootDir, filepath.FromSlash(tc.physical)), "att.txt", "att.txt", marker, tc.restricted)
		})
	}
}

func TestAttachmentIdentityFileLinks(t *testing.T) {
	cfg, marker := attachmentIdentityFixture(t)
	root := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir)
	public := filepath.Join(root, "public")
	outside := filepath.Join(cfg.Wiki.RootDir, "outside.txt")
	if err := os.WriteFile(outside, []byte(marker), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, target string }{
		{"restricted.txt", filepath.Join(root, "finance", "q3", "att.txt")},
		{"external.txt", outside},
		{"dangling.txt", filepath.Join(public, "missing.txt")},
		{"loop.txt", filepath.Join(public, "loop.txt")},
		{"markdown.txt", filepath.Join(public, "document.md")},
	} {
		if err := os.Symlink(tc.target, filepath.Join(public, tc.name)); err != nil {
			if errors.Is(err, os.ErrPermission) || errors.Is(err, errors.ErrUnsupported) || runtime.GOOS == "windows" && errors.Is(err, syscall.Errno(1314)) {
				t.Skipf("symlink creation unavailable: %v", err)
			}
			t.Fatal(err)
		}
	}
	editor := sessionCookieForUser(t, cfg, "editor", config.RoleEditor)
	for _, name := range []string{"restricted.txt", "external.txt", "dangling.txt", "loop.txt", "markdown.txt"} {
		for _, operation := range []string{"read", "upload", "rename", "delete"} {
			assertAttachmentIdentityDenied(t, attachmentIdentityRequest(t, cfg, operation, "public", name, editor), marker)
		}
	}
	for _, filename := range []string{outside, filepath.Join(root, "finance", "q3", "att.txt")} {
		contents, err := os.ReadFile(filename)
		if err != nil || string(contents) != marker {
			t.Fatalf("linked target changed: %s: %v", filename, err)
		}
	}
	listing := attachmentIdentityRequest(t, cfg, "list", "public", "", editor)
	if listing.Code != http.StatusOK || strings.Contains(listing.Body.String(), "restricted.txt") || strings.Contains(listing.Body.String(), "external.txt") || strings.Contains(listing.Body.String(), "markdown.txt") {
		t.Fatalf("listing disclosed invalid file links: %d %s", listing.Code, listing.Body.String())
	}
}

func TestAttachmentIdentityDocumentProtection(t *testing.T) {
	cfg, _ := attachmentIdentityFixture(t)
	editor := sessionCookieForUser(t, cfg, "editor", config.RoleEditor)
	filename := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, "public", "document.md")
	before, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	for _, operation := range []string{"read", "upload", "rename", "delete"} {
		assertAttachmentIdentityDenied(t, attachmentIdentityRequest(t, cfg, operation, "public", "document.md", editor), string(before))
	}
	request := httptest.NewRequest(http.MethodPost, "/api/files/rename", strings.NewReader(`{"currentPath":"public/att.txt","newName":"document.md"}`))
	request.AddCookie(editor)
	response := httptest.NewRecorder()
	RenameFileHandler(response, request, cfg)
	assertAttachmentIdentityDenied(t, response, string(before))
	after, err := os.ReadFile(filename)
	if err != nil || string(after) != string(before) {
		t.Fatalf("document changed through attachment API: %v", err)
	}
}

func TestAttachmentIdentityServingControls(t *testing.T) {
	cfg, marker := attachmentIdentityFixture(t)
	cfg.Wiki.DisableFileUploadChecking = false
	root := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, "public")
	for name, content := range map[string]string{
		"download.pdf": "%PDF-1.4\n" + marker,
		"no-extension": marker,
		"fake.png":     "<html>" + marker + "</html>",
		"notes.md":     marker,
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"no-extension", "download.pdf"} {
		response := attachmentIdentityRequest(t, cfg, "read", "public", name, nil)
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), marker) {
			t.Fatalf("public attachment %s status=%d", name, response.Code)
		}
		if name == "download.pdf" && !strings.Contains(response.Header().Get("Content-Disposition"), "attachment;") {
			t.Fatal("binary attachment lost download disposition")
		}
	}
	for _, name := range []string{"fake.png", "notes.md", "missing.txt"} {
		response := attachmentIdentityRequest(t, cfg, "read", "public", name, nil)
		if response.Code != http.StatusNotFound || strings.Contains(response.Body.String(), marker) {
			t.Fatalf("invalid attachment %s status=%d", name, response.Code)
		}
	}
	// Serving through the resolved resource retains range support.
	request := httptest.NewRequest(http.MethodGet, "/api/files/public/att.txt", nil)
	request.Header.Set("Range", "bytes=0-3")
	response := httptest.NewRecorder()
	ServeFileHandler(response, request, cfg)
	if response.Code != http.StatusPartialContent || response.Body.String() != marker[:4] {
		t.Fatalf("attachment range response = %d %q", response.Code, response.Body.String())
	}
}

func TestAttachmentIdentityOwnerPaths(t *testing.T) {
	for input, want := range map[string]string{
		"finance/q3/att.txt":        "/finance/q3",
		`finance\q3\att.txt`:        "/finance/q3",
		"pages/home/att.txt":        "/",
		`pages\home\assets\att.txt`: "/",
	} {
		got, err := attachmentDocumentPath(input)
		if err != nil || got != want || strings.Contains(got, "\\") {
			t.Errorf("attachmentDocumentPath(%q) = %q, %v; want %q", input, got, err, want)
		}
	}
}
