package goldext

import (
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

// VimeoPreprocessor transforms vimeo code blocks into HTML embeds
// and avoids processing nested vimeo blocks inside other code blocks
func VimeoPreprocessor(markdown string, _ string) string {
	// Use same approach as MermaidPreprocessor for consistent code block handling
	lines := strings.Split(markdown, "\n")
	processedLines := make([]string, len(lines))
	copy(processedLines, lines)

	// Maps to store replacements
	replacements := make(map[int]string) // Line index -> replacement HTML
	linesToRemove := make(map[int]bool)  // Lines to be removed

	// State tracking
	backtickStack := 0         // Track nesting of ``` blocks
	tildeStack := 0            // Track nesting of ~~~ blocks
	inVimeoBlock := false      // Are we in a vimeo block?
	vimeoStart := -1           // Start line of current vimeo block
	vimeoContent := []string{} // Content of current vimeo block
	vimeoBlockType := ""       // Type of block: "backtick" or "tilde"

	// Scan through all lines
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for code block markers
		if strings.HasPrefix(trimmed, "```") {
			if backtickStack == 0 {
				// Opening a backtick block
				backtickStack++

				// Check if it's a vimeo block and we're not inside any other block
				if tildeStack == 0 && strings.Contains(trimmed, "vimeo") {
					inVimeoBlock = true
					vimeoStart = i
					vimeoContent = []string{}
					vimeoBlockType = "backtick"
					// Mark this line for removal
					linesToRemove[i] = true
				}
			} else {
				// Closing a backtick block
				backtickStack--

				// If we're closing a vimeo block
				if inVimeoBlock && vimeoBlockType == "backtick" && backtickStack == 0 && tildeStack == 0 {
					// Get video ID from content
					videoID := ExtractVimeoID(strings.Join(vimeoContent, "\n"))

					if videoID != "" {
						// Create replacement HTML
						videoURL := "https://vimeo.com/" + videoID
						replacement := `<div class="video-container">
<iframe src="https://player.vimeo.com/video/` + videoID + `"
width="560" height="315" frameborder="0"
allow="autoplay; fullscreen; picture-in-picture"></iframe>
</div>
<div class="video-print-placeholder">
<p><strong>Vimeo Video</strong></p>
<p>This embedded video is not available in print. You can view it online at:</p>
<p><a href="` + videoURL + `">` + videoURL + `</a></p>
</div>`

						replacements[vimeoStart] = replacement
					}

					// Mark this line for removal
					linesToRemove[i] = true

					// Reset state
					inVimeoBlock = false
					vimeoStart = -1
					vimeoContent = []string{}
					vimeoBlockType = ""
				}
			}
		} else if strings.HasPrefix(trimmed, "~~~") {
			if tildeStack == 0 {
				// Opening a tilde block
				tildeStack++

				// Check if it's a vimeo block and we're not inside any other block
				if backtickStack == 0 && strings.Contains(trimmed, "vimeo") {
					inVimeoBlock = true
					vimeoStart = i
					vimeoContent = []string{}
					vimeoBlockType = "tilde"
					// Mark this line for removal
					linesToRemove[i] = true
				}
			} else {
				// Closing a tilde block
				tildeStack--

				// If we're closing a vimeo block
				if inVimeoBlock && vimeoBlockType == "tilde" && tildeStack == 0 && backtickStack == 0 {
					// Get video ID from content
					videoID := ExtractVimeoID(strings.Join(vimeoContent, "\n"))

					if videoID != "" {
						// Create replacement HTML
						videoURL := "https://vimeo.com/" + videoID
						replacement := `<div class="video-container">
<iframe src="https://player.vimeo.com/video/` + videoID + `"
width="560" height="315" frameborder="0"
allow="autoplay; fullscreen; picture-in-picture"></iframe>
</div>
<div class="video-print-placeholder">
<p><strong>Vimeo Video</strong></p>
<p>This embedded video is not available in print. You can view it online at:</p>
<p><a href="` + videoURL + `">` + videoURL + `</a></p>
</div>`

						replacements[vimeoStart] = replacement
					}

					// Mark this line for removal
					linesToRemove[i] = true

					// Reset state
					inVimeoBlock = false
					vimeoStart = -1
					vimeoContent = []string{}
					vimeoBlockType = ""
				}
			}
		} else if inVimeoBlock && ((vimeoBlockType == "backtick" && backtickStack > 0 && tildeStack == 0) ||
			(vimeoBlockType == "tilde" && tildeStack > 0 && backtickStack == 0)) {
			// We're inside a vimeo block, collect the content
			vimeoContent = append(vimeoContent, line)
			// Mark this line for removal
			linesToRemove[i] = true
		}
	}

	// Handle unclosed blocks at end of document
	if inVimeoBlock && vimeoStart >= 0 {
		// Get video ID from content
		videoID := ExtractVimeoID(strings.Join(vimeoContent, "\n"))

		if videoID != "" {
			// Create replacement HTML
			videoURL := "https://vimeo.com/" + videoID
			replacement := `<div class="video-container">
<iframe src="https://player.vimeo.com/video/` + videoID + `"
width="560" height="315" frameborder="0"
allow="autoplay; fullscreen; picture-in-picture"></iframe>
</div>
<div class="video-print-placeholder">
<p><strong>Vimeo Video</strong></p>
<p>This embedded video is not available in print. You can view it online at:</p>
<p><a href="` + videoURL + `">` + videoURL + `</a></p>
</div>`

			replacements[vimeoStart] = replacement
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
