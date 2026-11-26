package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
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
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		sendJSONResponse(w, false, "Method not allowed", http.StatusMethodNotAllowed, "", "")
		return
	}
	session := auth.GetSession(r)
	if session == nil || (session.Role != config.RoleAdmin && session.Role != config.RoleEditor) {
		sendJSONResponse(w, false, "Editor access required", http.StatusUnauthorized, "", "")
		return
	}
	var request MoveRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.SourcePath == "" || request.TargetPath == "" && request.NewSlug == "" {
		sendJSONResponse(w, false, "Invalid move request", http.StatusBadRequest, "", "")
		return
	}
	source, ok := requireAPIDocument(w, cfg, request.SourcePath, session, false, true)
	if !ok {
		return
	}
	if source.LogicalPath == "/" {
		sendJSONResponse(w, false, "Cannot move homepage", http.StatusBadRequest, "", "")
		return
	}
	if request.NewSlug != "" {
		canonical, err := utils.CanonicalDocumentRequestPath(request.NewSlug)
		if err != nil || canonical != "/"+request.NewSlug || strings.ContainsAny(request.NewSlug, "/\\") || request.NewSlug == "." || request.NewSlug == ".." {
			sendJSONResponse(w, false, "Invalid new slug", http.StatusBadRequest, "", "")
			return
		}
	}
	parent := path.Dir(source.LogicalPath)
	if request.TargetPath != "" {
		// A move targets the documents storage root, not the logical homepage.
		canonical, err := utils.CanonicalDocumentRequestPath(request.TargetPath)
		if err != nil {
			sendJSONResponse(w, false, "Invalid target path", http.StatusBadRequest, "", "")
			return
		}
		if canonical == "/" {
			parent = "/"
		} else {
			targetParent, ok := requireAPIDocument(w, cfg, request.TargetPath, session, true, true)
			if !ok {
				return
			}
			if targetParent.LogicalPath == "/" {
				sendJSONResponse(w, false, "Cannot move into homepage", http.StatusBadRequest, "", "")
				return
			}
			parent = targetParent.LogicalPath
		}
	}
	name := path.Base(source.LogicalPath)
	if request.NewSlug != "" {
		name = request.NewSlug
	}
	newLogical := path.Join(parent, name)
	target, ok := requireAPIDocument(w, cfg, newLogical, session, true, false)
	if !ok {
		return
	}
	fullTarget := target.FilesystemPath
	// Windows resolves an existing case-only destination to the source spelling.
	// Use the validated new component when renaming case in the same directory.
	if runtime.GOOS == "windows" && fullTarget == source.FilesystemPath && request.NewSlug != "" && name != filepath.Base(source.FilesystemPath) && strings.EqualFold(name, filepath.Base(source.FilesystemPath)) {
		fullTarget = filepath.Join(filepath.Dir(source.FilesystemPath), name)
		newLogical = path.Join(path.Dir(source.LogicalPath), name)
	} else {
		newLogical = target.LogicalPath
	}
	if fullTarget == source.FilesystemPath {
		sendJSONResponse(w, false, "Source and target are the same", http.StatusBadRequest, "", "")
		return
	}
	caseRename := runtime.GOOS == "windows" && strings.EqualFold(source.FilesystemPath, fullTarget)
	relative, err := filepath.Rel(source.FilesystemPath, fullTarget)
	if err != nil || !caseRename && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative) {
		sendJSONResponse(w, false, "Cannot move into a descendant", http.StatusBadRequest, "", "")
		return
	}
	if err := authorizeDocumentTree(cfg, session, source.ResolvedDocumentPath, newLogical); err != nil {
		sendJSONResponse(w, false, "Document subtree access denied", http.StatusForbidden, "", "")
		return
	}
	if !caseRename {
		if entries, err := os.ReadDir(fullTarget); err == nil && len(entries) > 0 {
			sendJSONResponse(w, false, "Target directory is not empty", http.StatusConflict, "", "")
			return
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			sendJSONResponse(w, false, "Invalid target directory", http.StatusBadRequest, "", "")
			return
		}
	}
	type storageMove struct{ source, target string }
	var metadata []storageMove
	for _, namespace := range []string{"versions", "comments"} {
		from, err := metadataPath(cfg, namespace, source.LogicalPath)
		if err != nil {
			sendJSONResponse(w, false, "Invalid metadata source", http.StatusBadRequest, "", "")
			return
		}
		to, err := metadataPath(cfg, namespace, newLogical)
		if err != nil {
			sendJSONResponse(w, false, "Invalid metadata destination", http.StatusBadRequest, "", "")
			return
		}
		if from != to {
			if _, err := os.Stat(to); err == nil {
				sendJSONResponse(w, false, "Target metadata already exists", http.StatusConflict, "", "")
				return
			} else if !errors.Is(err, os.ErrNotExist) {
				sendJSONResponse(w, false, "Invalid target metadata", http.StatusBadRequest, "", "")
				return
			}
			metadata = append(metadata, storageMove{from, to})
		}
	}
	if err := os.MkdirAll(filepath.Dir(fullTarget), 0755); err != nil {
		sendJSONResponse(w, false, "Failed to create target parent", http.StatusInternalServerError, "", "")
		return
	}
	if err := os.Rename(source.FilesystemPath, fullTarget); err != nil {
		sendJSONResponse(w, false, "Failed to move document", http.StatusInternalServerError, "", "")
		return
	}
	for _, move := range metadata {
		if _, err := os.Stat(move.source); err == nil {
			if err := os.MkdirAll(filepath.Dir(move.target), 0755); err != nil {
				logger.Warn("Failed to create metadata parent: %v", err)
				continue
			}
			if err := os.Rename(move.source, move.target); err != nil {
				logger.Warn("Failed to move metadata: %v", err)
			}
		}
	}
	sendJSONResponse(w, true, "Document moved successfully", http.StatusOK, strings.TrimPrefix(newLogical, "/"), strings.TrimPrefix(source.LogicalPath, "/"))
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
