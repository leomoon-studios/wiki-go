package markdown

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gomarkdown/markdown/ast"
)

// Stats package contains functionality for generating document statistics
// This file has been cleared for a fresh implementation

// Regular expression to match the stats shortcodes
var statsShortcodeRegex = regexp.MustCompile(`:::stats\s+(.*?):::`)
var statsParamRegex = regexp.MustCompile(`(\w+)=([*\w-]+)`)

// ProcessStatsCodeBlock processes a code block that might contain stats shortcodes
// Returns true if processed, false otherwise
func ProcessStatsCodeBlock(w io.Writer, codeBlock *ast.CodeBlock) bool {
	// Only process code blocks without language specification
	if len(codeBlock.Info) > 0 {
		return false
	}

	// Get the code block content
	content := string(codeBlock.Literal)

	// Check if this contains a stats shortcode
	if !statsShortcodeRegex.MatchString(content) {
		return false
	}

	// Process the content
	matches := statsShortcodeRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			// Extract parameters
			params := parseStatsParams(match[1])

			// Process based on the parameters
			if countParam, ok := params["count"]; ok {
				renderDocumentCount(w, countParam)
			} else if recentParam, ok := params["recent"]; ok {
				count, err := strconv.Atoi(recentParam)
				if err != nil || count <= 0 {
					count = 5 // Default to 5 if invalid
				}
				renderRecentEdits(w, count)
			} else {
				// No valid parameters
				w.Write([]byte("<div class=\"wiki-stats-error\">Invalid stats shortcode parameters. Use count=* or recent=N.</div>"))
			}
		}
	}

	return true
}

// ProcessStatsText processes stats shortcodes in text nodes
// Returns true if processed, false otherwise, and the remaining text
func ProcessStatsText(w io.Writer, text string) (bool, string) {
	// Check if the text contains a stats shortcode
	if !statsShortcodeRegex.MatchString(text) {
		return false, text
	}

	// We'll track if we processed any shortcodes
	processed := false

	// We'll build the result text without the shortcodes
	var resultBuilder strings.Builder
	lastEnd := 0

	// Find all matches
	matches := statsShortcodeRegex.FindAllStringSubmatchIndex(text, -1)

	for _, match := range matches {
		// Add the text before this match to the result
		resultBuilder.WriteString(text[lastEnd:match[0]])

		// Get the parameter text
		paramText := text[match[2]:match[3]]

		// Parse parameters
		params := parseStatsParams(paramText)

		// Process based on the parameters
		if countParam, ok := params["count"]; ok {
			renderDocumentCount(w, countParam)
		} else if recentParam, ok := params["recent"]; ok {
			count, err := strconv.Atoi(recentParam)
			if err != nil || count <= 0 {
				count = 5 // Default to 5 if invalid
			}
			renderRecentEdits(w, count)
		} else {
			// No valid parameters
			w.Write([]byte("<div class=\"wiki-stats-error\">Invalid stats shortcode parameters. Use count=* or recent=N.</div>"))
		}

		// Update the last end position and mark as processed
		lastEnd = match[1]
		processed = true
	}

	// Add any remaining text
	if lastEnd < len(text) {
		resultBuilder.WriteString(text[lastEnd:])
	}

	return processed, resultBuilder.String()
}

// ProcessStatsTextWithTracking processes a text node for stats shortcodes
// It returns true if any shortcodes were processed, along with the remaining text
// The tracking map ensures each shortcode is only processed once per document
func ProcessStatsTextWithTracking(w io.Writer, text string, processed map[string]bool) (bool, string) {
	if !strings.Contains(text, ":::stats") {
		return false, text
	}

	// Used to keep track of whether we found any shortcuts
	wasProcessed := false

	// Build the result without processed shortcodes
	var result strings.Builder

	// Regex to find stats shortcodes
	re := regexp.MustCompile(`:::stats\s+(count|recent)=([^:]+):::`)
	matches := re.FindAllStringSubmatchIndex(text, -1)

	lastIndex := 0
	for _, match := range matches {
		// Extract the full shortcode from the text using the match indices
		shortcode := text[match[0]:match[1]]

		// Skip if this shortcode has already been processed
		if processed[shortcode] {
			// Add text up to and including the shortcode
			result.WriteString(text[lastIndex:match[1]])
			lastIndex = match[1]
			continue
		}

		// Mark this shortcode as processed
		processed[shortcode] = true
		wasProcessed = true

		// Add text before the shortcode
		result.WriteString(text[lastIndex:match[0]])

		// Process the shortcode
		shortcodeType := text[match[2]:match[3]]
		shortcodeValue := text[match[4]:match[5]]

		// Process the shortcode based on its type
		if shortcodeType == "count" {
			renderDocumentCount(w, shortcodeValue)
		} else if shortcodeType == "recent" {
			count, _ := strconv.Atoi(shortcodeValue)
			if count <= 0 {
				count = 5 // Default limit
			}
			renderRecentEdits(w, count)
		}

		// Update the last index
		lastIndex = match[1]
	}

	// Add remaining text
	result.WriteString(text[lastIndex:])

	return wasProcessed, result.String()
}

// parseStatsParams extracts parameters from a stats shortcode
func parseStatsParams(paramString string) map[string]string {
	params := make(map[string]string)

	// Find all parameter matches
	matches := statsParamRegex.FindAllStringSubmatch(paramString, -1)
	for _, match := range matches {
		if len(match) == 3 {
			params[match[1]] = match[2]
		}
	}

	return params
}

// Document represents a document in the wiki
type Document struct {
	Title   string    // Document title
	Path    string    // Document path
	ModTime time.Time // Last modified time
}

// renderDocumentCount renders the document count HTML
func renderDocumentCount(w io.Writer, countParam string) {
	var buf strings.Builder
	renderDocumentCountToString(&buf, countParam)
	w.Write([]byte(buf.String()))
}

// renderDocumentCountToString renders the document count to a string builder
func renderDocumentCountToString(w *strings.Builder, countParam string) {
	var count int
	var title string
	var description string

	// Only count documents in the documents directory
	docsDir := "data/documents"

	// Count all documents (using * or all as wildcard)
	if countParam == "*" || countParam == "all" {
		count = countDocuments(docsDir)
		title = "Total Documents"
		description = "Total number of documents in the wiki"
	} else {
		// Count documents in a specific folder
		folderPath := filepath.Join(docsDir, countParam)
		count = countDocuments(folderPath)
		title = "Documents in " + formatDirName(countParam)
		description = "Number of documents in the " + formatDirName(countParam) + " section"
	}

	// Generate HTML for the document count - make it more compact
	w.WriteString("<div class=\"wiki-stats doc-count\">\n")
	w.WriteString("<h4>" + title + "</h4>\n")
	w.WriteString("<div class=\"count-container\">\n")
	w.WriteString("<div class=\"count-number\">" + strconv.Itoa(count) + "</div>\n")
	w.WriteString("<div class=\"count-description\">" + description + "</div>\n")
	w.WriteString("</div>\n")
	w.WriteString("</div>\n")
}

// renderRecentEdits renders the recent edits HTML
func renderRecentEdits(w io.Writer, count int) {
	var buf strings.Builder
	renderRecentEditsToString(&buf, count)
	w.Write([]byte(buf.String()))
}

// renderRecentEditsToString renders the recent edits to a string builder
func renderRecentEditsToString(w *strings.Builder, count int) {
	// Get recent documents
	docs := getRecentDocuments("data/documents", count)

	// Generate HTML for the recent edits
	w.WriteString("<div class=\"wiki-stats recent-edits\">\n")
	w.WriteString("<h4>Recently Edited Documents</h4>\n")

	if len(docs) == 0 {
		w.WriteString("<p>No recently edited documents found.</p>\n")
	} else {
		w.WriteString("<ul>\n")

		for _, doc := range docs {
			// Create link to the document's folder
			folderPath := "/" + doc.Path

			// New structure with elements on one line
			w.WriteString("<li>\n")
			w.WriteString("  <div class=\"doc-info\">\n")
			w.WriteString(fmt.Sprintf("    <a href=\"%s\">%s</a>\n", folderPath, doc.Title))
			w.WriteString(fmt.Sprintf("    <span class=\"doc-path\">%s</span>\n", folderPath))
			w.WriteString("  </div>\n")
			w.WriteString(fmt.Sprintf("  <span class=\"edit-date\">%s</span>\n", doc.ModTime.Format("2006-01-02 15:04")))
			w.WriteString("</li>\n")
		}

		w.WriteString("</ul>\n")
	}

	w.WriteString("</div>\n")
}

// countDocuments counts the number of document.md files in a directory
func countDocuments(dirPath string) int {
	count := 0

	// Check if the directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return 0
	}

	// Walk through the directory
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Count only document.md files
		if !info.IsDir() && filepath.Base(path) == "document.md" {
			count++
		}

		return nil
	})

	return count
}

// getRecentDocuments returns the most recently modified documents
func getRecentDocuments(dirPath string, count int) []Document {
	var docs []Document

	// Check if the directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return docs
	}

	// Walk through the directory
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Process only document.md files
		if !info.IsDir() && filepath.Base(path) == "document.md" {
			// Get the document directory
			docDir := filepath.Dir(path)

			// Get the relative path from the documents directory
			relPath, err := filepath.Rel(dirPath, docDir)
			if err != nil {
				relPath = filepath.Base(docDir)
			}

			// Replace backslashes with forward slashes for URLs
			relPath = strings.ReplaceAll(relPath, "\\", "/")

			// Extract the document title
			title := extractDocumentTitle(path)
			if title == "" {
				title = formatDirName(filepath.Base(docDir))
			}

			// Add to the documents list
			docs = append(docs, Document{
				Title:   title,
				Path:    relPath,
				ModTime: info.ModTime(),
			})
		}

		return nil
	})

	// Sort by modification time (newest first)
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].ModTime.After(docs[j].ModTime)
	})

	// Limit to the requested count
	if len(docs) > count {
		docs = docs[:count]
	}

	return docs
}

// extractDocumentTitle extracts the first H1 title from a markdown file
func extractDocumentTitle(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}

	return ""
}

// formatDirName formats a directory name by replacing dashes with spaces and title casing
func formatDirName(name string) string {
	// Replace dashes with spaces
	name = strings.ReplaceAll(name, "-", " ")

	// Title case the words
	words := strings.Fields(name)
	for i, word := range words {
		if len(word) > 0 {
			r := []rune(word)
			r[0] = []rune(strings.ToUpper(string(r[0])))[0]
			words[i] = string(r)
		}
	}

	return strings.Join(words, " ")
}
