package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/fazt-sh/fazt/internal/api"
	"github.com/fazt-sh/fazt/internal/database"
)

// Alias represents a routing alias
type Alias struct {
	Subdomain string          `json:"subdomain"`
	Type      string          `json:"type"`
	Targets   json.RawMessage `json:"targets,omitempty"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}

// AliasTarget represents a proxy target
type AliasTarget struct {
	AppID string `json:"app_id"`
}

// SplitTarget represents a traffic split target
type SplitTarget struct {
	AppID  string `json:"app_id"`
	Weight int    `json:"weight"`
}

// RedirectTarget represents a redirect target
type RedirectTarget struct {
	URL string `json:"url"`
}

// AliasesListHandler returns the list of all aliases
func AliasesListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	query := `
		SELECT subdomain, type, targets, created_at, updated_at
		FROM aliases
		ORDER BY subdomain
	`

	rows, err := db.Query(query)
	if err != nil {
		api.InternalError(w, err)
		return
	}
	defer rows.Close()

	var aliases []Alias
	for rows.Next() {
		var a Alias
		var targets *string
		var createdAt, updatedAt interface{}

		err := rows.Scan(&a.Subdomain, &a.Type, &targets, &createdAt, &updatedAt)
		if err != nil {
			continue
		}

		if targets != nil && *targets != "" {
			a.Targets = json.RawMessage(*targets)
		}

		if createdAt != nil {
			a.CreatedAt = formatTime(createdAt)
		}
		if updatedAt != nil {
			a.UpdatedAt = formatTime(updatedAt)
		}

		aliases = append(aliases, a)
	}

	api.Success(w, http.StatusOK, aliases)
}

// AliasDetailHandler returns details for a single alias
func AliasDetailHandler(w http.ResponseWriter, r *http.Request) {
	subdomain := r.PathValue("subdomain")
	if subdomain == "" {
		api.BadRequest(w, "subdomain required")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	query := `
		SELECT subdomain, type, targets, created_at, updated_at
		FROM aliases WHERE subdomain = ?
	`

	var a Alias
	var targets *string
	var createdAt, updatedAt interface{}

	err := db.QueryRow(query, subdomain).Scan(&a.Subdomain, &a.Type, &targets, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		api.NotFound(w, "ALIAS_NOT_FOUND", "Alias not found")
		return
	}
	if err != nil {
		api.InternalError(w, err)
		return
	}

	if targets != nil && *targets != "" {
		a.Targets = json.RawMessage(*targets)
	}

	if createdAt != nil {
		a.CreatedAt = formatTime(createdAt)
	}
	if updatedAt != nil {
		a.UpdatedAt = formatTime(updatedAt)
	}

	api.Success(w, http.StatusOK, a)
}

// AliasCreateRequest is the request body for creating an alias
type AliasCreateRequest struct {
	Subdomain string `json:"subdomain"`
	Type      string `json:"type"`
	AppID     string `json:"app_id,omitempty"`
	URL       string `json:"url,omitempty"`
}

// AliasCreateHandler creates a new alias (link)
func AliasCreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	var req AliasCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}

	if req.Subdomain == "" {
		api.BadRequest(w, "subdomain is required")
		return
	}

	// Validate subdomain format
	if !isValidSubdomain(req.Subdomain) {
		api.BadRequest(w, "invalid subdomain format")
		return
	}

	// Default type is proxy
	if req.Type == "" {
		req.Type = "proxy"
	}

	// Validate type
	if req.Type != "proxy" && req.Type != "redirect" && req.Type != "reserved" {
		api.BadRequest(w, "type must be 'proxy', 'redirect', or 'reserved'")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	// Build targets JSON
	var targets *string
	switch req.Type {
	case "proxy":
		if req.AppID == "" {
			api.BadRequest(w, "app_id is required for proxy aliases")
			return
		}
		// Verify app exists
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM apps WHERE id = ?", req.AppID).Scan(&count)
		if err != nil || count == 0 {
			api.BadRequest(w, "app_id not found")
			return
		}
		t := `{"app_id":"` + req.AppID + `"}`
		targets = &t
	case "redirect":
		if req.URL == "" {
			api.BadRequest(w, "url is required for redirect aliases")
			return
		}
		t := `{"url":"` + req.URL + `"}`
		targets = &t
	case "reserved":
		// No targets needed
	}

	// Insert alias
	query := `
		INSERT INTO aliases (subdomain, type, targets, created_at, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(subdomain) DO UPDATE SET
			type = excluded.type,
			targets = excluded.targets,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := db.Exec(query, req.Subdomain, req.Type, targets)
	if err != nil {
		api.InternalError(w, err)
		return
	}

	api.Success(w, http.StatusCreated, map[string]interface{}{
		"subdomain": req.Subdomain,
		"type":      req.Type,
		"message":   "Alias created",
	})
}

// AliasUpdateHandler updates an existing alias
func AliasUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	subdomain := r.PathValue("subdomain")
	if subdomain == "" {
		api.BadRequest(w, "subdomain required")
		return
	}

	var req AliasCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	// Check alias exists
	var currentType string
	err := db.QueryRow("SELECT type FROM aliases WHERE subdomain = ?", subdomain).Scan(&currentType)
	if err == sql.ErrNoRows {
		api.NotFound(w, "ALIAS_NOT_FOUND", "Alias not found")
		return
	}
	if err != nil {
		api.InternalError(w, err)
		return
	}

	// Use current type if not specified
	if req.Type == "" {
		req.Type = currentType
	}

	// Build targets JSON
	var targets *string
	switch req.Type {
	case "proxy":
		if req.AppID == "" {
			api.BadRequest(w, "app_id is required for proxy aliases")
			return
		}
		// Verify app exists
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM apps WHERE id = ?", req.AppID).Scan(&count)
		if err != nil || count == 0 {
			api.BadRequest(w, "app_id not found")
			return
		}
		t := `{"app_id":"` + req.AppID + `"}`
		targets = &t
	case "redirect":
		if req.URL == "" {
			api.BadRequest(w, "url is required for redirect aliases")
			return
		}
		t := `{"url":"` + req.URL + `"}`
		targets = &t
	case "reserved":
		// No targets needed
	}

	// Update alias
	query := `UPDATE aliases SET type = ?, targets = ?, updated_at = CURRENT_TIMESTAMP WHERE subdomain = ?`
	_, err = db.Exec(query, req.Type, targets, subdomain)
	if err != nil {
		api.InternalError(w, err)
		return
	}

	api.Success(w, http.StatusOK, map[string]interface{}{
		"subdomain": subdomain,
		"type":      req.Type,
		"message":   "Alias updated",
	})
}

// AliasDeleteHandler deletes an alias
func AliasDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	subdomain := r.PathValue("subdomain")
	if subdomain == "" {
		api.BadRequest(w, "subdomain required")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	// Check alias exists
	var aliasType string
	err := db.QueryRow("SELECT type FROM aliases WHERE subdomain = ?", subdomain).Scan(&aliasType)
	if err == sql.ErrNoRows {
		api.NotFound(w, "ALIAS_NOT_FOUND", "Alias not found")
		return
	}
	if err != nil {
		api.InternalError(w, err)
		return
	}

	// Don't allow deleting system reserved aliases
	if aliasType == "reserved" && (subdomain == "admin" || subdomain == "api") {
		api.ErrorResponse(w, http.StatusForbidden, "SYSTEM_ALIAS", "Cannot delete system alias", "")
		return
	}

	_, err = db.Exec("DELETE FROM aliases WHERE subdomain = ?", subdomain)
	if err != nil {
		api.InternalError(w, err)
		return
	}

	api.Success(w, http.StatusOK, map[string]interface{}{
		"subdomain": subdomain,
		"message":   "Alias deleted",
	})
}

// AliasReserveHandler reserves a subdomain
func AliasReserveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	subdomain := r.PathValue("subdomain")
	if subdomain == "" {
		api.BadRequest(w, "subdomain required")
		return
	}

	if !isValidSubdomain(subdomain) {
		api.BadRequest(w, "invalid subdomain format")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	query := `
		INSERT INTO aliases (subdomain, type, targets, created_at, updated_at)
		VALUES (?, 'reserved', NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(subdomain) DO UPDATE SET
			type = 'reserved',
			targets = NULL,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := db.Exec(query, subdomain)
	if err != nil {
		api.InternalError(w, err)
		return
	}

	api.Success(w, http.StatusCreated, map[string]interface{}{
		"subdomain": subdomain,
		"type":      "reserved",
		"message":   "Subdomain reserved",
	})
}

// SwapRequest is the request body for swapping aliases
type SwapRequest struct {
	Alias1 string `json:"alias1"`
	Alias2 string `json:"alias2"`
}

// AliasSwapHandler atomically swaps two aliases' targets
func AliasSwapHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	var req SwapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}

	if req.Alias1 == "" || req.Alias2 == "" {
		api.BadRequest(w, "alias1 and alias2 are required")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		api.InternalError(w, err)
		return
	}
	defer tx.Rollback()

	// Get both aliases
	var type1, type2, targets1, targets2 string
	err = tx.QueryRow("SELECT type, COALESCE(targets, '') FROM aliases WHERE subdomain = ?", req.Alias1).Scan(&type1, &targets1)
	if err == sql.ErrNoRows {
		api.BadRequest(w, "alias1 not found")
		return
	}
	if err != nil {
		api.InternalError(w, err)
		return
	}

	err = tx.QueryRow("SELECT type, COALESCE(targets, '') FROM aliases WHERE subdomain = ?", req.Alias2).Scan(&type2, &targets2)
	if err == sql.ErrNoRows {
		api.BadRequest(w, "alias2 not found")
		return
	}
	if err != nil {
		api.InternalError(w, err)
		return
	}

	// Both must be proxy type
	if type1 != "proxy" || type2 != "proxy" {
		api.BadRequest(w, "both aliases must be proxy type to swap")
		return
	}

	// Swap targets
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	_, err = tx.Exec("UPDATE aliases SET targets = ?, updated_at = ? WHERE subdomain = ?", targets2, now, req.Alias1)
	if err != nil {
		api.InternalError(w, err)
		return
	}
	_, err = tx.Exec("UPDATE aliases SET targets = ?, updated_at = ? WHERE subdomain = ?", targets1, now, req.Alias2)
	if err != nil {
		api.InternalError(w, err)
		return
	}

	if err := tx.Commit(); err != nil {
		api.InternalError(w, err)
		return
	}

	api.Success(w, http.StatusOK, map[string]interface{}{
		"alias1":  req.Alias1,
		"alias2":  req.Alias2,
		"message": "Aliases swapped",
	})
}

// SplitRequest is the request body for traffic splitting
type SplitRequest struct {
	Targets []SplitTarget `json:"targets"`
}

// AliasSplitHandler configures traffic splitting
func AliasSplitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	subdomain := r.PathValue("subdomain")
	if subdomain == "" {
		api.BadRequest(w, "subdomain required")
		return
	}

	var req SplitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}

	if len(req.Targets) < 2 {
		api.BadRequest(w, "at least 2 targets required for split")
		return
	}

	// Validate weights sum to 100
	totalWeight := 0
	for _, t := range req.Targets {
		if t.Weight < 1 || t.Weight > 100 {
			api.BadRequest(w, "weights must be between 1 and 100")
			return
		}
		totalWeight += t.Weight
	}
	if totalWeight != 100 {
		api.BadRequest(w, "weights must sum to 100")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	// Verify all app IDs exist
	for _, t := range req.Targets {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM apps WHERE id = ?", t.AppID).Scan(&count)
		if err != nil || count == 0 {
			api.BadRequest(w, "app_id not found: "+t.AppID)
			return
		}
	}

	// Build targets JSON
	targetsJSON, err := json.Marshal(req.Targets)
	if err != nil {
		api.InternalError(w, err)
		return
	}

	query := `
		INSERT INTO aliases (subdomain, type, targets, created_at, updated_at)
		VALUES (?, 'split', ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(subdomain) DO UPDATE SET
			type = 'split',
			targets = excluded.targets,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err = db.Exec(query, subdomain, string(targetsJSON))
	if err != nil {
		api.InternalError(w, err)
		return
	}

	api.Success(w, http.StatusCreated, map[string]interface{}{
		"subdomain": subdomain,
		"type":      "split",
		"targets":   req.Targets,
		"message":   "Traffic split configured",
	})
}

// ResolveAlias resolves a subdomain to an app ID
func ResolveAlias(subdomain string) (appID string, aliasType string, err error) {
	db := database.GetDB()
	if db == nil {
		return "", "", sql.ErrConnDone
	}

	var targets *string
	err = db.QueryRow("SELECT type, targets FROM aliases WHERE subdomain = ?", subdomain).Scan(&aliasType, &targets)
	if err == sql.ErrNoRows {
		return "", "", nil // No alias found
	}
	if err != nil {
		return "", "", err
	}

	switch aliasType {
	case "proxy":
		if targets != nil {
			var t AliasTarget
			if err := json.Unmarshal([]byte(*targets), &t); err == nil {
				return t.AppID, aliasType, nil
			}
		}
	case "split":
		if targets != nil {
			var splits []SplitTarget
			if err := json.Unmarshal([]byte(*targets), &splits); err == nil && len(splits) > 0 {
				// TODO: Implement weighted random selection with sticky sessions
				// For now, just return the first target
				return splits[0].AppID, aliasType, nil
			}
		}
	case "reserved":
		return "", "reserved", nil
	case "redirect":
		// Return empty app ID but indicate redirect type
		return "", "redirect", nil
	}

	return "", "", nil
}

// GetRedirectURL gets the redirect URL for a redirect alias
func GetRedirectURL(subdomain string) (string, error) {
	db := database.GetDB()
	if db == nil {
		return "", sql.ErrConnDone
	}

	var targets *string
	var aliasType string
	err := db.QueryRow("SELECT type, targets FROM aliases WHERE subdomain = ?", subdomain).Scan(&aliasType, &targets)
	if err != nil {
		return "", err
	}

	if aliasType != "redirect" || targets == nil {
		return "", nil
	}

	var t RedirectTarget
	if err := json.Unmarshal([]byte(*targets), &t); err != nil {
		return "", err
	}

	return t.URL, nil
}

func isValidSubdomain(s string) bool {
	if len(s) < 1 || len(s) > 63 {
		return false
	}
	// Strict validation - no normalization, must be lowercase
	if s[0] == '-' || s[len(s)-1] == '-' {
		return false
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	return true
}
