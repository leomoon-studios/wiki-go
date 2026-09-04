package resources

import (
	"encoding/json"
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
		`width: min(var(--chapter-links-width), 85vw);`,
		`padding: 20px 0 24px;`,
		`.chapter-links-body {
    padding-block: 0;
    padding-inline: 14px;
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
    padding-inline-start: 0.75rem;`,
		`.chapter-links-toggle-text {
    display: block;
    width: 8px;
    height: 8px;
    border-top: 2px solid currentColor;
    border-right: 2px solid currentColor;
    transform: rotate(45deg);`,
		`transform: rotate(225deg);`,
		`.chapter-links-toggle:focus-visible {`,
		`.chapter-links .toc-list a:hover {
    color: var(--primary-color);
    background-color: var(--hover-bg);
    text-decoration: none;`,
		`.chapter-links .toc-list a:focus-visible {`,
		`.chapter-links,
    .chapter-links-toggle {
        display: none !important;`,
	} {
		if !strings.Contains(layoutCSS+"\n"+markdownCSS, expected) {
			t.Errorf("chapter-links styles omitted scoped rule %q", expected)
		}
	}
}

func TestChapterLinksControlsUseEnglishTranslationsAndAccessibleState(t *testing.T) {
	englishFile, err := fs.ReadFile(languageFiles, "langs/en.json")
	if err != nil {
		t.Fatal(err)
	}
	var english map[string]string
	if err := json.Unmarshal(englishFile, &english); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{
		"chapter_links.on_this_page": "On this page",
		"chapter_links.expand":       "Expand chapter links",
		"chapter_links.retract":      "Retract chapter links",
	} {
		if got := english[key]; got != want {
			t.Errorf("English translation %q: want %q, got %q", key, want, got)
		}
	}

	templateFile, err := fs.ReadFile(templateFiles, "templates/chapter-links.html")
	if err != nil {
		t.Fatal(err)
	}
	markup := string(templateFile)
	for _, expected := range []string{
		`{{t "chapter_links.on_this_page"}}`,
		`data-label-expand="{{t "chapter_links.expand"}}"`,
		`data-label-retract="{{t "chapter_links.retract"}}"`,
		`type="button"`,
		`aria-expanded="false"`,
		`aria-hidden="true"`,
		` inert>`,
	} {
		if !strings.Contains(markup, expected) {
			t.Errorf("chapter-links template omitted %q", expected)
		}
	}
	for _, forbidden := range []string{"On this page", "Expand chapter links", "Retract chapter links", "&lt;&lt;", "&gt;&gt;"} {
		if strings.Contains(markup, forbidden) {
			t.Errorf("chapter-links template hard-coded %q", forbidden)
		}
	}

	controllerFile, err := fs.ReadFile(staticFiles, "static/js/markdown-extensions.js")
	if err != nil {
		t.Fatal(err)
	}
	controller := string(controllerFile)
	for _, expected := range []string{
		"chapterLinksToggle.dataset.labelExpand",
		"chapterLinksToggle.dataset.labelRetract",
		"chapterLinksToggle.setAttribute('aria-expanded', String(!isRetracted))",
		"chapterLinksPanel.setAttribute('aria-hidden', String(isRetracted))",
		"chapterLinksPanel.toggleAttribute('inert', isRetracted)",
		"chapterLinksPanel.contains(document.activeElement)",
		"chapterLinksToggle.focus()",
	} {
		if !strings.Contains(controller, expected) {
			t.Errorf("chapter-links controller omitted %q", expected)
		}
	}
	for _, forbidden := range []string{"'Expand chapter links'", "'Retract chapter links'", "'<<'", "'>>'"} {
		if strings.Contains(controller, forbidden) {
			t.Errorf("chapter-links controller hard-coded %q", forbidden)
		}
	}
}

func TestEveryLocaleHasChapterLinksTranslations(t *testing.T) {
	entries, err := fs.ReadDir(languageFiles, "langs")
	if err != nil {
		t.Fatal(err)
	}
	keys := []string{
		"chapter_links.on_this_page",
		"chapter_links.expand",
		"chapter_links.retract",
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		contents, err := fs.ReadFile(languageFiles, "langs/"+entry.Name())
		if err != nil {
			t.Fatal(err)
		}
		var translations map[string]string
		if err := json.Unmarshal(contents, &translations); err != nil {
			t.Fatalf("parse %s: %v", entry.Name(), err)
		}
		for _, key := range keys {
			if strings.TrimSpace(translations[key]) == "" {
				t.Errorf("%s is missing translation %q", entry.Name(), key)
			}
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
