package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"wiki-go/internal/config"
	"wiki-go/internal/i18n"
	"wiki-go/internal/types"
	"wiki-go/internal/utils"
)

// PageHandler handles requests for pages
func PageHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// Add cache control headers to prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Get the requested path
	path := r.URL.Path
	if path == "/" {
		HomeHandler(w, r, cfg)
		return
	}

	// Clean and decode the path
	path = filepath.Clean(path)
	path = strings.TrimSuffix(path, "/")
	path = strings.ReplaceAll(path, "\\", "/")
	decodedPath, err := url.QueryUnescape(path)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// Build navigation
	nav, err := utils.BuildNavigation(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Mark active navigation item
	utils.MarkActiveNavItem(nav, path)

	// Get the full filesystem path - adjust to use documents subdirectory
	fsPath := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, decodedPath)

	// Check if path exists
	info, err := os.Stat(fsPath)
	if err != nil || !info.IsDir() {
		http.NotFound(w, r)
		return
	}

	// Find the navigation item for breadcrumbs
	navItem := utils.FindNavItem(nav, path)
	if navItem == nil {
		navItem = &types.NavItem{
			Title: utils.FormatDirName(filepath.Base(decodedPath)),
			Path:  path,
			IsDir: true,
		}
	}

	// Generate breadcrumbs
	breadcrumbs := generateBreadcrumbs(nav, path)

	var content template.HTML
	var lastModified time.Time
	var dirContent template.HTML

	// Look for document.md in the directory
	docPath := filepath.Join(fsPath, "document.md")
	docInfo, err := os.Stat(docPath)
	if err == nil {
		// Read and render document.md if it exists
		mdContent, err := os.ReadFile(docPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Use the document path for rendering to handle local file references
		content = template.HTML(utils.RenderMarkdownWithPath(string(mdContent), decodedPath))
		lastModified = docInfo.ModTime()
	}

	// List directory contents
	files, err := os.ReadDir(fsPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build directory listing HTML
	var dirItems []string
	for _, f := range files {
		if !f.IsDir() || strings.HasPrefix(f.Name(), ".") || f.Name() == "document.md" {
			continue // Skip non-directories, hidden files, and document.md
		}

		dirName := f.Name()
		urlPath := filepath.Join(path, dirName)

		// Check if subdirectory has a document.md
		subDocPath := filepath.Join(fsPath, dirName, "document.md")
		if _, err := os.Stat(subDocPath); err == nil {
			// Read the document.md to get its title
			mdContent, err := os.ReadFile(subDocPath)
			if err == nil {
				// Extract first line and remove # prefix for title
				lines := strings.Split(string(mdContent), "\n")
				if len(lines) > 0 {
					title := strings.TrimSpace(lines[0])
					if strings.HasPrefix(title, "# ") {
						dirTitle := strings.TrimSpace(strings.TrimPrefix(title, "# "))
						dirItems = append(dirItems, fmt.Sprintf(`<div class="directory-item is-dir"><a href="%s">%s</a></div>`,
							urlPath, dirTitle))
						continue
					}
				}
			}
		}

		// Fallback to formatted directory name if no document.md or no title found
		dirTitle := utils.FormatDirName(dirName)
		dirItems = append(dirItems, fmt.Sprintf(`<div class="directory-item is-dir"><a href="%s">%s</a></div>`,
			urlPath, dirTitle))
	}

	if len(dirItems) > 0 {
		dirContent = template.HTML(strings.Join(dirItems, "\n"))
	}

	// If no document.md exists, show directory title and listing
	if docInfo == nil {
		content = template.HTML(fmt.Sprintf("<h1>%s</h1>", navItem.Title))
		lastModified = info.ModTime()
	}

	// Prepare template data
	data := &types.PageData{
		Navigation:         nav,
		Content:            content,
		DirContent:         dirContent,
		Breadcrumbs:        breadcrumbs,
		Config:             cfg,
		LastModified:       lastModified,
		CurrentDir:         navItem,
		AvailableLanguages: i18n.GetAvailableLanguages(),
	}

	renderTemplate(w, data)
}

// generateBreadcrumbs creates a breadcrumb trail from a path
func generateBreadcrumbs(nav *types.NavItem, path string) []types.BreadcrumbItem {
	if path == "" || path == "/" {
		return []types.BreadcrumbItem{{Title: "Home", Path: "/", IsLast: true}}
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	breadcrumbs := make([]types.BreadcrumbItem, 0, len(parts)+1)

	// Always start with Home
	breadcrumbs = append(breadcrumbs, types.BreadcrumbItem{
		Title:  "Home",
		Path:   "/",
		IsLast: false,
	})

	currentPath := ""
	for i, part := range parts {
		if currentPath == "" {
			currentPath = part
		} else {
			currentPath = currentPath + "/" + part
		}

		item := utils.FindNavItem(nav, "/"+currentPath)
		if item != nil {
			breadcrumbs = append(breadcrumbs, types.BreadcrumbItem{
				Title:  item.Title,
				Path:   "/" + currentPath,
				IsLast: i == len(parts)-1,
			})
		}
	}

	return breadcrumbs
}
