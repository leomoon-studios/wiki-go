package routes

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"wiki-go/internal/auth"
	"wiki-go/internal/config"
	"wiki-go/internal/handlers"
	"wiki-go/internal/resources"
)

// addCacheControlHeaders adds appropriate Cache-Control headers based on file type
func addCacheControlHeaders(w http.ResponseWriter, filename string) {
	ext := filepath.Ext(filename)

	// Check if filename contains a version identifier (hash or version number)
	isVersioned := strings.Contains(filename, ".v") ||
	               strings.Contains(filename, ".min.") ||
	               strings.Contains(filename, "-bundle")

	// 1 year for versioned resources
	if isVersioned {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		return
	}

	// Set specific cache times based on file type
	switch ext {
	case ".css", ".js", ".woff", ".woff2", ".ttf", ".otf":
		// 1 week for stylesheets, scripts, and fonts
		w.Header().Set("Cache-Control", "public, max-age=604800")

	case ".png", ".jpg", ".jpeg", ".gif", ".ico", ".svg", ".webp":
		// 2 weeks for images
		w.Header().Set("Cache-Control", "public, max-age=1209600")

	default:
		// 1 day for other static files
		w.Header().Set("Cache-Control", "public, max-age=86400")
	}
}

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

		// Add preload header for emojis.json to avoid AJAX loading
		// This works by telling the browser to preload this resource before it's needed
		if strings.HasSuffix(r.URL.Path, ".html") || r.URL.Path == "/" || strings.HasSuffix(r.URL.Path, "/") {
			w.Header().Add("Link", "</api/data/emojis>; rel=preload; as=fetch; crossorigin=anonymous")
		}

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

// Helper function to check if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Helper function to check if any custom favicon exists
func anyCustomFaviconExists(rootDir string) bool {
	for _, format := range []string{"favicon.ico", "favicon.png", "favicon.svg"} {
		if fileExists(filepath.Join(rootDir, "static", format)) {
			return true
		}
	}
	return false
}

// Helper function to handle favicon requests with proper fallback logic
func handleFaviconRequest(w http.ResponseWriter, r *http.Request, cfg *config.Config, format string) {
	// Add cache headers for favicon (1 week)
	w.Header().Set("Cache-Control", "public, max-age=604800")

	// Check if this specific favicon exists in custom path
	customPath := filepath.Join(cfg.Wiki.RootDir, "static", "favicon."+format)
	if fileExists(customPath) {
		http.ServeFile(w, r, customPath)
		return
	}

	// If any custom favicon exists but not this one, return 404
	if anyCustomFaviconExists(cfg.Wiki.RootDir) {
		http.NotFound(w, r)
		return
	}

	// Fallback to embedded favicon only if no custom favicons exist
	http.ServeFile(w, r, filepath.Join("internal", "resources", "static", "favicon."+format))
}

// SetupRoutes configures all routes for the application
func SetupRoutes(cfg *config.Config) {
	// Create a new ServeMux to apply middleware to all routes
	mux := http.NewServeMux()

	// Role-based middleware
	adminMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !auth.RequireRole(r, "admin") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"message": "Admin access required",
				})
				return
			}
			next(w, r)
		}
	}

	editorMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !auth.RequireRole(r, "editor") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"message": "Unauthorized. Admin or editor access required.",
				})
				return
			}
			next(w, r)
		}
	}

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

		// Add appropriate cache headers for static files
		addCacheControlHeaders(w, filename)

		// Check if this is a favicon file
		isFaviconFile := false
		for _, ext := range []string{"ico", "png", "svg"} {
			if filename == "favicon."+ext {
				isFaviconFile = true
				break
			}
		}

		// First check if the file exists in data/static
		customPath := filepath.Join(cfg.Wiki.RootDir, "static", filename)
		if fileExists(customPath) {
			// File exists in data/static, serve it directly
			http.ServeFile(w, r, customPath)
			return
		}

		// For favicon files, check if any custom favicon exists before falling back
		if isFaviconFile && anyCustomFaviconExists(cfg.Wiki.RootDir) {
			// If any custom favicon exists but not this one, return 404
			http.NotFound(w, r)
			return
		}

		// Fall back to embedded static files
		http.StripPrefix("/static/", http.FileServer(resources.GetFileSystem())).ServeHTTP(w, r)
	})

	// Serve favicons directly from root path
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		handleFaviconRequest(w, r, cfg, "ico")
	})

	mux.HandleFunc("/favicon.png", func(w http.ResponseWriter, r *http.Request) {
		handleFaviconRequest(w, r, cfg, "png")
	})

	mux.HandleFunc("/favicon.svg", func(w http.ResponseWriter, r *http.Request) {
		handleFaviconRequest(w, r, cfg, "svg")
	})

	mux.HandleFunc("/logo.png", func(w http.ResponseWriter, r *http.Request) {
		// Add cache headers for logo (1 week)
		w.Header().Set("Cache-Control", "public, max-age=604800")

		customPath := filepath.Join(cfg.Wiki.RootDir, "static", "logo.png")
		if _, err := os.Stat(customPath); err == nil {
			http.ServeFile(w, r, customPath)
			return
		}
		http.ServeFile(w, r, filepath.Join("internal", "resources", "static", "logo.png"))
	})

	mux.HandleFunc("/manifest.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/manifest+json")
		w.Header().Set("Cache-Control", "public, max-age=604800")

		// Check if custom manifest exists
		customPath := filepath.Join(cfg.Wiki.RootDir, "static", "manifest.json")
		if fileExists(customPath) {
			http.ServeFile(w, r, customPath)
			return
		}

		// Read the embedded manifest
		dataFS := resources.GetFileSystem()
		file, err := dataFS.Open("manifest.json")
		if err != nil {
			http.Error(w, "Manifest not found", http.StatusNotFound)
			return
		}
		defer file.Close()

		// Decode the JSON
		var manifest map[string]interface{}
		if err := json.NewDecoder(file).Decode(&manifest); err != nil {
			http.Error(w, "Invalid manifest", http.StatusInternalServerError)
			return
		}

		// Update name and short_name from config
		manifest["name"] = cfg.Wiki.Title
		manifest["short_name"] = cfg.Wiki.Title

		// Encode back to JSON
		if err := json.NewEncoder(w).Encode(manifest); err != nil {
			http.Error(w, "Error encoding manifest", http.StatusInternalServerError)
			return
		}
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

	mux.HandleFunc("/api/files/rename", func(w http.ResponseWriter, r *http.Request) {
		handlers.RenameFileHandler(w, r, cfg)
	})

	mux.HandleFunc("/api/files/", func(w http.ResponseWriter, r *http.Request) {
		handlers.ServeFileHandler(w, r, cfg)
	})

	// Comment API Routes
	mux.HandleFunc("/api/comments/add/", handlers.AddCommentHandler)
	mux.HandleFunc("/api/comments/delete/", handlers.DeleteCommentHandler)
	mux.HandleFunc("/api/comments/", handlers.GetCommentsHandler)

	// Search handler with wrapper to include config
	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		handlers.SearchHandler(w, r, cfg)
	})

	// Settings API - Admin only
	mux.HandleFunc("/api/settings/wiki", adminMiddleware(handlers.WikiSettingsHandler))
	mux.HandleFunc("/api/settings/security", adminMiddleware(handlers.SecuritySettingsHandler))

	// User Management API - Admin only
	mux.HandleFunc("/api/users", adminMiddleware(handlers.UsersHandler))

	// Access Rules API - Admin only
	mux.HandleFunc("/api/access-rules", adminMiddleware(handlers.AccessRulesHandler))
	mux.HandleFunc("/api/access-rules/", adminMiddleware(handlers.AccessRulesHandler))
	mux.HandleFunc("/api/folders", adminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		handlers.ListFoldersHandler(w, r, cfg)
	}))

	// Version history API - Editor or Admin
	mux.HandleFunc("/api/versions/", editorMiddleware(func(w http.ResponseWriter, r *http.Request) {
		handlers.VersionsHandler(w, r, cfg)
	}))

	// Document move/rename API - Editor or Admin
	mux.HandleFunc("/api/document/move", editorMiddleware(func(w http.ResponseWriter, r *http.Request) {
		handlers.MoveDocumentHandler(w, r, cfg)
	}))

	// Markdown rendering API - No auth required
	mux.HandleFunc("/api/render-markdown", handlers.RenderMarkdownHandler)

	// Emoji data API - serve the emojis.json file
	mux.HandleFunc("/api/data/emojis", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Add cache headers for emoji data (1 day)
		w.Header().Set("Cache-Control", "public, max-age=86400")
		dataFS := resources.GetDataFS()
		data, err := fs.ReadFile(dataFS, "emojis.json")
		if err != nil {
			http.Error(w, "Error reading emoji data", http.StatusInternalServerError)
			return
		}
		w.Write(data)
	})

	// Documents list API - for document linking
	mux.HandleFunc("/api/documents/list", func(w http.ResponseWriter, r *http.Request) {
		handlers.ListDocumentsHandler(w, r, cfg)
	})

	// Import API - Admin only
	mux.HandleFunc("/api/import", func(w http.ResponseWriter, r *http.Request) {
		handlers.ImportHandler(w, r, cfg)
	})

	mux.HandleFunc("/api/import/status/", func(w http.ResponseWriter, r *http.Request) {
		handlers.ImportStatusHandler(w, r, cfg)
	})

	// Sitemap routes
	mux.HandleFunc("/sitemap/", func(w http.ResponseWriter, r *http.Request) {
		handlers.SitemapHandler(w, r, cfg)
	})

	mux.HandleFunc("/sitemap.xml", func(w http.ResponseWriter, r *http.Request) {
		handlers.SitemapHandler(w, r, cfg)
	})

	// Utility API endpoints
	mux.HandleFunc("/api/utils/slugify", handlers.SlugifyHandler)

	// Links Metadata API - Editor or Admin only
	mux.HandleFunc("/api/links/fetch-metadata", editorMiddleware(handlers.FetchMetadataHandler))

	// Login page
	mux.HandleFunc("/login", handlers.LoginPageHandler)

	// Home page and other pages
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check if authentication is required
		if !auth.RequireAuth(r, cfg) {
			// If private and not authenticated, redirect to login page
			// Preserve original requested path so that we can jump back after login.
			target := r.URL.RequestURI()
			if target == "" {
				target = "/"
			}
			q := url.QueryEscape(target)
			http.Redirect(w, r, "/login?redirect="+q, http.StatusSeeOther)
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
