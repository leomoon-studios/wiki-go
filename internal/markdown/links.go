package markdown

import (
	"path/filepath"
	"strings"
	"wiki-go/internal/config"
)

// IsLocalFileReference checks if a file reference is a local path
func IsLocalFileReference(path string) bool {
	// Simple check: if it doesn't start with http/https/ftp and doesn't have a colon,
	// it's likely a local reference
	path = strings.TrimSpace(path)
	return !strings.HasPrefix(path, "http://") &&
		!strings.HasPrefix(path, "https://") &&
		!strings.HasPrefix(path, "ftp://") &&
		!strings.Contains(path, "://") &&
		path != ""
}

// TransformLocalFileReference transforms a local file reference to a full URL
func TransformLocalFileReference(path string, docPath string) string {
	// Handle relative paths based on the document path
	// This is a simplified version - you might need more complex logic
	// depending on your application
	if strings.HasPrefix(path, "/") {
		// Absolute path from content root
		return "/api/files" + path
	}

	// Relative path - construct URL based on document path
	return "/api/files" + docPath + "/" + path
}

// IsLocalFileReference checks if a link is a simple filename reference to a local file
// Returns true if the link is a simple filename with a supported extension
func IsLocalFileReferenceOld(link string) bool {
	// If it's empty or starts with a scheme or path separator, it's not a local file reference
	if link == "" || strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") ||
		strings.HasPrefix(link, "/") || strings.Contains(link, "\\") || strings.Contains(link, "/") {
		return false
	}

	// Check if the extension is supported using the centralized config
	ext := strings.ToLower(filepath.Ext(link))
	return config.IsAllowedExtension(ext)
}

// TransformLocalFileReference transforms a simple filename to a full URL based on the document path
// docPath should be the path to the document directory (not including the document.md file)
func TransformLocalFileReferenceOld(filename string, docPath string) string {
	// Clean and normalize the document path
	docPath = strings.TrimPrefix(docPath, "/")
	docPath = strings.TrimSuffix(docPath, "/")

	// Build the full URL
	urlPath := filepath.Join("/api/files", docPath, filename)

	// Replace backslashes with forward slashes for URLs
	urlPath = strings.ReplaceAll(urlPath, "\\", "/")

	return urlPath
}
