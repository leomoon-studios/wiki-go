package goldext

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

// VideoProvider is a fixed external video provider accepted by VideoEmbedBlock.
type VideoProvider uint8

const (
	VideoProviderInvalid VideoProvider = iota
	VideoProviderYouTube
	VideoProviderVimeo
)

// VideoEmbedBlock represents a validated external video ID. Embed and public
// URLs are constructed by the renderer and never accepted from Markdown.
type VideoEmbedBlock struct {
	ast.BaseBlock
	Provider VideoProvider
	VideoID  string
}

// KindVideoEmbedBlock is the Goldmark kind for VideoEmbedBlock.
var KindVideoEmbedBlock = ast.NewNodeKind("WikiGoVideoEmbedBlock")

// Kind implements ast.Node.
func (n *VideoEmbedBlock) Kind() ast.NodeKind {
	return KindVideoEmbedBlock
}

// Dump implements ast.Node.
func (n *VideoEmbedBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"VideoID": n.VideoID}, nil)
}

func newVideoEmbedBlock(provider VideoProvider, videoID string) *VideoEmbedBlock {
	return &VideoEmbedBlock{Provider: provider, VideoID: videoID}
}

func (r *trustedNodeRenderer) renderVideoEmbedBlock(writer util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	video := node.(*VideoEmbedBlock)
	var label, embedURL, publicURL, allow, title string
	switch video.Provider {
	case VideoProviderYouTube:
		if !youtubeIDPattern.MatchString(video.VideoID) {
			return ast.WalkSkipChildren, nil
		}
		label = "YouTube Video"
		title = "YouTube video player"
		embedURL = "https://www.youtube.com/embed/" + video.VideoID
		publicURL = "https://www.youtube.com/watch?v=" + video.VideoID
		allow = "accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; fullscreen"
	case VideoProviderVimeo:
		if !vimeoIDPattern.MatchString(video.VideoID) {
			return ast.WalkSkipChildren, nil
		}
		label = "Vimeo Video"
		title = "Vimeo video player"
		embedURL = "https://player.vimeo.com/video/" + video.VideoID
		publicURL = "https://vimeo.com/" + video.VideoID
		allow = "autoplay; fullscreen; picture-in-picture"
	default:
		return ast.WalkSkipChildren, nil
	}

	_, _ = writer.WriteString("<div class=\"video-container\">\n<iframe width=\"560\" height=\"315\"")
	writeAttribute(writer, "src", embedURL)
	writeAttribute(writer, "title", title)
	_, _ = writer.WriteString(" frameborder=\"0\"")
	writeAttribute(writer, "allow", allow)
	_, _ = writer.WriteString(" referrerpolicy=\"strict-origin-when-cross-origin\" allowfullscreen></iframe>\n</div>\n")
	_, _ = writer.WriteString("<div class=\"video-print-placeholder\">\n<p><strong>" + label + "</strong></p>\n")
	_, _ = writer.WriteString("<p>This embedded video is not available in print. You can view it online at:</p>\n<p><a")
	writeAttribute(writer, "href", publicURL)
	_, _ = writer.WriteString(">")
	_, _ = writer.WriteString(publicURL)
	_, _ = writer.WriteString("</a></p>\n</div>\n")
	return ast.WalkSkipChildren, nil
}
