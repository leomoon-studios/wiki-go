package goldext

import (
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// trustedFenceTransformer converts only exact, recognized fenced block info
// strings. Extra values remain ordinary escaped code blocks.
type trustedFenceTransformer struct{}

func (t *trustedFenceTransformer) Transform(document *ast.Document, reader text.Reader, _ parser.Context) {
	source := reader.Source()
	var fences []*ast.FencedCodeBlock
	_ = ast.Walk(document, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if fence, ok := node.(*ast.FencedCodeBlock); ok && fence.Info != nil {
				fences = append(fences, fence)
			}
		}
		return ast.WalkContinue, nil
	})

	for _, fence := range fences {
		var replacement ast.Node
		switch strings.TrimSpace(string(fence.Info.Text(source))) {
		case "ltr":
			replacement = newDirectionBlock(DirectionLTR)
		case "rtl":
			replacement = newDirectionBlock(DirectionRTL)
		case "mermaid":
			replacement = newMermaidBlock()
		default:
			continue
		}

		replacement.SetLines(fence.Lines())
		parent := fence.Parent()
		parent.ReplaceChild(parent, fence, replacement)
	}
}
