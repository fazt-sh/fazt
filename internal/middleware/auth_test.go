package middleware

import (
	"database/sql"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fazt-sh/fazt/internal/auth"
	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/hosting"

	_ "modernc.org/sqlite"
)

func setupAuthMiddlewareDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS auth_users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		name TEXT,
		picture TEXT,
		provider TEXT NOT NULL,
		provider_id TEXT,
		password_hash TEXT,
		role TEXT DEFAULT 'user',
		invited_by TEXT,
		created_at INTEGER NOT NULL DEFAULT (unixepoch()),
		last_login INTEGER
	);
	CREATE TABLE IF NOT EXISTS auth_sessions (
		token_hash TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		created_at INTEGER NOT NULL DEFAULT (unixepoch()),
		expires_at INTEGER NOT NULL,
		last_seen INTEGER,
		FOREIGN KEY (user_id) REFERENCES auth_users(id) ON DELETE CASCADE
	);
	CREATE TABLE IF NOT EXISTS api_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		key_hash TEXT NOT NULL,
		scopes TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_used_at DATETIME
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	database.SetDB(db)

	t.Cleanup(func() {
		db.Close()
		database.SetDB(nil)
	})

	return db
}

func setupAuthMiddlewareEnv(t *testing.T) (*sql.DB, *auth.Service) {
	t.Helper()

	log.SetOutput(io.Discard)
	t.Cleanup(func() {
		log.SetOutput(io.Discard)
	})

	db := setupAuthMiddlewareDB(t)
	service := auth.NewService(db, "test.local", false)
	return db, service
}

func createTestUser(t *testing.T, service *auth.Service, email, role string) *auth.User {
	t.Helper()

	user, err := service.CreateUser(email, "Test User", "", "test", nil)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	if role != "user" {
		if err := service.UpdateUserRole(user.ID, role); err != nil {
			t.Fatalf("Failed to update user role: %v", err)
		}
		user.Role = role
	}

	return user
}

func createTestSession(t *testing.T, service *auth.Service, userID string) string {
	t.Helper()

	token, err := service.CreateSession(userID)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	return token
}

func createTestAPIKey(t *testing.T, db *sql.DB) string {
	t.Helper()

	token, err := hosting.CreateAPIKey(db, "test-key", "deploy")
	if err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}

	return token
}

func TestAuthMiddleware_NoAuthRequired(t *testing.T) {
	_, service := setupAuthMiddlewareEnv(t)

	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	AuthMiddleware(service)(handler).ServeHTTP(rr, req)

	if !called {
		t.Fatal("Handler should be called for public path")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestAuthMiddleware_BearerToken_Valid(t *testing.T) {
	db, service := setupAuthMiddlewareEnv(t)
	apiToken := createTestAPIKey(t, db)

	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/apps", nil)
	req.Header.Set("Authorization", "Bearer "+apiToken)
	rr := httptest.NewRecorder()

	AuthMiddleware(service)(handler).ServeHTTP(rr, req)

	if !called {
		t.Fatal("Handler should be called for valid API token")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestAuthMiddleware_BearerToken_Invalid(t *testing.T) {
	_, service := setupAuthMiddlewareEnv(t)

	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/apps", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()

	AuthMiddleware(service)(handler).ServeHTTP(rr, req)

	if called {
		t.Fatal("Handler should not be called for invalid API token")
	}
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_EmptyBearerHeader(t *testing.T) {
	_, service := setupAuthMiddlewareEnv(t)

	req := httptest.NewRequest("GET", "/api/apps", nil)
	req.Header.Set("Authorization", "Bearer ")
	rr := httptest.NewRecorder()

	AuthMiddleware(service)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_MalformedBearer(t *testing.T) {
	_, service := setupAuthMiddlewareEnv(t)

	tests := []string{"Bearer", "Token abc", "BearerToken abc"}
	for _, header := range tests {
		t.Run(header, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/apps", nil)
			req.Header.Set("Authorization", header)
			rr := httptest.NewRecorder()

			AuthMiddleware(service)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})).ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestAuthMiddleware_Session_Valid(t *testing.T) {
	_, service := setupAuthMiddlewareEnv(t)

	user := createTestUser(t, service, "user@test.local", "user")
	session := createTestSession(t, service, user.ID)

	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/dashboard", nil)
	req.AddCookie(&http.Cookie{Name: "fazt_session", Value: session})
	rr := httptest.NewRecorder()

	AuthMiddleware(service)(handler).ServeHTTP(rr, req)

	if !called {
		t.Fatal("Handler should be called for valid session")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestAuthMiddleware_Session_Expired(t *testing.T) {
	db, service := setupAuthMiddlewareEnv(t)

	user := createTestUser(t, service, "expired@test.local", "user")
	session := createTestSession(t, service, user.ID)

	_, err := db.Exec("UPDATE auth_sessions SET expires_at = ?", time.Now().Add(-time.Hour).Unix())
	if err != nil {
		t.Fatalf("Failed to expire session: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/apps", nil)
	req.AddCookie(&http.Cookie{Name: "fazt_session", Value: session})
	rr := httptest.NewRecorder()

	AuthMiddleware(service)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_Session_Invalid(t *testing.T) {
	_, service := setupAuthMiddlewareEnv(t)

	req := httptest.NewRequest("GET", "/api/apps", nil)
	req.AddCookie(&http.Cookie{Name: "fazt_session", Value: "invalid-token"})
	rr := httptest.NewRecorder()

	AuthMiddleware(service)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_NoAuth(t *testing.T) {
	_, service := setupAuthMiddlewareEnv(t)

	req := httptest.NewRequest("GET", "/dashboard", nil)
	rr := httptest.NewRecorder()

	AuthMiddleware(service)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusSeeOther)
	}
	if location := rr.Header().Get("Location"); location != "/login.html" {
		t.Fatalf("Location = %q, want %q", location, "/login.html")
	}
}

func TestAuthMiddleware_APIvsHTML(t *testing.T) {
	_, service := setupAuthMiddlewareEnv(t)

	apiReq := httptest.NewRequest("GET", "/api/apps", nil)
	apiRR := httptest.NewRecorder()
	AuthMiddleware(service)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(apiRR, apiReq)
	if apiRR.Code != http.StatusUnauthorized {
		t.Fatalf("api status = %d, want %d", apiRR.Code, http.StatusUnauthorized)
	}

	htmlReq := httptest.NewRequest("GET", "/dashboard", nil)
	htmlRR := httptest.NewRecorder()
	AuthMiddleware(service)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(htmlRR, htmlReq)
	if htmlRR.Code != http.StatusSeeOther {
		t.Fatalf("html status = %d, want %d", htmlRR.Code, http.StatusSeeOther)
	}
}

func TestAdminMiddleware_NoSession(t *testing.T) {
	_, service := setupAuthMiddlewareEnv(t)

	req := httptest.NewRequest("GET", "/api/admin", nil)
	rr := httptest.NewRecorder()

	AdminMiddleware(service)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestAdminMiddleware_InvalidSession(t *testing.T) {
	_, service := setupAuthMiddlewareEnv(t)

	req := httptest.NewRequest("GET", "/api/admin", nil)
	req.AddCookie(&http.Cookie{Name: "fazt_session", Value: "invalid-token"})
	rr := httptest.NewRecorder()

	AdminMiddleware(service)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestAdminMiddleware_UserRole(t *testing.T) {
	_, service := setupAuthMiddlewareEnv(t)

	user := createTestUser(t, service, "user-role@test.local", "user")
	session := createTestSession(t, service, user.ID)

	req := httptest.NewRequest("GET", "/api/admin", nil)
	req.AddCookie(&http.Cookie{Name: "fazt_session", Value: session})
	rr := httptest.NewRecorder()

	AdminMiddleware(service)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestAdminMiddleware_AdminRole(t *testing.T) {
	_, service := setupAuthMiddlewareEnv(t)

	user := createTestUser(t, service, "admin@test.local", "admin")
	session := createTestSession(t, service, user.ID)

	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/admin", nil)
	req.AddCookie(&http.Cookie{Name: "fazt_session", Value: session})
	rr := httptest.NewRecorder()

	AdminMiddleware(service)(handler).ServeHTTP(rr, req)

	if !called {
		t.Fatal("Handler should be called for admin role")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestAdminMiddleware_OwnerRole(t *testing.T) {
	_, service := setupAuthMiddlewareEnv(t)

	user := createTestUser(t, service, "owner@test.local", "owner")
	session := createTestSession(t, service, user.ID)

	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/admin", nil)
	req.AddCookie(&http.Cookie{Name: "fazt_session", Value: session})
	rr := httptest.NewRecorder()

	AdminMiddleware(service)(handler).ServeHTTP(rr, req)

	if !called {
		t.Fatal("Handler should be called for owner role")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestRequiresAuth_PublicPaths(t *testing.T) {
	public := []string{
		"/",
		"/index.html",
		"/login.html",
		"/manifest.webmanifest",
		"/registerSW.js",
		"/sw.js",
		"/favicon.png",
		"/favicon.ico",
		"/logo.png",
		"/vite.svg",
		"/health",
	}

	for _, path := range public {
		if requiresAuth(path) {
			t.Fatalf("requiresAuth(%q) = true, want false", path)
		}
	}
}

func TestRequiresAuth_PublicPrefixes(t *testing.T) {
	public := []string{
		"/track",
		"/track/evt",
		"/pixel.gif",
		"/r/slug",
		"/webhook/test",
		"/static/app.css",
		"/assets/logo.png",
		"/workbox-123.js",
		"/api/login",
		"/api/deploy",
		"/auth/login",
		"/auth/callback",
	}

	for _, path := range public {
		if requiresAuth(path) {
			t.Fatalf("requiresAuth(%q) = true, want false", path)
		}
	}
}

func TestRequiresAuth_ProtectedPaths(t *testing.T) {
	protected := []string{
		"/api/apps",
		"/api/system",
		"/dashboard",
		"/admin",
		"/settings",
	}

	for _, path := range protected {
		if !requiresAuth(path) {
			t.Fatalf("requiresAuth(%q) = false, want true", path)
		}
	}
}

func TestRequiresAuth_CaseSensitivity(t *testing.T) {
	if !requiresAuth("/API/login") {
		t.Fatal("requiresAuth should be case-sensitive")
	}
}

func TestRequiresAuth_TrailingSlash(t *testing.T) {
	if requiresAuth("/static/") {
		t.Fatal("/static/ should be public")
	}
	if !requiresAuth("/static") {
		t.Fatal("/static should require auth")
	}
}

func TestRequiresAuth_EdgeCases(t *testing.T) {
	edge := []string{"", ".", "..", "/auth"}
	for _, path := range edge {
		if !requiresAuth(path) {
			t.Fatalf("requiresAuth(%q) = false, want true", path)
		}
	}
}

func TestRequiresAuth_PathTraversal(t *testing.T) {
	path := "/static/../api/apps"
	if requiresAuth(path) {
		t.Fatalf("requiresAuth(%q) = true, want false", path)
	}
}
