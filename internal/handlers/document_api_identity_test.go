package handlers

import (
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"syscall"
	"testing"

	"wiki-go/internal/comments"
	"wiki-go/internal/config"
)

const identityVersion = "20250102030405"
const identityComment = "20250102030405_member.md"

func documentAPIFixture(t *testing.T, canonical string) (*config.Config, string) {
	t.Helper()
	cfg, marker, _ := financeDisclosureFixture(t)
	if err := os.Rename(filepath.Join(cfg.Wiki.RootDir, "documents"), filepath.Join(cfg.Wiki.RootDir, "knowledge")); err != nil {
		t.Fatal(err)
	}
	cfg.Wiki.DocumentsDir = "knowledge"
	cfg.Wiki.DisableComments = false
	cfg.Wiki.MaxVersions = 10
	for name, contents := range map[string]string{
		"versions/documents/" + canonical + "/" + identityVersion + ".md": "VERSION" + marker,
		"comments/" + canonical + "/" + identityComment:                   "COMMENT" + marker,
		"pages/home/document.md":                                          "# Home",
	} {
		writeAPIIdentityFile(t, filepath.Join(cfg.Wiki.RootDir, filepath.FromSlash(name)), contents)
	}
	return cfg, marker
}

func writeAPIIdentityFile(t *testing.T, filename, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filename, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func apiIdentityRequest(t *testing.T, cfg *config.Config, operation, document string, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	method, target, body := http.MethodGet, "", ""
	var handler http.HandlerFunc
	escape := func(value string) string { return (&url.URL{Path: value}).EscapedPath() }
	payload := func(value any) string {
		result, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return string(result)
	}
	switch operation {
	case "source":
		target = "/api/source/" + escape(document)
		handler = SourceHandler
	case "save":
		method = http.MethodPost
		target = "/api/save/" + escape(document)
		body = "# UPDATED"
		handler = SaveHandler
	case "create":
		method = http.MethodPost
		target = "/api/document/create"
		body = payload(CreateDocumentRequest{Title: "New document", Path: document + "/new", Type: "document"})
		handler = CreateDocumentHandler
	case "delete":
		method = http.MethodDelete
		target = "/api/document/" + escape(document)
		handler = DeleteDocumentHandler
	case "move source", "move target":
		method = http.MethodPost
		target = "/api/document/move"
		move := MoveRequest{SourcePath: document, TargetPath: "public", NewSlug: "moved"}
		if document == "public" {
			move.TargetPath = "archive"
		}
		if operation == "move target" {
			move = MoveRequest{SourcePath: "public", TargetPath: document, NewSlug: "moved"}
		}
		body = payload(move)
		handler = func(w http.ResponseWriter, r *http.Request) { MoveDocumentHandler(w, r, cfg) }
	case "versions", "version read", "version restore", "version delete":
		target = "/api/versions/" + escape(document)
		if operation != "versions" {
			target += "/" + identityVersion
		}
		if operation == "version restore" {
			method = http.MethodPost
			target += "/restore"
		}
		if operation == "version delete" {
			method = http.MethodDelete
		}
		handler = func(w http.ResponseWriter, r *http.Request) { VersionsHandler(w, r, cfg) }
	case "comments":
		target = "/api/comments/" + escape(document)
		handler = GetCommentsHandler
	case "comment add":
		method = http.MethodPost
		target = "/api/comments/add/" + escape(document)
		body = `{"content":"ADDED_COMMENT"}`
		handler = AddCommentHandler
	case "comment delete":
		method = http.MethodDelete
		target = "/api/comments/delete/" + escape(document) + "/" + identityComment
		handler = DeleteCommentHandler
	case "link add", "link edit", "link delete":
		method = http.MethodPost
		action := strings.TrimPrefix(operation, "link ")
		target = "/api/links/" + action + "/" + escape(document)
		body = `{"url":"https://example.com","title":"Example","category":"Links","description":"Example link"}`
		if action == "edit" {
			body = `{"oldUrl":"https://example.com","newLink":{"url":"https://example.com","title":"Example","category":"Links","description":"Example link"}}`
		}
		switch action {
		case "add":
			handler = AddLinkHandler
		case "edit":
			method = http.MethodPut
			handler = EditLinkHandler
		case "delete":
			method = http.MethodDelete
			handler = DeleteLinkHandler
		}
	default:
		t.Fatalf("unknown API operation %q", operation)
	}
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	if cookie != nil {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	handler(response, request)
	return response
}

func apiIdentitySnapshot(t *testing.T, cfg *config.Config) map[string]string {
	t.Helper()
	result := map[string]string{}
	for _, root := range []string{cfg.Wiki.DocumentsDir, "versions", "comments"} {
		err := filepath.WalkDir(filepath.Join(cfg.Wiki.RootDir, root), func(filename string, entry fs.DirEntry, err error) error {
			if os.IsNotExist(err) {
				return nil
			}
			if err != nil {
				return err
			}
			relative, err := filepath.Rel(cfg.Wiki.RootDir, filename)
			if err != nil {
				return err
			}
			if entry.IsDir() {
				result[relative] = "directory"
				return nil
			}
			if entry.Type()&os.ModeSymlink != 0 {
				target, err := os.Readlink(filename)
				result[relative] = "link:" + target
				return err
			}
			contents, err := os.ReadFile(filename)
			result[relative] = string(contents)
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return result
}

func testDocumentAPIAliasDenial(t *testing.T, cfg *config.Config, alias, marker string) {
	t.Helper()
	for _, private := range []bool{false, true} {
		cfg.Wiki.Private = private
		for _, user := range []struct {
			name   string
			cookie *http.Cookie
		}{{"anonymous", nil}, {"viewer", sessionCookieForUser(t, cfg, "viewer", config.RoleViewer)}, {"editor", sessionCookieForUser(t, cfg, "editor", config.RoleEditor)}} {
			t.Run(user.name+map[bool]string{false: "/public", true: "/private"}[private], func(t *testing.T) {
				before := apiIdentitySnapshot(t, cfg)
				for _, operation := range []string{"source", "save", "create", "delete", "move source", "move target", "versions", "version read", "version restore", "version delete", "comments", "comment add", "comment delete", "link add", "link edit", "link delete"} {
					response := apiIdentityRequest(t, cfg, operation, alias, user.cookie)
					if response.Code < 400 || response.Code >= 500 || strings.Contains(response.Body.String(), marker) {
						t.Fatalf("%s via %q: status=%d marker_disclosed=%v", operation, alias, response.Code, strings.Contains(response.Body.String(), marker))
					}
					if after := apiIdentitySnapshot(t, cfg); !reflect.DeepEqual(before, after) {
						t.Fatalf("denied %s changed storage", operation)
					}
				}
			})
		}
	}
}

func testDocumentAPIAliasAuthorized(t *testing.T, cfg *config.Config, alias, canonical, marker string) {
	t.Helper()
	cfg.Wiki.Private = false
	member := sessionCookieForUserWithGroups(t, cfg, "member", config.RoleEditor, []string{"finance-team"})
	admin := sessionCookieForUser(t, cfg, "admin", config.RoleAdmin)
	for _, operation := range []string{"source", "versions", "version read", "comments", "comment add", "save", "version restore", "create", "comment delete", "version delete"} {
		cookie := member
		if operation == "comment delete" {
			cookie = admin
		}
		response := apiIdentityRequest(t, cfg, operation, alias, cookie)
		if response.Code != http.StatusOK {
			t.Fatalf("authorized %s via %q: %d %s", operation, alias, response.Code, response.Body.String())
		}
		if operation == "source" && !strings.Contains(response.Body.String(), marker) {
			t.Fatal("source omitted canonical content")
		}
		if operation == "version read" && !strings.Contains(response.Body.String(), "VERSION"+marker) {
			t.Fatal("alias selected a separate version namespace")
		}
		if operation == "comments" && !strings.Contains(response.Body.String(), "COMMENT"+marker) {
			t.Fatal("alias selected a separate comment namespace")
		}
	}
	actual := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, filepath.FromSlash(canonical))
	content, err := os.ReadFile(filepath.Join(actual, "document.md"))
	if err != nil || string(content) != "VERSION"+marker {
		t.Fatalf("restore did not update canonical document: %v", err)
	}
	if _, err := os.Stat(filepath.Join(actual, "new", "document.md")); err != nil {
		t.Fatalf("new document missed canonical destination: %v", err)
	}
	list, err := comments.GetComments(filepath.Join(cfg.Wiki.RootDir, "comments"), "/"+canonical)
	if err != nil || len(list) != 1 || list[0].Content != "ADDED_COMMENT" {
		t.Fatalf("comment alias storage differs: %+v %v", list, err)
	}
	response := apiIdentityRequest(t, cfg, "move source", alias, member)
	if response.Code != http.StatusOK {
		t.Fatalf("authorized move: %d %s", response.Code, response.Body.String())
	}
	for _, filename := range []string{"knowledge/public/moved/document.md", "versions/documents/public/moved", "comments/public/moved"} {
		if _, err := os.Stat(filepath.Join(cfg.Wiki.RootDir, filepath.FromSlash(filename))); err != nil {
			t.Fatalf("move did not carry %s: %v", filename, err)
		}
	}
	response = apiIdentityRequest(t, cfg, "delete", "public/moved", member)
	if response.Code != http.StatusOK {
		t.Fatalf("authorized delete: %d %s", response.Code, response.Body.String())
	}
	for _, filename := range []string{"knowledge/public/moved", "versions/documents/public/moved", "comments/public/moved"} {
		if _, err := os.Stat(filepath.Join(cfg.Wiki.RootDir, filepath.FromSlash(filename))); !os.IsNotExist(err) {
			t.Fatalf("delete left %s: %v", filename, err)
		}
	}
}

func TestDocumentAPIIdentityControls(t *testing.T) {
	for _, canonical := range []string{"finance/q3", "finance-quarterly-2026"} {
		t.Run(canonical, func(t *testing.T) {
			cfg, marker := documentAPIFixture(t, canonical)
			testDocumentAPIAliasDenial(t, cfg, canonical, marker)
			testDocumentAPIAliasAuthorized(t, cfg, canonical, canonical, marker)
		})
	}
}

func TestDocumentAPIIdentityLinksAuthorized(t *testing.T) {
	cfg, marker := documentAPIFixture(t, "finance/q3")
	writeAPIIdentityFile(t, filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, "finance", "q3", "document.md"), "---\nlayout: links\n---\n\n# "+marker+"\n\n## Links\n")
	member := sessionCookieForUserWithGroups(t, cfg, "member", config.RoleEditor, []string{"finance-team"})
	for _, operation := range []string{"link add", "link edit", "link delete"} {
		response := apiIdentityRequest(t, cfg, operation, "finance/q3", member)
		if response.Code != http.StatusOK {
			t.Fatalf("authorized %s = %d %s", operation, response.Code, response.Body.String())
		}
	}
}

func assertAPIIdentitySnapshot(t *testing.T, cfg *config.Config, before map[string]string) {
	t.Helper()
	if !reflect.DeepEqual(before, apiIdentitySnapshot(t, cfg)) {
		t.Fatal("denied operation changed document or metadata storage")
	}
}

func TestDocumentAPIIdentityParentAndDescendantPolicies(t *testing.T) {
	cfg, _ := documentAPIFixture(t, "finance/q3")
	cfg.AccessRules = append([]config.AccessRule{{Pattern: "/finance/q3/new", Access: "public"}}, cfg.AccessRules...)
	editor := sessionCookieForUser(t, cfg, "editor", config.RoleEditor)
	before := apiIdentitySnapshot(t, cfg)
	if response := apiIdentityRequest(t, cfg, "create", "finance/q3", editor); response.Code != http.StatusForbidden {
		t.Fatalf("public destination bypassed restricted parent: %d", response.Code)
	}
	// A public category containing a separately restricted child must not be
	// recursively deleted or moved by an outsider.
	writeAPIIdentityFile(t, filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, "public", "secret", "document.md"), "CHILD_SECRET")
	cfg.AccessRules = append([]config.AccessRule{{Pattern: "/public/secret/**", Access: "restricted", Groups: []string{"finance-team"}}}, cfg.AccessRules...)
	before = apiIdentitySnapshot(t, cfg)
	for _, operation := range []string{"delete", "move source"} {
		if response := apiIdentityRequest(t, cfg, operation, "public", editor); response.Code != http.StatusForbidden {
			t.Fatalf("restricted descendant %s = %d", operation, response.Code)
		}
	}
	assertAPIIdentitySnapshot(t, cfg, before)
}

func TestDocumentAPIIdentityLinksAndMetadata(t *testing.T) {
	for _, kind := range []string{"document", "versions", "comments"} {
		t.Run(kind, func(t *testing.T) {
			cfg, marker := documentAPIFixture(t, "finance/q3")
			var link, target string
			switch kind {
			case "document":
				link = filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, "public", "document.md")
				target = filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, "finance", "q3", "document.md")
				if err := os.Remove(link); err != nil {
					t.Fatal(err)
				}
			case "versions":
				link = filepath.Join(cfg.Wiki.RootDir, "versions", "documents", "public")
				target = filepath.Join(cfg.Wiki.RootDir, "versions", "documents", "finance", "q3")
			case "comments":
				link = filepath.Join(cfg.Wiki.RootDir, "comments", "public")
				target = filepath.Join(cfg.Wiki.RootDir, "comments", "finance", "q3")
			}
			if err := os.Symlink(target, link); err != nil {
				if errors.Is(err, os.ErrPermission) || errors.Is(err, errors.ErrUnsupported) || runtime.GOOS == "windows" && errors.Is(err, syscall.Errno(1314)) {
					t.Skipf("symlink creation unavailable: %v", err)
				}
				t.Fatal(err)
			}
			editor := sessionCookieForUser(t, cfg, "editor", config.RoleEditor)
			before := apiIdentitySnapshot(t, cfg)
			operations := []string{"source", "save", "create", "delete", "move source"}
			if kind == "versions" {
				operations = []string{"versions", "version read", "version restore", "version delete", "save", "delete", "move source"}
			}
			if kind == "comments" {
				operations = []string{"comments", "comment add", "delete", "move source"}
			}
			for _, operation := range operations {
				response := apiIdentityRequest(t, cfg, operation, "public", editor)
				if response.Code < 400 || response.Code >= 500 || strings.Contains(response.Body.String(), marker) {
					t.Fatalf("%s via linked %s = %d", operation, kind, response.Code)
				}
			}
			assertAPIIdentitySnapshot(t, cfg, before)
		})
	}
}
