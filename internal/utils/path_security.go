package utils

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ResolvedDocumentPath contains the canonical URL path used for authorization
// and the corresponding filesystem path contained by the documents root.
type ResolvedDocumentPath struct {
	LogicalPath    string
	FilesystemPath string
}

// ResolveExistingDocumentPath resolves an existing resource to its filesystem
// identity and returns the corresponding logical authorization path. The
// selected documents root must exist; "/" identifies that root, leaving homepage
// root selection to the caller. Windows aliases resolve to their stored long
// names. Other platforms preserve case. Links must remain within the root.
// Callers must use both returned paths rather than reconstructing the original
// spelling. Resolution does not protect a later open against concurrent renames.
func ResolveExistingDocumentPath(rootDir, documentsDir, logicalPath string) (ResolvedDocumentPath, error) {
	return resolveDocumentIdentity(rootDir, documentsDir, logicalPath, false)
}

// ResolveDocumentDestination resolves the existing prefix of a creation
// destination before appending validated missing components. An existing
// destination is resolved just like ResolveExistingDocumentPath. A missing
// selected root, dangling link, inaccessible parent, or non-directory parent is
// an error; none is treated as a new destination.
func ResolveDocumentDestination(rootDir, documentsDir, logicalPath string) (ResolvedDocumentPath, error) {
	return resolveDocumentIdentity(rootDir, documentsDir, logicalPath, true)
}

// CanonicalDocumentRequestPath validates URL input using the platform's path
// rules before cleaning it. It performs no lookup and does not establish a
// filesystem identity; use a resolver before authorization or file access.
func CanonicalDocumentRequestPath(logicalPath string) (string, error) {
	normalized := strings.ReplaceAll(logicalPath, "\\", "/")
	// Logical paths may start with one URL slash, but never a UNC/device prefix
	// or a drive-qualified path, including a drive-relative spelling such as C:x.
	withoutSlash := strings.TrimPrefix(normalized, "/")
	if strings.HasPrefix(normalized, "//") || len(withoutSlash) >= 2 &&
		((withoutSlash[0] >= 'A' && withoutSlash[0] <= 'Z') || (withoutSlash[0] >= 'a' && withoutSlash[0] <= 'z')) && withoutSlash[1] == ':' {
		return "", fmt.Errorf("document path contains a filesystem namespace")
	}
	if err := validateIdentityPath(normalized); err != nil {
		return "", err
	}
	return CanonicalRequestPath(normalized)
}

func resolveDocumentIdentity(rootDir, documentsDir, logicalPath string, allowMissing bool) (ResolvedDocumentPath, error) {
	canonical, err := CanonicalDocumentRequestPath(logicalPath)
	if err != nil {
		return ResolvedDocumentPath{}, err
	}
	root, err := filepath.Abs(filepath.Join(rootDir, documentsDir))
	if err != nil {
		return ResolvedDocumentPath{}, fmt.Errorf("resolve documents root: %w", err)
	}
	root, err = existingPathIdentity(root)
	if err != nil {
		return ResolvedDocumentPath{}, fmt.Errorf("resolve documents root identity: %w", err)
	}
	current := root
	components := strings.Split(strings.TrimPrefix(canonical, "/"), "/")
	for index, component := range components {
		info, err := os.Stat(current)
		if err != nil {
			return ResolvedDocumentPath{}, fmt.Errorf("stat document parent: %w", err)
		}
		if !info.IsDir() {
			return ResolvedDocumentPath{}, fmt.Errorf("document parent is not a directory")
		}
		if component == "" { // The selected root itself.
			break
		}
		candidate := filepath.Join(current, filepath.FromSlash(component))
		if allowMissing {
			// Lstat distinguishes a missing component from a dangling symlink.
			if _, err := os.Lstat(candidate); os.IsNotExist(err) {
				current = filepath.Join(current, filepath.Join(components[index:]...))
				break
			} else if err != nil {
				return ResolvedDocumentPath{}, fmt.Errorf("inspect document destination: %w", err)
			}
		}
		current, err = existingPathIdentity(candidate)
		if err != nil {
			return ResolvedDocumentPath{}, fmt.Errorf("resolve document identity: %w", err)
		}
		// Check every resolved prefix, including links that might leave the root
		// and then point back inside through another component.
		if _, err := containedIdentityRelative(root, current); err != nil {
			return ResolvedDocumentPath{}, err
		}
	}
	relative, err := containedIdentityRelative(root, current)
	if err != nil {
		return ResolvedDocumentPath{}, err
	}
	logical := "/"
	if relative != "." {
		logical += filepath.ToSlash(relative)
	}
	// A link target must also be representable by an unambiguous URL path.
	if err := validateIdentityPath(logical); err != nil {
		return ResolvedDocumentPath{}, err
	}
	if clean, err := CanonicalRequestPath(logical); err != nil || clean != logical {
		return ResolvedDocumentPath{}, fmt.Errorf("resolved identity is not a canonical logical path")
	}
	return ResolvedDocumentPath{LogicalPath: logical, FilesystemPath: current}, nil
}

func containedIdentityRelative(root, candidate string) (string, error) {
	relative, err := filepath.Rel(root, candidate)
	if err != nil {
		return "", fmt.Errorf("resolve identity relative to documents root: %w", err)
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) || filepath.Join(root, relative) != candidate {
		// Both paths have already been resolved to stored names. An exact prefix
		// also avoids conflating distinct Windows directories with case
		// sensitivity enabled, where filepath.Rel compares case-insensitively.
		return "", fmt.Errorf("document identity escapes the documents root")
	}
	return relative, nil
}

// CanonicalRequestPath converts the path already decoded once by net/http into
// the single logical form used for authorization and filesystem resolution.
func CanonicalRequestPath(requestPath string) (string, error) {
	if requestPath == "" {
		requestPath = "/"
	}
	if strings.ContainsRune(requestPath, '\x00') {
		return "", fmt.Errorf("path contains a null byte")
	}

	requestPath = strings.ReplaceAll(requestPath, "\\", "/")
	if !strings.HasPrefix(requestPath, "/") {
		requestPath = "/" + requestPath
	}

	// A valid percent escape here was encoded twice on the wire. Reject it so no
	// later decoder can turn the authorized path into a different filesystem path.
	if containsPercentEscape(requestPath) {
		return "", fmt.Errorf("path contains residual percent encoding")
	}

	for _, segment := range strings.Split(requestPath, "/") {
		if segment == ".." {
			return "", fmt.Errorf("path contains parent traversal")
		}
	}

	canonical := path.Clean(requestPath)
	if !strings.HasPrefix(canonical, "/") {
		return "", fmt.Errorf("path is not absolute")
	}
	return canonical, nil
}

// ResolveRelativeDocumentPath resolves a relative URL document path below the
// configured documents root. URL-derived paths must be accepted and
// canonicalized here before any part of them is passed to filepath.Join.
// This legacy helper checks lexical containment only. New identity-aware callers
// must choose ResolveExistingDocumentPath or ResolveDocumentDestination.
func ResolveRelativeDocumentPath(rootDir, documentsDir, relativePath string) (ResolvedDocumentPath, error) {
	if relativePath == "" {
		return ResolvedDocumentPath{}, fmt.Errorf("document path is empty")
	}

	normalized := strings.ReplaceAll(relativePath, "\\", "/")
	if isAbsoluteRelativePath(normalized) {
		return ResolvedDocumentPath{}, fmt.Errorf("document path is absolute")
	}
	for _, segment := range strings.Split(normalized, "/") {
		if segment == ".." {
			return ResolvedDocumentPath{}, fmt.Errorf("document path contains parent traversal")
		}
	}

	canonical, err := CanonicalRequestPath(normalized)
	if err != nil {
		return ResolvedDocumentPath{}, err
	}
	containedRelativePath := strings.TrimPrefix(canonical, "/")
	if containedRelativePath == "" {
		return ResolvedDocumentPath{}, fmt.Errorf("document path is empty")
	}

	documentsRoot, err := filepath.Abs(filepath.Join(rootDir, documentsDir))
	if err != nil {
		return ResolvedDocumentPath{}, fmt.Errorf("resolve documents root: %w", err)
	}
	candidate := filepath.Join(documentsRoot, filepath.FromSlash(containedRelativePath))
	relative, err := filepath.Rel(documentsRoot, candidate)
	if err != nil {
		return ResolvedDocumentPath{}, fmt.Errorf("resolve document path: %w", err)
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return ResolvedDocumentPath{}, fmt.Errorf("document path escapes the documents root")
	}
	return ResolvedDocumentPath{LogicalPath: canonical, FilesystemPath: candidate}, nil
}

// ResolveDocumentPath resolves a canonical logical path below the configured
// documents root and verifies containment before returning a filesystem path.
func ResolveDocumentPath(rootDir, documentsDir, logicalPath string) (string, error) {
	canonical, err := CanonicalRequestPath(logicalPath)
	if err != nil {
		return "", err
	}
	resolved, err := ResolveRelativeDocumentPath(rootDir, documentsDir, strings.TrimPrefix(canonical, "/"))
	if err != nil {
		return "", err
	}
	return resolved.FilesystemPath, nil
}

func isAbsoluteRelativePath(value string) bool {
	if path.IsAbs(value) || filepath.IsAbs(filepath.FromSlash(value)) {
		return true
	}
	return len(value) >= 3 &&
		((value[0] >= 'A' && value[0] <= 'Z') || (value[0] >= 'a' && value[0] <= 'z')) &&
		value[1] == ':' && value[2] == '/'
}

func containsPercentEscape(value string) bool {
	for index := 0; index+2 < len(value); index++ {
		if value[index] == '%' && isHexDigit(value[index+1]) && isHexDigit(value[index+2]) {
			return true
		}
	}
	return false
}

func isHexDigit(value byte) bool {
	return value >= '0' && value <= '9' || value >= 'a' && value <= 'f' || value >= 'A' && value <= 'F'
}
