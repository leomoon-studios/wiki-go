package goldext

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// OnePasswordIgnore is an extension that adds data-1p-ignore attribute to all code blocks
var OnePasswordIgnore = &onePasswordIgnoreExtension{}

// onePasswordIgnoreExtension adds data-1p-ignore attribute to all code blocks
type onePasswordIgnoreExtension struct{}

// Extend implements goldmark.Extender
func (e *onePasswordIgnoreExtension) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(&onePasswordIgnoreRenderer{}, 100),
		),
	)
}

// onePasswordIgnoreRenderer renders code blocks with data-1p-ignore attribute
type onePasswordIgnoreRenderer struct{}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs
func (r *onePasswordIgnoreRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
}

func (r *onePasswordIgnoreRenderer) renderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.FencedCodeBlock)
	if entering {
		_, _ = w.WriteString("<pre data-1p-ignore")
		if n.Info != nil {
			segment := n.Info.Segment
			info := segment.Value(source)
			if len(info) > 0 {
				_, _ = w.WriteString(" class=\"language-")
				r.writeLanguage(w, info)
				_, _ = w.WriteString("\"")
			}
		}
		_, _ = w.WriteString("><code>")
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			r.writeEscaped(w, line.Value(source))
		}
	} else {
		_, _ = w.WriteString("</code></pre>\n")
	}
	return ast.WalkContinue, nil
}

func (r *onePasswordIgnoreRenderer) writeLanguage(w util.BufWriter, source []byte) {
	for _, b := range source {
		if b == ' ' || b == '\t' || b == '\n' || b == '\r' {
			break
		}
		_ = w.WriteByte(b)
	}
}

func (r *onePasswordIgnoreRenderer) writeEscaped(w util.BufWriter, source []byte) {
	// Simple HTML escaping
	for _, b := range source {
		switch b {
		case '&':
			_, _ = w.WriteString("&amp;")
		case '<':
			_, _ = w.WriteString("&lt;")
		case '>':
			_, _ = w.WriteString("&gt;")
		case '"':
			_, _ = w.WriteString("&quot;")
		default:
			_ = w.WriteByte(b)
		}
	}
}
