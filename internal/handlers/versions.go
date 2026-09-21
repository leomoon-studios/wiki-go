package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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

		// Fetch the version content
		handleGetVersion(w, r, cfg, docPath, timestamp)
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
			versions = append(versions, VersionInfo{
				Timestamp: timestamp,
				Path:      filepath.Join(docPath, timestamp),
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
	versionPath := filepath.Join(resolved.versionsDir, timestamp+".md")

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

	versionFilePath := filepath.Join(resolved.versionsDir, timestamp+".md")
	documentPath := filepath.Join(resolved.documentDir, "document.md")

	logger.Debug("Version file path: %s", versionFilePath)
	logger.Debug("Document path for restore: %s", documentPath)

	// Check if version file exists
	if _, err := os.Stat(versionFilePath); os.IsNotExist(err) {
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
	versionContent, err := os.ReadFile(versionFilePath)
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
				newVersionPath := filepath.Join(resolved.versionsDir, newTimestamp+".md")

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

	session := auth.GetSession(r)
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
	initial, err := utils.ResolveRelativeDocumentPath(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, documentPath)
	if err != nil {
		return resolvedVersionPaths{}, err
	}
	storagePath := strings.TrimPrefix(initial.LogicalPath, "/")
	if storagePath == "pages/home" {
		documentDir, err := filepath.Abs(filepath.Join(cfg.Wiki.RootDir, "pages", "home"))
		if err != nil {
			return resolvedVersionPaths{}, err
		}
		versionsDir, err := filepath.Abs(filepath.Join(cfg.Wiki.RootDir, "versions", "pages", "home"))
		if err != nil {
			return resolvedVersionPaths{}, err
		}
		return resolvedVersionPaths{
			logicalPath: "/",
			apiPath:     "pages/home",
			documentDir: documentDir,
			versionsDir: versionsDir,
		}, nil
	}

	storagePath = strings.TrimPrefix(storagePath, "documents/")
	document, err := utils.ResolveRelativeDocumentPath(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, storagePath)
	if err != nil {
		return resolvedVersionPaths{}, err
	}
	versions, err := utils.ResolveRelativeDocumentPath(filepath.Join(cfg.Wiki.RootDir, "versions"), "documents", storagePath)
	if err != nil {
		return resolvedVersionPaths{}, err
	}
	return resolvedVersionPaths{
		logicalPath: document.LogicalPath,
		apiPath:     strings.TrimPrefix(document.LogicalPath, "/"),
		documentDir: document.FilesystemPath,
		versionsDir: versions.FilesystemPath,
	}, nil
}
