package utils

import (
	"strings"
	"testing"
)

func TestDocumentLayoutsUseTrustedMediaNodes(t *testing.T) {
	body := "```mp4\ndemo video.mp4\n```\n\n```youtube\nLcuvxJNIgfE\n```\n\n```vimeo\n92060047\n```\n"
	tests := map[string]string{
		"regular": body,
		"kanban":  "---\nlayout: kanban\n---\n\n" + body + "\n#### Board\n##### Todo\n- [ ] task\n",
	}

	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			got := string(RenderMarkdownWithPath(input, "guides/start"))
			for _, want := range []string{
				`src="/api/files/guides/start/demo%20video.mp4"`,
				`src="https://www.youtube.com/embed/LcuvxJNIgfE"`,
				`src="https://player.vimeo.com/video/92060047"`,
			} {
				if !strings.Contains(got, want) {
					t.Errorf("media output omitted %q: %s", want, got)
				}
			}
		})
	}
}

func TestDocumentRendererRejectsUnsafeMediaInputs(t *testing.T) {
	input := "```mp4\nhttps://evil.example/video.mp4\n```\n\n```youtube\nhttps://youtube.com.evil.example/watch?v=LcuvxJNIgfE\n```\n\n```vimeo\n92060047\" onload=\"alert(1)\n```\n"
	got := string(RenderMarkdownWithPath(input, "guides/start"))

	if strings.Contains(got, "<video") || strings.Contains(got, "<iframe") || strings.Contains(strings.ToLower(got), " onload=\"") {
		t.Fatalf("unsafe media input created active embed markup: %s", got)
	}
}
