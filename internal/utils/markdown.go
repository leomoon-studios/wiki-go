package utils

import (
	"bytes"
	"os"
	"path/filepath"
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
	// Check for frontmatter
	metadata, contentWithoutFrontmatter, hasFrontmatter := frontmatter.Parse(md)

	// If this has kanban layout, render as kanban with full goldext support
	if hasFrontmatter && metadata.Layout == "kanban" {
		// Create preprocessor functions (excluding frontmatter since it's already processed)
		var preprocessors []frontmatter.PreprocessorFunc
		var postProcessors []frontmatter.PostProcessorFunc

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

		// Add post-processors for mermaid and direction blocks
		postProcessors = append(postProcessors, func(html string) string {
			result := goldext.RestoreMermaidBlocks(html)
			result = goldext.RestoreDirectionBlocks(result)
			return result
		})

		kanbanHTML := frontmatter.RenderKanbanWithProcessors(contentWithoutFrontmatter, preprocessors, postProcessors)
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
			extension.Footnote,       // Enable footnotes
			extension.DefinitionList, // Enable definition lists
			extension.GFM,            // GitHub Flavored Markdown
			goldext.OnePasswordIgnore, // Add data-1p-ignore to code blocks
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
	if err := markdown.Convert([]byte(md), &buf); err != nil {
		// If there's an error, return an error message
		errMsg := []byte("<p>Error rendering markdown with Goldmark: " + err.Error() + "</p>")
		return errMsg
	}

	// Post-process: Restore Mermaid blocks that were replaced with placeholders
	htmlResult := goldext.RestoreMermaidBlocks(buf.String())

	// Post-process: Restore Direction blocks that were replaced with placeholders
	// This ensures RTL/LTR content is properly rendered with Markdown formatting
	htmlResult = goldext.RestoreDirectionBlocks(htmlResult)

	// Return the post-processed HTML
	return []byte(htmlResult)
}