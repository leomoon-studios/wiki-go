package markdown

import (
	"io"
	"strings"

	"github.com/gomarkdown/markdown/ast"
)

// CurrentDocumentPath holds the path of the document being rendered
// This is set by the document handler before rendering
var CurrentDocumentPath string

// RenderNodeHook is a custom renderer for extended markdown features
// It processes special elements like diagrams, video embeds, and collapsible sections
func RenderNodeHook(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	// Check if this is a code block
	if cb, ok := node.(*ast.CodeBlock); ok && entering {
		// Try to process as a diagram
		if ProcessDiagramCodeBlock(w, cb) {
			return ast.GoToNext, true
		}

		// Try to process as a video
		if ProcessVideoCodeBlock(w, cb) {
			return ast.GoToNext, true
		}

		// Try to process as a collapsible section
		if ProcessCollapsibleCodeBlock(w, cb) {
			return ast.GoToNext, true
		}

		// Try to process as a stats shortcode
		if ProcessStatsBlock(w, cb) {
			return ast.GoToNext, true
		}

		// Try to process as a direction (LTR/RTL) block
		if ProcessDirectionCodeBlock(w, cb) {
			return ast.GoToNext, true
		}

		// Check if this code block is inside a list item
		isInListItem := false
		parent := node.GetParent()
		for parent != nil {
			if _, ok := parent.(*ast.ListItem); ok {
				isInListItem = true
				break
			}
			parent = parent.GetParent()
		}

		// If this is a code block in a list item, render it without extra blank lines
		if isInListItem {
			codeContent := string(cb.Literal)
			// Trim leading and trailing whitespace to remove blank lines
			codeContent = strings.TrimSpace(codeContent)

			language := string(cb.Info)
			w.Write([]byte("<pre>"))
			if language != "" {
				w.Write([]byte("<code class=\"language-" + language + "\">"))
			} else {
				w.Write([]byte("<code>"))
			}

			// Write the trimmed code content
			w.Write([]byte(codeContent))

			if language != "" {
				w.Write([]byte("</code></pre>"))
			} else {
				w.Write([]byte("</code></pre>"))
			}
			return ast.GoToNext, true
		}

		// Not a special code block, let the default renderer handle it
		return ast.GoToNext, false
	}

	// Process text nodes for highlighting (==text==)
	if text, ok := node.(*ast.Text); ok && entering {
		// Apply highlighting if needed
		if ProcessTextNodeForHighlighting(w, text) {
			return ast.GoToNext, true
		}
	}

	// Handle paragraphs in list items - prevent wrapping in <p> tags
	if para, ok := node.(*ast.Paragraph); ok {
		// Check if this paragraph is a child of a list item
		parent := para.Parent
		if _, isListItem := parent.(*ast.ListItem); isListItem {
			if entering {
				// For entering a paragraph in a list item, don't output the <p> tag
				// Just write the content directly
				return ast.GoToNext, true
			} else {
				// For exiting a paragraph in a list item, don't output the </p> tag
				return ast.GoToNext, true
			}
		}
	}

	// Check if this is a paragraph that might contain our shortcode
	if para, ok := node.(*ast.Paragraph); ok && entering {
		// Try to process as a stats shortcode
		if ProcessStatsParagraph(w, para) {
			return ast.GoToNext, true
		}
	}

	// Handle links - transform local file references
	if link, ok := node.(*ast.Link); ok && entering && CurrentDocumentPath != "" {
		if IsLocalFileReference(string(link.Destination)) {
			// Transform the link destination to a full URL
			link.Destination = []byte(TransformLocalFileReference(string(link.Destination), CurrentDocumentPath))
		}
		return ast.GoToNext, false
	}

	// Handle images - transform local file references
	if image, ok := node.(*ast.Image); ok && entering && CurrentDocumentPath != "" {
		if IsLocalFileReference(string(image.Destination)) {
			// Transform the image destination to a full URL
			image.Destination = []byte(TransformLocalFileReference(string(image.Destination), CurrentDocumentPath))
		}
		return ast.GoToNext, false
	}

	// Not a special node, let the default renderer handle it
	return ast.GoToNext, false
}
