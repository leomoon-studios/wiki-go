package routes

import (
	"bytes"
	"crypto/rand"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"wiki-go/internal/config"
)

type windowsPathHTTPResponse struct {
	status int
	body   string
}

func windowsPathHTTPRequest(t *testing.T, client *http.Client, serverURL, method, target, body, contentType string, cookie *http.Cookie) windowsPathHTTPResponse {
	t.Helper()
	request, err := http.NewRequest(method, serverURL+target, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if cookie != nil {
		request.AddCookie(cookie)
	}
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	contents, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	return windowsPathHTTPResponse{response.StatusCode, string(contents)}
}

func windowsPathUploadBody(t *testing.T, document, filename, contents string) (string, string) {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("docPath", document); err != nil {
		t.Fatal(err)
	}
	file, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(file, contents); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return body.String(), writer.FormDataContentType()
}

// Exercise the advisory through a live registered router. The client keeps
// redirects visible, so a login response cannot be mistaken for a document.
func TestWindowsPathAuthorizationHTTP(t *testing.T) {
	for _, private := range []bool{false, true} {
		name := "public"
		if private {
			name = "private"
		}
		t.Run(name, func(t *testing.T) {
			cfg, documentMarker, publicMarker, _ := pageIdentityFixture(t)
			cfg.Wiki.Private = private
			cfg.Wiki.MaxUploadSize = 10
			cfg.Wiki.DisableFileUploadChecking = true
			cfg.AccessRules = append(cfg.AccessRules, config.AccessRule{Pattern: "/**", Access: "restricted", Groups: []string{"finance"}})
			attachmentMarker, publicAttachment := "ATTACHMENT"+rand.Text(), "PUBLIC_ATTACHMENT"+rand.Text()
			for relative, contents := range map[string]string{
				"knowledge/finance/q3/att.txt":             attachmentMarker,
				"knowledge/finance-quarterly-2026/att.txt": attachmentMarker,
				"knowledge/public/att.txt":                 publicAttachment,
				"pages/home/att.txt":                       attachmentMarker,
			} {
				filename := filepath.Join(cfg.Wiki.RootDir, filepath.FromSlash(relative))
				if err := os.WriteFile(filename, []byte(contents), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			users := []struct {
				name   string
				cookie *http.Cookie
				member bool
			}{
				{"anonymous", nil, false},
				{"outsider viewer", routeSessionCookie(t, cfg, "viewer", config.RoleViewer, nil), false},
				{"outsider editor", routeSessionCookie(t, cfg, "editor", config.RoleEditor, nil), false},
				{"member", routeSessionCookie(t, cfg, "member", config.RoleEditor, []string{"finance"}), true},
				{"administrator", routeSessionCookie(t, cfg, "admin", config.RoleAdmin, nil), true},
			}
			server := httptest.NewServer(http.DefaultServeMux)
			defer server.Close()
			client := server.Client()
			client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
			type resource struct {
				target, marker    string
				restricted, valid bool
			}
			check := func(t *testing.T, resources []resource) {
				t.Helper()
				for _, user := range users {
					for _, file := range resources {
						response := windowsPathHTTPRequest(t, client, server.URL, http.MethodGet, file.target, "", "", user.cookie)
						allowed := file.valid && (!private || user.cookie != nil) && (!file.restricted || user.member)
						present := strings.Contains(response.body, file.marker)
						t.Logf("user=%q target=%q status=%d marker_present=%v", user.name, file.target, response.status, present)
						if allowed {
							if response.status != http.StatusOK || !present {
								t.Fatalf("authorized HTTP request failed: user=%s path=%s status=%d", user.name, file.target, response.status)
							}
						} else {
							if present {
								t.Fatalf("unauthorized HTTP request disclosed a marker: %s %s", user.name, file.target)
							}
							switch response.status {
							case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusFound, http.StatusSeeOther:
							default:
								t.Fatalf("refused HTTP request status=%d: %s", response.status, file.target)
							}
						}
					}
				}
			}
			check(t, []resource{
				{"/finance/q3", documentMarker, true, true},
				{"/Finance/Q3", documentMarker, true, true},
				{"/finance./q3", documentMarker, true, false},
				{"/finance%20/q3", documentMarker, true, false},
				{"/finance-quarterly-2026", documentMarker, true, true},
				{"/api/files/finance/q3/att.txt", attachmentMarker, true, true},
				{"/api/files/Finance/Q3/ATT.TXT", attachmentMarker, true, true},
				{"/api/files/finance%5cq3/att.txt", attachmentMarker, true, true},
				{"/api/files/finance./q3/att.txt", attachmentMarker, true, false},
				{"/api/files/finance%20/q3/att.txt", attachmentMarker, true, false},
				{"/api/files/finance-quarterly-2026/att.txt", attachmentMarker, true, true},
				{"/api/files/pages/home/att.txt", attachmentMarker, true, true},
				{"/api/files/PaGeS/HoMe/ATT.TXT", attachmentMarker, true, true},
				{"/api/files/FINANCE/Q3/DOCUMENT.MD", documentMarker, true, false},
				{"/PUBLIC", publicMarker, false, true},
				{"/api/files/PUBLIC/ATT.TXT", publicAttachment, false, true},
			})
			t.Run("8.3 alias", func(t *testing.T) {
				short := pageIdentityShortDirectory(t, filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, "finance-quarterly-2026"))
				check(t, []resource{{"/" + short, documentMarker, true, true}, {"/api/files/" + short + "/att.txt", attachmentMarker, true, true}})
			})
			// An outsider with the editor role still cannot mutate a restricted
			// resource, while a group member can use the same alias through APIs.
			outsider, member := users[2].cookie, users[3].cookie
			for _, cookie := range []*http.Cookie{outsider, member} {
				allowed := cookie == member
				body, contentType := windowsPathUploadBody(t, "Finance/Q3", "uploaded.txt", "UPLOADED_HTTP_ATTACHMENT")
				renameSource, deleteTarget := "uploaded.txt", "renamed.txt"
				if !allowed {
					renameSource, deleteTarget = "att.txt", "att.txt"
				}
				for _, operation := range []struct{ method, target, body, contentType string }{
					{http.MethodPost, "/api/save/Finance/Q3", "# UPDATED_HTTP_DOCUMENT", "text/plain"},
					{http.MethodPost, "/api/files/upload", body, contentType},
					{http.MethodPost, "/api/files/rename", `{"currentPath":"Finance/Q3/` + renameSource + `","newName":"renamed.txt"}`, "application/json"},
					{http.MethodDelete, "/api/files/delete/Finance/Q3/" + deleteTarget, "", ""},
				} {
					response := windowsPathHTTPRequest(t, client, server.URL, operation.method, operation.target, operation.body, operation.contentType, cookie)
					t.Logf("mutation=%s target=%q status=%d allowed=%v", operation.method, operation.target, response.status, allowed)
					if allowed && response.status != http.StatusOK || !allowed && (response.status < 400 || response.status >= 500) {
						t.Fatalf("HTTP mutation %s %s status=%d allowed=%v", operation.method, operation.target, response.status, allowed)
					}
				}
				original := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, "finance", "q3", "document.md")
				contents, err := os.ReadFile(original)
				want := "# " + documentMarker
				if allowed {
					want = "# UPDATED_HTTP_DOCUMENT"
				}
				if err != nil || string(contents) != want {
					t.Fatalf("HTTP save changed the wrong document: %v", err)
				}
				for _, filename := range []string{"uploaded.txt", "renamed.txt"} {
					if _, err := os.Stat(filepath.Join(filepath.Dir(original), filename)); !os.IsNotExist(err) {
						t.Fatalf("HTTP mutation left %s: %v", filename, err)
					}
				}
				if !allowed {
					contents, err := os.ReadFile(filepath.Join(filepath.Dir(original), "att.txt"))
					if err != nil || string(contents) != attachmentMarker {
						t.Fatalf("denied mutation changed attachment: %v", err)
					}
				}
			}
			response := windowsPathHTTPRequest(t, client, server.URL, http.MethodGet, "/api/source/finance/q3", "", "", member)
			if response.status != http.StatusOK || response.body != "# UPDATED_HTTP_DOCUMENT" {
				t.Fatal("HTTP canonical source did not observe the alias save")
			}
		})
	}
}
