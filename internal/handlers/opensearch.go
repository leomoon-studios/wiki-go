package handlers

import (
	"encoding/xml"
	"net/http"
	"strings"
	"unicode"

	"wiki-go/internal/config"
	"wiki-go/internal/i18n"
)

type openSearchDescription struct {
	XMLName       xml.Name      `xml:"http://a9.com/-/spec/opensearch/1.1/ OpenSearchDescription"`
	ShortName     string        `xml:"ShortName"`
	Description   string        `xml:"Description"`
	InputEncoding string        `xml:"InputEncoding"`
	URL           openSearchURL `xml:"Url"`
}

type openSearchURL struct {
	Type     string `xml:"type,attr"`
	Template string `xml:"template,attr"`
}

// OpenSearchHandler exposes a browser provider for the configured or requested origin.
func OpenSearchHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	base, err := requestPublicOrigin(r, cfg)
	if err != nil {
		http.Error(w, i18n.Translate("search.public_url_error", cfg.Wiki.Language), http.StatusServiceUnavailable)
		return
	}
	name := strings.TrimSpace(strings.Map(func(r rune) rune {
		if !unicode.IsPrint(r) {
			return -1
		}
		return r
	}, cfg.Wiki.Title))
	if name == "" {
		name = "Wiki-Go"
	}
	shortName := []rune(name)
	if len(shortName) > 16 {
		shortName = shortName[:16]
	}
	desc := openSearchDescription{
		ShortName:     string(shortName),
		Description:   "Search " + name,
		InputEncoding: "UTF-8",
		URL:           openSearchURL{Type: "text/html", Template: base + "/?search={searchTerms}"},
	}
	w.Header().Set("Content-Type", "application/opensearchdescription+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "private, no-store")
	_, _ = w.Write([]byte(xml.Header))
	_ = xml.NewEncoder(w).Encode(desc)
}
