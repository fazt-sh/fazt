package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/jikku/command-center/internal/assets"
	"github.com/jikku/command-center/internal/audit"
	"github.com/jikku/command-center/internal/auth"
	"github.com/jikku/command-center/internal/config"
)

var (
	sessionStore *auth.SessionStore
	rateLimiter  *auth.RateLimiter
)

// InitAuth initializes the auth handlers with session store and rate limiter
func InitAuth(store *auth.SessionStore, limiter *auth.RateLimiter) {
	sessionStore = store
	rateLimiter = limiter
}

// LoginPageHandler serves the login page
func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	// If already logged in, redirect to dashboard
	if sessionID, err := auth.GetSessionCookie(r); err == nil {
		if valid, _ := sessionStore.ValidateSession(sessionID); valid {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	// Render login page
	tmpl, err := template.ParseFS(assets.WebFS, "web/templates/login.html")
	if err != nil {
		log.Printf("Error loading login template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Error": r.URL.Query().Get("error"),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error rendering login template: %v", err)
	}
}

// LoginHandler handles login requests
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
