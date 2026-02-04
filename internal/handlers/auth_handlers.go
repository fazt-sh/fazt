package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/fazt-sh/fazt/internal/api"
	"github.com/fazt-sh/fazt/internal/audit"
	"github.com/fazt-sh/fazt/internal/auth"
	"github.com/fazt-sh/fazt/internal/config"
)

var (
	authService   *auth.Service
	rateLimiter   *auth.RateLimiter
	serverVersion string
)

// InitAuth initializes the auth handlers with the auth service and rate limiter
func InitAuth(service *auth.Service, limiter *auth.RateLimiter, version string) {
	authService = service
	rateLimiter = limiter
	serverVersion = version
}

// UserMeHandler returns the current authenticated user info
func UserMeHandler(w http.ResponseWriter, r *http.Request) {
	user, err := authService.GetSessionFromRequest(r)
	if err != nil {
		// All session errors result in 401 - the client should re-authenticate
		api.Unauthorized(w, "Authentication required")
		return
	}

	api.Success(w, http.StatusOK, map[string]interface{}{
		"username": user.Name,
		"email":    user.Email,
		"role":     user.Role,
		"version":  serverVersion,
	})
}

// LoginHandler handles the API login request (password authentication)
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.BadRequest(w, "Method not allowed")
		return
	}

	// Get client IP for rate limiting
	ip := getClientIP(r)

	// Check rate limit
	if !rateLimiter.AllowLogin(ip) {
		log.Printf("Rate limit exceeded for IP: %s", ip)
		api.RateLimitExceeded(w, "Too many failed attempts. Please try again in 15 minutes.")
		return
	}

	// Parse login request
	var req struct {
		Username   string `json:"username"`
		Password   string `json:"password"`
		RememberMe bool   `json:"remember_me"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.InvalidJSON(w, "Invalid request body")
		return
	}

	// Get config and verify credentials
	cfg := config.Get()

	if req.Username != cfg.Auth.Username {
		rateLimiter.RecordAttempt(ip)
		audit.LogFailure(req.Username, ip, "login", "/api/login", "invalid username")
		log.Printf("Login failed: invalid username from %s", ip)
		api.InvalidCredentials(w)
		return
	}

	if err := auth.VerifyPassword(req.Password, cfg.Auth.PasswordHash); err != nil {
		rateLimiter.RecordAttempt(ip)
		audit.LogFailure(req.Username, ip, "login", "/api/login", "invalid password")
		log.Printf("Login failed: invalid password from %s", ip)
		api.InvalidCredentials(w)
		return
	}

	// Get or create the local admin user
	user, err := authService.GetOrCreateLocalAdmin(req.Username)
	if err != nil {
		log.Printf("Failed to get/create local admin user: %v", err)
		api.InternalError(w, err)
		return
	}

	// Create database session
	token, err := authService.CreateSession(user.ID)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		api.InternalError(w, err)
		return
	}

	// Set session cookie
	maxAge := int(auth.DefaultSessionTTL.Seconds())
	if req.RememberMe {
		maxAge = int(auth.RememberMeTTL.Seconds())
	}
	http.SetCookie(w, authService.SessionCookie(token, maxAge))

	// Reset rate limit on successful login
	rateLimiter.Reset(ip)

	// Log successful login
	audit.LogSuccess(req.Username, ip, "login", "/api/login")
	log.Printf("Login successful: %s from %s", req.Username, ip)

	api.Success(w, http.StatusOK, map[string]interface{}{
		"message": "Login successful",
	})
}

// LogoutHandler handles logout requests
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Get session info for audit logging
	var username string
	user, err := authService.GetSessionFromRequest(r)
	if err == nil {
		username = user.Name
	}

	// Delete session from database
	cookie, err := r.Cookie("fazt_session")
	if err == nil && cookie.Value != "" {
		authService.DeleteSession(cookie.Value)
	}

	// Clear session cookie
	http.SetCookie(w, authService.ClearSessionCookie())

	// Log logout
	if username != "" {
		ip := getClientIP(r)
		audit.LogSuccess(username, ip, "logout", "/api/logout")
	}

	// Return success for API requests
	if r.Header.Get("Accept") == "application/json" || strings.HasPrefix(r.URL.Path, "/api/") {
		api.Success(w, http.StatusOK, map[string]interface{}{
			"message": "Logged out successfully",
		})
		return
	}

	// Redirect to login page for HTML requests
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// AuthStatusHandler returns the current authentication status
func AuthStatusHandler(w http.ResponseWriter, r *http.Request) {
	user, err := authService.GetSessionFromRequest(r)
	if err != nil {
		api.Success(w, http.StatusOK, map[string]interface{}{
			"authenticated": false,
		})
		return
	}

	api.Success(w, http.StatusOK, map[string]interface{}{
		"authenticated": true,
		"username":      user.Name,
		"email":         user.Email,
		"role":          user.Role,
	})
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP if multiple are present
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}
