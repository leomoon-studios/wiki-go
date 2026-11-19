package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"wiki-go/internal/auth"
	"wiki-go/internal/basepath"
	"wiki-go/internal/config"
	"wiki-go/internal/goldext"
	"wiki-go/internal/handlers"
	"wiki-go/internal/logger"
	"wiki-go/internal/migration"
	"wiki-go/internal/routes"
	"wiki-go/internal/static"
)

func main() {
	configfilepath := flag.String("configfile",
		GetEnvString("CONFIGFILE", config.ConfigFilePath),
		"where to find config.yaml")
	flag.Parse()

	config.ConfigFilePath = *configfilepath

	// Fix broken config file if it exists (from previous bug)
	if err := migration.FixBrokenConfig(config.ConfigFilePath); err != nil {
		logger.Warn("Failed to fix broken config: %v", err)
	}

	// Migrate user roles from old IsAdmin to new role-based system
	if err := migration.MigrateUserRoles(config.ConfigFilePath); err != nil {
		logger.Fatal("Error migrating user roles: %v", err)
	}

	// Load configuration (after migration)
	cfg, err := config.LoadConfig(config.ConfigFilePath)
	if err != nil {
		logger.Fatal("Error loading config: %v", err)
	}

	// Serve the wiki from server.base_path (for example /wiki) when configured.
	basepath.Set(cfg.Server.BasePath)

	// Propagate wiki timezone to the goldext shortcode renderer so that
	// :::stats recent=N::: formats edit times in the configured zone
	// (consistent with formatTime used elsewhere in templates).
	goldext.SetWikiTimezone(cfg.Wiki.Timezone)
	goldext.ConfigureContentPaths(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir)

	// Apply log level from config
	logger.Init(cfg.Wiki.LogLevel)

	// Initialize session store for persistent logins
	sessionPath := filepath.Join(cfg.Wiki.RootDir, "temp", "sessions.json")
	if err := auth.InitSessionStore(sessionPath); err != nil {
		logger.Warn("Failed to initialize session store: %v", err)
	}

	// Ensure the homepage exists
	if err := handlers.EnsureHomepageExists(cfg); err != nil {
		logger.Fatal("Error creating homepage: %v", err)
	}

	// Ensure static assets exist in data directory
	if err := static.EnsureStaticAssetsExist(cfg.Wiki.RootDir); err != nil {
		logger.Fatal("Error copying static assets: %v", err)
	}

	// Update handlers with config
	handlers.InitHandlers(cfg)

	// Setup all routes
	routes.SetupRoutes(cfg)

	// Start the server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	if cfg.Server.BasePath != "" {
		logger.Info("Serving wiki under base path %s/", cfg.Server.BasePath)
	}
	if cfg.Server.SSL && cfg.Server.SSLCert != "" && cfg.Server.SSLKey != "" {
		logger.Info("HTTPS server starting on %s...", addr)
		if err := http.ListenAndServeTLS(addr, cfg.Server.SSLCert, cfg.Server.SSLKey, nil); err != nil {
			logger.Fatal("HTTPS server error: %v", err)
		}
	} else {
		logger.Info("HTTP server starting on %s...", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			logger.Fatal("HTTP server error: %v", err)
		}
	}
}

func GetEnvString(name, defaultvalue string) string {
	value, ok := os.LookupEnv(name)
	if !ok {
		return defaultvalue
	}
	return value
}
