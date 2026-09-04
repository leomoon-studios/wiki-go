package utils

import (
	"reflect"
	"strings"
	"testing"

	"wiki-go/internal/goldext"
)

func TestRenderMarkdownWithPathResultCollectsRenderedHeadings(t *testing.T) {
	input := "# **Formatted** [link](https://example.com) `code`\n\n" +
		"Setext Café 東京\n=================\n\n" +
		"## Repeat\n\n## Repeat\n\n" +
		"## Explicit {#Mixed_ID}\n"

	result := RenderMarkdownWithPathResult(input, "guides/outline")
	want := []goldext.TOCHeading{
		{Level: 1, Text: "Formatted link code", ID: "formatted-link-code"},
		{Level: 1, Text: "Setext Café 東京", ID: "setext-café-東京"},
		{Level: 2, Text: "Repeat", ID: "repeat"},
		{Level: 2, Text: "Repeat", ID: "repeat-1"},
		{Level: 2, Text: "Explicit", ID: "mixed_id"},
	}
	if !reflect.DeepEqual(result.Headings, want) {
		t.Fatalf("unexpected outline headings:\nwant: %#v\n got: %#v", want, result.Headings)
	}

	rendered := string(result.HTML)
	for _, heading := range want {
		if !strings.Contains(rendered, `id="`+heading.ID+`"`) {
			t.Errorf("rendered document omitted outline ID %q: %s", heading.ID, rendered)
		}
	}
}

func TestRenderMarkdownWithPathResultExcludesNonDocumentHeadings(t *testing.T) {
	input := "> # Quoted\n\n" +
		"```markdown\n# Fenced\n```\n\n" +
		"    # Indented\n\n" +
		"# Visible\n"

	result := RenderMarkdownWithPathResult(input, "guides/outline")
	want := []goldext.TOCHeading{{Level: 1, Text: "Visible", ID: "visible"}}
	if !reflect.DeepEqual(result.Headings, want) {
		t.Fatalf("non-document headings entered outline:\nwant: %#v\n got: %#v", want, result.Headings)
	}
}

func TestRenderMarkdownWithPathResultReportsTrustedInlineTOC(t *testing.T) {
	withTOC := RenderMarkdownWithPathResult("[toc]\n\n# Visible\n", "guides/outline")
	if !withTOC.HasInlineTOC {
		t.Fatal("standalone [toc] node was not reported")
	}

	withoutTOC := RenderMarkdownWithPathResult("`[toc]`\n\n# Visible\n", "guides/outline")
	if withoutTOC.HasInlineTOC {
		t.Fatal("inline code was incorrectly reported as a trusted [toc] node")
	}
}

func TestRenderMarkdownWithPathResultKeepsExistingWrappersCompatible(t *testing.T) {
	input := "# Wrapper\n\n**content**\n"
	result := RenderMarkdownWithPathResult(input, "guides/outline")
	if got := RenderMarkdownWithPath(input, "guides/outline"); !reflect.DeepEqual(got, result.HTML) {
		t.Fatalf("path renderer wrapper changed output:\nresult: %s\nwrapper: %s", result.HTML, got)
	}
	if got := RenderMarkdown(input); !reflect.DeepEqual(got, RenderMarkdownWithPathResult(input, "").HTML) {
		t.Fatalf("default renderer wrapper changed output:\n%s", got)
	}
}
