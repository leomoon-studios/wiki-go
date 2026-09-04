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

// NavTree wraps the navigation root with render-time options that apply
// uniformly to every node (e.g. whether all directories should render expanded).
// It exists so per-request, per-tree flags don't have to be duplicated onto
// every NavItem.
type NavTree struct {
	Root       *NavItem
	AlwaysOpen bool
}

// BreadcrumbItem represents an item in the breadcrumb trail
type BreadcrumbItem struct {
	Title  string
	Path   string
	IsLast bool
}

// ChapterHeading is an escaped template view of one rendered document
// heading. Level is limited to 1-6 and ID is normalized before construction.
type ChapterHeading struct {
	Level    int
	Text     string
	ID       string
	Children []ChapterHeading
}

// PageData represents the data passed to the template
type PageData struct {
	Navigation         *NavTree
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
	ChapterHeadings    []ChapterHeading   // Structured outline for the chapter-links panel
}
