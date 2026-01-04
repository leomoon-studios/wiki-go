package goldext

import (
	"strings"
)

// SubscriptPreprocessor adds support for ~subscript~ syntax
func SubscriptPreprocessor(markdown string, _ string) string {
	// Split markdown into lines for processing
	lines := strings.Split(markdown, "\n")
	processedLines := make([]string, 0, len(lines))

	// State tracking for code blocks and math blocks
	inBacktickBlock := false
	inTildeBlock := false
	inMathBlock := false // For $$ math blocks

	// Process each line
	for _, line := range lines {
		// Check for code block markers
		trimmedLine := strings.TrimSpace(line)

		// Strip blockquote prefix(es) to detect code blocks inside blockquotes
		// e.g., "> ```" or "> > ```" should be detected as code block markers
		contentLine := trimmedLine
		for strings.HasPrefix(contentLine, ">") {
			contentLine = strings.TrimSpace(strings.TrimPrefix(contentLine, ">"))
		}

		// Check for math block markers ($$)
		if strings.Contains(contentLine, "$$") {
			// Toggle math block state if we have an odd number of $$ markers
			count := strings.Count(contentLine, "$$")
			if count%2 != 0 {
				inMathBlock = !inMathBlock
			}
			processedLines = append(processedLines, line)
			continue
		}

		if strings.HasPrefix(contentLine, "```") {
			inBacktickBlock = !inBacktickBlock
			processedLines = append(processedLines, line)
			continue
		}

		if strings.HasPrefix(contentLine, "~~~") {
			inTildeBlock = !inTildeBlock
			processedLines = append(processedLines, line)
			continue
		}

		// Skip processing if in a code block or math block
		if inBacktickBlock || inTildeBlock || inMathBlock {
			processedLines = append(processedLines, line)
			continue
		}

		// Process subscript, being careful to distinguish from strikethrough
		if strings.Contains(line, "~") {
			// Process the line character by character to handle single tildes vs double tildes
			var inCode bool
			var inInlineMath bool
			var result strings.Builder

			for i := 0; i < len(line); i++ {
				// Handle inline code
				if line[i] == '`' {
					inCode = !inCode
					result.WriteByte('`')
					// Handle inline math ($)
				} else if line[i] == '$' && !inCode {
					inInlineMath = !inInlineMath
					result.WriteByte('$')
					// Process subscript only when not in code or math
				} else if line[i] == '~' && !inCode && !inInlineMath {
					// Check if it's a double tilde (~~) for strikethrough
					if i+1 < len(line) && line[i+1] == '~' {
						// It's strikethrough, write both tildes
						result.WriteString("~~")
						i++
					} else {
						// Check for a closing single tilde
						end := -1
						for j := i + 1; j < len(line); j++ {
							if line[j] == '~' && (j+1 >= len(line) || line[j+1] != '~') {
								end = j
								break
							}
						}

						if end != -1 {
							// Extract the text and wrap in <sub> tags
							subText := line[i+1 : end]
							result.WriteString("<sub>")
							result.WriteString(subText)
							result.WriteString("</sub>")

							// Skip past the processed text
							i = end
						} else {
							// No closing tilde, write as-is
							result.WriteByte('~')
						}
					}
				} else {
					result.WriteByte(line[i])
				}
			}

			processedLines = append(processedLines, result.String())
		} else {
			processedLines = append(processedLines, line)
		}
	}

	return strings.Join(processedLines, "\n")
}
