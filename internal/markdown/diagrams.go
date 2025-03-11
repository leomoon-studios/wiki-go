package markdown

import (
	"io"

	"github.com/gomarkdown/markdown/ast"
)

// ProcessDiagramCodeBlock processes a code block that might contain a diagram
// Returns true if the block was processed as a diagram, false otherwise
func ProcessDiagramCodeBlock(w io.Writer, codeBlock *ast.CodeBlock) bool {
	// Check if it's a Mermaid diagram
	if ProcessMermaidDiagram(w, codeBlock) {
		return true
	}

	// Add other diagram types here in the future

	return false
}
