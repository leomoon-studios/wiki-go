package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"wiki-go/internal/auth"
	"wiki-go/internal/i18n"
	"wiki-go/internal/logger"
	"wiki-go/internal/roles"
	"wiki-go/internal/utils"
)

// SourceHandler handles requests to get the raw markdown content of a page
func SourceHandler(w http.ResponseWriter, r *http.Request) {
	// Add cache control headers to prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Method not allowed",
		})
		return
	}

	// Check if user is authenticated and has appropriate permissions
	session := auth.GetSession(r)
	if session == nil || (session.Role != roles.RoleAdmin && session.Role != roles.RoleEditor) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Unauthorized. Admin or editor access required.",
		})
		return
	}

	// Get the path from the URL, removing the /api/source prefix
	path, err := logicalDocumentPath(strings.TrimPrefix(r.URL.Path, "/api/source"))
	if err != nil {
		sendJSONError(w, "Invalid document path", http.StatusBadRequest, "")
		return
	}
	if !canAccessLogicalDocument(session, cfg, path) {
		sendJSONError(w, "Forbidden", http.StatusForbidden, "Document access denied")
		return
	}

	var docPath string
	var dirPath string

	// Special case for homepage (root path)
	if path == "" || path == "/" {
		// For the homepage, we use the pages directory
		docPath = filepath.Join(cfg.Wiki.RootDir, "pages", "home", "document.md")
		dirPath = filepath.Join(cfg.Wiki.RootDir, "pages", "home")
	} else {
		dirPath, err = utils.ResolveDocumentPath(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, path)
		if err != nil {
			sendJSONError(w, "Invalid document path", http.StatusBadRequest, "")
			return
		}
		docPath = filepath.Join(dirPath, "document.md")
	}

	// Read the markdown file
	content, err := os.ReadFile(docPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Check if the directory exists
			dirInfo, dirErr := os.Stat(dirPath)
			if dirErr == nil && dirInfo.IsDir() {
				// Directory exists but document.md doesn't
				// Create a default content with the directory name as title
				dirName := filepath.Base(path)
				if dirName == "" || dirName == "." {
					dirName = "Home"
				}

				// Format the directory name for display (replace dashes with spaces, capitalize)
				formattedName := utils.FormatDirName(dirName)

				// Create default content
				defaultContent := fmt.Sprintf("# %s\n\nEnter content here", formattedName)

				// Set content type and write response
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.Write([]byte(defaultContent))
				return
			}

			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Document not found",
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to read document",
		})
		return
	}

	// Reset content type for plain text response
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(content)
}

// SaveHandler handles requests to save the markdown content of a page
func SaveHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Method not allowed",
		})
		return
	}

	// Check if user is authenticated and has appropriate permissions
	session := auth.GetSession(r)
	if session == nil || (session.Role != roles.RoleAdmin && session.Role != roles.RoleEditor) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Unauthorized. Admin or editor access required.",
		})
		return
	}

	// Get the path from the URL, removing the /api/save prefix
	path, err := logicalDocumentPath(strings.TrimPrefix(r.URL.Path, "/api/save"))
	if err != nil {
		sendJSONError(w, "Invalid document path", http.StatusBadRequest, "")
		return
	}
	if !canAccessLogicalDocument(session, cfg, path) {
		sendJSONError(w, "Forbidden", http.StatusForbidden, "Document access denied")
		return
	}

	var docPath string
	var relativePath string // To store path relative to the documents dir

	// Special case for homepage (root path)
	if path == "" || path == "/" {
		// For the homepage, we use the pages directory
		docPath = filepath.Join(cfg.Wiki.RootDir, "pages", "home", "document.md")
		relativePath = "pages/home"
	} else {
		// Save relative path for versioning
		relativePath = "documents/" + strings.TrimPrefix(path, "/")

		docDir, resolveErr := utils.ResolveDocumentPath(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, path)
		if resolveErr != nil {
			sendJSONError(w, "Invalid document path", http.StatusBadRequest, "")
			return
		}
		docPath = filepath.Join(docDir, "document.md")
	}

	// Read the request body (new content)
	content, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to read request body",
		})
		return
	}
	defer r.Body.Close()

	// VERSION CONTROL: Save current version before overwriting
	// Check if the document already exists
	if _, err := os.Stat(docPath); err == nil && cfg.Wiki.MaxVersions > 0 {
		// Document exists, read its current content
		currentContent, err := os.ReadFile(docPath)
		if err == nil && len(currentContent) > 0 {
			// Create timestamp for version filename
			timestamp := time.Now().Format("20060102150405") // Format: yyyymmddhhmmss

			// Create versions directory path that mirrors the document path
			versionDir := filepath.Join(cfg.Wiki.RootDir, "versions", relativePath)

			// Ensure versions directory exists
			if err := os.MkdirAll(versionDir, 0755); err == nil {
				// Create version file path with timestamp
				versionPath := filepath.Join(versionDir, timestamp+".md")

				// Save the current content as a version
				_ = os.WriteFile(versionPath, currentContent, 0644) // Ignore error for now

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
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to create directory",
		})
		return
	}

	// Write the content to the file
	if err := os.WriteFile(docPath, content, 0644); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to save document",
		})
		return
	}

	logger.Info("User %s edited document %s", session.Username, filepath.Join(cfg.Wiki.RootDir, relativePath))

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Document saved successfully",
	})
}

// CreateDocumentRequest represents the JSON payload for creating a new document
type CreateDocumentRequest struct {
	Title string `json:"title"`
	Path  string `json:"path"`
	Type  string `json:"type"`
}

// CreateDocumentResponse represents the JSON response after creating a document
type CreateDocumentResponse struct {
	URL     string `json:"url"`
	Message string `json:"message"`
}

// ErrorResponse represents a JSON error response
type ErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// CreateDocumentHandler handles the API endpoint for creating new documents
func CreateDocumentHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		sendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed, "")
		return
	}

	// Check authentication and permissions
	session := auth.GetSession(r)
	if session == nil || (session.Role != roles.RoleAdmin && session.Role != roles.RoleEditor) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Unauthorized. Admin or editor access required.",
		})
		return
	}

	// Parse the request body
	var req CreateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request payload", http.StatusBadRequest, err.Error())
		return
	}

	// Validate the request
	if req.Title == "" {
		sendJSONError(w, "Title is required", http.StatusBadRequest, "")
		return
	}

	if req.Path == "" {
		sendJSONError(w, "Path is required", http.StatusBadRequest, "")
		return
	}

	// Clean the path (remove any unwanted characters)
	cleanPath := utils.SanitizePath(req.Path)
	if cleanPath == "" {
		sendJSONError(w, "Invalid path after sanitization", http.StatusBadRequest, "")
		return
	}
	logicalPath, err := logicalDocumentPath(cleanPath)
	if err != nil {
		sendJSONError(w, "Invalid path after sanitization", http.StatusBadRequest, "")
		return
	}
	if !canAccessLogicalDocument(session, cfg, logicalPath) {
		sendJSONError(w, "Forbidden", http.StatusForbidden, "Document access denied")
		return
	}

	logger.Debug("Creating document: Title=%s, Path=%s, CleanPath=%s", req.Title, req.Path, cleanPath)

	// Get the config from the package variable
	// cfg is defined in handlers.go and initialized by InitHandlers

	// Log the paths for debugging
	logger.Debug("Creating document: Title=%s, Path=%s, CleanPath=%s", req.Title, req.Path, cleanPath)

	// Build the file path
	fullPath, err := utils.ResolveDocumentPath(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, logicalPath)
	if err != nil {
		sendJSONError(w, "Invalid document path", http.StatusBadRequest, "")
		return
	}

	// Log the full path
	logger.Debug("Full path: %s", fullPath)

	// Create the directory if it doesn't exist
	err = os.MkdirAll(fullPath, 0755)
	if err != nil {
		logger.Error("Error creating directories: %v", err)
		sendJSONError(w, "Failed to create directories", http.StatusInternalServerError, err.Error())
		return
	}

	// Create the document.md file inside the directory
	docFile := filepath.Join(fullPath, "document.md")

	// Check if file already exists
	if _, err := os.Stat(docFile); err == nil {
		sendJSONError(w, "Document already exists", http.StatusConflict, "")
		return
	}

	// Create the file content with the title as H1
	var content string
	if req.Type == "kanban" {
		content = fmt.Sprintf(
			"---\nlayout: kanban\n---\n\n# %s\n\n%s\n\n#### %s\n\n##### %s\n- [ ] %s\n\n##### %s\n\n##### %s",
			req.Title,
			i18n.Translate("new_doc.default_content"),
			i18n.Translate("new_doc.kanban_title"),
			i18n.Translate("new_doc.kanban_todo"),
			i18n.Translate("new_doc.kanban_task_example"),
			i18n.Translate("new_doc.kanban_in_progress"),
			i18n.Translate("new_doc.kanban_done"),
		)
	} else if req.Type == "links" {
		content = fmt.Sprintf("---\nlayout: links\n---\n\n# %s\n\n## Web Tools\n- [Example Link](https://example.com) - Sample link description | %s\n\n## Documentation\n- [MDN Docs](https://developer.mozilla.org) - Web development reference | %s", req.Title, time.Now().Format("2006-01-02"), time.Now().Format("2006-01-02"))
	} else {
		// Default to markdown
		content = fmt.Sprintf("# %s\n\n%s", req.Title, i18n.Translate("new_doc.default_content"))
	}

	// Write to the file
	err = os.WriteFile(docFile, []byte(content), 0644)
	if err != nil {
		logger.Error("Error creating document: %v", err)
		sendJSONError(w, "Failed to create document", http.StatusInternalServerError, err.Error())
		return
	}

	logger.Info("User %s created document %s", session.Username, filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, cleanPath))

	// Return success
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"success": true,
		"url":     "/" + cleanPath,
		"message": "Document created successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// sendJSONError sends a JSON error response with status code
func sendJSONError(w http.ResponseWriter, message string, statusCode int, errorDetails string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"success": false,
		"message": message,
	}

	if errorDetails != "" {
		response["error"] = errorDetails
	}

	json.NewEncoder(w).Encode(response)
	logger.Error("Error response: %s (%d) - %s", message, statusCode, errorDetails)
}

// DocumentHandler is a combined handler for document operations (GET, DELETE)
func DocumentHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		DeleteDocumentHandler(w, r)
	case http.MethodGet:
		// For now just return the document path
		docPath := strings.TrimPrefix(r.URL.Path, "/api/document")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"path":    docPath,
			"message": "Document retrieval not yet implemented",
		})
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Method not allowed",
		})
	}
}

// DeleteDocumentHandler handles requests to delete a document
func DeleteDocumentHandler(w http.ResponseWriter, r *http.Request) {
	// Only process DELETE requests
	if r.Method != http.MethodDelete {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Method not allowed",
		})
		return
	}

	// Check authentication and permissions
	session := auth.GetSession(r)
	if session == nil || (session.Role != roles.RoleAdmin && session.Role != roles.RoleEditor) {
		sendJSONError(w, "Authentication required", http.StatusUnauthorized, "Admin or editor access required to delete documents")
		return
	}

	// Get the path from the URL
	urlPath := r.URL.Path
	if urlPath == "/" {
		sendJSONError(w, "Cannot delete homepage", http.StatusBadRequest, "The homepage cannot be deleted")
		return
	}

	// Remove "/api/document" prefix from the path
	docPath, err := logicalDocumentPath(strings.TrimPrefix(urlPath, "/api/document"))
	if err != nil || docPath == "/" {
		sendJSONError(w, "Invalid document path", http.StatusBadRequest, "Document path is required")
		return
	}
	if !canAccessLogicalDocument(session, cfg, docPath) {
		sendJSONError(w, "Forbidden", http.StatusForbidden, "Document access denied")
		return
	}

	// Build the file path
	fullPath, err := utils.ResolveDocumentPath(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, docPath)
	if err != nil {
		sendJSONError(w, "Invalid document path", http.StatusBadRequest, "")
		return
	}
	relativeDocPath := strings.TrimPrefix(docPath, "/")

	// Check if file exists
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Document not found",
			})
			return
		}
		sendJSONError(w, "Error accessing document", http.StatusInternalServerError, err.Error())
		return
	}

	// Delete the file or directory recursively
	if fileInfo.IsDir() {
		// Use RemoveAll to recursively delete the directory and all its contents
		if err := os.RemoveAll(fullPath); err != nil {
			sendJSONError(w, "Error deleting directory", http.StatusInternalServerError, err.Error())
			return
		}
		logger.Info("User %s deleted document %s", session.Username, fullPath)
	} else {
		// Delete the file
		if err := os.Remove(fullPath); err != nil {
			sendJSONError(w, "Error deleting document", http.StatusInternalServerError, err.Error())
			return
		}
		logger.Info("User %s deleted document %s", session.Username, fullPath)
	}

	// Also delete the corresponding versions directory
	var versionsPath string
	if relativeDocPath == "pages/home" {
		// For homepage, use the new path
		versionsPath = filepath.Join(cfg.Wiki.RootDir, "versions", "pages", "home")
	} else if strings.HasPrefix(relativeDocPath, "documents/") {
		// Path already includes "documents/" prefix
		versionsPath = filepath.Join(cfg.Wiki.RootDir, "versions", relativeDocPath)
	} else {
		// Add "documents/" prefix for regular documents
		versionsPath = filepath.Join(cfg.Wiki.RootDir, "versions", "documents", relativeDocPath)
		if strings.HasSuffix(fullPath, ".md") {
			// If we're deleting a .md file, remove the .md extension from the versions path
			versionsPath = filepath.Join(cfg.Wiki.RootDir, "versions", "documents", strings.TrimSuffix(relativeDocPath, ".md"))
		}
	}

	// Check if versions directory exists before attempting to delete
	if _, err := os.Stat(versionsPath); err == nil {
		if err := os.RemoveAll(versionsPath); err != nil {
			logger.Warn("Failed to delete versions directory: %s - %v", versionsPath, err)
			// Continue execution even if versions deletion fails
		} else {
			logger.Info("Deleted versions directory: %s", versionsPath)
		}
	}

	// Also delete the corresponding comments directory
	commentsPath := filepath.Join(cfg.Wiki.RootDir, "comments", relativeDocPath)

	// Check if comments directory exists before attempting to delete
	if _, err := os.Stat(commentsPath); err == nil {
		if err := os.RemoveAll(commentsPath); err != nil {
			logger.Warn("Failed to delete comments directory: %s - %v", commentsPath, err)
			// Continue execution even if comments deletion fails
		} else {
			logger.Info("Deleted comments directory: %s", commentsPath)
		}
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Document deleted successfully",
	})
}
