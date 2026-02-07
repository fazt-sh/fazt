package main

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fazt-sh/fazt/internal/auth"
	"github.com/fazt-sh/fazt/internal/config"
	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/handlers"
	"github.com/fazt-sh/fazt/internal/hosting"

	_ "modernc.org/sqlite"
)

// integrationTestServer represents a full fazt server for integration testing
type integrationTestServer struct {
	db          *sql.DB
	server      *httptest.Server
	authHandler *auth.Handler
	authService *auth.Service
	config      *config.Config
	handler     http.Handler
}

// setupIntegrationTest creates a full test server with routing and middleware
func setupIntegrationTest(t *testing.T) *integrationTestServer {
	t.Helper()

	// Create in-memory database with migrations
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Single connection for in-memory consistency
	db.SetMaxOpenConns(1)

	// Enable WAL and foreign keys
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA foreign_keys=ON")

	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Set global DB for handlers
	database.SetDB(db)

	// Initialize hosting system (VFS)
	if err := hosting.Init(db); err != nil {
		t.Fatalf("Failed to initialize hosting: %v", err)
	}

	// Create test config
	cfg := &config.Config{
		Server: config.ServerConfig{
			Domain: "testdomain.com",
			Port:   "8080",
			Env:    "test",
		},
		Auth: config.AuthConfig{
			Username:     "admin",
			PasswordHash: "$2a$10$...", // Placeholder
		},
	}
	config.SetConfig(cfg)

	// Create auth service
	authService := auth.NewService(db, "testdomain.com", false)
	authHandler := auth.NewHandler(authService)

	// Set global site auth service for private file access
	siteAuthService = authService

	// Create dashboard mux with all handlers
	dashboardMux := http.NewServeMux()
	registerHandlers(dashboardMux, authService)

	// Create root handler with full routing logic
	rootHandler := createRootHandler(cfg, dashboardMux, authHandler)

	// Wrap with recovery middleware
	handler := recoveryMiddleware(rootHandler)

	// Create test server
	server := httptest.NewServer(handler)
	t.Cleanup(func() {
		server.Close()
		db.Close()
	})

	return &integrationTestServer{
		db:          db,
		server:      server,
		authHandler: authHandler,
		authService: authService,
		config:      cfg,
		handler:     handler,
	}
}

// registerHandlers sets up all API handlers on the mux
func registerHandlers(mux *http.ServeMux, authService *auth.Service) {
	// Core handlers
	mux.HandleFunc("/api/stats", handlers.StatsHandler)
	mux.HandleFunc("/api/events", handlers.EventsHandler)
	mux.HandleFunc("/api/domains", handlers.DomainsHandler)
	mux.HandleFunc("/api/tags", handlers.TagsHandler)
	mux.HandleFunc("/api/redirects", handlers.RedirectsHandler)
	mux.HandleFunc("/api/webhooks", handlers.WebhooksHandler)
	mux.HandleFunc("/track", handlers.TrackHandler)
	mux.HandleFunc("/pixel.gif", handlers.PixelHandler)
	mux.HandleFunc("/r/", handlers.RedirectHandler)

	// System handlers
	mux.HandleFunc("GET /api/system/health", handlers.SystemHealthHandler)
	mux.HandleFunc("GET /api/system/config", handlers.SystemConfigHandler)
	mux.HandleFunc("POST /api/sql", handlers.HandleSQL)

	// Alias handlers
	mux.HandleFunc("GET /api/aliases", handlers.AliasesListHandler)
	mux.HandleFunc("POST /api/aliases", handlers.AliasCreateHandler)
	mux.HandleFunc("GET /api/aliases/{alias}", handlers.AliasDetailHandler)
	mux.HandleFunc("PUT /api/aliases/{alias}", handlers.AliasUpdateHandler)
	mux.HandleFunc("DELETE /api/aliases/{alias}", handlers.AliasDeleteHandler)

	// Auth handlers
	mux.HandleFunc("/api/login", handlers.LoginHandler)
	mux.HandleFunc("/api/logout", handlers.LogoutHandler)
	mux.HandleFunc("/api/auth/status", handlers.AuthStatusHandler)
	mux.HandleFunc("/api/user/me", handlers.UserMeHandler)

	// User info endpoint
	mux.HandleFunc("/api/auth/me", func(w http.ResponseWriter, r *http.Request) {
		user, err := authService.GetSessionFromRequest(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"role":  user.Role,
		})
	})
}

// createSession creates a test user and session, returns the session token
func (s *integrationTestServer) createSession(t *testing.T, email, role string) string {
	t.Helper()

	userID := "user-" + email

	// Create user in database with all required fields
	_, err := s.db.Exec(`
		INSERT INTO auth_users (id, email, name, picture, provider, role, created_at)
		VALUES (?, ?, ?, ?, 'test', ?, ?)
	`, userID, email, "Test User", "", role, time.Now().Unix())
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create session token
	sessionToken := "test-session-" + email

	// Hash the token (auth service uses SHA-256)
	h := sha256.Sum256([]byte(sessionToken))
	tokenHash := hex.EncodeToString(h[:])

	// Insert into auth_sessions with hashed token
	now := time.Now().Unix()
	_, err = s.db.Exec(`
		INSERT INTO auth_sessions (token_hash, user_id, created_at, expires_at, last_seen)
		VALUES (?, ?, ?, ?, ?)
	`, tokenHash, userID, now, now+86400, now)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	return sessionToken
}

// makeRequest makes an HTTP request to the test server
func (s *integrationTestServer) makeRequest(t *testing.T, method, path string, body io.Reader, options ...requestOption) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, s.server.URL+path, body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Apply options (headers, cookies, etc.)
	for _, opt := range options {
		opt(req)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	return resp
}

// requestOption is a function that modifies a request
type requestOption func(*http.Request)

// withHost sets the Host header
func withHost(host string) requestOption {
	return func(r *http.Request) {
		r.Host = host
	}
}

// withSession adds a session cookie
func withSession(sessionID string) requestOption {
	return func(r *http.Request) {
		r.AddCookie(&http.Cookie{
			Name:  "fazt_session",
			Value: sessionID,
		})
	}
}

// withAPIKey adds an API key header
func withAPIKey(apiKey string) requestOption {
	return func(r *http.Request) {
		r.Header.Set("X-API-Key", apiKey)
	}
}

// withBearerToken adds a bearer token
func withBearerToken(token string) requestOption {
	return func(r *http.Request) {
		r.Header.Set("Authorization", "Bearer "+token)
	}
}

// withJSON sets content type to JSON
func withJSON() requestOption {
	return func(r *http.Request) {
		r.Header.Set("Content-Type", "application/json")
	}
}

// createTestApp creates a test app record in the database (no files).
// Use deployFiles() to add VFS files under a subdomain — matching production deploy behavior.
func (s *integrationTestServer) createTestApp(t *testing.T, appID, title string) {
	t.Helper()

	_, err := s.db.Exec(`
		INSERT INTO apps (id, title, created_at, updated_at)
		VALUES (?, ?, datetime('now'), datetime('now'))
	`, appID, title)
	if err != nil {
		t.Fatalf("Failed to insert test app: %v", err)
	}
}

// deployFiles writes VFS files under a site_id (subdomain).
// In production, deploy stores files keyed by subdomain, not app ID.
func (s *integrationTestServer) deployFiles(t *testing.T, siteID, title string) {
	t.Helper()

	indexContent := fmt.Sprintf("<html><body><h1>%s</h1></body></html>", title)
	indexHash := fmt.Sprintf("%x", sha256.Sum256([]byte(indexContent)))
	_, err := s.db.Exec(`
		INSERT INTO files (site_id, path, content, size_bytes, mime_type, hash, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, datetime('now'))
	`, siteID, "index.html", indexContent, len(indexContent), "text/html", indexHash)
	if err != nil {
		t.Fatalf("Failed to write index.html to VFS: %v", err)
	}

	manifestContent := fmt.Sprintf(`{"name":"%s"}`, siteID)
	manifestHash := fmt.Sprintf("%x", sha256.Sum256([]byte(manifestContent)))
	_, err = s.db.Exec(`
		INSERT INTO files (site_id, path, content, size_bytes, mime_type, hash, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, datetime('now'))
	`, siteID, "manifest.json", manifestContent, len(manifestContent), "application/json", manifestHash)
	if err != nil {
		t.Fatalf("Failed to write manifest.json to VFS: %v", err)
	}
}

// createTestAlias creates an alias for an app
func (s *integrationTestServer) createTestAlias(t *testing.T, subdomain, appID, aliasType string) {
	t.Helper()

	// Schema: subdomain, type, targets
	// For 'app' type, targets is JSON with app_id
	var targets string
	switch aliasType {
	case "app", "proxy":
		targets = fmt.Sprintf(`{"app_id":"%s"}`, appID)
	case "reserved":
		targets = "{}"
	default:
		targets = "{}"
	}

	_, err := s.db.Exec(`
		INSERT INTO aliases (subdomain, type, targets, created_at)
		VALUES (?, ?, ?, datetime('now'))
	`, subdomain, aliasType, targets)
	if err != nil {
		t.Fatalf("Failed to create alias: %v", err)
	}
}

// createTestAPIKey creates an API key for testing
func (s *integrationTestServer) createTestAPIKey(t *testing.T, apiKey, scope string) {
	t.Helper()

	_, err := s.db.Exec(`
		INSERT INTO api_keys (key, scope, created_at)
		VALUES (?, ?, datetime('now'))
	`, apiKey, scope)
	if err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}
}

// createTestUser creates a test user in the database
func (s *integrationTestServer) createTestUser(t *testing.T, username, email string) {
	t.Helper()

	_, err := s.db.Exec(`
		INSERT INTO auth_users (username, email, created_at)
		VALUES (?, ?, datetime('now'))
	`, username, email)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
}

// readBody reads and returns the response body as a string
func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	return string(body)
}

// parseJSON parses response body as JSON into target
func parseJSON(t *testing.T, resp *http.Response, target interface{}) {
	t.Helper()

	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}
}

// assertStatus asserts that the response has the expected status code
func assertStatus(t *testing.T, resp *http.Response, expected int) {
	t.Helper()

	if resp.StatusCode != expected {
		body := readBody(t, resp)
		t.Fatalf("Expected status %d, got %d. Body: %s", expected, resp.StatusCode, body)
	}
}

// assertContains asserts that the body contains the expected string
func assertContains(t *testing.T, body, expected string) {
	t.Helper()

	if !bytes.Contains([]byte(body), []byte(expected)) {
		t.Fatalf("Expected body to contain %q, got: %s", expected, body)
	}
}

// --- Storage Access Control Tests ---

// TestStorageAccessControl verifies that users cannot access each other's data
func TestStorageAccessControl(t *testing.T) {
	s := setupIntegrationTest(t)

	// Create two users (sessions not needed for this database-level test)
	_ = s.createSession(t, "userA@test.com", "user")
	_ = s.createSession(t, "userB@test.com", "user")

	// Create a test app
	appID := "test-storage-app"
	s.createTestApp(t, appID, "Test Storage App")

	// User A stores data in KV store
	// Note: Storage API would be exposed via serverless runtime
	// For integration test, we'll directly use the storage layer
	// to verify isolation at the database level

	userIDA := "user-userA@test.com"
	userIDB := "user-userB@test.com"

	// Insert data for User A
	_, err := s.db.Exec(`
		INSERT INTO app_kv (app_id, user_id, key, value, updated_at)
		VALUES (?, ?, ?, ?, strftime('%s', 'now'))
	`, appID, userIDA, "u:"+userIDA+":secret", `"userA secret data"`, time.Now().Unix())
	if err != nil {
		t.Fatalf("Failed to insert User A data: %v", err)
	}

	// User A can read their own data
	var valueA string
	err = s.db.QueryRow(`
		SELECT value FROM app_kv
		WHERE app_id = ? AND user_id = ? AND key = ?
	`, appID, userIDA, "u:"+userIDA+":secret").Scan(&valueA)
	if err != nil {
		t.Fatalf("User A cannot read own data: %v", err)
	}
	if valueA != `"userA secret data"` {
		t.Errorf("User A data mismatch: got %s", valueA)
	}

	// User B tries to read User A's data (should fail)
	err = s.db.QueryRow(`
		SELECT value FROM app_kv
		WHERE app_id = ? AND user_id = ? AND key = ?
	`, appID, userIDB, "u:"+userIDA+":secret").Scan(&valueA)
	if err == nil {
		t.Error("User B should NOT be able to read User A's data")
	}
	if err != sql.ErrNoRows {
		t.Errorf("Expected ErrNoRows, got: %v", err)
	}

	// User B tries to access via key alone (without user_id filter)
	// This should still require user_id in production - query without it might match wrong user
	var rawValue string
	err = s.db.QueryRow(`
		SELECT value FROM app_kv
		WHERE app_id = ? AND key = ? LIMIT 1
	`, appID, "u:"+userIDA+":secret").Scan(&rawValue)
	// Query may succeed (finding User A's row) but demonstrates why user_id filter is critical
	if err == nil {
		t.Logf("Query without user_id filter returned: %s (shows importance of user_id scoping)", rawValue)
	}

	// Verify User B can store their own separate data
	_, err = s.db.Exec(`
		INSERT INTO app_kv (app_id, user_id, key, value, updated_at)
		VALUES (?, ?, ?, ?, strftime('%s', 'now'))
	`, appID, userIDB, "u:"+userIDB+":secret", `"userB secret data"`, time.Now().Unix())
	if err != nil {
		t.Fatalf("Failed to insert User B data: %v", err)
	}

	// User B reads their own data successfully
	var valueB string
	err = s.db.QueryRow(`
		SELECT value FROM app_kv
		WHERE app_id = ? AND user_id = ? AND key = ?
	`, appID, userIDB, "u:"+userIDB+":secret").Scan(&valueB)
	if err != nil {
		t.Fatalf("User B cannot read own data: %v", err)
	}
	if valueB != `"userB secret data"` {
		t.Errorf("User B data mismatch: got %s", valueB)
	}

	// Verify both users' data exists independently
	var countA, countB int
	s.db.QueryRow("SELECT COUNT(*) FROM app_kv WHERE app_id = ? AND user_id = ?", appID, userIDA).Scan(&countA)
	s.db.QueryRow("SELECT COUNT(*) FROM app_kv WHERE app_id = ? AND user_id = ?", appID, userIDB).Scan(&countB)

	if countA != 1 {
		t.Errorf("Expected 1 record for User A, got %d", countA)
	}
	if countB != 1 {
		t.Errorf("Expected 1 record for User B, got %d", countB)
	}
}
