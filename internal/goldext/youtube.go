package goldext

import (
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

// YouTubePreprocessor transforms youtube code blocks into HTML embeds
// and avoids processing nested youtube blocks inside other code blocks
func YouTubePreprocessor(markdown string, _ string) string {
	// Use same approach as MermaidPreprocessor for consistent code block handling
	lines := strings.Split(markdown, "\n")
	processedLines := make([]string, len(lines))
	copy(processedLines, lines)

	// Maps to store replacements
	replacements := make(map[int]string) // Line index -> replacement HTML
	linesToRemove := make(map[int]bool)  // Lines to be removed

	// State tracking
	backtickStack := 0           // Track nesting of ``` blocks
	tildeStack := 0              // Track nesting of ~~~ blocks
	inYouTubeBlock := false      // Are we in a youtube block?
	youtubeStart := -1           // Start line of current youtube block
	youtubeContent := []string{} // Content of current youtube block
	youtubeBlockType := ""       // Type of block: "backtick" or "tilde"

	// Scan through all lines
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for code block markers
		if strings.HasPrefix(trimmed, "```") {
			if backtickStack == 0 {
				// Opening a backtick block
				backtickStack++

				// Check if it's a youtube block and we're not inside any other block
				if tildeStack == 0 && strings.Contains(trimmed, "youtube") {
					inYouTubeBlock = true
					youtubeStart = i
					youtubeContent = []string{}
					youtubeBlockType = "backtick"
					// Mark this line for removal
					linesToRemove[i] = true
				}
			} else {
				// Closing a backtick block
				backtickStack--

				// If we're closing a youtube block
				if inYouTubeBlock && youtubeBlockType == "backtick" && backtickStack == 0 && tildeStack == 0 {
					// Get video ID from content
					videoID := ExtractYouTubeID(strings.Join(youtubeContent, "\n"))

					if videoID != "" {
						// Create replacement HTML
						videoURL := "https://www.youtube.com/watch?v=" + videoID
						replacement := `<div class="video-container">
<iframe width="560" height="315" src="https://www.youtube.com/embed/` + videoID + `"
frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; fullscreen"></iframe>
</div>
<div class="video-print-placeholder">
<p><strong>YouTube Video</strong></p>
<p>This embedded video is not available in print. You can view it online at:</p>
<p><a href="` + videoURL + `">` + videoURL + `</a></p>
</div>`

						replacements[youtubeStart] = replacement
					}

					// Mark this line for removal
					linesToRemove[i] = true

					// Reset state
					inYouTubeBlock = false
					youtubeStart = -1
					youtubeContent = []string{}
					youtubeBlockType = ""
				}
			}
		} else if strings.HasPrefix(trimmed, "~~~") {
			if tildeStack == 0 {
				// Opening a tilde block
				tildeStack++

				// Check if it's a youtube block and we're not inside any other block
				if backtickStack == 0 && strings.Contains(trimmed, "youtube") {
					inYouTubeBlock = true
					youtubeStart = i
					youtubeContent = []string{}
					youtubeBlockType = "tilde"
					// Mark this line for removal
					linesToRemove[i] = true
				}
			} else {
				// Closing a tilde block
				tildeStack--

				// If we're closing a youtube block
				if inYouTubeBlock && youtubeBlockType == "tilde" && tildeStack == 0 && backtickStack == 0 {
					// Get video ID from content
					videoID := ExtractYouTubeID(strings.Join(youtubeContent, "\n"))

					if videoID != "" {
						// Create replacement HTML
						videoURL := "https://www.youtube.com/watch?v=" + videoID
						replacement := `<div class="video-container">
<iframe width="560" height="315" src="https://www.youtube.com/embed/` + videoID + `"
frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; fullscreen"></iframe>
</div>
<div class="video-print-placeholder">
<p><strong>YouTube Video</strong></p>
<p>This embedded video is not available in print. You can view it online at:</p>
<p><a href="` + videoURL + `">` + videoURL + `</a></p>
</div>`

						replacements[youtubeStart] = replacement
					}

					// Mark this line for removal
					linesToRemove[i] = true

					// Reset state
					inYouTubeBlock = false
					youtubeStart = -1
					youtubeContent = []string{}
					youtubeBlockType = ""
				}
			}
		} else if inYouTubeBlock && ((youtubeBlockType == "backtick" && backtickStack > 0 && tildeStack == 0) ||
			(youtubeBlockType == "tilde" && tildeStack > 0 && backtickStack == 0)) {
			// We're inside a youtube block, collect the content
			youtubeContent = append(youtubeContent, line)
			// Mark this line for removal
			linesToRemove[i] = true
		}
	}

	// Handle unclosed blocks at end of document
	if inYouTubeBlock && youtubeStart >= 0 {
		// Get video ID from content
		videoID := ExtractYouTubeID(strings.Join(youtubeContent, "\n"))

		if videoID != "" {
			// Create replacement HTML
			videoURL := "https://www.youtube.com/watch?v=" + videoID
			replacement := `<div class="video-container">
<iframe width="560" height="315" src="https://www.youtube.com/embed/` + videoID + `"
frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; fullscreen"></iframe>
</div>
<div class="video-print-placeholder">
<p><strong>YouTube Video</strong></p>
<p>This embedded video is not available in print. You can view it online at:</p>
<p><a href="` + videoURL + `">` + videoURL + `</a></p>
</div>`

			replacements[youtubeStart] = replacement
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
