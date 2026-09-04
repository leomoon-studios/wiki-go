package handlers

import (
	"wiki-go/internal/goldext"
	"wiki-go/internal/types"
)

func chapterHeadingsForPage(headings []goldext.TOCHeading) []types.ChapterHeading {
	result := make([]types.ChapterHeading, 0, len(headings))
	for _, heading := range headings {
		if heading.Level < 1 || heading.Level > 6 || heading.ID == "" || goldext.NormalizeIdentifier(heading.ID) != heading.ID {
			continue
		}
		result = append(result, types.ChapterHeading{
			Level: heading.Level,
			Text:  heading.Text,
			ID:    heading.ID,
		})
	}
	return result
}
