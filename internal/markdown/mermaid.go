package markdown

import (
	"bytes"
	"io"

	"github.com/gomarkdown/markdown/ast"
)

// ProcessMermaidDiagram processes a code block that contains a Mermaid diagram
// Returns true if the block was processed as a Mermaid diagram, false otherwise
func ProcessMermaidDiagram(w io.Writer, codeBlock *ast.CodeBlock) bool {
	// Check if it's a Mermaid diagram
	if bytes.Equal(codeBlock.Info, []byte("mermaid")) {
		// Render as a Mermaid diagram
		w.Write([]byte("<div class=\"mermaid\">\n"))
		w.Write(codeBlock.Literal)
		w.Write([]byte("</div>\n"))
		return true
	}

	return false
}
