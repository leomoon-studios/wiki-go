package goldext

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// DirectionBlock is produced only from an exact ```rtl or ```ltr fenced block.
// Direction is kept as an enum so the renderer cannot emit arbitrary values.
type DirectionBlock struct {
	ast.BaseBlock
	Direction TextDirection
}

// TextDirection is a fixed direction accepted by DirectionBlock.
type TextDirection uint8

const (
	DirectionInvalid TextDirection = iota
	DirectionLTR
	DirectionRTL
)

// KindDirectionBlock is the Goldmark kind for DirectionBlock.
var KindDirectionBlock = ast.NewNodeKind("WikiGoDirectionBlock")

// Kind implements ast.Node.
func (n *DirectionBlock) Kind() ast.NodeKind {
	return KindDirectionBlock
}

// Dump implements ast.Node.
func (n *DirectionBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"Direction": directionValue(n.Direction),
	}, nil)
}

func newDirectionBlock(direction TextDirection) *DirectionBlock {
	return &DirectionBlock{Direction: direction}
}

func directionValue(direction TextDirection) string {
	switch direction {
	case DirectionLTR:
		return "ltr"
	case DirectionRTL:
		return "rtl"
	default:
		return ""
	}
}

func (r *trustedNodeRenderer) renderDirectionBlock(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	direction := directionValue(node.(*DirectionBlock).Direction)
	if direction == "" {
		return ast.WalkSkipChildren, nil
	}

	content := node.Lines().Value(source)
	var rendered bytes.Buffer
	if err := newSafeNestedMarkdown().Convert(content, &rendered); err != nil {
		return ast.WalkSkipChildren, err
	}

	_, _ = writer.WriteString(`<div class="` + direction + `" dir="` + direction + `">`)
	_, _ = writer.Write(rendered.Bytes())
	_, _ = writer.WriteString("</div>\n")
	return ast.WalkSkipChildren, nil
}

func newSafeNestedMarkdown() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.Table,
			extension.Strikethrough,
			extension.Linkify,
			extension.Footnote,
			extension.DefinitionList,
			extension.GFM,
			&safeInlineFormattingExtension{},
		),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(goldhtml.WithHardWraps()),
	)
}
