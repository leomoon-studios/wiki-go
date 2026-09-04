package handlers

import (
	"wiki-go/internal/goldext"
	"wiki-go/internal/types"
)

type chapterHeadingNode struct {
	heading  types.ChapterHeading
	children []*chapterHeadingNode
}

func chapterHeadingsForPage(headings []goldext.TOCHeading, hasInlineTOC bool) []types.ChapterHeading {
	if hasInlineTOC {
		return nil
	}

	roots := make([]*chapterHeadingNode, 0, len(headings))
	stack := make([]*chapterHeadingNode, 0, 6)
	for _, heading := range headings {
		if heading.Level < 1 || heading.Level > 6 || heading.ID == "" || goldext.NormalizeIdentifier(heading.ID) != heading.ID {
			continue
		}

		node := &chapterHeadingNode{heading: types.ChapterHeading{
			Level: heading.Level,
			Text:  heading.Text,
			ID:    heading.ID,
		}}
		for len(stack) > 0 && stack[len(stack)-1].heading.Level >= heading.Level {
			stack = stack[:len(stack)-1]
		}
		if len(stack) == 0 {
			roots = append(roots, node)
		} else {
			parent := stack[len(stack)-1]
			parent.children = append(parent.children, node)
		}
		stack = append(stack, node)
	}

	result := make([]types.ChapterHeading, 0, len(roots))
	for _, root := range roots {
		result = append(result, chapterHeadingFromNode(root))
	}
	return result
}

func chapterHeadingFromNode(node *chapterHeadingNode) types.ChapterHeading {
	heading := node.heading
	if len(node.children) == 0 {
		return heading
	}
	heading.Children = make([]types.ChapterHeading, 0, len(node.children))
	for _, child := range node.children {
		heading.Children = append(heading.Children, chapterHeadingFromNode(child))
	}
	return heading
}
