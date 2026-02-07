package handlers

import (
	"testing"

	"github.com/fazt-sh/fazt/internal/database"
)

func setupAliasTestDB(t *testing.T) {
	t.Helper()

	db := setupTestDB(t)

	// Create apps and aliases tables matching production schema
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
	CREATE TABLE IF NOT EXISTS files (
		site_id TEXT NOT NULL,
		path TEXT NOT NULL,
		content BLOB,
		size_bytes INTEGER NOT NULL,
		mime_type TEXT,
		hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (site_id, path)
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create alias test schema: %v", err)
	}

	database.SetDB(db)
	t.Cleanup(func() {
		db.Close()
		database.SetDB(nil)
	})
}

func TestResolveAlias_AppType(t *testing.T) {
	silenceTestLogs(t)
	setupAliasTestDB(t)

	db := database.GetDB()

	// Insert app
	_, err := db.Exec(`INSERT INTO apps (id, title) VALUES (?, ?)`, "app_test123", "test-app")
	if err != nil {
		t.Fatalf("Failed to insert app: %v", err)
	}

	// Insert alias with type='app' (production default)
	_, err = db.Exec(`INSERT INTO aliases (subdomain, type, targets) VALUES (?, ?, ?)`,
		"test", "app", `{"app_id":"app_test123"}`)
	if err != nil {
		t.Fatalf("Failed to insert alias: %v", err)
	}

	appID, aliasType, err := ResolveAlias("test")
	if err != nil {
		t.Fatalf("ResolveAlias failed: %v", err)
	}

	if appID != "app_test123" {
		t.Errorf("Expected app_id 'app_test123', got '%s'", appID)
	}

	if aliasType != "app" {
		t.Errorf("Expected type 'app', got '%s'", aliasType)
	}
}

func TestResolveAlias_ProxyType(t *testing.T) {
	silenceTestLogs(t)
	setupAliasTestDB(t)

	db := database.GetDB()

	// Insert app
	_, err := db.Exec(`INSERT INTO apps (id, title) VALUES (?, ?)`, "app_test456", "test-app-2")
	if err != nil {
		t.Fatalf("Failed to insert app: %v", err)
	}

	// Insert alias with type='proxy'
	_, err = db.Exec(`INSERT INTO aliases (subdomain, type, targets) VALUES (?, ?, ?)`,
		"proxy-test", "proxy", `{"app_id":"app_test456"}`)
	if err != nil {
		t.Fatalf("Failed to insert alias: %v", err)
	}

	appID, aliasType, err := ResolveAlias("proxy-test")
	if err != nil {
		t.Fatalf("ResolveAlias failed: %v", err)
	}

	if appID != "app_test456" {
		t.Errorf("Expected app_id 'app_test456', got '%s'", appID)
	}

	if aliasType != "proxy" {
		t.Errorf("Expected type 'proxy', got '%s'", aliasType)
	}
}

func TestResolveAlias_SplitType(t *testing.T) {
	silenceTestLogs(t)
	setupAliasTestDB(t)

	db := database.GetDB()

	// Insert apps
	_, err := db.Exec(`INSERT INTO apps (id, title) VALUES (?, ?), (?, ?)`,
		"app_split1", "split-app-1",
		"app_split2", "split-app-2")
	if err != nil {
		t.Fatalf("Failed to insert apps: %v", err)
	}

	// Insert alias with type='split'
	_, err = db.Exec(`INSERT INTO aliases (subdomain, type, targets) VALUES (?, ?, ?)`,
		"split-test", "split", `[{"app_id":"app_split1","weight":50},{"app_id":"app_split2","weight":50}]`)
	if err != nil {
		t.Fatalf("Failed to insert alias: %v", err)
	}

	appID, aliasType, err := ResolveAlias("split-test")
	if err != nil {
		t.Fatalf("ResolveAlias failed: %v", err)
	}

	// Should return first target
	if appID != "app_split1" {
		t.Errorf("Expected app_id 'app_split1', got '%s'", appID)
	}

	if aliasType != "split" {
		t.Errorf("Expected type 'split', got '%s'", aliasType)
	}
}

func TestResolveAlias_ReservedType(t *testing.T) {
	silenceTestLogs(t)
	setupAliasTestDB(t)

	db := database.GetDB()

	// Insert reserved alias
	_, err := db.Exec(`INSERT INTO aliases (subdomain, type, targets) VALUES (?, ?, NULL)`,
		"reserved-test", "reserved")
	if err != nil {
		t.Fatalf("Failed to insert alias: %v", err)
	}

	appID, aliasType, err := ResolveAlias("reserved-test")
	if err != nil {
		t.Fatalf("ResolveAlias failed: %v", err)
	}

	if appID != "" {
		t.Errorf("Expected empty app_id for reserved, got '%s'", appID)
	}

	if aliasType != "reserved" {
		t.Errorf("Expected type 'reserved', got '%s'", aliasType)
	}
}

func TestResolveAlias_NotFound(t *testing.T) {
	silenceTestLogs(t)
	setupAliasTestDB(t)

	appID, aliasType, err := ResolveAlias("nonexistent")
	if err != nil {
		t.Fatalf("ResolveAlias should not error on not found: %v", err)
	}

	if appID != "" {
		t.Errorf("Expected empty app_id for not found, got '%s'", appID)
	}

	if aliasType != "" {
		t.Errorf("Expected empty type for not found, got '%s'", aliasType)
	}
}

func TestResolveAlias_MalformedJSON(t *testing.T) {
	silenceTestLogs(t)
	setupAliasTestDB(t)

	db := database.GetDB()

	// Insert alias with malformed JSON
	_, err := db.Exec(`INSERT INTO aliases (subdomain, type, targets) VALUES (?, ?, ?)`,
		"malformed", "app", `{bad json}`)
	if err != nil {
		t.Fatalf("Failed to insert alias: %v", err)
	}

	appID, _, err := ResolveAlias("malformed")
	if err != nil {
		t.Fatalf("ResolveAlias should not error on malformed JSON: %v", err)
	}

	// Should return empty since JSON unmarshal fails
	if appID != "" {
		t.Errorf("Expected empty app_id for malformed JSON, got '%s'", appID)
	}
}
