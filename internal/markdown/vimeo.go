package markdown

import (
	"io"
	"regexp"
	"strings"
)

// ExtractVimeoID extracts the Vimeo video ID from a Vimeo URL
// If the input is already just an ID, it returns it as is
func ExtractVimeoID(input string) string {
	// If the input is already just an ID (numeric only), return it
	if matched, _ := regexp.MatchString(`^\d+$`, strings.TrimSpace(input)); matched {
		return strings.TrimSpace(input)
	}

	// Otherwise, try to extract ID from URL
	patterns := []string{
		`vimeo\.com\/(\d+)`,
		`vimeo\.com\/video\/(\d+)`,
		`player\.vimeo\.com\/video\/(\d+)`,
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

// RenderVimeoVideo renders a Vimeo video embed
func RenderVimeoVideo(w io.Writer, videoID string) {
	videoURL := "https://vimeo.com/" + videoID

	w.Write([]byte(`<div class="video-container">
<iframe src="https://player.vimeo.com/video/` + videoID + `"
width="560" height="315" frameborder="0"
allow="autoplay; fullscreen; picture-in-picture" allowfullscreen></iframe>
</div>`))

	// Add print-friendly placeholder with link
	w.Write([]byte(`<div class="video-print-placeholder">
<p><strong>Vimeo Video</strong></p>
<p>This embedded video is not available in print. You can view it online at:</p>
<p><a href="` + videoURL + `">` + videoURL + `</a></p>
</div>`))
}

// ProcessVimeoCodeBlock processes a code block that contains a Vimeo video ID or URL
// Returns true if the block was processed as a Vimeo video, false otherwise
func ProcessVimeoCodeBlock(w io.Writer, codeBlockInfo []byte, content string) bool {
	if string(codeBlockInfo) == "vimeo" {
		videoID := ExtractVimeoID(content)
		if videoID != "" {
			RenderVimeoVideo(w, videoID)
			return true
		}
	}
	return false
}
