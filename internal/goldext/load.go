package goldext

// This file controls the loading order of all preprocessors
// The order is important as some preprocessors may interfere with others if not run in the correct sequence

// These variables ensure the preprocessors are available for registration
// We don't actually use them directly, but they're needed for the compiler to include the preprocessors
var (
	_ = WikiLinkPreprocessor
	_ = LinkPreprocessor
	_ = ShortcodesPreprocessor
	_ = TypographyPreprocessor
	_ = EmojiPreprocessor
	// _ = TaskListPreprocessor
	_ = FrontmatterPreprocessor
)

func init() {
	// Clear any previously registered preprocessors to ensure consistent ordering
	RegisteredPreprocessors = nil

	// Step 0: Process frontmatter FIRST, before any other processors
	RegisterPreprocessor(FrontmatterPreprocessor) // Process frontmatter

	// Register Markdown/text transformations before Unicode substitutions.
	RegisterPreprocessor(WikiLinkPreprocessor)   // Convert [[wikilinks]] to safe local Markdown links
	RegisterPreprocessor(LinkPreprocessor)       // Process links and images
	RegisterPreprocessor(ShortcodesPreprocessor) // Replace text-only year shortcodes
	// RegisterPreprocessor(TaskListPreprocessor)  // Process task lists before rendering

	// Register text-only preprocessors.
	RegisterPreprocessor(TypographyPreprocessor) // Process typography replacements
	RegisterPreprocessor(EmojiPreprocessor)      // Process emoji shortcodes

}
