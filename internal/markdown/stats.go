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

// Regular expression to match the stats shortcode with parameters
// Format: :::stats param1=value1 param2=value2:::
var statsRegex = regexp.MustCompile(`^\s*:::stats\s+(.*?):::\s*$`)
var paramRegex = regexp.MustCompile(`(\w+)=(\w+)`)

// ProcessStatsBlock processes a code block that contains a stats shortcode
// Returns true if the block was processed as a stats section, false otherwise
func ProcessStatsBlock(w io.Writer, codeBlock *ast.CodeBlock) bool {
	// Check if it's a text block (not a code block with language)
	if len(codeBlock.Info) > 0 {
		return false
	}

	// Convert the content to string and check if it matches our pattern
	content := string(codeBlock.Literal)

	// Check if the content matches our pattern
	matches := statsRegex.FindStringSubmatch(content)
	if matches == nil {
		return false
	}

	// Extract parameters
	params := parseParameters(matches[1])

	// Process the stats based on parameters
	renderStats(w, params)

	return true
}

// ProcessStatsParagraph processes a paragraph that contains a stats shortcode
// Returns true if the paragraph was processed as a stats section, false otherwise
func ProcessStatsParagraph(w io.Writer, para *ast.Paragraph) bool {
	// We need to check if this paragraph contains only our shortcode
	if len(para.Children) != 1 {
		return false
	}

	// Check if the only child is a text node
	textNode, ok := para.Children[0].(*ast.Text)
	if !ok {
		return false
	}

	// Convert the content to string and check if it matches our pattern
	content := string(textNode.Literal)

	// Check if the content matches our pattern
	matches := statsRegex.FindStringSubmatch(content)
	if matches == nil {
		return false
	}

	// Extract parameters
	params := parseParameters(matches[1])

	// Process the stats based on parameters
	renderStats(w, params)

	// Important: Set the text node's literal to empty to remove the shortcode from output
	textNode.Literal = nil

	return true
}

// parseParameters extracts parameter key-value pairs from the shortcode
func parseParameters(paramString string) map[string]string {
	params := make(map[string]string)

	// Find all parameter matches
	matches := paramRegex.FindAllStringSubmatch(paramString, -1)
	for _, match := range matches {
		if len(match) == 3 {
			params[match[1]] = match[2]
		}
	}

	return params
}

// renderStats renders the appropriate stats based on parameters
func renderStats(w io.Writer, params map[string]string) {
	// Check for "recent" parameter to show recently edited documents
	if recentStr, ok := params["recent"]; ok {
		count, err := strconv.Atoi(recentStr)
		if err != nil || count <= 0 {
			// Default to 5 if invalid
			count = 5
		}
		renderRecentEditsStats(w, count)
		return
	}

	// Check for "count" parameter to show document count
	if countParam, ok := params["count"]; ok {
		renderDocumentCount(w, countParam)
		return
	}

	// If no recognized parameters, show a default message
	w.Write([]byte("<div class=\"wiki-stats-error\">No valid stats parameters specified. Try 'recent=5' or 'count=all'.</div>"))
}

// renderRecentEdits renders the recent edits HTML
func renderRecentEditsStats(w io.Writer, count int) {
	// Get recently edited documents
	recentDocs, err := getRecentlyEditedDocsForStats("data/documents", count)
	if err != nil {
		// If there's an error, show an error message
		w.Write([]byte("<div class=\"wiki-stats-error\">Error retrieving recent documents: " + err.Error() + "</div>"))
		return
	}

	// Generate HTML for the recent edits
	w.Write([]byte("<div class=\"wiki-stats recent-edits\">\n"))
	w.Write([]byte("<h4>Recently Edited Documents</h4>\n"))

	if len(recentDocs) == 0 {
		w.Write([]byte("<p>No recently edited documents found.</p>\n"))
	} else {
		w.Write([]byte("<ul>\n"))

		for _, doc := range recentDocs {
			// Create link to the document's folder, not the .md file itself
			folderPath := "/" + doc.FolderPath

			w.Write([]byte(fmt.Sprintf("<li><a href=\"%s\">%s</a> <span class=\"edit-date\">%s</span></li>\n",
				folderPath,
				doc.Title,
				doc.LastModified.Format("2006-01-02 15:04"))))
		}

		w.Write([]byte("</ul>\n"))
	}

	w.Write([]byte("</div>\n"))
}

// renderDocumentCount renders the document count HTML
func renderDocumentCount(w io.Writer, countType string) {
	var count int
	var err error
	var title string
	var description string

	// Determine what to count based on the parameter
	switch countType {
	case "all":
		count, err = countAllDocuments("data/documents")
		title = "Total Documents"
		description = "Total number of documents in the wiki"
	default:
		// Try to interpret as a specific folder path
		count, err = countDocumentsInFolder("data/documents/" + countType)
		title = "Documents in " + formatDirNameForStats(countType)
		description = "Number of documents in the " + formatDirNameForStats(countType) + " section"
	}

	if err != nil {
		w.Write([]byte("<div class=\"wiki-stats-error\">Error counting documents: " + err.Error() + "</div>"))
		return
	}

	// Generate HTML for the document count
	w.Write([]byte("<div class=\"wiki-stats doc-count\">\n"))
	w.Write([]byte("<h4>" + title + "</h4>\n"))
	w.Write([]byte("<div class=\"count-container\">\n"))
	w.Write([]byte("<div class=\"count-number\">" + strconv.Itoa(count) + "</div>\n"))
	w.Write([]byte("<div class=\"count-description\">" + description + "</div>\n"))
	w.Write([]byte("</div>\n"))
	w.Write([]byte("</div>\n"))
}

// countAllDocuments counts all document.md files in the wiki
func countAllDocuments(rootDir string) (int, error) {
	count := 0

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Count only document.md files
		if !info.IsDir() && filepath.Base(path) == "document.md" {
			count++
		}

		return nil
	})

	return count, err
}

// countDocumentsInFolder counts document.md files in a specific folder
func countDocumentsInFolder(folderPath string) (int, error) {
	count := 0

	// Check if the folder exists
	_, err := os.Stat(folderPath)
	if os.IsNotExist(err) {
		return 0, fmt.Errorf("folder does not exist: %s", folderPath)
	}

	err = filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Count only document.md files
		if !info.IsDir() && filepath.Base(path) == "document.md" {
			count++
		}

		return nil
	})

	return count, err
}

// DocumentInfo represents information about a document
type DocumentInfoStats struct {
	Title        string
	FolderPath   string
	LastModified time.Time
}

// getRecentlyEditedDocuments retrieves the last n edited documents
func getRecentlyEditedDocsForStats(rootDir string, count int) ([]DocumentInfoStats, error) {
	var documents []DocumentInfoStats

	// Walk through the documents directory
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process document.md files
		if !info.IsDir() && filepath.Base(path) == "document.md" {
			// Get the folder path (relative to rootDir)
			relPath, err := filepath.Rel(rootDir, filepath.Dir(path))
			if err != nil {
				return err
			}

			// Replace backslashes with forward slashes for URLs
			relPath = strings.ReplaceAll(relPath, "\\", "/")

			// Extract document title from the file
			title := extractDocumentTitleForStats(path)
			if title == "" {
				// Use folder name as fallback
				title = formatDirNameForStats(filepath.Base(filepath.Dir(path)))
			}

			// Add to documents list
			documents = append(documents, DocumentInfoStats{
				Title:        title,
				FolderPath:   relPath,
				LastModified: info.ModTime(),
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort by last modified time (newest first)
	sort.Slice(documents, func(i, j int) bool {
		return documents[i].LastModified.After(documents[j].LastModified)
	})

	// Limit to requested count
	if len(documents) > count {
		documents = documents[:count]
	}

	return documents, nil
}

// extractDocumentTitle extracts the first H1 title from a markdown file
func extractDocumentTitleForStats(filePath string) string {
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
func formatDirNameForStats(name string) string {
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
