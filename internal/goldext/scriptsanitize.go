package goldext

import (
	"regexp"
	"strings"
)

// ScriptSanitizePreprocessor removes script tags from markdown content
// but preserves them in code blocks (both fenced and inline)
func ScriptSanitizePreprocessor(markdown string, _ string) string {
	// Process line by line instead of relying on splitCodeSections
	// This ensures we properly handle both ``` and ~~~ code blocks
	lines := strings.Split(markdown, "\n")
	var result []string

	inCodeBlock := false

	for _, line := range lines {
		// Check if this line starts or ends a code block (either ``` or ~~~)
		trimmedLine := strings.TrimSpace(line)
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

		// Handle inline code blocks in this line
		var processedLine string
		segments := strings.Split(line, "`")

		// Process each segment of the line
		for i, segment := range segments {
			// Even segments (0, 2, 4...) are outside inline code - process them
			if i%2 == 0 {
				// Process script tags in non-code segments
				segment = regexp.MustCompile(`(?i)<\s*script\b[^>]*>(.*?)<\s*/\s*script\s*>`).
					ReplaceAllString(segment, "")
				segment = regexp.MustCompile(`(?i)<\s*script\b[^>]*\s*/>`).
					ReplaceAllString(segment, "")
				segment = regexp.MustCompile(`(?i)<\s*script\b[^>]*>`).
					ReplaceAllString(segment, "")
				segment = regexp.MustCompile(`(?i)<\s*/\s*script\s*>`).
					ReplaceAllString(segment, "")

				processedLine += segment
			} else {
				// Odd segments (1, 3, 5...) are inside inline code - preserve them
				processedLine += "`" + segment + "`"
			}
		}

		// The above logic handles a line with even number of backticks
		// If there's an uneven number of backticks, the last segment won't have a closing backtick
		// This is handled correctly since we only add backticks around odd segments

		result = append(result, processedLine)
	}

	return strings.Join(result, "\n")
}
