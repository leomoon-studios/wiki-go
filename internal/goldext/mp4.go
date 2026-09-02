package goldext

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

// TransformMP4Path transforms a local video file path to a proper API URL
// It prepends "/api/files/" to the document path and filename
func TransformMP4Path(videoPath string, docPath string) string {
	// Skip transformation if it already looks like a URL
	if strings.HasPrefix(videoPath, "http://") ||
		strings.HasPrefix(videoPath, "https://") ||
		strings.HasPrefix(videoPath, "/") {
		return videoPath
	}

	// URL encode the filename to handle spaces and special characters
	escapedPath := url.PathEscape(videoPath)

	// Handle the homepage special case
	if docPath == "" || docPath == "/" {
		// Homepage files are stored in "pages/home"
		return "/api/files/pages/home/" + escapedPath
	}

	// Regular document files
	return "/api/files/" + docPath + "/" + escapedPath
}

// MP4Preprocessor transforms MP4 code blocks into HTML video elements
// and avoids processing nested MP4 blocks inside other code blocks
func MP4Preprocessor(markdown string, docPath string) string {
	// Scan fenced blocks consistently with the other media preprocessors.
	lines := strings.Split(markdown, "\n")
	processedLines := make([]string, len(lines))
	copy(processedLines, lines)

	// Maps to store replacements
	replacements := make(map[int]string) // Line index -> replacement HTML
	linesToRemove := make(map[int]bool)  // Lines to be removed

	// State tracking
	backtickStack := 0       // Track nesting of ``` blocks
	tildeStack := 0          // Track nesting of ~~~ blocks
	inMP4Block := false      // Are we in an MP4 block?
	mp4Start := -1           // Start line of current MP4 block
	mp4Content := []string{} // Content of current MP4 block
	mp4BlockType := ""       // Type of block: "backtick" or "tilde"

	// Scan through all lines
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for code block markers
		if strings.HasPrefix(trimmed, "```") {
			if backtickStack == 0 {
				// Opening a backtick block
				backtickStack++

				// Check if it's an MP4 block and we're not inside any other block
				if tildeStack == 0 && strings.Contains(trimmed, "mp4") {
					inMP4Block = true
					mp4Start = i
					mp4Content = []string{}
					mp4BlockType = "backtick"
					// Mark this line for removal
					linesToRemove[i] = true
				}
			} else {
				// Closing a backtick block
				backtickStack--

				// If we're closing an MP4 block
				if inMP4Block && mp4BlockType == "backtick" && backtickStack == 0 && tildeStack == 0 {
					// Get video path from content
					videoPath := strings.TrimSpace(strings.Join(mp4Content, "\n"))

					if videoPath != "" {
						// Transform the path to proper API URL
						videoPath = TransformMP4Path(videoPath, docPath)

						// Get filename for display in print placeholder
						filename := filepath.Base(videoPath)

						// Create replacement HTML
						replacement := fmt.Sprintf(`<div class="video-container">
<video class="local-video-player" style="max-width: 100%%; height: auto;" controls>
<source src="%s" type="video/mp4">
Your browser does not support the video tag.
</video>
</div>
<div class="video-print-placeholder">
<p><strong>Video Content</strong></p>
<p>This embedded video (%s) is not available in print.</p>
<p>To view this video, access this document at your wiki URL.</p>
</div>`, videoPath, filename)

						replacements[mp4Start] = replacement
					}

					// Mark this line for removal
					linesToRemove[i] = true

					// Reset state
					inMP4Block = false
					mp4Start = -1
					mp4Content = []string{}
					mp4BlockType = ""
				}
			}
		} else if strings.HasPrefix(trimmed, "~~~") {
			if tildeStack == 0 {
				// Opening a tilde block
				tildeStack++

				// Check if it's an MP4 block and we're not inside any other block
				if backtickStack == 0 && strings.Contains(trimmed, "mp4") {
					inMP4Block = true
					mp4Start = i
					mp4Content = []string{}
					mp4BlockType = "tilde"
					// Mark this line for removal
					linesToRemove[i] = true
				}
			} else {
				// Closing a tilde block
				tildeStack--

				// If we're closing an MP4 block
				if inMP4Block && mp4BlockType == "tilde" && tildeStack == 0 && backtickStack == 0 {
					// Get video path from content
					videoPath := strings.TrimSpace(strings.Join(mp4Content, "\n"))

					if videoPath != "" {
						// Transform the path to proper API URL
						videoPath = TransformMP4Path(videoPath, docPath)

						// Get filename for display in print placeholder
						filename := filepath.Base(videoPath)

						// Create replacement HTML
						replacement := fmt.Sprintf(`<div class="video-container">
<video class="local-video-player" style="max-width: 100%%; height: auto;" controls>
<source src="%s" type="video/mp4">
Your browser does not support the video tag.
</video>
</div>
<div class="video-print-placeholder">
<p><strong>Video Content</strong></p>
<p>This embedded video (%s) is not available in print.</p>
<p>To view this video, access this document at your wiki URL.</p>
</div>`, videoPath, filename)

						replacements[mp4Start] = replacement
					}

					// Mark this line for removal
					linesToRemove[i] = true

					// Reset state
					inMP4Block = false
					mp4Start = -1
					mp4Content = []string{}
					mp4BlockType = ""
				}
			}
		} else if inMP4Block && ((mp4BlockType == "backtick" && backtickStack > 0 && tildeStack == 0) ||
			(mp4BlockType == "tilde" && tildeStack > 0 && backtickStack == 0)) {
			// We're inside an MP4 block, collect the content
			mp4Content = append(mp4Content, line)
			// Mark this line for removal
			linesToRemove[i] = true
		}
	}

	// Handle unclosed blocks at end of document
	if inMP4Block && mp4Start >= 0 {
		// Get video path from content
		videoPath := strings.TrimSpace(strings.Join(mp4Content, "\n"))

		if videoPath != "" {
			// Transform the path to proper API URL
			videoPath = TransformMP4Path(videoPath, docPath)

			// Get filename for display in print placeholder
			filename := filepath.Base(videoPath)

			// Create replacement HTML
			replacement := fmt.Sprintf(`<div class="video-container">
<video class="local-video-player" style="max-width: 100%%; height: auto;" controls>
<source src="%s" type="video/mp4">
Your browser does not support the video tag.
</video>
</div>
<div class="video-print-placeholder">
<p><strong>Video Content</strong></p>
<p>This embedded video (%s) is not available in print.</p>
<p>To view this video, access this document at your wiki URL.</p>
</div>`, videoPath, filename)

			replacements[mp4Start] = replacement
		}
	}

	// Process the lines, applying replacements and removing marked lines
	result := []string{}

	for i, line := range processedLines {
		if replacement, ok := replacements[i]; ok {
			// This line has a replacement
			result = append(result, replacement)
		} else if !linesToRemove[i] {
			// This line should not be removed
			result = append(result, line)
		}
		// Lines marked for removal are skipped
	}

	return strings.Join(result, "\n")
}
