package handlers

import (
	"bytes"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"wiki-go/internal/i18n"
	"wiki-go/internal/resources"
	"wiki-go/internal/types"
	"wiki-go/internal/utils"
	"wiki-go/internal/version"
)

// renderTemplate renders the base template with the given data
func renderTemplate(w http.ResponseWriter, data *types.PageData) {
	// Get the template from cache or load it
	tmpl, err := getTemplate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Execute template into buffer
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write the rendered HTML to the response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

// Cache for the parsed template
var templateCache *template.Template
var templateOnce sync.Once

// getTemplate returns the cached template or loads it if not cached
func getTemplate() (*template.Template, error) {
	var templateErr error

	templateOnce.Do(func() {
		// Create a function map with our timezone formatter
		funcMap := template.FuncMap{
			"formatTime": func(t time.Time, timezone string, format string) string {
				return utils.FormatTimeInTimezone(t, timezone, format)
			},
			"getVersion": func() string {
				return version.Version
			},
			"hasFavicon": func(rootDir string, extension string) bool {
				// Check if a specific favicon exists
				path := filepath.Join(rootDir, "static", "favicon."+extension)
				_, err := os.Stat(path)
				return err == nil
			},
			"hasLogo": func(rootDir string) string {
				// Check for logo.svg
				svgPath := filepath.Join(rootDir, "static", "logo.svg")
				if _, err := os.Stat(svgPath); err == nil {
					return "/static/logo.svg"
				}

				// Check for logo.png
				pngPath := filepath.Join(rootDir, "static", "logo.png")
				if _, err := os.Stat(pngPath); err == nil {
					return "/static/logo.png"
				}

				// No logo found
				return ""
			},
			"hasBanner": func(rootDir string) string {
				// Check for banner.png
				pngPath := filepath.Join(rootDir, "static", "banner.png")
				if _, err := os.Stat(pngPath); err == nil {
					return "/static/banner.png"
				}

				// Check for banner.jpg
				jpgPath := filepath.Join(rootDir, "static", "banner.jpg")
				if _, err := os.Stat(jpgPath); err == nil {
					return "/static/banner.jpg"
				}

				// No banner found
				return ""
			},
			"t": func(key string, params ...interface{}) string {
				// Check if we have a language override as the second parameter
				if len(params) > 0 {
					if lang, ok := params[0].(string); ok {
						// Translate using the i18n package with language override
						return i18n.Translate(key, lang)
					}
				}
				// Regular translation without language override
				return i18n.Translate(key)
			},
			"processShortcodes": func(text string) string {
				// Process shortcodes like :::year::: in text
				// This is a simplified version for non-markdown contexts (e.g., notice)
				currentYear := strconv.Itoa(time.Now().Year())
				return strings.ReplaceAll(text, ":::year:::", currentYear)
			},
		}

		// Load template from embedded resources with our function map
		templateCache, templateErr = resources.LoadTemplates(funcMap)
	})

	return templateCache, templateErr
}