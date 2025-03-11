package utils

import (
	"path/filepath"
	"regexp"
	"strings"
)

// SanitizePath cleans a path to ensure it's safe and follows the expected format
func SanitizePath(path string) string {
	// Remove leading/trailing slashes
	path = strings.Trim(path, "/")

	// Replace any unsafe characters with dashes
	re := regexp.MustCompile(`[^a-zA-Z0-9_\-/]`)
	path = re.ReplaceAllString(path, "-")

	// Replace consecutive slashes with a single slash
	re = regexp.MustCompile(`/+`)
	path = re.ReplaceAllString(path, "/")

	// Ensure no path traversal
	path = filepath.Clean(path)

	// Remove any leading ../ sequences
	for strings.HasPrefix(path, "../") || strings.HasPrefix(path, "..\\") {
		path = strings.TrimPrefix(path, "../")
		path = strings.TrimPrefix(path, "..\\")
	}

	// Remove any leading ./ sequences
	for strings.HasPrefix(path, "./") || strings.HasPrefix(path, ".\\") {
		path = strings.TrimPrefix(path, "./")
		path = strings.TrimPrefix(path, ".\\")
	}

	return path
}

// IsNumeric checks if a string contains only digits
func IsNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
