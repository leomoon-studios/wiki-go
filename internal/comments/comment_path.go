package comments

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"unicode"
)

type resolvedCommentPath struct {
	directoryPath string
	filePath      string
	logicalPath   string
	commentID     string
}

func resolveCommentDirectory(commentsRoot, logicalDocumentPath string) (resolvedCommentPath, error) {
	if commentsRoot == "" {
		return resolvedCommentPath{}, fmt.Errorf("comments root is empty")
	}
	if logicalDocumentPath == "" {
		return resolvedCommentPath{}, fmt.Errorf("document path is empty")
	}

	normalized := strings.ReplaceAll(logicalDocumentPath, "\\", "/")
	if isWindowsAbsoluteCommentPath(normalized) || strings.HasPrefix(normalized, "//") {
		return resolvedCommentPath{}, fmt.Errorf("document path is absolute")
	}
	canonical, err := canonicalCommentDocumentPath(normalized)
	if err != nil {
		return resolvedCommentPath{}, fmt.Errorf("canonicalize comment document path: %w", err)
	}

	rootPath, err := filepath.Abs(commentsRoot)
	if err != nil {
		return resolvedCommentPath{}, fmt.Errorf("resolve comments root: %w", err)
	}
	relativeDocumentPath := strings.TrimPrefix(canonical, "/")
	directoryPath := filepath.Join(rootPath, filepath.FromSlash(relativeDocumentPath))
	relativePath, err := filepath.Rel(rootPath, directoryPath)
	if err != nil {
		return resolvedCommentPath{}, fmt.Errorf("resolve comment directory: %w", err)
	}
	if relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) || filepath.IsAbs(relativePath) {
		return resolvedCommentPath{}, fmt.Errorf("comment directory escapes comments root")
	}

	return resolvedCommentPath{
		directoryPath: directoryPath,
		logicalPath:   canonical,
	}, nil
}

func canonicalCommentDocumentPath(documentPath string) (string, error) {
	if strings.ContainsRune(documentPath, '\x00') {
		return "", fmt.Errorf("document path contains a null byte")
	}
	if !strings.HasPrefix(documentPath, "/") {
		documentPath = "/" + documentPath
	}
	if containsCommentPercentEscape(documentPath) {
		return "", fmt.Errorf("document path contains residual percent encoding")
	}
	for _, segment := range strings.Split(documentPath, "/") {
		if segment == ".." {
			return "", fmt.Errorf("document path contains parent traversal")
		}
	}
	canonical := path.Clean(documentPath)
	if !strings.HasPrefix(canonical, "/") {
		return "", fmt.Errorf("document path is not logical")
	}
	return canonical, nil
}

func containsCommentPercentEscape(value string) bool {
	for index := 0; index+2 < len(value); index++ {
		if value[index] == '%' && isCommentHexDigit(value[index+1]) && isCommentHexDigit(value[index+2]) {
			return true
		}
	}
	return false
}

func isCommentHexDigit(value byte) bool {
	return value >= '0' && value <= '9' || value >= 'a' && value <= 'f' || value >= 'A' && value <= 'F'
}

func resolveCommentFile(commentsRoot, logicalDocumentPath, commentID string) (resolvedCommentPath, error) {
	if !isValidCommentID(commentID) {
		return resolvedCommentPath{}, fmt.Errorf("invalid comment ID")
	}
	resolved, err := resolveCommentDirectory(commentsRoot, logicalDocumentPath)
	if err != nil {
		return resolvedCommentPath{}, err
	}

	filePath := filepath.Join(resolved.directoryPath, commentID)
	relativePath, err := filepath.Rel(resolved.directoryPath, filePath)
	if err != nil {
		return resolvedCommentPath{}, fmt.Errorf("resolve comment file: %w", err)
	}
	if relativePath != commentID || filepath.IsAbs(relativePath) {
		return resolvedCommentPath{}, fmt.Errorf("comment file escapes document directory")
	}

	resolved.filePath = filePath
	resolved.commentID = commentID
	return resolved, nil
}

func isWindowsAbsoluteCommentPath(value string) bool {
	return len(value) >= 3 &&
		((value[0] >= 'A' && value[0] <= 'Z') || (value[0] >= 'a' && value[0] <= 'z')) &&
		value[1] == ':' && value[2] == '/'
}

func hasCommentFilenameControlCharacter(value string) bool {
	for _, character := range value {
		if unicode.IsControl(character) || character == '\u2028' || character == '\u2029' {
			return true
		}
	}
	return false
}
