package handlers

import (
	"log"
	"wiki-go/internal/config"
	"wiki-go/internal/i18n"
)

var cfg *config.Config

// InitHandlers initializes the handlers with the given configuration
func InitHandlers(config *config.Config) {
	cfg = config

	// Initialize i18n package
	if err := i18n.Initialize(cfg); err != nil {
		log.Printf("Warning: Failed to initialize i18n package: %v", err)
	}

	// Routes are now managed in the routes package
}

// We don't need this anymore since main.go handles the routing
// func handlePage(w http.ResponseWriter, r *http.Request) {
//     PageHandler(w, r, cfg)
// }
