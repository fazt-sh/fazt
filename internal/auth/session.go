// LEGACY_CODE: In-memory session store is no longer used.
// All sessions now use database-backed auth.Service (sessions_db.go).
// This file and session_test.go can be removed in future cleanup.
// Kept for reference only.

package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Session represents an authenticated user session
type Session struct {
	ID        string
	Username  string
	CreatedAt time.Time
	ExpiresAt time.Time
	LastSeen  time.Time
}

// SessionStore manages active sessions
type SessionStore struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	ttl      time.Duration
	stopChan chan struct{}
}

// NewSessionStore creates a new session store
func NewSessionStore(ttl time.Duration) *SessionStore {
	store := &SessionStore{
		sessions: make(map[string]*Session),
		ttl:      ttl,
		stopChan: make(chan struct{}),
	}

	// Start cleanup goroutine
	go store.cleanupExpired()

	return store
}

// CreateSession creates a new session for a user
func (s *SessionStore) CreateSession(username string) (string, error) {
	if username == "" {
		return "", errors.New("username cannot be empty")
	}

	// Generate secure session ID
	sessionID, err := generateSessionID()
	if err != nil {
		return "", fmt.Errorf("failed to generate session ID: %w", err)
	}

	now := time.Now()
	session := &Session{
		ID:        sessionID,
		Username:  username,
		CreatedAt: now,
		ExpiresAt: now.Add(s.ttl),
		LastSeen:  now,
	}

	s.mu.Lock()
	s.sessions[sessionID] = session
	s.mu.Unlock()

	return sessionID, nil
}

// ValidateSession checks if a session is valid and not expired
func (s *SessionStore) ValidateSession(sessionID string) (bool, error) {
	if sessionID == "" {
		return false, errors.New("session ID cannot be empty")
	}

	s.mu.RLock()
	session, exists := s.sessions[sessionID]
	s.mu.RUnlock()

	if !exists {
		return false, nil
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		// Remove expired session
		s.DeleteSession(sessionID)
		return false, nil
	}

	// Session is valid, refresh it
	s.RefreshSession(sessionID)

	return true, nil
}

// GetSession retrieves a session by ID
func (s *SessionStore) GetSession(sessionID string) (*Session, error) {
	if sessionID == "" {
		return nil, errors.New("session ID cannot be empty")
	}

	s.mu.RLock()
	session, exists := s.sessions[sessionID]
	s.mu.RUnlock()

	if !exists {
		return nil, errors.New("session not found")
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		s.DeleteSession(sessionID)
		return nil, errors.New("session expired")
	}

	return session, nil
}

// RefreshSession updates the LastSeen time and extends the expiry
func (s *SessionStore) RefreshSession(sessionID string) error {
	if sessionID == "" {
		return errors.New("session ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return errors.New("session not found")
	}

	now := time.Now()
	session.LastSeen = now
	session.ExpiresAt = now.Add(s.ttl)

	return nil
}

// DeleteSession removes a session from the store
func (s *SessionStore) DeleteSession(sessionID string) {
	if sessionID == "" {
		return
	}

	s.mu.Lock()
	delete(s.sessions, sessionID)
	s.mu.Unlock()
}

// DeleteUserSessions removes all sessions for a specific user
func (s *SessionStore) DeleteUserSessions(username string) int {
	if username == "" {
		return 0
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for id, session := range s.sessions {
		if session.Username == username {
			delete(s.sessions, id)
			count++
		}
	}

	return count
}

// GetUserSessions returns all active sessions for a user
func (s *SessionStore) GetUserSessions(username string) []*Session {
	if username == "" {
		return nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var sessions []*Session
	for _, session := range s.sessions {
		if session.Username == username {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// cleanupExpired runs periodically to remove expired sessions
func (s *SessionStore) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanup()
		case <-s.stopChan:
			return
		}
	}
}

// cleanup removes all expired sessions
func (s *SessionStore) cleanup() {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	for id, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			delete(s.sessions, id)
		}
	}
}

// Stop stops the cleanup goroutine
func (s *SessionStore) Stop() {
	close(s.stopChan)
}

// Count returns the number of active sessions
func (s *SessionStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}

// generateSessionID generates a cryptographically secure random session ID
func generateSessionID() (string, error) {
	// Generate 32 bytes of random data
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Encode to base64 URL-safe string
	return base64.URLEncoding.EncodeToString(b), nil
}
