package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"log"
	"net/http"
	"sync"
	"time"
	"wiki-go/internal/config"
	"wiki-go/internal/crypto"
)

// Session represents a user session
type Session struct {
	Username     string    `json:"username"`
	Role         string    `json:"role"` // User role: "admin", "editor", or "viewer"
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	LastAccessed time.Time `json:"last_accessed"`
}

var (
	sessions     = make(map[string]Session)
	mu           sync.RWMutex
	sessionStore *SessionStore
)

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// InitSessionStore initializes the session store and loads existing sessions
func InitSessionStore(filePath string) error {
	sessionStore = NewSessionStore(filePath)
	loadedSessions, err := sessionStore.LoadSessions()
	if err != nil {
		return err
	}

	mu.Lock()
	sessions = loadedSessions
	// Cleanup expired sessions on startup
	for token, session := range sessions {
		if session.IsExpired() {
			delete(sessions, token)
		}
	}
	mu.Unlock()

	// Start background cleanup goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			// We need to pass the map to CleanupExpiredSessions
			// But we need to be careful about concurrency.
			// The CleanupExpiredSessions in session_store.go as I wrote it
			// takes the map and modifies it.
			// However, the map 'sessions' is global here and protected by 'mu'.
			// The SessionStore.CleanupExpiredSessions I wrote assumes it owns the map or locks it.
			// But here 'sessions' is the global map.

			// Let's adjust the strategy.
			// We should lock 'mu', perform cleanup on 'sessions', and then save.
			mu.Lock()
			deleted := 0
			for token, session := range sessions {
				if session.IsExpired() {
					delete(sessions, token)
					deleted++
				}
			}
			if deleted > 0 {
				sessionStore.SaveSessions(sessions)
			}
			mu.Unlock()
		}
	}()

	return nil
}

// hashToken returns the SHA256 hash of the token
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// GenerateSessionToken generates a random session token
func GenerateSessionToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// CreateSession creates a new session for the user
func CreateSession(w http.ResponseWriter, username string, role string, keepLoggedIn bool, cfg *config.Config) error {
	token, err := GenerateSessionToken()
	if err != nil {
		return err
	}

	// Set cookie expiration time based on keepLoggedIn flag
	maxAge := 3600 * 24 // 24 hours by default
	if keepLoggedIn {
		maxAge = 3600 * 24 * 30 // 30 days for persistent login
	}

	hashedToken := hashToken(token)

	mu.Lock()
	sessions[hashedToken] = Session{
		Username:     username,
		Role:         role,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Duration(maxAge) * time.Second),
		LastAccessed: time.Now(),
	}
	if sessionStore != nil {
		if err := sessionStore.SaveSessions(sessions); err != nil {
			log.Printf("Error saving sessions in CreateSession: %v", err)
		}
	} else {
		log.Println("Warning: sessionStore is nil in CreateSession")
	}
	mu.Unlock()

	// Set the secure HTTP-only session token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   !cfg.Server.AllowInsecureCookies,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   maxAge,
	})

	// Set a non-HTTP-only cookie for the username to be accessible by JavaScript
	http.SetCookie(w, &http.Cookie{
		Name:     "session_user",
		Value:    username,
		Path:     "/",
		HttpOnly: false,
		Secure:   !cfg.Server.AllowInsecureCookies,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   maxAge,
	})

	return nil
}

// GetSession retrieves the session for the current request
func GetSession(r *http.Request) *Session {
	c, err := r.Cookie("session_token")
	if err != nil {
		return nil
	}

	mu.Lock()
	defer mu.Unlock()

	hashedToken := hashToken(c.Value)
	session, exists := sessions[hashedToken]
	if !exists {
		return nil
	}

	if session.IsExpired() {
		delete(sessions, hashedToken)
		if sessionStore != nil {
			sessionStore.SaveSessions(sessions)
		}
		return nil
	}

	// Update LastAccessed
	session.LastAccessed = time.Now()
	sessions[hashedToken] = session

	return &session
}

// ClearSession removes the session from the sessions map and clears the cookie
func ClearSession(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	c, err := r.Cookie("session_token")
	if err != nil {
		return
	}

	hashedToken := hashToken(c.Value)

	mu.Lock()
	if _, exists := sessions[hashedToken]; exists {
		delete(sessions, hashedToken)
		if sessionStore != nil {
			if err := sessionStore.SaveSessions(sessions); err != nil {
				log.Printf("Error saving sessions in ClearSession: %v", err)
			}
		}
	} else {
		log.Printf("Warning: Session not found during logout for token hash: %s", hashedToken)
	}
	mu.Unlock()

	// Clear the session token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   !cfg.Server.AllowInsecureCookies,
		MaxAge:   -1,
	})

	// Clear the session user cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_user",
		Value:    "",
		Path:     "/",
		HttpOnly: false,
		Secure:   !cfg.Server.AllowInsecureCookies,
		MaxAge:   -1,
	})
}

// ValidateCredentials validates user credentials against the config
func ValidateCredentials(username, password string, cfg *config.Config) (bool, string) {
	for _, user := range cfg.Users {
		if user.Username == username && crypto.CheckPasswordHash(password, user.Password) {
			// Use the user's role
			role := user.Role
			return true, role
		}
	}
	return false, ""
}

// CheckAuth verifies if the user is authenticated and returns their session
func CheckAuth(r *http.Request) *Session {
	return GetSession(r)
}

// RequireAuth checks if the wiki is private and if the user is authenticated
// Returns true if the user is allowed to access the page
func RequireAuth(r *http.Request, cfg *config.Config) bool {
	// If the wiki is not private, allow access
	if !cfg.Wiki.Private {
		return true
	}

	// If the wiki is private, check if the user is authenticated
	session := GetSession(r)
	return session != nil
}

// RequireRole checks if user has required role or higher
func RequireRole(r *http.Request, requiredRole string) bool {
	session := GetSession(r)
	if session == nil {
		return false
	}

	// Role hierarchy: admin > editor > viewer
	switch requiredRole {
	case "admin":
		return session.Role == "admin"
	case "editor":
		return session.Role == "admin" || session.Role == "editor"
	case "viewer":
		return session.Role == "admin" || session.Role == "editor" || session.Role == "viewer"
	default:
		return false
	}
}
