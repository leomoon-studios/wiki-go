package routes

import (
	"net/http"

	"wiki-go/internal/basepath"
)

// wikiMux routes public URLs, then removes the base path for application
// handlers. ServeMux's own redirects therefore retain the public base path.
type wikiMux struct {
	mux *http.ServeMux
}

func newWikiMux() *wikiMux {
	router := &wikiMux{mux: http.NewServeMux()}
	if base := basepath.Get(); base != "" {
		// Preserve the wiki-root redirects and reject paths outside the mount.
		fallback := basepath.Middleware(http.NotFoundHandler())
		router.mux.Handle("/", fallback)
		router.mux.Handle(base, fallback)
	}
	return router
}

func (router *wikiMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	router.mux.Handle(basepath.URL(pattern), basepath.Middleware(http.HandlerFunc(handler)))
}

func (router *wikiMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.mux.ServeHTTP(w, r)
}
