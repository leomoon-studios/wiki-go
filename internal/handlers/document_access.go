package handlers

import (
	"path"
	"strings"

	"wiki-go/internal/auth"
	"wiki-go/internal/config"
	"wiki-go/internal/utils"
)

// logicalDocumentPath converts stored document and attachment paths to the URL
// path form consumed by the access-rule engine.
func logicalDocumentPath(storedPath string) (string, error) {
	canonical, err := utils.CanonicalRequestPath(storedPath)
	if err != nil {
		return "", err
	}
	relative := strings.TrimPrefix(canonical, "/")
	if relative == "pages/home" {
		return "/", nil
	}
	relative = strings.TrimPrefix(relative, "documents/")
	return utils.CanonicalRequestPath("/" + relative)
}

func attachmentDocumentPath(attachmentPath string) (string, error) {
	canonical, err := utils.CanonicalDocumentRequestPath(attachmentPath)
	if err != nil {
		return "", err
	}
	components := strings.Split(strings.TrimPrefix(canonical, "/"), "/")
	if len(components) >= 3 && equalAttachmentComponent(components[0], "pages") && equalAttachmentComponent(components[1], "home") {
		return "/", nil
	}
	return logicalDocumentPath(path.Dir(canonical))
}

func canAccessLogicalDocument(rSession *auth.Session, cfg *config.Config, logicalPath string) bool {
	return auth.CanAccessDocument(logicalPath, rSession, cfg)
}
