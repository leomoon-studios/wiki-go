// Package basepath lets the wiki be served from a URL subdirectory such as
// https://example.com/wiki/ instead of the root of a host.
//
// Requests arrive with the base path and have it removed after routing, so
// handlers, access rules, document paths, and API payloads keep using
// wiki-relative paths like /docs/page. Every URL sent to the browser (pages,
// rendered Markdown, redirects, cookies, sitemaps) has the base path added.
package basepath

import (
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

// segmentPattern limits base path segments to characters that never need URL
// escaping, so the configured value matches request paths, cookie paths, and
// browser location.pathname literally.
var segmentPattern = regexp.MustCompile(`^[A-Za-z0-9._~-]+$`)

var state = struct {
	sync.RWMutex
	path string
}{}

// Normalize validates a configured base path and returns its canonical form:
// empty for the host root, otherwise a path such as "/wiki" with a leading
// slash and no trailing slash.
func Normalize(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	value = strings.Trim(value, "/")
	if value == "" {
		return "", nil
	}
	for _, segment := range strings.Split(value, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return "", errors.New("base_path must not contain empty, '.' or '..' segments")
		}
		if !segmentPattern.MatchString(segment) {
			return "", errors.New("base_path segments may only contain letters, digits, '.', '_', '~' and '-'")
		}
	}
	return "/" + value, nil
}

// Set stores the normalized base path used by the rest of the application.
func Set(path string) {
	state.Lock()
	state.path = path
	state.Unlock()
}

// Get returns the configured base path, or "" when served from the host root.
func Get() string {
	state.RLock()
	defer state.RUnlock()
	return state.path
}

// URL adds the base path to a wiki-relative URL such as "/static/app.js" or
// "/docs/page?x=1". Other values (absolute URLs, scheme-relative URLs,
// fragments, and relative references) are returned unchanged.
func URL(path string) string {
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return path
	}
	return Get() + path
}

// CookiePath returns the path attribute for cookies issued by the wiki.
func CookiePath() string {
	if base := Get(); base != "" {
		return base
	}
	return "/"
}

// Middleware removes the base path from incoming requests before application
// handlers run. Routes must be registered with the base path so the router's
// automatic redirects retain it.
// Requests outside the base path receive 404, except "/" which redirects to the
// wiki so the server remains reachable when accessed directly.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		base := Get()
		if base == "" {
			next.ServeHTTP(w, r)
			return
		}

		if r.URL.Path == base || r.URL.Path == "/" {
			target := base + "/"
			if r.URL.RawQuery != "" {
				target += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, target, http.StatusFound)
			return
		}

		rest, ok := strings.CutPrefix(r.URL.Path, base+"/")
		if !ok {
			http.NotFound(w, r)
			return
		}

		stripped := new(http.Request)
		*stripped = *r
		stripped.URL = new(url.URL)
		*stripped.URL = *r.URL
		stripped.URL.Path = "/" + rest
		if rawRest, found := strings.CutPrefix(r.URL.RawPath, base+"/"); found {
			stripped.URL.RawPath = "/" + rawRest
		} else {
			stripped.URL.RawPath = ""
		}
		next.ServeHTTP(w, stripped)
	})
}

// localURLAttribute matches double-quoted URL attributes whose value is a
// wiki-relative path. Rendered Markdown always double-quotes attribute values
// and escapes quotes in text, so text and code content are never matched.
var localURLAttribute = regexp.MustCompile(`(\s(?:href|src|poster|action)=")(/(?:[^/"][^"]*)?)"`)

// RewriteHTML adds the base path to wiki-relative href, src, poster, and action
// attributes in HTML produced by the Markdown renderer. Authors write links
// relative to the wiki root, for example [Guide](/docs/guide).
func RewriteHTML(html []byte) []byte {
	base := Get()
	if base == "" {
		return html
	}
	return localURLAttribute.ReplaceAll(html, []byte(`${1}`+base+`${2}"`))
}
