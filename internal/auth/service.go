package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"
)

// Common errors
var (
	ErrUserNotFound     = errors.New("user not found")
	ErrUserExists       = errors.New("user already exists")
	ErrInvalidSession   = errors.New("invalid session")
	ErrSessionExpired   = errors.New("session expired")
	ErrProviderDisabled = errors.New("provider not enabled")
	ErrInvalidState     = errors.New("invalid or expired state")
	ErrInvalidInvite    = errors.New("invalid or expired invite code")
	ErrInviteUsed       = errors.New("invite code already used")
)

// Service is the main authentication service
type Service struct {
	db     *sql.DB
	domain string // Base domain for cookies (e.g., "zyt.app")
	secure bool   // Whether to use secure cookies (HTTPS)
}

// NewService creates a new auth service
func NewService(db *sql.DB, domain string, secure bool) *Service {
	return &Service{
		db:     db,
		domain: domain,
		secure: secure,
	}
}

// Domain returns the base domain for this service
func (s *Service) Domain() string {
	return s.domain
}

// IsSecure returns whether secure cookies should be used
func (s *Service) IsSecure() bool {
	return s.secure
}

// DB returns the database connection
func (s *Service) DB() *sql.DB {
	return s.db
}

// generateToken generates a cryptographically secure random token
func generateToken(bytes int) (string, error) {
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// hashToken creates a SHA-256 hash of a token for storage
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// generateUUID generates a simple UUID v4
func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	// Set version (4) and variant bits
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return hex.EncodeToString(b[:4]) + "-" +
		hex.EncodeToString(b[4:6]) + "-" +
		hex.EncodeToString(b[6:8]) + "-" +
		hex.EncodeToString(b[8:10]) + "-" +
		hex.EncodeToString(b[10:])
}

// CleanupExpired removes expired sessions, states, and invites
func (s *Service) CleanupExpired() error {
	now := time.Now().Unix()

	// Clean expired sessions
	_, err := s.db.Exec(`DELETE FROM auth_sessions WHERE expires_at < ?`, now)
	if err != nil {
		return err
	}

	// Clean expired states
	_, err = s.db.Exec(`DELETE FROM auth_states WHERE expires_at < ?`, now)
	if err != nil {
		return err
	}

	return nil
}

// StartCleanupRoutine starts a background goroutine to clean expired data
func (s *Service) StartCleanupRoutine(stop <-chan struct{}) {
	ticker := time.NewTicker(15 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.CleanupExpired()
			case <-stop:
				ticker.Stop()
				return
			}
		}
	}()
}

// SessionCookie creates a session cookie for the given token
func (s *Service) SessionCookie(token string, maxAge int) *http.Cookie {
	domain := s.domain
	// Add leading dot for subdomain cookie sharing
	if !strings.HasPrefix(domain, ".") && !strings.Contains(domain, "localhost") {
		domain = "." + domain
	}

	sameSite := http.SameSiteLaxMode
	// For localhost/dev, we can't use Secure or cross-domain cookies
	if strings.Contains(s.domain, "localhost") || strings.Contains(s.domain, "127.0.0.1") {
		domain = "" // Don't set domain for localhost
	}

	return &http.Cookie{
		Name:     "fazt_session",
		Value:    token,
		Domain:   domain,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   s.secure,
		SameSite: sameSite,
	}
}

// ClearSessionCookie returns a cookie that clears the session
func (s *Service) ClearSessionCookie() *http.Cookie {
	cookie := s.SessionCookie("", -1)
	cookie.MaxAge = -1
	return cookie
}

// GetSessionFromRequest extracts and validates the session from a request
func (s *Service) GetSessionFromRequest(r *http.Request) (*User, error) {
	cookie, err := r.Cookie("fazt_session")
	if err != nil {
		return nil, ErrInvalidSession
	}

	return s.ValidateSession(cookie.Value)
}

// GetSessionFromRequestInterface is a wrapper that returns interface{} for runtime compatibility
func (s *Service) GetSessionFromRequestInterface(r *http.Request) (interface{}, error) {
	user, err := s.GetSessionFromRequest(r)
	if err != nil {
		return nil, err
	}
	// Convert to map for runtime package (avoids import cycle)
	return map[string]interface{}{
		"id":       user.ID,
		"email":    user.Email,
		"name":     user.Name,
		"picture":  user.Picture,
		"role":     user.Role,
		"provider": user.Provider,
	}, nil
}

// AuthProviderAdapter adapts the Service to the runtime.AuthProvider interface
type AuthProviderAdapter struct {
	service *Service
}

// NewAuthProviderAdapter creates an adapter for use with the runtime
func NewAuthProviderAdapter(service *Service) *AuthProviderAdapter {
	return &AuthProviderAdapter{service: service}
}

// GetSessionFromRequest implements runtime.AuthProvider
func (a *AuthProviderAdapter) GetSessionFromRequest(r *http.Request) (interface{}, error) {
	return a.service.GetSessionFromRequestInterface(r)
}

// Domain implements runtime.AuthProvider
func (a *AuthProviderAdapter) Domain() string {
	return a.service.Domain()
}
