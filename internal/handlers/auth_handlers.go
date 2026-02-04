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
	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/hosting"
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

// requireAPIKeyAuth validates API key from Authorization header
func requireAPIKeyAuth(w http.ResponseWriter, r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		api.Unauthorized(w, "API key required (Authorization: Bearer <token>)")
		return false
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	_, _, err := hosting.ValidateAPIKey(database.GetDB(), token)
	if err != nil {
		api.InvalidAPIKey(w)
		return false
	}
	return true
}

// requireAdminAuth allows EITHER API key auth OR session auth with admin/owner role
// Returns the authenticated user's role (or "owner" for API key auth)
func requireAdminAuth(w http.ResponseWriter, r *http.Request) (role string, ok bool) {
	// Try API key auth first (CLI usage) - API key holders are treated as owners
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		_, _, err := hosting.ValidateAPIKey(database.GetDB(), token)
		if err == nil {
			return "owner", true // API key holders have full access
		}
	}

	// Try session auth (UI usage)
	user, err := authService.GetSessionFromRequest(r)
	if err != nil {
		api.Unauthorized(w, "Authentication required")
		return "", false
	}

	// Check role
	if user.Role != "admin" && user.Role != "owner" {
		api.Error(w, http.StatusForbidden, "FORBIDDEN", "Admin or owner role required", nil)
		return "", false
	}

	return user.Role, true
}

// UsersListHandler returns a list of all users
func UsersListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.BadRequest(w, "Method not allowed")
		return
	}

	// Require admin auth (API key or session with admin/owner role)
	if _, ok := requireAdminAuth(w, r); !ok {
		return
	}

	users, err := authService.ListUsers()
	if err != nil {
		api.InternalError(w, err)
		return
	}

	api.Success(w, http.StatusOK, users)
}

// UserSetRoleHandler sets a user's role
// POST /api/users/role
// Body: { "user_id": "...", "role": "admin|owner|user" }
//
// RBAC rules:
//   - owner: can set any role on any user
//   - admin: can set user/admin roles, but NOT owner; cannot modify owners
func UserSetRoleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.BadRequest(w, "Method not allowed")
		return
	}

	// Require admin auth and get caller's role
	callerRole, ok := requireAdminAuth(w, r)
	if !ok {
		return
	}

	var req struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"` // Alternative to user_id
		Role   string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.InvalidJSON(w, "Invalid request body")
		return
	}

	// Validate role
	if req.Role != "user" && req.Role != "admin" && req.Role != "owner" {
		api.BadRequest(w, "Invalid role. Must be: user, admin, or owner")
		return
	}

	// Get target user by ID or email
	var targetUser *auth.User
	var err error
	if req.UserID != "" {
		targetUser, err = authService.GetUserByID(req.UserID)
	} else if req.Email != "" {
		targetUser, err = authService.GetUserByEmail(req.Email)
	} else {
		api.BadRequest(w, "Must provide user_id or email")
		return
	}

	if err != nil {
		api.NotFound(w, "USER_NOT_FOUND", "User not found")
		return
	}

	// RBAC checks for non-owners
	if callerRole != "owner" {
		// Admins cannot promote to owner
		if req.Role == "owner" {
			api.Error(w, http.StatusForbidden, "FORBIDDEN", "Only owners can promote users to owner role", nil)
			return
		}
		// Admins cannot modify owners
		if targetUser.Role == "owner" {
			api.Error(w, http.StatusForbidden, "FORBIDDEN", "Only owners can modify other owners", nil)
			return
		}
	}

	// Update role
	if err := authService.UpdateUserRole(targetUser.ID, req.Role); err != nil {
		api.InternalError(w, err)
		return
	}

	log.Printf("User role updated: %s (%s) -> %s (by %s)", targetUser.Email, targetUser.ID, req.Role, callerRole)

	api.Success(w, http.StatusOK, map[string]interface{}{
		"message": "Role updated successfully",
		"user_id": targetUser.ID,
		"email":   targetUser.Email,
		"role":    req.Role,
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
