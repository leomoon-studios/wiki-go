package goldext

import (
	"fmt"

	"github.com/yuin/goldmark/ast"
)

// TrustedBlockElement identifies a fixed block element that Wiki-Go may emit.
// It is intentionally an enum so callers cannot supply arbitrary tag names.
type TrustedBlockElement uint8

const (
	TrustedBlockInvalid TrustedBlockElement = iota
	TrustedBlockDiv
	TrustedBlockNav
	TrustedBlockDetails
)

// TrustedBlock is a Wiki-Go-generated block container. Its renderer validates
// every field and ignores normal Goldmark attributes attached to the node.
type TrustedBlock struct {
	ast.BaseBlock
	Element   TrustedBlockElement
	Class     string
	ID        string
	Direction string
	AriaLabel string
	Open      bool
}

// KindTrustedBlock is the Goldmark kind for TrustedBlock.
var KindTrustedBlock = ast.NewNodeKind("WikiGoTrustedBlock")

// Kind implements ast.Node.
func (n *TrustedBlock) Kind() ast.NodeKind {
	return KindTrustedBlock
}

// Dump implements ast.Node.
func (n *TrustedBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"Element": fmt.Sprintf("%d", n.Element),
	}, nil)
}

// NewTrustedBlock creates a block node for a fixed Wiki-Go element.
func NewTrustedBlock(element TrustedBlockElement) *TrustedBlock {
	return &TrustedBlock{Element: element}
}

// TrustedInlineElement identifies a fixed inline element that Wiki-Go may
// emit. Arbitrary element names are never accepted.
type TrustedInlineElement uint8

const (
	TrustedInlineInvalid TrustedInlineElement = iota
	TrustedInlineSpan
	TrustedInlineMark
	TrustedInlineSup
	TrustedInlineSub
	TrustedInlineAnchor
)

// TrustedInline is a Wiki-Go-generated inline container.
type TrustedInline struct {
	ast.BaseInline
	Element   TrustedInlineElement
	Class     string
	ID        string
	Href      string
	Title     string
	AriaLabel string
}

// KindTrustedInline is the Goldmark kind for TrustedInline.
var KindTrustedInline = ast.NewNodeKind("WikiGoTrustedInline")

// Kind implements ast.Node.
func (n *TrustedInline) Kind() ast.NodeKind {
	return KindTrustedInline
}

// Dump implements ast.Node.
func (n *TrustedInline) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"Element": fmt.Sprintf("%d", n.Element),
	}, nil)
}

// NewTrustedInline creates an inline node for a fixed Wiki-Go element.
func NewTrustedInline(element TrustedInlineElement) *TrustedInline {
	return &TrustedInline{Element: element}
}
