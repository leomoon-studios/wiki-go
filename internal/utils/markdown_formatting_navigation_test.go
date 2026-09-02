package utils

import (
	"strings"
	"testing"
)

func TestDocumentLayoutsUseTrustedFormattingAndNavigationNodes(t *testing.T) {
	body := "[toc]\n\n# Safe Heading\n\n==highlight== H^2^O h~2~o\n\n```details More\n**details**\n```\n\n> [!NOTE]\n> **alert**\n"
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
			} {
				if !strings.Contains(got, want) {
					t.Errorf("trusted extension output omitted %q: %s", want, got)
				}
			}
		})
	}
}
