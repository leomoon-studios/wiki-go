package utils

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

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

// ResolveDocumentPath resolves a canonical logical path below the configured
// documents root and verifies containment before returning a filesystem path.
func ResolveDocumentPath(rootDir, documentsDir, logicalPath string) (string, error) {
	canonical, err := CanonicalRequestPath(logicalPath)
	if err != nil {
		return "", err
	}

	documentsRoot, err := filepath.Abs(filepath.Join(rootDir, documentsDir))
	if err != nil {
		return "", fmt.Errorf("resolve documents root: %w", err)
	}
	candidate := filepath.Join(documentsRoot, filepath.FromSlash(strings.TrimPrefix(canonical, "/")))
	relative, err := filepath.Rel(documentsRoot, candidate)
	if err != nil {
		return "", fmt.Errorf("resolve document path: %w", err)
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return "", fmt.Errorf("document path escapes the documents root")
	}
	return candidate, nil
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
