package goldext

import (
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// missingLinkTransformer restores Wiki-Go's dead-link indicator without
// generating raw HTML. Missing internal document and attachment links are
// wrapped in a typed span rendered by the trusted-node renderer.
type missingLinkTransformer struct{}

func (t *missingLinkTransformer) Transform(document *ast.Document, _ text.Reader, _ parser.Context) {
	missing := make([]*ast.Link, 0)
	_ = ast.Walk(document, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		link, ok := node.(*ast.Link)
		if ok && localLinkTargetMissing(string(link.Destination)) {
			missing = append(missing, link)
		}
		return ast.WalkContinue, nil
	})

	for _, link := range missing {
		parent := link.Parent()
		if parent == nil {
			continue
		}
		wrapper := NewTrustedInline(TrustedInlineSpan)
		wrapper.Class = "notfound"
		parent.ReplaceChild(parent, link, wrapper)
		wrapper.AppendChild(wrapper, link)
	}
}

func localLinkTargetMissing(destination string) bool {
	parsed, err := url.Parse(destination)
	if err != nil || parsed.Scheme != "" || parsed.Host != "" || parsed.Path == "" {
		return false
	}

	if strings.HasPrefix(parsed.Path, "/api/files/") {
		return localAttachmentTargetMissing(parsed)
	}
	if !strings.HasPrefix(parsed.Path, "/") || parsed.Path == "/" || strings.HasPrefix(parsed.Path, "/api/") {
		return false
	}

	target, ok := safeLinkFilesystemPath(DocumentsRoot(), parsed.EscapedPath())
	if !ok {
		return false
	}
	_, err = os.Stat(filepath.Join(target, "document.md"))
	return os.IsNotExist(err)
}

func localAttachmentTargetMissing(parsed *url.URL) bool {
	escapedPath := strings.TrimPrefix(parsed.EscapedPath(), "/api/files/")
	base := DocumentsRoot()
	if escapedPath == "pages/home" || strings.HasPrefix(escapedPath, "pages/home/") {
		base = homepageRoot()
		escapedPath = strings.TrimPrefix(strings.TrimPrefix(escapedPath, "pages/home"), "/")
	}
	target, ok := safeLinkFilesystemPath(base, escapedPath)
	if !ok {
		return false
	}
	_, err := os.Stat(target)
	return os.IsNotExist(err)
}

func safeLinkFilesystemPath(base, escapedURLPath string) (string, bool) {
	decoded, err := url.PathUnescape(escapedURLPath)
	if err != nil || decoded == "" || strings.Contains(decoded, "\\") {
		return "", false
	}
	decoded = strings.TrimPrefix(decoded, "/")
	cleaned := path.Clean("/" + decoded)
	if cleaned == "/" || strings.TrimPrefix(cleaned, "/") != decoded {
		return "", false
	}

	target := filepath.Join(base, filepath.FromSlash(decoded))
	relative, err := filepath.Rel(base, target)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", false
	}
	return target, true
}
