package handlers

import (
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/fazt-sh/fazt/internal/analytics"
	"github.com/fazt-sh/fazt/internal/api"
	"github.com/fazt-sh/fazt/internal/models"
)

const maxBodySize = 10 * 1024 // 10KB

// TrackHandler handles tracking requests
func TrackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.BadRequest(w, "Method not allowed")
		return
	}

	// Limit body size
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

	// Parse JSON body
	var req models.TrackRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		api.InvalidJSON(w, "Invalid JSON")
		return
	}

	// Determine domain (priority: explicit domain > hostname > referer hostname > "unknown")
	domain := determineDomain(&req, r)

	// Validate required fields
	if domain == "" {
		domain = "unknown"
	}
	if req.EventType == "" {
		req.EventType = "pageview"
	}

	// Extract client information
	ipAddress := extractIPAddress(r)
	userAgent := r.UserAgent()
	referrer := req.Referrer
	if referrer == "" {
		referrer = r.Referer()
	}

	// Prepare tags as comma-separated string
	tagsStr := ""
	if len(req.Tags) > 0 {
		tagsStr = strings.Join(req.Tags, ",")
	}

	// Sanitize inputs
	domain = sanitizeInput(domain)
	req.Path = sanitizeInput(req.Path)
	referrer = sanitizeInput(referrer)

	// Convert query params to JSON string
	queryParamsJSON := req.ToQueryParamsJSON()

	// Add to analytics buffer
	analytics.Add(analytics.Event{
		Domain:      domain,
		Tags:        tagsStr,
		SourceType:  "web",
		EventType:   req.EventType,
		Path:        req.Path,
		Referrer:    referrer,
		UserAgent:   userAgent,
		IPAddress:   ipAddress,
		QueryParams: queryParamsJSON,
	})

	// Return 204 No Content on success
	w.WriteHeader(http.StatusNoContent)
}

// determineDomain extracts domain from request in order of priority
func determineDomain(req *models.TrackRequest, r *http.Request) string {
	// 1. Explicit domain parameter
	if req.Domain != "" {
		return req.Domain
	}

	// 2. Hostname from request
	if req.Hostname != "" {
		return req.Hostname
	}

	// 3. Extract from Referer header
	referer := r.Referer()
	if referer != "" {
		if parsedURL, err := url.Parse(referer); err == nil {
			if parsedURL.Host != "" {
				return parsedURL.Host
			}
		}
	}

	// 4. Default to unknown
	return "unknown"
}

// extractIPAddress gets the client's IP address from the request
func extractIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies/load balancers)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP in the list
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

// sanitizeInput removes potentially dangerous characters and limits length
func sanitizeInput(input string) string {
	// Trim whitespace
	input = strings.TrimSpace(input)

	// Limit length to prevent abuse
	maxLen := 500
	if len(input) > maxLen {
		input = input[:maxLen]
	}

	return input
}
