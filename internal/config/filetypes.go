package config

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FileTypeConfig defines a file type with its extension and MIME type
type FileTypeConfig struct {
	Extension   string
	MimeType    string
	DisplayName string // Optional, for display purposes
}

// AllowedFileTypes defines all allowed file types in the application
var AllowedFileTypes = []FileTypeConfig{
	{Extension: "jpg", MimeType: "image/jpeg", DisplayName: "JPEG Image"},
	{Extension: "jpeg", MimeType: "image/jpeg", DisplayName: "JPEG Image"},
	{Extension: "png", MimeType: "image/png", DisplayName: "PNG Image"},
	{Extension: "gif", MimeType: "image/gif", DisplayName: "GIF Image"},
	{Extension: "txt", MimeType: "text/plain", DisplayName: "Text File"},
	{Extension: "zip", MimeType: "application/zip", DisplayName: "ZIP Archive"},
	{Extension: "pdf", MimeType: "application/pdf", DisplayName: "PDF Document"},
	{Extension: "docx", MimeType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document", DisplayName: "Word Document"},
	{Extension: "mp4", MimeType: "video/mp4", DisplayName: "MP4 Video"},
}

// GetAllowedExtensions returns a slice of all allowed file extensions
func GetAllowedExtensions() []string {
	extensions := make([]string, len(AllowedFileTypes))
	for i, fileType := range AllowedFileTypes {
		extensions[i] = fileType.Extension
	}
	return extensions
}

// GetMimeTypeForExtension returns the MIME type for a given file extension
func GetMimeTypeForExtension(ext string) string {
	// Remove the leading dot if present
	if ext != "" && ext[0] == '.' {
		ext = ext[1:]
	}

	for _, fileType := range AllowedFileTypes {
		if fileType.Extension == ext {
			return fileType.MimeType
		}
	}
	return "application/octet-stream" // Default fallback
}

// IsAllowedExtension checks if a given extension is allowed
func IsAllowedExtension(ext string) bool {
	// Remove the leading dot if present
	if ext != "" && ext[0] == '.' {
		ext = ext[1:]
	}

	for _, fileType := range AllowedFileTypes {
		if fileType.Extension == ext {
			return true
		}
	}
	return false
}

// GetAllowedExtensionsDisplayText returns a comma-separated list of allowed extensions
func GetAllowedExtensionsDisplayText() string {
	extensions := GetAllowedExtensions()
	return strings.Join(extensions, ", ")
}

// GetAllowedExtensionsJSON returns a JSON-formatted array of allowed extensions for JS
func GetAllowedExtensionsJSON() string {
	extensions := GetAllowedExtensions()

	// First try proper JSON marshaling
	jsonData, err := json.Marshal(extensions)
	if err != nil {
		// Fallback to manual method if marshaling fails
		jsonArray := "["
		for i, ext := range extensions {
			if i > 0 {
				jsonArray += ", "
			}
			jsonArray += "\"" + ext + "\""
		}
		jsonArray += "]"
		return jsonArray
	}

	return string(jsonData)
}

// GetExtensionMimeTypesJSON returns a JSON-formatted object mapping extensions to MIME types
func GetExtensionMimeTypesJSON() string {
	jsonObj := "{"
	for i, fileType := range AllowedFileTypes {
		if i > 0 {
			jsonObj += ", "
		}
		jsonObj += "\"" + fileType.Extension + "\": \"" + fileType.MimeType + "\""
	}
	jsonObj += "}"
	return jsonObj
}

// GetMaxUploadSizeBytes returns the maximum upload size in bytes based on the config
func GetMaxUploadSizeBytes(cfg *Config) int64 {
	if cfg != nil && cfg.Wiki.MaxUploadSize > 0 {
		return int64(cfg.Wiki.MaxUploadSize) * 1024 * 1024
	}
	// Default fallback - 20MB
	return 20 * 1024 * 1024
}

// GetMaxUploadSizeFormatted returns the maximum upload size in a human-readable format
func GetMaxUploadSizeFormatted(cfg *Config) string {
	if cfg != nil && cfg.Wiki.MaxUploadSize > 0 {
		return fmt.Sprintf("%dMB", cfg.Wiki.MaxUploadSize)
	}
	return "20MB"
}
