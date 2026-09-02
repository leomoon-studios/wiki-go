package goldext

import (
	"bytes"
	"html"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

func TestMediaFencesRenderValidatedTypedNodes(t *testing.T) {
	markdown := newTrustedMediaTestMarkdown()
	input := "```mp4\ndemo video.mp4\n```\n\n```youtube\nhttps://youtu.be/LcuvxJNIgfE\n```\n\n~~~vimeo\nhttps://player.vimeo.com/video/92060047\n~~~\n"
	source := []byte(input)
	document := markdown.Parser().Parse(text.NewReader(source), parser.WithContext(NewRenderContext("guides/getting-started")))
	localVideos := 0
	embeddedVideos := 0
	_ = ast.Walk(document, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch node.(type) {
			case *LocalVideoBlock:
				localVideos++
			case *VideoEmbedBlock:
				embeddedVideos++
			}
		}
		return ast.WalkContinue, nil
	})
	if localVideos != 1 || embeddedVideos != 2 {
		t.Fatalf("media fences were not converted to typed nodes: local=%d embedded=%d", localVideos, embeddedVideos)
	}

	var output bytes.Buffer
	if err := markdown.Renderer().Render(&output, source, document); err != nil {
		t.Fatalf("render valid media fences: %v", err)
	}

	got := output.String()
	for _, want := range []string{
		`<video class="local-video-player"`,
		`src="/api/files/guides/getting-started/demo%20video.mp4"`,
		`src="https://www.youtube.com/embed/LcuvxJNIgfE"`,
		`href="https://www.youtube.com/watch?v=LcuvxJNIgfE"`,
		`src="https://player.vimeo.com/video/92060047"`,
		`href="https://vimeo.com/92060047"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("media output omitted %q: %s", want, got)
		}
	}
	if strings.Count(got, "<iframe ") != 2 {
		t.Errorf("expected exactly two external embeds: %s", got)
	}
}

func TestResolveLocalMP4URL(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		docPath  string
		want     string
		valid    bool
	}{
		{name: "document", filename: "demo video.mp4", docPath: "guides/start", want: "/api/files/guides/start/demo%20video.mp4", valid: true},
		{name: "homepage", filename: "intro.MP4", want: "/api/files/pages/home/intro.MP4", valid: true},
		{name: "escaped quote", filename: `demo" onerror="alert(1).mp4`, docPath: "guide", want: "/api/files/guide/demo%22%20onerror=%22alert%281%29.mp4", valid: true},
		{name: "external https", filename: "https://evil.example/video.mp4", docPath: "guide"},
		{name: "external scheme relative", filename: "//evil.example/video.mp4", docPath: "guide"},
		{name: "absolute API URL", filename: "/api/files/other/video.mp4", docPath: "guide"},
		{name: "javascript scheme", filename: "javascript:payload.mp4", docPath: "guide"},
		{name: "encoded scheme", filename: "javas&#x63;ript:payload.mp4", docPath: "guide"},
		{name: "traversal", filename: "../video.mp4", docPath: "guide"},
		{name: "nested attachment", filename: "folder/video.mp4", docPath: "guide"},
		{name: "wrong extension", filename: "video.webm", docPath: "guide"},
		{name: "multiline", filename: "video.mp4\nnext.mp4", docPath: "guide"},
		{name: "bad document path", filename: "video.mp4", docPath: "guide/../private"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, filename, valid := ResolveLocalMP4URL(test.filename, test.docPath)
			if valid != test.valid || got != test.want {
				t.Fatalf("ResolveLocalMP4URL(%q, %q) = (%q, %q, %t), want URL %q valid %t", test.filename, test.docPath, got, filename, valid, test.want, test.valid)
			}
		})
	}
}

func TestVideoProviderIDValidation(t *testing.T) {
	youtubeValid := map[string]string{
		"LcuvxJNIgfE": "LcuvxJNIgfE",
		"https://www.youtube.com/watch?v=LcuvxJNIgfE":    "LcuvxJNIgfE",
		"https://m.youtube.com/watch?v=LcuvxJNIgfE&t=10": "LcuvxJNIgfE",
		"https://youtu.be/LcuvxJNIgfE?t=10":              "LcuvxJNIgfE",
		"https://youtube.com/embed/LcuvxJNIgfE":          "LcuvxJNIgfE",
	}
	for input, want := range youtubeValid {
		if got := ExtractYouTubeID(input); got != want {
			t.Errorf("ExtractYouTubeID(%q) = %q, want %q", input, got, want)
		}
	}
	for _, input := range []string{
		`LcuvxJNIgfE" onload="alert(1)`,
		"javascript:alert(1)",
		"javas&#x63;ript:alert(1)",
		"https://youtube.com.evil.example/watch?v=LcuvxJNIgfE",
		"https://evil.example/youtube.com/watch?v=LcuvxJNIgfE",
		"https://youtube.com/embed/LcuvxJNIgfE/extra",
		"too-short",
		"LcuvxJNIgfEextra",
	} {
		if got := ExtractYouTubeID(input); got != "" {
			t.Errorf("ExtractYouTubeID(%q) accepted %q", input, got)
		}
	}

	vimeoValid := map[string]string{
		"92060047":                   "92060047",
		"https://vimeo.com/92060047": "92060047",
		"https://www.vimeo.com/92060047?autoplay=1": "92060047",
		"https://player.vimeo.com/video/92060047":   "92060047",
	}
	for input, want := range vimeoValid {
		if got := ExtractVimeoID(input); got != want {
			t.Errorf("ExtractVimeoID(%q) = %q, want %q", input, got, want)
		}
	}
	for _, input := range []string{
		`92060047" onload="alert(1)`,
		"javascript:alert(1)",
		"https://vimeo.com.evil.example/92060047",
		"https://evil.example/vimeo.com/92060047",
		"https://player.vimeo.com/video/92060047/extra",
		"92060047extra",
		"123456789012345678901",
	} {
		if got := ExtractVimeoID(input); got != "" {
			t.Errorf("ExtractVimeoID(%q) accepted %q", input, got)
		}
	}
}

func TestInvalidMediaFencesRemainEscapedCode(t *testing.T) {
	markdown := newTrustedMediaTestMarkdown()
	input := "```mp4\nhttps://evil.example/video.mp4\n```\n\n```youtube\nLcuvxJNIgfE\" onload=\"alert(1)\n```\n\n```vimeo\nhttps://vimeo.com.evil.example/92060047\n```\n"
	var output bytes.Buffer
	if err := markdown.Convert([]byte(input), &output, parser.WithContext(NewRenderContext("guide"))); err != nil {
		t.Fatalf("render invalid media fences: %v", err)
	}

	got := output.String()
	if strings.Contains(got, "<video") || strings.Contains(got, "<iframe") {
		t.Fatalf("invalid media input created an embed: %s", got)
	}
	for _, language := range []string{"mp4", "youtube", "vimeo"} {
		if !strings.Contains(got, `class="language-`+language+`"`) {
			t.Errorf("invalid %s input was not retained as escaped code: %s", language, got)
		}
	}
	if strings.Contains(strings.ToLower(html.UnescapeString(got)), `<iframe`) {
		t.Fatalf("invalid media code contained active iframe markup: %s", got)
	}
}

func TestMediaNodeRenderersFailClosedAndEscapeFilename(t *testing.T) {
	document := ast.NewDocument()
	badLocal := newLocalVideoBlock("https://evil.example/video.mp4", `x</p><img src=x onerror=alert(1)>.mp4`)
	badProvider := newVideoEmbedBlock(VideoProviderYouTube, `LcuvxJNIgfE" onload="alert(2)`)
	goodLocal := newLocalVideoBlock("/api/files/guide/video.mp4", `x</p><img src=x onerror=alert(3)>.mp4`)
	document.AppendChild(document, badLocal)
	document.AppendChild(document, badProvider)
	document.AppendChild(document, goodLocal)

	got := renderTrustedTree(t, document)
	if strings.Contains(got, "evil.example") || strings.Contains(got, "<iframe") || strings.Contains(got, "<img") {
		t.Fatalf("media renderer accepted an invalid node or emitted active filename HTML: %s", got)
	}
	if !strings.Contains(got, `x&lt;/p&gt;&lt;img src=x onerror=alert(3)&gt;.mp4`) {
		t.Fatalf("local video filename was not safely escaped: %s", got)
	}
}

func newTrustedMediaTestMarkdown() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(TrustedNodes),
	)
}
