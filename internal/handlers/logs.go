package handlers

import (
	"net/http"
	"strconv"

	"github.com/fazt-sh/fazt/internal/api"
	"github.com/fazt-sh/fazt/internal/database"
)

// LogsHandler returns logs for a specific site
func LogsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.BadRequest(w, "Method not allowed")
		return
	}

	siteID := r.URL.Query().Get("site_id")
	if siteID == "" {
		api.BadRequest(w, "site_id required")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	db := database.GetDB()
	rows, err := db.Query(`
		SELECT id, level, message, created_at
		FROM site_logs
		WHERE site_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, siteID, limit)
	if err != nil {
		api.InternalError(w, err)
		return
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var id int64
		var level, message, createdAt string
		if err := rows.Scan(&id, &level, &message, &createdAt); err != nil {
			continue
		}
		logs = append(logs, map[string]interface{}{
			"id":         id,
			"level":      level,
			"message":    message,
			"created_at": createdAt,
		})
	}

	api.Success(w, http.StatusOK, map[string]interface{}{
		"logs": logs,
	})
}
