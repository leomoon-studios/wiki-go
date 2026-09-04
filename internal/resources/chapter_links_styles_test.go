package resources

import (
	"io/fs"
	"strings"
	"testing"
)

func TestChapterLinksPreserveInlineTOCStyles(t *testing.T) {
	markdownStyles, err := fs.ReadFile(staticFiles, "static/css/markdown-extensions.css")
	if err != nil {
		t.Fatal(err)
	}
	markdownCSS := string(markdownStyles)
	for _, expected := range []string{
		"background-color: var(--hover-bg);",
		`:root[data-theme="dark"] .wiki-toc {
    background-color: var(--code-bg);`,
		`:root[data-theme="dark"] .chapter-links .wiki-toc {
    background-color: transparent;`,
		"background-color: #f8f8f8 !important;",
	} {
		if !strings.Contains(markdownCSS, expected) {
			t.Errorf("inline TOC styles omitted %q", expected)
		}
	}

	layoutStyles, err := fs.ReadFile(staticFiles, "static/css/layout.css")
	if err != nil {
		t.Fatal(err)
	}
	layoutCSS := string(layoutStyles)
	for _, expected := range []string{
		`padding: 20px 0 24px;`,
		`.chapter-links-body {
    padding: 0 14px 0 16px;
    background: transparent;
    opacity: 1;
    height: 100%;
    min-height: 0;
    overflow: hidden;
    display: flex;
    flex-direction: column;`,
		`.chapter-links .wiki-toc {
    background: transparent;
    border: none;
    padding: 0;
    margin: 0;
    flex: 1 1 auto;
    min-height: 0;
    overflow-y: auto;`,
		`.chapter-links .toc-list-nested {
    margin: 2px 0 0;
    padding-left: 0.75rem;`,
		`.chapter-links .toc-list a:hover {
    color: var(--primary-color);
    background-color: var(--hover-bg);
    text-decoration: none;`,
		`.chapter-links .toc-list a:focus-visible {`,
		`.chapter-links {
        display: none !important;`,
	} {
		if !strings.Contains(layoutCSS+"\n"+markdownCSS, expected) {
			t.Errorf("chapter-links styles omitted scoped rule %q", expected)
		}
	}
}

func TestMobileNavigationStacksAboveChapterLinks(t *testing.T) {
	navigationStyles, err := fs.ReadFile(staticFiles, "static/css/navigation.css")
	if err != nil {
		t.Fatal(err)
	}
	navigationCSS := string(navigationStyles)
	for _, expected := range []string{
		"z-index: 1101;",
		"z-index: 1100;",
		"z-index: 1090;",
	} {
		if !strings.Contains(navigationCSS, expected) {
			t.Errorf("mobile navigation stacking omitted %q", expected)
		}
	}
}
