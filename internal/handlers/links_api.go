package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"wiki-go/internal/auth"
	"wiki-go/internal/config"
	"wiki-go/internal/frontmatter"
	"wiki-go/internal/logger"
	"wiki-go/internal/utils"
)

// LinkRequest represents the JSON payload for link operations
type LinkRequest struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Date        string `json:"date,omitempty"` // Optional, format: YYYY-MM-DD
}

// LinkResponse represents the JSON response after link operations
type LinkResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// AddLinkHandler handles POST requests to add a new link to a links document
func AddLinkHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		sendLinkError(w, "Method not allowed", http.StatusMethodNotAllowed, "")
		return
	}

	// Authentication is handled by editorMiddleware, so we can proceed directly

	// Get document path from URL
	path := strings.TrimPrefix(r.URL.Path, "/api/links/add/")
	if path == "" {
		sendLinkError(w, "Document path is required", http.StatusBadRequest, "")
		return
	}

	// Authorize the once-decoded URL path and retain its resolved identity.
	session := auth.GetSession(r)
	if session == nil || (session.Role != config.RoleAdmin && session.Role != config.RoleEditor) {
		sendLinkError(w, "Editor access required", http.StatusUnauthorized, "")
		return
	}
	document, ok := requireAPIDocument(w, cfg, path, session, false, true)
	if !ok {
		return
	}
	docPath, err := documentContentPath(document.ResolvedDocumentPath)
	if err != nil {
		sendLinkError(w, "Invalid document file", http.StatusBadRequest, "")
		return
	}
	path = document.LogicalPath

	// Parse request body
	var req LinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendLinkError(w, "Invalid request payload", http.StatusBadRequest, err.Error())
		return
	}

	// Validate request
	if err := validateLinkRequest(req); err != nil {
		sendLinkError(w, "Validation failed", http.StatusBadRequest, err.Error())
		return
	}

	// Get document path
	// docPath is the identity authorized above.

	// Read current document
	content, err := os.ReadFile(docPath)
	if err != nil {
		sendLinkError(w, "Failed to read document", http.StatusInternalServerError, err.Error())
		return
	}

	// Parse current content to get links data
	linksData, err := frontmatter.ParseLinksContent(string(content))
	if err != nil {
		sendLinkError(w, "Failed to parse links document", http.StatusBadRequest, err.Error())
		return
	}

	// Create new link
	newLink := frontmatter.Link{
		Title:       req.Title,
		URL:         req.URL,
		Description: req.Description,
		Category:    req.Category,
		AddedAt:     parseDate(req.Date),
	}

	// Add link to the appropriate category
	if linksData.Categories == nil {
		linksData.Categories = make(map[string][]frontmatter.Link)
	}
	linksData.Categories[req.Category] = append(linksData.Categories[req.Category], newLink)

	// Generate updated markdown
	updatedContent, err := generateLinksMarkdown(linksData, string(content))
	if err != nil {
		sendLinkError(w, "Failed to generate updated content", http.StatusInternalServerError, err.Error())
		return
	}

	// Save document with version control
	if err := saveDocumentWithVersioning(docPath, path, []byte(updatedContent)); err != nil {
		sendLinkError(w, "Failed to save document", http.StatusInternalServerError, err.Error())
		return
	}

	// Success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LinkResponse{
		Success: true,
		Message: "Link added successfully",
	})
}

// EditLinkHandler handles PUT requests to edit an existing link
func EditLinkHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPut {
		sendLinkError(w, "Method not allowed", http.StatusMethodNotAllowed, "")
		return
	}

	// Authentication is handled by editorMiddleware, so we can proceed directly

	// Get document path from URL
	path := strings.TrimPrefix(r.URL.Path, "/api/links/edit/")
	if path == "" {
		sendLinkError(w, "Document path is required", http.StatusBadRequest, "")
		return
	}

	// Authorize the once-decoded URL path and retain its resolved identity.
	session := auth.GetSession(r)
	if session == nil || (session.Role != config.RoleAdmin && session.Role != config.RoleEditor) {
		sendLinkError(w, "Editor access required", http.StatusUnauthorized, "")
		return
	}
	document, ok := requireAPIDocument(w, cfg, path, session, false, true)
	if !ok {
		return
	}
	docPath, err := documentContentPath(document.ResolvedDocumentPath)
	if err != nil {
		sendLinkError(w, "Invalid document file", http.StatusBadRequest, "")
		return
	}
	path = document.LogicalPath

	// Parse request body (includes old and new link data)
	var req struct {
		OldURL  string      `json:"oldUrl"`
		NewLink LinkRequest `json:"newLink"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendLinkError(w, "Invalid request payload", http.StatusBadRequest, err.Error())
		return
	}

	// Validate new link data
	if err := validateLinkRequest(req.NewLink); err != nil {
		sendLinkError(w, "Validation failed", http.StatusBadRequest, err.Error())
		return
	}

	// Get document path
	// docPath is the identity authorized above.

	// Read current document
	content, err := os.ReadFile(docPath)
	if err != nil {
		sendLinkError(w, "Failed to read document", http.StatusInternalServerError, err.Error())
		return
	}

	// Parse current content
	linksData, err := frontmatter.ParseLinksContent(string(content))
	if err != nil {
		sendLinkError(w, "Failed to parse links document", http.StatusBadRequest, err.Error())
		return
	}

	// Find and update the link
	linkFound := false
	for category, links := range linksData.Categories {
		for i, link := range links {
			if link.URL == req.OldURL {
				// Update the link
				linksData.Categories[category][i] = frontmatter.Link{
					Title:       req.NewLink.Title,
					URL:         req.NewLink.URL,
					Description: req.NewLink.Description,
					Category:    req.NewLink.Category,
					AddedAt:     parseDate(req.NewLink.Date),
				}

				// If category changed, move the link
				if req.NewLink.Category != category {
					// Remove from old category
					linksData.Categories[category] = append(links[:i], links[i+1:]...)
					// Add to new category
					if linksData.Categories[req.NewLink.Category] == nil {
						linksData.Categories[req.NewLink.Category] = []frontmatter.Link{}
					}
					linksData.Categories[req.NewLink.Category] = append(
						linksData.Categories[req.NewLink.Category],
						linksData.Categories[category][i])
				}

				linkFound = true
				break
			}
		}
		if linkFound {
			break
		}
	}

	if !linkFound {
		sendLinkError(w, "Link not found", http.StatusNotFound, "")
		return
	}

	// Generate updated markdown
	updatedContent, err := generateLinksMarkdown(linksData, string(content))
	if err != nil {
		sendLinkError(w, "Failed to generate updated content", http.StatusInternalServerError, err.Error())
		return
	}

	// Save document with version control
	if err := saveDocumentWithVersioning(docPath, path, []byte(updatedContent)); err != nil {
		sendLinkError(w, "Failed to save document", http.StatusInternalServerError, err.Error())
		return
	}

	// Success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LinkResponse{
		Success: true,
		Message: "Link updated successfully",
	})
}

// DeleteLinkHandler handles DELETE requests to remove a link
func DeleteLinkHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodDelete {
		sendLinkError(w, "Method not allowed", http.StatusMethodNotAllowed, "")
		return
	}

	// Authentication is handled by editorMiddleware, so we can proceed directly

	// Get document path from URL
	path := strings.TrimPrefix(r.URL.Path, "/api/links/delete/")
	if path == "" {
		sendLinkError(w, "Document path is required", http.StatusBadRequest, "")
		return
	}

	// Authorize the once-decoded URL path and retain its resolved identity.
	session := auth.GetSession(r)
	if session == nil || (session.Role != config.RoleAdmin && session.Role != config.RoleEditor) {
		sendLinkError(w, "Editor access required", http.StatusUnauthorized, "")
		return
	}
	document, ok := requireAPIDocument(w, cfg, path, session, false, true)
	if !ok {
		return
	}
	docPath, err := documentContentPath(document.ResolvedDocumentPath)
	if err != nil {
		sendLinkError(w, "Invalid document file", http.StatusBadRequest, "")
		return
	}
	path = document.LogicalPath

	// Parse request body (link data for precise matching)
	var req struct {
		URL         string `json:"url"`
		Title       string `json:"title,omitempty"`
		Description string `json:"description,omitempty"`
		Category    string `json:"category,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendLinkError(w, "Invalid request payload", http.StatusBadRequest, err.Error())
		return
	}

	if req.URL == "" {
		sendLinkError(w, "URL is required", http.StatusBadRequest, "")
		return
	}

	// Get document path
	// docPath is the identity authorized above.

	// Read current document
	content, err := os.ReadFile(docPath)
	if err != nil {
		sendLinkError(w, "Failed to read document", http.StatusInternalServerError, err.Error())
		return
	}

	// Parse current content
	linksData, err := frontmatter.ParseLinksContent(string(content))
	if err != nil {
		sendLinkError(w, "Failed to parse links document", http.StatusBadRequest, err.Error())
		return
	}

	// Find and remove the specific link with more precise matching
	linkFound := false
	for category, links := range linksData.Categories {
		for i, link := range links {
			// First check URL match
			if link.URL == req.URL {
				// If category is provided, verify it matches
				if req.Category != "" && category != req.Category {
					continue // Skip this link, wrong category
				}

				// If title is provided, verify it matches
				if req.Title != "" && link.Title != req.Title {
					continue // Skip this link, wrong title
				}

				// If description is provided, verify it matches
				if req.Description != "" && link.Description != req.Description {
					continue // Skip this link, wrong description
				}

				// All checks passed, remove this link
				linksData.Categories[category] = append(links[:i], links[i+1:]...)
				linkFound = true
				break
			}
		}
		if linkFound {
			break
		}
	}

	if !linkFound {
		sendLinkError(w, "Link not found", http.StatusNotFound, "")
		return
	}

	// Generate updated markdown
	updatedContent, err := generateLinksMarkdown(linksData, string(content))
	if err != nil {
		sendLinkError(w, "Failed to generate updated content", http.StatusInternalServerError, err.Error())
		return
	}

	// Save document with version control
	if err := saveDocumentWithVersioning(docPath, path, []byte(updatedContent)); err != nil {
		sendLinkError(w, "Failed to save document", http.StatusInternalServerError, err.Error())
		return
	}

	// Success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LinkResponse{
		Success: true,
		Message: "Link deleted successfully",
	})
}

// Helper functions

func sendLinkError(w http.ResponseWriter, message string, statusCode int, error string) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(LinkResponse{
		Success: false,
		Message: message,
		Error:   error,
	})
}

func validateLinkRequest(req LinkRequest) error {
	if req.URL == "" {
		return fmt.Errorf("URL is required")
	}
	if req.Title == "" {
		return fmt.Errorf("Title is required")
	}
	if req.Category == "" {
		return fmt.Errorf("Category is required")
	}

	// Validate URL format
	if err := frontmatter.ValidateURL(req.URL); err != nil {
		return err
	}

	// Validate the full link structure
	link := frontmatter.Link{
		URL:      req.URL,
		Title:    req.Title,
		Category: req.Category,
	}

	if errors := frontmatter.ValidateLink(link); len(errors) > 0 {
		return errors[0] // Return the first error
	}

	return nil
}

func parseDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Now()
	}

	if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
		return parsed
	}

	return time.Now()
}

func generateLinksMarkdown(linksData *frontmatter.LinksData, originalContent string) (string, error) {
	// Extract frontmatter and title from original content
	lines := strings.Split(originalContent, "\n")

	var frontmatterLines []string
	var title string
	contentStartIndex := 0

	// Handle frontmatter
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		frontmatterLines = append(frontmatterLines, lines[0])
		for i := 1; i < len(lines); i++ {
			frontmatterLines = append(frontmatterLines, lines[i])
			if strings.TrimSpace(lines[i]) == "---" {
				contentStartIndex = i + 1
				break
			}
		}
	}

	// Find title (first # heading)
	for i := contentStartIndex; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "# ") {
			title = strings.TrimPrefix(line, "# ")
			contentStartIndex = i + 1
			break
		}
	}

	// Build new content
	var result strings.Builder

	// Add frontmatter
	if len(frontmatterLines) > 0 {
		for _, line := range frontmatterLines {
			result.WriteString(line)
			result.WriteString("\n")
		}
		result.WriteString("\n")
	}

	// Add title
	if title != "" {
		result.WriteString("# ")
		result.WriteString(title)
		result.WriteString("\n\n")
	}

	// Add categories and links
	for category, links := range linksData.Categories {
		if len(links) == 0 {
			continue
		}

		result.WriteString("## ")
		result.WriteString(category)
		result.WriteString("\n")

		for _, link := range links {
			result.WriteString("- [")
			result.WriteString(link.Title)
			result.WriteString("](")
			result.WriteString(link.URL)
			result.WriteString(")")

			if link.Description != "" {
				result.WriteString(" - ")
				result.WriteString(link.Description)
			}

			if !link.AddedAt.IsZero() {
				result.WriteString(" | ")
				result.WriteString(link.AddedAt.Format("2006-01-02"))
			}

			result.WriteString("\n")
		}
		result.WriteString("\n")
	}

	return result.String(), nil
}

func saveDocumentWithVersioning(docPath, relativePath string, content []byte) error {
	// VERSION CONTROL: Save current version before overwriting (same logic as SaveHandler)
	if _, err := os.Stat(docPath); err == nil && cfg.Wiki.MaxVersions > 0 {
		// Document exists, read its current content
		currentContent, err := os.ReadFile(docPath)
		if err == nil && len(currentContent) > 0 {
			// Create timestamp for version filename
			timestamp := time.Now().Format("20060102150405")

			// Create versions directory path that mirrors the document path
			versionDir, err := metadataPath(cfg, "versions", relativePath)
			if err != nil {
				return err
			}

			// Ensure versions directory exists
			if err := os.MkdirAll(versionDir, 0755); err == nil {
				// Create version file path with timestamp
				versionPath, err := versionFilePath(cfg, relativePath, timestamp)
				if err != nil {
					return err
				}

				// Save the current content as a version
				_ = os.WriteFile(versionPath, currentContent, 0644)

				// Log the versioning
				logger.Info("Created version: %s", versionPath)

				// Clean up old versions if needed
				utils.CleanupOldVersions(versionDir, cfg.Wiki.MaxVersions)
			}
		}
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(docPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Write the content to the file
	if err := os.WriteFile(docPath, content, 0644); err != nil {
		return fmt.Errorf("failed to save document: %v", err)
	}

	return nil
}
