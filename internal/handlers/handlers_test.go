package handlers

import (
	"database/sql"
	"io"
	"log"
	"testing"
	"time"

	"github.com/fazt-sh/fazt/internal/auth"
	"github.com/fazt-sh/fazt/internal/config"
	"github.com/fazt-sh/fazt/internal/database"

	_ "modernc.org/sqlite"
)

var (
	testDB           *sql.DB
	testSessionStore *auth.SessionStore
	testSessionID    string
	testUsername     = "admin"
	testAPIKey       = "test_api_key_12345"
)

// setupTestDB creates an in-memory database with full schema for testing
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Allow multiple connections for in-memory DBs in tests.
	// Note: In-memory DBs are isolated per connection normally, but with a shared
	// cache mode, multiple connections can work. We use 1 for migrations to avoid
	// race conditions, but tests can use more.
	db.SetMaxOpenConns(1)

	// Enable WAL and foreign keys
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA foreign_keys=ON")

	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}

// cleanupTestDB closes and cleans up the test database
func cleanupTestDB(t *testing.T, db *sql.DB) {
	t.Helper()
	if db != nil {
		db.Close()
	}
}

// setupTestAuth creates a test session store and returns a valid session ID
func setupTestAuth(t *testing.T) (store *auth.SessionStore, sessionID string) {
	t.Helper()

	// Create session store with 24h TTL
	store = auth.NewSessionStore(24 * time.Hour)

	// Create a test session
	sessionID, err := store.CreateSession(testUsername)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	return store, sessionID
}

// setupTestConfig initializes test configuration
func setupTestConfig(t *testing.T) {
	t.Helper()

	// Initialize with test config
	testCfg := &config.Config{
		Server: config.ServerConfig{
			Domain: "localhost",
			Port:   "8080",
			Env:    "test",
		},
		Auth: config.AuthConfig{
			Username:     testUsername,
			PasswordHash: "$2a$10$...", // Placeholder hash
		},
	}

	config.SetConfig(testCfg)
}

// createTestSite creates a test site in the VFS
func createTestSite(t *testing.T, db *sql.DB, siteName string) {
	t.Helper()

	// Insert a simple test file
	_, err := db.Exec(`
		INSERT INTO files (site_id, path, content, size_bytes, mime_type, hash)
		VALUES (?, ?, ?, ?, ?, ?)
	`, siteName, "index.html", []byte("<html>Test</html>"), 18, "text/html", "test-hash-123")

	if err != nil {
		t.Fatalf("Failed to create test site: %v", err)
	}
}

// createTestRedirect creates a test redirect
func createTestRedirect(t *testing.T, db *sql.DB, slug, destination string) int64 {
	t.Helper()

	result, err := db.Exec(`
		INSERT INTO redirects (slug, destination, tags)
		VALUES (?, ?, ?)
	`, slug, destination, "")

	if err != nil {
		t.Fatalf("Failed to create test redirect: %v", err)
	}

	id, _ := result.LastInsertId()
	return id
}

// createTestWebhook creates a test webhook
func createTestWebhook(t *testing.T, db *sql.DB, name, endpoint string) int64 {
	t.Helper()

	result, err := db.Exec(`
		INSERT INTO webhooks (name, endpoint, secret, is_active)
		VALUES (?, ?, '', 1)
	`, name, endpoint)

	if err != nil {
		t.Fatalf("Failed to create test webhook: %v", err)
	}

	id, _ := result.LastInsertId()
	return id
}

// createTestEvent creates a test analytics event
func createTestEvent(t *testing.T, db *sql.DB, domain, eventType string) {
	t.Helper()

	_, err := db.Exec(`
		INSERT INTO events (domain, event_type, source_type, path)
		VALUES (?, ?, ?, ?)
	`, domain, eventType, "web", "/test")

	if err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}
}

// silenceTestLogs redirects log output during tests
func silenceTestLogs(t *testing.T) {
	t.Helper()
	log.SetOutput(io.Discard)
	t.Cleanup(func() {
		log.SetOutput(io.Discard) // Keep it silenced for tests
	})
}
