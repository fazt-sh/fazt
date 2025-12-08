package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/fazt-sh/fazt/internal/api"
	"github.com/fazt-sh/fazt/internal/database"
)

// DeleteWebhookHandler handles DELETE /api/webhooks/{id}
func DeleteWebhookHandler(w http.ResponseWriter, r *http.Request) {
	// Get ID from path
	idStr := r.PathValue("id")
	if idStr == "" {
		api.BadRequest(w, "Webhook ID required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		api.BadRequest(w, "Invalid webhook ID")
		return
	}

	// Delete from database
	db := database.GetDB()
	result, err := db.Exec("DELETE FROM webhooks WHERE id = ?", id)
	if err != nil {
		api.ServerError(w, err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		api.NotFound(w, "WEBHOOK_NOT_FOUND", "Webhook not found")
		return
	}

	api.Success(w, http.StatusOK, map[string]string{"message": "Webhook deleted"})
}

// UpdateWebhookHandler handles PUT /api/webhooks/{id}
func UpdateWebhookHandler(w http.ResponseWriter, r *http.Request) {
	// Get ID from path
	idStr := r.PathValue("id")
	if idStr == "" {
		api.BadRequest(w, "Webhook ID required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		api.BadRequest(w, "Invalid webhook ID")
		return
	}

	// Parse request body
	var req struct {
		Name     *string `json:"name,omitempty"`
		Endpoint *string `json:"endpoint,omitempty"`
		Secret   *string `json:"secret,omitempty"`
		IsActive *bool   `json:"is_active,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "Invalid JSON")
		return
	}

	// Build update query dynamically based on provided fields
	db := database.GetDB()

	// First check if webhook exists
	var exists int
	err = db.QueryRow("SELECT COUNT(*) FROM webhooks WHERE id = ?", id).Scan(&exists)
	if err != nil || exists == 0 {
		api.NotFound(w, "WEBHOOK_NOT_FOUND", "Webhook not found")
		return
	}

	// Update fields
	if req.Name != nil {
		_, err = db.Exec("UPDATE webhooks SET name = ? WHERE id = ?", *req.Name, id)
		if err != nil {
			api.ServerError(w, err)
			return
		}
	}

	if req.Endpoint != nil {
		// Check if new endpoint conflicts with another webhook
		var conflictCount int
		db.QueryRow("SELECT COUNT(*) FROM webhooks WHERE endpoint = ? AND id != ?", *req.Endpoint, id).Scan(&conflictCount)
		if conflictCount > 0 {
			api.ErrorResponse(w, http.StatusConflict, "CONFLICT", "Endpoint already in use", "endpoint")
			return
		}

		_, err = db.Exec("UPDATE webhooks SET endpoint = ? WHERE id = ?", *req.Endpoint, id)
		if err != nil {
			api.ServerError(w, err)
			return
		}
	}

	if req.Secret != nil {
		_, err = db.Exec("UPDATE webhooks SET secret = ? WHERE id = ?", *req.Secret, id)
		if err != nil {
			api.ServerError(w, err)
			return
		}
	}

	if req.IsActive != nil {
		_, err = db.Exec("UPDATE webhooks SET is_active = ? WHERE id = ?", *req.IsActive, id)
		if err != nil {
			api.ServerError(w, err)
			return
		}
	}

	// Fetch updated webhook
	var name, endpoint, secret string
	var isActive bool

	err = db.QueryRow(`
		SELECT name, endpoint, secret, is_active
		FROM webhooks
		WHERE id = ?
	`, id).Scan(&name, &endpoint, &secret, &isActive)

	if err != nil {
		api.ServerError(w, err)
		return
	}

	result := map[string]interface{}{
		"id":         id,
		"name":       name,
		"endpoint":   endpoint,
		"has_secret": secret != "",
		"is_active":  isActive,
	}

	api.Success(w, http.StatusOK, result)
}
