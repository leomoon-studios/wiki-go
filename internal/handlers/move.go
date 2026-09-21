package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"wiki-go/internal/auth"
	"wiki-go/internal/config"
	"wiki-go/internal/logger"
	"wiki-go/internal/utils"
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
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Unauthorized. Admin or editor access required.",
		})
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

	// Resolve paths before authorization or filesystem access.
	source, err := utils.ResolveRelativeDocumentPath(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, moveReq.SourcePath)
	if err != nil {
		sendJSONResponse(w, false, "Invalid source path", http.StatusBadRequest, "", "")
		return
	}
	if !canAccessLogicalDocument(session, cfg, source.LogicalPath) {
		sendJSONResponse(w, false, "Document access denied", http.StatusForbidden, "", "")
		return
	}
	moveReq.SourcePath = strings.TrimPrefix(source.LogicalPath, "/")
	if moveReq.TargetPath != "" {
		target, targetErr := utils.ResolveRelativeDocumentPath(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, moveReq.TargetPath)
		if targetErr != nil {
			sendJSONResponse(w, false, "Invalid target path", http.StatusBadRequest, "", "")
			return
		}
		moveReq.TargetPath = strings.TrimPrefix(target.LogicalPath, "/")
	}

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

	// Check if this is a move to root operation
	moveToRoot := false
	sourceBase := path.Base(moveReq.SourcePath)
	sourceDir := path.Dir(moveReq.SourcePath)

	// If we're moving to the root (empty target path) and the source is not already at the root
	if moveReq.TargetPath == "" && sourceDir != "." {
		moveToRoot = true
		logger.Debug("Detected move to root operation: %s -> %s", moveReq.SourcePath, moveReq.NewSlug)
	}

	// Determine if this is a move operation
	isMove := moveReq.TargetPath != "" || moveToRoot

	logger.Debug("Operation analysis: sourceBase=%s, sourceDir=%s, moveToRoot=%v",
		sourceBase, sourceDir, moveToRoot)

	fullSourcePath := source.FilesystemPath

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

	// Log the request details for debugging
	logger.Debug("Move request: SourcePath=%s, TargetPath=%s, NewSlug=%s", moveReq.SourcePath, moveReq.TargetPath, moveReq.NewSlug)
	logger.Debug("Operation type: isRename=%v, isMove=%v", isRename, isMove)

	if isRename && !isMove {
		// Rename operation (change slug only)
		parentDir := path.Dir(moveReq.SourcePath)
		if parentDir == "." {
			parentDir = "" // Root directory
		}
		newPath = path.Join(parentDir, moveReq.NewSlug)
	} else if isMove && !isRename {
		// Move operation (change path only)
		sourceName := path.Base(moveReq.SourcePath)

		// Special case for moving to root
		if moveReq.TargetPath == "" {
			newPath = sourceName
		} else {
			newPath = path.Join(moveReq.TargetPath, sourceName)
		}
	} else if isMove && isRename {
		// Both move and rename

		// Special case for moving to root
		if moveReq.TargetPath == "" {
			newPath = moveReq.NewSlug
		} else {
			newPath = path.Join(moveReq.TargetPath, moveReq.NewSlug)
		}
	} else {
		// This case should not happen due to earlier validation
		sendJSONResponse(w, false, "Either new slug or target path must be provided", http.StatusBadRequest, "", "")
		return
	}

	target, err := utils.ResolveRelativeDocumentPath(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, newPath)
	if err != nil {
		sendJSONResponse(w, false, "Invalid target path", http.StatusBadRequest, "", "")
		return
	}
	targetParentPath := path.Dir(target.LogicalPath)
	if !canAccessLogicalDocument(session, cfg, targetParentPath) || !canAccessLogicalDocument(session, cfg, target.LogicalPath) {
		sendJSONResponse(w, false, "Document access denied", http.StatusForbidden, "", "")
		return
	}
	newPath = strings.TrimPrefix(target.LogicalPath, "/")
	fullTargetPath = target.FilesystemPath

	// Log the calculated paths
	logger.Debug("Calculated paths: newPath=%s, fullTargetPath=%s", newPath, fullTargetPath)

	// Check if target already exists, but only if it's not the same as the source
	// We need to check if the document.md file exists at the target path
	targetDocPath := filepath.Join(fullTargetPath, "document.md")

	// Skip the check if the target is the same as the source (just in a different location)
	// This allows moving a document to the root with the same name
	if fullTargetPath != fullSourcePath {
		if _, err := os.Stat(targetDocPath); err == nil {
			// Check if this is a case-only rename (e.g., "test" to "Test")
			sourceBaseLower := strings.ToLower(path.Base(moveReq.SourcePath))
			targetBaseLower := strings.ToLower(moveReq.NewSlug)

			// If it's not a case-only rename, then it's a conflict
			if sourceBaseLower != targetBaseLower || filepath.Dir(fullSourcePath) == filepath.Dir(fullTargetPath) {
				sendJSONResponse(w, false, "A document already exists at the target location", http.StatusConflict, "", "")
				return
			}
		}

		// Also check if the directory itself exists and is not empty
		if info, err := os.Stat(fullTargetPath); err == nil && info.IsDir() {
			// Check if the directory is empty
			entries, err := os.ReadDir(fullTargetPath)
			if err == nil && len(entries) > 0 {
				sendJSONResponse(w, false, "Target directory already exists and is not empty", http.StatusConflict, "", "")
				return
			}
		}
	}

	// Create target directory if it doesn't exist
	targetDir := filepath.Dir(fullTargetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		sendJSONResponse(w, false, "Failed to create target directory: "+err.Error(), http.StatusInternalServerError, "", "")
		return
	}

	// Check if source and target are the same
	if fullSourcePath == fullTargetPath {
		logger.Warn("Source and target paths are the same, aborting move")
		sendJSONResponse(w, false, "Source and target paths are the same", http.StatusBadRequest, "", "")
		return
	}

	// Move the document or category
	if err := os.Rename(fullSourcePath, fullTargetPath); err != nil {
		logger.Error("Error moving document: %v", err)
		sendJSONResponse(w, false, "Failed to move: "+err.Error(), http.StatusInternalServerError, "", "")
		return
	}

	logger.Info("User %s moved %s -> %s", session.Username,
		filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, moveReq.SourcePath),
		filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, newPath))

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
			logger.Warn("Failed to create versions target directory: %v", err)
		} else {
			// Move versions directory
			if err := os.Rename(versionsSourcePath, versionsTargetPath); err != nil {
				logger.Warn("Failed to move versions directory: %v", err)
			}
		}
	}

	// Handle comments directory
	commentsSourcePath := filepath.Join(cfg.Wiki.RootDir, "comments", moveReq.SourcePath)
	commentsTargetPath := filepath.Join(cfg.Wiki.RootDir, "comments", newPath)

	// Check if comments directory exists
	if _, err := os.Stat(commentsSourcePath); err == nil {
		// Create parent directory for comments if needed
		if err := os.MkdirAll(filepath.Dir(commentsTargetPath), 0755); err != nil {
			logger.Warn("Failed to create comments target directory: %v", err)
		} else {
			// Move comments directory
			if err := os.Rename(commentsSourcePath, commentsTargetPath); err != nil {
				logger.Warn("Failed to move comments directory: %v", err)
			}
		}
	}

	// Return success response with both old and new paths
	sendJSONResponse(w, true, "Document moved successfully", http.StatusOK, newPath, moveReq.SourcePath)
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
