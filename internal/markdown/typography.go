package markdown

import (
	"fmt"
	"regexp"
	"strings"
)

// ApplyTypographicReplacements handles common typographic replacements and misspelling corrections
func ApplyTypographicReplacements(input string) string {
	// First, identify and protect code blocks and inline code
	// We'll replace them with placeholders, apply typographic changes, then restore the code
	codeBlockRegex := regexp.MustCompile("```[\\s\\S]*?```|`[^`]+`")
	var codeBlocks []string

	// Extract and replace code blocks with placeholders
	input = codeBlockRegex.ReplaceAllStringFunc(input, func(match string) string {
		codeBlocks = append(codeBlocks, match)
		return fmt.Sprintf("_CODE_BLOCK_%d_", len(codeBlocks)-1)
	})

	// Create a map of replacements for typographic shortcodes
	typoShortcodes := map[string]string{
		"(c)":  "©", // Copyright symbol
		"(r)":  "®", // Registered trademark symbol
		"(tm)": "™", // Trademark symbol
		"(p)":  "¶", // Paragraph symbol
		"+-":   "±", // Plus-minus symbol
		"...":  "…", // Ellipsis
		"1/2":  "½", // One-half
		"1/4":  "¼", // One-quarter
		"3/4":  "¾", // Three-quarters
	}

	// Handle typographic shortcodes first using a more flexible approach
	// This pattern matches each shortcode even when they're consecutive
	for pattern, replacement := range typoShortcodes {
		escapedPattern := regexp.QuoteMeta(pattern)
		// Match the pattern when it appears at beginning, end, after space, or even consecutive
		re := regexp.MustCompile(escapedPattern)
		input = re.ReplaceAllString(input, replacement)
	}

	// Restore code blocks
	for i, block := range codeBlocks {
		placeholder := fmt.Sprintf("_CODE_BLOCK_%d_", i)
		input = strings.Replace(input, placeholder, block, 1)
	}

	return input
}
