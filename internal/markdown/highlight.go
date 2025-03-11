package markdown

import (
	"bytes"
	"io"
	"regexp"

	"github.com/gomarkdown/markdown/ast"
)

// Regular expression to find ==highlighted text==
var highlightRegex = regexp.MustCompile(`==([^=]+?)==`)

// ProcessHighlight converts ==text== to <mark>text</mark>
func ProcessHighlight(text []byte) []byte {
	return highlightRegex.ReplaceAll(text, []byte("<mark>$1</mark>"))
}

// ProcessTextNodeForHighlighting processes text nodes to apply highlighting
// Returns true if highlights were applied, false otherwise
func ProcessTextNodeForHighlighting(w io.Writer, node *ast.Text) bool {
	// Check if the text contains highlight markers
	if bytes.Contains(node.Literal, []byte("==")) {
		// Apply highlighting
		highlighted := ProcessHighlight(node.Literal)
		w.Write(highlighted)
		return true
	}
	return false
}
