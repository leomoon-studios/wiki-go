package markdown

import (
	"io"
	"regexp"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// ProcessCollapsibleCodeBlock processes a code block that contains a collapsible section
// Returns true if the block was processed as a collapsible section, false otherwise
func ProcessCollapsibleCodeBlock(w io.Writer, codeBlock *ast.CodeBlock) bool {
	// Convert Info to string for more reliable processing
	infoString := string(codeBlock.Info)

	// More robust check for details code blocks
	if strings.HasPrefix(strings.TrimSpace(infoString), "details") {
		// Extract the title from the info string
		title := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(infoString), "details"))

		// If no title is provided, use a default
		if title == "" {
			title = "Details"
		}

		// Get the content and ensure it has proper line endings for list parsing
		content := string(codeBlock.Literal)

		// Ensure content has proper line breaks
		// This is critical for proper list parsing
		content = strings.ReplaceAll(content, "\r\n", "\n")

		// Ensure there's a newline after each list item to help the parser
		// This regex finds list items not followed by blank lines and adds them
		listItemRegex := regexp.MustCompile(`(?m)^(\s*[-*+]|\s*[0-9]+\.)\s+.+$`)
		content = listItemRegex.ReplaceAllStringFunc(content, func(s string) string {
			// Check if the line is followed by another list item or blank line
			if strings.HasSuffix(s, "\n") {
				return s
			}
			return s + "\n"
		})

		// Create a new parser with the same extensions as the main parser
		extensions := parser.CommonExtensions | parser.AutoHeadingIDs |
			parser.Strikethrough | parser.Footnotes | parser.SuperSubscript | parser.MathJax

		p := parser.NewWithExtensions(extensions)

		// Parse the content
		doc := p.Parse([]byte(content))

		// Create HTML renderer with the same options but WITHOUT SmartyPants
		htmlFlags := html.CommonFlags | html.HrefTargetBlank
		// Remove SmartyPants flag if it's set
		htmlFlags &= ^html.Smartypants

		opts := html.RendererOptions{
			Flags:          htmlFlags,
			RenderNodeHook: RenderNodeHook, // Use our custom hook for nested elements
		}
		renderer := html.NewRenderer(opts)

		// Render the content as HTML
		renderedContent := markdown.Render(doc, renderer)

		// Add a unique ID for this details element
		detailsID := "details-" + strings.ReplaceAll(strings.ToLower(title), " ", "-")

		// Write the details/summary HTML with a print-class and data-print-open attribute
		w.Write([]byte("<details id=\"" + detailsID + "\" class=\"markdown-details\">\n"))
		w.Write([]byte("<summary>" + title + "</summary>\n"))
		w.Write([]byte("<div class=\"details-content\">\n"))
		w.Write(renderedContent)
		w.Write([]byte("</div>\n"))
		w.Write([]byte("</details>\n"))

		return true
	}

	return false
}
