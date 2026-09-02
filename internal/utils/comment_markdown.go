package utils

import (
	"bytes"
	"html/template"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
)

const commentRenderError = `<p>Unable to render comment.</p>`

// RenderCommentMarkdown renders the Markdown subset supported in comments.
// Raw HTML and document-only Wiki-Go preprocessors are deliberately disabled.
// The returned value is safe to insert into an html/template because Goldmark
// escapes text, omits raw HTML, and filters dangerous link destinations.
func RenderCommentMarkdown(markdownSource string) template.HTML {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			extension.Table,
			extension.Strikethrough,
			extension.Linkify,
			extension.Footnote,
			extension.DefinitionList,
			extension.GFM,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			goldmarkhtml.WithHardWraps(),
		),
	)

	var output bytes.Buffer
	if err := markdown.Convert([]byte(markdownSource), &output); err != nil {
		return template.HTML(commentRenderError)
	}

	return template.HTML(output.String())
}
