package goldext

import (
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// TrustedNodes registers renderers for Wiki-Go's typed, generated nodes. It
// deliberately registers no Markdown parser, so user input cannot create one.
var TrustedNodes = &trustedNodesExtension{}

type trustedNodesExtension struct{}

func (e *trustedNodesExtension) Extend(markdown goldmark.Markdown) {
	markdown.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(&trustedFenceTransformer{}, 100),
		),
	)
	markdown.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(&trustedNodeRenderer{}, 100),
		),
	)
}

type trustedNodeRenderer struct{}

func (r *trustedNodeRenderer) RegisterFuncs(registerer renderer.NodeRendererFuncRegisterer) {
	registerer.Register(KindTrustedBlock, r.renderBlock)
	registerer.Register(KindTrustedInline, r.renderInline)
	registerer.Register(KindDirectionBlock, r.renderDirectionBlock)
	registerer.Register(KindMermaidBlock, r.renderMermaidBlock)
}

func (r *trustedNodeRenderer) renderBlock(writer util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	block := node.(*TrustedBlock)
	tag, ok := trustedBlockTag(block.Element)
	if !ok {
		return ast.WalkSkipChildren, nil
	}

	if !entering {
		_, _ = writer.WriteString("</" + tag + ">\n")
		return ast.WalkContinue, nil
	}

	_, _ = writer.WriteString("<" + tag)
	writeCommonAttributes(writer, block.Class, block.ID, "", block.AriaLabel)
	if direction := strings.ToLower(block.Direction); direction == "ltr" || direction == "rtl" || direction == "auto" {
		writeAttribute(writer, "dir", direction)
	}
	if block.Element == TrustedBlockDetails && block.Open {
		_, _ = writer.WriteString(" open")
	}
	_ = writer.WriteByte('>')
	return ast.WalkContinue, nil
}

func (r *trustedNodeRenderer) renderInline(writer util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	inline := node.(*TrustedInline)
	tag, ok := trustedInlineTag(inline)
	if !ok {
		return ast.WalkSkipChildren, nil
	}

	if !entering {
		_, _ = writer.WriteString("</" + tag + ">")
		return ast.WalkContinue, nil
	}

	_, _ = writer.WriteString("<" + tag)
	writeCommonAttributes(writer, inline.Class, inline.ID, inline.Title, inline.AriaLabel)
	if inline.Element == TrustedInlineAnchor {
		if href, valid := ValidateTrustedURL(inline.Href); valid {
			writeAttribute(writer, "href", href)
			if strings.HasPrefix(strings.ToLower(href), "http://") || strings.HasPrefix(strings.ToLower(href), "https://") {
				writeAttribute(writer, "rel", "nofollow noreferrer")
			}
		}
	}
	_ = writer.WriteByte('>')
	return ast.WalkContinue, nil
}

func trustedBlockTag(element TrustedBlockElement) (string, bool) {
	switch element {
	case TrustedBlockDiv:
		return "div", true
	case TrustedBlockNav:
		return "nav", true
	case TrustedBlockDetails:
		return "details", true
	default:
		return "", false
	}
}

func trustedInlineTag(inline *TrustedInline) (string, bool) {
	switch inline.Element {
	case TrustedInlineSpan:
		return "span", true
	case TrustedInlineMark:
		return "mark", true
	case TrustedInlineSup:
		return "sup", true
	case TrustedInlineSub:
		return "sub", true
	case TrustedInlineAnchor:
		return "a", true
	default:
		return "", false
	}
}

func writeCommonAttributes(writer util.BufWriter, class, id, title, ariaLabel string) {
	if class = NormalizeClassList(class); class != "" {
		writeAttribute(writer, "class", class)
	}
	if id = NormalizeIdentifier(id); id != "" {
		writeAttribute(writer, "id", id)
	}
	if title != "" {
		writeAttribute(writer, "title", title)
	}
	if ariaLabel != "" {
		writeAttribute(writer, "aria-label", ariaLabel)
	}
}

func writeAttribute(writer util.BufWriter, name, value string) {
	_, _ = writer.WriteString(" " + name + "=\"")
	_, _ = writer.WriteString(EscapeHTMLText(value))
	_ = writer.WriteByte('"')
}
