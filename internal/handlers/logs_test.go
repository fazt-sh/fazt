package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/handlers/testutil"
)

func setupLogsTest(t *testing.T) {
	t.Helper()
	silenceTestLogs(t)

	db := setupTestDB(t)
	database.SetDB(db)
	t.Cleanup(func() {
		db.Close()
		database.SetDB(nil)
	})
}

func insertTestLog(t *testing.T, siteID, level, message string) {
	t.Helper()
	db := database.GetDB()
	_, err := db.Exec(`
		INSERT INTO site_logs (site_id, level, message)
		VALUES (?, ?, ?)
	`, siteID, level, message)
	if err != nil {
		t.Fatalf("Failed to insert test log: %v", err)
	}
}

// --- LogsHandler ---

func TestLogsHandler_Success(t *testing.T) {
	setupLogsTest(t)
	insertTestLog(t, "my-app", "info", "Server started")
	insertTestLog(t, "my-app", "error", "Something broke")

	req := httptest.NewRequest("GET", "/api/logs?site_id=my-app", nil)
	resp := httptest.NewRecorder()
	LogsHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusOK)
	logs, ok := data["logs"]
	if !ok {
		t.Fatal("Expected 'logs' in response")
	}
	logsList, ok := logs.([]interface{})
	if !ok {
		t.Fatal("Expected 'logs' to be an array")
	}
	if len(logsList) != 2 {
		t.Errorf("Expected 2 logs, got %d", len(logsList))
	}
}

func TestLogsHandler_Empty(t *testing.T) {
	setupLogsTest(t)

	req := httptest.NewRequest("GET", "/api/logs?site_id=empty-site", nil)
	resp := httptest.NewRecorder()
	LogsHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusOK)
	logs := data["logs"]
	if logs != nil {
		logsList, ok := logs.([]interface{})
		if ok && len(logsList) != 0 {
			t.Errorf("Expected empty logs, got %d", len(logsList))
		}
	}
}

func TestLogsHandler_MissingSiteID(t *testing.T) {
	setupLogsTest(t)

	req := httptest.NewRequest("GET", "/api/logs", nil)
	resp := httptest.NewRecorder()
	LogsHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

func TestLogsHandler_MethodNotAllowed(t *testing.T) {
	setupLogsTest(t)

	req := httptest.NewRequest("POST", "/api/logs?site_id=test", nil)
	resp := httptest.NewRecorder()
	LogsHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

func TestLogsHandler_WithLimit(t *testing.T) {
	setupLogsTest(t)
	for i := 0; i < 5; i++ {
		insertTestLog(t, "limit-app", "info", "Log entry")
	}

	req := httptest.NewRequest("GET", "/api/logs?site_id=limit-app&limit=2", nil)
	resp := httptest.NewRecorder()
	LogsHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusOK)
	logsList := data["logs"].([]interface{})
	if len(logsList) != 2 {
		t.Errorf("Expected 2 logs with limit, got %d", len(logsList))
	}
}

func TestLogsHandler_InvalidLimit(t *testing.T) {
	setupLogsTest(t)
	insertTestLog(t, "inv-app", "info", "test")

	// Invalid limit should fall back to default (50)
	req := httptest.NewRequest("GET", "/api/logs?site_id=inv-app&limit=abc", nil)
	resp := httptest.NewRecorder()
	LogsHandler(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.Code)
	}
}

// --- LogStreamManager ---

func TestLogStreamManager_SubscribeUnsubscribe(t *testing.T) {
	mgr := &LogStreamManager{
		channels: make(map[string][]chan LogEvent),
	}

	ch := make(chan LogEvent, 10)
	mgr.Subscribe("test-site", ch)

	// Verify subscribed
	mgr.mu.RLock()
	if len(mgr.channels["test-site"]) != 1 {
		t.Errorf("Expected 1 subscriber, got %d", len(mgr.channels["test-site"]))
	}
	mgr.mu.RUnlock()

	// Unsubscribe
	mgr.Unsubscribe("test-site", ch)

	mgr.mu.RLock()
	if len(mgr.channels["test-site"]) != 0 {
		t.Errorf("Expected 0 subscribers after unsubscribe, got %d", len(mgr.channels["test-site"]))
	}
	mgr.mu.RUnlock()
}

func TestLogStreamManager_Broadcast(t *testing.T) {
	mgr := &LogStreamManager{
		channels: make(map[string][]chan LogEvent),
	}

	ch1 := make(chan LogEvent, 10)
	ch2 := make(chan LogEvent, 10)
	mgr.Subscribe("site-a", ch1)
	mgr.Subscribe("site-a", ch2)

	event := LogEvent{
		ID:      1,
		SiteID:  "site-a",
		Level:   "info",
		Message: "test broadcast",
	}
	mgr.Broadcast(event)

	// Both channels should receive the event
	select {
	case e := <-ch1:
		if e.Message != "test broadcast" {
			t.Errorf("ch1: expected 'test broadcast', got %s", e.Message)
		}
	default:
		t.Error("ch1: expected event but none received")
	}

	select {
	case e := <-ch2:
		if e.Message != "test broadcast" {
			t.Errorf("ch2: expected 'test broadcast', got %s", e.Message)
		}
	default:
		t.Error("ch2: expected event but none received")
	}
}

func TestLogStreamManager_BroadcastDifferentSites(t *testing.T) {
	mgr := &LogStreamManager{
		channels: make(map[string][]chan LogEvent),
	}

	chA := make(chan LogEvent, 10)
	chB := make(chan LogEvent, 10)
	mgr.Subscribe("site-a", chA)
	mgr.Subscribe("site-b", chB)

	event := LogEvent{SiteID: "site-a", Message: "only for A"}
	mgr.Broadcast(event)

	// chA should receive
	select {
	case <-chA:
		// OK
	default:
		t.Error("chA: expected event")
	}

	// chB should NOT receive
	select {
	case <-chB:
		t.Error("chB: should not receive event for different site")
	default:
		// OK
	}
}

func TestLogStreamManager_BroadcastFullChannel(t *testing.T) {
	mgr := &LogStreamManager{
		channels: make(map[string][]chan LogEvent),
	}

	// Channel with 0 buffer — will be full immediately
	ch := make(chan LogEvent)
	mgr.Subscribe("full-site", ch)

	// Should not block
	event := LogEvent{SiteID: "full-site", Message: "dropped"}
	mgr.Broadcast(event)
	// If we get here without hanging, the test passes
}

// --- PersistLog ---

func TestPersistLog_Success(t *testing.T) {
	setupLogsTest(t)

	err := PersistLog("test-site", "info", "Hello from test")
	if err != nil {
		t.Fatalf("PersistLog failed: %v", err)
	}

	// Verify persisted
	db := database.GetDB()
	var msg string
	err = db.QueryRow("SELECT message FROM site_logs WHERE site_id = ?", "test-site").Scan(&msg)
	if err != nil {
		t.Fatalf("Failed to query log: %v", err)
	}
	if msg != "Hello from test" {
		t.Errorf("Expected 'Hello from test', got %q", msg)
	}
}

func TestPersistLog_NilDB(t *testing.T) {
	silenceTestLogs(t)
	database.SetDB(nil)

	// Should not error with nil DB
	err := PersistLog("test", "info", "should not crash")
	if err != nil {
		t.Errorf("Expected nil error with nil DB, got %v", err)
	}
}

// --- LogStreamHandler ---

func TestLogStreamHandler_MissingSiteID(t *testing.T) {
	silenceTestLogs(t)

	req := httptest.NewRequest("GET", "/api/logs/stream", nil)
	resp := httptest.NewRecorder()
	LogStreamHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}
