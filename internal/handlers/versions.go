package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	urlpath "path"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"wiki-go/internal/auth"
	"wiki-go/internal/config"
	"wiki-go/internal/logger"
	"wiki-go/internal/utils"
)

// VersionInfo holds metadata about a document version
type VersionInfo struct {
	Timestamp string `json:"timestamp"`
	Path      string `json:"path"`
}

// VersionsListResponse is the JSON response for listing versions
type VersionsListResponse struct {
	Success  bool          `json:"success"`
	Versions []VersionInfo `json:"versions"`
	Message  string        `json:"message,omitempty"`
}

// VersionResponse is the JSON response for retrieving a specific version
type VersionResponse struct {
	Success bool   `json:"success"`
	Content string `json:"content"`
	Message string `json:"message,omitempty"`
}

// Helper to send a JSON error response
func sendJSONErrorVersion(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]interface{}{
		"success": false,
		"message": message,
	}
	json.NewEncoder(w).Encode(response)
}

// VersionsHandler manages requests related to document versions
func VersionsHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// Set JSON content type header
	w.Header().Set("Content-Type", "application/json")

	logger.Debug("Version handler request: %s %s", r.Method, r.URL.Path)

	// Extract the path from the URL
	// URL format: /api/versions/{document-path} or /api/versions/{document-path}/{version-timestamp}
	pathParts := strings.Split(r.URL.Path, "/api/versions/")
	if len(pathParts) < 2 {
		sendJSONErrorVersion(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// Get the document path (everything after "/api/versions/")
	docPath := pathParts[1]
	if _, err := utils.CanonicalDocumentRequestPath(docPath); err != nil {
		sendJSONErrorVersion(w, "Invalid document path", http.StatusBadRequest)
		return
	}

	// If docPath is empty, return an error
	if docPath == "" {
		sendJSONErrorVersion(w, "Document path is required", http.StatusBadRequest)
		return
	}

	logger.Debug("Processing version request for document path: %s", docPath)

	// Check for restore action first
	if strings.HasSuffix(r.URL.Path, "/restore") && r.Method == "POST" {
		// For restore requests, path format is: /api/versions/{docPath}/{timestamp}/restore
		// Extract docPath and timestamp by removing "/restore" from the end
		pathWithoutRestore := strings.TrimSuffix(docPath, "/restore")
		pathParts := strings.Split(pathWithoutRestore, "/")

		if len(pathParts) < 1 {
			sendJSONErrorVersion(w, "Invalid path format for restore", http.StatusBadRequest)
			return
		}

		// Last part is the timestamp
		timestamp := pathParts[len(pathParts)-1]

		// Remove timestamp from path to get docPath
		docPath = strings.TrimSuffix(pathWithoutRestore, "/"+timestamp)

		// Check if timestamp is valid (14 digits)
		if len(timestamp) != 14 || !utils.IsNumeric(timestamp) {
			sendJSONErrorVersion(w, "Invalid timestamp format", http.StatusBadRequest)
			return
		}

		handleVersionRestore(w, r, cfg, docPath, timestamp)
		return
	}

	// Check if we're handling a specific version or listing versions
	parts := strings.Split(docPath, "/")

	// If the last part looks like a timestamp (numeric and 14 chars), it's a version request
	isVersionRequest := len(parts) > 0 && len(parts[len(parts)-1]) == 14 && utils.IsNumeric(parts[len(parts)-1])

	if isVersionRequest {
		// Handle specific version request
		timestamp := parts[len(parts)-1]
		// Remove timestamp from the docPath to get the document path
		docPath = strings.TrimSuffix(docPath, "/"+timestamp)

		if r.Method == http.MethodDelete {
			handleDeleteVersion(w, r, cfg, docPath, timestamp)
			return
		}
		if r.Method != http.MethodGet {
			sendJSONErrorVersion(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handleGetVersion(w, r, cfg, docPath, timestamp)
		return
	}

	if r.Method != http.MethodGet {
		sendJSONErrorVersion(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// If we're here, it's a request to list versions
	handleListVersions(w, r, cfg, docPath)
}

// handleListVersions lists all versions for a document
func handleListVersions(w http.ResponseWriter, r *http.Request, cfg *config.Config, docPath string) {
	resolved, err := resolveVersionDocumentPaths(cfg, docPath)
	if err != nil {
		sendJSONErrorVersion(w, "Invalid document path", http.StatusBadRequest)
		return
	}
	if !canAccessLogicalDocument(auth.GetSession(r), cfg, resolved.logicalPath) {
		sendJSONErrorVersion(w, "Document access denied", http.StatusForbidden)
		return
	}
	docPath = resolved.apiPath
	versionsDir := resolved.versionsDir

	// Check if versions directory exists
	if _, err := os.Stat(versionsDir); os.IsNotExist(err) {
		// Return empty list if directory doesn't exist
		response := VersionsListResponse{
			Success:  true,
			Versions: []VersionInfo{},
			Message:  "No versions found",
		}

		json.NewEncoder(w).Encode(response)
		return
	}

	// Read all files in the versions directory
	files, err := os.ReadDir(versionsDir)
	if err != nil {
		sendJSONErrorVersion(w, "Failed to read versions directory", http.StatusInternalServerError)
		return
	}

	// Filter and process version files
	var versions []VersionInfo
	for _, file := range files {
		// Skip directories and non-md files
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		// Extract timestamp from filename (remove .md extension)
		timestamp := strings.TrimSuffix(file.Name(), ".md")

		// Only add valid timestamp files (14 digits: yyyymmddhhmmss)
		if len(timestamp) == 14 && utils.IsNumeric(timestamp) {
			if _, err := versionFilePath(cfg, resolved.logicalPath, timestamp); err != nil {
				continue
			}
			versions = append(versions, VersionInfo{
				Timestamp: timestamp,
				Path:      urlpath.Join(docPath, timestamp),
			})
		}
	}

	// Sort versions by timestamp (newest first)
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Timestamp > versions[j].Timestamp
	})

	// Return the versions list
	response := VersionsListResponse{
		Success:  true,
		Versions: versions,
	}

	json.NewEncoder(w).Encode(response)
}

// handleGetVersion retrieves the content of a specific version
func handleGetVersion(w http.ResponseWriter, r *http.Request, cfg *config.Config, docPath, timestamp string) {
	resolved, err := resolveVersionDocumentPaths(cfg, docPath)
	if err != nil {
		sendJSONErrorVersion(w, "Invalid document path", http.StatusBadRequest)
		return
	}
	if !canAccessLogicalDocument(auth.GetSession(r), cfg, resolved.logicalPath) {
		sendJSONErrorVersion(w, "Document access denied", http.StatusForbidden)
		return
	}
	versionPath, err := versionFilePath(cfg, resolved.logicalPath, timestamp)
	if err != nil {
		sendJSONErrorVersion(w, "Invalid version file", http.StatusBadRequest)
		return
	}

	// Check if version file exists
	if _, err := os.Stat(versionPath); os.IsNotExist(err) {
		sendJSONErrorVersion(w, "Version not found", http.StatusNotFound)
		return
	}

	// Read the version file content
	content, err := os.ReadFile(versionPath)
	if err != nil {
		sendJSONErrorVersion(w, "Failed to read version file", http.StatusInternalServerError)
		return
	}

	// Return the version content
	response := VersionResponse{
		Success: true,
		Content: string(content),
	}

	json.NewEncoder(w).Encode(response)
}

// handleVersionRestore restores a document to a specific version
func handleVersionRestore(w http.ResponseWriter, r *http.Request, cfg *config.Config, docPath, timestamp string) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		sendJSONErrorVersion(w, "Method not allowed. Use POST to restore a version.", http.StatusMethodNotAllowed)
		return
	}
	session := auth.GetSession(r)
	if session == nil || (session.Role != config.RoleAdmin && session.Role != config.RoleEditor) {
		sendJSONErrorVersion(w, "Editor access required", http.StatusUnauthorized)
		return
	}
	resolved, err := resolveVersionDocumentPaths(cfg, docPath)
	if err != nil {
		sendJSONErrorVersion(w, "Invalid document path", http.StatusBadRequest)
		return
	}
	if !canAccessLogicalDocument(auth.GetSession(r), cfg, resolved.logicalPath) {
		sendJSONErrorVersion(w, "Document access denied", http.StatusForbidden)
		return
	}
	docPath = resolved.apiPath

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Disable all caching to ensure fresh content
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	logger.Debug("Restore request: docPath=%s, timestamp=%s", docPath, timestamp)

	versionFilename, err := versionFilePath(cfg, resolved.logicalPath, timestamp)
	if err != nil {
		sendJSONErrorVersion(w, "Invalid version file", http.StatusBadRequest)
		return
	}
	documentPath, err := documentContentPath(utils.ResolvedDocumentPath{LogicalPath: resolved.logicalPath, FilesystemPath: resolved.documentDir})
	if err != nil {
		sendJSONErrorVersion(w, "Invalid document file", http.StatusBadRequest)
		return
	}

	logger.Debug("Version file path: %s", versionFilename)
	logger.Debug("Document path for restore: %s", documentPath)

	// Check if version file exists
	if _, err := os.Stat(versionFilename); os.IsNotExist(err) {
		sendJSONErrorVersion(w, "Version not found", http.StatusNotFound)
		return
	}

	// Ensure the document directory exists
	docDir := filepath.Dir(documentPath)
	if err := os.MkdirAll(docDir, 0755); err != nil {
		logger.Error("Error creating directory: %v", err)
		sendJSONErrorVersion(w, "Failed to ensure document directory exists", http.StatusInternalServerError)
		return
	}

	// Read the version file content
	versionContent, err := os.ReadFile(versionFilename)
	if err != nil {
		logger.Error("Error reading version file: %v", err)
		sendJSONErrorVersion(w, "Failed to read version file", http.StatusInternalServerError)
		return
	}

	// Before overwriting current document, save it as a version
	if _, err := os.Stat(documentPath); err == nil && cfg.Wiki.MaxVersions > 0 {
		// Document exists, read its current content
		currentContent, err := os.ReadFile(documentPath)
		if err == nil && len(currentContent) > 0 {
			// Create timestamp for version filename
			newTimestamp := time.Now().Format("20060102150405") // Format: yyyymmddhhmmss

			// Ensure versions directory exists
			if err := os.MkdirAll(resolved.versionsDir, 0755); err == nil {
				// Create version file path with timestamp
				newVersionPath, pathErr := versionFilePath(cfg, resolved.logicalPath, newTimestamp)
				if pathErr != nil {
					sendJSONErrorVersion(w, "Invalid version backup", http.StatusBadRequest)
					return
				}

				// Save the current content as a version
				_ = os.WriteFile(newVersionPath, currentContent, 0644) // Ignore error for now

				// Clean up old versions if needed
				utils.CleanupOldVersions(resolved.versionsDir, cfg.Wiki.MaxVersions)
			}
		}
	}

	// Write the version content to the document file
	if err := os.WriteFile(documentPath, versionContent, 0644); err != nil {
		logger.Error("Error writing to document file: %v", err)
		sendJSONErrorVersion(w, "Failed to restore document", http.StatusInternalServerError)
		return
	}

	// Force update the file's modification time to ensure cache invalidation
	now := time.Now()
	if err := os.Chtimes(documentPath, now, now); err != nil {
		logger.Warn("Couldn't update file timestamp: %v", err)
		// Continue anyway, not critical
	}

	user := "unknown"
	if session != nil {
		user = session.Username
	}
	logger.Info("User %s restored version %s of %s", user, timestamp, documentPath)

	// Return success response
	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Document restored to version from %s", timestamp),
	}

	json.NewEncoder(w).Encode(response)
}

type resolvedVersionPaths struct {
	logicalPath string
	apiPath     string
	documentDir string
	versionsDir string
}

func resolveVersionDocumentPaths(cfg *config.Config, documentPath string) (resolvedVersionPaths, error) {
	document, err := resolveAPIDocument(cfg, documentPath, nil, false, true)
	if err != nil {
		return resolvedVersionPaths{}, err
	}
	versions, err := metadataPath(cfg, "versions", document.LogicalPath)
	if err != nil {
		return resolvedVersionPaths{}, err
	}
	apiPath := strings.TrimPrefix(document.LogicalPath, "/")
	if document.LogicalPath == "/" {
		apiPath = "pages/home"
	}
	return resolvedVersionPaths{document.LogicalPath, apiPath, document.FilesystemPath, versions}, nil
}

func versionFilePath(cfg *config.Config, logical, timestamp string) (string, error) {
	if len(timestamp) != 14 || !utils.IsNumeric(timestamp) {
		return "", fmt.Errorf("invalid timestamp")
	}
	return utils.ResolveStoragePath(cfg.Wiki.RootDir, "versions/"+versionStorageName(logical)+"/"+timestamp+".md")
}

func handleDeleteVersion(w http.ResponseWriter, r *http.Request, cfg *config.Config, docPath, timestamp string) {
	session := auth.GetSession(r)
	if session == nil || (session.Role != config.RoleAdmin && session.Role != config.RoleEditor) {
		sendJSONErrorVersion(w, "Editor access required", http.StatusUnauthorized)
		return
	}
	document, ok := requireAPIDocument(w, cfg, docPath, session, false, true)
	if !ok {
		return
	}
	filename, err := versionFilePath(cfg, document.LogicalPath, timestamp)
	if err != nil {
		sendJSONErrorVersion(w, "Invalid version file", http.StatusBadRequest)
		return
	}
	if err := os.Remove(filename); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, os.ErrNotExist) {
			status = http.StatusNotFound
		}
		sendJSONErrorVersion(w, "Failed to delete version", status)
		return
	}
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
