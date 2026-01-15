package goldext

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// Preprocessor defines a function that transforms markdown before rendering
type Preprocessor func(markdown string, docPath string) string

// RegisteredPreprocessors holds all registered preprocessors
var RegisteredPreprocessors []Preprocessor

// RegisterPreprocessor adds a preprocessor to the list
func RegisterPreprocessor(pp Preprocessor) {
	RegisteredPreprocessors = append(RegisteredPreprocessors, pp)
}

// ProcessMarkdown applies all registered preprocessors to the markdown
func ProcessMarkdown(markdown string, docPath string) string {
	result := markdown
	for _, preprocessor := range RegisteredPreprocessors {
		result = preprocessor(result, docPath)
	}
	return result
}

// Section represents a piece of markdown content that should or shouldn't be processed
type Section struct {
	content string
	isCode  bool
}

// splitCodeSections splits markdown into regular text and code/math/mermaid sections
func splitCodeSections(markdown string) []Section {
	var sections []Section

	// Define regex patterns to match code blocks, inline code, math, and mermaid divs
	codeBlockPattern := "```[\\s\\S]*?```"
	inlineCodePattern := "`[^`]*?`"
	inlineMathPattern := "\\$[^\\$\\n]+?\\$"
	blockMathPattern := "\\$\\$[\\s\\S]*?\\$\\$"
	mermaidDivPattern := "<div class=\"mermaid\">[\\s\\S]*?</div>"

	// Combine patterns to find all protected sections
	combinedPattern := fmt.Sprintf("(?s)(%s|%s|%s|%s|%s)",
		codeBlockPattern, inlineCodePattern, inlineMathPattern, blockMathPattern, mermaidDivPattern)

	// Use (?s) flag to make . match newlines and compile with DOTALL flag for better performance with large inputs
	re := regexp.MustCompile(combinedPattern)

	// Find all protected sections with a 1MB limit to ensure we process the entire document
	// Default limit might be causing issues with large documents
	matches := re.FindAllStringIndex(markdown, -1)

	// If no protected sections found, return the entire markdown as a single non-code section
	if len(matches) == 0 {
		return []Section{{content: markdown, isCode: false}}
	}

	// Process sections between and including protected blocks
	lastEnd := 0
	for _, match := range matches {
		start, end := match[0], match[1]

		// Add non-code section before this protected block (if any)
		if start > lastEnd {
			sections = append(sections, Section{
				content: markdown[lastEnd:start],
				isCode:  false,
			})
		}

		// Add the protected section
		sections = append(sections, Section{
			content: markdown[start:end],
			isCode:  true,
		})

		lastEnd = end
	}

	// Add any remaining text after the last protected section
	if lastEnd < len(markdown) {
		sections = append(sections, Section{
			content: markdown[lastEnd:],
			isCode:  false,
		})
	}

	return sections
}

// joinSections rejoins all sections into a single string
func joinSections(sections []Section) string {
	var result strings.Builder

	for _, section := range sections {
		result.WriteString(section.content)
	}

	return result.String()
}

// LinkPreprocessor resolves local file references
func LinkPreprocessor(markdown string, docPath string) string {
	// This is a simplified implementation
	// A more robust version would use a proper Markdown parser

	// Use the splitCodeSections function to break the markdown into regular text and code sections
	sections := splitCodeSections(markdown)

	// Process only regular text sections
	for i := range sections {
		if !sections[i].isCode {
			// Process image links: ![alt](local-path)
			imgRe := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
			sections[i].content = imgRe.ReplaceAllStringFunc(sections[i].content, func(match string) string {
				parts := imgRe.FindStringSubmatch(match)
				if len(parts) < 3 {
					return match
				}

				alt := parts[1]
				path := parts[2]

				if isLocalPath(path) {
					path = resolveLocalPath(path, docPath)
				}

				return "![" + alt + "](" + path + ")"
			})

			// Process regular links: [text](local-path)
			linkRe := regexp.MustCompile(`\[([^\]]*)\]\(([^)]+)\)`)
			sections[i].content = linkRe.ReplaceAllStringFunc(sections[i].content, func(match string) string {
				parts := linkRe.FindStringSubmatch(match)
				if len(parts) < 3 {
					return match
				}

				text := parts[1]
				path := parts[2]

				if isLocalPath(path) {
					path = resolveLocalPath(path, docPath)
				}
				// Only absolute paths now -- resolveLocalPath ought to have
				// made this an absolute local path, if it was a local path at
				// all.
				if strings.HasPrefix(path, "/") {
					if _, err := os.Stat("data/documents" + path); os.IsNotExist(err) {
						return "<span class=\"notfound\">[" + text + "](" + path + ")</span>"
					}
				}

				return "[" + text + "](" + path + ")"
			})
		}
	}

	// Rejoin all sections
	return joinSections(sections)
}

// isLocalPath returns true if the path is a local file reference
func isLocalPath(path string) bool {
	// Skip URLs with schemes (http://, https://, ftp://, etc)
	if strings.Contains(path, "://") {
		return false
	}

	// Skip fragment URLs that start with #
	if strings.HasPrefix(path, "#") {
		return false
	}

	// Skip absolute URLs that start with /
	if strings.HasPrefix(path, "/") {
		return false
	}

	// Skip data: URLs
	if strings.HasPrefix(path, "data:") {
		return false
	}

	// Skip mailto: links
	if strings.HasPrefix(path, "mailto:") {
		return false
	}

	// All other URLs are considered local file references
	return true
}

// resolveLocalPath resolves a local path relative to the document path
func resolveLocalPath(path, docPath string) string {
	// Remove any leading slashes from docPath
	docPath = strings.TrimLeft(docPath, "/")

	// URL encode the filename to handle spaces and special characters
	escapedPath := url.PathEscape(path)

	// Handle the homepage special case
	if docPath == "" || docPath == "/" {
		// Homepage files are stored in "pages/home"
		result := "/api/files/pages/home/" + escapedPath
		// fmt.Printf("[resolveLocalPath] docPath: '%s', path: '%s', result: '%s' (homepage)\n", docPath, path, result)
		return result
	}

	// Regular document files
	result := "/api/files/" + docPath + "/" + escapedPath
	// fmt.Printf("[resolveLocalPath] docPath: '%s', path: '%s', result: '%s'\n", docPath, path, result)
	return result
}
