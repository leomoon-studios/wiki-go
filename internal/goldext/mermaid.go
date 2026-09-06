package goldext

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

// MermaidBlock is produced only from an exact ```mermaid fenced block. Its
// source is always emitted as escaped text for the strict client-side renderer.
type MermaidBlock struct {
	ast.BaseBlock
}

// KindMermaidBlock is the Goldmark kind for MermaidBlock.
var KindMermaidBlock = ast.NewNodeKind("WikiGoMermaidBlock")

// Kind implements ast.Node.
func (n *MermaidBlock) Kind() ast.NodeKind {
	return KindMermaidBlock
}

// Dump implements ast.Node.
func (n *MermaidBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

func newMermaidBlock() *MermaidBlock {
	return &MermaidBlock{}
}

func (r *trustedNodeRenderer) renderMermaidBlock(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	_, _ = writer.WriteString(`<div class="mermaid">`)
	_, _ = writer.Write(util.EscapeHTML(node.Lines().Value(source)))
	_, _ = writer.WriteString("</div>\n")
	return ast.WalkSkipChildren, nil
}
