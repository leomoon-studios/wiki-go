package utils

import (
	"strings"
	"testing"
)

func TestDocumentLayoutsUseTrustedFormattingAndNavigationNodes(t *testing.T) {
	body := "[toc]\n\n# Safe Heading\n\n==highlight== H^2^O h~2~o\n\n```details More\n**details**\n```\n\n> [!NOTE]\n> **alert**\n>\n> ```go\n> return fmt.Errorf(\"failed: %w\", err)\n> ```\n"
	tests := map[string]string{
		"regular": body,
		"kanban":  "---\nlayout: kanban\n---\n\n" + body + "\n#### Board\n##### Todo\n- [ ] task\n",
	}

	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			got := string(RenderMarkdownWithPath(input, "guides/start"))
			for _, want := range []string{
				`<nav class="wiki-toc table-of-contents"`,
				`class="heading-anchor"`,
				`<mark>highlight</mark>`,
				`H<sup>2</sup>O`,
				`h<sub>2</sub>o`,
				`<details class="markdown-details">`,
				`class="markdown-alert markdown-alert-note"`,
				`language-go`,
				`return fmt.Errorf(&quot;failed: %w&quot;, err)`,
			} {
				if !strings.Contains(got, want) {
					t.Errorf("trusted extension output omitted %q: %s", want, got)
				}
			}
		})
	}
}

func TestGitHubAlertPreservesFencedCodeThroughDocumentPipeline(t *testing.T) {
	input := `### With Code Blocks

> [!IMPORTANT]
> Remember to include error handling in your code:
>
> ` + "```go" + `
> func processFile(filename string) error {
>     data, err := ioutil.ReadFile(filename)
>     if err != nil {
>         return fmt.Errorf("failed to read file: %w", err)
>     }
>     // Process data...
>     return nil
> }
> ` + "```" + `
`
	got := string(RenderMarkdownWithPath(input, "guides/alerts"))
	for _, want := range []string{
		`class="markdown-alert markdown-alert-important"`,
		`class="markdown-alert-content"`,
		`language-go`,
		`func processFile(filename string) error {`,
		`return fmt.Errorf(&quot;failed to read file: %w&quot;, err)`,
		`</code></pre>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("rendered alert omitted %q: %s", want, got)
		}
	}
	if strings.Contains(got, "```go") || strings.Contains(got, "func processFile(filename string) error {<br>") {
		t.Fatalf("document pipeline flattened alert code into paragraph text: %s", got)
	}
}
