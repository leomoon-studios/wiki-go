package handlers

import (
	"io"
	"net/http"
	"wiki-go/internal/auth"
	"wiki-go/internal/utils"
)

// RenderMarkdownHandler handles requests to render markdown to HTML
// This endpoint is used for client-side previewing to ensure consistent rendering
func RenderMarkdownHandler(w http.ResponseWriter, r *http.Request) {
	// Add cache control headers to prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	session := auth.GetSession(r)
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Read the markdown content from the request body
	markdown, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Get the document path from the query parameter
	docPath := r.URL.Query().Get("path")

	// Use the utility function to render markdown to HTML with the document path
	var html []byte
	if docPath != "" {
		html = utils.RenderMarkdownWithPath(string(markdown), docPath)
	} else {
		html = utils.RenderMarkdown(string(markdown))
	}

	// Set content type to HTML
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Write the rendered HTML to the response
	w.Write(html)
}
