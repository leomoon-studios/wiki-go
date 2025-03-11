package markdown

import (
	"io"

	"github.com/gomarkdown/markdown/ast"
)

// VideoProvider represents a video platform provider
type VideoProvider string

const (
	YouTube VideoProvider = "youtube"
	Vimeo   VideoProvider = "vimeo"
	Local   VideoProvider = "mp4"
)

// RenderVideo renders a responsive video embed for the given provider and ID
func RenderVideo(w io.Writer, provider VideoProvider, videoID string) {
	switch provider {
	case YouTube:
		RenderYouTubeVideo(w, videoID)
	case Vimeo:
		RenderVimeoVideo(w, videoID)
	case Local:
		RenderMP4Video(w, videoID)
	}
}

// ProcessVideoCodeBlock processes a code block that might contain a video embed
// Returns true if the block was processed as a video, false otherwise
func ProcessVideoCodeBlock(w io.Writer, codeBlock *ast.CodeBlock) bool {
	content := string(codeBlock.Literal)
	info := codeBlock.Info

	// Check for each video type
	if ProcessYouTubeCodeBlock(w, info, content) {
		return true
	}

	if ProcessVimeoCodeBlock(w, info, content) {
		return true
	}

	if ProcessMP4CodeBlock(w, info, content) {
		return true
	}

	return false
}
