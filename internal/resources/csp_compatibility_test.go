package resources

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

var (
	scriptTagPattern      = regexp.MustCompile(`(?is)<script\b([^>]*)>`)
	scriptSourcePattern   = regexp.MustCompile(`(?i)\bsrc\s*=`)
	eventAttributePattern = regexp.MustCompile(`(?i)\son[a-z0-9_-]+\s*=`)
	dynamicCodePattern    = regexp.MustCompile(`(?m)\beval\s*\(|\bnew\s+Function\s*\(|\b(?:setTimeout|setInterval)\s*\(\s*["']`)
)

func TestTemplatesAndFirstPartyScriptsAreCSPCompatible(t *testing.T) {
	err := fs.WalkDir(templateFiles, "templates", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}

		contents, err := fs.ReadFile(templateFiles, path)
		if err != nil {
			return err
		}
		markup := string(contents)
		for _, match := range scriptTagPattern.FindAllStringSubmatch(markup, -1) {
			if !scriptSourcePattern.MatchString(match[1]) {
				t.Errorf("%s contains an inline script blocked by script-src 'self': %s", path, match[0])
			}
		}
		if match := eventAttributePattern.FindString(markup); match != "" {
			t.Errorf("%s contains an inline event handler blocked by script-src-attr 'none': %s", path, match)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// Vendored libraries are separately version-pinned. This check protects code
	// maintained by Wiki-Go from reintroducing string-compiled JavaScript.
	err = fs.WalkDir(staticFiles, "static/js", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".js") {
			return nil
		}

		contents, err := fs.ReadFile(staticFiles, path)
		if err != nil {
			return err
		}
		if match := dynamicCodePattern.Find(contents); match != nil {
			t.Errorf("%s compiles JavaScript from a string: %q", path, match)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
