package frontmatter

import (
	"fmt"
	"html/template"
	"net/url"
	urlPkg "net/url"
	"regexp"
	"strings"
	"time"

	"wiki-go/internal/i18n"
)

// Link represents a single link in a links document
type Link struct {
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	AddedAt     time.Time `json:"added_at"`
}

// LinksData represents the complete collection of links organized by category
type LinksData struct {
	Title      string            `json:"title"`       // Document title (H1)
	Categories map[string][]Link `json:"categories"`  // Links organized by category
	TotalLinks int               `json:"total_links"` // Total number of links
	Stats      LinksStats        `json:"stats"`       // Statistics for the links collection
}

// LinksStats provides statistics about the links collection
type LinksStats struct {
	TotalLinks      int       `json:"total_links"`
	TotalCategories int       `json:"total_categories"`
	RecentLinks     int       `json:"recent_links"` // Links added in the last 30 days
	LatestAdded     time.Time `json:"latest_added"` // Most recent addition date
}

type linksLabels struct {
	NoResultsTitle   string
	NoResultsMessage string
	AddNewLink       string
}

type linksTemplateData struct {
	*LinksData
	Labels linksLabels
}

// NewLinksData creates a new LinksData instance with initialized maps
func NewLinksData() *LinksData {
	return &LinksData{
		Categories: make(map[string][]Link),
		Stats:      LinksStats{},
	}
}

// AddLink adds a link to the specified category
func (ld *LinksData) AddLink(link Link) {
	if ld.Categories == nil {
		ld.Categories = make(map[string][]Link)
	}

	ld.Categories[link.Category] = append(ld.Categories[link.Category], link)
	ld.updateStats()
}

// updateStats recalculates the statistics for the links collection
func (ld *LinksData) updateStats() {
	totalLinks := 0
	recentLinks := 0
	oneDayAgo := time.Now().AddDate(0, 0, -1)

	for _, links := range ld.Categories {
		totalLinks += len(links)
		for _, link := range links {
			if link.AddedAt.After(oneDayAgo) {
				recentLinks++
			}
		}
	}

	ld.Stats = LinksStats{
		TotalLinks:      totalLinks,
		TotalCategories: len(ld.Categories),
		RecentLinks:     recentLinks,
	}
	ld.TotalLinks = totalLinks
}

// ValidateURL checks if a URL is valid and well-formed
func ValidateURL(rawURL string) error {
	if strings.TrimSpace(rawURL) == "" {
		return &LinkValidationError{"URL cannot be empty"}
	}

	// Parse the URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return &LinkValidationError{"Invalid URL format: " + err.Error()}
	}

	// Check if scheme is present and valid
	if parsedURL.Scheme == "" {
		return &LinkValidationError{"URL must include a scheme (http:// or https://)"}
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return &LinkValidationError{"URL scheme must be http or https"}
	}

	// Check if host is present
	if parsedURL.Host == "" {
		return &LinkValidationError{"URL must include a valid host"}
	}

	return nil
}

// ValidateLink performs comprehensive validation on a Link struct
func ValidateLink(link Link) []error {
	var errors []error

	// Validate title
	if strings.TrimSpace(link.Title) == "" {
		errors = append(errors, &LinkValidationError{"Title cannot be empty"})
	}

	// Validate URL
	if err := ValidateURL(link.URL); err != nil {
		errors = append(errors, err)
	}

	// Validate category
	if strings.TrimSpace(link.Category) == "" {
		errors = append(errors, &LinkValidationError{"Category cannot be empty"})
	}

	// Description is optional, so no validation needed

	return errors
}

// ParseLinkDate parses a date string in YYYY-MM-DD format for links
func ParseLinkDate(dateStr string) (time.Time, error) {
	// Try YYYY-MM-DD format first
	if t, err := time.Parse("2006-01-02", dateStr); err == nil {
		return t, nil
	}

	// Try other common formats as fallback
	formats := []string{
		"2006/01/02",
		"01/02/2006",
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// ParseDateFromString parses a date string and returns a time.Time value
// Returns zero time if parsing fails or string is empty
func ParseDateFromString(dateStr string) time.Time {
	if strings.TrimSpace(dateStr) == "" {
		return time.Time{}
	}

	// Try to parse in YYYY-MM-DD format
	if t, err := time.Parse("2006-01-02", strings.TrimSpace(dateStr)); err == nil {
		return t
	}

	// If parsing fails, return zero time
	return time.Time{}
}

// FormatDateForDisplay formats a time.Time for display in the links interface
func FormatDateForDisplay(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

// SanitizeCategory sanitizes a category name for safe usage
func SanitizeCategory(category string) string {
	// Remove leading/trailing whitespace
	category = strings.TrimSpace(category)

	// Replace multiple consecutive spaces with single space
	category = regexp.MustCompile(`\s+`).ReplaceAllString(category, " ")

	// If empty after sanitization, return default
	if category == "" {
		return "General"
	}

	return category
}

// ParseLinksContent parses markdown content to extract structured links data
func ParseLinksContent(content string) (*LinksData, error) {
	data := &LinksData{
		Categories: make(map[string][]Link),
		Stats:      LinksStats{},
	}

	lines := strings.Split(content, "\n")
	currentCategory := "General"

	// Regular expressions for parsing
	h1Regex := regexp.MustCompile(`^#\s+(.+)$`)
	h2Regex := regexp.MustCompile(`^##\s+(.+)$`)
	// More robust regex that handles both formats:
	// - [Title](URL) - Description | Date
	// - [Title](URL) | Date
	linkRegex := regexp.MustCompile(`^\s*[-*+]\s+\[([^\]]+)\]\(([^)]+)\)(?:\s*-\s*(.+?))?(?:\s*\|\s*(\d{4}-\d{2}-\d{2}))?\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check for document title (# heading)
		if h1Match := h1Regex.FindStringSubmatch(line); h1Match != nil {
			data.Title = strings.TrimSpace(h1Match[1])
			continue
		}

		// Check for category headers (## heading)
		if h2Match := h2Regex.FindStringSubmatch(line); h2Match != nil {
			categoryName := strings.TrimSpace(h2Match[1])
			if categoryName == "" {
				currentCategory = "General"
			} else {
				currentCategory = SanitizeCategory(categoryName)
			}
			// Initialize category if it doesn't exist
			if _, exists := data.Categories[currentCategory]; !exists {
				data.Categories[currentCategory] = []Link{}
			}
			continue
		}

		// Check for link items (- [Title](URL) - Description | Date or - [Title](URL) | Date)
		if linkMatch := linkRegex.FindStringSubmatch(line); linkMatch != nil {
			title := strings.TrimSpace(linkMatch[1])
			url := strings.TrimSpace(linkMatch[2])
			description := ""
			if len(linkMatch) > 3 && linkMatch[3] != "" {
				description = strings.TrimSpace(linkMatch[3])
			}
			dateStr := ""
			if len(linkMatch) > 4 && linkMatch[4] != "" {
				dateStr = strings.TrimSpace(linkMatch[4])
			}

			// Create link
			link := Link{
				Title:       title,
				URL:         url,
				Description: description,
				Category:    currentCategory,
			}

			// Parse date if provided
			if dateStr != "" {
				if date, err := ParseLinkDate(dateStr); err == nil {
					link.AddedAt = date
				}
				// Ignore date parsing errors, just use zero time
			}

			// Validate link
			if err := ValidateLink(link); err != nil {
				// Skip invalid links but continue parsing
				continue
			}

			// Initialize category if it doesn't exist
			if _, exists := data.Categories[currentCategory]; !exists {
				data.Categories[currentCategory] = []Link{}
			}

			// Add link to category
			data.Categories[currentCategory] = append(data.Categories[currentCategory], link)
		}
	}

	// Calculate statistics
	data.calculateStats()

	return data, nil
}

// calculateStats calculates statistics for the links data
func (ld *LinksData) calculateStats() {
	totalLinks := 0
	recentLinks := 0
	var latestDate time.Time

	// 1 day ago for "recent" calculation
	oneDayAgo := time.Now().AddDate(0, 0, -1)

	for _, links := range ld.Categories {
		totalLinks += len(links)

		for _, link := range links {
			// Count recent links (added in last 1 day)
			if !link.AddedAt.IsZero() && link.AddedAt.After(oneDayAgo) {
				recentLinks++
			}

			// Track latest added date
			if !link.AddedAt.IsZero() && link.AddedAt.After(latestDate) {
				latestDate = link.AddedAt
			}
		}
	}

	ld.Stats = LinksStats{
		TotalLinks:      totalLinks,
		TotalCategories: len(ld.Categories),
		RecentLinks:     recentLinks,
		LatestAdded:     latestDate,
	}
}

// LinkValidationError represents an error in link validation
type LinkValidationError struct {
	Message string
}

func (e *LinkValidationError) Error() string {
	return e.Message
}

// RenderLinks renders links data as HTML using template functions
func RenderLinks(content string) (string, error) {
	// Parse the markdown content to extract links data
	linksData, err := ParseLinksContent(content)
	if err != nil {
		return "", fmt.Errorf("failed to parse links content: %v", err)
	}

	labels := linksLabels{
		NoResultsTitle:   i18n.Translate("links.no_results_title"),
		NoResultsMessage: i18n.Translate("links.no_results_message"),
		AddNewLink:       i18n.Translate("links.add_new_link"),
	}
	return renderLinksTemplate(linksData, labels)
}

func renderLinksTemplate(linksData *LinksData, labels linksLabels) (string, error) {
	t := template.New("links").Funcs(template.FuncMap{"getFaviconURL": getFaviconURL})

	tmpl := `<div class="links-container" data-total-links="{{.Stats.TotalLinks}}">
    {{if .Title}}
    <!-- Document Title -->
    <h1 class="links-document-title">{{.Title}}</h1>
    {{end}}
    
    <!-- Search and Filter Controls -->
    <div class="links-controls">
        <div class="search-filter-row">
            <div class="search-group">
                <input type="text" id="linksSearch" class="search-input" placeholder="Search links by title, description, or URL..." autocomplete="off">
                <button type="button" class="search-clear" id="searchClear" title="Clear search" hidden>×</button>
            </div>
            <div class="filter-group">
                <div class="language-selector-wrapper">
                    <select id="categoryFilter" class="language-selector">
                        <option value="">All Categories</option>
                        {{range $category, $links := .Categories}}
                        <option value="{{$category}}">{{$category}} ({{len $links}})</option>
                        {{end}}
                    </select>
                </div>
                <div class="language-selector-wrapper">
                    <select id="sortFilter" class="language-selector">
                        <option value="default">Default Order</option>
                        <option value="title-asc">Title A-Z</option>
                        <option value="title-desc">Title Z-A</option>
                        <option value="date-newest">Newest First</option>
                        <option value="date-oldest">Oldest First</option>
                        <option value="category">By Category</option>
                    </select>
                </div>
            </div>
        </div>
        <div class="search-results-info" id="searchResultsInfo" hidden>
            <span class="results-count">0 links found</span>
            <button type="button" class="clear-filters" id="clearFilters">Clear all filters</button>
        </div>
    </div>

    <!-- Stats bar -->
    <div class="links-stats-bar">
        <div class="links-stats-item">
            <div class="links-stats-number">{{.Stats.TotalLinks}}</div>
            <div class="links-stats-label">Total Links</div>
        </div>
        <div class="links-stats-item">
            <div class="links-stats-number">{{.Stats.TotalCategories}}</div>
            <div class="links-stats-label">Categories</div>
        </div>
        <div class="links-stats-item">
            <div class="links-stats-number">{{.Stats.RecentLinks}}</div>
			<div class="links-stats-label">Recent</div>
        </div>
        {{if not .Stats.LatestAdded.IsZero}}
        <div class="links-stats-item">
            <div class="links-stats-number">{{.Stats.LatestAdded.Format "Jan 2"}}</div>
            <div class="links-stats-label">Latest Added</div>
        </div>
        {{end}}
    </div>

    <!-- Links sections -->
    <div class="links-content" id="linksContent">
        <!-- Hidden data for all categories (including empty ones) -->
        <div id="allCategories" hidden data-categories="{{range $category, $links := .Categories}}{{$category}},{{end}}"></div>
        
        {{range $category, $links := .Categories}}
        {{if gt (len $links) 0}}
        <div class="links-category" data-category="{{$category}}">
            <h2 class="links-category-header">
                {{$category}} <span class="section-count">({{len $links}})</span>
            </h2>
            
            {{range $links}}
            <div class="link-item" data-category="{{.Category}}" data-title="{{.Title}}" data-description="{{.Description}}" data-url="{{.URL}}" data-date="{{.AddedAt.Unix}}">
                <div class="link-content">
                    <div class="link-title-row">
                        <div class="link-favicon">
                            <img src="{{getFaviconURL .URL}}" alt="{{.Title}}">
                        </div>
                        <div class="link-title">
                            <a href="{{.URL}}" target="_blank" rel="noopener noreferrer" title="{{.URL}}">{{.Title}}</a>
                            <span class="external-icon">↗</span>
                        </div>
                    </div>
                    <div class="link-description">{{.Description}}</div>
                    {{if not .AddedAt.IsZero}}
                    <div class="link-date">Added {{.AddedAt.Format "Jan 2, 2006"}}</div>
                    {{end}}
                </div>
            </div>
            {{end}}
        </div>
        {{end}}
        {{end}}
    </div>

    <!-- No results message -->
    <div class="no-results" id="noResults" hidden>
        <div class="no-results-icon">🔍</div>
        <div class="no-results-title">{{.Labels.NoResultsTitle}}</div>
        <div class="no-results-message">{{.Labels.NoResultsMessage}}</div>
    </div>

    <!-- Floating Add Link button for admin/editor users -->
    <div class="floating-add-link-container editor-admin-only">
        <button type="button" class="floating-add-link-btn" title="{{.Labels.AddNewLink}}">
            <i class="fa fa-plus"></i>
        </button>
    </div>
</div>`

	t, err := t.Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %v", err)
	}

	// Execute template
	var buf strings.Builder
	if err := t.Execute(&buf, linksTemplateData{LinksData: linksData, Labels: labels}); err != nil {
		return "", fmt.Errorf("failed to execute template: %v", err)
	}

	return buf.String(), nil
}

// Helper functions for template rendering

// getFaviconURL returns the favicon URL for a given domain
func getFaviconURL(url string) string {
	// Try to extract domain from URL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	u, err := urlPkg.Parse(url)
	if err != nil {
		return ""
	}

	// Use Google's favicon service
	return fmt.Sprintf("https://www.google.com/s2/favicons?domain=%s&sz=32", u.Host)
}
