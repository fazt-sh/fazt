package handlers

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/fazt-sh/fazt/internal/api"
	"github.com/fazt-sh/fazt/internal/appid"
	"github.com/fazt-sh/fazt/internal/assets"
	"github.com/fazt-sh/fazt/internal/config"
	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/hosting"
)

// AppV2 represents an app with v0.10 identity model
type AppV2 struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Description  string   `json:"description,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Visibility   string   `json:"visibility"`
	Source       string   `json:"source"`
	SourceURL    string   `json:"source_url,omitempty"`
	SourceRef    string   `json:"source_ref,omitempty"`
	SourceCommit string   `json:"source_commit,omitempty"`
	OriginalID   string   `json:"original_id,omitempty"`
	ForkedFromID string   `json:"forked_from_id,omitempty"`
	FileCount    int      `json:"file_count"`
	SizeBytes    int64    `json:"size_bytes"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
	Aliases      []string `json:"aliases,omitempty"` // Associated aliases
}

// AppsListHandlerV2 returns the list of apps with v0.10 schema
func AppsListHandlerV2(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	// Check if visibility filter is requested (public API vs admin API)
	showAll := r.URL.Query().Get("all") == "true"

	query := `
		SELECT
			a.id,
			COALESCE(a.title, '') as title,
			COALESCE(a.description, '') as description,
			COALESCE(a.tags, '[]') as tags,
			COALESCE(a.visibility, 'unlisted') as visibility,
			COALESCE(a.source, 'deploy') as source,
			COALESCE(a.source_url, '') as source_url,
			COALESCE(a.source_ref, '') as source_ref,
			COALESCE(a.source_commit, '') as source_commit,
			COALESCE(a.original_id, '') as original_id,
			COALESCE(a.forked_from_id, '') as forked_from_id,
			a.created_at,
			a.updated_at,
			COALESCE(COUNT(f.path), 0) as file_count,
			COALESCE(SUM(f.size_bytes), 0) as size_bytes
		FROM apps a
		LEFT JOIN files f ON a.id = f.app_id
	`

	if !showAll {
		query += " WHERE a.visibility = 'public'"
	}

	query += `
		GROUP BY a.id
		ORDER BY a.updated_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		api.InternalError(w, err)
		return
	}
	defer rows.Close()

	// Collect all apps first, then close cursor before querying aliases
	// (avoids nested query deadlock — Issue 05)
	var apps []AppV2
	for rows.Next() {
		var app AppV2
		var tagsJSON string
		var createdAt, updatedAt interface{}

		err := rows.Scan(
			&app.ID,
			&app.Title,
			&app.Description,
			&tagsJSON,
			&app.Visibility,
			&app.Source,
			&app.SourceURL,
			&app.SourceRef,
			&app.SourceCommit,
			&app.OriginalID,
			&app.ForkedFromID,
			&createdAt,
			&updatedAt,
			&app.FileCount,
			&app.SizeBytes,
		)
		if err != nil {
			continue
		}

		// Parse tags
		if tagsJSON != "" {
			json.Unmarshal([]byte(tagsJSON), &app.Tags)
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
	rows.Close()

	// Now safe to query aliases — cursor is closed
	for i := range apps {
		apps[i].Aliases = getAliasesForApp(db, apps[i].ID)
	}

	api.Success(w, http.StatusOK, apps)
}

// AppDetailHandlerV2 returns details for a single app by ID or alias
func AppDetailHandlerV2(w http.ResponseWriter, r *http.Request) {
	identifier := r.PathValue("id")
	if identifier == "" {
		api.BadRequest(w, "id or alias required")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	// Determine if this is an app ID or alias
	var appID string
	if appid.IsValid(identifier) {
		appID = identifier
	} else {
		// Try to resolve as alias
		resolvedID, aliasType, err := ResolveAlias(identifier)
		if err != nil {
			api.InternalError(w, err)
			return
		}
		if aliasType == "reserved" {
			api.NotFound(w, "RESERVED", "Subdomain is reserved")
			return
		}
		if resolvedID == "" {
			api.NotFound(w, "NOT_FOUND", "App or alias not found")
			return
		}
		appID = resolvedID
	}

	app, err := getAppByID(db, appID)
	if err == sql.ErrNoRows {
		api.NotFound(w, "APP_NOT_FOUND", "App not found")
		return
	}
	if err != nil {
		api.InternalError(w, err)
		return
	}

	api.Success(w, http.StatusOK, app)
}

// AppCreateRequest represents a request to create an app
type AppCreateRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Visibility  string   `json:"visibility"`
	Alias       string   `json:"alias"` // Optional alias to create
	Template    string   `json:"template"`
}

// AppCreateHandlerV2 creates a new app
func AppCreateHandlerV2(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	var req AppCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}

	if req.Title == "" {
		api.BadRequest(w, "title is required")
		return
	}

	// Validate visibility
	if req.Visibility == "" {
		req.Visibility = "unlisted"
	}
	if req.Visibility != "public" && req.Visibility != "unlisted" && req.Visibility != "private" {
		api.BadRequest(w, "visibility must be 'public', 'unlisted', or 'private'")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	// Generate new app ID
	newID := appid.GenerateApp()

	// Build tags JSON
	tagsJSON := "[]"
	if len(req.Tags) > 0 {
		b, _ := json.Marshal(req.Tags)
		tagsJSON = string(b)
	}

	// Insert app
	query := `
		INSERT INTO apps (id, original_id, title, description, tags, visibility, source, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, 'template', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	_, err := db.Exec(query, newID, newID, req.Title, req.Description, tagsJSON, req.Visibility)
	if err != nil {
		api.InternalError(w, err)
		return
	}

	// Create alias if requested
	if req.Alias != "" {
		if !isValidSubdomain(req.Alias) {
			api.BadRequest(w, "invalid alias format")
			return
		}

		aliasQuery := `
			INSERT INTO aliases (subdomain, type, targets, created_at, updated_at)
			VALUES (?, 'proxy', ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`
		targets := fmt.Sprintf(`{"app_id":"%s"}`, newID)
		_, err = db.Exec(aliasQuery, req.Alias, targets)
		if err != nil {
			// Alias might already exist, continue anyway
		}
	}

	// If template specified, deploy template files
	if req.Template != "" {
		if err := deployTemplate(db, newID, req.Template, req.Title); err != nil {
			// Log but don't fail - app is created
		}
	}

	cfg := config.Get()
	result := map[string]interface{}{
		"id":         newID,
		"title":      req.Title,
		"visibility": req.Visibility,
	}

	if req.Alias != "" {
		result["alias"] = req.Alias
		result["url"] = fmt.Sprintf("https://%s.%s", req.Alias, cfg.Server.Domain)
	}

	api.Success(w, http.StatusCreated, result)
}

// AppUpdateRequest represents a request to update app metadata
type AppUpdateRequest struct {
	Title       *string  `json:"title,omitempty"`
	Description *string  `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Visibility  *string  `json:"visibility,omitempty"`
}

// AppUpdateHandlerV2 updates app metadata
func AppUpdateHandlerV2(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	appID := r.PathValue("id")
	if appID == "" {
		api.BadRequest(w, "id required")
		return
	}

	var req AppUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	// Verify app exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM apps WHERE id = ?", appID).Scan(&count)
	if err != nil || count == 0 {
		api.NotFound(w, "APP_NOT_FOUND", "App not found")
		return
	}

	// Build update query dynamically
	updates := []string{}
	args := []interface{}{}

	if req.Title != nil {
		updates = append(updates, "title = ?")
		args = append(args, *req.Title)
	}
	if req.Description != nil {
		updates = append(updates, "description = ?")
		args = append(args, *req.Description)
	}
	if len(req.Tags) > 0 {
		tagsJSON, _ := json.Marshal(req.Tags)
		updates = append(updates, "tags = ?")
		args = append(args, string(tagsJSON))
	}
	if req.Visibility != nil {
		if *req.Visibility != "public" && *req.Visibility != "unlisted" && *req.Visibility != "private" {
			api.BadRequest(w, "visibility must be 'public', 'unlisted', or 'private'")
			return
		}
		updates = append(updates, "visibility = ?")
		args = append(args, *req.Visibility)
	}

	if len(updates) == 0 {
		api.BadRequest(w, "no fields to update")
		return
	}

	updates = append(updates, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, appID)

	query := "UPDATE apps SET " + strings.Join(updates, ", ") + " WHERE id = ?"
	_, err = db.Exec(query, args...)
	if err != nil {
		api.InternalError(w, err)
		return
	}

	api.Success(w, http.StatusOK, map[string]interface{}{
		"id":      appID,
		"message": "App updated",
	})
}

// AppDeleteHandlerV2 deletes an app
func AppDeleteHandlerV2(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	appID := r.PathValue("id")
	if appID == "" {
		api.BadRequest(w, "id required")
		return
	}

	withForks := r.URL.Query().Get("with-forks") == "true"

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	// Verify app exists and get info
	var title, source string
	err := db.QueryRow("SELECT COALESCE(title, ''), COALESCE(source, '') FROM apps WHERE id = ?", appID).Scan(&title, &source)
	if err == sql.ErrNoRows {
		api.NotFound(w, "APP_NOT_FOUND", "App not found")
		return
	}
	if err != nil {
		api.InternalError(w, err)
		return
	}

	// Don't allow deleting system apps
	if source == "system" {
		api.ErrorResponse(w, http.StatusForbidden, "SYSTEM_APP", "Cannot delete system app", "")
		return
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		api.InternalError(w, err)
		return
	}
	defer tx.Rollback()

	idsToDelete := []string{appID}

	// If with-forks, find all forks
	if withForks {
		// Find all apps with this original_id
		rows, err := tx.Query("SELECT id FROM apps WHERE original_id = ? AND id != ?", appID, appID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id string
				if rows.Scan(&id) == nil {
					idsToDelete = append(idsToDelete, id)
				}
			}
		}
	}

	// Delete files for all apps
	for _, id := range idsToDelete {
		_, err = tx.Exec("DELETE FROM files WHERE app_id = ?", id)
		if err != nil {
			api.InternalError(w, err)
			return
		}
	}

	// Delete apps
	for _, id := range idsToDelete {
		_, err = tx.Exec("DELETE FROM apps WHERE id = ?", id)
		if err != nil {
			api.InternalError(w, err)
			return
		}
	}

	// Remove aliases pointing to deleted apps (orphan cleanup)
	for _, id := range idsToDelete {
		_, err = tx.Exec("DELETE FROM aliases WHERE targets LIKE ?", `%"`+id+`"%`)
		if err != nil {
			// Non-fatal, continue
		}
	}

	if err := tx.Commit(); err != nil {
		api.InternalError(w, err)
		return
	}

	api.Success(w, http.StatusOK, map[string]interface{}{
		"id":      appID,
		"title":   title,
		"deleted": len(idsToDelete),
		"message": "App deleted",
	})
}

// ForkRequest represents a request to fork an app
type ForkRequest struct {
	Alias       string `json:"alias"`        // Optional new alias
	CopyStorage bool   `json:"copy_storage"` // Whether to copy KV storage
}

// AppForkHandler forks an app
func AppForkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	appID := r.PathValue("id")
	if appID == "" {
		api.BadRequest(w, "id required")
		return
	}

	var req ForkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Allow empty body
		req = ForkRequest{}
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	// Get source app
	sourceApp, err := getAppByID(db, appID)
	if err == sql.ErrNoRows {
		api.NotFound(w, "APP_NOT_FOUND", "App not found")
		return
	}
	if err != nil {
		api.InternalError(w, err)
		return
	}

	// Generate new app ID
	newID := appid.GenerateApp()

	// Build tags JSON
	tagsJSON := "[]"
	if len(sourceApp.Tags) > 0 {
		b, _ := json.Marshal(sourceApp.Tags)
		tagsJSON = string(b)
	}

	// Determine original_id (root of lineage)
	originalID := sourceApp.OriginalID
	if originalID == "" {
		originalID = sourceApp.ID
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		api.InternalError(w, err)
		return
	}
	defer tx.Rollback()

	// Insert forked app
	query := `
		INSERT INTO apps (id, original_id, forked_from_id, title, description, tags, visibility, source, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, 'unlisted', 'fork', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	_, err = tx.Exec(query, newID, originalID, sourceApp.ID, sourceApp.Title, sourceApp.Description, tagsJSON)
	if err != nil {
		api.InternalError(w, err)
		return
	}

	// Copy files
	copyQuery := `
		INSERT INTO files (site_id, app_id, path, content, size_bytes, mime_type, hash, created_at, updated_at)
		SELECT site_id, ?, path, content, size_bytes, mime_type, hash, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		FROM files WHERE app_id = ?
	`
	_, err = tx.Exec(copyQuery, newID, sourceApp.ID)
	if err != nil {
		api.InternalError(w, err)
		return
	}

	// Copy KV storage if requested
	if req.CopyStorage {
		kvQuery := `
			INSERT INTO kv_store (site_id, key, value, created_at, updated_at)
			SELECT ?, key, value, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
			FROM kv_store WHERE site_id = ?
		`
		tx.Exec(kvQuery, newID, sourceApp.ID) // Ignore errors if kv_store doesn't exist
	}

	// Create alias if requested
	if req.Alias != "" {
		if !isValidSubdomain(req.Alias) {
			api.BadRequest(w, "invalid alias format")
			return
		}

		aliasQuery := `
			INSERT INTO aliases (subdomain, type, targets, created_at, updated_at)
			VALUES (?, 'proxy', ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`
		targets := fmt.Sprintf(`{"app_id":"%s"}`, newID)
		_, err = tx.Exec(aliasQuery, req.Alias, targets)
		if err != nil {
			// Alias might already exist
			api.BadRequest(w, "alias already exists")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		api.InternalError(w, err)
		return
	}

	cfg := config.Get()
	result := map[string]interface{}{
		"id":             newID,
		"title":          sourceApp.Title,
		"forked_from_id": sourceApp.ID,
		"original_id":    originalID,
	}

	if req.Alias != "" {
		result["alias"] = req.Alias
		result["url"] = fmt.Sprintf("https://%s.%s", req.Alias, cfg.Server.Domain)
	}

	api.Success(w, http.StatusCreated, result)
}

// LineageNode represents a node in the lineage tree
type LineageNode struct {
	ID      string        `json:"id"`
	Title   string        `json:"title"`
	Aliases []string      `json:"aliases,omitempty"`
	Forks   []LineageNode `json:"forks,omitempty"`
}

// AppLineageHandler returns the lineage tree for an app
func AppLineageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	appID := r.PathValue("id")
	if appID == "" {
		api.BadRequest(w, "id required")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	// Get the app and its original_id
	var originalID string
	err := db.QueryRow("SELECT COALESCE(original_id, id) FROM apps WHERE id = ?", appID).Scan(&originalID)
	if err == sql.ErrNoRows {
		api.NotFound(w, "APP_NOT_FOUND", "App not found")
		return
	}
	if err != nil {
		api.InternalError(w, err)
		return
	}

	// Build lineage tree starting from original
	tree := buildLineageTree(db, originalID, nil)

	api.Success(w, http.StatusOK, tree)
}

// AppForksHandler returns direct forks of an app
func AppForksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	appID := r.PathValue("id")
	if appID == "" {
		api.BadRequest(w, "id required")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	query := `
		SELECT id, COALESCE(title, '') as title
		FROM apps WHERE forked_from_id = ?
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query, appID)
	if err != nil {
		api.InternalError(w, err)
		return
	}
	defer rows.Close()

	// Collect fork data first, then close cursor before querying aliases
	// (avoids nested query deadlock — Issue 05)
	type forkData struct {
		id, title string
	}
	var forkList []forkData
	for rows.Next() {
		var id, title string
		if rows.Scan(&id, &title) == nil {
			forkList = append(forkList, forkData{id, title})
		}
	}
	rows.Close()

	// Now safe to query aliases — cursor is closed
	var forks []map[string]interface{}
	for _, f := range forkList {
		forks = append(forks, map[string]interface{}{
			"id":      f.id,
			"title":   f.title,
			"aliases": getAliasesForApp(db, f.id),
		})
	}

	api.Success(w, http.StatusOK, forks)
}

// Helper functions

func getAppByID(db *sql.DB, appID string) (*AppV2, error) {
	query := `
		SELECT
			a.id,
			COALESCE(a.title, '') as title,
			COALESCE(a.description, '') as description,
			COALESCE(a.tags, '[]') as tags,
			COALESCE(a.visibility, 'unlisted') as visibility,
			COALESCE(a.source, 'deploy') as source,
			COALESCE(a.source_url, '') as source_url,
			COALESCE(a.source_ref, '') as source_ref,
			COALESCE(a.source_commit, '') as source_commit,
			COALESCE(a.original_id, '') as original_id,
			COALESCE(a.forked_from_id, '') as forked_from_id,
			a.created_at,
			a.updated_at,
			COALESCE(COUNT(f.path), 0) as file_count,
			COALESCE(SUM(f.size_bytes), 0) as size_bytes
		FROM apps a
		LEFT JOIN files f ON a.id = f.app_id
		WHERE a.id = ?
		GROUP BY a.id
	`

	var app AppV2
	var tagsJSON string
	var createdAt, updatedAt interface{}

	err := db.QueryRow(query, appID).Scan(
		&app.ID,
		&app.Title,
		&app.Description,
		&tagsJSON,
		&app.Visibility,
		&app.Source,
		&app.SourceURL,
		&app.SourceRef,
		&app.SourceCommit,
		&app.OriginalID,
		&app.ForkedFromID,
		&createdAt,
		&updatedAt,
		&app.FileCount,
		&app.SizeBytes,
	)
	if err != nil {
		return nil, err
	}

	// Parse tags
	if tagsJSON != "" {
		json.Unmarshal([]byte(tagsJSON), &app.Tags)
	}

	// Format timestamps
	if createdAt != nil {
		app.CreatedAt = formatTime(createdAt)
	}
	if updatedAt != nil {
		app.UpdatedAt = formatTime(updatedAt)
	}

	// Get aliases
	app.Aliases = getAliasesForApp(db, app.ID)

	return &app, nil
}

func getAliasesForApp(db *sql.DB, appID string) []string {
	var aliases []string
	rows, err := db.Query("SELECT subdomain FROM aliases WHERE targets LIKE ?", `%"`+appID+`"%`)
	if err != nil {
		return aliases
	}
	defer rows.Close()

	for rows.Next() {
		var subdomain string
		if rows.Scan(&subdomain) == nil {
			aliases = append(aliases, subdomain)
		}
	}
	return aliases
}

func buildLineageTree(db *sql.DB, appID string, visited map[string]bool) *LineageNode {
	if visited == nil {
		visited = make(map[string]bool)
	}
	if visited[appID] {
		return nil // Prevent cycles
	}
	visited[appID] = true

	var title string
	db.QueryRow("SELECT COALESCE(title, '') FROM apps WHERE id = ?", appID).Scan(&title)

	node := &LineageNode{
		ID:      appID,
		Title:   title,
		Aliases: getAliasesForApp(db, appID),
	}

	// Collect fork IDs first, then close cursor before recursing
	// (avoids nested query deadlock — Issue 05)
	rows, err := db.Query("SELECT id FROM apps WHERE forked_from_id = ?", appID)
	if err != nil {
		return node
	}
	var forkIDs []string
	for rows.Next() {
		var forkID string
		if rows.Scan(&forkID) == nil {
			forkIDs = append(forkIDs, forkID)
		}
	}
	rows.Close()

	// Now safe to recurse — cursor is closed
	for _, forkID := range forkIDs {
		if fork := buildLineageTree(db, forkID, visited); fork != nil {
			node.Forks = append(node.Forks, *fork)
		}
	}

	return node
}

func deployTemplate(db *sql.DB, appID, templateName, appTitle string) error {
	tmplFS, err := assets.GetTemplate(templateName)
	if err != nil {
		return err
	}

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "fazt-template-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	// Copy template with substitution
	data := map[string]string{"Name": appTitle}
	if err := copyTemplateFilesV2(tmplFS, tmpDir, data); err != nil {
		return err
	}

	// Create zip and deploy
	zipData, err := createZipFromDirV2(tmpDir)
	if err != nil {
		return err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return err
	}

	// Deploy to VFS with app_id
	return deployFilesToApp(db, appID, zipReader)
}

func deployFilesToApp(db *sql.DB, appID string, zipReader *zip.Reader) error {
	for _, file := range zipReader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		cleanPath := filepath.Clean(file.Name)
		cleanPath = filepath.ToSlash(cleanPath)

		src, err := file.Open()
		if err != nil {
			continue
		}

		data, err := io.ReadAll(src)
		src.Close()
		if err != nil {
			continue
		}

		// Write to files table with app_id
		query := `
			INSERT INTO files (site_id, app_id, path, content, size_bytes, mime_type, hash, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			ON CONFLICT(site_id, path) DO UPDATE SET
				content = excluded.content,
				size_bytes = excluded.size_bytes,
				mime_type = excluded.mime_type,
				hash = excluded.hash,
				updated_at = CURRENT_TIMESTAMP
		`
		_, err = db.Exec(query, appID, appID, cleanPath, data, len(data), hosting.GetMimeType(cleanPath), "", appID)
		if err != nil {
			return err
		}
	}

	return nil
}

func copyTemplateFilesV2(tmplFS fs.FS, destDir string, data map[string]string) error {
	return fs.WalkDir(tmplFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}

		destPath := filepath.Join(destDir, path)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		content, err := fs.ReadFile(tmplFS, path)
		if err != nil {
			return err
		}

		// Apply template substitution
		tmpl, err := template.New(path).Parse(string(content))
		if err != nil {
			return os.WriteFile(destPath, content, 0644)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return os.WriteFile(destPath, content, 0644)
		}

		return os.WriteFile(destPath, buf.Bytes(), 0644)
	})
}

func createZipFromDirV2(srcDir string) ([]byte, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || info.Name()[0] == '.' {
			return nil
		}

		relPath, _ := filepath.Rel(srcDir, path)
		if strings.HasPrefix(relPath, "node_modules") {
			return nil
		}

		writer, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	if err != nil {
		return nil, err
	}

	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// copyTemplateFiles is imported from apps_handler.go
var _ = template.New // Import template package
