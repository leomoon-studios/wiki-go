// Package safehtml owns Wiki-Go's trusted HTML boundary.
//
// Go sink inventory:
//   - trusted.go is the only approved location for template.HTML conversions.
//   - frontmatter/kanban.go, utils/markdown.go, and utils/comment_markdown.go
//     supply configured safe-renderer output to this package.
//   - handlers/page.go and handlers/error.go supply contextually escaped
//     html/template output to this package.
//   - comments/comments.go, frontmatter/kanban.go, handlers/types.go, and
//     types/types.go only hold trusted values produced through this boundary.
//   - formatted markup in goldext/shortcodes.go validates URLs and escapes
//     inserted values; handlers/metadata_api.go formats input regex patterns,
//     not HTML output.
//
// Browser sink inventory:
//   - search, file, version, access-rule, backup, import, settings, document,
//     and comment UI modules render API/user data with WikiDOM or native DOM
//     text and attribute APIs.
//   - version-history.js and editor-preview.js accept HTML only from the safe
//     server Markdown renderer.
//   - mermaid-init.js accepts SVG from Mermaid in strict mode, and
//     sidebar-navigation.js accepts a sidebar from a same-origin response.
//   - Kanban browser rendering encodes raw HTML, emits only fixed formatting
//     tags, and validates link/image schemes with WikiDOM.markdownURL.
//   - remaining innerHTML assignments insert fixed UI icons/scaffolding, clear
//     children, or read existing task DOM back into Markdown.
//
// The tests in this package enforce the Go boundary and high-risk JavaScript
// sinks. tests/xss_dom_test.js exercises the browser boundary with element-,
// attribute-, URL-, filename-, and error-shaped payloads.
package safehtml
