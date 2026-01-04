package goldext

import (
	"strings"
)

// SuperscriptPreprocessor adds support for ^superscript^ syntax
func SuperscriptPreprocessor(markdown string, _ string) string {
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

		// Process superscript
		if strings.Contains(line, "^") {
			// Match text between carets, avoiding inline code and inline math
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
					// Process superscript only when not in code or math
				} else if line[i] == '^' && !inCode && !inInlineMath {
					// Skip if this is a footnote reference [^...]
					if i > 0 && line[i-1] == '[' {
						result.WriteByte('^')
						continue
					}

					// Find the closing caret
					end := -1
					for j := i + 1; j < len(line); j++ {
						if line[j] == '^' && j > i+1 {
							end = j
							break
						}
					}

					if end != -1 {
						// Extract the text and wrap in <sup> tags
						superText := line[i+1 : end]
						result.WriteString("<sup>")
						result.WriteString(superText)
						result.WriteString("</sup>")

						// Skip past the processed text
						i = end
					} else {
						// No closing caret, write as-is
						result.WriteByte('^')
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
