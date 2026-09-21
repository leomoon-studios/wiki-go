package handlers

import (
	"path/filepath"
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
	canonical, err := utils.CanonicalRequestPath(attachmentPath)
	if err != nil {
		return "", err
	}
	return logicalDocumentPath(filepath.ToSlash(filepath.Dir(canonical)))
}

func canAccessLogicalDocument(rSession *auth.Session, cfg *config.Config, logicalPath string) bool {
	return auth.CanAccessDocument(logicalPath, rSession, cfg)
}
