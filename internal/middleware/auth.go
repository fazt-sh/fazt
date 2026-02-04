package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/fazt-sh/fazt/internal/auth"
	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/hosting"
)

// AuthMiddleware checks if a user is authenticated before allowing access to protected routes
// Uses database-backed sessions via auth.Service
func AuthMiddleware(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if the path requires authentication
			if !requiresAuth(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// 1. Check for Bearer Token (API Access)
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				db := database.GetDB()
				if db != nil {
					_, _, err := hosting.ValidateAPIKey(db, token)
					if err == nil {
						// Token is valid
						next.ServeHTTP(w, r)
						return
					}
					log.Printf("Invalid API Token: %v", err)
					redirectToLogin(w, r)
					return
				}
			}

			// 2. Check database-backed session
			user, err := authService.GetSessionFromRequest(r)
			if err == nil && user != nil {
				next.ServeHTTP(w, r)
				return
			}

			// No valid session found
			log.Printf("No valid session for %s %s", r.Method, r.URL.Path)
			redirectToLogin(w, r)
		})
	}
}

// requiresAuth returns true if the path requires authentication
func requiresAuth(path string) bool {
	// Exact match paths (no prefix matching)
	exactPaths := map[string]bool{
		"/":                     true,
		"/index.html":           true,
		"/login.html":           true,
		"/manifest.webmanifest": true,
		"/registerSW.js":        true,
		"/sw.js":                true,
		"/favicon.png":          true,
		"/favicon.ico":          true,
		"/logo.png":             true,
		"/vite.svg":             true,
		"/health":               true,
	}

	if exactPaths[path] {
		return false
	}

	// Prefix match paths
	publicPrefixes := []string{
		"/track",
		"/pixel.gif",
		"/r/",
		"/webhook/",
		"/static/",
		"/assets/",
		"/workbox-",
		"/api/login",
		"/api/deploy",
		"/auth/login",
		"/auth/",
	}

	for _, prefix := range publicPrefixes {
		if strings.HasPrefix(path, prefix) {
			return false
		}
	}

	// All other paths require authentication
	return true
}

// AdminMiddleware checks if a user has admin or owner role before allowing access
// This middleware should be applied to admin-only endpoints (apps, aliases, system management, etc.)
func AdminMiddleware(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user from session
			user, err := authService.GetSessionFromRequest(r)
			if err != nil || user == nil {
				// Not authenticated
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"Authentication required"}`))
				return
			}

			// Check role - must be admin or owner
			if user.Role != "admin" && user.Role != "owner" {
				// Authenticated but insufficient permissions
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"error":"Admin or owner role required","user_role":"` + user.Role + `"}`))
				log.Printf("Access denied: user %s (role: %s) attempted to access %s %s", user.Email, user.Role, r.Method, r.URL.Path)
				return
			}

			// User has required role, proceed
			next.ServeHTTP(w, r)
		})
	}
}

// redirectToLogin redirects the user to the login page
func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	// For API requests, return 401 Unauthorized
	if strings.HasPrefix(r.URL.Path, "/api/") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	// For HTML requests, redirect to login page
	http.Redirect(w, r, "/login.html", http.StatusSeeOther)
}
