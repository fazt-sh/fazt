package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/fazt-sh/fazt/internal/analytics"
	"github.com/fazt-sh/fazt/internal/api"
	"github.com/fazt-sh/fazt/internal/database"
)

// WebhookHandler handles incoming webhooks
func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.BadRequest(w, "Method not allowed")
		return
	}

	// Extract endpoint from URL path (/webhook/{endpoint})
	path := strings.TrimPrefix(r.URL.Path, "/webhook/")
	endpoint := strings.TrimSpace(path)

	if endpoint == "" {
		api.BadRequest(w, "Invalid webhook endpoint")
		return
	}

	// Lookup webhook configuration
	db := database.GetDB()
	var webhookID int64
	var secret string
	var isActive bool
	var name string

	err := db.QueryRow(`
		SELECT id, name, secret, is_active FROM webhooks WHERE endpoint = ?
	`, endpoint).Scan(&webhookID, &name, &secret, &isActive)

	if err == sql.ErrNoRows {
		api.NotFound(w, "WEBHOOK_NOT_FOUND", "Webhook endpoint not found")
		return
	} else if err != nil {
		log.Printf("Error looking up webhook: %v", err)
		api.InternalError(w, err)
		return
	}

	// Check if webhook is active
	if !isActive {
		api.Unauthorized(w, "Webhook is disabled")
		return
	}

	// Read body
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodySize))
	if err != nil {
		api.BadRequest(w, "Failed to read body")
		return
	}

	// Verify signature if secret is configured
	if secret != "" {
		signature := r.Header.Get("X-Webhook-Signature")
		if signature == "" {
			api.Unauthorized(w, "Missing signature")
			return
		}

		if !verifySignature(body, secret, signature) {
			api.Unauthorized(w, "Invalid signature")
			return
		}
	}

	// Parse JSON payload (flexible structure)
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		// If not JSON, store raw body
		payload = map[string]interface{}{
			"raw": string(body),
		}
	}

	// Extract useful fields if present
	eventType := "webhook"
	if et, ok := payload["event"].(string); ok {
		eventType = et
	} else if et, ok := payload["type"].(string); ok {
		eventType = et
	}

	// Convert payload back to JSON string for storage
	payloadJSON, _ := json.Marshal(payload)

	// Extract client info
	ipAddress := extractIPAddress(r)
	userAgent := r.UserAgent()

	// Add to analytics buffer
	analytics.Add(analytics.Event{
		Domain:      endpoint,
		Tags:        "",
		SourceType:  "webhook",
		EventType:   eventType,
		Path:        "/webhook/" + endpoint,
		UserAgent:   userAgent,
		IPAddress:   ipAddress,
		QueryParams: string(payloadJSON),
	})

	// Return success response
	api.Success(w, http.StatusOK, map[string]interface{}{
		"message": "Webhook received",
		"webhook": name,
	})
}

// verifySignature verifies HMAC SHA256 signature
func verifySignature(body []byte, secret, signature string) bool {
	// Compute HMAC SHA256
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures (constant time comparison)
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}
