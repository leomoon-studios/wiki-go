package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"wiki-go/internal/auth"
	"wiki-go/internal/utils"
)

// SourceHandler handles requests to get the raw markdown content of a page
func SourceHandler(w http.ResponseWriter, r *http.Request) {
	// Add cache control headers to prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	session := auth.GetSession(r)
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the path from the URL, removing the /api/source prefix
	path := strings.TrimPrefix(r.URL.Path, "/api/source")

	var docPath string

	// Special case for homepage (root path)
	if path == "" || path == "/" {
		// For the homepage, we use the pages directory
		docPath = filepath.Join(cfg.Wiki.RootDir, "pages", "home", "document.md")
	} else {
		// Clean and normalize the path
		path = filepath.Clean(path)
		path = strings.TrimSuffix(path, "/")
		path = strings.ReplaceAll(path, "\\", "/")

		// Get the full filesystem path, adding the documents subdirectory
		docPath = filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, path, "document.md")
	}

	// Read the markdown file
	content, err := os.ReadFile(docPath)
	if err != nil {
		if os.IsNotExist(err) {
			// If document doesn't exist, create it with a default title
			dirName := filepath.Base(filepath.Dir(docPath))
			title := utils.FormatDirName(dirName)
			content = []byte(fmt.Sprintf("# %s\n", title))

			// Create directory if it doesn't exist
			dir := filepath.Dir(docPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				http.Error(w, "Failed to create directory", http.StatusInternalServerError)
				return
			}

			// Write the initial content
			if err := os.WriteFile(docPath, content, 0644); err != nil {
				http.Error(w, "Failed to create document", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Failed to read document", http.StatusInternalServerError)
			return
		}
	}

	// Set content type and write response
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(content)
}

// SaveHandler handles requests to save the markdown content of a page
func SaveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	session := auth.GetSession(r)
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the path from the URL, removing the /api/save prefix
	path := strings.TrimPrefix(r.URL.Path, "/api/save")

	var docPath string
	var relativePath string // To store path relative to the documents dir

	// Special case for homepage (root path)
	if path == "" || path == "/" {
		// For the homepage, we use the pages directory
		docPath = filepath.Join(cfg.Wiki.RootDir, "pages", "home", "document.md")
		relativePath = "pages/home"
	} else {
		// Clean and normalize the path
		path = filepath.Clean(path)
		path = strings.TrimSuffix(path, "/")
		path = strings.ReplaceAll(path, "\\", "/")

		// Save relative path for versioning
		relativePath = "documents/" + path

		// Get the full filesystem path, adding the documents subdirectory
		docPath = filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, path, "document.md")
	}

	// Read the request body (new content)
	content, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
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
				log.Printf("Created version: %s", versionPath)

				// Clean up old versions if needed
				utils.CleanupOldVersions(versionDir, cfg.Wiki.MaxVersions)
			}
		}
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(docPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		http.Error(w, "Failed to create directory", http.StatusInternalServerError)
		return
	}

	// Write the content to the file
	if err := os.WriteFile(docPath, content, 0644); err != nil {
		http.Error(w, "Failed to save document", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// CreateDocumentRequest represents the JSON payload for creating a new document
type CreateDocumentRequest struct {
	Title string `json:"title"`
	Path  string `json:"path"`
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

	// Check authentication
	session := auth.GetSession(r)
	if session == nil {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized, "")
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

	log.Printf("Creating document: Title=%s, Path=%s, CleanPath=%s", req.Title, req.Path, cleanPath)

	// Get the config from the package variable
	// cfg is defined in handlers.go and initialized by InitHandlers

	// Log the paths for debugging
	log.Printf("Creating document: Title=%s, Path=%s, CleanPath=%s", req.Title, req.Path, cleanPath)

	// Build the file path
	documentDir := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir)
	fullPath := filepath.Join(documentDir, cleanPath)

	// Log the full path
	log.Printf("Full path: %s", fullPath)

	// Create the directory if it doesn't exist
	err := os.MkdirAll(fullPath, 0755)
	if err != nil {
		log.Printf("Error creating directories: %v", err)
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
	content := fmt.Sprintf("# %s\n\nEnter content here", req.Title)

	// Write to the file
	err = os.WriteFile(docFile, []byte(content), 0644)
	if err != nil {
		log.Printf("Error creating document: %v", err)
		sendJSONError(w, "Failed to create document", http.StatusInternalServerError, err.Error())
		return
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	response := CreateDocumentResponse{
		URL:     "/" + cleanPath,
		Message: "Document created successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// sendJSONError sends a JSON error response with status code
func sendJSONError(w http.ResponseWriter, message string, statusCode int, errorDetails string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Message: message,
	}

	if errorDetails != "" {
		response.Error = errorDetails
	}

	json.NewEncoder(w).Encode(response)
	log.Printf("Error response: %s (%d) - %s", message, statusCode, errorDetails)
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
		json.NewEncoder(w).Encode(map[string]string{
			"path":    docPath,
			"message": "Document retrieval not yet implemented",
		})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// DeleteDocumentHandler handles requests to delete a document
func DeleteDocumentHandler(w http.ResponseWriter, r *http.Request) {
	// Only process DELETE requests
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check authentication
	session := auth.GetSession(r)
	if session == nil {
		sendJSONError(w, "Authentication required", http.StatusUnauthorized, "You must be logged in to delete documents")
		return
	}

	// Get the path from the URL
	urlPath := r.URL.Path
	if urlPath == "/" {
		sendJSONError(w, "Cannot delete homepage", http.StatusBadRequest, "The homepage cannot be deleted")
		return
	}

	// Remove "/api/document" prefix from the path
	docPath := strings.TrimPrefix(urlPath, "/api/document")
	if docPath == "" {
		sendJSONError(w, "Invalid document path", http.StatusBadRequest, "Document path is required")
		return
	}

	// Build the file path
	docPath = filepath.Clean(docPath)
	documentDir := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir)
	fullPath := filepath.Join(documentDir, docPath)

	// Check if file exists
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Try with .md extension
			if !strings.HasSuffix(fullPath, ".md") {
				fullPath += ".md"
				fileInfo, err = os.Stat(fullPath)
				if err != nil {
					sendJSONError(w, "Document not found", http.StatusNotFound, "The specified document does not exist")
					return
				}
			} else {
				sendJSONError(w, "Document not found", http.StatusNotFound, "The specified document does not exist")
				return
			}
		} else {
			sendJSONError(w, "Error accessing document", http.StatusInternalServerError, err.Error())
			return
		}
	}

	// Delete the file or directory recursively
	if fileInfo.IsDir() {
		// Use RemoveAll to recursively delete the directory and all its contents
		if err := os.RemoveAll(fullPath); err != nil {
			sendJSONError(w, "Error deleting directory", http.StatusInternalServerError, err.Error())
			return
		}
		log.Printf("Recursively deleted directory: %s", fullPath)
	} else {
		// Delete the file
		if err := os.Remove(fullPath); err != nil {
			sendJSONError(w, "Error deleting document", http.StatusInternalServerError, err.Error())
			return
		}
		log.Printf("Deleted file: %s", fullPath)
	}

	// Also delete the corresponding versions directory
	var versionsPath string
	if docPath == "pages/home" {
		// For homepage, use the new path
		versionsPath = filepath.Join(cfg.Wiki.RootDir, "versions", "pages", "home")
	} else if strings.HasPrefix(docPath, "documents/") {
		// Path already includes "documents/" prefix
		versionsPath = filepath.Join(cfg.Wiki.RootDir, "versions", docPath)
	} else {
		// Add "documents/" prefix for regular documents
		versionsPath = filepath.Join(cfg.Wiki.RootDir, "versions", "documents", docPath)
		if strings.HasSuffix(fullPath, ".md") {
			// If we're deleting a .md file, remove the .md extension from the versions path
			versionsPath = filepath.Join(cfg.Wiki.RootDir, "versions", "documents", strings.TrimSuffix(docPath, ".md"))
		}
	}

	// Check if versions directory exists before attempting to delete
	if _, err := os.Stat(versionsPath); err == nil {
		if err := os.RemoveAll(versionsPath); err != nil {
			log.Printf("Warning: Failed to delete versions directory: %s - %v", versionsPath, err)
			// Continue execution even if versions deletion fails
		} else {
			log.Printf("Deleted versions directory: %s", versionsPath)
		}
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Document deleted successfully",
	})
}
