package handlers

import (
	"encoding/json"
	"html/template"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"wiki-go/internal/auth"
	"wiki-go/internal/ban"
	"wiki-go/internal/config"
	"wiki-go/internal/crypto"
	"wiki-go/internal/i18n"
	"wiki-go/internal/logger"
	"wiki-go/internal/resources"
	"wiki-go/internal/roles"
	"wiki-go/internal/version"
)

type LoginRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	KeepLoggedIn bool   `json:"keepLoggedIn"`
}

// loginBan handles IP-based banning for failed login attempts.
var loginBan *ban.BanList

// clientIP returns a forwarding header address only when the request arrived
// from a configured trusted proxy. Otherwise the direct peer is authoritative.
func clientIP(r *http.Request, cfg *config.Config) string {
	peer := parseRemoteIP(r.RemoteAddr)
	if peer == nil {
		return r.RemoteAddr
	}
	if cfg == nil || !isTrustedProxy(peer, cfg.Server.TrustedProxies) {
		return peer.String()
	}

	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		chain := make([]net.IP, len(parts))
		for i, part := range parts {
			chain[i] = net.ParseIP(strings.TrimSpace(part))
			if chain[i] == nil {
				return peer.String()
			}
		}

		// Work from the nearest address toward the client and discard only
		// proxies that are explicitly trusted. The first untrusted address is
		// the client as observed by the trusted proxy chain.
		for i := len(chain) - 1; i >= 0; i-- {
			if !isTrustedProxy(chain[i], cfg.Server.TrustedProxies) {
				return chain[i].String()
			}
		}
		return chain[0].String()
	}

	if realIP := net.ParseIP(strings.TrimSpace(r.Header.Get("X-Real-IP"))); realIP != nil {
		return realIP.String()
	}
	return peer.String()
}

func parseRemoteIP(remoteAddr string) net.IP {
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return net.ParseIP(host)
	}
	return net.ParseIP(strings.Trim(remoteAddr, "[]"))
}

func isTrustedProxy(ip net.IP, trustedProxies []string) bool {
	for _, value := range trustedProxies {
		value = strings.TrimSpace(value)
		if trustedIP := net.ParseIP(value); trustedIP != nil && trustedIP.Equal(ip) {
			return true
		}
		if _, network, err := net.ParseCIDR(value); err == nil && network.Contains(ip) {
			return true
		}
	}
	return false
}

// LoginHandler handles API login requests
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Method not allowed",
		})
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	ip := clientIP(r, cfg)

	// If IP is currently banned, short-circuit before doing any work.
	if loginBan != nil {
		if remaining := loginBan.IsBanned(ip); remaining > 0 {
			logger.Warn("Login attempt from banned IP %s (retry in %ds)", ip, int(remaining.Seconds()))
			w.Header().Set("Retry-After", strconv.Itoa(int(remaining.Seconds())))
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":    false,
				"retryAfter": int(remaining.Seconds()),
				"message":    "Too many failed logins; try again later",
			})
			return
		}
	}

	// Validate credentials
	valid, role, groups := auth.ValidateCredentials(req.Username, req.Password, cfg)
	if !valid {
		if loginBan != nil {
			if dur, bannedNow := loginBan.RegisterFailure(ip); bannedNow {
				logger.Warn("IP %s banned after too many failed login attempts (ban duration: %ds)", ip, int(dur.Seconds()))
				// Immediately inform client of new ban
				w.Header().Set("Retry-After", strconv.Itoa(int(dur.Seconds())))
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success":    false,
					"retryAfter": int(dur.Seconds()),
					"message":    "Too many failed logins; try again later",
				})
				return
			}
		}
		logger.Warn("Login failed for user %s from %s", req.Username, ip)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid credentials",
		})
		return
	}

	if loginBan != nil {
		loginBan.Clear(ip) // successful login resets failures / ban
	}

	// Create session
	if err := auth.CreateSession(w, req.Username, role, groups, req.KeepLoggedIn, cfg); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to create session",
		})
		return
	}

	logger.Info("User %s logged in from %s", req.Username, ip)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Login successful",
	})
}

// CheckAuthHandler checks if the user is authenticated
func CheckAuthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	session := auth.GetSession(r)
	if session == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Unauthorized",
		})
		return
	}

	// Return user information including role
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"username": session.Username,
		"role":     session.Role,
		"groups":   session.Groups,
	})
}

// LogoutHandler handles user logout
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	session := auth.GetSession(r)
	if session != nil {
		logger.Info("User %s logged out", session.Username)
	}
	auth.ClearSession(w, r, cfg)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Logout successful",
	})
}

// LoginPageHandler renders the login page
func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	// If user is already logged in, redirect to home page
	session := auth.GetSession(r)
	if session != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Prepare the data for the template
	data := struct {
		Config *config.Config
		Theme  string
	}{
		Config: cfg,
		Theme:  "light", // Default theme
	}

	// Get theme from cookie if available
	if cookie, err := r.Cookie("theme"); err == nil {
		data.Theme = cookie.Value
	}

	// Create function map with translation function
	funcMap := template.FuncMap{
		"t": func(key string) string {
			return i18n.Translate(key)
		},
		"getVersion": func() string {
			return version.Version
		},
	}

	// Get and execute login template with translation function
	tmpl, err := template.New("login.html").Funcs(funcMap).ParseFS(resources.GetTemplatesFS(), "templates/login.html")
	if err != nil {
		http.Error(w, "Error loading login template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Set content type header before writing response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Execute template directly to response writer
	err = tmpl.Execute(w, data)
	if err != nil {
		// Since we've already started writing the response, we can't use http.Error here
		// But we can log the error
		logger.Error("Error rendering login template: %v", err)
	}
}

// CheckDefaultPasswordHandler checks if the default admin password is still in use
func CheckDefaultPasswordHandler(w http.ResponseWriter, r *http.Request) {
	// Default admin credentials (typically admin/admin)
	defaultUsername := "admin"
	defaultPassword := "admin"

	// Check if any admin user still has the default password
	defaultPasswordInUse := false

	for _, user := range cfg.Users {
		if user.Role == roles.RoleAdmin && user.Username == defaultUsername {
			// Check if password is still the default
			if crypto.CheckPasswordHash(defaultPassword, user.Password) {
				defaultPasswordInUse = true
				break
			}
		}
	}

	// Return the result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"defaultPasswordInUse": defaultPasswordInUse,
	})
}

// InitLoginBan initialises IP banning using cfg.Wiki.RootDir/temp/login_ban.json.
func InitLoginBan(cfg *config.Config) {
	if !cfg.Security.LoginBan.Enabled {
		loginBan = nil
		return
	}

	// Apply policy overrides from config
	ban.UpdatePolicy(
		cfg.Security.LoginBan.MaxFailures,
		cfg.Security.LoginBan.WindowSeconds,
		cfg.Security.LoginBan.InitialBanSeconds,
		cfg.Security.LoginBan.MaxBanSeconds,
	)

	path := filepath.Join(cfg.Wiki.RootDir, "temp", "login_ban.json")
	bl, err := ban.NewBanList(path)
	if err != nil {
		logger.Warn("Failed to initialise login ban list: %v", err)
	} else {
		loginBan = bl
	}
}
