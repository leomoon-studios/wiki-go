package routes

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"wiki-go/internal/auth"
	"wiki-go/internal/config"
	"wiki-go/internal/handlers"
	"wiki-go/internal/resources"
)

// CSPMiddleware adds Content Security Policy headers to all responses
func CSPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set a less restrictive Content Security Policy that allows inline scripts and styles
		csp := []string{
			// Default to allowing same-origin resources and inline scripts/styles
			"default-src 'self' 'unsafe-inline' 'unsafe-eval'",
			// Allow inline styles and styles from same origin
			"style-src 'self' 'unsafe-inline'",
			// Allow scripts from same origin and inline scripts
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'",
			// Images from same origin and data: URLs (for embedded images)
			"img-src 'self' data: https://*.ytimg.com https://*.vimeocdn.com",
			// Connect only to same origin
			"connect-src 'self'",
			// Fonts from same origin
			"font-src 'self'",
			// Allow object/embed only from same origin
			"object-src 'self'",
			// Media only from same origin
			"media-src 'self' https://*.youtube.com https://*.vimeo.com",
			// Allow frames from YouTube and Vimeo for video embeds
			"frame-src 'self' https://*.youtube.com https://*.vimeo.com",
			// Form submissions only to same origin
			"form-action 'self'",
			// Base URI restricted to same origin
			"base-uri 'self'",
		}

		// Set the CSP header - using Content-Security-Policy-Report-Only first to avoid breaking things
		// This will report violations but not enforce them
		w.Header().Set("Content-Security-Policy-Report-Only", strings.Join(csp, "; "))

		// Once you've fixed all the violations, you can switch to enforcing mode:
		// w.Header().Set("Content-Security-Policy", strings.Join(csp, "; "))

		// Add other security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

/*
// Example of how to implement nonce-based CSP (for future reference)
func CSPMiddlewareWithNonce(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate a random nonce for this request
		nonce := generateNonce()

		// Store the nonce in the request context so templates can access it
		ctx := context.WithValue(r.Context(), "csp-nonce", nonce)
		r = r.WithContext(ctx)

		// Set a strict Content Security Policy with nonce
		csp := []string{
			// Default to only allowing same-origin resources
			"default-src 'self'",
			// Allow styles from same origin and inline styles with nonce
			fmt.Sprintf("style-src 'self' 'nonce-%s'", nonce),
			// Allow scripts from same origin and inline scripts with nonce
			fmt.Sprintf("script-src 'self' https://cdn.jsdelivr.net 'nonce-%s'", nonce),
			// Images from same origin and data: URLs
			"img-src 'self' data:",
			// Other directives...
		}

		// Set the CSP header
		w.Header().Set("Content-Security-Policy", strings.Join(csp, "; "))

		// Add other security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// Generate a random nonce for CSP
func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
*/

// SetupRoutes configures all routes for the application
func SetupRoutes(cfg *config.Config) {
	// Create a new ServeMux to apply middleware to all routes
	mux := http.NewServeMux()

	// Serve static files with custom handling to check data/static first
	mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		// Extract the file path from the URL
		filename := strings.TrimPrefix(r.URL.Path, "/static/")

		// Special handling for file-extensions.js to inject dynamic content
		if filename == "js/file-extensions.js" {
			w.Header().Set("Content-Type", "application/javascript")
			// Inject the file extensions configuration
			fmt.Fprintf(w, "// File extensions configuration - dynamically generated\n")
			fmt.Fprintf(w, "var ALLOWED_FILE_EXTENSIONS = %s;\n", config.GetAllowedExtensionsJSON())
			fmt.Fprintf(w, "var FILE_EXTENSION_MIME_TYPES = %s;\n", config.GetExtensionMimeTypesJSON())
			return
		}

		// First check if the file exists in data/static
		customPath := filepath.Join(cfg.Wiki.RootDir, "static", filename)
		if _, err := os.Stat(customPath); err == nil {
			// File exists in data/static, serve it directly
			http.ServeFile(w, r, customPath)
			return
		}

		// Fall back to embedded static files
		http.StripPrefix("/static/", http.FileServer(resources.GetFileSystem())).ServeHTTP(w, r)
	})

	// Serve favicons directly from root path
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		customPath := filepath.Join(cfg.Wiki.RootDir, "static", "favicon.ico")
		if _, err := os.Stat(customPath); err == nil {
			http.ServeFile(w, r, customPath)
			return
		}
		http.ServeFile(w, r, filepath.Join("internal", "resources", "static", "favicon.ico"))
	})
	mux.HandleFunc("/favicon.png", func(w http.ResponseWriter, r *http.Request) {
		customPath := filepath.Join(cfg.Wiki.RootDir, "static", "favicon.png")
		if _, err := os.Stat(customPath); err == nil {
			http.ServeFile(w, r, customPath)
			return
		}
		http.ServeFile(w, r, filepath.Join("internal", "resources", "static", "favicon.png"))
	})
	mux.HandleFunc("/favicon.svg", func(w http.ResponseWriter, r *http.Request) {
		customPath := filepath.Join(cfg.Wiki.RootDir, "static", "favicon.svg")
		if _, err := os.Stat(customPath); err == nil {
			http.ServeFile(w, r, customPath)
			return
		}
		http.ServeFile(w, r, filepath.Join("internal", "resources", "static", "favicon.svg"))
	})
	mux.HandleFunc("/logo.png", func(w http.ResponseWriter, r *http.Request) {
		customPath := filepath.Join(cfg.Wiki.RootDir, "static", "logo.png")
		if _, err := os.Stat(customPath); err == nil {
			http.ServeFile(w, r, customPath)
			return
		}
		http.ServeFile(w, r, filepath.Join("internal", "resources", "static", "logo.png"))
	})

	// API Routes
	mux.HandleFunc("/api/login", handlers.LoginHandler)
	mux.HandleFunc("/api/check-auth", handlers.CheckAuthHandler)
	mux.HandleFunc("/api/logout", handlers.LogoutHandler)
	mux.HandleFunc("/api/check-default-password", handlers.CheckDefaultPasswordHandler)
	mux.HandleFunc("/api/document/create", handlers.CreateDocumentHandler)
	mux.HandleFunc("/api/document/", handlers.DocumentHandler)
	mux.HandleFunc("/api/source/", handlers.SourceHandler)
	mux.HandleFunc("/api/save/", handlers.SaveHandler)

	// File API Routes
	mux.HandleFunc("/api/files/upload", func(w http.ResponseWriter, r *http.Request) {
		handlers.UploadFileHandler(w, r, cfg)
	})

	mux.HandleFunc("/api/files/list/", func(w http.ResponseWriter, r *http.Request) {
		handlers.ListFilesHandler(w, r, cfg)
	})

	mux.HandleFunc("/api/files/delete/", func(w http.ResponseWriter, r *http.Request) {
		handlers.DeleteFileHandler(w, r, cfg)
	})

	mux.HandleFunc("/api/files/", func(w http.ResponseWriter, r *http.Request) {
		handlers.ServeFileHandler(w, r, cfg)
	})

	// Search handler with wrapper to include config
	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		handlers.SearchHandler(w, r, cfg)
	})

	// Settings API
	mux.HandleFunc("/api/settings/wiki", handlers.WikiSettingsHandler)

	// User Management API
	mux.HandleFunc("/api/users", handlers.UsersHandler)

	// Version history API
	mux.HandleFunc("/api/versions/", func(w http.ResponseWriter, r *http.Request) {
		handlers.VersionsHandler(w, r, cfg)
	})

	// Document move/rename API
	mux.HandleFunc("/api/document/move", func(w http.ResponseWriter, r *http.Request) {
		handlers.MoveDocumentHandler(w, r, cfg)
	})

	// Markdown rendering API
	mux.HandleFunc("/api/render-markdown", handlers.RenderMarkdownHandler)

	// Login page
	mux.HandleFunc("/login", handlers.LoginPageHandler)

	// Home page and other pages
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check if authentication is required
		if !auth.RequireAuth(r, cfg) {
			// If private and not authenticated, redirect to login page
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// If the URL path is just /, serve the home page
		if r.URL.Path == "/" {
			handlers.HomeHandler(w, r, cfg)
			return
		}

		// Otherwise, serve the page based on the URL path
		handlers.PageHandler(w, r, cfg)
	})

	// Apply middleware to all routes
	handler := CSPMiddleware(mux)

	// Set the handler for the default ServeMux
	http.Handle("/", handler)
}
