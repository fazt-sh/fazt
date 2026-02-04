package auth

import (
	"net/http"
	"strings"
	"time"
)

const (
	// SessionCookieName is the name of the session cookie
	// Must match service.go for unified auth across simple login and OAuth flows
	SessionCookieName = "fazt_session"

	// SessionTTL is the default session time-to-live (24 hours)
	SessionTTL = 24 * time.Hour

	// RememberMeTTL is the extended session time for "remember me" (7 days)
	RememberMeTTL = 7 * 24 * time.Hour
)

// SetSessionCookie sets a secure session cookie
// domain should be the parent domain (e.g., "192.168.64.3.nip.io") to allow cross-subdomain auth
func SetSessionCookie(w http.ResponseWriter, sessionID string, ttl time.Duration, isProduction bool, domain string) {
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,         // Prevent JavaScript access
		Secure:   isProduction, // HTTPS only in production
		SameSite: http.SameSiteLaxMode, // Lax allows cross-subdomain requests
	}

	// Set domain for cross-subdomain cookie sharing
	if domain != "" {
		cookie.Domain = domain
	}

	http.SetCookie(w, cookie)
}

// ClearSessionCookie removes the session cookie
// domain should match the domain used when setting the cookie
func ClearSessionCookie(w http.ResponseWriter, domain string) {
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Delete cookie
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	if domain != "" {
		cookie.Domain = domain
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

// ExtractCookieDomain extracts the parent domain from a URL for cookie sharing
// Input: "https://admin.192.168.64.3.nip.io:8080" → Output: ".192.168.64.3.nip.io"
// Returns empty string for localhost (cookies don't need domain for localhost)
func ExtractCookieDomain(url string) string {
	domain := url
	// Strip protocol
	if strings.HasPrefix(domain, "http://") {
		domain = strings.TrimPrefix(domain, "http://")
	}
	if strings.HasPrefix(domain, "https://") {
		domain = strings.TrimPrefix(domain, "https://")
	}
	// Strip path
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}
	// Strip port
	if idx := strings.LastIndex(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}
	// Don't set domain for localhost
	if strings.Contains(domain, "localhost") || strings.Contains(domain, "127.0.0.1") {
		return ""
	}
	// Add leading dot for subdomain cookie sharing
	if !strings.HasPrefix(domain, ".") {
		domain = "." + domain
	}
	return domain
}
