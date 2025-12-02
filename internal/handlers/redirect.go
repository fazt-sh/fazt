package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/fazt-sh/fazt/internal/database"
)

// RedirectHandler handles redirect tracking
func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	// Extract slug from URL path (/r/{slug})
	path := strings.TrimPrefix(r.URL.Path, "/r/")
	slug := strings.TrimSpace(path)

	if slug == "" {
		http.Error(w, "Invalid redirect slug", http.StatusBadRequest)
		return
	}

	// Lookup redirect in database
	db := database.GetDB()
	var destination string
	var tags string
	var id int64

	err := db.QueryRow(`
		SELECT id, destination, tags FROM redirects WHERE slug = ?
	`, slug).Scan(&id, &destination, &tags)

	if err == sql.ErrNoRows {
		http.Error(w, "Redirect not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Error looking up redirect: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Parse additional tags from query string
	query := r.URL.Query()
	extraTags := query.Get("tags")
	if extraTags != "" {
		if tags != "" {
			tags = tags + "," + extraTags
		} else {
			tags = extraTags
		}
	}

	// Extract client info
	ipAddress := extractIPAddress(r)
	userAgent := r.UserAgent()
	referrer := r.Referer()

	// Log the click event
	_, err = db.Exec(`
		INSERT INTO events (domain, tags, source_type, event_type, path, referrer, user_agent, ip_address)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, slug, tags, "redirect", "click", "/r/"+slug, referrer, userAgent, ipAddress)

	if err != nil {
		log.Printf("Error logging redirect event: %v", err)
		// Don't fail the redirect - continue
	}

	// Increment click count
	_, err = db.Exec(`
		UPDATE redirects SET click_count = click_count + 1 WHERE id = ?
	`, id)

	if err != nil {
		log.Printf("Error incrementing click count: %v", err)
		// Don't fail the redirect - continue
	}

	// Perform redirect
	http.Redirect(w, r, destination, http.StatusFound)
}
