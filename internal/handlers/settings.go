package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
	"wiki-go/internal/auth"
	"wiki-go/internal/config"
	"wiki-go/internal/i18n"
)

// WikiSettingsRequest represents the request body for updating wiki settings
type WikiSettingsRequest struct {
	Title         string `json:"title"`
	Owner         string `json:"owner"`
	Notice        string `json:"notice"`
	Timezone      string `json:"timezone"`
	Private       bool   `json:"private"`
	MaxVersions   int    `json:"max_versions"`
	MaxUploadSize int    `json:"max_upload_size"`
	Language      string `json:"language"`
}

// WikiSettingsResponse represents the response for wiki settings
type WikiSettingsResponse struct {
	Title         string   `json:"title"`
	Owner         string   `json:"owner"`
	Notice        string   `json:"notice"`
	Timezone      string   `json:"timezone"`
	Private       bool     `json:"private"`
	MaxVersions   int      `json:"max_versions"`
	MaxUploadSize int      `json:"max_upload_size"`
	Language      string   `json:"language"`
	Languages     []string `json:"languages"`
}

// WikiSettingsHandler handles both GET and POST requests for wiki settings
func WikiSettingsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetWikiSettingsHandler(w, r)
	case http.MethodPost:
		UpdateWikiSettingsHandler(w, r)
	default:
		sendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed, "")
	}
}

// GetWikiSettingsHandler handles requests to get the current wiki settings
func GetWikiSettingsHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated and is an admin
	session := auth.GetSession(r)
	if session == nil || !session.IsAdmin {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized, "")
		return
	}

	// Get the current wiki settings from the global config
	response := WikiSettingsResponse{
		Title:         cfg.Wiki.Title,
		Owner:         cfg.Wiki.Owner,
		Notice:        cfg.Wiki.Notice,
		Timezone:      cfg.Wiki.Timezone,
		Private:       cfg.Wiki.Private,
		MaxVersions:   cfg.Wiki.MaxVersions,
		MaxUploadSize: cfg.Wiki.MaxUploadSize,
		Language:      cfg.Wiki.Language,
		Languages:     i18n.GetAvailableLanguages(),
	}

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateWikiSettingsHandler handles requests to update the wiki settings
func UpdateWikiSettingsHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated and is an admin
	session := auth.GetSession(r)
	if session == nil || !session.IsAdmin {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized, "")
		return
	}

	// Parse the request body
	var req WikiSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request payload", http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	// Validate the request
	if req.Title == "" || req.Owner == "" || req.Notice == "" || req.Timezone == "" {
		sendJSONError(w, "All fields are required", http.StatusBadRequest, "")
		return
	}

	// Create a copy of the current config
	updatedConfig := *cfg

	// Update the wiki settings
	updatedConfig.Wiki.Title = req.Title
	updatedConfig.Wiki.Owner = req.Owner
	updatedConfig.Wiki.Notice = req.Notice
	updatedConfig.Wiki.Timezone = req.Timezone
	updatedConfig.Wiki.Private = req.Private
	updatedConfig.Wiki.MaxVersions = req.MaxVersions
	updatedConfig.Wiki.MaxUploadSize = req.MaxUploadSize
	updatedConfig.Wiki.Language = req.Language

	// Save the updated config to file
	configPath := config.ConfigFilePath
	if err := saveConfig(configPath, &updatedConfig); err != nil {
		sendJSONError(w, "Failed to save configuration", http.StatusInternalServerError, err.Error())
		return
	}

	// Update the global config
	*cfg = updatedConfig

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Settings updated successfully",
	})
}

// saveConfig saves the configuration to a file
func saveConfig(path string, cfg *config.Config) error {
	// Create a temporary file with a unique name using timestamp
	tempFile := fmt.Sprintf("%s.%d.tmp", path, time.Now().UnixNano())
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}

	// Close the file properly regardless of what happens later
	defer func() {
		file.Close()
		// Attempt to clean up the temporary file
		os.Remove(tempFile)
	}()

	// Save to the temporary file
	if err := config.SaveConfig(cfg, file); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Make sure data is written to disk
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync temporary file: %w", err)
	}

	// Close the file explicitly to ensure it's fully written
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Read the content of the temporary file
	content, err := os.ReadFile(tempFile)
	if err != nil {
		return fmt.Errorf("failed to read temporary file: %w", err)
	}

	// Write directly to the target file
	// This approach avoids the rename operation that can cause locking issues
	err = os.WriteFile(path, content, 0644)
	if err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	return nil
}
