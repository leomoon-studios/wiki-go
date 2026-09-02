package goldext

import (
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// trustedFenceTransformer converts only exact, recognized fenced block info
// strings. Extra values remain ordinary escaped code blocks.
type trustedFenceTransformer struct {
	documentPath string
}

func (t *trustedFenceTransformer) Transform(document *ast.Document, reader text.Reader, context parser.Context) {
	source := reader.Source()
	documentPath := RenderStateFromContext(context).DocumentPath
	if documentPath == "" {
		documentPath = t.documentPath
	}
	var fences []*ast.FencedCodeBlock
	_ = ast.Walk(document, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if fence, ok := node.(*ast.FencedCodeBlock); ok && fence.Info != nil {
				fences = append(fences, fence)
			}
		}
		return ast.WalkContinue, nil
	})

	for _, fence := range fences {
		var replacement ast.Node
		content := strings.TrimSpace(string(fence.Lines().Value(source)))
		switch strings.TrimSpace(string(fence.Info.Text(source))) {
		case "ltr":
			replacement = newDirectionBlock(DirectionLTR)
		case "rtl":
			replacement = newDirectionBlock(DirectionRTL)
		case "mermaid":
			replacement = newMermaidBlock()
		case "mp4":
			sourceURL, filename, ok := ResolveLocalMP4URL(content, documentPath)
			if !ok {
				continue
			}
			replacement = newLocalVideoBlock(sourceURL, filename)
		case "youtube":
			videoID := ExtractYouTubeID(content)
			if videoID == "" {
				continue
			}
			replacement = newVideoEmbedBlock(VideoProviderYouTube, videoID)
		case "vimeo":
			videoID := ExtractVimeoID(content)
			if videoID == "" {
				continue
			}
			replacement = newVideoEmbedBlock(VideoProviderVimeo, videoID)
		default:
			continue
		}

		replacement.SetLines(fence.Lines())
		parent := fence.Parent()
		parent.ReplaceChild(parent, fence, replacement)
	}
}
