package safehtml

import (
	"go/ast"
	"go/parser"
	"go/token"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestExecuteContextuallyEscapesUserValues(t *testing.T) {
	tmpl := template.Must(template.New("payloads").Parse(
		`<a href="{{.Path}}" title="{{.Title}}">{{.Title}}</a><p>{{.Error}}</p>`,
	))
	data := struct {
		Path  string
		Title string
		Error string
	}{
		Path:  `javascript:alert(1)`,
		Title: `"><img src=x onerror="alert(2)">`,
		Error: `<svg onload="alert(3)"></svg>`,
	}

	got, err := Execute(tmpl, data)
	if err != nil {
		t.Fatal(err)
	}
	output := string(got)
	for _, forbidden := range []string{`href="javascript:`, `<img`, `<svg`, `onerror="alert`, `onload="alert`} {
		if strings.Contains(strings.ToLower(output), forbidden) {
			t.Fatalf("trusted template output retained active payload %q: %s", forbidden, output)
		}
	}
	for _, want := range []string{`href="#ZgotmplZ"`, `&lt;img`, `&lt;svg`} {
		if !strings.Contains(output, want) {
			t.Errorf("trusted template output omitted escaped value %q: %s", want, output)
		}
	}
}

func TestTrustedHTMLConversionsStayInsidePackage(t *testing.T) {
	repoRoot := repositoryRoot(t)
	approvedFile := filepath.Join(repoRoot, "internal", "safehtml", "trusted.go")

	err := filepath.WalkDir(repoRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == "vendor") {
			return filepath.SkipDir
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}

		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}

		templateAliases := map[string]bool{}
		dotTemplateImport := false
		for _, imported := range parsed.Imports {
			if strings.Trim(imported.Path.Value, `"`) != "html/template" {
				continue
			}
			alias := "template"
			if imported.Name != nil {
				alias = imported.Name.Name
			}
			if alias == "." {
				dotTemplateImport = true
				continue
			}
			templateAliases[alias] = true
		}

		ast.Inspect(parsed, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			if identifier, ok := call.Fun.(*ast.Ident); ok {
				if identifier.Name == "WithUnsafe" {
					t.Errorf("unapproved html.WithUnsafe call in %s", relativePath(repoRoot, path))
				}
				if identifier.Name == "HTML" && dotTemplateImport && path != approvedFile {
					t.Errorf("unapproved template.HTML conversion in %s; use internal/safehtml", relativePath(repoRoot, path))
				}
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if selector.Sel.Name == "WithUnsafe" {
				t.Errorf("unapproved html.WithUnsafe call in %s", relativePath(repoRoot, path))
			}
			packageName, ok := selector.X.(*ast.Ident)
			if ok && selector.Sel.Name == "HTML" && templateAliases[packageName.Name] && path != approvedFile {
				t.Errorf("unapproved template.HTML conversion in %s; use internal/safehtml", relativePath(repoRoot, path))
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestHighRiskJavaScriptRenderersDoNotParseAPIValuesAsHTML(t *testing.T) {
	repoRoot := repositoryRoot(t)
	javascriptDir := filepath.Join(repoRoot, "internal", "resources", "static", "js")
	filesWithNoHTMLAssignments := []string{
		"access-rules-manager.js",
		"backup-manager.js",
		"comments.js",
		"document-management.js",
		"file-utilities.js",
		"import-manager.js",
		"search.js",
		"settings-manager.js",
	}
	for _, name := range filesWithNoHTMLAssignments {
		content, err := os.ReadFile(filepath.Join(javascriptDir, name))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(content), "innerHTML =") {
			t.Errorf("%s assigns innerHTML; render API and user values with DOM nodes", name)
		}
	}

	versionHistory, err := os.ReadFile(filepath.Join(javascriptDir, "version-history.js"))
	if err != nil {
		t.Fatal(err)
	}
	versionSource := string(versionHistory)
	if strings.Count(versionSource, "innerHTML =") != 1 || !strings.Contains(versionSource, "versionContent.innerHTML = renderedHTML") {
		t.Errorf("version-history.js must reserve its only innerHTML assignment for safe server-rendered Markdown")
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate security policy test")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

func relativePath(root, path string) string {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(relative)
}
