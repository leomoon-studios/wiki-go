package auth

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// SessionStore manages session persistence
type SessionStore struct {
	mu       sync.RWMutex
	filePath string
}

// NewSessionStore creates a new session store
func NewSessionStore(filePath string) *SessionStore {
	// Ensure directory exists
	if dir := filepath.Dir(filePath); dir != "" {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			log.Printf("Error creating session store directory: %v", err)
		}
	}

	return &SessionStore{
		filePath: filePath,
	}
}

// SaveSessions saves the sessions to disk
func (s *SessionStore) SaveSessions(sessions map[string]Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a temporary file
	tempFile := s.filePath + ".tmp"
	f, err := os.OpenFile(tempFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}

	// Encode sessions to JSON
	encoder := json.NewEncoder(f)
	if err := encoder.Encode(sessions); err != nil {
		f.Close()
		return err
	}
	f.Close()

	// Rename temporary file to actual file (atomic operation)
	if err := os.Rename(tempFile, s.filePath); err != nil {
		return err
	}

	return nil
}

// LoadSessions loads sessions from disk
func (s *SessionStore) LoadSessions() (map[string]Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make(map[string]Session)

	// Check if file exists
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return sessions, nil
	}

	// Read file
	data, err := ioutil.ReadFile(s.filePath)
	if err != nil {
		return nil, err
	}

	// Decode JSON
	if err := json.Unmarshal(data, &sessions); err != nil {
		// If file is corrupted, return empty map and log error
		log.Printf("Error decoding session file: %v", err)
		return sessions, nil
	}

	return sessions, nil
}

// CleanupExpiredSessions removes expired sessions from the map and saves to disk
func (s *SessionStore) CleanupExpiredSessions(sessions map[string]Session) {
	s.mu.Lock()
	defer s.mu.Unlock() // Lock for the whole operation to ensure consistency

	deleted := 0

	for token, session := range sessions {
		if session.IsExpired() {
			delete(sessions, token)
			deleted++
		}
	}

	if deleted > 0 {
		// We need to save the cleaned sessions.
		// Since we are already holding the lock, we can't call SaveSessions directly
		// because it also tries to acquire the lock.
		// We should refactor SaveSessions or duplicate the logic here.
		// For simplicity and correctness with the current structure, let's duplicate the save logic
		// but without the lock acquisition.

		// Create a temporary file
		tempFile := s.filePath + ".tmp"
		f, err := os.OpenFile(tempFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
		if err != nil {
			log.Printf("Error creating temp session file during cleanup: %v", err)
			return
		}

		// Encode sessions to JSON
		encoder := json.NewEncoder(f)
		if err := encoder.Encode(sessions); err != nil {
			f.Close()
			log.Printf("Error encoding sessions during cleanup: %v", err)
			return
		}
		f.Close()

		// Rename temporary file to actual file (atomic operation)
		if err := os.Rename(tempFile, s.filePath); err != nil {
			log.Printf("Error renaming session file during cleanup: %v", err)
		}
	}
}
