package handlers

import (
	"net/http"
	"net/url"
	"strings"

	"wiki-go/internal/auth"
	"wiki-go/internal/config"
	"wiki-go/internal/i18n"
	"wiki-go/internal/types"
	"wiki-go/internal/utils"
)

const searchPagePath = "/"

// SearchPageHandler renders a bookmarkable search in the standard wiki layout.
func SearchPageHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	w.Header().Set("Cache-Control", "private, no-store")
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := auth.GetSession(r)
	if cfg.Wiki.Private && session == nil {
		target := searchPagePath
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, "/login?redirect="+url.QueryEscape(target), http.StatusSeeOther)
		return
	}

	nav, err := utils.BuildNavigation(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir)
	if err != nil {
		http.Error(w, "Failed to build navigation", http.StatusInternalServerError)
		return
	}
	nav = utils.FilterNavigation(nav, func(path string) bool {
		return auth.CanAccessDocument(path, session, cfg)
	})

	query, queryErr := searchQueryParam(r.URL.RawQuery, "search")
	results := []SearchResult{}
	searched := queryErr == nil && strings.TrimSpace(query) != ""
	if searched {
		results = performSearch(query, session, cfg)
	}

	role := ""
	if session != nil {
		role = session.Role
	}
	data := &types.PageData{
		IsSearchPage:       true,
		SearchQuery:        query,
		SearchSubmitted:    searched,
		SearchError:        queryErr != nil,
		SearchResults:      results,
		Navigation:         &types.NavTree{Root: nav, AlwaysOpen: cfg.Wiki.AlwaysOpenChildrenInSidebar},
		Breadcrumbs:        []types.BreadcrumbItem{{Title: "Home", Path: "/"}, {Title: "Search", Path: searchPagePath, IsLast: true}},
		Config:             cfg,
		CurrentDir:         &types.NavItem{Title: "Search", Path: "/?search="},
		AvailableLanguages: i18n.GetAvailableLanguages(),
		IsAuthenticated:    session != nil,
		UserRole:           role,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if queryErr != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	renderTemplate(w, data)
}
