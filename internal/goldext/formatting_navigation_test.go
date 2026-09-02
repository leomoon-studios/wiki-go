package goldext

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldhtml "github.com/yuin/goldmark/renderer/html"
)

func TestTrustedInlineFormattingIsSafe(t *testing.T) {
	input := "==highlight== H^2^O h~2~o ~~gone~~ ==<img src=x onerror=alert(1)>== ^<img src=x onerror=alert(2)>^ ~<svg onload=alert(3)>~ `==code==` $x^2^$"
	got := renderStep5Markdown(t, input)

	for _, want := range []string{
		"<mark>highlight</mark>",
		"H<sup>2</sup>O",
		"h<sub>2</sub>o",
		"<del>gone</del>",
		"<mark>&lt;img src=x onerror=alert(1)&gt;</mark>",
		"<sup>&lt;img src=x onerror=alert(2)&gt;</sup>",
		"<sub>&lt;svg onload=alert(3)&gt;</sub>",
		"<code>==code==</code>",
		"$x^2^$",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("formatted output omitted %q: %s", want, got)
		}
	}
	if strings.Contains(strings.ToLower(got), "<img") || strings.Contains(strings.ToLower(got), "<svg") {
		t.Fatalf("inline formatting emitted active HTML: %s", got)
	}

	mathAndCode := renderStep5Markdown(t, "```\n$$\n```\n\n==after code==\n\n$$\nx^2^\n$$")
	if !strings.Contains(mathAndCode, "<mark>after code</mark>") || !strings.Contains(mathAndCode, "x^2^") {
		t.Fatalf("formatting parser did not preserve code/math boundaries: %s", mathAndCode)
	}
}

func TestDetailsAndAlertsRenderNestedMarkdownSafely(t *testing.T) {
	input := "```details <img src=x onerror=alert(1)>\n**details bold**\n\n<script>alert(2)</script>\n```\n\n> [!WARNING]\n> **alert bold**\n> <img src=x onerror=alert(3)>\n"
	got := renderStep5Markdown(t, input)

	for _, want := range []string{
		`<details class="markdown-details"><summary>&lt;img src=x onerror=alert(1)&gt;</summary>`,
		"<strong>details bold</strong>",
		`<div class="markdown-alert markdown-alert-warning">`,
		"<strong>alert bold</strong>",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("block extension output omitted %q: %s", want, got)
		}
	}
	for _, forbidden := range []string{"<script", "<img", ` onerror="`} {
		if strings.Contains(strings.ToLower(got), forbidden) {
			t.Errorf("block extension output retained active content %q: %s", forbidden, got)
		}
	}
}

func TestTOCAndHeadingAnchorsNormalizeAndEscapeValues(t *testing.T) {
	input := "[toc]\n\n# Hello <img src=x onerror=alert(1)>\n\n## Custom {#Mixed_CASE}\n\n## Custom Again {#mixed_case}\n"
	got := renderStep5Markdown(t, input)

	for _, want := range []string{
		`<nav class="wiki-toc table-of-contents" aria-label="Table of Contents">`,
		`href="#hello-img-srcx-onerroralert1"`,
		`href="#mixed_case"`,
		`href="#mixed_case-1"`,
		`class="heading-anchor"`,
		`aria-label="Permalink"`,
		`&lt;img src=x onerror=alert(1)&gt;`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("navigation output omitted %q: %s", want, got)
		}
	}
	if strings.Contains(strings.ToLower(got), "<img") || strings.Contains(strings.ToLower(got), ` onerror="`) {
		t.Fatalf("navigation output emitted active heading HTML: %s", got)
	}
}

func TestStatsShortcodesEscapeFilesystemDataAndMalformedArguments(t *testing.T) {
	root := t.TempDir()
	ConfigureContentPaths(root, "documents")
	t.Cleanup(func() { ConfigureContentPaths("data", "documents") })

	documentDir := filepath.Join(root, "documents", `bad"><img src=x onerror=alert(1)>`)
	if err := os.MkdirAll(documentDir, 0o755); err != nil {
		t.Fatal(err)
	}
	documentPath := filepath.Join(documentDir, "document.md")
	if err := os.WriteFile(documentPath, []byte("# <img src=x onerror=alert(2)>\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	input := ":::stats recent=5:::\n\n:::stats count=*:::\n\n:::stats count=../private:::\n\n:::stats count=<img src=x onerror=alert(3)>:::\n\nYear :::year:::"
	got := renderStep5Markdown(t, ShortcodesPreprocessor(input, ""))

	if !strings.Contains(got, `class="wiki-stats recent-edits"`) || !strings.Contains(got, `class="wiki-stats doc-count"`) {
		t.Fatalf("valid stats shortcodes did not render: %s", got)
	}
	if !strings.Contains(got, strconv.Itoa(time.Now().Year())) {
		t.Errorf("year shortcode was not replaced with text: %s", got)
	}
	if !strings.Contains(got, ":::stats count=../private:::") || !strings.Contains(got, "&lt;img src=x onerror=alert(3)&gt;") {
		t.Errorf("malformed stats shortcode did not fail closed as text: %s", got)
	}
	if strings.Contains(strings.ToLower(got), "<img") || strings.Contains(strings.ToLower(got), ` onerror="`) {
		t.Fatalf("stats shortcode emitted active filesystem or argument HTML: %s", got)
	}
}

func TestTextOnlyPreprocessorsDoNotGenerateHTML(t *testing.T) {
	wikiLink := WikiLinkPreprocessor(`[[safe|<img src=x onerror=alert(1)>]] [[javascript:alert(2)|bad]]`, "")
	if strings.Contains(wikiLink, "<img") || !strings.Contains(wikiLink, "&lt;img") {
		t.Fatalf("wiki-link preprocessing retained active label HTML: %q", wikiLink)
	}
	if strings.Contains(wikiLink, "](javascript:") {
		t.Fatalf("wiki-link preprocessing created a dangerous URL: %q", wikiLink)
	}

	attachment := LinkPreprocessor("[file](guide.pdf) ![image](photo.png)", "docs/start")
	for _, want := range []string{
		"[file](/api/files/docs/start/guide.pdf)",
		"![image](/api/files/docs/start/photo.png)",
	} {
		if !strings.Contains(attachment, want) {
			t.Errorf("attachment preprocessing omitted %q: %s", want, attachment)
		}
	}
	if strings.Contains(attachment, "<") {
		t.Fatalf("attachment preprocessing generated HTML: %q", attachment)
	}
	dangerousLinks := LinkPreprocessor(`[one](javascript:alert(1)) [two](javas&#x63;ript:alert(2)) [three](//evil.example/x)`, "docs/start")
	if strings.Contains(strings.ToLower(dangerousLinks), "javascript:") || strings.Contains(dangerousLinks, "//evil.example") {
		t.Fatalf("link preprocessing retained a dangerous destination: %q", dangerousLinks)
	}
	dangerousLabel := LinkPreprocessor(`[<img src=x onerror=alert(4)>](https://example.com)`, "docs/start")
	if strings.Contains(dangerousLabel, "<img") || !strings.Contains(dangerousLabel, "&lt;img") {
		t.Fatalf("link preprocessing retained active label HTML: %q", dangerousLabel)
	}

	typography := TypographyPreprocessor("(c) ...", "")
	emoji := EmojiPreprocessor(":smile:", "")
	if typography != "© …" || strings.Contains(typography+emoji, "<") {
		t.Fatalf("text preprocessors emitted unexpected output: typography=%q emoji=%q", typography, emoji)
	}
}

func TestMalformedExtensionSyntaxRemainsInert(t *testing.T) {
	input := "```detailsonclick=alert(1)\n<img src=x onerror=alert(2)>\n```\n\n> [!UNKNOWN]\n> ordinary quote\n\n==unterminated\n\n:::stats recent=999999999999999999999999999999999999999999:::\n"
	got := renderStep5Markdown(t, input)
	if strings.Contains(got, "<details") || strings.Contains(got, "markdown-alert-unknown") || strings.Contains(got, "wiki-stats") {
		t.Fatalf("malformed syntax created trusted extension markup: %s", got)
	}
	if strings.Contains(strings.ToLower(got), "<img") {
		t.Fatalf("malformed fenced extension emitted active HTML: %s", got)
	}
}

func renderStep5Markdown(t *testing.T, input string) string {
	t.Helper()
	markdown := goldmark.New(
		goldmark.WithExtensions(
			extension.Table,
			extension.Strikethrough,
			extension.Footnote,
			extension.GFM,
			TrustedNodes,
		),
		goldmark.WithParserOptions(parser.WithAutoHeadingID(), parser.WithAttribute()),
		goldmark.WithRendererOptions(goldhtml.WithUnsafe(), goldhtml.WithHardWraps()),
	)
	var output bytes.Buffer
	if err := markdown.Convert([]byte(input), &output, parser.WithContext(NewRenderContext("test/document"))); err != nil {
		t.Fatalf("render Markdown: %v", err)
	}
	return output.String()
}
