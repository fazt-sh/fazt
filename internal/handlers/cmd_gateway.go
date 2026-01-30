package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/fazt-sh/fazt/internal/api"
	"github.com/fazt-sh/fazt/internal/appid"
	"github.com/fazt-sh/fazt/internal/config"
	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/hosting"
)

// CmdRequest represents a command gateway request
type CmdRequest struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// CmdResponse represents a command gateway response
type CmdResponse struct {
	Success bool        `json:"success"`
	Output  string      `json:"output,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// CmdGatewayHandler handles POST /api/cmd for remote command execution
func CmdGatewayHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	var req CmdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}

	if req.Command == "" {
		api.BadRequest(w, "command is required")
		return
	}

	// Route command to appropriate handler
	result, err := executeCommand(req.Command, req.Args)
	if err != nil {
		api.Success(w, http.StatusOK, CmdResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	api.Success(w, http.StatusOK, CmdResponse{
		Success: true,
		Data:    result,
	})
}

// executeCommand routes a command to the appropriate handler
func executeCommand(command string, args []string) (interface{}, error) {
	db := database.GetDB()
	if db == nil {
		return nil, ErrDatabaseNotInitialized
	}

	switch command {
	case "app":
		return executeAppCommand(args)
	case "server":
		return executeServerCommand(args)
	default:
		return nil, ErrUnknownCommand
	}
}

// executeAppCommand handles app subcommands
func executeAppCommand(args []string) (interface{}, error) {
	if len(args) < 1 {
		return nil, ErrMissingSubcommand
	}

	db := database.GetDB()
	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "list":
		return cmdAppList(db, subArgs)
	case "info":
		return cmdAppInfo(db, subArgs)
	case "remove":
		return cmdAppRemove(db, subArgs)
	case "link":
		return cmdAppLink(db, subArgs)
	case "unlink":
		return cmdAppUnlink(db, subArgs)
	case "reserve":
		return cmdAppReserve(db, subArgs)
	case "fork":
		return cmdAppFork(db, subArgs)
	case "lineage":
		return cmdAppLineage(db, subArgs)
	default:
		return nil, ErrUnknownSubcommand
	}
}

// executeServerCommand handles server subcommands
func executeServerCommand(args []string) (interface{}, error) {
	if len(args) < 1 {
		return nil, ErrMissingSubcommand
	}

	subcommand := args[0]

	switch subcommand {
	case "info":
		return cmdServerInfo()
	default:
		return nil, ErrUnknownSubcommand
	}
}

// Command implementations

func cmdAppList(db interface{}, args []string) (interface{}, error) {
	sqlDB := database.GetDB()

	// Check for --aliases flag
	showAliases := false
	for _, arg := range args {
		if arg == "--aliases" {
			showAliases = true
		}
	}

	if showAliases {
		// Return aliases list
		query := `
			SELECT subdomain, type, targets
			FROM aliases
			ORDER BY subdomain
		`
		rows, err := sqlDB.Query(query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var aliases []map[string]interface{}
		for rows.Next() {
			var subdomain, aliasType string
			var targets *string
			if rows.Scan(&subdomain, &aliasType, &targets) == nil {
				alias := map[string]interface{}{
					"subdomain": subdomain,
					"type":      aliasType,
				}
				if targets != nil {
					var t map[string]interface{}
					if json.Unmarshal([]byte(*targets), &t) == nil {
						alias["targets"] = t
					}
				}
				aliases = append(aliases, alias)
			}
		}
		return aliases, nil
	}

	// Return apps list
	query := `
		SELECT
			a.id,
			COALESCE(a.title, '') as title,
			COALESCE(a.visibility, 'unlisted') as visibility,
			COALESCE(a.tags, '[]') as tags,
			COALESCE(a.forked_from_id, '') as forked_from
		FROM apps a
		ORDER BY a.updated_at DESC
	`
	rows, err := sqlDB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []map[string]interface{}
	for rows.Next() {
		var id, title, visibility, tags, forkedFrom string
		if rows.Scan(&id, &title, &visibility, &tags, &forkedFrom) == nil {
			app := map[string]interface{}{
				"id":         id,
				"title":      title,
				"visibility": visibility,
			}

			var tagsList []string
			if json.Unmarshal([]byte(tags), &tagsList) == nil && len(tagsList) > 0 {
				app["tags"] = tagsList
			}

			if forkedFrom != "" {
				app["forked_from"] = forkedFrom
			}

			// Get aliases
			aliases := getAliasesForApp(sqlDB, id)
			if len(aliases) > 0 {
				app["aliases"] = aliases
			}

			apps = append(apps, app)
		}
	}

	return apps, nil
}

func cmdAppInfo(db interface{}, args []string) (interface{}, error) {
	if len(args) < 1 {
		return nil, ErrMissingArgument
	}

	sqlDB := database.GetDB()
	identifier := args[0]

	// Check for --alias or --id flags
	useAlias := false
	useID := false
	for i, arg := range args {
		if arg == "--alias" && i+1 < len(args) {
			identifier = args[i+1]
			useAlias = true
		} else if arg == "--id" && i+1 < len(args) {
			identifier = args[i+1]
			useID = true
		}
	}

	var appID string
	if useAlias || (!useID && !appid.IsValid(identifier)) {
		// Resolve alias
		resolvedID, aliasType, err := ResolveAlias(identifier)
		if err != nil {
			return nil, err
		}
		if aliasType == "reserved" {
			return nil, ErrReservedSubdomain
		}
		if resolvedID == "" {
			return nil, ErrNotFound
		}
		appID = resolvedID
	} else {
		appID = identifier
	}

	app, err := getAppByID(sqlDB, appID)
	if err != nil {
		return nil, ErrNotFound
	}

	cfg := config.Get()
	result := map[string]interface{}{
		"id":         app.ID,
		"title":      app.Title,
		"visibility": app.Visibility,
		"source":     app.Source,
		"file_count": app.FileCount,
		"size_bytes": app.SizeBytes,
		"created_at": app.CreatedAt,
		"updated_at": app.UpdatedAt,
	}

	if app.Description != "" {
		result["description"] = app.Description
	}
	if len(app.Tags) > 0 {
		result["tags"] = app.Tags
	}
	if app.OriginalID != "" && app.OriginalID != app.ID {
		result["original_id"] = app.OriginalID
	}
	if app.ForkedFromID != "" {
		result["forked_from_id"] = app.ForkedFromID
	}
	if len(app.Aliases) > 0 {
		result["aliases"] = app.Aliases
		result["url"] = "https://" + app.Aliases[0] + "." + cfg.Server.Domain
	}

	return result, nil
}

func cmdAppRemove(db interface{}, args []string) (interface{}, error) {
	if len(args) < 1 {
		return nil, ErrMissingArgument
	}

	sqlDB := database.GetDB()
	identifier := args[0]
	withForks := false

	// Parse flags
	useAlias := false
	useID := false
	for i, arg := range args {
		if arg == "--alias" && i+1 < len(args) {
			identifier = args[i+1]
			useAlias = true
		} else if arg == "--id" && i+1 < len(args) {
			identifier = args[i+1]
			useID = true
		} else if arg == "--with-forks" {
			withForks = true
		}
	}

	if useAlias && !useID {
		// Remove alias only
		_, err := sqlDB.Exec("DELETE FROM aliases WHERE subdomain = ?", identifier)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"alias":   identifier,
			"message": "Alias removed",
		}, nil
	}

	// Remove app
	var appID string
	if appid.IsValid(identifier) {
		appID = identifier
	} else {
		// Try as alias first
		resolvedID, _, err := ResolveAlias(identifier)
		if err != nil || resolvedID == "" {
			return nil, ErrNotFound
		}
		appID = resolvedID
	}

	// Get app title
	var title string
	err := sqlDB.QueryRow("SELECT COALESCE(title, '') FROM apps WHERE id = ?", appID).Scan(&title)
	if err != nil {
		return nil, ErrNotFound
	}

	// Delete files and app
	idsToDelete := []string{appID}
	if withForks {
		rows, _ := sqlDB.Query("SELECT id FROM apps WHERE original_id = ? AND id != ?", appID, appID)
		if rows != nil {
			defer rows.Close()
			for rows.Next() {
				var id string
				if rows.Scan(&id) == nil {
					idsToDelete = append(idsToDelete, id)
				}
			}
		}
	}

	for _, id := range idsToDelete {
		sqlDB.Exec("DELETE FROM files WHERE app_id = ?", id)
		sqlDB.Exec("DELETE FROM apps WHERE id = ?", id)
		sqlDB.Exec("DELETE FROM aliases WHERE targets LIKE ?", `%"`+id+`"%`)
	}

	return map[string]interface{}{
		"id":      appID,
		"title":   title,
		"deleted": len(idsToDelete),
		"message": "App removed",
	}, nil
}

func cmdAppLink(db interface{}, args []string) (interface{}, error) {
	if len(args) < 1 {
		return nil, ErrMissingArgument
	}

	sqlDB := database.GetDB()
	subdomain := args[0]
	var appID string

	// Parse --id flag
	for i, arg := range args {
		if arg == "--id" && i+1 < len(args) {
			appID = args[i+1]
		}
	}

	if appID == "" {
		return nil, ErrMissingArgument
	}

	// Verify app exists
	var count int
	err := sqlDB.QueryRow("SELECT COUNT(*) FROM apps WHERE id = ?", appID).Scan(&count)
	if err != nil || count == 0 {
		return nil, ErrNotFound
	}

	// Create/update alias
	targets := `{"app_id":"` + appID + `"}`
	query := `
		INSERT INTO aliases (subdomain, type, targets, created_at, updated_at)
		VALUES (?, 'proxy', ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(subdomain) DO UPDATE SET
			type = 'proxy',
			targets = excluded.targets,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err = sqlDB.Exec(query, subdomain, targets)
	if err != nil {
		return nil, err
	}

	cfg := config.Get()
	return map[string]interface{}{
		"subdomain": subdomain,
		"app_id":    appID,
		"url":       "https://" + subdomain + "." + cfg.Server.Domain,
		"message":   "Alias created",
	}, nil
}

func cmdAppUnlink(db interface{}, args []string) (interface{}, error) {
	if len(args) < 1 {
		return nil, ErrMissingArgument
	}

	sqlDB := database.GetDB()
	subdomain := args[0]

	_, err := sqlDB.Exec("DELETE FROM aliases WHERE subdomain = ?", subdomain)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"subdomain": subdomain,
		"message":   "Alias removed",
	}, nil
}

func cmdAppReserve(db interface{}, args []string) (interface{}, error) {
	if len(args) < 1 {
		return nil, ErrMissingArgument
	}

	sqlDB := database.GetDB()
	subdomain := args[0]

	query := `
		INSERT INTO aliases (subdomain, type, targets, created_at, updated_at)
		VALUES (?, 'reserved', NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(subdomain) DO UPDATE SET
			type = 'reserved',
			targets = NULL,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := sqlDB.Exec(query, subdomain)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"subdomain": subdomain,
		"message":   "Subdomain reserved",
	}, nil
}

func cmdAppFork(db interface{}, args []string) (interface{}, error) {
	if len(args) < 1 {
		return nil, ErrMissingArgument
	}

	sqlDB := database.GetDB()
	identifier := args[0]
	var newAlias string
	copyStorage := true

	// Parse flags
	for i, arg := range args {
		if arg == "--as" && i+1 < len(args) {
			newAlias = args[i+1]
		} else if arg == "--no-storage" {
			copyStorage = false
		} else if arg == "--alias" && i+1 < len(args) {
			identifier = args[i+1]
		} else if arg == "--id" && i+1 < len(args) {
			identifier = args[i+1]
		}
	}

	// Resolve source app
	var sourceAppID string
	if appid.IsValid(identifier) {
		sourceAppID = identifier
	} else {
		resolvedID, _, err := ResolveAlias(identifier)
		if err != nil || resolvedID == "" {
			return nil, ErrNotFound
		}
		sourceAppID = resolvedID
	}

	// Get source app
	sourceApp, err := getAppByID(sqlDB, sourceAppID)
	if err != nil {
		return nil, ErrNotFound
	}

	// Generate new ID
	newID := appid.GenerateApp()

	// Build tags JSON
	tagsJSON := "[]"
	if len(sourceApp.Tags) > 0 {
		b, _ := json.Marshal(sourceApp.Tags)
		tagsJSON = string(b)
	}

	// Determine original_id
	originalID := sourceApp.OriginalID
	if originalID == "" {
		originalID = sourceApp.ID
	}

	// Insert forked app
	query := `
		INSERT INTO apps (id, original_id, forked_from_id, title, description, tags, visibility, source, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, 'unlisted', 'fork', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	_, err = sqlDB.Exec(query, newID, originalID, sourceApp.ID, sourceApp.Title, sourceApp.Description, tagsJSON)
	if err != nil {
		return nil, err
	}

	// Copy files
	copyQuery := `
		INSERT INTO files (site_id, app_id, path, content, size_bytes, mime_type, hash, created_at, updated_at)
		SELECT site_id, ?, path, content, size_bytes, mime_type, hash, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		FROM files WHERE app_id = ?
	`
	sqlDB.Exec(copyQuery, newID, sourceApp.ID)

	// Copy KV storage
	if copyStorage {
		kvQuery := `
			INSERT INTO kv_store (site_id, key, value, created_at, updated_at)
			SELECT ?, key, value, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
			FROM kv_store WHERE site_id = ?
		`
		sqlDB.Exec(kvQuery, newID, sourceApp.ID)
	}

	// Create alias if specified
	if newAlias != "" {
		aliasTargets := `{"app_id":"` + newID + `"}`
		aliasQuery := `
			INSERT INTO aliases (subdomain, type, targets, created_at, updated_at)
			VALUES (?, 'proxy', ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`
		sqlDB.Exec(aliasQuery, newAlias, aliasTargets)
	}

	cfg := config.Get()
	result := map[string]interface{}{
		"id":             newID,
		"title":          sourceApp.Title,
		"forked_from_id": sourceApp.ID,
		"original_id":    originalID,
		"message":        "App forked",
	}

	if newAlias != "" {
		result["alias"] = newAlias
		result["url"] = "https://" + newAlias + "." + cfg.Server.Domain
	}

	return result, nil
}

func cmdAppLineage(db interface{}, args []string) (interface{}, error) {
	if len(args) < 1 {
		return nil, ErrMissingArgument
	}

	sqlDB := database.GetDB()
	identifier := args[0]

	// Parse flags
	for i, arg := range args {
		if (arg == "--alias" || arg == "--id") && i+1 < len(args) {
			identifier = args[i+1]
		}
	}

	// Resolve app
	var appID string
	if appid.IsValid(identifier) {
		appID = identifier
	} else {
		resolvedID, _, err := ResolveAlias(identifier)
		if err != nil || resolvedID == "" {
			return nil, ErrNotFound
		}
		appID = resolvedID
	}

	// Get original_id
	var originalID string
	err := sqlDB.QueryRow("SELECT COALESCE(original_id, id) FROM apps WHERE id = ?", appID).Scan(&originalID)
	if err != nil {
		return nil, ErrNotFound
	}

	// Build tree
	tree := buildLineageTree(sqlDB, originalID, nil)
	return tree, nil
}

func cmdServerInfo() (interface{}, error) {
	cfg := config.Get()
	stats := hosting.GetStats()

	return map[string]interface{}{
		"version": config.Version,
		"domain":  cfg.Server.Domain,
		"env":     cfg.Server.Env,
		"vfs": map[string]interface{}{
			"cached_files": stats.CachedFiles,
			"cache_size":   stats.CacheSizeBytes,
		},
	}, nil
}

// Error types
type cmdError string

func (e cmdError) Error() string { return string(e) }

const (
	ErrDatabaseNotInitialized cmdError = "database not initialized"
	ErrUnknownCommand         cmdError = "unknown command"
	ErrUnknownSubcommand      cmdError = "unknown subcommand"
	ErrMissingSubcommand      cmdError = "missing subcommand"
	ErrMissingArgument        cmdError = "missing required argument"
	ErrNotFound               cmdError = "not found"
	ErrReservedSubdomain      cmdError = "subdomain is reserved"
)

// parseFlags extracts flag values from args
func parseFlags(args []string) map[string]string {
	flags := make(map[string]string)
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "--") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
			key := strings.TrimPrefix(args[i], "--")
			flags[key] = args[i+1]
			i++
		}
	}
	return flags
}
