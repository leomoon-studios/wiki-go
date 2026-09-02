package goldext

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type documentNodeTransformer struct{}

func (t *documentNodeTransformer) Transform(document *ast.Document, reader text.Reader, _ parser.Context) {
	source := reader.Source()
	var headings []*ast.Heading
	var paragraphs []*ast.Paragraph
	var blockquotes []*ast.Blockquote
	_ = ast.Walk(document, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch typed := node.(type) {
		case *ast.Heading:
			headings = append(headings, typed)
		case *ast.Paragraph:
			paragraphs = append(paragraphs, typed)
		case *ast.Blockquote:
			blockquotes = append(blockquotes, typed)
		}
		return ast.WalkContinue, nil
	})

	tocHeadings := transformHeadings(headings, source)
	for _, paragraph := range paragraphs {
		if paragraph.Parent() == nil {
			continue
		}
		paragraphSource := strings.TrimSpace(string(paragraph.Lines().Value(source)))
		parent := paragraph.Parent()
		if paragraphSource == "[toc]" {
			parent.ReplaceChild(parent, paragraph, newTOCBlock(tocHeadings))
			continue
		}
		if stats, matched := parseStatsBlock(paragraphSource); matched {
			parent.ReplaceChild(parent, paragraph, stats)
		}
	}

	for _, blockquote := range blockquotes {
		if blockquote.Parent() == nil {
			continue
		}
		alert, ok := alertFromBlockquote(blockquote, source)
		if !ok {
			continue
		}
		parent := blockquote.Parent()
		parent.ReplaceChild(parent, blockquote, alert)
	}
}

func transformHeadings(headings []*ast.Heading, source []byte) []TOCHeading {
	usedIDs := make(map[string]bool)
	result := make([]TOCHeading, 0, len(headings))
	for _, heading := range headings {
		escapeRawHTMLDescendants(heading, source)
		text := strings.TrimSpace(plainNodeText(heading, source))
		if text == "" {
			text = "Heading"
		}

		id := ""
		if value, ok := heading.AttributeString("id"); ok {
			switch typed := value.(type) {
			case []byte:
				id = string(typed)
			case string:
				id = typed
			}
		}
		id = NormalizeIdentifier(id)
		if id == "" {
			id = NormalizeIdentifier(text)
		}
		if id == "" {
			id = "heading"
		}
		baseID := id
		for suffix := uint64(1); usedIDs[id]; suffix++ {
			id = baseID + "-" + uintToString(suffix)
		}
		usedIDs[id] = true
		heading.SetAttributeString("id", []byte(id))

		anchor := NewTrustedInline(TrustedInlineAnchor)
		anchor.Class = "heading-anchor"
		anchor.Href = "#" + id
		anchor.AriaLabel = "Permalink"
		anchor.AppendChild(anchor, ast.NewString([]byte("¶")))
		heading.AppendChild(heading, ast.NewString([]byte(" ")))
		heading.AppendChild(heading, anchor)

		if !hasAncestorKind(heading, ast.KindBlockquote) {
			result = append(result, TOCHeading{Level: heading.Level, Text: text, ID: id})
		}
	}
	return result
}

func escapeRawHTMLDescendants(root ast.Node, source []byte) {
	var rawNodes []*ast.RawHTML
	_ = ast.Walk(root, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if raw, ok := node.(*ast.RawHTML); ok {
				rawNodes = append(rawNodes, raw)
			}
		}
		return ast.WalkContinue, nil
	})
	for _, raw := range rawNodes {
		if parent := raw.Parent(); parent != nil {
			parent.ReplaceChild(parent, raw, ast.NewString(append([]byte(nil), raw.Text(source)...)))
		}
	}
}

func plainNodeText(root ast.Node, source []byte) string {
	var result strings.Builder
	_ = ast.Walk(root, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || node == root {
			return ast.WalkContinue, nil
		}
		switch typed := node.(type) {
		case *ast.Text:
			result.Write(typed.Value(source))
			return ast.WalkSkipChildren, nil
		case *ast.String:
			result.Write(typed.Value)
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	return result.String()
}

func hasAncestorKind(node ast.Node, kind ast.NodeKind) bool {
	for parent := node.Parent(); parent != nil; parent = parent.Parent() {
		if parent.Kind() == kind {
			return true
		}
	}
	return false
}

func alertFromBlockquote(blockquote *ast.Blockquote, source []byte) (*AlertBlock, bool) {
	paragraph, ok := blockquote.FirstChild().(*ast.Paragraph)
	if !ok || paragraph.Lines().Len() == 0 {
		return nil, false
	}
	markerSegment := paragraph.Lines().At(0)
	alertType := parseAlertType(strings.TrimSpace(string(markerSegment.Value(source))))
	if alertType == AlertInvalid {
		return nil, false
	}

	start := bytes.LastIndexByte(source[:markerSegment.Start], '\n') + 1
	stop := maximumSourceStop(blockquote)
	if stop < start {
		return nil, false
	}
	for stop < len(source) && source[stop] != '\n' {
		stop++
	}
	if stop < len(source) {
		stop++
	}

	physicalLines := strings.Split(string(source[start:stop]), "\n")
	contentLines := make([]string, 0, len(physicalLines))
	for index, line := range physicalLines {
		if index == 0 || index == len(physicalLines)-1 && line == "" {
			continue
		}
		contentLines = append(contentLines, stripOneBlockquoteMarker(line))
	}
	return newAlertBlock(alertType, []byte(strings.Join(contentLines, "\n"))), true
}

func maximumSourceStop(root ast.Node) int {
	stop := 0
	_ = ast.Walk(root, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if node.Type() != ast.TypeInline {
			for index := range node.Lines().Len() {
				if segmentStop := node.Lines().At(index).Stop; segmentStop > stop {
					stop = segmentStop
				}
			}
		}
		if textNode, ok := node.(*ast.Text); ok && textNode.Segment.Stop > stop {
			stop = textNode.Segment.Stop
		}
		return ast.WalkContinue, nil
	})
	return stop
}

func stripOneBlockquoteMarker(line string) string {
	leadingLength := len(line) - len(strings.TrimLeft(line, " \t"))
	content := line[leadingLength:]
	if !strings.HasPrefix(content, ">") {
		return line
	}
	content = strings.TrimPrefix(content, ">")
	if strings.HasPrefix(content, " ") {
		content = content[1:]
	}
	return content
}
