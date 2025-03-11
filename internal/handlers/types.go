package handlers

import (
	"html/template"
	"time"
	"wiki-go/internal/config"
	"wiki-go/internal/utils"
)

// PageData represents the data passed to templates
type PageData struct {
	Navigation   *utils.NavItem
	Content      template.HTML
	CurrentDir   *utils.NavItem
	Breadcrumbs  []BreadcrumbItem
	Config       *config.Config
	LastModified time.Time
}

// BreadcrumbItem represents a single item in the breadcrumb trail
type BreadcrumbItem struct {
	Title  string
	Path   string
	IsLast bool
}
