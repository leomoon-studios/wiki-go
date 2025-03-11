package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"wiki-go/internal/config"
)

type SearchRequest struct {
	Query string `json:"query"`
}

type SearchResult struct {
	Title   string `json:"title"`
	Path    string `json:"path"`
	Excerpt string `json:"excerpt"`
}

func SearchHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	results := performSearch(req.Query, cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func performSearch(query string, rootDir string, documentsDir string) []SearchResult {
	var results []SearchResult
	searchTerms := parseSearchQuery(query)

	// Full path to the documents directory
	docsPath := filepath.Join(rootDir, documentsDir)

	err := filepath.Walk(docsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process markdown files
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".md") {
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			if matches := matchContent(string(content), searchTerms); matches {
				title := extractTitle(string(content))
				excerpt := extractExcerpt(string(content), searchTerms)

				// Clean up the path
				cleanPath := path
				// First convert to forward slashes
				cleanPath = strings.ReplaceAll(cleanPath, "\\", "/")

				// Remove the rootDir/documentsDir prefix
				prefix := strings.ReplaceAll(docsPath, "\\", "/") + "/"
				cleanPath = strings.TrimPrefix(cleanPath, prefix)

				// Remove document.md and any remaining .md extension
				cleanPath = strings.TrimSuffix(strings.Replace(cleanPath, "document.md", "", 1), ".md")

				results = append(results, SearchResult{
					Title:   title,
					Path:    "/" + cleanPath,
					Excerpt: excerpt,
				})
			}
		}
		return nil
	})

	if err != nil {
		return []SearchResult{}
	}

	return results
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

func extractTitle(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return "Untitled"
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
