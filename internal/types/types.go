package types

import (
	"html/template"
	"time"
	"wiki-go/internal/comments"
	"wiki-go/internal/config"
)

// Role constants are now defined in the roles package

// NavItem represents a navigation item (directory)
type NavItem struct {
	Title          string
	Path           string
	IsDir          bool
	Children       []*NavItem
	IsActive       bool
	DocumentLayout string // Layout type from frontmatter
}

// BreadcrumbItem represents an item in the breadcrumb trail
type BreadcrumbItem struct {
	Title  string
	Path   string
	IsLast bool
}

// PageData represents the data passed to the template
type PageData struct {
	Navigation         *NavItem
	Content            template.HTML
	DirContent         template.HTML
	Breadcrumbs        []BreadcrumbItem
	Config             *config.Config
	LastModified       time.Time
	CurrentDir         *NavItem           // Current directory as a NavItem
	Title              string             // Page title
	IsLoginPage        bool               // Whether this is the login page
	AvailableLanguages []string           // Available languages for the UI
	Comments           []comments.Comment // Comments for the document
	CommentsAllowed    bool               // Whether comments are allowed for this document
	IsAuthenticated    bool               // Whether the user is authenticated
	UserRole           string             // User role: "admin", "editor", or "viewer"
	DocPath            string             // Document path for API calls
	DocumentLayout     string             // Document layout type from frontmatter (e.g., "kanban")
	IsEditMode         bool               // Whether page is in edit mode (separate edit page architecture)
	RawContent         string             // Raw markdown content with frontmatter for edit mode
}
