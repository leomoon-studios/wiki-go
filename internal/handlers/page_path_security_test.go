package handlers

import (
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"wiki-go/internal/auth"
	"wiki-go/internal/config"
)

func TestPageHandlerDoesNotDiscloseDoubleEncodedTraversalTarget(t *testing.T) {
	testConfig := installUserSessionTestConfig(t, nil)
	testConfig.Wiki.DocumentsDir = "documents"
	canary := "EXTERNAL-DOCUMENT-CANARY"
	externalDocument := filepath.Join(testConfig.Wiki.RootDir, "outside", "document.md")
	if err := os.MkdirAll(filepath.Dir(externalDocument), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(externalDocument, []byte(canary), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		cookie *http.Cookie
	}{
		{name: "anonymous"},
		{name: "viewer", cookie: sessionCookieForUser(t, testConfig, "viewer", config.RoleViewer)},
		{name: "editor", cookie: sessionCookieForUser(t, testConfig, "editor", config.RoleEditor)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/%252e%252e/outside", nil)
			if test.cookie != nil {
				request.AddCookie(test.cookie)
			}
			response := httptest.NewRecorder()

			PageHandler(response, request, testConfig)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; body: %s", response.Code, response.Body.String())
			}
			if strings.Contains(response.Body.String(), canary) {
				t.Fatal("response disclosed the external document canary")
			}
		})
	}
}

func TestPageHandlerDoesNotDiscloseBackslashTraversalTarget(t *testing.T) {
	testConfig := installUserSessionTestConfig(t, nil)
	testConfig.Wiki.DocumentsDir = "documents"
	canary := "EXTERNAL-BACKSLASH-DOCUMENT-CANARY"
	externalDocument := filepath.Join(testConfig.Wiki.RootDir, "outside", "document.md")
	if err := os.MkdirAll(filepath.Dir(externalDocument), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(externalDocument, []byte(canary), 0o644); err != nil {
		t.Fatal(err)
	}

	requests := []struct {
		name   string
		target string
	}{
		{name: "literal backslash", target: `/..\outside`},
		{name: "encoded backslash", target: "/..%5Coutside"},
		{name: "reporter depth encoded backslash", target: "/..%5C..%5C..%5C..%5C..%5Ctmp%5Ccanary_pg"},
	}
	users := []struct {
		name   string
		cookie *http.Cookie
	}{
		{name: "anonymous"},
		{name: "viewer", cookie: sessionCookieForUser(t, testConfig, "viewer", config.RoleViewer)},
		{name: "editor", cookie: sessionCookieForUser(t, testConfig, "editor", config.RoleEditor)},
		{name: "administrator", cookie: sessionCookieForUser(t, testConfig, "admin", config.RoleAdmin)},
	}
	for _, requestTest := range requests {
		for _, user := range users {
			t.Run(requestTest.name+"/"+user.name, func(t *testing.T) {
				request := httptest.NewRequest(http.MethodGet, requestTest.target, nil)
				if user.cookie != nil {
					request.AddCookie(user.cookie)
				}
				response := httptest.NewRecorder()
				PageHandler(response, request, testConfig)

				if response.Code != http.StatusBadRequest {
					t.Fatalf("status = %d, want 400; body: %s", response.Code, response.Body.String())
				}
				if strings.Contains(response.Body.String(), canary) {
					t.Fatal("response disclosed the external document canary")
				}
			})
		}
	}

	content, err := os.ReadFile(externalDocument)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != canary {
		t.Fatalf("external document changed to %q", content)
	}
}

func TestRestrictedDocumentAccessAllowsFinanceGroupAndAdministrators(t *testing.T) {
	testConfig := installUserSessionTestConfig(t, nil)
	testConfig.Wiki.DocumentsDir = "documents"
	testConfig.Wiki.Title = "Security Test Wiki"
	testConfig.Wiki.Owner = "wiki.example"
	testConfig.Wiki.Timezone = "UTC"
	testConfig.Wiki.Language = "en"
	testConfig.Wiki.DisableComments = true
	testConfig.Wiki.HideAttachments = true
	testConfig.AccessRules = []config.AccessRule{{
		Pattern: "/finance/**",
		Access:  "restricted",
		Groups:  []string{"finance"},
	}}
	financeDocument := filepath.Join(testConfig.Wiki.RootDir, "documents", "finance", "document.md")
	if err := os.MkdirAll(filepath.Dir(financeDocument), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(financeDocument, []byte("# Authorized finance content"), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		cookie *http.Cookie
	}{
		{
			name:   "finance group editor",
			cookie: sessionCookieForUserWithGroups(t, testConfig, "finance-editor", config.RoleEditor, []string{"finance"}),
		},
		{
			name:   "administrator",
			cookie: sessionCookieForUser(t, testConfig, "admin", config.RoleAdmin),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pageRequest := httptest.NewRequest(http.MethodGet, "/finance", nil)
			pageRequest.AddCookie(test.cookie)
			pageResponse := httptest.NewRecorder()
			PageHandler(pageResponse, pageRequest, testConfig)
			if pageResponse.Code != http.StatusOK {
				t.Fatalf("page status = %d, want 200; body: %s", pageResponse.Code, pageResponse.Body.String())
			}
			if !strings.Contains(pageResponse.Body.String(), "Authorized finance content") {
				t.Fatal("page response omitted authorized finance content")
			}

			sourceRequest := httptest.NewRequest(http.MethodGet, "/api/source/finance", nil)
			sourceRequest.AddCookie(test.cookie)
			sourceResponse := httptest.NewRecorder()
			SourceHandler(sourceResponse, sourceRequest)
			if sourceResponse.Code != http.StatusOK {
				t.Fatalf("source status = %d, want 200; body: %s", sourceResponse.Code, sourceResponse.Body.String())
			}
			if sourceResponse.Body.String() != "# Authorized finance content" {
				t.Fatalf("source body = %q, want authorized finance content", sourceResponse.Body.String())
			}
		})
	}
}

func sessionCookieForUserWithGroups(t *testing.T, testConfig *config.Config, username, role string, groups []string) *http.Cookie {
	t.Helper()
	response := httptest.NewRecorder()
	if err := auth.CreateSession(response, username, role, groups, false, testConfig); err != nil {
		t.Fatal(err)
	}
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == "session_token" {
			return cookie
		}
	}
	t.Fatal("session token cookie was not created")
	return nil
}

// financeDisclosureFixture supplies distinct markers so a generic 200 response
// cannot be mistaken for a successful document or attachment read.
func financeDisclosureFixture(t *testing.T) (*config.Config, string, string) {
	t.Helper()
	testConfig := installUserSessionTestConfig(t, nil)
	testConfig.Wiki.DocumentsDir = "documents"
	testConfig.Wiki.Title = "Security Test Wiki"
	testConfig.Wiki.Owner = "wiki.example"
	testConfig.Wiki.Timezone = "UTC"
	testConfig.Wiki.Language = "en"
	testConfig.Wiki.DisableComments = true
	testConfig.Wiki.HideAttachments = true
	testConfig.AccessRules = []config.AccessRule{
		{Pattern: "/finance/**", Access: "restricted", Groups: []string{"finance-team"}},
		{Pattern: "/finance-quarterly-2026/**", Access: "restricted", Groups: []string{"finance-team"}},
	}
	documentMarker := "DOCUMENT" + rand.Text()
	attachmentMarker := "ATTACHMENT" + rand.Text()
	for name, body := range map[string]string{
		"finance/q3/document.md":             documentMarker,
		"finance-quarterly-2026/document.md": documentMarker,
		"finance/att.txt":                    attachmentMarker,
		"finance/q3/att.txt":                 attachmentMarker,
		"finance-quarterly-2026/att.txt":     attachmentMarker,
		"public/document.md":                 "PUBLIC" + documentMarker,
		"public/att.txt":                     "PUBLIC" + attachmentMarker,
	} {
		filename := filepath.Join(testConfig.Wiki.RootDir, testConfig.Wiki.DocumentsDir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filename, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return testConfig, documentMarker, attachmentMarker
}

func financeDisclosureRequest(testConfig *config.Config, target string, cookie *http.Cookie) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, target, nil)
	if cookie != nil {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	if strings.HasPrefix(target, "/api/files/") {
		ServeFileHandler(response, request, testConfig)
	} else {
		PageHandler(response, request, testConfig)
	}
	return response
}

func assertFinanceRefused(t *testing.T, response *httptest.ResponseRecorder, marker string) {
	t.Helper()
	disclosed := strings.Contains(response.Body.String(), marker)
	t.Logf("status=%d marker_disclosed=%v", response.Code, disclosed)
	if disclosed {
		t.Error("response disclosed restricted content marker")
	}
	switch response.Code {
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusFound, http.StatusSeeOther:
	default:
		t.Errorf("status = %d, want refusal", response.Code)
	}
}

func TestFinanceDisclosureControls(t *testing.T) {
	testConfig, documentMarker, attachmentMarker := financeDisclosureFixture(t)
	member := sessionCookieForUserWithGroups(t, testConfig, "member", config.RoleViewer, []string{"finance-team"})
	admin := sessionCookieForUser(t, testConfig, "administrator", config.RoleAdmin)
	outsider := sessionCookieForUser(t, testConfig, "outsider", config.RoleViewer)
	for _, private := range []bool{false, true} {
		name := "public"
		if private {
			name = "private"
		}
		t.Run(name, func(t *testing.T) {
			testConfig.Wiki.Private = private
			for _, user := range []struct {
				name       string
				cookie     *http.Cookie
				restricted bool
			}{
				{"anonymous", nil, false}, {"outsider", outsider, false}, {"member", member, true}, {"administrator", admin, true},
			} {
				t.Run(user.name, func(t *testing.T) {
					resources := []struct{ target, marker string }{
						{"/finance/q3", documentMarker}, {"/finance-quarterly-2026", documentMarker},
						{"/api/files/finance/att.txt", attachmentMarker},
					}
					// Denial of the nested attachment has its own regression below;
					// keep its permitted requests here as a reachability control.
					if user.restricted {
						resources = append(resources, struct{ target, marker string }{"/api/files/finance/q3/att.txt", attachmentMarker})
					}
					for _, resource := range resources {
						t.Run(resource.target, func(t *testing.T) {
							response := financeDisclosureRequest(testConfig, resource.target, user.cookie)
							if !user.restricted {
								assertFinanceRefused(t, response, resource.marker)
								return
							}
							if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), resource.marker) {
								t.Fatalf("authorized response status=%d, marker_present=%v", response.Code, strings.Contains(response.Body.String(), resource.marker))
							}
						})
					}
					for _, resource := range []struct{ target, marker string }{
						{"/public", documentMarker},
						{"/api/files/public/att.txt", attachmentMarker},
					} {
						t.Run(resource.target, func(t *testing.T) {
							response := financeDisclosureRequest(testConfig, resource.target, user.cookie)
							if private && user.cookie == nil {
								assertFinanceRefused(t, response, resource.marker)
							} else if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "PUBLIC"+resource.marker) {
								t.Fatalf("public control status=%d, marker_present=%v", response.Code, strings.Contains(response.Body.String(), "PUBLIC"+resource.marker))
							}
						})
					}
				})
			}
		})
	}
}

// This is deliberately separate from the passing controls: filepath.Dir in the
// unfixed Windows handler produces an owner path that misses slash-based rules.
func TestFinanceNestedAttachmentDenied(t *testing.T) {
	testConfig, _, marker := financeDisclosureFixture(t)
	outsider := sessionCookieForUser(t, testConfig, "outsider", config.RoleViewer)
	for _, private := range []bool{false, true} {
		name := "public"
		if private {
			name = "private"
		}
		t.Run(name, func(t *testing.T) {
			testConfig.Wiki.Private = private
			for _, user := range []struct {
				name   string
				cookie *http.Cookie
			}{{"anonymous", nil}, {"outsider", outsider}} {
				t.Run(user.name, func(t *testing.T) {
					assertFinanceRefused(t, financeDisclosureRequest(testConfig, "/api/files/finance/q3/att.txt", user.cookie), marker)
				})
			}
		})
	}
}

func TestFinanceDistinctCaseSensitivePaths(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux case-sensitive filesystem control")
	}
	testConfig, documentMarker, _ := financeDisclosureFixture(t)
	root := filepath.Join(testConfig.Wiki.RootDir, testConfig.Wiki.DocumentsDir)
	alias := filepath.Join(root, "Finance", "q3")
	if err := os.MkdirAll(alias, 0o755); err != nil {
		t.Fatal(err)
	}
	lowerInfo, err := os.Stat(filepath.Join(root, "finance", "q3"))
	if err != nil {
		t.Fatal(err)
	}
	upperInfo, err := os.Stat(alias)
	if err != nil {
		t.Fatal(err)
	}
	if os.SameFile(lowerInfo, upperInfo) {
		t.Skip("filesystem does not support distinct case-sensitive directories")
	}
	publicMarker := "DISTINCT" + rand.Text()
	if err := os.WriteFile(filepath.Join(alias, "document.md"), []byte(publicMarker), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Run("restricted lowercase", func(t *testing.T) {
		assertFinanceRefused(t, financeDisclosureRequest(testConfig, "/finance/q3", nil), documentMarker)
	})
	t.Run("distinct uppercase", func(t *testing.T) {
		response := financeDisclosureRequest(testConfig, "/Finance/q3", nil)
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), publicMarker) || strings.Contains(response.Body.String(), documentMarker) {
			t.Fatal("distinct case-sensitive document was not served independently")
		}
	})
}
