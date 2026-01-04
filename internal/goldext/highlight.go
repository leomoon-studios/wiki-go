package goldext

import (
	"regexp"
	"strings"
)

// HighlightPreprocessor adds support for ==highlighted text==
// It correctly handles code blocks, math blocks, and other special sections
func HighlightPreprocessor(markdown string, _ string) string {
	// Process line by line instead of relying on regex which might fail on large documents
	lines := strings.Split(markdown, "\n")
	var result []string

	inCodeBlock := false
	highlightRegex := regexp.MustCompile(`([^=]|^)==([^=\n]+?)==([^=]|$)`)

	for _, line := range lines {
		// Check if this line starts or ends a code block
		trimmedLine := strings.TrimSpace(line)

		// Strip blockquote prefix(es) to detect code blocks inside blockquotes
		// e.g., "> ```" or "> > ```" should be detected as code block markers
		contentLine := trimmedLine
		for strings.HasPrefix(contentLine, ">") {
			contentLine = strings.TrimSpace(strings.TrimPrefix(contentLine, ">"))
		}

		if strings.HasPrefix(contentLine, "```") || strings.HasPrefix(contentLine, "~~~") {
			inCodeBlock = !inCodeBlock
			result = append(result, line)
			continue
		}

		// If we're in a code block, don't process
		if inCodeBlock {
			result = append(result, line)
			continue
		}

		// Process each segment of the line, preserving inline code
		var processedLine string
		segments := strings.Split(line, "`")

		for i, segment := range segments {
			// Even segments (0, 2, 4...) are outside inline code
			if i%2 == 0 {
				// Apply highlight replacement to non-code segments
				for {
					replaced := highlightRegex.ReplaceAllString(segment, "$1<mark>$2</mark>$3")
					if replaced == segment {
						break // No more matches
					}
					segment = replaced
				}
				processedLine += segment
			} else {
				// Odd segments (1, 3, 5...) are inside inline code - preserve them
				processedLine += "`" + segment + "`"
			}
		}

		result = append(result, processedLine)
	}

	return strings.Join(result, "\n")
}
