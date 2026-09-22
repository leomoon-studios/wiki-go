package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"wiki-go/internal/auth"
	"wiki-go/internal/comments"
	"wiki-go/internal/roles"
	"wiki-go/internal/utils"
)

// CommentRequest represents the request body for adding a comment
type CommentRequest struct {
	Content string `json:"content"`
}

// CommentResponse represents the response for a comment operation
type CommentResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type commentDocument struct {
	logicalPath string
	filePath    string
}

func resolveCommentDocument(rawPath string) (commentDocument, error) {
	if rawPath == "" {
		return commentDocument{}, fmt.Errorf("document path is empty")
	}
	logicalPath, err := logicalDocumentPath(rawPath)
	if err != nil {
		return commentDocument{}, err
	}
	if logicalPath == "/" {
		return commentDocument{
			logicalPath: logicalPath,
			filePath:    filepath.Join(cfg.Wiki.RootDir, "pages", "home", "document.md"),
		}, nil
	}
	documentDir, err := utils.ResolveDocumentPath(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, logicalPath)
	if err != nil {
		return commentDocument{}, err
	}
	return commentDocument{
		logicalPath: logicalPath,
		filePath:    filepath.Join(documentDir, "document.md"),
	}, nil
}

func commentsRoot() string {
	return filepath.Join(cfg.Wiki.RootDir, "comments")
}

func requireCommentDocumentAccess(w http.ResponseWriter, session *auth.Session, logicalPath string) bool {
	if auth.CanAccessDocument(logicalPath, session, cfg) {
		return true
	}
	if session == nil {
		sendJSONError(w, "Authentication required", http.StatusUnauthorized, "")
	} else {
		sendJSONError(w, "Document access denied", http.StatusForbidden, "")
	}
	return false
}

// AddCommentHandler handles requests to add a comment to a document
func AddCommentHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		sendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed, "")
		return
	}

	// FIRST CHECK: If comments are disabled system-wide, reject IMMEDIATELY
	if cfg.Wiki.DisableComments {
		sendJSONError(w, "Comments are disabled system-wide", http.StatusForbidden, "")
		return
	}

	// Check if user is authenticated
	session := auth.GetSession(r)
	if session == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Authentication required",
		})
		return
	}

	// Get the document path from the request
	rawDocPath := strings.TrimPrefix(r.URL.Path, "/api/comments/add/")
	if rawDocPath == "" {
		sendJSONError(w, "Document path is required", http.StatusBadRequest, "")
		return
	}
	document, err := resolveCommentDocument(rawDocPath)
	if err != nil {
		sendJSONError(w, "Invalid document path", http.StatusBadRequest, "")
		return
	}
	if !requireCommentDocumentAccess(w, session, document.logicalPath) {
		return
	}

	if _, err := os.Stat(document.filePath); os.IsNotExist(err) {
		sendJSONError(w, "Document not found", http.StatusNotFound, "")
		return
	}

	// Read document content to check if comments are allowed
	content, err := os.ReadFile(document.filePath)
	if err != nil {
		sendJSONError(w, "Failed to read document", http.StatusInternalServerError, err.Error())
		return
	}

	// Check if comments are allowed for this document
	if !comments.AreCommentsAllowed(string(content)) {
		sendJSONError(w, "Comments are not allowed for this document", http.StatusForbidden, "")
		return
	}

	// Parse the request body
	var req CommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest, err.Error())
		return
	}

	// Validate the comment content
	if strings.TrimSpace(req.Content) == "" {
		sendJSONError(w, "Comment content cannot be empty", http.StatusBadRequest, "")
		return
	}

	// Add the comment
	err = comments.AddComment(commentsRoot(), document.logicalPath, req.Content, session.Username)
	if err != nil {
		sendJSONError(w, "Failed to add comment", http.StatusInternalServerError, err.Error())
		return
	}

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CommentResponse{
		Success: true,
		Message: "Comment added successfully",
	})
}

// GetCommentsHandler handles requests to get comments for a document
func GetCommentsHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		sendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed, "")
		return
	}

	// FIRST CHECK: If comments are disabled system-wide, return empty list IMMEDIATELY
	if cfg.Wiki.DisableComments {
		// Return empty list if comments are disabled system-wide
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"comments": []comments.Comment{},
		})
		return
	}

	// Get the document path from the request
	rawDocPath := strings.TrimPrefix(r.URL.Path, "/api/comments/")
	if rawDocPath == "" {
		sendJSONError(w, "Document path is required", http.StatusBadRequest, "")
		return
	}
	document, err := resolveCommentDocument(rawDocPath)
	if err != nil {
		sendJSONError(w, "Invalid document path", http.StatusBadRequest, "")
		return
	}
	session := auth.GetSession(r)
	if !requireCommentDocumentAccess(w, session, document.logicalPath) {
		return
	}

	// Get comments for the document
	commentsList, err := comments.GetComments(commentsRoot(), document.logicalPath)
	if err != nil {
		sendJSONError(w, "Failed to get comments", http.StatusInternalServerError, err.Error())
		return
	}

	// Process comments for rendering
	for i := range commentsList {
		commentsList[i].RenderedHTML = utils.RenderCommentMarkdown(commentsList[i].Content)
		// Format timestamp
		commentsList[i].FormattedTime = comments.FormatCommentTime(commentsList[i].Timestamp)
	}

	// Send comments as JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"comments": commentsList,
	})
}

// DeleteCommentHandler handles requests to delete a comment
func DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow DELETE requests
	if r.Method != http.MethodDelete {
		sendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed, "")
		return
	}

	// Check if user is authenticated and has admin role
	session := auth.GetSession(r)
	if session == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Authentication required",
		})
		return
	}

	if session.Role != roles.RoleAdmin {
		sendJSONError(w, "Admin privileges required", http.StatusForbidden, "")
		return
	}

	// Parse the URL to get document path and comment ID
	// Format: /api/comments/delete/{docPath}/{commentID}
	path := strings.TrimPrefix(r.URL.Path, "/api/comments/delete/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		sendJSONError(w, "Invalid request format", http.StatusBadRequest, "")
		return
	}

	// Last element is the comment ID
	commentID := parts[len(parts)-1]
	// Everything else is the document path
	rawDocPath := strings.Join(parts[:len(parts)-1], "/")
	document, err := resolveCommentDocument(rawDocPath)
	if err != nil {
		sendJSONError(w, "Invalid document path", http.StatusBadRequest, "")
		return
	}

	// Delete the comment
	err = comments.DeleteComment(commentsRoot(), commentID, document.logicalPath, true)
	if err != nil {
		sendJSONError(w, "Failed to delete comment", http.StatusInternalServerError, err.Error())
		return
	}

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Comment deleted successfully",
	})
}
