package config

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"wiki-go/internal/crypto"
	"wiki-go/internal/roles"

	"gopkg.in/yaml.v3"
)

// ConfigFilePath defines the global path to the configuration file
const ConfigFilePath = "data/config.yaml"

// User represents a user with authentication credentials
type User struct {
	Username string   `yaml:"username" json:"username"`
	Password string   `yaml:"password" json:"password,omitempty"`
	Role     string   `yaml:"role" json:"role"`                       // "admin", "editor", or "viewer"
	Groups   []string `yaml:"groups,omitempty" json:"groups,omitempty"` // Optional groups for access control
}

// AccessRule defines a path-based access control rule
type AccessRule struct {
	Pattern     string   `yaml:"pattern" json:"pattern"`
	Access      string   `yaml:"access" json:"access"` // "public", "private", "restricted"
	Groups      []string `yaml:"groups,omitempty" json:"groups,omitempty"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
}

// Role constants - using the ones defined in roles package
var (
	RoleAdmin  = roles.RoleAdmin  // Can do anything
	RoleEditor = roles.RoleEditor // Can edit documents and post comments
	RoleViewer = roles.RoleViewer // Can only view documents and post comments
)

// Config represents the server configuration
type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
		// When set to true, allows cookies to be sent over non-HTTPS connections.
		// WARNING: Only enable this in trusted environments like a homelab
		// where HTTPS is not available. This reduces security by allowing
		// cookies to be transmitted in plain text.
		AllowInsecureCookies bool `yaml:"allow_insecure_cookies"`
		// Enable native TLS. When true, application will run over HTTPS using the
		// supplied certificate and key paths.
		SSL      bool   `yaml:"ssl"`
		SSLCert  string `yaml:"ssl_cert"`
		SSLKey   string `yaml:"ssl_key"`
	} `yaml:"server"`
	Wiki struct {
		RootDir                   string `yaml:"root_dir"`
		DocumentsDir              string `yaml:"documents_dir"`
		Title                     string `yaml:"title"`
		Owner                     string `yaml:"owner"`
		Notice                    string `yaml:"notice"`
		Timezone                  string `yaml:"timezone"`
		Private                   bool   `yaml:"private"`
		DisableComments           bool   `yaml:"disable_comments"`              // Disable comments system-wide when true
		DisableFileUploadChecking bool   `yaml:"disable_file_upload_checking"`  // Disable mimetype checking for file uploads when true
		EnableLinkEmbedding       bool   `yaml:"enable_link_embedding"`         // Enable automatic link embedding from clipboard when true
		HideAttachments           bool   `yaml:"hide_attachments"`              // Hide attachments section in documents when true
		DisableContentMaxWidth    bool   `yaml:"disable_content_max_width"`     // Disable 900px content width limit when true
		MaxVersions               int    `yaml:"max_versions"`
		MaxUploadSize             int    `yaml:"max_upload_size"` // Maximum upload file size in MB
		Language                  string `yaml:"language"`        // Default language for the wiki
	} `yaml:"wiki"`
	Users       []User       `yaml:"users"`
	AccessRules []AccessRule `yaml:"access_rules,omitempty"`
	Security    struct {
		LoginBan struct {
			Enabled           bool `yaml:"enabled"`
			MaxFailures       int  `yaml:"max_failures"`
			WindowSeconds     int  `yaml:"window_seconds"`
			InitialBanSeconds int  `yaml:"initial_ban_seconds"`
			MaxBanSeconds     int  `yaml:"max_ban_seconds"`
		} `yaml:"login_ban"`
	} `yaml:"security"`
}

// LoadConfig loads the configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	// Set default values
	config := &Config{}
	config.Server.Host = "0.0.0.0" // Set to localhost for local development
	config.Server.Port = 8080
	config.Server.AllowInsecureCookies = false // Default to secure cookies
	config.Server.SSL = false
	config.Server.SSLCert = ""
	config.Server.SSLKey = ""
	config.Wiki.RootDir = "data"
	config.Wiki.DocumentsDir = "documents"
	config.Wiki.Title = "📚 Wiki-Go"
	config.Wiki.Owner = "wiki.example.com"
	config.Wiki.Notice = "Copyright 2025 © All rights reserved."
	config.Wiki.Timezone = "America/Vancouver"
	config.Wiki.Private = false
	config.Wiki.DisableComments = false
	config.Wiki.DisableFileUploadChecking = false // Default to false - always check file uploads
	config.Wiki.EnableLinkEmbedding = false
	config.Wiki.HideAttachments = false
	config.Wiki.DisableContentMaxWidth = false
	config.Wiki.MaxVersions = 10   // Default value
	config.Wiki.MaxUploadSize = 10 // Default value
	config.Wiki.Language = "en"    // Default to English
	config.Users = []User{}        // Initialize empty users array

	// Security defaults
	config.Security.LoginBan.Enabled = true
	config.Security.LoginBan.MaxFailures = 5
	config.Security.LoginBan.WindowSeconds = 180
	config.Security.LoginBan.InitialBanSeconds = 60
	config.Security.LoginBan.MaxBanSeconds = 86400 // 24h

	// Read config file
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Ensure the directory exists
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
			}

			// Hash the default admin password
			hashedPassword, err := crypto.HashPassword("admin")
			if err != nil {
				return nil, err
			}

			// Add default admin user
			config.Users = append(config.Users, User{
				Username: "admin",
				Password: hashedPassword,
				Role:     RoleAdmin,
			})

			// Format all users
			var usersStr strings.Builder
			for _, user := range config.Users {
				if usersStr.Len() > 0 {
					usersStr.WriteString("\n")
				}
				usersStr.WriteString(FormatUserEntry(user))
			}

			// Fill in the template with values from the config
			configData := fmt.Sprintf(
				GetConfigTemplate(),
				config.Server.Host,
				config.Server.Port,
				config.Server.AllowInsecureCookies,
				config.Server.SSL,
				config.Server.SSLCert,
				config.Server.SSLKey,
				config.Wiki.RootDir,
				config.Wiki.DocumentsDir,
				config.Wiki.Title,
				config.Wiki.Owner,
				config.Wiki.Notice,
				config.Wiki.Timezone,
				config.Wiki.Private,
				config.Wiki.DisableComments,
				config.Wiki.DisableFileUploadChecking,
				config.Wiki.EnableLinkEmbedding,
				config.Wiki.HideAttachments,
				config.Wiki.DisableContentMaxWidth,
				config.Wiki.MaxVersions,
				config.Wiki.MaxUploadSize,
				config.Wiki.Language,
				config.Security.LoginBan.Enabled,
				config.Security.LoginBan.MaxFailures,
				config.Security.LoginBan.WindowSeconds,
				config.Security.LoginBan.InitialBanSeconds,
				config.Security.LoginBan.MaxBanSeconds,
				usersStr.String(),
			)

			// Write the config file
			err = os.WriteFile(path, []byte(configData), 0644)
			if err != nil {
				return nil, err
			}
			return config, nil
		}
		return nil, err
	}

	// Parse YAML
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	// Ensure the on-disk configuration includes every setting present in the current template.
	// This will rewrite the file ONLY when new settings have been introduced that are not
	// present in the user's existing config.yaml. The user's current values will be
	// preserved because we first unmarshalled them into the config struct above before
	// regenerating the file via the template.
	if err := ensureCompleteConfig(path, config, data); err != nil {
		return nil, err
	}

	// Migrate user roles from is_admin to role - this is now done in main.go

	return config, nil
}

// GetConfigTemplate returns the template for the config file with comments
func GetConfigTemplate() string {
	return `server:
    host: "%s"
    port: %d
    # When set to true, allows cookies to be sent over non-HTTPS connections.
    # WARNING: Only enable this in trusted environments like a homelab
    # where HTTPS is not available. This reduces security by allowing
    # cookies to be transmitted in plain text.
    allow_insecure_cookies: %t
    # Enable native TLS. When true, application will run over HTTPS using the
    # supplied certificate and key paths.
    ssl: %t
    ssl_cert: "%s"
    ssl_key: "%s"
wiki:
    root_dir: "%s"
    documents_dir: "%s"
    title: "%s"
    owner: "%s"
    notice: "%s"
    timezone: "%s"
    private: %t
    disable_comments: %t
    disable_file_upload_checking: %t
    enable_link_embedding: %t
    hide_attachments: %t
    disable_content_max_width: %t
    max_versions: %d
    # Maximum file upload size in MB
    max_upload_size: %d
    # Default language for the wiki interface (en, es, etc.)
    language: "%s"
security:
    login_ban:
        # Enable protection against brute force login attacks
        enabled: %t
        # Number of failed attempts before triggering a ban
        max_failures: %d
        # Time window in seconds for counting failures
        window_seconds: %d
        # Duration in seconds for the first ban
        initial_ban_seconds: %d
        # Maximum ban duration in seconds (24 hours)
        max_ban_seconds: %d
users:
%s
access_rules:
%s`
}

// FormatUserEntry formats a single user entry for the config file
func FormatUserEntry(user User) string {
	entry := fmt.Sprintf("    - username: %s\n      password: %s\n      role: %s",
		user.Username, user.Password, user.Role)

	if len(user.Groups) > 0 {
		entry += "\n      groups:"
		for _, group := range user.Groups {
			entry += fmt.Sprintf("\n        - %s", group)
		}
	}
	return entry
}

// FormatAccessRuleEntry formats a single access rule entry for the config file
func FormatAccessRuleEntry(rule AccessRule) string {
	entry := fmt.Sprintf("    - pattern: \"%s\"\n      access: %s", rule.Pattern, rule.Access)
	if len(rule.Groups) > 0 {
		entry += fmt.Sprintf("\n      groups: [%s]", strings.Join(rule.Groups, ", "))
	}
	if rule.Description != "" {
		entry += fmt.Sprintf("\n      description: \"%s\"", rule.Description)
	}
	return entry
}

// SaveConfig saves the configuration to a writer
func SaveConfig(cfg *Config, w io.Writer) error {
	// Format all users
	var usersStr strings.Builder
	for _, user := range cfg.Users {
		if usersStr.Len() > 0 {
			usersStr.WriteString("\n")
		}
		usersStr.WriteString(FormatUserEntry(user))
	}

	// Format all access rules
	var accessRulesStr strings.Builder
	for _, rule := range cfg.AccessRules {
		if accessRulesStr.Len() > 0 {
			accessRulesStr.WriteString("\n")
		}
		accessRulesStr.WriteString(FormatAccessRuleEntry(rule))
	}

	// Fill in the template with values from the config
	configData := fmt.Sprintf(
		GetConfigTemplate(),
		cfg.Server.Host,
		cfg.Server.Port,
		cfg.Server.AllowInsecureCookies,
		cfg.Server.SSL,
		cfg.Server.SSLCert,
		cfg.Server.SSLKey,
		cfg.Wiki.RootDir,
		cfg.Wiki.DocumentsDir,
		cfg.Wiki.Title,
		cfg.Wiki.Owner,
		cfg.Wiki.Notice,
		cfg.Wiki.Timezone,
		cfg.Wiki.Private,
		cfg.Wiki.DisableComments,
		cfg.Wiki.DisableFileUploadChecking,
		cfg.Wiki.EnableLinkEmbedding,
		cfg.Wiki.HideAttachments,
		cfg.Wiki.DisableContentMaxWidth,
		cfg.Wiki.MaxVersions,
		cfg.Wiki.MaxUploadSize,
		cfg.Wiki.Language,
		cfg.Security.LoginBan.Enabled,
		cfg.Security.LoginBan.MaxFailures,
		cfg.Security.LoginBan.WindowSeconds,
		cfg.Security.LoginBan.InitialBanSeconds,
		cfg.Security.LoginBan.MaxBanSeconds,
		usersStr.String(),
		accessRulesStr.String(),
	)

	_, err := w.Write([]byte(configData))
	return err
}

// ensureCompleteConfig regenerates the configuration file using the current template and
// writes it back to disk ONLY if the newly rendered file differs from what already exists.
// This means that when new settings are added to the application template, running the app
// once will automatically add them to an existing config.yaml while preserving the user's
// current values for existing settings.
func ensureCompleteConfig(path string, cfg *Config, original []byte) error {
	var buf bytes.Buffer
	if err := SaveConfig(cfg, &buf); err != nil {
		return err
	}

	newData := buf.Bytes()

	// If the generated configuration exactly matches what is already on disk, no update is
	// necessary. This quick equality check avoids unnecessary writes.
	if bytes.Equal(original, newData) {
		return nil
	}

	// Otherwise, overwrite the file with the fully rendered configuration that now contains
	// any newly introduced settings.
	if err := os.WriteFile(path, newData, 0644); err != nil {
		return fmt.Errorf("failed to update config file with new settings: %w", err)
	}

	return nil
}
