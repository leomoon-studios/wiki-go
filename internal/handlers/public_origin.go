package handlers

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"wiki-go/internal/basepath"
	"wiki-go/internal/config"
)

// publicOrigin validates the configured external origin without consulting request headers.
func publicOrigin(raw string) (string, error) {
	base, err := url.Parse(raw)
	if err != nil || base == nil || (base.Scheme != "http" && base.Scheme != "https") || base.Hostname() == "" || base.User != nil || !isBasePathURLPath(base.Path) || base.RawQuery != "" || base.ForceQuery || base.Fragment != "" || strings.ContainsAny(base.Host, "\\ \t\r\n") {
		return "", errors.New("invalid public URL")
	}
	if port := base.Port(); port != "" {
		value, err := strconv.Atoi(port)
		if err != nil || value < 1 || value > 65535 {
			return "", errors.New("invalid public URL port")
		}
	}
	return base.Scheme + "://" + base.Host, nil
}

// isBasePathURLPath reports whether a public URL path names the wiki root:
// empty, "/", or the configured base path with or without a trailing slash.
func isBasePathURLPath(path string) bool {
	path = strings.TrimSuffix(path, "/")
	return path == "" || path == basepath.Get()
}

// requestPublicOrigin uses the configured origin, or derives it from this request.
func requestPublicOrigin(r *http.Request, cfg *config.Config) (string, error) {
	if cfg.Server.PublicURL != "" {
		return publicOrigin(cfg.Server.PublicURL)
	}
	scheme := "http"
	if cfg.Server.SSL || r.TLS != nil {
		scheme = "https"
	}
	host := r.Host
	peer := parseRemoteIP(r.RemoteAddr)
	if peer != nil && isTrustedProxy(peer, cfg.Server.TrustedProxies) {
		if value := r.Header.Get("X-Forwarded-Proto"); value != "" {
			parts := strings.Split(value, ",")
			proto := strings.TrimSpace(parts[len(parts)-1])
			if proto != "http" && proto != "https" {
				return "", errors.New("invalid forwarded scheme")
			}
			scheme = proto
		}
		if value := r.Header.Get("X-Forwarded-Host"); value != "" {
			parts := strings.Split(value, ",")
			host = strings.TrimSpace(parts[len(parts)-1])
		}
	}
	if host == "" || strings.ContainsAny(host, "/?#@\\") {
		return "", errors.New("invalid request host")
	}
	return publicOrigin(scheme + "://" + host)
}
