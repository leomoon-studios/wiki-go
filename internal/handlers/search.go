package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"wiki-go/internal/auth"
	"wiki-go/internal/config"
	"wiki-go/internal/types"
	"wiki-go/internal/utils"
)

type SearchRequest struct {
	Query string `json:"query"`
}

type SearchResult = types.SearchResult

func SearchHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	w.Header().Set("Cache-Control", "private, no-store")

	var query string
	switch r.Method {
	case http.MethodGet:
		var err error
		query, err = searchQueryFromURL(r.URL.RawQuery)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
	case http.MethodPost:
		var req SearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		query = req.Query
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := auth.GetSession(r)
	if cfg.Wiki.Private && session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	results := []SearchResult{}
	if strings.TrimSpace(query) != "" {
		results = performSearch(query, session, cfg)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(results)
}

func searchQueryFromURL(rawQuery string) (string, error) {
	return searchQueryParam(rawQuery, "q")
}

func searchQueryParam(rawQuery, key string) (string, error) {
	params, err := url.ParseQuery(rawQuery)
	if err != nil || len(params[key]) > 1 {
		return "", errors.New("invalid search query")
	}
	if len(params[key]) == 0 {
		return "", nil
	}
	return params[key][0], nil
}

func performSearch(query string, session *auth.Session, cfg *config.Config) []SearchResult {
	type rankedResult struct {
		SearchResult
		titleMatch bool
	}
	var ranked []rankedResult
	searchTerms := parseSearchQuery(query)
	hasPositiveTerms := len(searchTerms.ExactPhrases) > 0 || len(searchTerms.IncludeWords) > 0
	positiveTerms := SearchTerms{ExactPhrases: searchTerms.ExactPhrases, IncludeWords: searchTerms.IncludeWords}

	// Full path to the documents directory
	docsPath := filepath.Join(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir)

	err := filepath.Walk(docsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only document.md represents a wiki page; other Markdown files are attachments.
		if info.Mode().IsRegular() && info.Name() == "document.md" {
			urlPath, err := searchDocumentPath(docsPath, path)
			if err != nil {
				return nil
			}

			// Skip documents the user cannot access
			document, allowed, err := auth.ResolveDocumentAccess(urlPath, session, cfg)
			if err != nil || !allowed {
				return nil
			}
			filename, err := documentContentPath(document)
			if err != nil {
				return nil
			}
			urlPath = document.LogicalPath
			content, err := os.ReadFile(filename)
			if err != nil {
				return nil
			}

			text := string(content)
			if matchContent(text, searchTerms) {
				title, hasTitle := extractTitle(text)
				excerpt := extractExcerpt(text, searchTerms)

				ranked = append(ranked, rankedResult{
					SearchResult: SearchResult{Title: title, Path: urlPath, Excerpt: excerpt},
					titleMatch:   hasTitle && hasPositiveTerms && matchContent(title, positiveTerms),
				})
			}
		}
		return nil
	})

	if err != nil {
		return []SearchResult{}
	}

	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].titleMatch != ranked[j].titleMatch {
			return ranked[i].titleMatch
		}
		return ranked[i].Path < ranked[j].Path
	})
	results := make([]SearchResult, len(ranked))
	for i, result := range ranked {
		results[i] = result.SearchResult
	}
	return results
}

// searchDocumentPath returns the page identity corresponding to a document file.
func searchDocumentPath(documentsRoot, filePath string) (string, error) {
	if filepath.Base(filePath) != "document.md" {
		return "", fmt.Errorf("not a document file")
	}
	relative, err := filepath.Rel(documentsRoot, filepath.Dir(filePath))
	if err != nil || relative == "." || relative == ".." || filepath.IsAbs(relative) || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("document is outside the searchable documents root")
	}
	logicalPath, err := utils.CanonicalRequestPath(filepath.ToSlash(relative))
	if err != nil || logicalPath == "/" || strings.TrimPrefix(logicalPath, "/") != filepath.ToSlash(relative) {
		return "", fmt.Errorf("invalid search document path")
	}
	return logicalPath, nil
}

type SearchTerms struct {
	ExactPhrases []string
	IncludeWords []string
	ExcludeWords []string
}

func parseSearchQuery(query string) SearchTerms {
	var terms SearchTerms

	// Extract exact phrases (text within quotes)
	exactMatches := strings.Count(query, "\"")
	if exactMatches >= 2 {
		for {
			start := strings.Index(query, "\"")
			if start == -1 {
				break
			}
			end := strings.Index(query[start+1:], "\"")
			if end == -1 {
				break
			}
			end += start + 1

			phrase := query[start+1 : end]
			if phrase != "" {
				terms.ExactPhrases = append(terms.ExactPhrases, strings.ToLower(phrase))
			}

			query = query[:start] + query[end+1:]
		}
	}

	// Split remaining words
	words := strings.Fields(query)
	for i := 0; i < len(words); i++ {
		word := strings.ToLower(words[i])

		if word == "not" && i+1 < len(words) {
			terms.ExcludeWords = append(terms.ExcludeWords, strings.ToLower(words[i+1]))
			i++
		} else if word == "and" {
			continue
		} else {
			terms.IncludeWords = append(terms.IncludeWords, word)
		}
	}

	return terms
}

func matchContent(content string, terms SearchTerms) bool {
	content = strings.ToLower(content)

	// Check exact phrases
	for _, phrase := range terms.ExactPhrases {
		if !strings.Contains(content, phrase) {
			return false
		}
	}

	// Check included words
	for _, word := range terms.IncludeWords {
		if !strings.Contains(content, word) {
			return false
		}
	}

	// Check excluded words
	for _, word := range terms.ExcludeWords {
		if strings.Contains(content, word) {
			return false
		}
	}

	return true
}

func extractTitle(content string) (string, bool) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			if title := strings.TrimSpace(strings.TrimPrefix(line, "# ")); title != "" {
				return title, true
			}
			return "Untitled", false
		}
	}
	return "Untitled", false
}

func extractExcerpt(content string, terms SearchTerms) string {
	const excerptLength = 200
	content = strings.ToLower(content)

	// First try to find a match with exact phrases
	var matchIndex int
	if len(terms.ExactPhrases) > 0 {
		for _, phrase := range terms.ExactPhrases {
			if idx := strings.Index(content, phrase); idx != -1 {
				matchIndex = idx
				break
			}
		}
	} else if len(terms.IncludeWords) > 0 {
		// Then try with included words
		for _, word := range terms.IncludeWords {
			if idx := strings.Index(content, word); idx != -1 {
				matchIndex = idx
				break
			}
		}
	}

	// Calculate excerpt range
	start := matchIndex - excerptLength/2
	if start < 0 {
		start = 0
	}
	end := start + excerptLength
	if end > len(content) {
		end = len(content)
	}

	// Trim to word boundaries
	excerpt := content[start:end]
	if start > 0 {
		if idx := strings.Index(excerpt, " "); idx != -1 {
			excerpt = "..." + excerpt[idx:]
		}
	}
	if end < len(content) {
		if idx := strings.LastIndex(excerpt, " "); idx != -1 {
			excerpt = excerpt[:idx] + "..."
		}
	}

	return excerpt
}
