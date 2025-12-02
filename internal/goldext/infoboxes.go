package goldext

import (
	"fmt"
	"regexp"
	"strings"
)

// InfoBoxPreprocessor adds support for GitHub-flavored alerts
// Syntax:
// > [!NOTE]
// > Content
func InfoBoxPreprocessor(markdown string, _ string) string {
	lines := strings.Split(markdown, "\n")
	var result []string

	var inCodeBlock bool
	var codeBlockMarker string

	// Regex to detect the start of an alert
	// Matches "> [!TYPE]" where TYPE is one of the supported types
	alertStartRegex := regexp.MustCompile(`^\s*>\s*\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]\s*$`)

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		// Track code blocks
		if strings.HasPrefix(trimmedLine, "```") || strings.HasPrefix(trimmedLine, "~~~") {
			if !inCodeBlock {
				inCodeBlock = true
				if strings.HasPrefix(trimmedLine, "```") {
					codeBlockMarker = "```"
				} else {
					codeBlockMarker = "~~~"
				}
			} else {
				// Check if it's the closing marker
				// We need to be careful about nested blocks or length of markers, 
				// but for this simple preprocessor, checking prefix is usually enough
				if strings.HasPrefix(trimmedLine, codeBlockMarker) {
					inCodeBlock = false
					codeBlockMarker = ""
				}
			}
			result = append(result, line)
			continue
		}

		if inCodeBlock {
			result = append(result, line)
			continue
		}

		// Check for alert start
		matches := alertStartRegex.FindStringSubmatch(trimmedLine)
		if len(matches) > 1 {
			alertType := strings.ToUpper(matches[1])

			// Collect content
			var content []string
			j := i + 1
			for ; j < len(lines); j++ {
				nextLine := lines[j]
				nextTrimmed := strings.TrimSpace(nextLine)

				// Check if line starts with >
				if strings.HasPrefix(nextTrimmed, ">") {
					// Remove the > and optional space
					// We want to preserve indentation relative to the >
					// Find where the > is in the original line (to handle indentation before >)
					idx := strings.Index(nextLine, ">")
					if idx != -1 {
						contentLine := nextLine[idx+1:]
						if strings.HasPrefix(contentLine, " ") {
							contentLine = contentLine[1:]
						}
						content = append(content, contentLine)
					} else {
						// Should not happen given HasPrefix check on trimmed, but fallback
						content = append(content, strings.TrimPrefix(nextTrimmed, ">"))
					}
				} else if nextTrimmed == "" {
					// Empty line. In GitHub alerts, empty lines are usually quoted too ">"
					// But if the user leaves a blank line, we can treat it as part of the block 
					// if the next line is also part of the block.
					// However, standard markdown breaks blockquote on empty line unless lazy.
					// For this preprocessor, we will stop at empty line to be safe and simple,
					// unless the user explicitly quoted it with ">".
					// Actually, let's stop at non-quoted line.
					break
				} else {
					// Non-empty line not starting with >. End of block.
					break
				}
			}

			// Generate HTML
			html := generateAlertHTML(alertType, content)
			result = append(result, html)

			// Skip processed lines
			i = j - 1
			continue
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

func generateAlertHTML(alertType string, content []string) string {
	var iconHTML string
	var title string
	var classSuffix string

	switch alertType {
	case "NOTE":
		title = "Note"
		classSuffix = "note"
		iconHTML = `<i class="fa fa-info-circle" aria-hidden="true"></i>`
	case "TIP":
		title = "Tip"
		classSuffix = "tip"
		iconHTML = `<i class="fa fa-lightbulb-o" aria-hidden="true"></i>`
	case "IMPORTANT":
		title = "Important"
		classSuffix = "important"
		iconHTML = `<i class="fa fa-exclamation-circle" aria-hidden="true"></i>`
	case "WARNING":
		title = "Warning"
		classSuffix = "warning"
		iconHTML = `<i class="fa fa-exclamation-triangle" aria-hidden="true"></i>`
	case "CAUTION":
		title = "Caution"
		classSuffix = "caution"
		iconHTML = `<i class="fa fa-ban" aria-hidden="true"></i>`
	}

	html := fmt.Sprintf(`<div class="markdown-alert markdown-alert-%s">
<p class="markdown-alert-title">
  %s
  %s
</p>
<div class="markdown-alert-content">

%s

</div>
</div>`, classSuffix, iconHTML, title, strings.Join(content, "\n"))

	return html
}
