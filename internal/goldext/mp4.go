package goldext

import (
	"net/url"
	"path"
	"strings"
	"unicode"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

// LocalVideoBlock represents an MP4 attachment resolved relative to the
// document being rendered. SourceURL must remain under /api/files/.
type LocalVideoBlock struct {
	ast.BaseBlock
	SourceURL string
	Filename  string
}

// KindLocalVideoBlock is the Goldmark kind for LocalVideoBlock.
var KindLocalVideoBlock = ast.NewNodeKind("WikiGoLocalVideoBlock")

// Kind implements ast.Node.
func (n *LocalVideoBlock) Kind() ast.NodeKind {
	return KindLocalVideoBlock
}

// Dump implements ast.Node.
func (n *LocalVideoBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"SourceURL": n.SourceURL}, nil)
}

func newLocalVideoBlock(sourceURL, filename string) *LocalVideoBlock {
	return &LocalVideoBlock{SourceURL: sourceURL, Filename: filename}
}

// TransformMP4Path resolves an attachment filename to its local API URL. An
// empty result means the input was not a valid local MP4 attachment filename.
func TransformMP4Path(videoPath string, docPath string) string {
	resolved, _, ok := ResolveLocalMP4URL(videoPath, docPath)
	if !ok {
		return ""
	}
	return resolved
}

// ResolveLocalMP4URL converts one local MP4 attachment filename into a URL for
// the current document. External URLs, absolute paths, traversal, and nested
// paths are deliberately rejected.
func ResolveLocalMP4URL(videoPath, docPath string) (sourceURL, filename string, ok bool) {
	filename = strings.TrimSpace(videoPath)
	if !validMP4Filename(filename) {
		return "", "", false
	}

	documentURLPath, ok := attachmentDocumentURLPath(docPath)
	if !ok {
		return "", "", false
	}

	sourceURL = "/api/files/" + documentURLPath + "/" + url.PathEscape(filename)
	if _, ok := validateResolvedLocalMP4URL(sourceURL); !ok {
		return "", "", false
	}
	return sourceURL, filename, true
}

func validMP4Filename(filename string) bool {
	if filename == "" || filename == "." || filename == ".." ||
		strings.ContainsAny(filename, `/\\?#:`) || !strings.EqualFold(path.Ext(filename), ".mp4") {
		return false
	}
	for _, character := range filename {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func attachmentDocumentURLPath(docPath string) (string, bool) {
	docPath = strings.Trim(docPath, "/")
	if docPath == "" {
		return "pages/home", true
	}
	if strings.Contains(docPath, "\\") {
		return "", false
	}

	segments := strings.Split(docPath, "/")
	for index, segment := range segments {
		if segment == "" || segment == "." || segment == ".." {
			return "", false
		}
		for _, character := range segment {
			if unicode.IsControl(character) {
				return "", false
			}
		}
		segments[index] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/"), true
}

func validateResolvedLocalMP4URL(value string) (string, bool) {
	validated, ok := ValidateTrustedURL(value)
	if !ok || !strings.HasPrefix(validated, "/api/files/") {
		return "", false
	}

	parsed, err := url.Parse(validated)
	if err != nil || parsed.Scheme != "" || parsed.Host != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", false
	}
	decodedPath, err := url.PathUnescape(parsed.EscapedPath())
	if err != nil || strings.Contains(decodedPath, "\\") || !strings.HasPrefix(decodedPath, "/api/files/") {
		return "", false
	}

	segments := strings.Split(strings.TrimPrefix(decodedPath, "/api/files/"), "/")
	if len(segments) < 2 {
		return "", false
	}
	for _, segment := range segments {
		if segment == "" || segment == "." || segment == ".." {
			return "", false
		}
	}
	if !strings.EqualFold(path.Ext(segments[len(segments)-1]), ".mp4") {
		return "", false
	}
	return validated, true
}

func (r *trustedNodeRenderer) renderLocalVideoBlock(writer util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	video := node.(*LocalVideoBlock)
	sourceURL, ok := validateResolvedLocalMP4URL(video.SourceURL)
	if !ok {
		return ast.WalkSkipChildren, nil
	}

	_, _ = writer.WriteString("<div class=\"video-container\">\n")
	_, _ = writer.WriteString("<video class=\"local-video-player\" style=\"max-width: 100%; height: auto;\" controls>\n")
	_, _ = writer.WriteString("<source")
	writeAttribute(writer, "src", sourceURL)
	_, _ = writer.WriteString(" type=\"video/mp4\">\n")
	_, _ = writer.WriteString("Your browser does not support the video tag.\n</video>\n</div>\n")
	_, _ = writer.WriteString("<div class=\"video-print-placeholder\">\n<p><strong>Video Content</strong></p>\n<p>This embedded video (")
	_, _ = writer.Write(util.EscapeHTML([]byte(video.Filename)))
	_, _ = writer.WriteString(") is not available in print.</p>\n<p>To view this video, access this document at your wiki URL.</p>\n</div>\n")
	return ast.WalkSkipChildren, nil
}
