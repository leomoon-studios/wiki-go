package goldext

import (
	"encoding/json"
	"io/fs"
	"log"
	"strings"

	"wiki-go/internal/resources"
)

// EmojiData represents an emoji entry in the JSON file
type EmojiData struct {
	Emoji      string   `json:"emoji"`
	Shortcodes []string `json:"shortcodes"`
}

// Global emoji map
var emojis map[string]string

// init loads emoji data from the JSON file
func init() {
	// Initialize the map
	emojis = make(map[string]string)

	// Get the data filesystem
	dataFS := resources.GetDataFS()

	// Read the emoji JSON file
	emojiFile, err := fs.ReadFile(dataFS, "emojis.json")
	if err != nil {
		log.Printf("Error reading emoji file: %v", err)
		return
	}

	// Parse the JSON data
	var emojiList []EmojiData
	if err := json.Unmarshal(emojiFile, &emojiList); err != nil {
		log.Printf("Error parsing emoji data: %v", err)
		return
	}

	// Convert to map for faster lookups
	for _, emoji := range emojiList {
		// Add all shortcodes in the array
		for _, code := range emoji.Shortcodes {
			// Ensure shortcode has colon prefix/suffix if not already present
			formattedCode := code
			if !strings.HasPrefix(formattedCode, ":") {
				formattedCode = ":" + formattedCode
			}
			if !strings.HasSuffix(formattedCode, ":") {
				formattedCode = formattedCode + ":"
			}
			emojis[formattedCode] = emoji.Emoji
		}
	}

	log.Printf("Loaded %d emojis from emojis.json", len(emojis))
}

// EmojiPreprocessor replaces emoji shortcodes with Unicode emoji characters
// but avoids processing text inside code blocks
func EmojiPreprocessor(markdown string, _ string) string {
	// Process line by line instead of relying on regex which might fail on large documents
	lines := strings.Split(markdown, "\n")
	var result []string

	inCodeBlock := false

	for _, line := range lines {
		// Check if this line starts or ends a code block
		trimmedLine := strings.TrimSpace(line)

		// Strip blockquote prefix(es) to detect code blocks inside blockquotes
		// e.g., "> ```" or "> > ```" should be detected as code block markers
		contentLine := trimmedLine
		for strings.HasPrefix(contentLine, ">") {
			contentLine = strings.TrimSpace(strings.TrimPrefix(contentLine, ">"))
		}

		if strings.HasPrefix(contentLine, "```") || strings.HasPrefix(contentLine, "~~~") {
			inCodeBlock = !inCodeBlock
			result = append(result, line)
			continue
		}

		// If we're in a code block, don't process
		if inCodeBlock {
			result = append(result, line)
			continue
		}

		// Process each segment of the line, preserving inline code
		var processedLine string
		segments := strings.Split(line, "`")

		for i, segment := range segments {
			// Even segments (0, 2, 4...) are outside inline code
			if i%2 == 0 {
				// Apply all replacements to non-code segments
				for shortcode, emoji := range emojis {
					segment = strings.ReplaceAll(segment, shortcode, emoji)
				}
				processedLine += segment
			} else {
				// Odd segments (1, 3, 5...) are inside inline code - preserve them
				processedLine += "`" + segment + "`"
			}
		}

		result = append(result, processedLine)
	}

	return strings.Join(result, "\n")
}