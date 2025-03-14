package utils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"

	mdext "wiki-go/internal/markdown"
)

// preprocessMarkdown fixes common issues with markdown before parsing
func preprocessMarkdown(input string) string {
	// Apply typographic replacements
	input = mdext.ApplyTypographicReplacements(input)

	// Process stats shortcodes
	input = processStatsShortcodes(input)

	lines := strings.Split(input, "\n")
	result := make([]string, 0, len(lines))

	// Track list structure
	listItems := make([]bool, len(lines))
	listIndentations := make([]string, len(lines))
	inList := false
	inCodeBlock := false
	inDetailsBlock := false
	inTable := false
	lastLineWasList := false
	lastLineWasText := false
	listIndent := ""
	tableIndent := ""
	codeBlockInList := false
	currentListIndent := ""

	// First pass: identify list items
	for i, line := range lines {
		isList := regexp.MustCompile(`^\s*[0-9]+\.\s`).MatchString(line) ||
			regexp.MustCompile(`^\s*[-*+]\s`).MatchString(line)

		if isList {
			listItems[i] = true
			listIndentations[i] = line[:len(line)-len(strings.TrimLeft(line, " \t"))]
		}

		// Pre-process task list items and convert them to HTML
		if listItems[i] {
			trimmedLine := strings.TrimSpace(line)
			uncheckedRegex := regexp.MustCompile(`^[-*+]\s+\[\s*\]\s+(.*)$`)
			checkedRegex := regexp.MustCompile(`^[-*+]\s+\[(x|X)\]\s+(.*)$`)

			if matches := uncheckedRegex.FindStringSubmatch(trimmedLine); len(matches) > 1 {
				// Found an unchecked task list item
				indent := listIndentations[i]
				taskText := matches[1]

				// Calculate the list nesting level based on indentation
				indentLevel := len(indent) / 2

				// Replace the original list item with a task list item but preserve the indentation structure
				if indentLevel > 0 {
					// This is a nested list item, preserve the class to allow proper CSS nesting
					line = indent + "<li class=\"task-list-item-container\" style=\"list-style-type: none;\" data-indent-level=\"" +
						strconv.Itoa(indentLevel) + "\"><span class=\"task-list-item\"><input type=\"checkbox\" class=\"task-checkbox\" disabled> <span class=\"task-text\">" +
						taskText + "</span></span></li>"
				} else {
					// Top level list item
					line = "<li class=\"task-list-item-container\" style=\"list-style-type: none;\"><span class=\"task-list-item\"><input type=\"checkbox\" class=\"task-checkbox\" disabled> <span class=\"task-text\">" +
						taskText + "</span></span></li>"
				}
				lines[i] = line
			} else if matches := checkedRegex.FindStringSubmatch(trimmedLine); len(matches) > 2 {
				// Found a checked task list item
				indent := listIndentations[i]
				taskText := matches[2]

				// Calculate the list nesting level based on indentation
				indentLevel := len(indent) / 2

				// Replace the original list item with a task list item but preserve the indentation structure
				if indentLevel > 0 {
					// This is a nested list item, preserve the class to allow proper CSS nesting
					line = indent + "<li class=\"task-list-item-container\" style=\"list-style-type: none;\" data-indent-level=\"" +
						strconv.Itoa(indentLevel) + "\"><span class=\"task-list-item\"><input type=\"checkbox\" class=\"task-checkbox\" checked disabled> <span class=\"task-text\">" +
						taskText + "</span></span></li>"
				} else {
					// Top level list item
					line = "<li class=\"task-list-item-container\" style=\"list-style-type: none;\"><span class=\"task-list-item\"><input type=\"checkbox\" class=\"task-checkbox\" checked disabled> <span class=\"task-text\">" +
						taskText + "</span></span></li>"
				}
				lines[i] = line
			}
		}
	}

	// Second pass: process the lines with knowledge of list structure
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		// Check for code block start/end
		if strings.HasPrefix(trimmedLine, "```") {
			if !inCodeBlock {
				// Start of a code block
				inCodeBlock = true

				// Check if this is a special block type - we keep the detection
				// but don't need to track it anymore since we don't do special indentation
				if len(trimmedLine) > 3 {
					codeInfo := trimmedLine[3:]
					// Track if we're in a details block specially
					if strings.HasPrefix(codeInfo, "details") {
						inDetailsBlock = true
					}
					_ = strings.HasPrefix(codeInfo, "youtube") ||
						strings.HasPrefix(codeInfo, "vimeo") ||
						strings.HasPrefix(codeInfo, "mp4") ||
						strings.HasPrefix(codeInfo, "mermaid")
				}

				// Check if the previous line was a list item (we're inside a list)
				if lastLineWasList && !inTable {
					codeBlockInList = true
					currentListIndent = listIndent + "    " // Add 4 spaces to indent code block inside list item

					// Remove any blank lines between list item and code block
					for len(result) > 0 && strings.TrimSpace(result[len(result)-1]) == "" {
						result = result[:len(result)-1]
					}

					// Add the code block opening with proper indentation
					result = append(result, currentListIndent+trimmedLine)
				} else {
					// Regular code block, not in a list
					result = append(result, trimmedLine)
				}
				continue
			} else {
				// End of a code block
				inCodeBlock = false
				if codeBlockInList {
					// Close the code block with proper indentation
					result = append(result, currentListIndent+trimmedLine)
					codeBlockInList = false
					currentListIndent = ""
				} else {
					result = append(result, trimmedLine)
				}

				// Reset details block tracking
				inDetailsBlock = false
				continue
			}
		}

		// Check if this is a list item
		if listItems[i] {
			// If this is a list item following a paragraph inside a details block,
			// insert an empty line to ensure proper list formatting
			if lastLineWasText && inDetailsBlock && !lastLineWasList {
				result = append(result, "")
			}

			inList = true
			lastLineWasList = true
			lastLineWasText = false
			listIndent = listIndentations[i]

			// Add the list item to the result
			result = append(result, line)
			continue
		}

		// Process lines inside a code block
		if inCodeBlock {
			// Check if we're inside a list and need to indent
			if codeBlockInList {
				// Add proper indentation to code content to keep it inside list
				result = append(result, currentListIndent+line)
			} else {
				// No special indentation needed
				result = append(result, line)
			}

			// Check if this is a line of text that could be followed by a list
			// in a details block
			if inDetailsBlock && trimmedLine != "" {
				// Check if this is a table line
				isTableLine := strings.Contains(trimmedLine, "|") &&
					(strings.HasPrefix(trimmedLine, "|") || strings.HasSuffix(trimmedLine, "|"))

				// Check if the line looks like regular text (not a heading, list, etc.)
				isHeading := strings.HasPrefix(trimmedLine, "#")
				isHorizontalRule := regexp.MustCompile(`^[-*_]{3,}$`).MatchString(trimmedLine)
				isLinkDefinition := regexp.MustCompile(`^\[.+\]:\s+`).MatchString(trimmedLine)

				// If it's regular text, mark it for potential list detection later
				if !isHeading && !isHorizontalRule && !isLinkDefinition && !isTableLine {
					lastLineWasText = true
				} else {
					lastLineWasText = false
				}
			}

			continue
		}

		// Check for table start or continuation
		isTableLine := strings.Contains(trimmedLine, "|") &&
			(strings.HasPrefix(trimmedLine, "|") || strings.HasSuffix(trimmedLine, "|"))

		// Special handling for tables in lists
		if isTableLine && inList {
			// If this is the first line of a table after a list item, add an empty line
			if !inTable && lastLineWasList {
				// Add an empty line before the table to ensure proper parsing
				result = append(result, "")
			}

			// Start or continue a table
			inTable = true

			// For tables in lists, we need to indent them properly
			// The indentation should be the list item indentation + 4 spaces
			tableIndent = listIndent + "    "

			// Add the table line with proper indentation
			result = append(result, tableIndent+trimmedLine)
			continue
		} else if inTable && !isTableLine {
			// End of table
			inTable = false

			// Add an empty line after the table for better separation
			result = append(result, "")
		}

		// Reset lastLineWasList if this line is not a list item and not in a code block or table
		if !listItems[i] && !inCodeBlock && !inTable {
			lastLineWasList = false
		}

		// Check if we're exiting a list
		if inList && trimmedLine == "" {
			// Empty line might be the end of a list
			// We'll check the next non-empty line to confirm
			nextNonEmptyLine := ""
			nextNonEmptyLineIndex := -1
			for j := i + 1; j < len(lines); j++ {
				if strings.TrimSpace(lines[j]) != "" {
					nextNonEmptyLine = lines[j]
					nextNonEmptyLineIndex = j
					break
				}
			}

			// If the next non-empty line is not a list item or table, we're exiting the list
			if nextNonEmptyLine != "" && !listItems[nextNonEmptyLineIndex] &&
				!strings.Contains(strings.TrimSpace(nextNonEmptyLine), "|") {
				inList = false
				inTable = false
			}
		}

		// Add the line to the result
		result = append(result, line)
	}

	// Join the lines back into a single string
	return strings.Join(result, "\n")
}

// processStatsShortcodes pre-processes stats shortcodes in the markdown
// It replaces them directly with the rendered HTML
func processStatsShortcodes(input string) string {
	// Define a regex to match stats shortcodes
	statsRegex := regexp.MustCompile(`:::stats\s+(.*?):::`)

	// Replace each shortcode with its rendered HTML
	processed := statsRegex.ReplaceAllStringFunc(input, func(match string) string {
		// Extract the parameters
		paramText := statsRegex.FindStringSubmatch(match)[1]

		// Parse the parameters
		paramParts := strings.Fields(paramText)
		statsParams := make(map[string]string)

		for _, part := range paramParts {
			if strings.Contains(part, "=") {
				kv := strings.SplitN(part, "=", 2)
				statsParams[kv[0]] = kv[1]
			}
		}

		// Create a buffer to hold the stats HTML
		var buf bytes.Buffer

		// Render the stats
		mdext.RenderStats(&buf, statsParams)

		// Return the rendered HTML wrapped in a div to protect it from markdown processing
		return fmt.Sprintf("<div class=\"stats-shortcode-wrapper\">%s</div>", buf.String())
	})

	return processed
}

// RenderMarkdownFile reads a markdown file and returns its HTML representation
func RenderMarkdownFile(filePath string) ([]byte, error) {
	// Read the markdown file
	mdContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Get the directory path for the document
	docDir := filepath.Dir(filePath)

	// Convert to a relative path for URL construction
	relPath, err := filepath.Rel(filepath.Join("data", "documents"), docDir)
	if err != nil {
		// If we can't get a relative path, just use the directory name
		relPath = filepath.Base(docDir)
	}

	// Replace backslashes with forward slashes for URLs
	relPath = strings.ReplaceAll(relPath, "\\", "/")

	// Use the path-aware rendering function
	return RenderMarkdownWithPath(string(mdContent), relPath), nil
}

// RenderMarkdown converts markdown text to HTML
func RenderMarkdown(md string) []byte {
	return RenderMarkdownWithPath(md, "")
}

// RenderMarkdownWithPath converts markdown text to HTML with the current document path
func RenderMarkdownWithPath(md string, docPath string) []byte {
	// Preprocess the markdown to fix common issues
	preprocessed := preprocessMarkdown(md)

	// Create markdown extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs |
		parser.Strikethrough | // Enable ~~strikethrough~~
		parser.Footnotes | // Enable footnotes
		parser.SuperSubscript | // Enable super^script^ and sub~script~
		parser.Mmark |
		parser.MathJax |
		parser.NoEmptyLineBeforeBlock // Help with nested blocks in lists

	// Create the parser
	p := parser.NewWithExtensions(extensions)

	// Set the current document path for link resolution
	mdext.CurrentDocumentPath = docPath

	// Parse the markdown text
	doc := p.Parse([]byte(preprocessed))

	// Create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank | html.CompletePage
	// Disable SmartyPants to prevent backticks from being converted to curly quotes
	htmlFlags &= ^html.Smartypants

	// Force tight lists (no paragraphs in list items) even when there are code blocks
	// Note: We need to implement this ourselves since the flag isn't available

	opts := html.RendererOptions{
		Flags:          htmlFlags,
		RenderNodeHook: mdext.RenderNodeHook,
	}
	renderer := html.NewRenderer(opts)

	// Render to HTML
	return markdown.Render(doc, renderer)
}
