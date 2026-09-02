package goldext

import (
	"bytes"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

// DetailsBlock is produced from a details fence. Its title is plain text and
// its body is rendered through the nested safe Markdown renderer.
type DetailsBlock struct {
	ast.BaseBlock
	Title string
}

// KindDetailsBlock is the Goldmark kind for DetailsBlock.
var KindDetailsBlock = ast.NewNodeKind("WikiGoDetailsBlock")

// Kind implements ast.Node.
func (n *DetailsBlock) Kind() ast.NodeKind {
	return KindDetailsBlock
}

// Dump implements ast.Node.
func (n *DetailsBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"Title": n.Title}, nil)
}

func newDetailsBlock(title string) *DetailsBlock {
	if title == "" {
		title = "Details"
	}
	return &DetailsBlock{Title: title}
}

func (r *trustedNodeRenderer) renderDetailsBlock(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	details := node.(*DetailsBlock)
	var rendered bytes.Buffer
	if err := newSafeNestedMarkdown().Convert(node.Lines().Value(source), &rendered); err != nil {
		return ast.WalkSkipChildren, err
	}

	_, _ = writer.WriteString("<details class=\"markdown-details\"><summary>")
	_, _ = writer.Write(util.EscapeHTML([]byte(details.Title)))
	_, _ = writer.WriteString("</summary><div class=\"details-content\">")
	_, _ = writer.Write(rendered.Bytes())
	_, _ = writer.WriteString("</div></details>\n")
	return ast.WalkSkipChildren, nil
}
