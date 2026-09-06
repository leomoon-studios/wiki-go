package goldext

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type trustedInlineSyntaxParser struct{}

func (p *trustedInlineSyntaxParser) Trigger() []byte {
	return []byte{'=', '^', '~'}
}

func (p *trustedInlineSyntaxParser) Parse(_ ast.Node, reader text.Reader, _ parser.Context) ast.Node {
	line, _ := reader.PeekLine()
	if len(line) == 0 {
		return nil
	}
	_, position := reader.Position()
	if positionInsideMath(reader.Source(), position.Start) {
		return nil
	}

	var marker []byte
	var element TrustedInlineElement
	switch line[0] {
	case '=':
		if len(line) < 4 || line[1] != '=' || (len(line) > 2 && line[2] == '=') || reader.PrecendingCharacter() == '=' {
			return nil
		}
		marker = []byte("==")
		element = TrustedInlineMark
	case '^':
		marker = []byte("^")
		element = TrustedInlineSup
	case '~':
		if len(line) < 3 || line[1] == '~' || reader.PrecendingCharacter() == '~' {
			return nil
		}
		marker = []byte("~")
		element = TrustedInlineSub
	default:
		return nil
	}

	closing := findInlineClosingDelimiter(line, marker)
	if closing < len(marker)+1 {
		return nil
	}

	node := NewTrustedInline(element)
	content := append([]byte(nil), line[len(marker):closing]...)
	node.AppendChild(node, ast.NewString(content))
	reader.Advance(closing + len(marker))
	return node
}

func findInlineClosingDelimiter(line, marker []byte) int {
	for offset := len(marker); offset+len(marker) <= len(line); {
		relative := bytes.Index(line[offset:], marker)
		if relative < 0 {
			return -1
		}
		closing := offset + relative
		if closing > len(marker) && !delimiterTouchesSameMarker(line, closing, marker) {
			return closing
		}
		offset = closing + 1
	}
	return -1
}

func delimiterTouchesSameMarker(line []byte, closing int, marker []byte) bool {
	character := marker[0]
	return closing > 0 && line[closing-1] == character ||
		closing+len(marker) < len(line) && line[closing+len(marker)] == character
}

func positionInsideMath(source []byte, position int) bool {
	blockMath := false
	inlineMath := false
	fenceMarker := byte(0)
	if position > len(source) {
		position = len(source)
	}
	lines := bytes.Split(source[:position], []byte("\n"))
	for lineIndex, line := range lines {
		trimmed := bytes.TrimSpace(line)
		for len(trimmed) > 0 && trimmed[0] == '>' {
			trimmed = bytes.TrimSpace(trimmed[1:])
		}
		if len(trimmed) >= 3 && (bytes.HasPrefix(trimmed, []byte("```")) || bytes.HasPrefix(trimmed, []byte("~~~"))) {
			marker := trimmed[0]
			if fenceMarker == 0 {
				fenceMarker = marker
			} else if fenceMarker == marker {
				fenceMarker = 0
			}
			inlineMath = false
			continue
		}
		if fenceMarker != 0 {
			continue
		}

		inlineCode := false
		for index := 0; index < len(line); index++ {
			switch line[index] {
			case '\\':
				index++
			case '`':
				inlineCode = !inlineCode
			case '$':
				if inlineCode {
					continue
				}
				if index+1 < len(line) && line[index+1] == '$' {
					blockMath = !blockMath
					inlineMath = false
					index++
				} else if !blockMath {
					inlineMath = !inlineMath
				}
			}
		}
		if lineIndex < len(lines)-1 && !blockMath {
			inlineMath = false
		}
	}
	return blockMath || inlineMath
}

type safeInlineFormattingExtension struct{}

func (e *safeInlineFormattingExtension) Extend(markdown goldmark.Markdown) {
	markdown.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(&trustedInlineSyntaxParser{}, 250),
	))
	markdown.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&trustedNodeRenderer{}, 100),
	))
}
