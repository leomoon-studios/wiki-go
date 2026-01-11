package goldext

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ShortcodesPreprocessor processes shortcodes in markdown text
// Supports: :::year:::, :::stats count=*:::, :::stats recent=N:::
// Avoids processing shortcodes inside code blocks
func ShortcodesPreprocessor(markdown string, _ string) string {
	// Split markdown into lines for processing
	lines := strings.Split(markdown, "\n")
	processedLines := make([]string, 0, len(lines))

	// State tracking for code blocks
	inBacktickBlock := false
	inTildeBlock := false

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

		// Skip processing if in a code block
		if inBacktickBlock || inTildeBlock {
			processedLines = append(processedLines, line)
			continue
		}

		// Process shortcodes with respect to inline code blocks
		if strings.Contains(line, ":::year:::") || strings.Contains(line, ":::stats") {
			// Process each segment of the line, preserving inline code
			var processedLine string
			segments := strings.Split(line, "`")

			for i, segment := range segments {
				// Even segments (0, 2, 4...) are outside inline code
				if i%2 == 0 {
					// Process :::year::: shortcode
					if strings.Contains(segment, ":::year:::") {
						currentYear := strconv.Itoa(time.Now().Year())
						segment = strings.ReplaceAll(segment, ":::year:::", currentYear)
					}

					// Match stats shortcode pattern
					if strings.Contains(segment, ":::stats") {
						statsRegex := regexp.MustCompile(`:::stats\s+(recent|count)=([^:]+):::`)
						segment = statsRegex.ReplaceAllStringFunc(segment, func(match string) string {
							params := statsRegex.FindStringSubmatch(match)
							if len(params) < 3 {
								return match
							}

							shortcodeType := params[1]
							shortcodeValue := params[2]

							if shortcodeType == "count" {
								var buf strings.Builder
								renderDocumentCount(&buf, shortcodeValue)
								return buf.String()
							} else if shortcodeType == "recent" {
								count, err := strconv.Atoi(shortcodeValue)
								if err != nil || count <= 0 {
									count = 5 // Default to 5 if invalid
								}
								var buf strings.Builder
								renderRecentEdits(&buf, count)
								return buf.String()
							}

							return match
						})
					}
					processedLine += segment
				} else {
					// Odd segments (1, 3, 5...) are inside inline code - preserve them
					processedLine += "`" + segment + "`"
				}
			}

			processedLines = append(processedLines, processedLine)
		} else {
			processedLines = append(processedLines, line)
		}
	}

	return strings.Join(processedLines, "\n")
}

// Document represents a document in the wiki
type Document struct {
	Title   string    // Document title
	Path    string    // Document path
	ModTime time.Time // Last modified time
}

// renderDocumentCount renders the document count HTML
func renderDocumentCount(w *strings.Builder, countParam string) {
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

	// Generate HTML for the document count
	w.WriteString("<div class=\"wiki-stats doc-count\">\n")
	w.WriteString("<h4>" + title + "</h4>\n")
	w.WriteString("<div class=\"count-container\">\n")
	w.WriteString("<div class=\"count-number\">" + strconv.Itoa(count) + "</div>\n")
	w.WriteString("<div class=\"count-description\">" + description + "</div>\n")
	w.WriteString("</div>\n")
	w.WriteString("</div>\n")
}

// renderRecentEdits renders the recent edits HTML
func renderRecentEdits(w *strings.Builder, count int) {
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

			// Structure with elements on one line
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
