package handlers

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"wiki-go/internal/goldext"
	"wiki-go/internal/types"
	"wiki-go/internal/utils"
)

func TestChapterHeadingsForPageAcceptsOnlyValidatedMetadata(t *testing.T) {
	input := []goldext.TOCHeading{
		{Level: 1, Text: `Safe <img src=x onerror="alert(1)">`, ID: "safe-heading"},
		{Level: 0, Text: "Invalid level", ID: "invalid-level"},
		{Level: 7, Text: "Invalid level", ID: "invalid-level-2"},
		{Level: 2, Text: "Empty ID", ID: ""},
		{Level: 2, Text: "Invalid ID", ID: `bad" onclick="alert(2)`},
	}

	got := chapterHeadingsForPage(input, false)
	want := []types.ChapterHeading{
		{Level: 1, Text: input[0].Text, ID: "safe-heading"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected validated headings:\nwant: %#v\n got: %#v", want, got)
	}
}

func TestChapterHeadingsForPagePreservesHierarchy(t *testing.T) {
	input := []goldext.TOCHeading{
		{Level: 1, Text: "Guide", ID: "guide"},
		{Level: 3, Text: "Skipped level", ID: "skipped-level"},
		{Level: 4, Text: "Nested detail", ID: "nested-detail"},
		{Level: 2, Text: "Sibling section", ID: "sibling-section"},
		{Level: 4, Text: "Sibling detail", ID: "sibling-detail"},
		{Level: 1, Text: "Second guide", ID: "second-guide"},
	}

	want := []types.ChapterHeading{
		{
			Level: 1, Text: "Guide", ID: "guide",
			Children: []types.ChapterHeading{
				{
					Level: 3, Text: "Skipped level", ID: "skipped-level",
					Children: []types.ChapterHeading{
						{Level: 4, Text: "Nested detail", ID: "nested-detail"},
					},
				},
				{
					Level: 2, Text: "Sibling section", ID: "sibling-section",
					Children: []types.ChapterHeading{
						{Level: 4, Text: "Sibling detail", ID: "sibling-detail"},
					},
				},
			},
		},
		{Level: 1, Text: "Second guide", ID: "second-guide"},
	}

	if got := chapterHeadingsForPage(input, false); !reflect.DeepEqual(got, want) {
		t.Fatalf("chapter hierarchy was flattened:\nwant: %#v\n got: %#v", want, got)
	}
}

func TestChapterLinksTemplateContextuallyEscapesHeadings(t *testing.T) {
	tmpl, err := getTemplate()
	if err != nil {
		t.Fatal(err)
	}
	renderResult := utils.RenderMarkdownWithPathResult("# Safe <img src=x onerror=\"alert(1)\">\n\n### Nested\n", "guides/outline")
	if len(renderResult.Headings) != 2 {
		t.Fatalf("renderer returned unexpected chapter headings: %#v", renderResult.Headings)
	}
	data := &types.PageData{
		ChapterHeadings: chapterHeadingsForPage(renderResult.Headings, renderResult.HasInlineTOC),
	}

	var output bytes.Buffer
	if err := tmpl.ExecuteTemplate(&output, "chapter-links", data); err != nil {
		t.Fatal(err)
	}
	rendered := output.String()
	for _, expected := range []string{
		`<aside class="chapter-links"`,
		`class="toc-level-1"`,
		`href="#` + renderResult.Headings[0].ID + `"`,
		`Safe &lt;img src=x onerror=&#34;alert(1)&#34;&gt;`,
		`class="toc-level-3"`,
		`href="#` + renderResult.Headings[1].ID + `"`,
	} {
		if !strings.Contains(rendered, expected) {
			t.Errorf("chapter links omitted %q: %s", expected, rendered)
		}
	}
	if strings.Contains(strings.ToLower(rendered), "<img") || strings.Contains(strings.ToLower(rendered), ` onclick=`) {
		t.Fatalf("chapter links emitted active heading markup: %s", rendered)
	}
	for _, heading := range renderResult.Headings {
		if !strings.Contains(string(renderResult.HTML), `id="`+heading.ID+`"`) {
			t.Errorf("chapter link %q does not resolve to a rendered heading: %s", heading.ID, renderResult.HTML)
		}
	}
	parentLink := strings.Index(rendered, `href="#`+renderResult.Headings[0].ID+`"`)
	nestedList := strings.Index(rendered, `class="toc-list toc-list-nested"`)
	childLink := strings.Index(rendered, `href="#`+renderResult.Headings[1].ID+`"`)
	if parentLink < 0 || nestedList < parentLink || childLink < nestedList {
		t.Fatalf("chapter template did not nest the child heading under its parent: %s", rendered)
	}
}

func TestChapterHeadingsForPageSuppressesDuplicateInlineTOC(t *testing.T) {
	renderResult := utils.RenderMarkdownWithPathResult("[toc]\n\n# Visible\n", "guides/outline")
	if !renderResult.HasInlineTOC {
		t.Fatal("renderer did not report the trusted inline TOC")
	}
	if got := chapterHeadingsForPage(renderResult.Headings, renderResult.HasInlineTOC); len(got) != 0 {
		t.Fatalf("chapter panel duplicated an inline TOC: %#v", got)
	}

	codeResult := utils.RenderMarkdownWithPathResult("`[toc]`\n\n# Visible\n", "guides/outline")
	if codeResult.HasInlineTOC {
		t.Fatal("inline code was treated as a trusted inline TOC")
	}
	if got := chapterHeadingsForPage(codeResult.Headings, codeResult.HasInlineTOC); len(got) != 1 {
		t.Fatalf("inline code incorrectly suppressed the chapter panel: %#v", got)
	}
}

func TestChapterLinksTemplateIsAbsentWithoutHeadings(t *testing.T) {
	tmpl, err := getTemplate()
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := tmpl.ExecuteTemplate(&output, "chapter-links", &types.PageData{}); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(output.String()) != "" {
		t.Fatalf("empty chapter outline rendered controls: %q", output.String())
	}
}
