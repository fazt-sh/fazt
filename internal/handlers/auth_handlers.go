package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/fazt-sh/fazt/internal/audit"
	"github.com/fazt-sh/fazt/internal/auth"
	"github.com/fazt-sh/fazt/internal/config"
)

var (
	sessionStore *auth.SessionStore
	rateLimiter  *auth.RateLimiter
	serverVersion string
)

// InitAuth initializes the auth handlers with the session store and rate limiter
func InitAuth(store *auth.SessionStore, limiter *auth.RateLimiter, version string) {
	sessionStore = store
	rateLimiter = limiter
	serverVersion = version
}

// UserMeHandler returns the current authenticated user info
func UserMeHandler(w http.ResponseWriter, r *http.Request) {
	sessionID, err := auth.GetSessionCookie(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	session, err := sessionStore.GetSession(sessionID)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"username": session.Username,
		"version":  serverVersion,
	})
}

// LoginHandler handles the API login request
func LoginHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get client IP for rate limiting
	ip := getClientIP(r)

	// Check rate limit
	if !rateLimiter.AllowLogin(ip) {
		log.Printf("Rate limit exceeded for IP: %s", ip)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Too many failed attempts. Please try again in 15 minutes.",
		})
		return
	}

	// Parse login request
	var req struct {
		Username   string `json:"username"`
		Password   string `json:"password"`
		RememberMe bool   `json:"remember_me"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request",
		})
		return
	}

	// Get config
	cfg := config.Get()

	// Verify credentials
	if req.Username != cfg.Auth.Username {
		rateLimiter.RecordAttempt(ip)
		audit.LogFailure(req.Username, ip, "login", "/api/login", "invalid username")
		log.Printf("Login failed: invalid username from %s", ip)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid username or password",
		})
		return
	}

	if err := auth.VerifyPassword(req.Password, cfg.Auth.PasswordHash); err != nil {
		rateLimiter.RecordAttempt(ip)
		audit.LogFailure(req.Username, ip, "login", "/api/login", "invalid password")
		log.Printf("Login failed: invalid password from %s", ip)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid username or password",
		})
		return
	}

	// Credentials valid - create session
	sessionID, err := sessionStore.CreateSession(req.Username)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to create session",
		})
		return
	}

	// Set session cookie
	ttl := auth.SessionTTL
	if req.RememberMe {
		ttl = auth.RememberMeTTL
	}

	auth.SetSessionCookie(w, sessionID, ttl, cfg.IsProduction())

	// Reset rate limit on successful login
	rateLimiter.Reset(ip)

	// Log successful login
	audit.LogSuccess(req.Username, ip, "login", "/api/login")
	log.Printf("Login successful: %s from %s", req.Username, ip)

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Login successful",
	})
}

// LogoutHandler handles logout requests
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Get session info for audit logging
	var username string
	sessionID, err := auth.GetSessionCookie(r)
	if err == nil {
		if session, err := sessionStore.GetSession(sessionID); err == nil {
			username = session.Username
		}
		// Delete session
		sessionStore.DeleteSession(sessionID)
	}

	// Clear session cookie
	auth.ClearSessionCookie(w)

	// Log logout
	if username != "" {
		ip := getClientIP(r)
		audit.LogSuccess(username, ip, "logout", "/api/logout")
	}

	// Return success for API requests
	if r.Header.Get("Accept") == "application/json" || strings.HasPrefix(r.URL.Path, "/api/") {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Logged out successfully",
		})
		return
	}

	// Redirect to login page for HTML requests
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// AuthStatusHandler returns the current authentication status
func AuthStatusHandler(w http.ResponseWriter, r *http.Request) {
	sessionID, err := auth.GetSessionCookie(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"authenticated": false,
		})
		return
	}

	session, err := sessionStore.GetSession(sessionID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"authenticated": false,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"authenticated": true,
		"username":      session.Username,
		"expiresAt":     session.ExpiresAt,
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
