package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/handlers/testutil"
)

func setupAgentHandlerTest(t *testing.T) *sql.DB {
	t.Helper()

	silenceTestLogs(t)

	db := setupTestDB(t)
	database.SetDB(db)
	t.Cleanup(func() {
		db.Close()
		database.SetDB(nil)
	})

	return db
}

func insertTestApp(t *testing.T, appID string) {
	t.Helper()

	db := database.GetDB()
	_, err := db.Exec(`INSERT INTO apps (id, title) VALUES (?, ?)`, appID, "Test App")
	if err != nil {
		t.Fatalf("Failed to insert app: %v", err)
	}

	_, err = db.Exec(`INSERT INTO files (site_id, path, content, size_bytes, mime_type, hash, app_id) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		appID, "index.html", []byte("<html>ok</html>"), 15, "text/html", "hash-1", appID)
	if err != nil {
		t.Fatalf("Failed to insert file: %v", err)
	}
}

func TestAgentInfoHandler_MissingHeader(t *testing.T) {
	setupAgentHandlerTest(t)

	req := httptest.NewRequest(http.MethodGet, "/_fazt/info", nil)
	rr := httptest.NewRecorder()

	AgentInfoHandler(rr, req)

	testutil.CheckError(t, rr, http.StatusBadRequest, "BAD_REQUEST")
}

func TestAgentInfoHandler_Success(t *testing.T) {
	setupAgentHandlerTest(t)

	appID := "app_test123"
	insertTestApp(t, appID)

	_, err := database.GetDB().Exec(`INSERT INTO kv_store (site_id, key, value) VALUES (?, ?, ?)`, appID, "k1", `{"foo":"bar"}`)
	if err != nil {
		t.Fatalf("Failed to insert kv data: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/_fazt/info", nil)
	req.Header.Set("X-Fazt-App-ID", appID)
	rr := httptest.NewRecorder()

	AgentInfoHandler(rr, req)

	data := testutil.CheckSuccess(t, rr, http.StatusOK)
	testutil.AssertFieldEquals(t, data, "id", appID)

	if got, ok := data["storage_keys"].(float64); !ok || got != 1 {
		t.Fatalf("storage_keys = %v, want 1", data["storage_keys"])
	}
}

func TestAgentStorageGetHandler_NotFound(t *testing.T) {
	setupAgentHandlerTest(t)

	req := httptest.NewRequest(http.MethodGet, "/_fazt/storage/missing", nil)
	req.Header.Set("X-Fazt-App-ID", "app_test123")
	req.SetPathValue("key", "missing")
	rr := httptest.NewRecorder()

	AgentStorageGetHandler(rr, req)

	testutil.CheckError(t, rr, http.StatusNotFound, "KEY_NOT_FOUND")
}

func TestAgentStorageGetHandler_JSONValue(t *testing.T) {
	setupAgentHandlerTest(t)

	appID := "app_test123"
	_, err := database.GetDB().Exec(`INSERT INTO kv_store (site_id, key, value) VALUES (?, ?, ?)`, appID, "config", `{"foo":"bar"}`)
	if err != nil {
		t.Fatalf("Failed to insert kv data: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/_fazt/storage/config", nil)
	req.Header.Set("X-Fazt-App-ID", appID)
	req.SetPathValue("key", "config")
	rr := httptest.NewRecorder()

	AgentStorageGetHandler(rr, req)

	data := testutil.CheckSuccess(t, rr, http.StatusOK)
	testutil.AssertFieldEquals(t, data, "foo", "bar")
}

func TestAgentSnapshotAndRestore(t *testing.T) {
	setupAgentHandlerTest(t)

	appID := "app_test123"
	db := database.GetDB()
	_, err := db.Exec(`INSERT INTO kv_store (site_id, key, value) VALUES (?, ?, ?)`, appID, "k1", `"v1"`)
	if err != nil {
		t.Fatalf("Failed to insert kv data: %v", err)
	}

	body, _ := json.Marshal(SnapshotRequest{Name: "snap1"})
	req := httptest.NewRequest(http.MethodPost, "/_fazt/snapshot", bytes.NewReader(body))
	req.Header.Set("X-Fazt-App-ID", appID)
	rr := httptest.NewRecorder()

	AgentSnapshotHandler(rr, req)

	testutil.CheckSuccess(t, rr, http.StatusCreated)

	// Change KV data before restore
	_, err = db.Exec(`DELETE FROM kv_store WHERE site_id = ?`, appID)
	if err != nil {
		t.Fatalf("Failed to clear kv data: %v", err)
	}
	_, err = db.Exec(`INSERT INTO kv_store (site_id, key, value) VALUES (?, ?, ?)`, appID, "k2", `"v2"`)
	if err != nil {
		t.Fatalf("Failed to insert new kv data: %v", err)
	}

	restoreReq := httptest.NewRequest(http.MethodPost, "/_fazt/restore/snap1", nil)
	restoreReq.Header.Set("X-Fazt-App-ID", appID)
	restoreReq.SetPathValue("name", "snap1")
	restoreRR := httptest.NewRecorder()

	AgentRestoreHandler(restoreRR, restoreReq)

	testutil.CheckSuccess(t, restoreRR, http.StatusOK)

	// Verify original key restored
	var value string
	err = db.QueryRow(`SELECT value FROM kv_store WHERE site_id = ? AND key = ?`, appID, "k1").Scan(&value)
	if err != nil {
		t.Fatalf("Failed to read restored key: %v", err)
	}
	if value != `"v1"` {
		t.Fatalf("restored value = %q, want %q", value, `"v1"`)
	}
}
