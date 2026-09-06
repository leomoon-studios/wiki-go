package migration

import (
	"fmt"
	"io"
	"os"
	"time"
	"wiki-go/internal/logger"
	"wiki-go/internal/roles"

	"gopkg.in/yaml.v3"
)

// User represents a user with authentication credentials (copied from config to avoid import cycle)
type User struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Role     string `yaml:"role"`               // "admin", "editor", or "viewer"
	IsAdmin  bool   `yaml:"is_admin,omitempty"` // Old field for migration
}

// Config represents a simplified version of the server configuration (copied from config to avoid import cycle)
type Config struct {
	Server struct {
		Host                 string   `yaml:"host"`
		Port                 int      `yaml:"port"`
		AllowInsecureCookies bool     `yaml:"allow_insecure_cookies"`
		TrustedProxies       []string `yaml:"trusted_proxies,omitempty"`
		SSL                  bool     `yaml:"ssl"`
		SSLCert              string   `yaml:"ssl_cert"`
		SSLKey               string   `yaml:"ssl_key"`
	} `yaml:"server"`
	Wiki struct {
		RootDir                   string `yaml:"root_dir"`
		DocumentsDir              string `yaml:"documents_dir"`
		Title                     string `yaml:"title"`
		Owner                     string `yaml:"owner"`
		Notice                    string `yaml:"notice"`
		Timezone                  string `yaml:"timezone"`
		Private                   bool   `yaml:"private"`
		DisableComments           bool   `yaml:"disable_comments"`
		DisableFileUploadChecking bool   `yaml:"disable_file_upload_checking"`
		MaxVersions               int    `yaml:"max_versions"`
		MaxUploadSize             int    `yaml:"max_upload_size"`
		Language                  string `yaml:"language"`
	} `yaml:"wiki"`
	Users []User `yaml:"users"`
}

// copyFile creates a copy of src file at dst path
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

// MigrateUserRoles updates the user configuration from the old IsAdmin boolean
// to the new role-based system. This function should be called during application
// startup to ensure all users have a proper role assigned.
func MigrateUserRoles(configPath string) error {
	logger.Debug("Checking if user role migration is needed...")

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file does not exist yet – nothing to migrate. It will be
			// created with defaults when the main application loads the config.
			logger.Info("Config file %s not found, skipping user-role migration", configPath)
			return nil
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the config
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	migrationNeeded := false

	// Check if any users need migration (empty Role or has IsAdmin field)
	for _, user := range cfg.Users {
		if user.Role == "" || user.IsAdmin {
			migrationNeeded = true
			break
		}
	}

	if !migrationNeeded {
		logger.Debug("No role migration needed, all users have roles assigned")
		return nil
	}

	// Create backup of config file before migration (only if migration is needed)
	backupPath := configPath + "." + time.Now().Format("20060102-150405") + ".bak"
	if err := copyFile(configPath, backupPath); err != nil {
		return fmt.Errorf("failed to create config backup: %w", err)
	}
	logger.Info("Created config backup at: %s", backupPath)

	logger.Info("Migrating users from IsAdmin to role-based system...")

	// Create a new users slice to hold the migrated users
	migratedUsers := make([]User, 0, len(cfg.Users))

	// Migrate each user
	for _, user := range cfg.Users {
		// If role is already set and no IsAdmin field, keep it
		if user.Role != "" && !user.IsAdmin {
			migratedUsers = append(migratedUsers, user)
			continue
		}

		// Set role based on IsAdmin field if present
		if user.IsAdmin {
			user.Role = roles.RoleAdmin
		} else {
			user.Role = roles.RoleViewer
		}

		// Clear the IsAdmin field after migration
		user.IsAdmin = false

		migratedUsers = append(migratedUsers, user)
		logger.Info("Migrated user %s: Role=%s", user.Username, user.Role)
	}

	// Update the config with migrated users
	cfg.Users = migratedUsers

	// Save the updated config
	updatedData, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal updated config: %w", err)
	}

	if err := os.WriteFile(configPath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to save migrated config: %w", err)
	}

	logger.Info("User role migration completed successfully")
	return nil
}
