package goldext

import (
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type documentNodeTransformer struct{}

func (t *documentNodeTransformer) Transform(document *ast.Document, reader text.Reader, context parser.Context) {
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
	hasInlineTOC := false
	for _, paragraph := range paragraphs {
		if paragraph.Parent() == nil {
			continue
		}
		paragraphSource := strings.TrimSpace(string(paragraph.Lines().Value(source)))
		parent := paragraph.Parent()
		if paragraphSource == "[toc]" {
			hasInlineTOC = true
			parent.ReplaceChild(parent, paragraph, newTOCBlock(tocHeadings))
			continue
		}
		if statsBlocks, matched := parseStatsBlocks(paragraphSource); matched {
			for _, stats := range statsBlocks {
				parent.InsertBefore(parent, paragraph, stats)
			}
			parent.RemoveChild(parent, paragraph)
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

	storeDocumentOutline(context, DocumentOutline{
		Headings:     tocHeadings,
		HasInlineTOC: hasInlineTOC,
	})
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

	removeAlertMarker(paragraph, markerSegment)
	if paragraph.Lines().Len() == 0 {
		blockquote.RemoveChild(blockquote, paragraph)
	}

	alert := newAlertBlock(alertType)
	for child := blockquote.FirstChild(); child != nil; {
		next := child.NextSibling()
		blockquote.RemoveChild(blockquote, child)
		alert.AppendChild(alert, child)
		child = next
	}
	escapeRawHTMLBlocksAndInlines(alert, source)
	return alert, true
}

func removeAlertMarker(paragraph *ast.Paragraph, markerSegment text.Segment) {
	for child := paragraph.FirstChild(); child != nil; {
		next := child.NextSibling()
		if child.Pos() >= markerSegment.Stop {
			break
		}
		paragraph.RemoveChild(paragraph, child)
		child = next
	}

	remainingLines := text.NewSegments()
	for index := 1; index < paragraph.Lines().Len(); index++ {
		remainingLines.Append(paragraph.Lines().At(index))
	}
	paragraph.SetLines(remainingLines)
}

func escapeRawHTMLBlocksAndInlines(root ast.Node, source []byte) {
	type replacement struct {
		old ast.Node
		new ast.Node
	}
	replacements := make([]replacement, 0)
	_ = ast.Walk(root, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch typed := node.(type) {
		case *ast.RawHTML:
			replacements = append(replacements, replacement{
				old: typed,
				new: ast.NewString(append([]byte(nil), typed.Text(source)...)),
			})
		case *ast.HTMLBlock:
			paragraph := ast.NewParagraph()
			paragraph.AppendChild(paragraph, ast.NewString(append([]byte(nil), typed.Text(source)...)))
			replacements = append(replacements, replacement{old: typed, new: paragraph})
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	for _, item := range replacements {
		if parent := item.old.Parent(); parent != nil {
			parent.ReplaceChild(parent, item.old, item.new)
		}
	}
}
