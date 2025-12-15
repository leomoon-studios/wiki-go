package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"wiki-go/internal/auth"
	"wiki-go/internal/config"
)

// AccessRulesHandler handles requests for access rules
func AccessRulesHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated and has admin role
	session := auth.GetSession(r)
	if session == nil || session.Role != config.RoleAdmin {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized, "")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/access-rules")

	// Handle reorder endpoint
	if path == "/reorder" && r.Method == http.MethodPost {
		ReorderAccessRulesHandler(w, r)
		return
	}

	// Handle root endpoint /api/access-rules
	if path == "" || path == "/" {
		switch r.Method {
		case http.MethodGet:
			GetAccessRulesHandler(w, r)
		case http.MethodPost:
			CreateAccessRuleHandler(w, r)
		default:
			sendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed, "")
		}
		return
	}

	// Handle specific rule endpoint /api/access-rules/{index}
	// Extract index from path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 1 {
		index, err := strconv.Atoi(parts[0])
		if err != nil {
			sendJSONError(w, "Invalid rule index", http.StatusBadRequest, err.Error())
			return
		}

		switch r.Method {
		case http.MethodPut:
			UpdateAccessRuleHandler(w, r, index)
		case http.MethodDelete:
			DeleteAccessRuleHandler(w, r, index)
		default:
			sendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed, "")
		}
		return
	}

	sendJSONError(w, "Not found", http.StatusNotFound, "")
}

// GetAccessRulesHandler returns the list of access rules
func GetAccessRulesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rules": cfg.AccessRules,
	})
}

// CreateAccessRuleHandler creates a new access rule
func CreateAccessRuleHandler(w http.ResponseWriter, r *http.Request) {
	var rule config.AccessRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		sendJSONError(w, "Invalid request payload", http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	// Validate rule
	if rule.Pattern == "" {
		sendJSONError(w, "Pattern is required", http.StatusBadRequest, "")
		return
	}

	// Prevent recursive rules on root
	if rule.Pattern == "/**" {
		sendJSONError(w, "Recursive rules on root (/**) are not allowed.", http.StatusBadRequest, "")
		return
	}

	// Create a copy of the current config
	updatedConfig := *cfg

	// Add the new rule
	updatedConfig.AccessRules = append(updatedConfig.AccessRules, rule)

	// Save the updated config
	if err := saveConfig(config.ConfigFilePath, &updatedConfig); err != nil {
		sendJSONError(w, "Failed to save configuration", http.StatusInternalServerError, err.Error())
		return
	}

	// Update global config
	*cfg = updatedConfig

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Access rule created successfully",
		"rule":    rule,
	})
}

// UpdateAccessRuleHandler updates an existing access rule
func UpdateAccessRuleHandler(w http.ResponseWriter, r *http.Request, index int) {
	if index < 0 || index >= len(cfg.AccessRules) {
		sendJSONError(w, "Rule not found", http.StatusNotFound, "")
		return
	}

	var rule config.AccessRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		sendJSONError(w, "Invalid request payload", http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	// Validate rule
	if rule.Pattern == "" {
		sendJSONError(w, "Pattern is required", http.StatusBadRequest, "")
		return
	}

	// Prevent recursive rules on root
	if rule.Pattern == "/**" {
		sendJSONError(w, "Recursive rules on root (/**) are not allowed.", http.StatusBadRequest, "")
		return
	}

	// Create a copy of the current config
	updatedConfig := *cfg

	// Update the rule
	updatedConfig.AccessRules[index] = rule

	// Save the updated config
	if err := saveConfig(config.ConfigFilePath, &updatedConfig); err != nil {
		sendJSONError(w, "Failed to save configuration", http.StatusInternalServerError, err.Error())
		return
	}

	// Update global config
	*cfg = updatedConfig

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Access rule updated successfully",
		"rule":    rule,
	})
}

// DeleteAccessRuleHandler deletes an access rule
func DeleteAccessRuleHandler(w http.ResponseWriter, r *http.Request, index int) {
	if index < 0 || index >= len(cfg.AccessRules) {
		sendJSONError(w, "Rule not found", http.StatusNotFound, "")
		return
	}

	// Create a copy of the current config
	updatedConfig := *cfg

	// Remove the rule
	updatedConfig.AccessRules = append(updatedConfig.AccessRules[:index], updatedConfig.AccessRules[index+1:]...)

	// Save the updated config
	if err := saveConfig(config.ConfigFilePath, &updatedConfig); err != nil {
		sendJSONError(w, "Failed to save configuration", http.StatusInternalServerError, err.Error())
		return
	}

	// Update global config
	*cfg = updatedConfig

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Access rule deleted successfully",
	})
}

// ReorderAccessRulesHandler reorders the access rules
func ReorderAccessRulesHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Indices []int `json:"indices"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request payload", http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	if len(req.Indices) != len(cfg.AccessRules) {
		sendJSONError(w, "Invalid number of indices", http.StatusBadRequest, "")
		return
	}

	// Create a copy of the current config
	updatedConfig := *cfg
	newRules := make([]config.AccessRule, len(cfg.AccessRules))

	// Reorder rules
	for i, oldIndex := range req.Indices {
		if oldIndex < 0 || oldIndex >= len(cfg.AccessRules) {
			sendJSONError(w, "Invalid index", http.StatusBadRequest, "")
			return
		}
		newRules[i] = cfg.AccessRules[oldIndex]
	}

	updatedConfig.AccessRules = newRules

	// Save the updated config
	if err := saveConfig(config.ConfigFilePath, &updatedConfig); err != nil {
		sendJSONError(w, "Failed to save configuration", http.StatusInternalServerError, err.Error())
		return
	}

	// Update global config
	*cfg = updatedConfig

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Access rules reordered successfully",
		"rules":   newRules,
	})
}
