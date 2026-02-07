package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/handlers/testutil"

	"golang.org/x/crypto/bcrypt"
)

const testCmdAPIKey = "test_cmd_api_key_12345"

// setupCmdTestDB creates an in-memory DB with all tables needed by cmd gateway
func setupCmdTestDB(t *testing.T) {
	t.Helper()

	db := setupTestDB(t)

	// Add apps and aliases tables (not in base test schema)
	// IMPORTANT: Match production schema - type defaults to 'app' not 'proxy'
	// Note: files table is in base schema with app_id column
	schema := `
	CREATE TABLE IF NOT EXISTS apps (
		id TEXT PRIMARY KEY,
		original_id TEXT,
		forked_from_id TEXT,
		title TEXT,
		description TEXT,
		tags TEXT,
		visibility TEXT DEFAULT 'unlisted',
		source TEXT DEFAULT 'deploy',
		source_url TEXT,
		source_ref TEXT,
		source_commit TEXT,
		spa INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS aliases (
		subdomain TEXT PRIMARY KEY,
		type TEXT DEFAULT 'app',
		targets TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create cmd test schema: %v", err)
	}

	// Insert a test API key (bcrypt hash of testCmdAPIKey)
	hash, err := bcrypt.GenerateFromPassword([]byte(testCmdAPIKey), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("Failed to hash API key: %v", err)
	}
	_, err = db.Exec(`INSERT INTO api_keys (name, key_hash) VALUES (?, ?)`, "test-key", string(hash))
	if err != nil {
		t.Fatalf("Failed to insert test API key: %v", err)
	}

	// Insert test apps
	_, err = db.Exec(`INSERT INTO apps (id, title, visibility) VALUES (?, ?, ?)`,
		"app_test123", "test-app", "unlisted")
	if err != nil {
		t.Fatalf("Failed to insert test app: %v", err)
	}

	// Insert test alias with type='app' (production default)
	_, err = db.Exec(`INSERT INTO aliases (subdomain, type, targets) VALUES (?, ?, ?)`,
		"test-alias", "app", `{"app_id":"app_test123"}`)
	if err != nil {
		t.Fatalf("Failed to insert test alias: %v", err)
	}

	// Insert test files for app (with app_id for proper JOIN)
	_, err = db.Exec(`INSERT INTO files (site_id, path, content, size_bytes, mime_type, hash, app_id) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"app_test123", "index.html", []byte("<html>test</html>"), 18, "text/html", "testhash123", "app_test123")
	if err != nil {
		t.Fatalf("Failed to insert test file: %v", err)
	}

	// Set as global DB (cmd gateway uses database.GetDB())
	database.SetDB(db)
	t.Cleanup(func() {
		db.Close()
		database.SetDB(nil)
	})
}

func TestCmdGateway_RejectsUnauthenticated(t *testing.T) {
	silenceTestLogs(t)
	setupTestConfig(t)
	setupCmdTestDB(t)

	req := testutil.JSONRequest("POST", "/api/cmd", map[string]interface{}{
		"command": "app",
		"args":    []string{"list"},
	})
	// No auth header

	rr := httptest.NewRecorder()
	CmdGatewayHandler(rr, req)

	testutil.CheckError(t, rr, 401, "UNAUTHORIZED")
}

func TestCmdGateway_RejectsInvalidToken(t *testing.T) {
	silenceTestLogs(t)
	setupTestConfig(t)
	setupCmdTestDB(t)

	req := testutil.JSONRequest("POST", "/api/cmd", map[string]interface{}{
		"command": "app",
		"args":    []string{"list"},
	})
	testutil.WithAuth(req, "invalid_token_xyz")

	rr := httptest.NewRecorder()
	CmdGatewayHandler(rr, req)

	testutil.CheckError(t, rr, 401, "INVALID_API_KEY")
}

func TestCmdGateway_AcceptsValidAPIKey(t *testing.T) {
	silenceTestLogs(t)
	setupTestConfig(t)
	setupCmdTestDB(t)

	req := testutil.JSONRequest("POST", "/api/cmd", map[string]interface{}{
		"command": "app",
		"args":    []string{"list"},
	})
	testutil.WithAuth(req, testCmdAPIKey)

	rr := httptest.NewRecorder()
	CmdGatewayHandler(rr, req)

	data := testutil.CheckSuccess(t, rr, 200)
	testutil.AssertFieldEquals(t, data, "success", true)
}

func TestCmdGateway_RejectsInvalidMethod(t *testing.T) {
	silenceTestLogs(t)
	setupTestConfig(t)
	setupCmdTestDB(t)

	req := testutil.JSONRequest("GET", "/api/cmd", nil)
	testutil.WithAuth(req, testCmdAPIKey)

	rr := httptest.NewRecorder()
	CmdGatewayHandler(rr, req)

	testutil.CheckError(t, rr, 405, "METHOD_NOT_ALLOWED")
}

func TestCmdGateway_AppListReturnsApps(t *testing.T) {
	silenceTestLogs(t)
	setupTestConfig(t)
	setupCmdTestDB(t)

	req := testutil.JSONRequest("POST", "/api/cmd", map[string]interface{}{
		"command": "app",
		"args":    []string{"list"},
	})
	testutil.WithAuth(req, testCmdAPIKey)

	rr := httptest.NewRecorder()
	CmdGatewayHandler(rr, req)

	data := testutil.CheckSuccess(t, rr, 200)
	testutil.AssertFieldEquals(t, data, "success", true)
	testutil.AssertFieldExists(t, data, "data")
}

func TestCmdGateway_AppInfoByAlias(t *testing.T) {
	silenceTestLogs(t)
	setupTestConfig(t)
	setupCmdTestDB(t)

	req := testutil.JSONRequest("POST", "/api/cmd", map[string]interface{}{
		"command": "app",
		"args":    []string{"info", "test-alias"},
	})
	testutil.WithAuth(req, testCmdAPIKey)

	rr := httptest.NewRecorder()
	CmdGatewayHandler(rr, req)

	data := testutil.CheckSuccess(t, rr, 200)

	// Debug: print the actual response
	if success, ok := data["success"].(bool); !ok || !success {
		t.Logf("App info failed. Full response: %+v", data)
		if errMsg, ok := data["error"].(string); ok {
			t.Logf("Error message: %s", errMsg)
		}
	}

	testutil.AssertFieldEquals(t, data, "success", true)
	testutil.AssertFieldExists(t, data, "data")

	// Verify the returned data contains app info
	dataMap, ok := data["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}
	if id, ok := dataMap["id"].(string); !ok || id != "app_test123" {
		t.Errorf("Expected id 'app_test123', got '%v'", dataMap["id"])
	}
}

func TestCmdGateway_AppInfoByID(t *testing.T) {
	silenceTestLogs(t)
	setupTestConfig(t)
	setupCmdTestDB(t)

	req := testutil.JSONRequest("POST", "/api/cmd", map[string]interface{}{
		"command": "app",
		"args":    []string{"info", "--id", "app_test123"},
	})
	testutil.WithAuth(req, testCmdAPIKey)

	rr := httptest.NewRecorder()
	CmdGatewayHandler(rr, req)

	data := testutil.CheckSuccess(t, rr, 200)
	testutil.AssertFieldEquals(t, data, "success", true)
	testutil.AssertFieldExists(t, data, "data")
}

func TestCmdGateway_AppInfoNotFound(t *testing.T) {
	silenceTestLogs(t)
	setupTestConfig(t)
	setupCmdTestDB(t)

	req := testutil.JSONRequest("POST", "/api/cmd", map[string]interface{}{
		"command": "app",
		"args":    []string{"info", "nonexistent"},
	})
	testutil.WithAuth(req, testCmdAPIKey)

	rr := httptest.NewRecorder()
	CmdGatewayHandler(rr, req)

	data := testutil.CheckSuccess(t, rr, 200)
	testutil.AssertFieldEquals(t, data, "success", false)
	testutil.AssertFieldExists(t, data, "error")
}

func TestCmdGateway_UnknownCommand(t *testing.T) {
	silenceTestLogs(t)
	setupTestConfig(t)
	setupCmdTestDB(t)

	req := testutil.JSONRequest("POST", "/api/cmd", map[string]interface{}{
		"command": "nonexistent",
		"args":    []string{},
	})
	testutil.WithAuth(req, testCmdAPIKey)

	rr := httptest.NewRecorder()
	CmdGatewayHandler(rr, req)

	// Gateway returns 200 with success: false for command errors
	data := testutil.CheckSuccess(t, rr, 200)
	testutil.AssertFieldEquals(t, data, "success", false)
	testutil.AssertFieldExists(t, data, "error")
}
