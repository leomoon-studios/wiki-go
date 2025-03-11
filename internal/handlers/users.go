package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"wiki-go/internal/auth"
	"wiki-go/internal/config"
	"wiki-go/internal/crypto"
)

// User represents a user in the response
type UserResponse struct {
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
}

// UserCreateRequest represents the request body for creating a user
type UserCreateRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"is_admin"`
}

// UserUpdateRequest represents the request body for updating a user
type UserUpdateRequest struct {
	Username    string `json:"username"`
	NewPassword string `json:"new_password,omitempty"`
	IsAdmin     bool   `json:"is_admin"`
}

// UsersHandler handles user management endpoints
func UsersHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated and is an admin
	session := auth.GetSession(r)
	if session == nil || !session.IsAdmin {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized, "")
		return
	}

	switch r.Method {
	case http.MethodGet:
		GetUsersHandler(w, r)
	case http.MethodPost:
		CreateUserHandler(w, r)
	case http.MethodPut:
		UpdateUserHandler(w, r)
	case http.MethodDelete:
		DeleteUserHandler(w, r)
	default:
		sendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed, "")
	}
}

// GetUsersHandler returns a list of all users (without passwords)
func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated and is an admin
	session := auth.GetSession(r)
	if session == nil || !session.IsAdmin {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized, "")
		return
	}

	// Convert users to response objects (without passwords)
	users := make([]UserResponse, 0, len(cfg.Users))
	for _, user := range cfg.Users {
		users = append(users, UserResponse{
			Username: user.Username,
			IsAdmin:  user.IsAdmin,
		})
	}

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users": users,
	})
}

// CreateUserHandler creates a new user
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated and is an admin
	session := auth.GetSession(r)
	if session == nil || !session.IsAdmin {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized, "")
		return
	}

	// Parse the request body
	var req UserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request payload", http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	// Validate request
	if req.Username == "" || req.Password == "" {
		sendJSONError(w, "Username and password are required", http.StatusBadRequest, "")
		return
	}

	// Check if username already exists
	for _, user := range cfg.Users {
		if user.Username == req.Username {
			sendJSONError(w, "Username already exists", http.StatusConflict, "")
			return
		}
	}

	// Hash the password
	hashedPassword, err := crypto.HashPassword(req.Password)
	if err != nil {
		sendJSONError(w, "Failed to hash password", http.StatusInternalServerError, err.Error())
		return
	}

	// Create a copy of the current config
	updatedConfig := *cfg

	// Add the new user
	updatedConfig.Users = append(updatedConfig.Users, config.User{
		Username: req.Username,
		Password: hashedPassword,
		IsAdmin:  req.IsAdmin,
	})

	// Save the updated config
	configPath := config.ConfigFilePath
	if err := saveConfig(configPath, &updatedConfig); err != nil {
		sendJSONError(w, "Failed to save configuration", http.StatusInternalServerError, err.Error())
		return
	}

	// Update the global config
	*cfg = updatedConfig

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User created successfully",
	})
}

// UpdateUserHandler updates an existing user
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated and is an admin
	session := auth.GetSession(r)
	if session == nil || !session.IsAdmin {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized, "")
		return
	}

	// Parse the request body
	var req UserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request payload", http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	// Validate request
	if req.Username == "" {
		sendJSONError(w, "Username is required", http.StatusBadRequest, "")
		return
	}

	// Create a copy of the current config
	updatedConfig := *cfg

	// Find and update the user
	userFound := false
	for i, user := range updatedConfig.Users {
		if user.Username == req.Username {
			// Update the user's admin status
			updatedConfig.Users[i].IsAdmin = req.IsAdmin

			// Update password if provided
			if req.NewPassword != "" {
				hashedPassword, err := crypto.HashPassword(req.NewPassword)
				if err != nil {
					sendJSONError(w, "Failed to hash password", http.StatusInternalServerError, err.Error())
					return
				}
				updatedConfig.Users[i].Password = hashedPassword
			}

			userFound = true
			break
		}
	}

	if !userFound {
		sendJSONError(w, "User not found", http.StatusNotFound, "")
		return
	}

	// Save the updated config
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
		"message": "User updated successfully",
	})
}

// DeleteUserHandler deletes a user
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated and is an admin
	session := auth.GetSession(r)
	if session == nil || !session.IsAdmin {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized, "")
		return
	}

	// Get username from query parameters
	username := r.URL.Query().Get("username")
	if username == "" {
		sendJSONError(w, "Username is required", http.StatusBadRequest, "")
		return
	}

	// Don't allow deleting your own account
	if session.Username == username {
		sendJSONError(w, "Cannot delete your own account", http.StatusBadRequest, "")
		return
	}

	// Create a copy of the current config
	updatedConfig := *cfg

	// Find and remove the user
	userFound := false
	for i, user := range updatedConfig.Users {
		if user.Username == username {
			// Remove this user from the slice
			updatedConfig.Users = append(updatedConfig.Users[:i], updatedConfig.Users[i+1:]...)
			userFound = true
			break
		}
	}

	if !userFound {
		sendJSONError(w, "User not found", http.StatusNotFound, "")
		return
	}

	// Make sure we don't delete the last admin
	adminCount := 0
	for _, user := range updatedConfig.Users {
		if user.IsAdmin {
			adminCount++
		}
	}

	if adminCount == 0 {
		sendJSONError(w, "Cannot delete the last admin user", http.StatusBadRequest, "")
		return
	}

	// Save the updated config
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
		"message": "User deleted successfully",
	})
}

// GetUserByUsername retrieves a user by username (for internal use)
func GetUserByUsername(username string) (*config.User, error) {
	for _, user := range cfg.Users {
		if user.Username == username {
			return &user, nil
		}
	}
	return nil, errors.New("user not found")
}
