package goldext

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"
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

// splitCodeSections splits Markdown into regular text and code/math sections.
func splitCodeSections(markdown string) []Section {
	var sections []Section

	// Protect both supported fence styles plus inline code and math.
	codeBlockPattern := "(?:```[\\s\\S]*?```|~~~[\\s\\S]*?~~~)"
	inlineCodePattern := "`[^`]*?`"
	inlineMathPattern := "\\$[^\\$\\n]+?\\$"
	blockMathPattern := "\\$\\$[\\s\\S]*?\\$\\$"

	// Combine patterns to find all protected sections
	combinedPattern := fmt.Sprintf("(?s)(%s|%s|%s|%s)",
		codeBlockPattern, inlineCodePattern, inlineMathPattern, blockMathPattern)

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

var (
	imgLinkRe     = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
	regularLinkRe = regexp.MustCompile(`\[([^\]]*)\]\(([^)]+)\)`)
)

// LinkPreprocessor resolves local attachment references while returning only
// Markdown. It never emits HTML wrappers for missing files.
func LinkPreprocessor(markdown string, docPath string) string {
	sections := splitCodeSections(markdown)
	for i := range sections {
		if sections[i].isCode {
			continue
		}
		sections[i].content = imgLinkRe.ReplaceAllStringFunc(sections[i].content, func(match string) string {
			parts := imgLinkRe.FindStringSubmatch(match)
			if len(parts) < 3 {
				return match
			}
			path, title := splitMarkdownLinkTitle(parts[2])
			if destination, ok := safeMarkdownDestination(path, docPath); ok {
				path = destination
			} else {
				path = "#"
			}
			return "![" + escapeMarkdownHTML(parts[1]) + "](" + path + title + ")"
		})
		sections[i].content = regularLinkRe.ReplaceAllStringFunc(sections[i].content, func(match string) string {
			parts := regularLinkRe.FindStringSubmatch(match)
			if len(parts) < 3 {
				return match
			}
			path, title := splitMarkdownLinkTitle(parts[2])
			if destination, ok := safeMarkdownDestination(path, docPath); ok {
				path = destination
			} else {
				path = "#"
			}
			return "[" + escapeMarkdownHTML(parts[1]) + "](" + path + title + ")"
		})
	}
	return joinSections(sections)
}

func escapeMarkdownHTML(value string) string {
	return strings.NewReplacer("<", "&lt;", ">", "&gt;").Replace(value)
}

func safeMarkdownDestination(value, docPath string) (string, bool) {
	if resolved, ok := resolveLocalAttachmentReference(value, docPath); ok {
		return resolved, true
	}
	return ValidateTrustedURL(value)
}

func splitMarkdownLinkTitle(rawPath string) (path, title string) {
	path = rawPath
	if index := strings.Index(rawPath, " \""); index != -1 {
		return rawPath[:index], rawPath[index:]
	}
	if index := strings.Index(rawPath, " '"); index != -1 {
		return rawPath[:index], rawPath[index:]
	}
	return path, ""
}

// isLocalPath returns true if the path is a local file reference
func isLocalPath(path string) bool {
	_, ok := localAttachmentParts(path)
	return ok
}

// resolveLocalPath resolves a local path relative to the document path
func resolveLocalPath(path, docPath string) string {
	if resolved, ok := resolveLocalAttachmentReference(path, docPath); ok {
		return resolved
	}
	return path
}

func resolveLocalAttachmentReference(value, docPath string) (string, bool) {
	parts, ok := localAttachmentParts(value)
	if !ok {
		return "", false
	}
	documentPath, ok := attachmentDocumentURLPath(docPath)
	if !ok {
		return "", false
	}
	result := "/api/files/" + documentPath + "/" + url.PathEscape(parts[0])
	if parts[1] != "" {
		result += "#" + url.PathEscape(parts[1])
	}
	return result, true
}

func localAttachmentParts(value string) ([2]string, bool) {
	var result [2]string
	value = strings.TrimSpace(value)
	if value == "" || strings.HasPrefix(value, "/") || strings.HasPrefix(value, "#") ||
		strings.ContainsAny(value, `\\?:`) {
		return result, false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return result, false
		}
	}
	path := value
	if index := strings.IndexByte(value, '#'); index != -1 {
		path, result[1] = value[:index], value[index+1:]
	}
	if path == "" || strings.Contains(path, "/") || path == "." || path == ".." {
		return result, false
	}
	result[0] = path
	return result, true
}
