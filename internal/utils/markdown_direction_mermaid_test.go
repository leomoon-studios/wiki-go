package utils

import (
	"strings"
	"testing"
)

func TestDocumentLayoutsUseTrustedDirectionAndMermaidNodes(t *testing.T) {
	body := "```rtl\n**safe direction**\n<img src=x onerror=alert(1)>\n```\n\n```mermaid\ngraph TD\nA-->B\n<img src=x onerror=alert(2)>\n```\n"
	tests := map[string]string{
		"regular": body,
		"kanban":  "---\nlayout: kanban\n---\n\n" + body + "\n#### Board\n##### Todo\n- [ ] task\n",
	}

	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			got := string(RenderMarkdownWithPath(input, "security-test"))
			for _, want := range []string{
				`<div class="rtl" dir="rtl">`,
				`<strong>safe direction</strong>`,
				`<div class="mermaid">`,
				`&lt;img src=x onerror=alert(2)&gt;`,
			} {
				if !strings.Contains(got, want) {
					t.Errorf("output omitted %q: %s", want, got)
				}
			}
			if strings.Contains(strings.ToLower(got), "<img") {
				t.Fatalf("document layout emitted active attachment payload markup: %s", got)
			}
		})
	}
}
