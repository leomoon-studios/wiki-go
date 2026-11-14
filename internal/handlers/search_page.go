package handlers

import (
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"wiki-go/internal/auth"
	"wiki-go/internal/config"
)

const searchPagePath = "/api/search/page"

var searchPageTemplate = template.Must(template.New("search-page").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Search - {{.SiteTitle}}</title>
<link rel="stylesheet" href="/static/css/theme.css">
<link rel="stylesheet" href="/static/css/layout.css">
<link rel="stylesheet" href="/static/css/search.css">
</head>
<body>
<main>
<h1>Search</h1>
<form method="get" action="/api/search/page" role="search">
<label for="search-query">Search this wiki</label>
<input id="search-query" type="search" name="q" value="{{.Query}}">
<button type="submit">Search</button>
</form>
{{if .Searched}}
{{if .Results}}
<ul class="search-page-results">
{{range .Results}}<li><a href="{{.Path}}">{{.Title}}</a><p>{{.Excerpt}}</p></li>{{end}}
</ul>
{{else}}<p>No results found.</p>{{end}}
{{end}}
</main>
</body>
</html>`))

// SearchPageHandler renders a bookmarkable search without JavaScript.
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

	query, err := searchQueryFromURL(r.URL.RawQuery)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	results := []SearchResult{}
	if strings.TrimSpace(query) != "" {
		results = performSearch(query, session, cfg)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := searchPageTemplate.Execute(w, struct {
		SiteTitle string
		Query     string
		Searched  bool
		Results   []SearchResult
	}{SiteTitle: cfg.Wiki.Title, Query: query, Searched: strings.TrimSpace(query) != "", Results: results}); err != nil {
		http.Error(w, "Failed to render search page", http.StatusInternalServerError)
	}
}
