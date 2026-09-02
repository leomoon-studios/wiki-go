package utils

import (
	"html"
	"html/template"
	"strings"
	"testing"
)

func TestRenderCommentMarkdownPreservesSupportedMarkdown(t *testing.T) {
	input := `# Comment heading

**bold**, *emphasis*, ~~removed~~, and ` + "`code`" + `.

- first
- second

[local guide](/guide)

| Name | Value |
| --- | --- |
| Project | Wiki-Go |`

	var rendered template.HTML = RenderCommentMarkdown(input)
	got := string(rendered)

	for _, want := range []string{
		`<h1 id="comment-heading">Comment heading</h1>`,
		`<strong>bold</strong>`,
		`<em>emphasis</em>`,
		`<del>removed</del>`,
		`<code>code</code>`,
		`<ul>`,
		`<a href="/guide">local guide</a>`,
		`<table>`,
		`<td>Wiki-Go</td>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("rendered comment omitted %q\noutput: %s", want, got)
		}
	}
}

func TestRenderCommentMarkdownRejectsActiveHTML(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"script", `<script>alert(1)</script><p>safe</p>`},
		{"event image", `<img src=x oNeRrOr="alert(1)">`},
		{"svg event", `<svg><circle onload="alert(1)"></circle></svg>`},
		{"iframe srcdoc", `<iframe srcdoc="<script>alert(1)</script>"></iframe>`},
		{"malformed handler", `<IMG SRC=/safe.png ONERROR=alert(1)//><strong>safe`},
		{"encoded raw URL", `<a href="javas&#x63;ript:alert(1)">click</a>`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strings.ToLower(html.UnescapeString(string(RenderCommentMarkdown(tt.input))))
			for _, forbidden := range []string{
				"<script", "<img", "<svg", "<circle", "<iframe", "onerror", "onload", "srcdoc", "javascript:",
			} {
				if strings.Contains(got, forbidden) {
					t.Errorf("rendered comment retained %q: %s", forbidden, got)
				}
			}
		})
	}
}

func TestRenderCommentMarkdownFiltersDangerousLinkDestinations(t *testing.T) {
	tests := []struct {
		name        string
		destination string
	}{
		{"javascript", `javascript:alert(1)`},
		{"mixed case javascript", `JaVaScRiPt:alert(1)`},
		{"encoded javascript", `javas&#x63;ript:alert(1)`},
		{"data HTML", `data:text/html;base64,PHNjcmlwdD4=`},
		{"vbscript", `vbscript:msgbox(1)`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strings.ToLower(html.UnescapeString(string(RenderCommentMarkdown(`[click](` + tt.destination + `)`))))
			for _, forbidden := range []string{"href=\"javascript:", "href=\"data:", "href=\"vbscript:"} {
				if strings.Contains(got, forbidden) {
					t.Errorf("rendered comment retained dangerous link destination %q: %s", tt.destination, got)
				}
			}
		})
	}
}

func TestRenderCommentMarkdownDoesNotEnableDocumentExtensions(t *testing.T) {
	input := "```mermaid\ngraph TD; A-->B\n```\n\n:::toc:::"
	got := string(RenderCommentMarkdown(input))

	if strings.Contains(got, `<div class="mermaid">`) || strings.Contains(got, `<nav class="wiki-toc`) {
		t.Fatalf("comment renderer enabled a document-only extension: %s", got)
	}
	if !strings.Contains(got, `<code class="language-mermaid">`) {
		t.Fatalf("comment renderer did not preserve Mermaid input as inert code: %s", got)
	}
}
