package handlers

import (
	"encoding/base64"
	"log"
	"net/http"
	"strings"

	"github.com/jikku/command-center/internal/database"
)

// 1x1 transparent GIF pixel (base64 encoded)
const transparentGIF = "R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7"

// PixelHandler serves a 1x1 transparent GIF for tracking
func PixelHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	domain := query.Get("domain")
	tagsParam := query.Get("tags")
	source := query.Get("source")

	// Default domain if not provided
	if domain == "" {
		// Try to extract from referer
		if referer := r.Referer(); referer != "" {
			domain = extractDomainFromReferer(referer)
		}
		if domain == "" {
			domain = "unknown"
		}
	}

	// Parse tags
	tagsStr := ""
	if tagsParam != "" {
		tagsStr = tagsParam
	}

	// Extract client info
	ipAddress := extractIPAddress(r)
	userAgent := r.UserAgent()
	referrer := r.Referer()

	// Default source
	if source == "" {
		source = "pixel"
	}

	// Log event to database
	db := database.GetDB()
	_, err := db.Exec(`
		INSERT INTO events (domain, tags, source_type, event_type, path, referrer, user_agent, ip_address)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, domain, tagsStr, "pixel", source, "", referrer, userAgent, ipAddress)

	if err != nil {
		log.Printf("Error logging pixel event: %v", err)
		// Don't fail - still return pixel
	}

	// Decode base64 GIF
	gifBytes, err := base64.StdEncoding.DecodeString(transparentGIF)
	if err != nil {
		log.Printf("Error decoding GIF: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set headers for GIF response
	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Write GIF
	w.WriteHeader(http.StatusOK)
	w.Write(gifBytes)
}

// extractDomainFromReferer extracts the hostname from a referer URL
func extractDomainFromReferer(referer string) string {
	// Remove protocol
	referer = strings.TrimPrefix(referer, "http://")
	referer = strings.TrimPrefix(referer, "https://")

	// Extract domain (everything before first slash)
	parts := strings.Split(referer, "/")
	if len(parts) > 0 {
		return parts[0]
	}

	return ""
}
