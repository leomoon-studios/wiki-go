package handlers

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"wiki-go/internal/auth"
	"wiki-go/internal/config"
	"wiki-go/internal/utils"
)

type resolvedAttachmentPath struct {
	filesystemPath string
	rootPath       string
	storagePath    string
	owner          utils.ResolvedDocumentPath
}

// resolveAttachmentPath resolves an existing owner and its attachment's stored
// identity. Only the final filename may be missing, so callers can authorize
// before returning 404 or recognizing an already completed rename. Lookup
// failures and dangling links never fall back to an unchecked request spelling.
func resolveAttachmentPath(cfg *config.Config, attachmentPath string) (resolvedAttachmentPath, error) {
	return resolveAttachmentStoragePath(cfg, attachmentPath, false)
}

// resolveAttachmentDirectory resolves an existing attachment directory below
// either the documents root or the separate homepage root.
func resolveAttachmentDirectory(cfg *config.Config, documentPath string) (resolvedAttachmentPath, error) {
	return resolveAttachmentStoragePath(cfg, documentPath, true)
}

func resolveAttachmentStoragePath(cfg *config.Config, input string, directory bool) (resolvedAttachmentPath, error) {
	if cfg == nil || input == "" {
		return resolvedAttachmentPath{}, fmt.Errorf("attachment path is empty or unconfigured")
	}
	normalized := strings.ReplaceAll(input, "\\", "/")
	if isAbsoluteAttachmentPath(normalized) {
		return resolvedAttachmentPath{}, fmt.Errorf("attachment path is absolute")
	}
	canonical, err := utils.CanonicalDocumentRequestPath(normalized)
	if err != nil {
		return resolvedAttachmentPath{}, err
	}
	storagePath := strings.TrimPrefix(canonical, "/")
	if storagePath == "" {
		return resolvedAttachmentPath{}, fmt.Errorf("attachment path is empty")
	}

	rootDir := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir)
	relativePath := storagePath
	homepage := false
	components := strings.Split(storagePath, "/")
	if equalAttachmentComponent(components[0], "pages") {
		if len(components) < 2 || !equalAttachmentComponent(components[1], "home") {
			return resolvedAttachmentPath{}, fmt.Errorf("attachment path is outside the homepage root")
		}
		homepage = true
		rootDir = filepath.Join(cfg.Wiki.RootDir, "pages", "home")
		relativePath = strings.Join(components[2:], "/")
	}
	root, err := utils.ResolveExistingDocumentPath(rootDir, ".", "/")
	if err != nil {
		return resolvedAttachmentPath{}, fmt.Errorf("resolve attachment root: %w", err)
	}
	logicalDirectory := "/" + relativePath
	if !directory {
		if relativePath == "" {
			return resolvedAttachmentPath{}, fmt.Errorf("attachment path does not name a file")
		}
		logicalDirectory = path.Dir(logicalDirectory)
	}
	parent, err := utils.ResolveExistingDocumentPath(root.FilesystemPath, ".", logicalDirectory)
	if err != nil {
		return resolvedAttachmentPath{}, fmt.Errorf("resolve attachment directory: %w", err)
	}
	info, err := os.Stat(parent.FilesystemPath)
	if err != nil {
		return resolvedAttachmentPath{}, fmt.Errorf("stat attachment directory: %w", err)
	}
	if !info.IsDir() || !homepage && parent.LogicalPath == "/" {
		return resolvedAttachmentPath{}, fmt.Errorf("attachment has no owning document directory")
	}
	owner := parent
	if homepage {
		// Every subdirectory in homepage storage belongs to logical /, rather
		// than to a similarly named directory in the documents root.
		owner = root
	}
	filesystemPath := parent.FilesystemPath
	if !directory {
		file, err := utils.ResolveDocumentDestination(parent.FilesystemPath, ".", "/"+path.Base(relativePath))
		if err != nil {
			return resolvedAttachmentPath{}, fmt.Errorf("resolve attachment file: %w", err)
		}
		if filepath.Dir(file.FilesystemPath) != parent.FilesystemPath {
			return resolvedAttachmentPath{}, fmt.Errorf("attachment file alias changes its owning directory")
		}
		filesystemPath = file.FilesystemPath
	}
	relative, err := filepath.Rel(root.FilesystemPath, filesystemPath)
	if err != nil {
		return resolvedAttachmentPath{}, err
	}
	storagePath = filepath.ToSlash(relative)
	if homepage {
		storagePath = path.Join("pages/home", storagePath)
	}
	return resolvedAttachmentPath{filesystemPath, root.FilesystemPath, storagePath, owner}, nil
}

func canAccessAttachment(session *auth.Session, cfg *config.Config, resolved resolvedAttachmentPath) bool {
	owner, allowed, err := auth.ResolveDocumentAccess(resolved.owner.LogicalPath, session, cfg)
	return err == nil && allowed && owner.FilesystemPath == resolved.owner.FilesystemPath
}

func equalAttachmentComponent(left, right string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(left, right)
	}
	return left == right
}

func validateAttachmentFilename(filename string) error {
	if filename == "." || filename == ".." || !utils.IsValidFilename(filename) {
		return fmt.Errorf("invalid attachment filename")
	}
	_, err := utils.CanonicalDocumentRequestPath(filename)
	return err
}

func resolveAttachmentRenameDestination(source resolvedAttachmentPath, filename string) (resolvedAttachmentPath, error) {
	if err := validateAttachmentFilename(filename); err != nil {
		return resolvedAttachmentPath{}, err
	}
	if source.filesystemPath == "" || source.storagePath == "" {
		return resolvedAttachmentPath{}, fmt.Errorf("source attachment path is incomplete")
	}
	directory := source
	directory.filesystemPath = filepath.Dir(source.filesystemPath)
	directory.storagePath = path.Dir(source.storagePath)
	return resolveAttachmentFileDestination(directory, filename)
}

// resolveAttachmentFileDestination is shared by uploads and renames. The parent
// is already resolved and authorized; existing filename aliases must stay in
// that directory, and missing names are validated before any file is created.
func resolveAttachmentFileDestination(source resolvedAttachmentPath, filename string) (resolvedAttachmentPath, error) {
	canonical, err := utils.CanonicalDocumentRequestPath(filename)
	if err != nil || filename == "" || strings.ContainsAny(filename, "/\\") || canonical != "/"+filename || filename == "." || filename == ".." {
		return resolvedAttachmentPath{}, fmt.Errorf("invalid attachment filename")
	}
	if source.filesystemPath == "" || source.rootPath == "" || source.storagePath == "" || source.owner.FilesystemPath == "" {
		return resolvedAttachmentPath{}, fmt.Errorf("source attachment path is incomplete")
	}
	parent := source.filesystemPath
	file, err := utils.ResolveDocumentDestination(parent, ".", filename)
	if err != nil {
		return resolvedAttachmentPath{}, err
	}
	if filepath.Dir(file.FilesystemPath) != parent || strings.EqualFold(filepath.Base(file.FilesystemPath), "document.md") {
		return resolvedAttachmentPath{}, fmt.Errorf("attachment destination changes its owner or names a document")
	}
	info, err := os.Stat(file.FilesystemPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return resolvedAttachmentPath{}, fmt.Errorf("stat attachment destination: %w", err)
	}
	if err == nil && !info.Mode().IsRegular() {
		return resolvedAttachmentPath{}, fmt.Errorf("attachment destination is not a regular file")
	}
	return resolvedAttachmentPath{
		filesystemPath: file.FilesystemPath,
		rootPath:       source.rootPath,
		storagePath:    path.Join(source.storagePath, filepath.Base(file.FilesystemPath)),
		owner:          source.owner,
	}, nil
}

func isAbsoluteAttachmentPath(attachmentPath string) bool {
	if strings.HasPrefix(attachmentPath, "/") || filepath.IsAbs(filepath.FromSlash(attachmentPath)) {
		return true
	}
	return len(attachmentPath) >= 2 &&
		((attachmentPath[0] >= 'A' && attachmentPath[0] <= 'Z') || (attachmentPath[0] >= 'a' && attachmentPath[0] <= 'z')) &&
		attachmentPath[1] == ':'
}
