package goldext

// This file controls the loading order of all preprocessors
// The order is important as some preprocessors may interfere with others if not run in the correct sequence

// These variables ensure the preprocessors are available for registration
// We don't actually use them directly, but they're needed for the compiler to include the preprocessors
var (
	_ = LinkPreprocessor
	_ = MermaidPreprocessor
	_ = DirectionPreprocessor
	_ = MP4Preprocessor
	_ = YouTubePreprocessor
	_ = VimeoPreprocessor
	_ = StatsPreprocessor
	_ = HighlightPreprocessor
	_ = TypographyPreprocessor
	_ = EmojiPreprocessor
	_ = DetailsPreprocessor
	_ = InfoBoxPreprocessor
	// _ = TaskListPreprocessor
	_ = TocPreprocessor
	_ = HeadingAnchorPreprocessor
	_ = SuperscriptPreprocessor
	_ = SubscriptPreprocessor
	_ = ScriptSanitizePreprocessor
	_ = FrontmatterPreprocessor
)

func init() {
	// Clear any previously registered preprocessors to ensure consistent ordering
	RegisteredPreprocessors = nil

	// Step 0: Process frontmatter FIRST, before any other processors
	RegisterPreprocessor(FrontmatterPreprocessor) // Process frontmatter

	// Step 1: Process Mermaid FIRST, before any other processors can touch the content
	RegisterPreprocessor(MermaidPreprocessor) // Process mermaid diagrams first

	// Step 3: Register preprocessors that handle code blocks
	RegisterPreprocessor(LinkPreprocessor)      // Process links and images
	RegisterPreprocessor(DirectionPreprocessor) // Process RTL/LTR blocks
	RegisterPreprocessor(MP4Preprocessor)       // Process MP4 video blocks
	RegisterPreprocessor(YouTubePreprocessor)   // Process YouTube video blocks
	RegisterPreprocessor(VimeoPreprocessor)     // Process Vimeo video blocks
	RegisterPreprocessor(StatsPreprocessor)     // Process stats shortcodes
	RegisterPreprocessor(DetailsPreprocessor)   // Process details blocks
	RegisterPreprocessor(InfoBoxPreprocessor)   // Process GitHub-flavored alerts
	// RegisterPreprocessor(TaskListPreprocessor)  // Process task lists before rendering
	RegisterPreprocessor(TocPreprocessor)       // Process table of contents markers
	RegisterPreprocessor(HeadingAnchorPreprocessor) // Add ¶ anchors to headings

	// Step 4: Register text formatting preprocessors
	RegisterPreprocessor(HighlightPreprocessor)  // Process highlighting
	RegisterPreprocessor(TypographyPreprocessor) // Process typography replacements
	RegisterPreprocessor(EmojiPreprocessor)      // Process emoji shortcodes

	// Step 5: Register these last to avoid interference with other syntax
	// These preprocessors will skip content inside MathJax blocks ($ and $$)
	RegisterPreprocessor(SuperscriptPreprocessor) // Process superscript (avoids MathJax content)
	RegisterPreprocessor(SubscriptPreprocessor)   // Process subscript (avoids MathJax content)

	// Step 6: Security-related preprocessing (run last to sanitize all generated content and handle unwrapped code blocks)
	RegisterPreprocessor(ScriptSanitizePreprocessor) // Sanitize script tags
}
