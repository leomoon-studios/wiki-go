package markdown

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// RenderMP4Video renders a local MP4 video player
func RenderMP4Video(w io.Writer, videoPath string) {
	w.Write([]byte(`<div class="video-container">
<video class="local-video-player" style="max-width: 100%; height: auto;" controls>
<source src="` + videoPath + `" type="video/mp4">
Your browser does not support the video tag.
</video>
</div>`))

	// Get just the filename for cleaner display
	filename := filepath.Base(videoPath)

	// Add print-friendly placeholder
	printHTML := fmt.Sprintf(`<div class="video-print-placeholder">
<p><strong>Video Content</strong></p>
<p>This embedded video (%s) is not available in print.</p>
<p>To view this video, access this document at your wiki URL.</p>
</div>`, filename)

	w.Write([]byte(printHTML))
}

// ProcessMP4CodeBlock processes a code block that contains a local MP4 video path
// Returns true if the block was processed as an MP4 video, false otherwise
func ProcessMP4CodeBlock(w io.Writer, codeBlockInfo []byte, content string) bool {
	if string(codeBlockInfo) == "mp4" {
		videoPath := strings.TrimSpace(content)
		if videoPath != "" {
			// Check if it's a simple filename or already a full path
			if IsLocalFileReference(videoPath) && CurrentDocumentPath != "" {
				// Transform to a full URL based on the current document path
				videoPath = TransformLocalFileReference(videoPath, CurrentDocumentPath)
			}
			// Render the video
			RenderMP4Video(w, videoPath)
			return true
		}
	}
	return false
}
