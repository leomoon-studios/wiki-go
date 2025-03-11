package markdown

import (
	"io"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// ProcessDirectionCodeBlock processes a code block that should have forced text direction
// Returns true if the block was processed as a direction block, false otherwise
func ProcessDirectionCodeBlock(w io.Writer, codeBlock *ast.CodeBlock) bool {
	// Convert Info to string for more reliable processing
	infoString := strings.TrimSpace(string(codeBlock.Info))

	// Check if this is an LTR or RTL block
	isLTR := infoString == "ltr"
	isRTL := infoString == "rtl"

	if !isLTR && !isRTL {
		return false
	}

	// Get the direction attribute value based on the info string
	direction := "ltr"
	if isRTL {
		direction = "rtl"
	}

	// Get the content
	content := string(codeBlock.Literal)

	// Simple content cleanup
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.TrimSpace(content)

	// Use the same parser configuration as the main application
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs |
		parser.NoEmptyLineBeforeBlock | parser.NoIntraEmphasis |
		parser.Tables | parser.FencedCode | parser.Autolink |
		parser.Strikethrough | parser.SpaceHeadings

	// Create the parser
	p := parser.NewWithExtensions(extensions)

	// Parse markdown to AST
	doc := p.Parse([]byte(content))

	// Configure renderer with tight lists (no paragraphs in list items)
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	htmlFlags &= ^html.Smartypants

	opts := html.RendererOptions{
		Flags: htmlFlags,
		RenderNodeHook: func(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
			// Prevent paragraphs inside list items
			if para, ok := node.(*ast.Paragraph); ok {
				parent := para.Parent
				if _, isListItem := parent.(*ast.ListItem); isListItem {
					if entering {
						// For entering a paragraph in a list item, don't output the <p> tag
						return ast.GoToNext, true
					} else {
						// For exiting a paragraph in a list item, don't output the </p> tag
						return ast.GoToNext, true
					}
				}
			}
			return ast.GoToNext, false
		},
	}

	renderer := html.NewRenderer(opts)

	// Render to HTML
	renderedContent := markdown.Render(doc, renderer)

	// Write the div with direction attributes
	w.Write([]byte("<div class=\"" + direction + "\">\n"))
	w.Write(renderedContent)
	w.Write([]byte("</div>\n"))

	return true
}
