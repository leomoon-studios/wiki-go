package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"wiki-go/internal/auth"
	"wiki-go/internal/config"
)

// MoveRequest represents the request to move or rename a document or category
type MoveRequest struct {
	SourcePath string `json:"sourcePath"` // Current path of the document or category
	TargetPath string `json:"targetPath"` // New path for the document or category
	NewSlug    string `json:"newSlug"`    // New slug/name for the document or category (if renaming)
}

// MoveResponse represents the response for a move/rename operation
type MoveResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	NewPath string `json:"newPath,omitempty"`
	OldPath string `json:"oldPath,omitempty"`
}

// MoveDocumentHandler handles requests to move or rename a document or category
func MoveDocumentHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// Set JSON content type header
	w.Header().Set("Content-Type", "application/json")

	// Only process POST requests
	if r.Method != http.MethodPost {
		sendJSONResponse(w, false, "Method not allowed", http.StatusMethodNotAllowed, "", "")
		return
	}

	// Check authentication
	session := auth.GetSession(r)
	if session == nil {
		sendJSONResponse(w, false, "Authentication required", http.StatusUnauthorized, "", "")
		return
	}

	// Parse the request body
	var moveReq MoveRequest
	err := json.NewDecoder(r.Body).Decode(&moveReq)
	if err != nil {
		sendJSONResponse(w, false, "Invalid request format", http.StatusBadRequest, "", "")
		return
	}

	// Validate request
	if moveReq.SourcePath == "" {
		sendJSONResponse(w, false, "Source path is required", http.StatusBadRequest, "", "")
		return
	}

	// Clean and normalize paths
	moveReq.SourcePath = cleanPath(moveReq.SourcePath)
	moveReq.TargetPath = cleanPath(moveReq.TargetPath)

	// If target path is empty, set it to root
	if moveReq.TargetPath == "" && moveReq.NewSlug == "" {
		sendJSONResponse(w, false, "Either target path or new slug must be provided", http.StatusBadRequest, "", "")
		return
	}

	// Prevent moving the homepage
	if moveReq.SourcePath == "" || moveReq.SourcePath == "/" ||
		moveReq.SourcePath == "pages/home" || strings.EqualFold(moveReq.SourcePath, "pages/home") ||
		strings.HasSuffix(moveReq.SourcePath, "/homepage") {
		sendJSONResponse(w, false, "Cannot move or rename the home page", http.StatusBadRequest, "", "")
		return
	}

	// Also prevent setting the target path to the homepage
	if moveReq.TargetPath == "pages/home" || strings.EqualFold(moveReq.TargetPath, "pages/home") {
		sendJSONResponse(w, false, "Cannot move or rename to the home page location", http.StatusBadRequest, "", "")
		return
	}

	// Determine if this is a rename or move operation
	isRename := moveReq.NewSlug != ""
	isMove := moveReq.TargetPath != ""

	// Build the full source path
	documentDir := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir)
	fullSourcePath := filepath.Join(documentDir, moveReq.SourcePath)

	// Check if source exists
	_, err = os.Stat(fullSourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			sendJSONResponse(w, false, "Source document or category not found", http.StatusNotFound, "", "")
			return
		}
		sendJSONResponse(w, false, "Error accessing source: "+err.Error(), http.StatusInternalServerError, "", "")
		return
	}

	// Determine the target path based on operation type
	var fullTargetPath string
	var newPath string

	if isRename && !isMove {
		// Rename operation (change slug only)
		parentDir := filepath.Dir(moveReq.SourcePath)
		if parentDir == "." {
			parentDir = "" // Root directory
		}
		newPath = filepath.Join(parentDir, moveReq.NewSlug)
		fullTargetPath = filepath.Join(documentDir, newPath)
	} else if isMove && !isRename {
		// Move operation (change path only)
		sourceName := filepath.Base(moveReq.SourcePath)
		newPath = filepath.Join(moveReq.TargetPath, sourceName)
		fullTargetPath = filepath.Join(documentDir, newPath)
	} else if isMove && isRename {
		// Both move and rename
		newPath = filepath.Join(moveReq.TargetPath, moveReq.NewSlug)
		fullTargetPath = filepath.Join(documentDir, newPath)
	} else {
		// This case should not happen due to earlier validation
		sendJSONResponse(w, false, "Either new slug or target path must be provided", http.StatusBadRequest, "", "")
		return
	}

	// Check if target already exists
	if _, err := os.Stat(fullTargetPath); err == nil {
		sendJSONResponse(w, false, "Target already exists", http.StatusConflict, "", "")
		return
	}

	// Create target directory if it doesn't exist
	targetDir := filepath.Dir(fullTargetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		sendJSONResponse(w, false, "Failed to create target directory: "+err.Error(), http.StatusInternalServerError, "", "")
		return
	}

	// Move the document or category
	if err := os.Rename(fullSourcePath, fullTargetPath); err != nil {
		sendJSONResponse(w, false, "Failed to move: "+err.Error(), http.StatusInternalServerError, "", "")
		return
	}

	// Handle versions directory
	var versionsSourcePath, versionsTargetPath string

	if moveReq.SourcePath == "pages/home" {
		// For homepage, use the new paths
		versionsSourcePath = filepath.Join(cfg.Wiki.RootDir, "versions", "pages", "home")
	} else if strings.HasPrefix(moveReq.SourcePath, "documents/") {
		// Source path already includes "documents/" prefix
		versionsSourcePath = filepath.Join(cfg.Wiki.RootDir, "versions", moveReq.SourcePath)
	} else {
		// Add "documents/" prefix for regular documents
		versionsSourcePath = filepath.Join(cfg.Wiki.RootDir, "versions", "documents", moveReq.SourcePath)
	}

	if newPath == "pages/home" {
		// For homepage, use the new paths
		versionsTargetPath = filepath.Join(cfg.Wiki.RootDir, "versions", "pages", "home")
	} else if strings.HasPrefix(newPath, "documents/") {
		// Target path already includes "documents/" prefix
		versionsTargetPath = filepath.Join(cfg.Wiki.RootDir, "versions", newPath)
	} else {
		// Add "documents/" prefix for regular documents
		versionsTargetPath = filepath.Join(cfg.Wiki.RootDir, "versions", "documents", newPath)
	}

	// Check if versions directory exists
	if _, err := os.Stat(versionsSourcePath); err == nil {
		// Create parent directory for versions if needed
		if err := os.MkdirAll(filepath.Dir(versionsTargetPath), 0755); err != nil {
			log.Printf("Warning: Failed to create versions target directory: %v", err)
		} else {
			// Move versions directory
			if err := os.Rename(versionsSourcePath, versionsTargetPath); err != nil {
				log.Printf("Warning: Failed to move versions directory: %v", err)
			}
		}
	}

	// Return success response with both old and new paths
	sendJSONResponse(w, true, "Document moved successfully", http.StatusOK, newPath, moveReq.SourcePath)
}

// Helper function to clean and normalize a path
func cleanPath(path string) string {
	if path == "" {
		return ""
	}

	// Clean and normalize the path
	path = filepath.Clean(path)
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	path = strings.ReplaceAll(path, "\\", "/")

	return path
}

// Helper function to send a JSON response
func sendJSONResponse(w http.ResponseWriter, success bool, message string, statusCode int, newPath string, oldPath string) {
	w.WriteHeader(statusCode)
	response := MoveResponse{
		Success: success,
		Message: message,
	}

	if newPath != "" {
		response.NewPath = newPath
	}

	if oldPath != "" {
		response.OldPath = oldPath
	}

	json.NewEncoder(w).Encode(response)
}
