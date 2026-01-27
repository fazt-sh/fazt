package auth

import (
	"database/sql"
	"time"
)

const (
	// DefaultSessionTTL is 30 days
	DefaultSessionTTL = 30 * 24 * time.Hour
)

// DBSession represents a session stored in SQLite
type DBSession struct {
	TokenHash string
	UserID    string
	CreatedAt int64
	ExpiresAt int64
	LastSeen  int64
}

// CreateSession creates a new session for a user and returns the token
func (s *Service) CreateSession(userID string) (string, error) {
	// Generate a secure random token
	token, err := generateToken(32)
	if err != nil {
		return "", err
	}

	tokenHash := hashToken(token)
	now := time.Now().Unix()
	expiresAt := now + int64(DefaultSessionTTL.Seconds())

	_, err = s.db.Exec(`
		INSERT INTO auth_sessions (token_hash, user_id, created_at, expires_at, last_seen)
		VALUES (?, ?, ?, ?, ?)
	`, tokenHash, userID, now, expiresAt, now)

	if err != nil {
		return "", err
	}

	// Update user's last login
	s.UpdateLastLogin(userID)

	return token, nil
}

// ValidateSession validates a session token and returns the associated user
func (s *Service) ValidateSession(token string) (*User, error) {
	if token == "" {
		return nil, ErrInvalidSession
	}

	tokenHash := hashToken(token)
	now := time.Now().Unix()

	var session DBSession
	err := s.db.QueryRow(`
		SELECT token_hash, user_id, created_at, expires_at, last_seen
		FROM auth_sessions WHERE token_hash = ?
	`, tokenHash).Scan(
		&session.TokenHash, &session.UserID,
		&session.CreatedAt, &session.ExpiresAt, &session.LastSeen,
	)

	if err == sql.ErrNoRows {
		return nil, ErrInvalidSession
	}
	if err != nil {
		return nil, err
	}

	// Check if expired
	if now > session.ExpiresAt {
		// Clean up expired session
		s.db.Exec(`DELETE FROM auth_sessions WHERE token_hash = ?`, tokenHash)
		return nil, ErrSessionExpired
	}

	// Update last_seen (with some throttling to avoid too many writes)
	if now-session.LastSeen > 60 { // Only update if more than 1 minute since last update
		s.db.Exec(`UPDATE auth_sessions SET last_seen = ? WHERE token_hash = ?`, now, tokenHash)
	}

	// Get the user
	return s.GetUserByID(session.UserID)
}

// DeleteSession removes a session by token
func (s *Service) DeleteSession(token string) error {
	if token == "" {
		return nil
	}
	tokenHash := hashToken(token)
	_, err := s.db.Exec(`DELETE FROM auth_sessions WHERE token_hash = ?`, tokenHash)
	return err
}

// DeleteUserSessions removes all sessions for a user
func (s *Service) DeleteUserSessions(userID string) (int64, error) {
	result, err := s.db.Exec(`DELETE FROM auth_sessions WHERE user_id = ?`, userID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// ListUserSessions returns all active sessions for a user
func (s *Service) ListUserSessions(userID string) ([]*DBSession, error) {
	now := time.Now().Unix()
	rows, err := s.db.Query(`
		SELECT token_hash, user_id, created_at, expires_at, last_seen
		FROM auth_sessions
		WHERE user_id = ? AND expires_at > ?
		ORDER BY last_seen DESC
	`, userID, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*DBSession
	for rows.Next() {
		var session DBSession
		err := rows.Scan(
			&session.TokenHash, &session.UserID,
			&session.CreatedAt, &session.ExpiresAt, &session.LastSeen,
		)
		if err != nil {
			continue
		}
		sessions = append(sessions, &session)
	}

	return sessions, nil
}

// CountActiveSessions returns the number of active sessions
func (s *Service) CountActiveSessions() (int, error) {
	now := time.Now().Unix()
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM auth_sessions WHERE expires_at > ?`, now).Scan(&count)
	return count, err
}

// RefreshSession extends the session expiry
func (s *Service) RefreshSession(token string) error {
	tokenHash := hashToken(token)
	now := time.Now().Unix()
	newExpiry := now + int64(DefaultSessionTTL.Seconds())

	_, err := s.db.Exec(`
		UPDATE auth_sessions
		SET expires_at = ?, last_seen = ?
		WHERE token_hash = ?
	`, newExpiry, now, tokenHash)

	return err
}
