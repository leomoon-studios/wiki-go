package handlers

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"wiki-go/internal/auth"
	"wiki-go/internal/config"
	"wiki-go/internal/resources"
)

// XML sitemap types
type SitemapURL struct {
	XMLName    xml.Name `xml:"url"`
	Location   string   `xml:"loc"`
	LastMod    string   `xml:"lastmod,omitempty"`
	ChangeFreq string   `xml:"changefreq,omitempty"`
	Priority   string   `xml:"priority,omitempty"`
}

type Sitemap struct {
	XMLName xml.Name    `xml:"urlset"`
	XMLNS   string      `xml:"xmlns,attr"`
	URLs    []SitemapURL `xml:"url"`
}

// Page data for HTML sitemap
type SitemapPage struct {
	Title      string
	Config     *config.Config
	BaseURL    string
	Pages      []SitemapPageEntry
	Categories map[string][]SitemapPageEntry
	UserRole   string
}

type SitemapPageEntry struct {
	URL      string
	Title    string
	Path     string
	Category string
	LastMod  time.Time
	Depth    int
}

// SitemapHandler handles requests for both XML and HTML sitemaps
func SitemapHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// Get current session for access control filtering
	session := auth.CheckAuth(r)

	// Determine if XML format is requested
	isXML := strings.HasSuffix(r.URL.Path, ".xml")

	// Get base URL for the sitemap
	baseURL := getBaseURL(r, cfg)

	// Get current user role for conditional display in HTML sitemap
	userRole := ""
	if session != nil {
		userRole = session.Role
	}

	// Gather all pages (filtered by access rules)
	urls, pageEntries, err := gatherPages(baseURL, cfg, session)
	if err != nil {
		http.Error(w, "Error generating sitemap: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if isXML {
		renderXMLSitemap(w, urls)
	} else {
		renderHTMLSitemap(w, r, pageEntries, baseURL, cfg, userRole)
	}
}

// renderXMLSitemap renders the sitemap in XML format for search engines
func renderXMLSitemap(w http.ResponseWriter, urls []SitemapURL) {
	// Create the sitemap XML
	sitemap := Sitemap{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}

	// Set correct content type for XML
	w.Header().Set("Content-Type", "application/xml; charset=UTF-8")

	// Add cache control headers - sitemaps can be cached for a day
	w.Header().Set("Cache-Control", "public, max-age=86400")

	// Write XML header
	w.Write([]byte(xml.Header))

	// Encode the sitemap
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(sitemap); err != nil {
		http.Error(w, "Error encoding sitemap", http.StatusInternalServerError)
		return
	}
}

// renderHTMLSitemap renders a user-friendly HTML sitemap
func renderHTMLSitemap(w http.ResponseWriter, r *http.Request, pages []SitemapPageEntry, baseURL string, cfg *config.Config, userRole string) {
	// Organize pages by category
	categories := make(map[string][]SitemapPageEntry)

	// Add the home page to a special category
	homeCategory := "Home"
	categories[homeCategory] = []SitemapPageEntry{}

	// Find the home page in the entries and add it to the home category
	for _, page := range pages {
		if page.Path == "/" {
			categories[homeCategory] = append(categories[homeCategory], page)
			break
		}
	}

	// Organize other pages by their top-level category
	for _, page := range pages {
		if page.Path == "/" {
			continue // Skip home page as we already added it
		}

		if page.Category != "" {
			if _, exists := categories[page.Category]; !exists {
				categories[page.Category] = []SitemapPageEntry{}
			}
			categories[page.Category] = append(categories[page.Category], page)
		}
	}

	// Create the sitemap page data
	sitemapData := SitemapPage{
		Title:      fmt.Sprintf("Sitemap - %s", cfg.Wiki.Title),
		Config:     cfg,
		BaseURL:    baseURL,
		Pages:      pages,
		Categories: categories,
		UserRole:   userRole,
	}

	// Set content type
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")

	// Parse and execute the template using the embedded filesystem
	tmpl, err := template.ParseFS(resources.GetTemplatesFS(), "templates/sitemap.html")
	if err != nil {
		http.Error(w, "Error parsing sitemap template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, sitemapData); err != nil {
		http.Error(w, "Error rendering sitemap template: "+err.Error(), http.StatusInternalServerError)
	}
}

// gatherPages collects all pages for the sitemap, filtered by access rules
func gatherPages(baseURL string, cfg *config.Config, session *auth.Session) ([]SitemapURL, []SitemapPageEntry, error) {
	urls := []SitemapURL{}
	pageEntries := []SitemapPageEntry{}

	// Add homepage only if user has access
	if auth.CanAccessDocument("/", session, cfg) {
		homeURL := SitemapURL{
			Location:   baseURL,
			ChangeFreq: "weekly",
			Priority:   "1.0",
		}
		urls = append(urls, homeURL)

		// Add homepage to page entries
		homePage := SitemapPageEntry{
			URL:      "/",
			Title:    "Home",
			Path:     "/",
			Category: "",
			LastMod:  time.Now(),
			Depth:    0,
		}
		pageEntries = append(pageEntries, homePage)
	}

	// Walk the documents directory
	docsDir := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir)
	err := filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only include document.md files
		if !info.IsDir() && filepath.Base(path) == "document.md" {
			// Get the directory that contains this document.md
			dirPath := filepath.Dir(path)

			// Get path relative to documents directory - this is the URL path
			relDirPath, err := filepath.Rel(docsDir, dirPath)
			if err != nil {
				return err
			}

			// Convert path separators to forward slashes
			relDirPath = filepath.ToSlash(relDirPath)

			// Skip hidden directories
			if strings.HasPrefix(filepath.Base(relDirPath), ".") {
				return nil
			}

			// Add to XML sitemap URLs
			urlPath := "/" + relDirPath
			if urlPath == "//" {
				urlPath = "/"
			}

			// Check access rules - skip documents the user can't access
			if !auth.CanAccessDocument(urlPath, session, cfg) {
				return nil
			}

			// Format last modified time for XML
			lastModStr := info.ModTime().Format(time.RFC3339)

			url := SitemapURL{
				Location:   baseURL + urlPath,
				LastMod:    lastModStr,
				ChangeFreq: "monthly",
				Priority:   "0.8",
			}
			urls = append(urls, url)

			// Get document title from document.md
			title := getDocumentTitle(path)
			if title == "" {
				title = filepath.Base(relDirPath)
			}

			// Determine category from path
			category := ""
			pathParts := strings.Split(relDirPath, "/")
			if len(pathParts) > 0 && pathParts[0] != "" {
				category = pathParts[0]
			}

			// Add to HTML sitemap entries
			pageEntry := SitemapPageEntry{
				URL:      urlPath,
				Title:    title,
				Path:     urlPath,
				Category: category,
				LastMod:  info.ModTime(),
				Depth:    len(pathParts) - 1,
			}
			pageEntries = append(pageEntries, pageEntry)
		}

		return nil
	})

	return urls, pageEntries, err
}

// getDocumentTitle extracts the title from a markdown file's first heading
func getDocumentTitle(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}

	// Look for the first heading in the file
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}

	return ""
}

// getBaseURL constructs the base URL from request and config
func getBaseURL(r *http.Request, cfg *config.Config) string {
	scheme := "http"
	if cfg.Server.SSL || r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	host := r.Host
	if host == "" {
		// Fallback to configuration
		host = cfg.Server.Host
		if cfg.Server.Port != 80 && cfg.Server.Port != 443 {
			host = fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		}
	}

	return fmt.Sprintf("%s://%s", scheme, host)
}