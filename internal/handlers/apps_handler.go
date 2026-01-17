package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fazt-sh/fazt/internal/api"
	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/hosting"
)

// App represents an app with metadata
type App struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Source    string      `json:"source"`
	Manifest  interface{} `json:"manifest,omitempty"`
	FileCount int         `json:"file_count"`
	SizeBytes int64       `json:"size_bytes"`
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"updated_at"`
}

// AppsListHandler returns the list of apps with metadata
func AppsListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	// Query apps with file stats
	query := `
		SELECT
			a.id,
			a.name,
			a.source,
			a.manifest,
			a.created_at,
			a.updated_at,
			COALESCE(COUNT(f.path), 0) as file_count,
			COALESCE(SUM(f.size_bytes), 0) as size_bytes
		FROM apps a
		LEFT JOIN files f ON a.name = f.site_id
		GROUP BY a.id
		ORDER BY a.updated_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		api.InternalError(w, err)
		return
	}
	defer rows.Close()

	var apps []App
	for rows.Next() {
		var app App
		var manifest *string
		var createdAt, updatedAt interface{}

		err := rows.Scan(
			&app.ID,
			&app.Name,
			&app.Source,
			&manifest,
			&createdAt,
			&updatedAt,
			&app.FileCount,
			&app.SizeBytes,
		)
		if err != nil {
			continue
		}

		// Parse manifest if present
		if manifest != nil && *manifest != "" {
			json.Unmarshal([]byte(*manifest), &app.Manifest)
		}

		// Format timestamps
		if createdAt != nil {
			app.CreatedAt = formatTime(createdAt)
		}
		if updatedAt != nil {
			app.UpdatedAt = formatTime(updatedAt)
		}

		apps = append(apps, app)
	}

	api.Success(w, http.StatusOK, apps)
}

// AppDetailHandler returns details for a single app
func AppDetailHandler(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("id")
	if appID == "" {
		api.BadRequest(w, "app_id required")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	// Query app by id or name
	query := `
		SELECT
			a.id,
			a.name,
			a.source,
			a.manifest,
			a.created_at,
			a.updated_at,
			COALESCE(COUNT(f.path), 0) as file_count,
			COALESCE(SUM(f.size_bytes), 0) as size_bytes
		FROM apps a
		LEFT JOIN files f ON a.name = f.site_id
		WHERE a.id = ? OR a.name = ?
		GROUP BY a.id
	`

	var app App
	var manifest *string
	var createdAt, updatedAt interface{}

	err := db.QueryRow(query, appID, appID).Scan(
		&app.ID,
		&app.Name,
		&app.Source,
		&manifest,
		&createdAt,
		&updatedAt,
		&app.FileCount,
		&app.SizeBytes,
	)
	if err != nil {
		api.NotFound(w, "APP_NOT_FOUND", "App not found")
		return
	}

	// Parse manifest if present
	if manifest != nil && *manifest != "" {
		json.Unmarshal([]byte(*manifest), &app.Manifest)
	}

	// Format timestamps
	if createdAt != nil {
		app.CreatedAt = formatTime(createdAt)
	}
	if updatedAt != nil {
		app.UpdatedAt = formatTime(updatedAt)
	}

	api.Success(w, http.StatusOK, app)
}

// AppDeleteHandler deletes an app
func AppDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	appID := r.PathValue("id")
	if appID == "" {
		api.BadRequest(w, "app_id required")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	// Get app name for file deletion
	var name string
	err := db.QueryRow("SELECT name FROM apps WHERE id = ? OR name = ?", appID, appID).Scan(&name)
	if err != nil {
		api.NotFound(w, "APP_NOT_FOUND", "App not found")
		return
	}

	// Don't allow deleting system apps
	if name == "root" || name == "404" || name == "admin" {
		api.ErrorResponse(w, http.StatusForbidden, "SYSTEM_APP", "Cannot delete system app", "")
		return
	}

	// Delete files via hosting
	if err := hosting.DeleteSite(name); err != nil {
		api.InternalError(w, err)
		return
	}

	// Delete from apps table
	_, err = db.Exec("DELETE FROM apps WHERE id = ? OR name = ?", appID, appID)
	if err != nil {
		api.InternalError(w, err)
		return
	}

	api.Success(w, http.StatusOK, map[string]interface{}{
		"message": "App deleted",
		"name":    name,
	})
}

// AppFilesHandler returns the file tree for an app
func AppFilesHandler(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("id")
	if appID == "" {
		api.BadRequest(w, "app_id required")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	// Get app name
	var name string
	err := db.QueryRow("SELECT name FROM apps WHERE id = ? OR name = ?", appID, appID).Scan(&name)
	if err != nil {
		api.NotFound(w, "APP_NOT_FOUND", "App not found")
		return
	}

	fs := hosting.GetFileSystem()
	files, err := fs.ListFiles(name)
	if err != nil {
		api.InternalError(w, err)
		return
	}

	api.Success(w, http.StatusOK, files)
}

// formatTime converts interface{} time to string
func formatTime(t interface{}) string {
	if s, ok := t.(string); ok {
		return s
	}
	return ""
}

// AppSourceHandler returns source tracking info for an app
func AppSourceHandler(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("id")
	if appID == "" {
		api.BadRequest(w, "app_id required")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	query := `
		SELECT source, source_url, source_ref, source_commit
		FROM apps WHERE name = ? OR id = ?
	`

	var sourceType string
	var sourceURL, sourceRef, sourceCommit *string

	err := db.QueryRow(query, appID, appID).Scan(&sourceType, &sourceURL, &sourceRef, &sourceCommit)
	if err != nil {
		api.NotFound(w, "APP_NOT_FOUND", "App not found")
		return
	}

	result := map[string]string{
		"type": sourceType,
	}
	if sourceURL != nil {
		result["url"] = *sourceURL
	}
	if sourceRef != nil {
		result["ref"] = *sourceRef
	}
	if sourceCommit != nil {
		result["commit"] = *sourceCommit
	}

	api.Success(w, http.StatusOK, result)
}

// AppFileContentHandler returns the content of a specific file
func AppFileContentHandler(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("id")
	filePath := r.PathValue("path")

	if appID == "" || filePath == "" {
		api.BadRequest(w, "app_id and path required")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	// Get app name
	var name string
	err := db.QueryRow("SELECT name FROM apps WHERE id = ? OR name = ?", appID, appID).Scan(&name)
	if err != nil {
		api.NotFound(w, "APP_NOT_FOUND", "App not found")
		return
	}

	fs := hosting.GetFileSystem()
	file, err := fs.ReadFile(name, filePath)
	if err != nil {
		api.NotFound(w, "FILE_NOT_FOUND", "File not found")
		return
	}
	defer file.Content.Close()

	// Return raw content with correct MIME type
	w.Header().Set("Content-Type", file.MimeType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", file.Size))
	w.WriteHeader(http.StatusOK)

	buf := make([]byte, file.Size)
	file.Content.Read(buf)
	w.Write(buf)
}
