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

// createTestApp creates a test app in the database and VFS
func (s *integrationTestServer) createTestApp(t *testing.T, appID, title string) {
	t.Helper()

	// Insert app into database
	_, err := s.db.Exec(`
		INSERT INTO apps (id, title, created_at, updated_at)
		VALUES (?, ?, datetime('now'), datetime('now'))
	`, appID, title)
	if err != nil {
		t.Fatalf("Failed to insert test app: %v", err)
	}

	// Write files to VFS (database-backed)
	// Schema: site_id, path, content, size_bytes, mime_type, hash
	indexContent := fmt.Sprintf("<html><body><h1>%s</h1></body></html>", title)
	indexHash := fmt.Sprintf("%x", sha256.Sum256([]byte(indexContent)))
	_, err = s.db.Exec(`
		INSERT INTO files (site_id, path, content, size_bytes, mime_type, hash, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, datetime('now'))
	`, appID, "index.html", indexContent, len(indexContent), "text/html", indexHash)
	if err != nil {
		t.Fatalf("Failed to write index.html to VFS: %v", err)
	}

	manifestContent := fmt.Sprintf(`{"name":"%s"}`, appID)
	manifestHash := fmt.Sprintf("%x", sha256.Sum256([]byte(manifestContent)))
	_, err = s.db.Exec(`
		INSERT INTO files (site_id, path, content, size_bytes, mime_type, hash, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, datetime('now'))
	`, appID, "manifest.json", manifestContent, len(manifestContent), "application/json", manifestHash)
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
