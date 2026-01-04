package goldext

import (
	"strings"
)

// DetailsPreprocessor adds support for ```details and ~~~details blocks
func DetailsPreprocessor(markdown string, _ string) string {
	lines := strings.Split(markdown, "\n")
	var result []string
	
	var inCodeBlock bool
	var codeBlockMarker string
	
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		// Strip blockquote prefix(es) to detect code blocks inside blockquotes
		// e.g., "> ```" or "> > ```" should be detected as code block markers
		contentLine := trimmedLine
		for strings.HasPrefix(contentLine, ">") {
			contentLine = strings.TrimSpace(strings.TrimPrefix(contentLine, ">"))
		}
		
		// Track regular code blocks (not details blocks)
		if (strings.HasPrefix(contentLine, "```") || strings.HasPrefix(contentLine, "~~~")) &&
			!strings.HasPrefix(contentLine, "```details") && !strings.HasPrefix(contentLine, "~~~details") {
			
			if !inCodeBlock {
				// Starting a code block
				inCodeBlock = true
				if strings.HasPrefix(contentLine, "```") {
					codeBlockMarker = "```"
				} else {
					codeBlockMarker = "~~~"
				}
			} else if contentLine == codeBlockMarker {
				// Ending a code block
				inCodeBlock = false
				codeBlockMarker = ""
			}
			result = append(result, line)
			continue
		}
		
		// If we're inside a regular code block, don't process details
		if inCodeBlock {
			result = append(result, line)
			continue
		}
		
		// Check for details block start
		if strings.HasPrefix(trimmedLine, "```details") || strings.HasPrefix(trimmedLine, "~~~details") {
			var detailsMarker string
			if strings.HasPrefix(trimmedLine, "```details") {
				detailsMarker = "```"
			} else {
				detailsMarker = "~~~"
			}
			
			// Extract the title (everything after "```details" or "~~~details")
			detailsTitle := strings.TrimSpace(strings.TrimPrefix(
				strings.TrimPrefix(trimmedLine, "```details"), "~~~details"))
			
			// Find the end of the details block
			var detailsContent []string
			j := i + 1
			
			for ; j < len(lines); j++ {
				if strings.TrimSpace(lines[j]) == detailsMarker {
					break
				}
				detailsContent = append(detailsContent, lines[j])
			}
			
			// Generate the details HTML with proper markdown content
			detailsHTML := "<details class=\"markdown-details\">"
			if detailsTitle != "" {
				detailsHTML += "<summary>" + detailsTitle + "</summary>"
			} else {
				detailsHTML += "<summary>Details</summary>"
			}
			detailsHTML += "<div class=\"details-content\">"
			
			// Add the content as-is so it can be processed by markdown renderer
			if len(detailsContent) > 0 {
				detailsHTML += "\n\n" + strings.Join(detailsContent, "\n") + "\n\n"
			}
			
			detailsHTML += "</div></details>"
			
			result = append(result, detailsHTML)
			
			// Skip to the end of the details block
			if j < len(lines) {
				i = j
			}
			continue
		}
		
		// Regular line
		result = append(result, line)
	}
	
	return strings.Join(result, "\n")
}

// Register Details preprocessor in the list of known processors
var _ = DetailsPreprocessor
