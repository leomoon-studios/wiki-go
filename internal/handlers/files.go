package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"wiki-go/internal/auth"
	"wiki-go/internal/config"
	"wiki-go/internal/i18n"
	"wiki-go/internal/utils"
)

// FileResponse represents the response for file operations
type FileResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message"`
	URL     string     `json:"url,omitempty"`
	Files   []FileInfo `json:"files,omitempty"`
}

// FileInfo represents information about a file
type FileInfo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Size int64  `json:"size"` // Size in bytes
	Type string `json:"type"` // MIME type or extension
}

// Document represents a document in the wiki
type Document struct {
	Title string `json:"title"`
	Path  string `json:"path"`
}

// DocumentsResponse represents the response for the documents list API
type DocumentsResponse struct {
	Success   bool       `json:"success"`
	Message   string     `json:"message,omitempty"`
	Documents []Document `json:"documents"`
}

// FolderInfo represents information about a folder
type FolderInfo struct {
	Path  string `json:"path"`
	Name  string `json:"name"`
	Level int    `json:"level"`
}

// UploadFileHandler handles file uploads to the document's directory
func UploadFileHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// Set appropriate headers
	w.Header().Set("Content-Type", "application/json")

	// Check if user is authenticated and has appropriate permissions
	session := auth.GetSession(r)
	if session == nil || (session.Role != config.RoleAdmin && session.Role != config.RoleEditor) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Unauthorized. Admin or editor access required.",
		})
		return
	}

	// Only allow POST method
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Method not allowed. Use POST for file uploads.",
		})
		return
	}

	// Get the maximum upload size from the config
	maxUploadSize := config.GetMaxUploadSizeBytes(cfg)
	maxUploadSizeFormatted := config.GetMaxUploadSizeFormatted(cfg)

	// Parse the multipart form with a maximum size
	err := r.ParseMultipartForm(maxUploadSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Failed to parse form or file too large. Maximum size is " + maxUploadSizeFormatted + ".",
		})
		return
	}

	// Get the document path from the request
	docPath := r.FormValue("docPath")
	if docPath == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Document path is required.",
		})
		return
	}

	// Clean and normalize the path
	docPath = filepath.Clean(docPath)
	docPath = strings.TrimSuffix(docPath, "/")
	docPath = strings.ReplaceAll(docPath, "\\", "/")

	// Special case for homepage
	if docPath == "" || docPath == "/" {
		docPath = "pages/home"
	}

	// Determine the full filesystem path to the document's directory
	var uploadDir string
	if strings.HasPrefix(docPath, "pages/") {
		// For pages directory (like homepage), don't add the documents directory
		uploadDir = filepath.Join(cfg.Wiki.RootDir, docPath)
	} else {
		// For regular documents
		uploadDir = filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, docPath)
	}

	// Check if directory exists
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Document directory does not exist.",
		})
		return
	}

	// Get the uploaded file
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Failed to get uploaded file.",
		})
		return
	}
	defer file.Close()

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !cfg.Wiki.DisableFileUploadChecking && !config.IsAllowedExtension(ext) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Invalid file type. Allowed extensions: " + config.GetAllowedExtensionsDisplayText(),
		})
		return
	}

	// Read a larger buffer to better detect the actual content type
	buffer := make([]byte, 8192)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Failed to read file content.",
		})
		return
	}
	buffer = buffer[:n] // Resize to actual read size

	// Reset the file pointer to the beginning
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Failed to process file.",
		})
		return
	}

	// Check if MIME type validation is disabled in settings
	if !cfg.Wiki.DisableFileUploadChecking {
		// Use enhanced detection for content type
		detectedContentType, err := detectFileContentType(buffer, fileHeader.Filename)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(FileResponse{
				Success: false,
				Message: "Failed to detect file content type.",
			})
			return
		}
		expectedContentType := config.GetMimeTypeForExtension(ext)

		// Check if the detected content type matches what we expect for this extension
		// Note: http.DetectContentType is limited and may return generic types like "application/octet-stream"
		// So we need to be careful with the validation logic
		if !isContentTypeCompatible(detectedContentType, expectedContentType, buffer, fileHeader.Filename) {
			// Debug info for file validation issues
			debugFileValidation(buffer, fileHeader.Filename, detectedContentType, expectedContentType)

			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(FileResponse{
				Success: false,
				Message: i18n.Translate("attachments.error_content_mismatch"),
			})
			return
		}
	}

	// Create safe filename - remove any potentially unsafe characters
	filename := sanitizeFilename(fileHeader.Filename)

	// Special handling for SVG files to prevent XSS attacks
	if strings.ToLower(filepath.Ext(filename)) == ".svg" && !cfg.Wiki.DisableFileUploadChecking {
		// Read the entire file content
		svgContent, err := io.ReadAll(file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(FileResponse{
				Success: false,
				Message: "Failed to read SVG file content.",
			})
			return
		}

		// Sanitize the SVG content
		sanitizedSVG, err := sanitizeSVG(svgContent)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(FileResponse{
				Success: false,
				Message: i18n.Translate("attachments.error_svg_sanitization"),
			})
			return
		}

		// Full path where the file will be saved
		savePath := filepath.Join(uploadDir, filename)

		// Write the sanitized SVG directly to the file
		if err := os.WriteFile(savePath, sanitizedSVG, 0644); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(FileResponse{
				Success: false,
				Message: "Failed to save sanitized SVG file.",
			})
			return
		}

		// Create URL path for the file
		urlPath := filepath.Join("/api/files", docPath, filename)
		// Replace backslashes with forward slashes for URLs
		urlPath = strings.ReplaceAll(urlPath, "\\", "/")

		// Return success response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(FileResponse{
			Success: true,
			Message: "File uploaded successfully.",
			URL:     urlPath,
		})
		return
	}

	// Full path where the file will be saved
	savePath := filepath.Join(uploadDir, filename)

	// Create destination file
	dst, err := os.Create(savePath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Failed to create destination file.",
		})
		return
	}
	defer dst.Close()

	// Copy the uploaded file to the destination file
	_, err = io.Copy(dst, file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Failed to save uploaded file.",
		})
		return
	}

	// Create URL path for the file
	urlPath := filepath.Join("/api/files", docPath, filename)
	// Replace backslashes with forward slashes for URLs
	urlPath = strings.ReplaceAll(urlPath, "\\", "/")

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(FileResponse{
		Success: true,
		Message: "File uploaded successfully.",
		URL:     urlPath,
	})
}

// ListFilesHandler returns a list of files in the document's directory
func ListFilesHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// Set appropriate headers
	w.Header().Set("Content-Type", "application/json")

	// Only allow GET method
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Method not allowed. Use GET to list files.",
		})
		return
	}

	// Get the document path from the URL
	path := strings.TrimPrefix(r.URL.Path, "/api/files/list")

	// Remove leading slash if present
	path = strings.TrimPrefix(path, "/")

	// Special case for homepage
	if path == "" || path == "/" {
		path = "pages/home"
	}

	// Clean and normalize the path
	path = filepath.Clean(path)
	path = strings.TrimSuffix(path, "/")
	path = strings.ReplaceAll(path, "\\", "/")

	// Determine logical path for access check
	logicalPath := "/" + path
	if path == "pages/home" {
		logicalPath = "/"
	} else if strings.HasPrefix(path, "pages/home/") {
		// Handle potential sub-resources of homepage if any
		logicalPath = "/" + strings.TrimPrefix(path, "pages/home/")
	}

	// Check access permissions
	session := auth.GetSession(r)
	if !auth.CanAccessDocument(logicalPath, session, cfg) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Unauthorized. You do not have permission to access files for this document.",
		})
		return
	}

	// Determine the full filesystem path to the document's directory
	var dirPath string
	if strings.HasPrefix(path, "pages/") {
		// For pages directory (like homepage), don't add the documents directory
		dirPath = filepath.Join(cfg.Wiki.RootDir, path)
	} else {
		// For regular documents
		dirPath = filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, path)
	}

	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Document directory does not exist.",
		})
		return
	}

	// Read the directory contents
	files, err := os.ReadDir(dirPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Failed to read directory contents.",
		})
		return
	}

	// Filter for valid files and create response list
	var filesList []FileInfo
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Skip the document.md file
		if file.Name() == "document.md" {
			continue
		}

		// Get file extension
		ext := strings.ToLower(filepath.Ext(file.Name()))

		// Check if it's a valid file type we want to show
		if !cfg.Wiki.DisableFileUploadChecking && !config.IsAllowedExtension(ext) {
			continue
		}

		// Get file info for size
		fileInfo, err := file.Info()
		if err != nil {
			continue // Skip files with errors
		}

		// Create URL path for the file
		urlPath := filepath.Join("/api/files", path, file.Name())
		// Replace backslashes with forward slashes for URLs
		urlPath = strings.ReplaceAll(urlPath, "\\", "/")

		// Get the MIME type based on extension
		fileType := config.GetMimeTypeForExtension(ext)

		filesList = append(filesList, FileInfo{
			Name: file.Name(),
			URL:  urlPath,
			Size: fileInfo.Size(),
			Type: fileType,
		})
	}

	// Return success response with file list
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(FileResponse{
		Success: true,
		Message: "Files retrieved successfully.",
		Files:   filesList,
	})
}

// DeleteFileHandler handles deletion of a file
func DeleteFileHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// Set appropriate headers
	w.Header().Set("Content-Type", "application/json")

	// Check if user is authenticated and has appropriate permissions
	session := auth.GetSession(r)
	if session == nil || (session.Role != config.RoleAdmin && session.Role != config.RoleEditor) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Unauthorized. Admin or editor access required.",
		})
		return
	}

	// Only allow DELETE method
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Method not allowed. Use DELETE to remove files.",
		})
		return
	}

	// Get the file path from the URL
	path := strings.TrimPrefix(r.URL.Path, "/api/files/delete")

	// Remove leading slash if present
	path = strings.TrimPrefix(path, "/")

	// Clean and normalize the path
	path = filepath.Clean(path)
	path = strings.TrimSuffix(path, "/")
	path = strings.ReplaceAll(path, "\\", "/")

	// Verify we have a path
	if path == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Invalid file path.",
		})
		return
	}

	// Determine the full filesystem path to the file
	var filePath string
	if strings.HasPrefix(path, "pages/") {
		// For pages directory (like homepage), don't add the documents directory
		filePath = filepath.Join(cfg.Wiki.RootDir, path)
	} else {
		// For regular documents
		filePath = filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, path)
	}

	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "File not found.",
		})
		return
	}

	// Ensure it's not a directory
	if fileInfo.IsDir() {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Path is a directory, not a file.",
		})
		return
	}

	// Don't allow deleting document.md
	if filepath.Base(filePath) == "document.md" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Cannot delete the document file itself.",
		})
		return
	}

	// Delete the file
	err = os.Remove(filePath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Failed to delete file.",
		})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(FileResponse{
		Success: true,
		Message: "File deleted successfully.",
	})
}

// ServeFileHandler serves the actual files
func ServeFileHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the file path from the URL
	path := strings.TrimPrefix(r.URL.Path, "/api/files/")

	// Remove leading slash if present
	path = strings.TrimPrefix(path, "/")

	// Clean and normalize the path
	path = filepath.Clean(path)
	path = strings.ReplaceAll(path, "\\", "/")

	// Check access permissions
	// We need to determine the document path from the file path
	// The path is like "pages/home/image.png" or "finance/doc/image.png"
	docPath := filepath.Dir(path)
	
	// Determine logical path for access check
	logicalPath := "/" + docPath
	if docPath == "pages/home" {
		logicalPath = "/"
	} else if strings.HasPrefix(docPath, "pages/home/") {
		logicalPath = "/" + strings.TrimPrefix(docPath, "pages/home/")
	} else if docPath == "." {
		// Should not happen for valid document attachments, but handle gracefully
		logicalPath = "/"
	}

	session := auth.GetSession(r)
	if !auth.CanAccessDocument(logicalPath, session, cfg) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Determine the full filesystem path to the file
	var filePath string
	if strings.HasPrefix(path, "pages/") {
		// For pages directory (like homepage), don't add the documents directory
		filePath = filepath.Join(cfg.Wiki.RootDir, path)
	} else {
		// For regular documents
		filePath = filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, path)
	}

	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Ensure it's not a directory
	if fileInfo.IsDir() {
		http.Error(w, "Path is a directory, not a file", http.StatusBadRequest)
		return
	}

	// Get file extension
	ext := strings.ToLower(filepath.Ext(filePath))

	// SECURITY CHECK: Block access to markdown files
	if ext == ".md" {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// SECURITY CHECK: Block access to files named "document" with any extension
	if strings.ToLower(filepath.Base(filePath)) == "document.md" {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// SECURITY CHECK: Block access if path contains ".." to prevent directory traversal
	if strings.Contains(path, "..") {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Determine content type based on file extension
	contentType := config.GetMimeTypeForExtension(ext)

	// Additional security check: Verify content type for certain file types
	// This helps prevent serving files that have been tampered with after upload
	if !cfg.Wiki.DisableFileUploadChecking && config.ShouldVerifyContentType(ext[1:]) { // Remove leading dot with ext[1:]
		// Open the file to check its content
		f, err := os.Open(filePath)
		if err == nil {
			defer f.Close()

			// Read a larger buffer to detect content type
			buffer := make([]byte, 8192) // Increased buffer size for better detection
			n, err := f.Read(buffer)
			if err == nil {
				buffer = buffer[:n] // Resize to actual read bytes

				// Use enhanced content type detection
				detectedType, err := detectFileContentType(buffer, filePath)
				if err == nil {
					// Verify the content type matches what we expect
					if !isContentTypeCompatible(detectedType, contentType, buffer, filePath) {
						// If content doesn't match extension, block access
						http.Error(w, "File not found", http.StatusNotFound)
						return
					}
				}
			}

			// Reset file pointer for serving
			_, err = f.Seek(0, io.SeekStart)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		}
	}

	// Set content type and other headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// For SVG files, add security headers to prevent script execution
	if ext == ".svg" {
		// Add Content-Security-Policy header to prevent script execution in SVG
		w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'self'; img-src 'self'; object-src 'none'")
		// Add X-Content-Type-Options to prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")
	}

	// For binary files, set content disposition header for download
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/gif" && contentType != "text/plain" {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(filePath)))
	}

	// Serve the file
	http.ServeFile(w, r, filePath)
}

// Helper function to sanitize filenames
func sanitizeFilename(filename string) string {
	// Remove path information
	filename = filepath.Base(filename)

	// Replace potentially problematic characters
	filename = strings.ReplaceAll(filename, " ", "_")
	filename = strings.ReplaceAll(filename, "#", "")
	filename = strings.ReplaceAll(filename, "%", "")
	filename = strings.ReplaceAll(filename, "&", "")
	filename = strings.ReplaceAll(filename, "{", "")
	filename = strings.ReplaceAll(filename, "}", "")
	filename = strings.ReplaceAll(filename, "\\", "")
	filename = strings.ReplaceAll(filename, ":", "")
	filename = strings.ReplaceAll(filename, "<", "")
	filename = strings.ReplaceAll(filename, ">", "")
	filename = strings.ReplaceAll(filename, "*", "")
	filename = strings.ReplaceAll(filename, "?", "")
	filename = strings.ReplaceAll(filename, "|", "")
	filename = strings.ReplaceAll(filename, "\"", "")
	filename = strings.ReplaceAll(filename, "'", "")
	filename = strings.ReplaceAll(filename, ";", "")

	return filename
}

// Helper function to check if detected content type is compatible with expected type
func isContentTypeCompatible(detected, expected string, fileContent []byte, filename string) bool {
	// If they match exactly, it's compatible
	if detected == expected {
		return true
	}

	ext := strings.ToLower(filepath.Ext(filename))

	// Special handling for text-based files
	if ext == ".svg" || ext == ".txt" || ext == ".log" || ext == ".csv" {
		// For SVGs, check if content is XML or text-based
		if ext == ".svg" {
			return detected == "image/svg+xml" ||
				detected == "text/xml" ||
				detected == "application/xml" ||
				detected == "text/plain" ||
				isSVGContent(fileContent)
		}

		// For TXT, LOG, and CSV files, check if content is primarily text
		if ext == ".txt" || ext == ".log" || ext == ".csv" {
			return detected == "text/plain" ||
				strings.HasPrefix(detected, "text/") ||
				isTextContent(fileContent)
		}
	}

	// Special cases for common file types
	switch expected {
	case "image/jpeg":
		return detected == "image/jpeg" || detected == "image/pjpeg"
	case "image/png":
		return detected == "image/png"
	case "image/gif":
		return detected == "image/gif"
	case "image/webp":
		return detected == "image/webp"
	case "application/pdf":
		return detected == "application/pdf"
	case "text/plain":
		return detected == "text/plain" || strings.HasPrefix(detected, "text/")
	case "application/zip":
		return detected == "application/zip" || detected == "application/x-zip-compressed"
	case "video/mp4":
		return detected == "video/mp4" || strings.HasPrefix(detected, "video/")
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		// Word documents - check specific Office file structure
		return detected == "application/vnd.openxmlformats-officedocument.wordprocessingml.document" ||
			isOfficeFile(fileContent, "docx")
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		// Excel files - check specific Office file structure
		return detected == "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" ||
			isOfficeFile(fileContent, "xlsx")
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		// PowerPoint files - check specific Office file structure
		// Also accept application/zip for PPTX files since they are essentially ZIP files
		return detected == "application/vnd.openxmlformats-officedocument.presentationml.presentation" ||
			detected == "application/zip" || // Accept ZIP MIME type for PPTX files
			isOfficeFile(fileContent, "pptx")
	case "image/svg+xml":
		// SVG files might be detected as XML or plain text
		return detected == "image/svg+xml" || detected == "text/xml" ||
			detected == "application/xml" || detected == "text/plain"
	}

	// For unknown types, be conservative
	return false
}

// sanitizeSVG removes potentially harmful elements and attributes from SVG files
func sanitizeSVG(content []byte) ([]byte, error) {
	// Convert to string for easier manipulation
	svgStr := string(content)

	// Remove script tags and their content
	scriptRegex := regexp.MustCompile(`(?i)<script\b[^>]*>.*?</script>`)
	svgStr = scriptRegex.ReplaceAllString(svgStr, "")

	// Remove event handlers (attributes starting with "on")
	eventHandlerRegex := regexp.MustCompile(`(?i)\s+on\w+\s*=\s*["'][^"']*["']`)
	svgStr = eventHandlerRegex.ReplaceAllString(svgStr, "")

	// Remove javascript: URLs
	jsUrlRegex := regexp.MustCompile(`(?i)(href|xlink:href)\s*=\s*["']javascript:[^"']*["']`)
	svgStr = jsUrlRegex.ReplaceAllString(svgStr, `$1=""`)

	// Remove data: URLs
	dataUrlRegex := regexp.MustCompile(`(?i)(href|xlink:href)\s*=\s*["']data:[^"']*["']`)
	svgStr = dataUrlRegex.ReplaceAllString(svgStr, `$1=""`)

	// Remove external references (can be used for data exfiltration)
	externalRefRegex := regexp.MustCompile(`(?i)(href|xlink:href)\s*=\s*["']https?:[^"']*["']`)
	svgStr = externalRefRegex.ReplaceAllString(svgStr, `$1=""`)

	// Remove potentially dangerous tags
	dangerousTags := []string{"foreignObject", "use", "embed", "object", "iframe"}
	for _, tag := range dangerousTags {
		openTagRegex := regexp.MustCompile(`(?i)<` + tag + `\b[^>]*>`)
		closeTagRegex := regexp.MustCompile(`(?i)<\/` + tag + `\s*>`)
		svgStr = openTagRegex.ReplaceAllString(svgStr, "")
		svgStr = closeTagRegex.ReplaceAllString(svgStr, "")
	}

	return []byte(svgStr), nil
}

// isTextContent checks if content is primarily text-based by sampling bytes
func isTextContent(content []byte) bool {
	// Check a larger sample (up to 8KB)
	sampleSize := min(8192, len(content))
	sample := content[:sampleSize]

	// Count text characters vs. binary characters
	textCount := 0
	for _, b := range sample {
		// Check if the byte is a printable ASCII character or common whitespace
		if (b >= 32 && b <= 126) || b == 9 || b == 10 || b == 13 {
			textCount++
		}
	}

	// If more than 90% of characters are text, consider it text content
	return float64(textCount)/float64(sampleSize) > 0.9
}

// isSVGContent checks if content appears to be an SVG file
func isSVGContent(content []byte) bool {
	// Convert to string for easier pattern matching
	contentStr := string(content)

	// Look for SVG-specific patterns
	hasSVGTag := strings.Contains(strings.ToLower(contentStr), "<svg")
	hasXMLNamespace := strings.Contains(contentStr, "http://www.w3.org/2000/svg")

	// Check for XML structure
	isXML := len(contentStr) > 0 &&
		(strings.HasPrefix(strings.TrimSpace(contentStr), "<?xml") ||
			strings.HasPrefix(strings.TrimSpace(contentStr), "<svg"))

	return isXML && (hasSVGTag || hasXMLNamespace)
}

// detectFileContentType provides more sophisticated file type detection
func detectFileContentType(fileContent []byte, filename string) (string, error) {
	// Basic detection from standard library
	basicMimeType := http.DetectContentType(fileContent)
	ext := strings.ToLower(filepath.Ext(filename))

	// Check for text-based files specifically
	if isTextContent(fileContent) {
		// Special handling for SVG files
		if ext == ".svg" || isSVGContent(fileContent) {
			return "image/svg+xml", nil
		}

		// For TXT files or files detected as text
		if ext == ".txt" {
			return "text/plain", nil
		}
	}

	// Special handling for Office files based on extension and ZIP signature
	if (ext == ".docx" || ext == ".xlsx" || ext == ".pptx") &&
		len(fileContent) >= 4 && string(fileContent[0:4]) == "PK\x03\x04" {

		// Check the Office file structure
		switch ext {
		case ".docx":
			if isOfficeFile(fileContent, "docx") {
				return "application/vnd.openxmlformats-officedocument.wordprocessingml.document", nil
			}
		case ".xlsx":
			if isOfficeFile(fileContent, "xlsx") {
				return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", nil
			}
		case ".pptx":
			if isOfficeFile(fileContent, "pptx") {
				return "application/vnd.openxmlformats-officedocument.presentationml.presentation", nil
			}
		}
	}

	return basicMimeType, nil
}

// Helper function for min of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// isOfficeFile checks if content matches the structure of the specified Office file type
func isOfficeFile(content []byte, officeType string) bool {
	// Check for ZIP file signature (PK..)
	if len(content) < 4 || string(content[0:4]) != "PK\x03\x04" {
		return false
	}

	// Special handling for PPTX files - if they have the PK header, we'll be more lenient
	if officeType == "pptx" {
		// For PPTX, just having a valid ZIP signature is often good enough
		// since we've already validated the file extension
		return true
	}

	// Create a reader from the content
	reader := bytes.NewReader(content)

	// Open content as a ZIP archive
	zipReader, err := zip.NewReader(reader, int64(len(content)))
	if err != nil {
		// If we can't open as ZIP but it has PK header, give DOCX and XLSX a chance too
		// just not as permissive as PPTX
		if officeType == "docx" || officeType == "xlsx" {
			// Look for common patterns in the raw content
			contentStr := string(content)
			if officeType == "docx" && strings.Contains(contentStr, "word/document.xml") {
				return true
			}
			if officeType == "xlsx" && strings.Contains(contentStr, "xl/workbook.xml") {
				return true
			}
		}
		return false
	}

	// Define marker files for each format
	var markerFiles []string
	switch officeType {
	case "docx":
		markerFiles = []string{"word/document.xml"}
	case "xlsx":
		markerFiles = []string{"xl/workbook.xml"}
	case "pptx":
		markerFiles = []string{"ppt/presentation.xml"}
	default:
		return false
	}

	// Check file listing for any of the marker files
	for _, markerFile := range markerFiles {
		for _, file := range zipReader.File {
			// Normalize path separators for comparison
			normalizedPath := strings.ReplaceAll(file.Name, "\\", "/")
			if normalizedPath == markerFile {
				return true
			}
		}
	}

	// Double-check for partial matches - sometimes paths might have prefix differences
	if officeType == "pptx" {
		for _, file := range zipReader.File {
			if strings.Contains(file.Name, "presentation.xml") {
				return true
			}
		}
	}

	return false
}

// debugFileValidation helps diagnose file validation issues (only active in dev mode)
func debugFileValidation(fileContent []byte, filename string, detected, expected string) {
	// Skip if not in dev mode
	// if !config.IsDevMode() {
	//    return
	// }

	ext := strings.ToLower(filepath.Ext(filename))
	fmt.Printf("File Validation Debug (%s):\n", filename)
	fmt.Printf("  - Detected MIME: %s\n", detected)
	fmt.Printf("  - Expected MIME: %s\n", expected)
	fmt.Printf("  - Content starts with: %x\n", fileContent[:min(16, len(fileContent))])

	// For Office files, try to show ZIP contents
	if ext == ".docx" || ext == ".xlsx" || ext == ".pptx" {
		// Check for ZIP signature
		if len(fileContent) >= 4 && string(fileContent[0:4]) == "PK\x03\x04" {
			fmt.Println("  - Has valid ZIP signature")

			// Try to open as ZIP
			reader := bytes.NewReader(fileContent)
			zipReader, err := zip.NewReader(reader, int64(len(fileContent)))
			if err != nil {
				fmt.Printf("  - Failed to open as ZIP: %v\n", err)
			} else {
				fmt.Println("  - ZIP contents:")
				for i, f := range zipReader.File {
					if i < 10 { // Limit to first 10 files
						fmt.Printf("    * %s\n", f.Name)
					} else if i == 10 {
						fmt.Println("    * ... (more files)")
						break
					}
				}
			}
		} else {
			fmt.Println("  - Missing ZIP signature!")
		}
	}
}

// ListDocumentsHandler handles requests to list all documents for the document picker
func ListDocumentsHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Check if user is authenticated and has appropriate permissions
	session := auth.GetSession(r)
	if session == nil || (session.Role != config.RoleAdmin && session.Role != config.RoleEditor) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(DocumentsResponse{
			Success: false,
			Message: "Unauthorized. Admin or editor access required.",
		})
		return
	}

	// Paths to scan for documents
	documentsPath := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir)

	var documents []Document

	// Find all documents in documents directory
	err := filepath.WalkDir(documentsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check if this is a document.md file
		if !d.IsDir() && d.Name() == "document.md" {
			// Get the directory path (document path)
			docDir := filepath.Dir(path)
			// Convert to relative path
			relPath, err := filepath.Rel(cfg.Wiki.RootDir, docDir)
			if err != nil {
				return nil // Skip this file
			}

			// Format the path for use in URLs
			relPath = strings.ReplaceAll(relPath, "\\", "/")

			// Get the document title from the markdown file
			title := extractTitleFromMarkdown(path)
			if title == "" {
				// If no title found, use the parent directory name
				title = filepath.Base(docDir)
			}

			// For regular documents, we want to remove the documents/ prefix
			// since it's not part of the visible URL
			if strings.HasPrefix(relPath, cfg.Wiki.DocumentsDir) {
				relPath = strings.TrimPrefix(relPath, cfg.Wiki.DocumentsDir)
				// Remove any leading slash that might remain
				relPath = strings.TrimPrefix(relPath, "/")
			}

			// Add to documents list
			documents = append(documents, Document{
				Title: title,
				Path:  "/" + relPath,
			})
		}

		return nil
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(DocumentsResponse{
			Success: false,
			Message: "Failed to list documents: " + err.Error(),
		})
		return
	}

	// Return the documents list
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(DocumentsResponse{
		Success:   true,
		Documents: documents,
	})
}

// extractTitleFromMarkdown reads a markdown file and extracts the first h1 heading
func extractTitleFromMarkdown(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}

	// Look for the first h1 heading (# Title)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}

	return ""
}

// RenameFileHandler handles renaming of a file
func RenameFileHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// Set appropriate headers
	w.Header().Set("Content-Type", "application/json")

	// Check if user is authenticated and has appropriate permissions
	session := auth.GetSession(r)
	if session == nil || (session.Role != config.RoleAdmin && session.Role != config.RoleEditor) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Unauthorized. Admin or editor access required.",
		})
		return
	}

	// Only allow POST method
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Method not allowed. Use POST to rename files.",
		})
		return
	}

	// Parse request body
	var renameReq struct {
		CurrentPath string `json:"currentPath"`
		NewName     string `json:"newName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&renameReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Invalid request format.",
		})
		return
	}

	// Validate request
	if renameReq.CurrentPath == "" || renameReq.NewName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Current path and new name are required.",
		})
		return
	}

	// Debug incoming request
	fmt.Printf("Rename request received - CurrentPath: %s, NewName: %s\n", renameReq.CurrentPath, renameReq.NewName)

	// Clean and normalize the path
	path := renameReq.CurrentPath
	path = filepath.Clean(path)
	path = strings.TrimSuffix(path, "/")
	path = strings.ReplaceAll(path, "\\", "/")

	// Extract directory and filename
	dir := filepath.Dir(path)
	filename := filepath.Base(path)
	newPath := filepath.Join(dir, renameReq.NewName)

	// Log paths for debugging
	fmt.Printf("Path components: path=%s, dir=%s, filename=%s, newPath=%s\n", path, dir, filename, newPath)

	// Determine file paths based on two possible locations
	var currentFilePath, newFilePath string
	var fileFound bool

	// Try the documents directory first (most common case)
	currentDocumentsPath := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, path)
	if fileExists(currentDocumentsPath) {
		currentFilePath = currentDocumentsPath
		newFilePath = filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, newPath)
		fileFound = true
		fmt.Printf("File found in documents path: %s\n", currentFilePath)
	}

	// If not found in documents, try pages directory
	if !fileFound && strings.HasPrefix(path, "pages/") {
		currentPagesPath := filepath.Join(cfg.Wiki.RootDir, path)
		if fileExists(currentPagesPath) {
			currentFilePath = currentPagesPath
			newFilePath = filepath.Join(cfg.Wiki.RootDir, newPath)
			fileFound = true
			fmt.Printf("File found in pages path: %s\n", currentFilePath)
		}
	}

	// If file not found in either location, check if the target file already exists
	// which could mean the file was already renamed
	if !fileFound {
		// Check if the destination file already exists with the new name
		possibleNewPathInDocuments := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, newPath)
		possibleNewPathInPages := filepath.Join(cfg.Wiki.RootDir, newPath)

		if fileExists(possibleNewPathInDocuments) ||
		   (strings.HasPrefix(newPath, "pages/") && fileExists(possibleNewPathInPages)) {
			// The file with the new name already exists, likely was already renamed
			fmt.Printf("File already appears to have been renamed to: %s\n", newPath)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(FileResponse{
				Success: true,
				Message: "File already renamed.",
				URL:     "/api/files/" + newPath,
			})
			return
		}

		// If we get here, the file truly doesn't exist
		fmt.Printf("Error: File not found in documents or pages path\n")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "File not found.",
		})
		return
	}

	// Log the full paths for debugging
	fmt.Printf("Renaming file: %s -> %s\n", currentFilePath, newFilePath)

	// Check if target already exists
	if _, err := os.Stat(newFilePath); err == nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "A file with this name already exists.",
		})
		return
	}

	// Rename the file
	err := os.Rename(currentFilePath, newFilePath)
	if err != nil {
		fmt.Printf("Error renaming file: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Failed to rename file: " + err.Error(),
		})
		return
	}

	// Create URL for the renamed file
	urlPath := filepath.Join("/api/files", newPath)
	// Replace backslashes with forward slashes for URLs
	urlPath = strings.ReplaceAll(urlPath, "\\", "/")

	fmt.Printf("File renamed successfully: %s -> %s\n", currentFilePath, newFilePath)
	fmt.Printf("URL path: %s\n", urlPath)

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(FileResponse{
		Success: true,
		Message: "File renamed successfully.",
		URL:     urlPath,
	})
}

// Helper function to check if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Helper function to check if a directory exists
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// ListFoldersHandler returns a hierarchical list of folders in the documents directory
func ListFoldersHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// Check if user is authenticated and has admin role
	session := auth.GetSession(r)
	if session == nil || session.Role != config.RoleAdmin {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Unauthorized. Admin access required.",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var folders []FolderInfo

	// Add root folder (homepage)
	folders = append(folders, FolderInfo{
		Path:  "/",
		Name:  "/ (homepage)",
		Level: 0,
	})

	rootDir := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir)

	// Walk the directory
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root dir itself
		if path == rootDir {
			return nil
		}

		// Only process directories
		if !d.IsDir() {
			return nil
		}

		// Skip hidden directories
		if strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}

		// Calculate relative path
		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return nil
		}

		// Normalize path separators
		relPath = filepath.ToSlash(relPath)

		// Calculate level (depth)
		// Top-level folders in documents should be level 0
		parts := strings.Split(relPath, "/")
		level := len(parts) - 1

		// Construct the full virtual path (e.g. /finance/reports)
		virtualPath := "/" + relPath

		folders = append(folders, FolderInfo{
			Path:  virtualPath,
			Name:  utils.GetDocumentTitle(path),
			Level: level,
		})

		return nil
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to list folders: " + err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"folders": folders,
	})
}
