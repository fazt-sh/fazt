package auth

import (
	"net/http"
	"time"
)

const (
	// SessionCookieName is the name of the session cookie
	SessionCookieName = "cc_session"

	// SessionTTL is the default session time-to-live (24 hours)
	SessionTTL = 24 * time.Hour

	// RememberMeTTL is the extended session time for "remember me" (7 days)
	RememberMeTTL = 7 * 24 * time.Hour
)

// SetSessionCookie sets a secure session cookie
func SetSessionCookie(w http.ResponseWriter, sessionID string, ttl time.Duration, isProduction bool) {
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,         // Prevent JavaScript access
		Secure:   isProduction, // HTTPS only in production
		SameSite: 3,            // SameSiteStrictMode (3) for CSRF protection
	}

	http.SetCookie(w, cookie)
}

// ClearSessionCookie removes the session cookie
func ClearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Delete cookie
		HttpOnly: true,
		SameSite: 3, // SameSiteStrictMode (3)
	}

	http.SetCookie(w, cookie)
}

// GetSessionCookie retrieves the session cookie from a request
func GetSessionCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}
