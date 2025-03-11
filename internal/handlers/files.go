package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"wiki-go/internal/auth"
	"wiki-go/internal/config"
	"wiki-go/internal/i18n"
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

// UploadFileHandler handles file uploads to the document's directory
func UploadFileHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// Set appropriate headers
	w.Header().Set("Content-Type", "application/json")

	// Check if user is authenticated and is admin
	session := auth.GetSession(r)
	if session == nil || !session.IsAdmin {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Unauthorized. Admin access required.",
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
	if !config.IsAllowedExtension(ext) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Invalid file type. Allowed extensions: " + config.GetAllowedExtensionsDisplayText(),
		})
		return
	}

	// Read a small buffer to detect the actual content type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Failed to read file content.",
		})
		return
	}

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

	// Detect content type from file content
	detectedContentType := http.DetectContentType(buffer)
	expectedContentType := config.GetMimeTypeForExtension(ext)

	// Check if the detected content type matches what we expect for this extension
	// Note: http.DetectContentType is limited and may return generic types like "application/octet-stream"
	// So we need to be careful with the validation logic
	if !isContentTypeCompatible(detectedContentType, expectedContentType) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: i18n.Translate("attachments.error_content_mismatch"),
		})
		return
	}

	// Create safe filename - remove any potentially unsafe characters
	filename := sanitizeFilename(fileHeader.Filename)

	// Special handling for SVG files to prevent XSS attacks
	if strings.ToLower(filepath.Ext(filename)) == ".svg" {
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
		if !config.IsAllowedExtension(ext) {
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

	// Check if user is authenticated and is admin
	session := auth.GetSession(r)
	if session == nil || !session.IsAdmin {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(FileResponse{
			Success: false,
			Message: "Unauthorized. Admin access required.",
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
	if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".pdf" {
		// Open the file to check its content
		f, err := os.Open(filePath)
		if err == nil {
			defer f.Close()

			// Read a buffer to detect content type
			buffer := make([]byte, 512)
			_, err = f.Read(buffer)
			if err == nil {
				detectedType := http.DetectContentType(buffer)

				// For images and PDFs, verify the content type matches what we expect
				if !isContentTypeCompatible(detectedType, contentType) {
					// If content doesn't match extension, block access
					http.Error(w, "File not found", http.StatusNotFound)
					return
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
func isContentTypeCompatible(detected, expected string) bool {
	// Some content types are generic or may vary slightly
	// For example, a JPEG might be detected as "image/jpeg" or "image/pjpeg"

	// If they match exactly, it's compatible
	if detected == expected {
		return true
	}

	// Special cases for common file types
	switch expected {
	case "image/jpeg":
		return detected == "image/jpeg" || detected == "image/pjpeg"
	case "image/png":
		return detected == "image/png"
	case "image/gif":
		return detected == "image/gif"
	case "application/pdf":
		return detected == "application/pdf"
	case "text/plain":
		return detected == "text/plain" || strings.HasPrefix(detected, "text/")
	case "application/zip":
		return detected == "application/zip" || detected == "application/x-zip-compressed"
	case "video/mp4":
		return detected == "video/mp4" || strings.HasPrefix(detected, "video/")
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		// Word documents might be detected as generic binary
		return detected == "application/vnd.openxmlformats-officedocument.wordprocessingml.document" ||
			detected == "application/octet-stream"
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
