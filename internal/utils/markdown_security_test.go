package utils

import (
	"strings"
	"testing"
)

const rawHTMLAttack = `<script>alert("script")</script>
<svg><circle onmouseover="alert('svg')"></circle></svg>
<img src=x onerror="alert('image')">
<iframe srcdoc="<script>alert('frame')</script>"></iframe>
<div style="background:url(javascript:alert(1))" onclick="alert('div')">unsafe</div>
[unsafe link](javascript:alert(1))`

func TestDocumentRenderingPathsMakeRawHTMLInert(t *testing.T) {
	regularMarkdown := `# Safe heading

**bold** and ==highlight==

> [!IMPORTANT]
> Alert content
>
> ` + "```go" + `
> return nil
> ` + "```" + `

` + "```rtl" + `
**right-to-left**
<svg onload="alert(1)"></svg>
` + "```" + `

` + rawHTMLAttack

	kanbanMarkdown := `---
layout: kanban
---

# Kanban heading

` + rawHTMLAttack + `

#### Board <svg onload="alert(2)">
##### Column <img src=x onerror="alert(3)">
- [ ] **task** <svg onload="alert(4)"></svg>
`

	linksMarkdown := `---
layout: links
---

# Links <svg onload="alert(5)">

## Category <img src=x onerror="alert(6)">
- [**safe link**](https://example.com/docs?q=wiki) - Description <script>alert(7)</script> | 2026-09-01
- [unsafe link](javascript:alert(8)) - rejected | 2026-09-01
`

	tests := []struct {
		name     string
		render   func() string
		expected []string
	}{
		{
			name:   "regular document",
			render: func() string { return string(RenderMarkdownWithPath(regularMarkdown, "guides/security")) },
			expected: []string{
				`<strong>bold</strong>`,
				`<mark>highlight</mark>`,
				`markdown-alert-important`,
				`language-go`,
				`class="rtl" dir="rtl"`,
				`<strong>right-to-left</strong>`,
			},
		},
		{
			name:   "home page",
			render: func() string { return string(RenderMarkdownHTML(regularMarkdown)) },
			expected: []string{
				`<strong>bold</strong>`,
				`<mark>highlight</mark>`,
				`markdown-alert-important`,
			},
		},
		{
			name:   "editor preview with source lines",
			render: func() string { return string(RenderMarkdownWithSourceLines(regularMarkdown, "guides/security")) },
			expected: []string{
				`data-source-line="0"`,
				`<strong>bold</strong>`,
				`markdown-alert-important`,
			},
		},
		{
			name:   "version preview",
			render: func() string { return string(RenderMarkdownWithPath(regularMarkdown, "guides/security")) },
			expected: []string{
				`<strong>bold</strong>`,
				`markdown-alert-important`,
				`language-go`,
			},
		},
		{
			name:   "kanban layout",
			render: func() string { return string(RenderMarkdownWithPathHTML(kanbanMarkdown, "boards/security")) },
			expected: []string{
				`class="kanban-container"`,
				`<strong>task</strong>`,
				`&lt;svg onload=&#34;alert(2)&#34;&gt;`,
			},
		},
		{
			name:   "links layout",
			render: func() string { return string(RenderMarkdownWithPath(linksMarkdown, "links/security")) },
			expected: []string{
				`class="links-container"`,
				`href="https://example.com/docs?q=wiki"`,
				`Description &lt;script&gt;alert(7)&lt;/script&gt;`,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.render()
			assertNoActiveRawHTML(t, got)
			for _, expected := range test.expected {
				if !strings.Contains(got, expected) {
					t.Errorf("safe rendering omitted %q:\n%s", expected, got)
				}
			}
		})
	}
}

func TestRawHTMLIsHandledByGoldmarkWithoutRegexSanitizing(t *testing.T) {
	input := `<ScRiPt data-value=">"><svg onload=alert(1)>alert(2)</ScRiPt>

<div><span malformed="'><img src=x onerror=alert(3)>">content</div>`
	got := string(RenderMarkdownWithPath(input, "security"))
	assertNoActiveRawHTML(t, got)
	if !strings.Contains(got, "raw HTML omitted") {
		t.Fatalf("expected Goldmark to omit raw HTML nodes: %s", got)
	}
}

func TestAdjacentStatsShortcodesRenderThroughSafeDocumentBoundary(t *testing.T) {
	got := string(RenderMarkdownWithPath(":::stats recent=5::: :::stats count=*:::", "dashboard"))
	if strings.Count(got, `class="wiki-stats recent-edits"`) != 1 || strings.Count(got, `class="wiki-stats doc-count"`) != 1 {
		t.Fatalf("adjacent stats shortcodes did not render through the document boundary: %s", got)
	}
	assertNoActiveRawHTML(t, got)
}

func assertNoActiveRawHTML(t *testing.T, rendered string) {
	t.Helper()
	lower := strings.ToLower(rendered)
	for _, forbidden := range []string{
		"<script", "<svg", "<circle", "<iframe", "<img src=x", "<div style=",
		`onmouseover="`, `onerror="`, `onload="`, `onclick="`, `srcdoc="`,
		`href="javascript:`, `href="data:text/html`, `style="background:url(javascript:`,
	} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("rendered output contains active raw HTML %q:\n%s", forbidden, rendered)
		}
	}
}
