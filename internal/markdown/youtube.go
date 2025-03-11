package markdown

import (
	"io"
	"regexp"
	"strings"
)

// ExtractYouTubeID extracts the YouTube video ID from a YouTube URL
// If the input is already just an ID, it returns it as is
func ExtractYouTubeID(input string) string {
	// If the input is already just an ID (no slashes or dots), return it
	if !strings.Contains(input, "/") && !strings.Contains(input, ".") && len(input) >= 11 {
		return strings.TrimSpace(input)
	}

	// Otherwise, try to extract ID from URL
	patterns := []string{
		`(?:youtube\.com\/watch\?v=|youtu\.be\/)([^&\?]+)`,
		`youtube\.com\/embed\/([^&\?]+)`,
		`youtube\.com\/v\/([^&\?]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(input)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

// RenderYouTubeVideo renders a YouTube video embed
func RenderYouTubeVideo(w io.Writer, videoID string) {
	videoURL := "https://www.youtube.com/watch?v=" + videoID

	w.Write([]byte(`<div class="video-container">
<iframe width="560" height="315" src="https://www.youtube.com/embed/` + videoID + `"
frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
allowfullscreen></iframe>
</div>`))

	// Add print-friendly placeholder with link
	w.Write([]byte(`<div class="video-print-placeholder">
<p><strong>YouTube Video</strong></p>
<p>This embedded video is not available in print. You can view it online at:</p>
<p><a href="` + videoURL + `">` + videoURL + `</a></p>
</div>`))
}

// ProcessYouTubeCodeBlock processes a code block that contains a YouTube video ID or URL
// Returns true if the block was processed as a YouTube video, false otherwise
func ProcessYouTubeCodeBlock(w io.Writer, codeBlockInfo []byte, content string) bool {
	if string(codeBlockInfo) == "youtube" {
		videoID := ExtractYouTubeID(content)
		if videoID != "" {
			RenderYouTubeVideo(w, videoID)
			return true
		}
	}
	return false
}
