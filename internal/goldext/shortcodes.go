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
	"sync"
	"time"
	"wiki-go/internal/logger"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

// wikiLocation is the resolved *time.Location for the configured wiki
// timezone. It is guarded by tzMu so callers in main can set it at
// startup while requests format times concurrently.
var (
	tzMu         sync.RWMutex
	wikiTimezone string
	wikiLocation *time.Location // nil means UTC
)

// SetWikiTimezone configures the timezone used by shortcodes that render
// timestamps. It should be called once at startup, after config is loaded.
// Invalid timezone strings are logged once here and fall back to UTC.
func SetWikiTimezone(tz string) {
	tzMu.Lock()
	defer tzMu.Unlock()
	wikiTimezone = tz
	if tz == "" {
		wikiLocation = nil
		return
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		logger.Warn("goldext: invalid wiki timezone %q: %v, falling back to UTC", tz, err)
		wikiLocation = nil
		return
	}
	wikiLocation = loc
}

// formatModTime formats t using the configured wiki timezone. If no
// timezone is configured or the configured value was invalid, it falls
// back to UTC.
func formatModTime(t time.Time, format string) string {
	tzMu.RLock()
	loc := wikiLocation
	tzMu.RUnlock()

	if loc == nil {
		return t.UTC().Format(format)
	}
	return t.In(loc).Format(format)
}

// ShortcodesPreprocessor replaces the text-only year shortcode outside code.
// Stats shortcodes are converted to trusted AST nodes after Markdown parsing.
func ShortcodesPreprocessor(markdown string, _ string) string {
	sections := splitCodeSections(markdown)
	year := strconv.Itoa(time.Now().Year())
	for index := range sections {
		if !sections[index].isCode {
			sections[index].content = strings.ReplaceAll(sections[index].content, ":::year:::", year)
		}
	}
	return joinSections(sections)
}

// Document represents a document in the wiki
type Document struct {
	Title   string    // Document title
	Path    string    // Document path
	ModTime time.Time // Last modified time
}

// StatsMode is a fixed stats shortcode operation.
type StatsMode uint8

const (
	StatsInvalid StatsMode = iota
	StatsDocumentCount
	StatsRecentEdits
)

// StatsBlock is produced from a standalone, validated stats shortcode.
type StatsBlock struct {
	ast.BaseBlock
	Mode        StatsMode
	Folder      string
	RecentCount int
	Original    string
}

// KindStatsBlock is the Goldmark kind for StatsBlock.
var KindStatsBlock = ast.NewNodeKind("WikiGoStatsBlock")

// Kind implements ast.Node.
func (n *StatsBlock) Kind() ast.NodeKind { return KindStatsBlock }

// Dump implements ast.Node.
func (n *StatsBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"Original": n.Original}, nil)
}

var statsShortcodePattern = regexp.MustCompile(`^:::stats\s+(count|recent)=([^:\r\n]+):::$`)
var statsFolderPattern = regexp.MustCompile(`^[A-Za-z0-9_.\-/]+$`)

func parseStatsBlock(value string) (*StatsBlock, bool) {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, ":::stats") {
		return nil, false
	}

	node := &StatsBlock{Original: value}
	match := statsShortcodePattern.FindStringSubmatch(value)
	if len(match) != 3 {
		return node, true
	}
	parameter := strings.TrimSpace(match[2])
	switch match[1] {
	case "count":
		if parameter != "*" && parameter != "all" && !validStatsFolder(parameter) {
			return node, true
		}
		node.Mode = StatsDocumentCount
		node.Folder = parameter
	case "recent":
		count, err := strconv.Atoi(parameter)
		if err != nil || count < 1 || count > 100 {
			return node, true
		}
		node.Mode = StatsRecentEdits
		node.RecentCount = count
	}
	return node, true
}

func validStatsFolder(folder string) bool {
	if folder == "" || strings.HasPrefix(folder, "/") || strings.Contains(folder, "\\") || !statsFolderPattern.MatchString(folder) {
		return false
	}
	for _, segment := range strings.Split(folder, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return false
		}
	}
	return true
}

func (r *trustedNodeRenderer) renderStatsBlock(writer util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	stats := node.(*StatsBlock)
	var output strings.Builder
	switch stats.Mode {
	case StatsDocumentCount:
		if stats.Folder != "*" && stats.Folder != "all" && !validStatsFolder(stats.Folder) {
			return ast.WalkSkipChildren, nil
		}
		renderDocumentCount(&output, stats.Folder)
	case StatsRecentEdits:
		if stats.RecentCount < 1 || stats.RecentCount > 100 {
			return ast.WalkSkipChildren, nil
		}
		renderRecentEdits(&output, stats.RecentCount)
	default:
		_, _ = writer.WriteString("<p>")
		_, _ = writer.Write(util.EscapeHTML([]byte(stats.Original)))
		_, _ = writer.WriteString("</p>\n")
		return ast.WalkSkipChildren, nil
	}
	_, _ = writer.WriteString(output.String())
	return ast.WalkSkipChildren, nil
}

// renderDocumentCount renders the document count HTML
func renderDocumentCount(w *strings.Builder, countParam string) {
	var count int
	var title string
	var description string

	// Only count documents in the documents directory
	docsDir := DocumentsRoot()

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
	w.WriteString("<h4>" + EscapeHTMLText(title) + "</h4>\n")
	w.WriteString("<div class=\"count-container\">\n")
	w.WriteString("<div class=\"count-number\">" + strconv.Itoa(count) + "</div>\n")
	w.WriteString("<div class=\"count-description\">" + EscapeHTMLText(description) + "</div>\n")
	w.WriteString("</div>\n")
	w.WriteString("</div>\n")
}

// renderRecentEdits renders the recent edits HTML
func renderRecentEdits(w *strings.Builder, count int) {
	renderRecentEditsFromDir(w, DocumentsRoot(), count)
}

// renderRecentEditsFromDir is the testable core of renderRecentEdits and
// renders the list of recent edits from a given documents directory.
func renderRecentEditsFromDir(w *strings.Builder, dirPath string, count int) {
	docs := getRecentDocuments(dirPath, count)

	w.WriteString("<div class=\"wiki-stats recent-edits\">\n")
	w.WriteString("<h4>Recently Edited Documents</h4>\n")

	if len(docs) == 0 {
		w.WriteString("<p>No recently edited documents found.</p>\n")
	} else {
		w.WriteString("<ul>\n")

		for _, doc := range docs {
			folderPath := "/" + doc.Path

			w.WriteString("<li>\n")
			w.WriteString("  <div class=\"doc-info\">\n")
			if href, ok := ValidateTrustedURL(folderPath); ok {
				w.WriteString(fmt.Sprintf("    <a href=\"%s\">%s</a>\n", EscapeHTMLText(href), EscapeHTMLText(doc.Title)))
			} else {
				w.WriteString("    <span>" + EscapeHTMLText(doc.Title) + "</span>\n")
			}
			w.WriteString(fmt.Sprintf("    <span class=\"doc-path\">%s</span>\n", EscapeHTMLText(folderPath)))
			w.WriteString("  </div>\n")
			w.WriteString(fmt.Sprintf("  <span class=\"edit-date\">%s</span>\n", formatModTime(doc.ModTime, "2006-01-02 15:04")))
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
