package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"wiki-go/internal/auth"
	"wiki-go/internal/config"
	"wiki-go/internal/ban"
	"wiki-go/internal/crypto"
	"wiki-go/internal/resources"
	"wiki-go/internal/i18n"
	"wiki-go/internal/roles"
	"wiki-go/internal/version"
)

type LoginRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	KeepLoggedIn bool   `json:"keepLoggedIn"`
}

// loginBan handles IP-based banning for failed login attempts.
var loginBan *ban.BanList

// clientIP extracts the real client IP address, considering proxy headers.
func clientIP(r *http.Request) string {
	// Prioritise common proxy headers
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For may contain multiple IPs, the first is the client
		if comma := strings.Index(ip, ","); comma != -1 {
			return strings.TrimSpace(ip[:comma])
		}
		return strings.TrimSpace(ip)
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	// Fallback to RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr // as-is (unlikely path)
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

	ip := clientIP(r)

	// If IP is currently banned, short-circuit before doing any work.
	if loginBan != nil {
		if remaining := loginBan.IsBanned(ip); remaining > 0 {
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
		log.Printf("Error rendering login template: %v", err)
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
		log.Printf("Warning: failed to initialise login ban list: %v", err)
	} else {
		loginBan = bl
	}
}
