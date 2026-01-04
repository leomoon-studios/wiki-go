package goldext

import (
	"strings"
)

// TypographyPreprocessor replaces common typography shortcuts with proper Unicode symbols
// but avoids processing text inside code blocks
func TypographyPreprocessor(markdown string, _ string) string {
	// Process line by line instead of relying on regex which might fail on large documents
	lines := strings.Split(markdown, "\n")
	var result []string

	inCodeBlock := false

	// Define replacements mapping
	replacements := map[string]string{
		"(c)":  "©", // Copyright symbol
		"(r)":  "®", // Registered trademark symbol
		"(tm)": "™", // Trademark symbol
		"(p)":  "¶", // Paragraph symbol
		"+-":   "±", // Plus-minus symbol
		"...":  "…", // Ellipsis
		"(1/2)":  "½", // One-half
		"(1/4)":  "¼", // One-quarter
		"(3/4)":  "¾", // Three-quarters
	}

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
				// Apply all replacements to non-code segments
				for shortcut, symbol := range replacements {
					segment = strings.ReplaceAll(segment, shortcut, symbol)
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
