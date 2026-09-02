package goldext

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

// AlertType is a fixed GitHub-style alert type.
type AlertType uint8

const (
	AlertInvalid AlertType = iota
	AlertNote
	AlertTip
	AlertImportant
	AlertWarning
	AlertCaution
)

// AlertBlock contains Markdown copied from a recognized alert blockquote.
type AlertBlock struct {
	ast.BaseBlock
	Alert   AlertType
	Content []byte
}

// KindAlertBlock is the Goldmark kind for AlertBlock.
var KindAlertBlock = ast.NewNodeKind("WikiGoAlertBlock")

// Kind implements ast.Node.
func (n *AlertBlock) Kind() ast.NodeKind {
	return KindAlertBlock
}

// Dump implements ast.Node.
func (n *AlertBlock) Dump(source []byte, level int) {
	_, title, _ := alertPresentation(n.Alert)
	ast.DumpHelper(n, source, level, map[string]string{"Type": title}, nil)
}

func parseAlertType(marker string) AlertType {
	switch strings.ToUpper(strings.TrimSpace(marker)) {
	case "[!NOTE]":
		return AlertNote
	case "[!TIP]":
		return AlertTip
	case "[!IMPORTANT]":
		return AlertImportant
	case "[!WARNING]":
		return AlertWarning
	case "[!CAUTION]":
		return AlertCaution
	default:
		return AlertInvalid
	}
}

func newAlertBlock(alertType AlertType, content []byte) *AlertBlock {
	return &AlertBlock{Alert: alertType, Content: append([]byte(nil), content...)}
}

func (r *trustedNodeRenderer) renderAlertBlock(writer util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	alert := node.(*AlertBlock)
	classSuffix, title, iconClass := alertPresentation(alert.Alert)
	if classSuffix == "" {
		return ast.WalkSkipChildren, nil
	}

	var rendered bytes.Buffer
	if err := newSafeNestedMarkdown().Convert(alert.Content, &rendered); err != nil {
		return ast.WalkSkipChildren, err
	}

	_, _ = writer.WriteString(`<div class="markdown-alert markdown-alert-` + classSuffix + `">`)
	_, _ = writer.WriteString(`<p class="markdown-alert-title"><i class="fa ` + iconClass + `" aria-hidden="true"></i> `)
	_, _ = writer.WriteString(title)
	_, _ = writer.WriteString(`</p><div class="markdown-alert-content">`)
	_, _ = writer.Write(rendered.Bytes())
	_, _ = writer.WriteString("</div></div>\n")
	return ast.WalkSkipChildren, nil
}

func alertPresentation(alertType AlertType) (classSuffix, title, iconClass string) {
	switch alertType {
	case AlertNote:
		return "note", "Note", "fa-info-circle"
	case AlertTip:
		return "tip", "Tip", "fa-lightbulb-o"
	case AlertImportant:
		return "important", "Important", "fa-exclamation-circle"
	case AlertWarning:
		return "warning", "Warning", "fa-exclamation-triangle"
	case AlertCaution:
		return "caution", "Caution", "fa-ban"
	default:
		return "", "", ""
	}
}
