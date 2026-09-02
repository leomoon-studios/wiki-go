package goldext

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

// TOCHeading is a safe snapshot of one document heading.
type TOCHeading struct {
	Level int
	Text  string
	ID    string
}

// TOCBlock is produced from a standalone [toc] marker.
type TOCBlock struct {
	ast.BaseBlock
	Headings []TOCHeading
}

// KindTOCBlock is the Goldmark kind for TOCBlock.
var KindTOCBlock = ast.NewNodeKind("WikiGoTOCBlock")

// Kind implements ast.Node.
func (n *TOCBlock) Kind() ast.NodeKind {
	return KindTOCBlock
}

// Dump implements ast.Node.
func (n *TOCBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

func newTOCBlock(headings []TOCHeading) *TOCBlock {
	return &TOCBlock{Headings: append([]TOCHeading(nil), headings...)}
}

func (r *trustedNodeRenderer) renderTOCBlock(writer util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	toc := node.(*TOCBlock)
	_, _ = writer.WriteString(`<nav class="wiki-toc table-of-contents" aria-label="Table of Contents"><div class="toc-title">Table of Contents</div>`)
	if len(toc.Headings) == 0 {
		_, _ = writer.WriteString(`<p class="toc-empty">No headings found in this document.</p></nav>` + "\n")
		return ast.WalkSkipChildren, nil
	}

	_, _ = writer.WriteString(`<ul class="toc-list">`)
	currentLevel := toc.Headings[0].Level
	listDepth := 1
	for index, heading := range toc.Headings {
		level := heading.Level
		if level < 1 || level > 6 || NormalizeIdentifier(heading.ID) != heading.ID {
			continue
		}
		if index == 0 {
			_, _ = writer.WriteString("<li>")
		} else if level > currentLevel {
			for depth := currentLevel; depth < level; depth++ {
				_, _ = writer.WriteString("<ul><li>")
				listDepth++
			}
		} else if level < currentLevel {
			levelsToClose := currentLevel - level
			if levelsToClose >= listDepth {
				levelsToClose = listDepth - 1
			}
			for range levelsToClose {
				_, _ = writer.WriteString("</li></ul>")
				listDepth--
			}
			_, _ = writer.WriteString("</li><li>")
		} else {
			_, _ = writer.WriteString("</li><li>")
		}
		_, _ = writer.WriteString("<a")
		writeAttribute(writer, "href", "#"+heading.ID)
		_, _ = writer.WriteString(">")
		_, _ = writer.Write(util.EscapeHTML([]byte(heading.Text)))
		_, _ = writer.WriteString("</a>")
		currentLevel = level
	}
	for listDepth > 1 {
		_, _ = writer.WriteString("</li></ul>")
		listDepth--
	}
	_, _ = writer.WriteString("</li></ul></nav>\n")
	return ast.WalkSkipChildren, nil
}
