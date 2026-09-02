package utils

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"wiki-go/internal/frontmatter"
	"wiki-go/internal/goldext"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

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
	relPath, err := filepath.Rel(goldext.DocumentsRoot(), docDir)
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
	// Check for frontmatter
	metadata, contentWithoutFrontmatter, hasFrontmatter := frontmatter.Parse(md)

	// If this has kanban layout, render as kanban with full goldext support
	if hasFrontmatter && metadata.Layout == "kanban" {
		// Create preprocessor functions (excluding frontmatter since it's already processed)
		var preprocessors []frontmatter.PreprocessorFunc

		// Add all goldext preprocessors (frontmatter will be a no-op since it's already processed)
		for _, preprocessor := range goldext.RegisteredPreprocessors {
			if preprocessor != nil {
				// Create a closure that captures the docPath for kanban rendering
				capturedPreprocessor := preprocessor
				capturedDocPath := docPath
				wrappedPreprocessor := func(md string, _ string) string {
					return capturedPreprocessor(md, capturedDocPath)
				}
				preprocessors = append(preprocessors, wrappedPreprocessor)
			}
		}

		kanbanHTML := frontmatter.RenderKanbanWithProcessors(
			contentWithoutFrontmatter,
			preprocessors,
			nil,
			goldext.TrustedNodes,
		)
		return []byte(kanbanHTML)
	}

	// If this has links layout, render as links document
	if hasFrontmatter && metadata.Layout == "links" {
		linksHTML, err := frontmatter.RenderLinks(contentWithoutFrontmatter)
		if err != nil {
			// If links rendering fails, fall back to regular markdown
			md = contentWithoutFrontmatter
		} else {
			return []byte(linksHTML)
		}
	}

	// If there's frontmatter but not kanban layout, use content without frontmatter
	if hasFrontmatter {
		md = contentWithoutFrontmatter
	}

	// Apply any custom extensions via pre-processing
	md = goldext.ProcessMarkdown(md, docPath)

	// Configure Goldmark with all needed extensions
	markdown := goldmark.New(
		// Enable common extensions
		goldmark.WithExtensions(
			extension.Table,         // Enable tables
			extension.Strikethrough, // Enable ~~strikethrough~~
			extension.Linkify,       // Auto-link URLs
			// extension.TaskList,    // Disabled - we use our own task list processor
			extension.Footnote,        // Enable footnotes
			extension.DefinitionList,  // Enable definition lists
			extension.GFM,             // GitHub Flavored Markdown
			goldext.OnePasswordIgnore, // Add data-1p-ignore to code blocks
			goldext.TrustedNodes,      // Render typed Wiki-Go extension nodes
			// MathJax is now handled via client-side JavaScript
		),
		// Parser options
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(), // Enable auto heading IDs
			parser.WithAttribute(),     // Enable attributes
		),
		// Renderer options
		goldmark.WithRendererOptions(
			html.WithUnsafe(), // Allow raw HTML in the markdown
			html.WithHardWraps(),
		),
	)

	// Create a buffer to store the rendered HTML
	var buf bytes.Buffer

	// Convert markdown to HTML
	if err := markdown.Convert([]byte(md), &buf, parser.WithContext(goldext.NewRenderContext(docPath))); err != nil {
		// If there's an error, return an error message
		errMsg := []byte("<p>Error rendering markdown with Goldmark: " + err.Error() + "</p>")
		return errMsg
	}

	return buf.Bytes()
}

// blockLineRe matches the start of a fenced code block, ATX heading, or paragraph
// opening tag in the rendered HTML — used by injectSourceLines.
var blockTagRe = regexp.MustCompile(`(?i)<(h[1-6]|pre|blockquote|table|hr|ol|ul)([\s>])`)

// RenderMarkdownWithSourceLines renders markdown to HTML and injects
// data-source-line="N" attributes on block-level elements so the editor
// preview scroll-sync can map HTML elements back to source lines.
// It works by scanning the raw markdown for block boundaries and annotating
// the rendered HTML with the corresponding line numbers via a post-process pass.
func RenderMarkdownWithSourceLines(md string, docPath string) []byte {
	html := RenderMarkdownWithPath(md, docPath)
	return []byte(injectSourceLines(string(html), md))
}

// injectSourceLines walks the rendered HTML and the source markdown in parallel
// to stamp each block-level opening tag with its originating source line number.
func injectSourceLines(htmlStr string, md string) string {
	lines := strings.Split(md, "\n")

	// Special code fences that goldext renders as <div> elements (not <pre>).
	// These must NOT be recorded in blockLines because they produce no matching HTML tag.
	specialFences := map[string]bool{
		"mermaid": true, "youtube": true, "vimeo": true,
		"mp4": true, "details": true, "rtl": true, "ltr": true,
	}

	var blockLines []int
	inList := false
	inTable := false
	inCodeBlock := false
	var codeFence string
	prevBQDepth := 0  // previous blockquote depth
	inBQList := false // inside a list within a blockquote
	inBQCode := false // inside a code block within a blockquote

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Top-level code fence open/close
		if !inCodeBlock && (strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~")) {
			inCodeBlock = true
			codeFence = trimmed[:3]
			inList = false
			inTable = false
			// Extract fence language (e.g. "bash" from "```bash")
			lang := strings.ToLower(strings.TrimSpace(trimmed[3:]))
			// Only record if it will produce a <pre> in the final HTML;
			// special fences get preprocessed into <div> elements.
			if !specialFences[lang] {
				blockLines = append(blockLines, i)
			}
			continue
		}
		if inCodeBlock {
			if strings.HasPrefix(trimmed, codeFence) {
				inCodeBlock = false
				codeFence = ""
			}
			continue
		}

		if trimmed == "" {
			inList = false
			inTable = false
			prevBQDepth = 0
			inBQList = false
			inBQCode = false
			continue
		}

		// ATX headings
		if strings.HasPrefix(trimmed, "#") {
			inList = false
			inTable = false
			blockLines = append(blockLines, i)
			continue
		}

		// Blockquotes — use depth tracking so multi-line blockquotes emit ONE entry
		if strings.HasPrefix(trimmed, ">") {
			depth := bqDepth(trimmed)
			inner := strings.TrimSpace(trimmed[depth:])

			if depth > prevBQDepth {
				// Entering a new (possibly nested) blockquote block
				blockLines = append(blockLines, i)
				inBQList = false
				inBQCode = false
			}
			prevBQDepth = depth

			// Check inner content for sub-block elements that generate HTML tags
			if inBQCode {
				// Inside a code block within the blockquote
				if inner != "" && (strings.HasPrefix(inner, "```") || strings.HasPrefix(inner, "~~~")) {
					inBQCode = false // closing fence
				}
			} else if inner != "" {
				switch {
				case strings.HasPrefix(inner, "#"):
					blockLines = append(blockLines, i)
					inBQList = false
				case strings.HasPrefix(inner, "```") || strings.HasPrefix(inner, "~~~"):
					blockLines = append(blockLines, i)
					inBQCode = true
					inBQList = false
				case isListItemStart(inner):
					if !inBQList {
						blockLines = append(blockLines, i)
						inBQList = true
					}
				case strings.HasPrefix(inner, "|"):
					blockLines = append(blockLines, i)
					inBQList = false
				case inner == "---" || inner == "***" || inner == "___":
					blockLines = append(blockLines, i)
					inBQList = false
				default:
					// Regular paragraph content inside blockquote — no tag, reset list
					inBQList = false
				}
			}

			inList = false
			continue
		}

		// Leaving blockquote context
		prevBQDepth = 0
		inBQList = false
		inBQCode = false

		// Tables — record only the FIRST row of each table block
		if strings.HasPrefix(trimmed, "|") {
			if !inTable {
				blockLines = append(blockLines, i)
				inTable = true
			}
			inList = false
			continue
		}
		inTable = false

		// Thematic breaks
		if trimmed == "---" || trimmed == "***" || trimmed == "___" {
			inList = false
			inTable = false
			blockLines = append(blockLines, i)
			continue
		}

		// List items — record only the first item of each list block
		if isListItemStart(trimmed) {
			if !inList {
				blockLines = append(blockLines, i)
				inList = true
			}
			inTable = false
			continue
		}

		// Regular paragraph — not in regex, not recorded
		inList = false
		inTable = false
	}

	blockIdx := 0
	tableDepth := 0
	listDepth := 0

	// Also match </table>, </ol>, </ul> so we can track when we leave those contexts
	anyTagRe := regexp.MustCompile(`(?i)<(/?)` +
		`(h[1-6]|pre|blockquote|table|hr|ol|ul)` +
		`([\s>])`)

	return anyTagRe.ReplaceAllStringFunc(htmlStr, func(match string) string {
		sub := anyTagRe.FindStringSubmatchIndex(match)
		if sub == nil || len(sub) < 8 {
			return match
		}
		closing := match[sub[2]:sub[3]]  // "/" or ""
		tagName := match[sub[4]:sub[5]]  // tag name
		trailing := match[sub[6]:sub[7]] // space or >

		lowerTag := strings.ToLower(tagName)

		// Track table depth
		if lowerTag == "table" {
			if closing == "/" {
				if tableDepth > 0 {
					tableDepth--
				}
			} else {
				tableDepth++
			}
		}

		// Track list depth (ol and ul share the same counter)
		if lowerTag == "ol" || lowerTag == "ul" {
			if closing == "/" {
				if listDepth > 0 {
					listDepth--
				}
			} else {
				listDepth++
			}
		}

		// Closing tags are never annotated
		if closing == "/" {
			return match
		}

		// Tags inside a table (other than the table itself) are raw HTML in cells — skip
		if tableDepth > 0 && lowerTag != "table" {
			return match
		}

		// Only the outermost list tag is annotated; nested <ul>/<ol> are skipped
		// (listDepth is already incremented above, so >1 means we are nested)
		if (lowerTag == "ol" || lowerTag == "ul") && listDepth > 1 {
			return match
		}

		// Assign next blockLines entry
		if blockIdx >= len(blockLines) {
			return match
		}
		lineNum := blockLines[blockIdx]
		blockIdx++

		attr := ` data-source-line="` + strconv.Itoa(lineNum) + `"`
		if trailing == ">" {
			return "<" + tagName + attr + ">"
		}
		return "<" + tagName + attr + trailing
	})
}

// bqDepth counts the number of leading '>' characters in a trimmed line.
func bqDepth(s string) int {
	count := 0
	for _, ch := range s {
		if ch == '>' {
			count++
		} else {
			break
		}
	}
	return count
}

// isListItemStart returns true if s looks like the start of a list item.
func isListItemStart(s string) bool {
	if len(s) < 2 {
		return false
	}
	return ((s[0] == '-' || s[0] == '*' || s[0] == '+') && (s[1] == ' ' || s[1] == '\t')) ||
		(len(s) > 2 && s[0] >= '0' && s[0] <= '9' && s[1] == '.')
}
