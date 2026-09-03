package resources

import (
	"io/fs"
	"strings"
	"testing"
)

func TestMissingLinkStylesCoverEveryMarkdownPreview(t *testing.T) {
	stylesheet, err := fs.ReadFile(staticFiles, "static/css/typography.css")
	if err != nil {
		t.Fatal(err)
	}

	styles := string(stylesheet)
	for _, selector := range []string{
		".markdown-content span.notfound a",
		".editor-preview span.notfound a",
		".version-content span.notfound a",
	} {
		if !strings.Contains(styles, selector) {
			t.Errorf("missing-link styling does not cover %s", selector)
		}
	}
}
