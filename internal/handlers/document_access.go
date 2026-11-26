package handlers

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"wiki-go/internal/auth"
	"wiki-go/internal/config"
	"wiki-go/internal/utils"
)

// logicalDocumentPath converts stored document and attachment paths to the URL
// path form consumed by the access-rule engine.
func logicalDocumentPath(storedPath string) (string, error) {
	canonical, err := utils.CanonicalDocumentRequestPath(storedPath)
	if err != nil {
		return "", err
	}
	relative := strings.TrimPrefix(canonical, "/")
	if equalAttachmentComponent(relative, "pages/home") {
		return "/", nil
	}
	if components := strings.SplitN(relative, "/", 2); len(components) == 2 && equalAttachmentComponent(components[0], "documents") {
		relative = components[1]
	}
	return utils.CanonicalRequestPath("/" + relative)
}

type resolvedAPIDocument struct {
	utils.ResolvedDocumentPath
	allowed bool
}

// API logical paths and stored paths have distinct wire contracts. In
// particular /api/source/documents/x is the actual document /documents/x,
// whereas versions and comments also accept the legacy documents/ prefix.
func resolveAPIDocument(cfg *config.Config, raw string, session *auth.Session, creation, stored bool) (resolvedAPIDocument, error) {
	logical := raw
	var err error
	if stored {
		logical, err = logicalDocumentPath(raw)
	}
	if err != nil {
		return resolvedAPIDocument{}, err
	}
	resolver := auth.ResolveDocumentAccess
	if creation {
		resolver = auth.ResolveDocumentCreationAccess
	}
	document, allowed, err := resolver(logical, session, cfg)
	if err != nil {
		return resolvedAPIDocument{}, err
	}
	if creation && document.LogicalPath != "/" {
		// Check each parent that a recursive create would traverse, excluding
		// the storage root, which is distinct from the logical homepage.
		for parent := path.Dir(document.LogicalPath); parent != "/"; parent = path.Dir(parent) {
			if !auth.CanCreateDocument(parent, session, cfg) {
				allowed = false
			}
		}
	}
	return resolvedAPIDocument{document, allowed}, nil
}

func requireAPIDocument(w http.ResponseWriter, cfg *config.Config, raw string, session *auth.Session, creation, stored bool) (resolvedAPIDocument, bool) {
	document, err := resolveAPIDocument(cfg, raw, session, creation, stored)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, os.ErrNotExist) {
			status = http.StatusNotFound
		}
		sendJSONError(w, "Invalid document path", status, "")
		return resolvedAPIDocument{}, false
	}
	if !document.allowed {
		sendJSONError(w, "Document access denied", http.StatusForbidden, "")
		return resolvedAPIDocument{}, false
	}
	return document, true
}

func versionStorageName(logical string) string {
	if logical == "/" {
		return "pages/home"
	}
	return "documents/" + strings.TrimPrefix(logical, "/")
}

func metadataPath(cfg *config.Config, namespace, logical string) (string, error) {
	relative := strings.TrimPrefix(logical, "/")
	if namespace == "versions" {
		relative = versionStorageName(logical)
	}
	return utils.ResolveStoragePath(cfg.Wiki.RootDir, path.Join("/", namespace, relative))
}

// Resolve the fixed document file within its already resolved directory. A
// missing directory is permitted for creation; file links cannot switch owners.
func documentContentPath(document utils.ResolvedDocumentPath) (string, error) {
	parent := document.FilesystemPath
	suffix := "document.md"
	for {
		if _, err := os.Lstat(parent); err == nil {
			break
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		suffix = path.Join(filepath.Base(parent), suffix)
		parent = filepath.Dir(parent)
	}
	resolved, err := utils.ResolveDocumentDestination(parent, ".", suffix)
	if err != nil {
		return "", err
	}
	if !equalAttachmentComponent(resolved.FilesystemPath, filepath.Join(document.FilesystemPath, "document.md")) {
		return "", fmt.Errorf("document file changes its owning identity")
	}
	return resolved.FilesystemPath, nil
}

func authorizeDocumentTree(cfg *config.Config, session *auth.Session, document utils.ResolvedDocumentPath, targetLogical string) error {
	return filepath.WalkDir(document.FilesystemPath, func(filename string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(document.FilesystemPath, filename)
		if err != nil {
			return err
		}
		logical := path.Join(document.LogicalPath, filepath.ToSlash(relative))
		if !auth.CanAccessDocument(logical, session, cfg) || targetLogical != "" && !auth.CanCreateDocument(path.Join(targetLogical, filepath.ToSlash(relative)), session, cfg) {
			return fmt.Errorf("document subtree access denied")
		}
		return nil
	})
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
