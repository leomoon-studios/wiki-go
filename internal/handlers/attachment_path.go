package handlers

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"wiki-go/internal/config"
	"wiki-go/internal/utils"
)

type resolvedAttachmentPath struct {
	filesystemPath string
	rootPath       string
	storagePath    string
}

// resolveAttachmentPath resolves an attachment path beneath either the
// documents root or the homepage root. The input is the storage path used by
// attachment APIs, such as "finance/report.pdf" or "pages/home/logo.png".
func resolveAttachmentPath(cfg *config.Config, attachmentPath string) (resolvedAttachmentPath, error) {
	resolved, relativePath, err := resolveAttachmentStoragePath(cfg, attachmentPath)
	if err != nil {
		return resolvedAttachmentPath{}, err
	}
	if relativePath == "" {
		return resolvedAttachmentPath{}, fmt.Errorf("attachment path does not name a file")
	}
	return resolved, nil
}

// resolveAttachmentDirectory applies the attachment storage boundary while
// allowing the homepage root itself as an upload destination.
func resolveAttachmentDirectory(cfg *config.Config, documentPath string) (resolvedAttachmentPath, error) {
	resolved, _, err := resolveAttachmentStoragePath(cfg, documentPath)
	return resolved, err
}

func resolveAttachmentStoragePath(cfg *config.Config, attachmentPath string) (resolvedAttachmentPath, string, error) {
	if attachmentPath == "" {
		return resolvedAttachmentPath{}, "", fmt.Errorf("attachment path is empty")
	}

	normalized := strings.ReplaceAll(attachmentPath, "\\", "/")
	if isAbsoluteAttachmentPath(normalized) {
		return resolvedAttachmentPath{}, "", fmt.Errorf("attachment path is absolute")
	}
	for _, segment := range strings.Split(normalized, "/") {
		if segment == ".." {
			return resolvedAttachmentPath{}, "", fmt.Errorf("attachment path contains parent traversal")
		}
	}

	canonical, err := utils.CanonicalRequestPath(normalized)
	if err != nil {
		return resolvedAttachmentPath{}, "", fmt.Errorf("canonicalize attachment path: %w", err)
	}
	storagePath := strings.TrimPrefix(canonical, "/")
	if storagePath == "" {
		return resolvedAttachmentPath{}, "", fmt.Errorf("attachment path is empty")
	}

	var allowedRoot string
	var relativePath string
	switch {
	case storagePath == "pages/home":
		allowedRoot = filepath.Join(cfg.Wiki.RootDir, "pages", "home")
	case strings.HasPrefix(storagePath, "pages/home/"):
		allowedRoot = filepath.Join(cfg.Wiki.RootDir, "pages", "home")
		relativePath = strings.TrimPrefix(storagePath, "pages/home/")
	case strings.HasPrefix(storagePath, "pages/"):
		return resolvedAttachmentPath{}, "", fmt.Errorf("attachment path is outside the homepage root")
	default:
		allowedRoot = filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir)
		relativePath = storagePath
	}

	rootPath, err := filepath.Abs(allowedRoot)
	if err != nil {
		return resolvedAttachmentPath{}, "", fmt.Errorf("resolve attachment root: %w", err)
	}
	filesystemPath := filepath.Join(rootPath, filepath.FromSlash(relativePath))
	containedPath, err := filepath.Rel(rootPath, filesystemPath)
	if err != nil {
		return resolvedAttachmentPath{}, "", fmt.Errorf("resolve attachment path: %w", err)
	}
	if containedPath == ".." || strings.HasPrefix(containedPath, ".."+string(filepath.Separator)) || filepath.IsAbs(containedPath) {
		return resolvedAttachmentPath{}, "", fmt.Errorf("attachment path escapes its allowed root")
	}

	return resolvedAttachmentPath{
		filesystemPath: filesystemPath,
		rootPath:       rootPath,
		storagePath:    storagePath,
	}, relativePath, nil
}

func validateAttachmentFilename(filename string) error {
	if filename == "." || filename == ".." || !utils.IsValidFilename(filename) {
		return fmt.Errorf("invalid attachment filename")
	}
	return nil
}

func resolveAttachmentRenameDestination(source resolvedAttachmentPath, filename string) (resolvedAttachmentPath, error) {
	if err := validateAttachmentFilename(filename); err != nil {
		return resolvedAttachmentPath{}, err
	}
	if source.filesystemPath == "" || source.rootPath == "" || source.storagePath == "" {
		return resolvedAttachmentPath{}, fmt.Errorf("source attachment path is incomplete")
	}

	filesystemPath := filepath.Join(filepath.Dir(source.filesystemPath), filename)
	containedPath, err := filepath.Rel(source.rootPath, filesystemPath)
	if err != nil {
		return resolvedAttachmentPath{}, fmt.Errorf("resolve attachment destination: %w", err)
	}
	if containedPath == ".." || strings.HasPrefix(containedPath, ".."+string(filepath.Separator)) || filepath.IsAbs(containedPath) {
		return resolvedAttachmentPath{}, fmt.Errorf("attachment destination escapes its allowed root")
	}

	return resolvedAttachmentPath{
		filesystemPath: filesystemPath,
		rootPath:       source.rootPath,
		storagePath:    path.Join(path.Dir(source.storagePath), filename),
	}, nil
}

func isAbsoluteAttachmentPath(attachmentPath string) bool {
	if strings.HasPrefix(attachmentPath, "/") || filepath.IsAbs(filepath.FromSlash(attachmentPath)) {
		return true
	}
	return len(attachmentPath) >= 3 &&
		((attachmentPath[0] >= 'A' && attachmentPath[0] <= 'Z') || (attachmentPath[0] >= 'a' && attachmentPath[0] <= 'z')) &&
		attachmentPath[1] == ':' && attachmentPath[2] == '/'
}
