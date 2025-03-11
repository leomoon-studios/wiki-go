package utils

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"wiki-go/internal/types"
)

// NavItem represents a navigation item (directory)
type NavItem struct {
	Title    string
	Path     string
	IsDir    bool
	Children []*NavItem
	IsActive bool
}

// GetDocumentTitle extracts the first H1 title from document.md
func GetDocumentTitle(dirPath string) string {
	docPath := filepath.Join(dirPath, "document.md")
	file, err := os.Open(docPath)
	if err != nil {
		// If no document.md or can't read it, use directory name
		return FormatDirName(filepath.Base(dirPath))
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}

	// If no H1 found, use directory name
	return FormatDirName(filepath.Base(dirPath))
}

// FormatDirName formats a directory name by replacing dashes with spaces and title casing
func FormatDirName(name string) string {
	// Replace dashes with spaces
	name = strings.ReplaceAll(name, "-", " ")

	// Title case the words
	words := strings.Fields(name)
	for i, word := range words {
		words[i] = strings.Title(word)
	}

	return strings.Join(words, " ")
}

// ToURLPath converts a filesystem path to a URL path
func ToURLPath(path string) string {
	// Convert spaces to dashes
	return strings.ReplaceAll(path, " ", "-")
}

// BuildNavigation builds the navigation structure from the root directory
func BuildNavigation(rootDir string, documentsDir string) (*types.NavItem, error) {
	root := &types.NavItem{
		Title:    "Wiki-Go",
		Path:     "/",
		IsDir:    true,
		Children: make([]*types.NavItem, 0),
	}

	// Create the documents directory path
	docsPath := filepath.Join(rootDir, documentsDir)

	// Check if documents directory exists
	if _, err := os.Stat(docsPath); os.IsNotExist(err) {
		// Create documents directory if it doesn't exist
		if err := os.MkdirAll(docsPath, 0755); err != nil {
			return nil, err
		}
	}

	err := filepath.Walk(docsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the documents directory itself and non-directories
		if path == docsPath || !info.IsDir() {
			return nil
		}

		// Skip hidden directories and document.md files
		if strings.HasPrefix(filepath.Base(path), ".") || filepath.Base(path) == "document.md" {
			return nil
		}

		// Skip the pages/home directory in navigation
		if path == filepath.Join(rootDir, "pages", "home") || path == filepath.Join(rootDir, "pages") {
			return filepath.SkipDir
		}

		// Create relative path for the URL
		relPath := strings.TrimPrefix(path, docsPath)
		relPath = strings.TrimPrefix(relPath, string(os.PathSeparator))
		relPath = filepath.ToSlash(relPath)

		// Get the title from document.md's H1 or fallback to formatted directory name
		title := GetDocumentTitle(path)

		// Split the path into components
		parts := strings.Split(relPath, "/")
		current := root

		// Build the directory structure
		for i := 0; i < len(parts); i++ {
			// Create URL path with dashes
			urlPath := "/" + ToURLPath(filepath.ToSlash(filepath.Join(parts[:i+1]...)))

			// Look for existing directory at this level
			var found *types.NavItem
			for _, child := range current.Children {
				if child.Path == urlPath {
					found = child
					break
				}
			}

			if found == nil {
				// Create new directory item
				dirTitle := ""
				if i == len(parts)-1 {
					dirTitle = title // Use document.md title for leaf nodes
				} else {
					dirTitle = FormatDirName(parts[i])
				}

				found = &types.NavItem{
					Title:    dirTitle,
					Path:     urlPath,
					IsDir:    true,
					Children: make([]*types.NavItem, 0),
				}
				current.Children = append(current.Children, found)
			}
			current = found
		}

		return nil
	})

	return root, err
}

// FindNavItem finds a navigation item by its path
func FindNavItem(root *types.NavItem, path string) *types.NavItem {
	if root == nil {
		return nil
	}

	// Clean up the path
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		path = "/"
	}

	if root.Path == path {
		return root
	}

	for _, child := range root.Children {
		if found := FindNavItem(child, path); found != nil {
			return found
		}
	}

	return nil
}

// MarkActiveNavItem marks the active navigation item and its parents
func MarkActiveNavItem(root *types.NavItem, currentPath string) {
	if root == nil {
		return
	}

	// Clean up the path
	currentPath = strings.TrimSuffix(currentPath, "/")
	if currentPath == "" {
		currentPath = "/"
	}

	// Mark this item if it matches
	if root.Path == currentPath {
		root.IsActive = true
	}

	// Mark this item if any child is active
	for _, child := range root.Children {
		MarkActiveNavItem(child, currentPath)
		if child.IsActive {
			root.IsActive = true
		}
	}
}
