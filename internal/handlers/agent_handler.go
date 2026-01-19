package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/fazt-sh/fazt/internal/api"
	"github.com/fazt-sh/fazt/internal/database"
)

// AgentInfoHandler returns app metadata for agent workflows
// GET /_fazt/info
func AgentInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	// Extract app_id from context or Host header
	appID := r.Header.Get("X-Fazt-App-ID")
	if appID == "" {
		api.BadRequest(w, "X-Fazt-App-ID header required")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	app, err := getAppByID(db, appID)
	if err != nil {
		api.NotFound(w, "APP_NOT_FOUND", "App not found")
		return
	}

	// Get storage stats
	var kvCount int
	db.QueryRow("SELECT COUNT(*) FROM kv_store WHERE site_id = ?", appID).Scan(&kvCount)

	result := map[string]interface{}{
		"id":           app.ID,
		"title":        app.Title,
		"description":  app.Description,
		"visibility":   app.Visibility,
		"source":       app.Source,
		"original_id":  app.OriginalID,
		"forked_from":  app.ForkedFromID,
		"file_count":   app.FileCount,
		"size_bytes":   app.SizeBytes,
		"storage_keys": kvCount,
	}

	api.Success(w, http.StatusOK, result)
}

// AgentStorageListHandler lists all storage keys
// GET /_fazt/storage
func AgentStorageListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	appID := r.Header.Get("X-Fazt-App-ID")
	if appID == "" {
		api.BadRequest(w, "X-Fazt-App-ID header required")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	rows, err := db.Query("SELECT key, LENGTH(value) as size FROM kv_store WHERE site_id = ? ORDER BY key", appID)
	if err != nil {
		api.InternalError(w, err)
		return
	}
	defer rows.Close()

	var keys []map[string]interface{}
	for rows.Next() {
		var key string
		var size int
		if rows.Scan(&key, &size) == nil {
			keys = append(keys, map[string]interface{}{
				"key":  key,
				"size": size,
			})
		}
	}

	api.Success(w, http.StatusOK, keys)
}

// AgentStorageGetHandler gets a specific storage key
// GET /_fazt/storage/:key
func AgentStorageGetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	appID := r.Header.Get("X-Fazt-App-ID")
	key := r.PathValue("key")
	if appID == "" || key == "" {
		api.BadRequest(w, "X-Fazt-App-ID header and key required")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	var value string
	err := db.QueryRow("SELECT value FROM kv_store WHERE site_id = ? AND key = ?", appID, key).Scan(&value)
	if err == sql.ErrNoRows {
		api.NotFound(w, "KEY_NOT_FOUND", "Storage key not found")
		return
	}
	if err != nil {
		api.InternalError(w, err)
		return
	}

	// Try to parse as JSON, return raw if not
	var jsonValue interface{}
	if err := json.Unmarshal([]byte(value), &jsonValue); err == nil {
		api.Success(w, http.StatusOK, jsonValue)
	} else {
		api.Success(w, http.StatusOK, value)
	}
}

// Snapshot represents a named snapshot
type Snapshot struct {
	Name      string `json:"name"`
	AppID     string `json:"app_id"`
	CreatedAt string `json:"created_at"`
	KeyCount  int    `json:"key_count"`
}

// SnapshotRequest is the request body for creating a snapshot
type SnapshotRequest struct {
	Name string `json:"name"`
}

// AgentSnapshotHandler creates a named snapshot
// POST /_fazt/snapshot
func AgentSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	appID := r.Header.Get("X-Fazt-App-ID")
	if appID == "" {
		api.BadRequest(w, "X-Fazt-App-ID header required")
		return
	}

	var req SnapshotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}

	if req.Name == "" {
		api.BadRequest(w, "name is required")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	// Ensure snapshots table exists
	db.Exec(`CREATE TABLE IF NOT EXISTS snapshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		app_id TEXT NOT NULL,
		data TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(name, app_id)
	)`)

	// Get current KV data
	rows, err := db.Query("SELECT key, value FROM kv_store WHERE site_id = ?", appID)
	if err != nil {
		api.InternalError(w, err)
		return
	}
	defer rows.Close()

	kvData := make(map[string]string)
	for rows.Next() {
		var key, value string
		if rows.Scan(&key, &value) == nil {
			kvData[key] = value
		}
	}

	// Serialize to JSON
	dataJSON, err := json.Marshal(kvData)
	if err != nil {
		api.InternalError(w, err)
		return
	}

	// Save snapshot
	_, err = db.Exec(`
		INSERT INTO snapshots (name, app_id, data, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(name, app_id) DO UPDATE SET
			data = excluded.data,
			created_at = CURRENT_TIMESTAMP
	`, req.Name, appID, string(dataJSON))
	if err != nil {
		api.InternalError(w, err)
		return
	}

	api.Success(w, http.StatusCreated, map[string]interface{}{
		"name":      req.Name,
		"app_id":    appID,
		"key_count": len(kvData),
		"message":   "Snapshot created",
	})
}

// AgentRestoreHandler restores a named snapshot
// POST /_fazt/restore/:name
func AgentRestoreHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	appID := r.Header.Get("X-Fazt-App-ID")
	snapshotName := r.PathValue("name")
	if appID == "" || snapshotName == "" {
		api.BadRequest(w, "X-Fazt-App-ID header and snapshot name required")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	// Get snapshot data
	var dataJSON string
	err := db.QueryRow("SELECT data FROM snapshots WHERE name = ? AND app_id = ?", snapshotName, appID).Scan(&dataJSON)
	if err == sql.ErrNoRows {
		api.NotFound(w, "SNAPSHOT_NOT_FOUND", "Snapshot not found")
		return
	}
	if err != nil {
		api.InternalError(w, err)
		return
	}

	// Parse data
	var kvData map[string]string
	if err := json.Unmarshal([]byte(dataJSON), &kvData); err != nil {
		api.InternalError(w, err)
		return
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		api.InternalError(w, err)
		return
	}
	defer tx.Rollback()

	// Clear current KV data
	_, err = tx.Exec("DELETE FROM kv_store WHERE site_id = ?", appID)
	if err != nil {
		api.InternalError(w, err)
		return
	}

	// Restore from snapshot
	for key, value := range kvData {
		_, err = tx.Exec("INSERT INTO kv_store (site_id, key, value) VALUES (?, ?, ?)", appID, key, value)
		if err != nil {
			api.InternalError(w, err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		api.InternalError(w, err)
		return
	}

	api.Success(w, http.StatusOK, map[string]interface{}{
		"name":      snapshotName,
		"app_id":    appID,
		"key_count": len(kvData),
		"message":   "Snapshot restored",
	})
}

// AgentSnapshotsListHandler lists available snapshots
// GET /_fazt/snapshots
func AgentSnapshotsListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	appID := r.Header.Get("X-Fazt-App-ID")
	if appID == "" {
		api.BadRequest(w, "X-Fazt-App-ID header required")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	rows, err := db.Query("SELECT name, created_at FROM snapshots WHERE app_id = ? ORDER BY created_at DESC", appID)
	if err != nil {
		// Table might not exist yet
		api.Success(w, http.StatusOK, []Snapshot{})
		return
	}
	defer rows.Close()

	var snapshots []Snapshot
	for rows.Next() {
		var s Snapshot
		var createdAt time.Time
		if rows.Scan(&s.Name, &createdAt) == nil {
			s.AppID = appID
			s.CreatedAt = createdAt.Format(time.RFC3339)
			snapshots = append(snapshots, s)
		}
	}

	api.Success(w, http.StatusOK, snapshots)
}

// LogEntry represents a serverless execution log
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Path      string `json:"path,omitempty"`
	Duration  int64  `json:"duration_ms,omitempty"`
}

// AgentLogsHandler returns recent serverless execution logs
// GET /_fazt/logs
func AgentLogsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	appID := r.Header.Get("X-Fazt-App-ID")
	if appID == "" {
		api.BadRequest(w, "X-Fazt-App-ID header required")
		return
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	rows, err := db.Query(`
		SELECT level, message, created_at
		FROM site_logs
		WHERE site_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, appID, limit)
	if err != nil {
		api.InternalError(w, err)
		return
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var log LogEntry
		var createdAt time.Time
		if rows.Scan(&log.Level, &log.Message, &createdAt) == nil {
			log.Timestamp = createdAt.Format(time.RFC3339)
			logs = append(logs, log)
		}
	}

	api.Success(w, http.StatusOK, logs)
}

// AgentErrorsHandler returns recent errors with stack traces
// GET /_fazt/errors
func AgentErrorsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	appID := r.Header.Get("X-Fazt-App-ID")
	if appID == "" {
		api.BadRequest(w, "X-Fazt-App-ID header required")
		return
	}

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	rows, err := db.Query(`
		SELECT level, message, created_at
		FROM site_logs
		WHERE site_id = ? AND level = 'error'
		ORDER BY created_at DESC
		LIMIT ?
	`, appID, limit)
	if err != nil {
		api.InternalError(w, err)
		return
	}
	defer rows.Close()

	var errors []LogEntry
	for rows.Next() {
		var log LogEntry
		var createdAt time.Time
		if rows.Scan(&log.Level, &log.Message, &createdAt) == nil {
			log.Timestamp = createdAt.Format(time.RFC3339)
			errors = append(errors, log)
		}
	}

	api.Success(w, http.StatusOK, errors)
}
