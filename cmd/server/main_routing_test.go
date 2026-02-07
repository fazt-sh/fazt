package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fazt-sh/fazt/internal/auth"
	"github.com/fazt-sh/fazt/internal/config"
	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/handlers"
	"github.com/fazt-sh/fazt/internal/hosting"

	_ "modernc.org/sqlite"
)

// Test infrastructure

func setupRoutingTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Minimal schema for routing tests
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
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS files (
		site_id TEXT NOT NULL,
		path TEXT NOT NULL,
		content BLOB,
		size_bytes INTEGER NOT NULL,
		mime_type TEXT,
		hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		app_id TEXT,
		PRIMARY KEY (site_id, path)
	);
	CREATE INDEX IF NOT EXISTS idx_files_app_id ON files(app_id);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	database.SetDB(db)

	// Initialize hosting system (VFS)
	if err := hosting.Init(db); err != nil {
		t.Fatalf("Failed to initialize hosting: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
		database.SetDB(nil)
	})

	return db
}

// setupTestHandlers initializes handler globals (auth service, rate limiter)
func setupTestHandlers(t *testing.T, authService *auth.Service) {
	t.Helper()

	// Initialize rate limiter
	rateLimiter := auth.NewRateLimiter()

	// Initialize auth handlers
	handlers.InitAuth(authService, rateLimiter, "test")
}

func setupRoutingTestConfig(t *testing.T) *config.Config {
	t.Helper()

	cfg := &config.Config{
		Server: config.ServerConfig{
			Domain: "test.local",
			Port:   "8080",
			Env:    "test",
		},
		Auth: config.AuthConfig{
			Username:     "admin",
			PasswordHash: "$2a$10$test",
		},
	}

	config.SetConfig(cfg)
	return cfg
}

func createTestUser(t *testing.T, authService *auth.Service, email, role string) string {
	t.Helper()

	// Create user via auth service (generates UUID, etc.)
	user, err := authService.CreateUser(email, "Test User", "", "test", nil)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Update role if needed (CreateUser defaults to 'user')
	if role != "user" {
		db := database.GetDB()
		_, err = db.Exec(`UPDATE auth_users SET role = ? WHERE id = ?`, role, user.ID)
		if err != nil {
			t.Fatalf("Failed to update user role: %v", err)
		}
	}

	return user.ID
}

func createTestSession(t *testing.T, authService *auth.Service, userID string) string {
	t.Helper()

	// Use auth service to create session properly (handles token hashing, etc.)
	token, err := authService.CreateSession(userID)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	return token
}

// Test Cases

// ============================================================================
// 1. Host Routing Tests
// ============================================================================

func TestRouting_AdminDomain_APIBypass(t *testing.T) {
	db := setupRoutingTestDB(t)
	cfg := setupRoutingTestConfig(t)

	authService := auth.NewService(db, cfg.Server.Domain, false)
	authHandler := auth.NewHandler(authService)

	dashboardMux := http.NewServeMux()
	dashboardMux.HandleFunc("/api/deploy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("deploy"))
	})
	dashboardMux.HandleFunc("/api/cmd", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("cmd"))
	})
	dashboardMux.HandleFunc("/api/sql", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("sql"))
	})
	dashboardMux.HandleFunc("/api/upgrade", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("upgrade"))
	})
	dashboardMux.HandleFunc("/api/system/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("health"))
	})
	dashboardMux.HandleFunc("/api/system/logs", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("logs"))
	})
	dashboardMux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("users"))
	})
	dashboardMux.HandleFunc("/api/aliases/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("aliases"))
	})
	dashboardMux.HandleFunc("/api/apps/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/status") {
			w.WriteHeader(200)
			w.Write([]byte("status"))
		} else {
			// This should NOT be reached without session
			w.WriteHeader(401)
			w.Write([]byte("unauthorized"))
		}
	})

	rootHandler := createRootHandler(cfg, dashboardMux, authHandler)

	// Test bypass endpoints (no session required)
	bypassPaths := []string{
		"/api/deploy",
		"/api/cmd",
		"/api/sql",
		"/api/upgrade",
		"/api/system/health",
		"/api/system/logs",
		"/api/users/list",
		"/api/aliases/list",
		"/api/apps/test123/status",
	}

	for _, path := range bypassPaths {
		t.Run("Bypass_"+path, func(t *testing.T) {
			req := httptest.NewRequest("GET", path, nil)
			req.Host = "admin.test.local"

			rr := httptest.NewRecorder()
			rootHandler.ServeHTTP(rr, req)

			// Should succeed (200) or fail with app-specific error, NOT middleware auth failure
			if rr.Code == 401 && strings.Contains(rr.Body.String(), "Authentication required") {
				t.Errorf("%s hit middleware auth check instead of bypass", path)
			}
		})
	}

	// Cleanup
	db.Close()
}

func TestRouting_AdminDomain_AdminMiddleware(t *testing.T) {
	db := setupRoutingTestDB(t)
	cfg := setupRoutingTestConfig(t)

	authService := auth.NewService(db, cfg.Server.Domain, false)
	authHandler := auth.NewHandler(authService)

	// Create test users
	adminUserID := createTestUser(t, authService, "admin@test.local", "admin")
	regularUserID := createTestUser(t, authService, "user@test.local", "user")

	// Create sessions
	adminSession := createTestSession(t, authService, adminUserID)
	userSession := createTestSession(t, authService, regularUserID)

	dashboardMux := http.NewServeMux()
	dashboardMux.HandleFunc("/api/apps", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("apps"))
	})
	dashboardMux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("stats"))
	})

	rootHandler := createRootHandler(cfg, dashboardMux, authHandler)

	adminProtectedPaths := []string{
		"/api/apps",
		"/api/stats",
	}

	for _, path := range adminProtectedPaths {
		t.Run("NoAuth_"+path, func(t *testing.T) {
			req := httptest.NewRequest("GET", path, nil)
			req.Host = "admin.test.local"

			rr := httptest.NewRecorder()
			rootHandler.ServeHTTP(rr, req)

			if rr.Code != 401 {
				t.Errorf("Expected 401 without auth, got %d", rr.Code)
			}
			if !strings.Contains(rr.Body.String(), "Authentication required") {
				t.Errorf("Expected auth error message, got: %s", rr.Body.String())
			}
		})

		t.Run("UserRole_"+path, func(t *testing.T) {
			req := httptest.NewRequest("GET", path, nil)
			req.Host = "admin.test.local"
			req.AddCookie(&http.Cookie{
				Name:  "fazt_session",
				Value: userSession,
			})

			rr := httptest.NewRecorder()
			rootHandler.ServeHTTP(rr, req)

			if rr.Code != 403 {
				t.Errorf("Expected 403 for user role, got %d", rr.Code)
			}
			if !strings.Contains(rr.Body.String(), "Admin or owner role required") {
				t.Errorf("Expected role error message, got: %s", rr.Body.String())
			}
		})

		t.Run("AdminRole_"+path, func(t *testing.T) {
			req := httptest.NewRequest("GET", path, nil)
			req.Host = "admin.test.local"
			req.AddCookie(&http.Cookie{
				Name:  "fazt_session",
				Value: adminSession,
			})

			rr := httptest.NewRecorder()
			rootHandler.ServeHTTP(rr, req)

			if rr.Code != 200 {
				t.Errorf("Expected 200 for admin role, got %d", rr.Code)
			}
		})
	}

	// Cleanup
	db.Close()
}

func TestRouting_AdminDomain_TrackEndpoint(t *testing.T) {
	db := setupRoutingTestDB(t)
	cfg := setupRoutingTestConfig(t)

	authService := auth.NewService(db, cfg.Server.Domain, false)
	authHandler := auth.NewHandler(authService)

	dashboardMux := http.NewServeMux()
	dashboardMux.HandleFunc("/track", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("tracked"))
	})

	rootHandler := createRootHandler(cfg, dashboardMux, authHandler)

	// /track should be public (no auth required)
	req := httptest.NewRequest("POST", "/track", nil)
	req.Host = "admin.test.local"

	rr := httptest.NewRecorder()
	rootHandler.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Errorf("Expected 200 for /track, got %d", rr.Code)
	}
	if rr.Body.String() != "tracked" {
		t.Errorf("Expected 'tracked', got: %s", rr.Body.String())
	}

	// Cleanup
	db.Close()
}

func TestRouting_LocalhostSpecialCase(t *testing.T) {
	db := setupRoutingTestDB(t)
	cfg := setupRoutingTestConfig(t)

	authService := auth.NewService(db, cfg.Server.Domain, false)
	authHandler := auth.NewHandler(authService)

	// Create admin user and session
	adminUserID := createTestUser(t, authService, "admin@test.local", "admin")
	adminSession := createTestSession(t, authService, adminUserID)

	dashboardMux := http.NewServeMux()
	dashboardMux.HandleFunc("/api/apps", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("apps"))
	})

	rootHandler := createRootHandler(cfg, dashboardMux, authHandler)

	t.Run("Localhost_NoAuth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/apps", nil)
		req.Host = "localhost"

		rr := httptest.NewRecorder()
		rootHandler.ServeHTTP(rr, req)

		// localhost should go through AuthMiddleware and require session
		if rr.Code != 401 {
			t.Errorf("Expected 401 for localhost without session, got %d", rr.Code)
		}
	})

	t.Run("Localhost_WithSession", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/apps", nil)
		req.Host = "localhost"
		req.AddCookie(&http.Cookie{
			Name:  "fazt_session",
			Value: adminSession,
		})

		rr := httptest.NewRecorder()
		rootHandler.ServeHTTP(rr, req)

		// With session, should reach handler
		if rr.Code != 200 {
			t.Errorf("Expected 200 for localhost with session, got %d", rr.Code)
		}
	})

	// Cleanup
	db.Close()
}

func TestRouting_RootDomain(t *testing.T) {
	db := setupRoutingTestDB(t)
	cfg := setupRoutingTestConfig(t)

	authService := auth.NewService(db, cfg.Server.Domain, false)
	authHandler := auth.NewHandler(authService)

	// Note: "root" site is automatically seeded by hosting.Init() in setupRoutingTestDB

	dashboardMux := http.NewServeMux()
	rootHandler := createRootHandler(cfg, dashboardMux, authHandler)

	testCases := []struct {
		host string
		name string
	}{
		{"root.test.local", "root_subdomain"},
		{"test.local", "bare_domain"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Host = tc.host

			rr := httptest.NewRecorder()
			rootHandler.ServeHTTP(rr, req)

			// Should serve root site (200) or 404 if not found, NOT auth error
			if rr.Code == 401 || rr.Code == 403 {
				t.Errorf("Root domain should not require auth, got %d", rr.Code)
			}
		})
	}

	// Cleanup
	db.Close()
}

func TestRouting_404Domain(t *testing.T) {
	db := setupRoutingTestDB(t)
	cfg := setupRoutingTestConfig(t)

	authService := auth.NewService(db, cfg.Server.Domain, false)
	authHandler := auth.NewHandler(authService)

	// Note: "404" site is automatically seeded by hosting.Init() in setupRoutingTestDB

	dashboardMux := http.NewServeMux()
	rootHandler := createRootHandler(cfg, dashboardMux, authHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "404.test.local"

	rr := httptest.NewRecorder()
	rootHandler.ServeHTTP(rr, req)

	// Should serve 404 site, NOT require auth
	if rr.Code == 401 || rr.Code == 403 {
		t.Errorf("404 domain should not require auth, got %d", rr.Code)
	}

	// Cleanup
	db.Close()
}

func TestRouting_SubdomainRouting(t *testing.T) {
	db := setupRoutingTestDB(t)
	cfg := setupRoutingTestConfig(t)

	authService := auth.NewService(db, cfg.Server.Domain, false)
	authHandler := auth.NewHandler(authService)

	// Create subdomain site
	_, err := db.Exec(`INSERT INTO files (site_id, path, content, size_bytes, mime_type, hash) VALUES (?, ?, ?, ?, ?, ?)`,
		"myapp", "index.html", []byte("<html>my app</html>"), 20, "text/html", "test-hash-myapp")
	if err != nil {
		t.Fatalf("Failed to create subdomain site: %v", err)
	}

	dashboardMux := http.NewServeMux()
	rootHandler := createRootHandler(cfg, dashboardMux, authHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "myapp.test.local"

	rr := httptest.NewRecorder()
	rootHandler.ServeHTTP(rr, req)

	// Should serve subdomain site, NOT require auth
	if rr.Code == 401 || rr.Code == 403 {
		t.Errorf("Subdomain should not require auth, got %d", rr.Code)
	}

	// Cleanup
	db.Close()
}

// ============================================================================
// 2. Local-Only Routes (/_app/<id>/)
// ============================================================================

func TestRouting_LocalOnlyRoutes_FromLocal(t *testing.T) {
	db := setupRoutingTestDB(t)
	cfg := setupRoutingTestConfig(t)

	authService := auth.NewService(db, cfg.Server.Domain, false)
	authHandler := auth.NewHandler(authService)

	// Create app
	_, err := db.Exec(`INSERT INTO files (site_id, path, content, size_bytes, mime_type, hash) VALUES (?, ?, ?, ?, ?, ?)`,
		"app_test123", "index.html", []byte("<html>app</html>"), 17, "text/html", "test-hash-app")
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	dashboardMux := http.NewServeMux()
	rootHandler := createRootHandler(cfg, dashboardMux, authHandler)

	// Simulate local request (127.0.0.1, ::1, 10.*, 192.168.*, etc.)
	localIPs := []string{
		"127.0.0.1:12345",
		"[::1]:12345",
		"192.168.1.10:12345",
		"10.0.0.5:12345",
	}

	for _, remoteAddr := range localIPs {
		t.Run("Local_"+remoteAddr, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/_app/app_test123/", nil)
			req.Host = "test.local"
			req.RemoteAddr = remoteAddr

			rr := httptest.NewRecorder()
			rootHandler.ServeHTTP(rr, req)

			// Should allow access from local IPs
			if rr.Code == 404 && strings.Contains(rr.Body.String(), "page not found") {
				// This means routing worked but app has no file at /
				// That's OK - routing succeeded
			} else if rr.Code == 401 || rr.Code == 403 {
				t.Errorf("Local IP should have access, got %d", rr.Code)
			}
		})
	}

	// Cleanup
	db.Close()
}

func TestRouting_LocalOnlyRoutes_FromPublic(t *testing.T) {
	db := setupRoutingTestDB(t)
	cfg := setupRoutingTestConfig(t)

	authService := auth.NewService(db, cfg.Server.Domain, false)
	authHandler := auth.NewHandler(authService)

	// Create app
	_, err := db.Exec(`INSERT INTO files (site_id, path, content, size_bytes, mime_type, hash) VALUES (?, ?, ?, ?, ?, ?)`,
		"app_test123", "index.html", []byte("<html>app</html>"), 17, "text/html", "test-hash-app")
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	dashboardMux := http.NewServeMux()
	rootHandler := createRootHandler(cfg, dashboardMux, authHandler)

	// Simulate public IPs
	publicIPs := []string{
		"1.2.3.4:12345",
		"203.0.113.5:12345",
		"2001:db8::1:12345",
	}

	for _, remoteAddr := range publicIPs {
		t.Run("Public_"+remoteAddr, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/_app/app_test123/", nil)
			req.Host = "test.local"
			req.RemoteAddr = remoteAddr

			rr := httptest.NewRecorder()
			rootHandler.ServeHTTP(rr, req)

			// Should return 404 (not reveal route exists)
			if rr.Code != 404 {
				t.Errorf("Expected 404 for public IP, got %d", rr.Code)
			}
		})
	}

	// Cleanup
	db.Close()
}

// ============================================================================
// 3. Auth Routes (/auth/*)
// ============================================================================

func TestRouting_AuthRoutes_AvailableEverywhere(t *testing.T) {
	db := setupRoutingTestDB(t)
	cfg := setupRoutingTestConfig(t)

	authService := auth.NewService(db, cfg.Server.Domain, false)
	authHandler := auth.NewHandler(authService)

	dashboardMux := http.NewServeMux()
	rootHandler := createRootHandler(cfg, dashboardMux, authHandler)

	// /auth/* routes should be available on all hosts
	hosts := []string{
		"admin.test.local",
		"root.test.local",
		"test.local",
		"myapp.test.local",
		"localhost",
	}

	for _, host := range hosts {
		t.Run("AuthRoute_"+host, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/auth/callback", nil)
			req.Host = host

			rr := httptest.NewRecorder()
			rootHandler.ServeHTTP(rr, req)

			// Should reach auth handler (NOT 401/403 from middleware)
			// Auth handler may return various codes depending on OAuth state
			if rr.Code == 401 && strings.Contains(rr.Body.String(), "Authentication required") {
				t.Errorf("/auth/* route should not hit auth middleware on %s", host)
			}
		})
	}

	// Cleanup
	db.Close()
}

func TestRouting_LoginRoute_PostOnly(t *testing.T) {
	db := setupRoutingTestDB(t)
	cfg := setupRoutingTestConfig(t)

	authService := auth.NewService(db, cfg.Server.Domain, false)
	authHandler := auth.NewHandler(authService)
	setupTestHandlers(t, authService)

	dashboardMux := http.NewServeMux()
	rootHandler := createRootHandler(cfg, dashboardMux, authHandler)

	// POST /auth/login should go to LoginHandler
	req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(`{"email":"test@test.com","password":"pass"}`))
	req.Host = "test.local"
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	rootHandler.ServeHTTP(rr, req)

	// Should reach LoginHandler, not auth middleware
	// LoginHandler will fail with invalid credentials, but routing should work
	if rr.Code == 401 && strings.Contains(rr.Body.String(), "Authentication required") {
		t.Errorf("POST /auth/login should not hit auth middleware")
	}

	// Cleanup
	db.Close()
}

// ============================================================================
// 4. Middleware Order Tests
// ============================================================================

func TestRouting_MiddlewareOrder_AuthBeforeAdmin(t *testing.T) {
	db := setupRoutingTestDB(t)
	cfg := setupRoutingTestConfig(t)

	authService := auth.NewService(db, cfg.Server.Domain, false)
	authHandler := auth.NewHandler(authService)

	dashboardMux := http.NewServeMux()
	dashboardMux.HandleFunc("/api/apps", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("apps"))
	})

	rootHandler := createRootHandler(cfg, dashboardMux, authHandler)

	// Request to admin-protected endpoint without any auth
	req := httptest.NewRequest("GET", "/api/apps", nil)
	req.Host = "admin.test.local"

	rr := httptest.NewRecorder()
	rootHandler.ServeHTTP(rr, req)

	// Should get 401 from AdminMiddleware (which checks auth first)
	if rr.Code != 401 {
		t.Errorf("Expected 401 from AdminMiddleware, got %d", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Authentication required") {
		t.Errorf("Expected auth error from AdminMiddleware, got: %s", body)
	}

	// Cleanup
	db.Close()
}

// ============================================================================
// 5. Port Stripping Tests
// ============================================================================

func TestRouting_PortStripping(t *testing.T) {
	db := setupRoutingTestDB(t)
	cfg := setupRoutingTestConfig(t)

	authService := auth.NewService(db, cfg.Server.Domain, false)
	authHandler := auth.NewHandler(authService)

	adminUserID := createTestUser(t, authService, "admin@test.local", "admin")
	adminSession := createTestSession(t, authService, adminUserID)

	dashboardMux := http.NewServeMux()
	dashboardMux.HandleFunc("/api/apps", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("apps"))
	})

	rootHandler := createRootHandler(cfg, dashboardMux, authHandler)

	// Hosts with ports should have port stripped
	hostsWithPorts := []struct {
		host string
		name string
	}{
		{"admin.test.local:8080", "standard_port"},
		{"admin.test.local:443", "https_port"},
		{"admin.test.local:3000", "dev_port"},
	}

	for _, tc := range hostsWithPorts {
		t.Run("PortStrip_"+tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/apps", nil)
			req.Host = tc.host
			req.AddCookie(&http.Cookie{
				Name:  "fazt_session",
				Value: adminSession,
			})

			rr := httptest.NewRecorder()
			rootHandler.ServeHTTP(rr, req)

			// Should route correctly (port stripped)
			if rr.Code != 200 {
				t.Errorf("Port stripping failed for %s, got %d", tc.host, rr.Code)
			}
		})
	}

	// Cleanup
	db.Close()
}

func TestRouting_IPv6_PortStripping(t *testing.T) {
	db := setupRoutingTestDB(t)
	cfg := setupRoutingTestConfig(t)

	authService := auth.NewService(db, cfg.Server.Domain, false)
	authHandler := auth.NewHandler(authService)

	adminUserID := createTestUser(t, authService, "admin@test.local", "admin")
	adminSession := createTestSession(t, authService, adminUserID)

	dashboardMux := http.NewServeMux()
	dashboardMux.HandleFunc("/api/apps", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("apps"))
	})

	rootHandler := createRootHandler(cfg, dashboardMux, authHandler)

	// IPv6 with port - should NOT strip port (brackets indicate IPv6)
	req := httptest.NewRequest("GET", "/api/apps", nil)
	req.Host = "[::1]:8080"
	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: adminSession,
	})

	rr := httptest.NewRecorder()
	rootHandler.ServeHTTP(rr, req)

	// IPv6 routing may not match "admin.test.local", but should not crash
	// This test verifies the port-stripping logic doesn't break on IPv6
	if rr.Code == 500 {
		t.Errorf("IPv6 port handling caused server error")
	}

	// Cleanup
	db.Close()
}

// ============================================================================
// 6. Edge Cases
// ============================================================================

func TestRouting_EmptyHost(t *testing.T) {
	db := setupRoutingTestDB(t)
	cfg := setupRoutingTestConfig(t)

	authService := auth.NewService(db, cfg.Server.Domain, false)
	authHandler := auth.NewHandler(authService)

	dashboardMux := http.NewServeMux()
	rootHandler := createRootHandler(cfg, dashboardMux, authHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "" // Empty host

	rr := httptest.NewRecorder()
	rootHandler.ServeHTTP(rr, req)

	// Should not crash, likely 404
	if rr.Code == 500 {
		t.Errorf("Empty host caused server error")
	}

	// Cleanup
	db.Close()
}

func TestRouting_UnknownSubdomain_Fallback(t *testing.T) {
	db := setupRoutingTestDB(t)
	cfg := setupRoutingTestConfig(t)

	authService := auth.NewService(db, cfg.Server.Domain, false)
	authHandler := auth.NewHandler(authService)

	dashboardMux := http.NewServeMux()
	rootHandler := createRootHandler(cfg, dashboardMux, authHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "nonexistent.test.local"

	rr := httptest.NewRecorder()
	rootHandler.ServeHTTP(rr, req)

	// Should serve 404 site or fallback, not crash
	if rr.Code == 500 {
		t.Errorf("Unknown subdomain caused server error")
	}

	// Cleanup
	db.Close()
}

func TestRouting_CaseSensitivity(t *testing.T) {
	db := setupRoutingTestDB(t)
	cfg := setupRoutingTestConfig(t)

	authService := auth.NewService(db, cfg.Server.Domain, false)
	authHandler := auth.NewHandler(authService)

	adminUserID := createTestUser(t, authService, "admin@test.local", "admin")
	adminSession := createTestSession(t, authService, adminUserID)

	dashboardMux := http.NewServeMux()
	dashboardMux.HandleFunc("/api/apps", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("apps"))
	})

	rootHandler := createRootHandler(cfg, dashboardMux, authHandler)

	// Test case variations of admin domain
	caseVariations := []string{
		"admin.test.local",   // lowercase (expected)
		"ADMIN.test.local",   // uppercase
		"Admin.test.local",   // mixed case
		"AdMiN.TEST.local",   // chaotic case
	}

	for _, host := range caseVariations {
		t.Run("Case_"+host, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/apps", nil)
			req.Host = host
			req.AddCookie(&http.Cookie{
				Name:  "fazt_session",
				Value: adminSession,
			})

			rr := httptest.NewRecorder()
			rootHandler.ServeHTTP(rr, req)

			// Behavior depends on case sensitivity in extractDomain/extractSubdomain
			// Document current behavior (likely case-sensitive)
			if rr.Code == 200 {
				t.Logf("Case variation %s routed correctly", host)
			} else {
				t.Logf("Case variation %s did NOT route to admin (code: %d)", host, rr.Code)
			}
		})
	}

	// Cleanup
	db.Close()
}

// ============================================================================
// 7. Path Precedence Tests
// ============================================================================

func TestRouting_PathPrecedence_BypassBeforeAdmin(t *testing.T) {
	db := setupRoutingTestDB(t)
	cfg := setupRoutingTestConfig(t)

	authService := auth.NewService(db, cfg.Server.Domain, false)
	authHandler := auth.NewHandler(authService)

	dashboardMux := http.NewServeMux()
	dashboardMux.HandleFunc("/api/deploy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("deploy"))
	})

	rootHandler := createRootHandler(cfg, dashboardMux, authHandler)

	// /api/deploy should bypass AdminMiddleware even though it starts with /api/
	req := httptest.NewRequest("POST", "/api/deploy", nil)
	req.Host = "admin.test.local"

	rr := httptest.NewRecorder()
	rootHandler.ServeHTTP(rr, req)

	// Should NOT get 401 from AdminMiddleware
	if rr.Code == 401 && strings.Contains(rr.Body.String(), "Authentication required") {
		t.Errorf("/api/deploy hit AdminMiddleware instead of bypass list")
	}

	// Cleanup
	db.Close()
}

func TestRouting_PathPrecedence_AppsStatusBypass(t *testing.T) {
	db := setupRoutingTestDB(t)
	cfg := setupRoutingTestConfig(t)

	authService := auth.NewService(db, cfg.Server.Domain, false)
	authHandler := auth.NewHandler(authService)

	dashboardMux := http.NewServeMux()
	dashboardMux.HandleFunc("/api/apps/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/status") {
			w.WriteHeader(200)
			w.Write([]byte("status"))
		} else {
			w.WriteHeader(401)
			w.Write([]byte("unauthorized"))
		}
	})

	rootHandler := createRootHandler(cfg, dashboardMux, authHandler)

	t.Run("StatusEndpoint_Bypass", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/apps/app123/status", nil)
		req.Host = "admin.test.local"

		rr := httptest.NewRecorder()
		rootHandler.ServeHTTP(rr, req)

		// Should bypass AdminMiddleware
		if rr.Code != 200 {
			t.Errorf("Expected 200 for /status bypass, got %d", rr.Code)
		}
	})

	t.Run("NonStatusEndpoint_AdminRequired", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/apps/app123/info", nil)
		req.Host = "admin.test.local"

		rr := httptest.NewRecorder()
		rootHandler.ServeHTTP(rr, req)

		// Should require AdminMiddleware (401 without session)
		if rr.Code != 401 {
			t.Errorf("Expected 401 for non-status endpoint, got %d", rr.Code)
		}
	})

	// Cleanup
	db.Close()
}

// ============================================================================
// 8. Admin Domain Fallthrough Tests
// ============================================================================

func TestRouting_AdminDomain_Fallthrough(t *testing.T) {
	db := setupRoutingTestDB(t)
	cfg := setupRoutingTestConfig(t)

	authService := auth.NewService(db, cfg.Server.Domain, false)
	authHandler := auth.NewHandler(authService)

	// Create "admin" site in VFS
	_, err := db.Exec(`INSERT INTO files (site_id, path, content, size_bytes, mime_type, hash) VALUES (?, ?, ?, ?, ?, ?)`,
		"admin", "index.html", []byte("<html>admin app</html>"), 23, "text/html", "test-hash-admin")
	if err != nil {
		t.Fatalf("Failed to create admin site: %v", err)
	}

	dashboardMux := http.NewServeMux()
	rootHandler := createRootHandler(cfg, dashboardMux, authHandler)

	// Non-API, non-track paths on admin.* should fall through to app serving
	req := httptest.NewRequest("GET", "/dashboard", nil)
	req.Host = "admin.test.local"

	rr := httptest.NewRecorder()
	rootHandler.ServeHTTP(rr, req)

	// Should NOT hit middleware, should try to serve from VFS
	if rr.Code == 401 && strings.Contains(rr.Body.String(), "Authentication required") {
		t.Errorf("Non-API path on admin.* should fall through to app serving, not hit middleware")
	}

	// Cleanup
	db.Close()
}
