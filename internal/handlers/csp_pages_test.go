package handlers

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
	"time"

	"wiki-go/internal/comments"
	"wiki-go/internal/config"
	"wiki-go/internal/safehtml"
	"wiki-go/internal/types"
)

func TestPageTemplatesRenderWithoutInlineJavaScript(t *testing.T) {
	tmpl, err := getTemplate()
	if err != nil {
		t.Fatal(err)
	}
	content, err := safehtml.Execute(template.Must(template.New("test-content").Parse("<p>Safe document</p>")), nil)
	if err != nil {
		t.Fatal(err)
	}

	newPageData := func() *types.PageData {
		cfg := &config.Config{}
		cfg.Wiki.RootDir = t.TempDir()
		cfg.Wiki.Title = "Test Wiki"
		cfg.Wiki.Language = "en"
		cfg.Wiki.Timezone = "UTC"
		return &types.PageData{
			Navigation:      &types.NavTree{Root: &types.NavItem{}},
			Content:         content,
			Config:          cfg,
			CurrentDir:      &types.NavItem{Title: "Security", Path: "/security"},
			LastModified:    time.Unix(1, 0),
			IsAuthenticated: true,
			UserRole:        "admin",
		}
	}

	tests := []struct {
		name      string
		configure func(*types.PageData)
		expected  string
	}{
		{name: "document", configure: func(*types.PageData) {}, expected: "markdown-content"},
		{name: "editing", configure: func(data *types.PageData) { data.IsEditMode = true }, expected: "editor-container"},
		{name: "comments", configure: func(data *types.PageData) {
			data.CommentsAllowed = true
			data.Comments = []comments.Comment{{ID: "1_user.md", Author: "user", FormattedTime: "now"}}
		}, expected: "comments-section"},
		{name: "kanban", configure: func(data *types.PageData) { data.DocumentLayout = "kanban" }, expected: "/static/js/kanban-core.js"},
		{name: "links", configure: func(data *types.PageData) { data.DocumentLayout = "links" }, expected: "/static/js/links.js"},
		{name: "attachments", configure: func(*types.PageData) {}, expected: "file-attachments-section"},
		{name: "error", configure: func(data *types.PageData) {
			data.CurrentDir.Path = `/missing/"><script>alert(1)</script>`
			notFound, executeErr := safehtml.ExecuteTemplate(tmpl, "notfound", data)
			if executeErr != nil {
				t.Fatal(executeErr)
			}
			data.Content = notFound
		}, expected: "/static/js/404.js"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := newPageData()
			test.configure(data)

			var output bytes.Buffer
			if err := tmpl.Execute(&output, data); err != nil {
				t.Fatal(err)
			}
			html := output.String()
			if !strings.Contains(html, test.expected) {
				t.Fatalf("rendered page is missing %q", test.expected)
			}
			assertNoInlineJavaScript(t, html)
		})
	}
}

func TestLoginTemplateUsesExternalScripts(t *testing.T) {
	tmpl, err := getTemplate()
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{}
	cfg.Wiki.Title = `Wiki "><script>alert(1)</script>`
	data := struct {
		Config *config.Config
		Theme  string
	}{Config: cfg, Theme: "light"}

	var output bytes.Buffer
	if err := tmpl.ExecuteTemplate(&output, "login.html", data); err != nil {
		t.Fatal(err)
	}
	html := output.String()
	for _, script := range []string{"/static/js/theme-manager.js", "/static/js/login.js"} {
		if !strings.Contains(html, script) {
			t.Errorf("login page is missing external script %q", script)
		}
	}
	assertNoInlineJavaScript(t, html)
}

func assertNoInlineJavaScript(t *testing.T, html string) {
	t.Helper()
	lower := strings.ToLower(html)
	if strings.Contains(lower, "<script>alert(1)</script>") {
		t.Fatal("injected inline script survived template rendering")
	}
	if strings.Contains(lower, " onclick=") || strings.Contains(lower, " onerror=") || strings.Contains(lower, " onmouseover=") {
		t.Fatal("rendered page contains an inline event handler")
	}

	remainder := html
	for {
		start := strings.Index(strings.ToLower(remainder), "<script")
		if start < 0 {
			return
		}
		end := strings.Index(remainder[start:], ">")
		if end < 0 {
			t.Fatal("rendered page contains an unterminated script tag")
		}
		tag := strings.ToLower(remainder[start : start+end+1])
		if !strings.Contains(tag, " src=") {
			t.Fatalf("rendered page contains inline JavaScript: %s", tag)
		}
		remainder = remainder[start+end+1:]
	}
}
