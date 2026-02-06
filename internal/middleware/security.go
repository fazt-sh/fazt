package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/fazt-sh/fazt/internal/config"
	"github.com/fazt-sh/fazt/internal/system"
)

// MaxBodySize is the default maximum request body size (1MB)
const MaxBodySize = 1 << 20 // 1MB

// BodySizeLimit limits the size of request bodies to prevent memory exhaustion
func BodySizeLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip for paths that have their own limits (deploy has 100MB)
			if r.URL.Path == "/api/deploy" {
				next.ServeHTTP(w, r)
				return
			}

			// Allow larger bodies for multipart file uploads
			contentType := r.Header.Get("Content-Type")
			if strings.Contains(contentType, "multipart/form-data") {
				maxUpload := system.GetLimits().Storage.MaxUpload
				r.Body = http.MaxBytesReader(w, r.Body, maxUpload)
				next.ServeHTTP(w, r)
				return
			}

			// Limit request body size
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// RequestTracing adds a unique request ID header for tracing
func RequestTracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request already has an ID (from load balancer)
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Set on response for client visibility
		w.Header().Set("X-Request-ID", requestID)

		// Add to request context (can be used in handlers)
		r.Header.Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r)
	})
}

// generateRequestID creates a short random hex string
func generateRequestID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Comprehensive CDN whitelist for CSP
// Organized by category for easy maintenance
var (
	// JavaScript CDNs and ES module hosts
	scriptCDNs = []string{
		"https://cdn.jsdelivr.net",
		"https://unpkg.com",
		"https://cdnjs.cloudflare.com",
		"https://cdn.tailwindcss.com",
		"https://esm.sh",
		"https://esm.run",
		"https://cdn.skypack.dev",
		"https://ga.jspm.io",
		"https://ajax.googleapis.com",
		"https://ajax.aspnetcdn.com",
		"https://code.jquery.com",
		"https://stackpath.bootstrapcdn.com",
		"https://cdn.bootcdn.net",
		"https://lib.baomitu.com",
		"https://polyfill.io",
		"https://kit.fontawesome.com",
		"https://cdn.statically.io",
		"https://rawcdn.githack.com",
		"https://raw.githubusercontent.com",
		"https://static.cloudflareinsights.com", // Cloudflare Web Analytics
	}

	// Style/CSS CDNs
	styleCDNs = []string{
		"https://cdn.jsdelivr.net",
		"https://unpkg.com",
		"https://cdnjs.cloudflare.com",
		"https://fonts.googleapis.com",
		"https://fonts.bunny.net",
		"https://cdn.fontshare.com",
		"https://use.typekit.net",
		"https://stackpath.bootstrapcdn.com",
		"https://cdn.statically.io",
	}

	// Font hosts
	fontCDNs = []string{
		"https://cdn.jsdelivr.net",
		"https://fonts.gstatic.com",
		"https://fonts.bunny.net",
		"https://cdn.fontshare.com",
		"https://use.typekit.net",
		"https://kit.fontawesome.com",
	}

	// Connect sources (fetch/XHR/WebSocket)
	connectCDNs = []string{
		"https://cdn.jsdelivr.net",
		"https://unpkg.com",
		"https://esm.sh",
		"https://esm.run",
		"https://cdn.skypack.dev",
		"https://ga.jspm.io",
		"https://api.github.com",
	}
)

// buildCSP constructs the Content-Security-Policy header
func buildCSP(domain string, port string, isSecure bool) string {
	// Allow both http and https for cross-subdomain requests
	connectDomain := "https://*." + domain
	if !isSecure {
		// In development with non-standard ports, include port in CSP
		if port != "80" && port != "443" && port != "" {
			connectDomain += " http://*." + domain + ":" + port
		} else {
			connectDomain += " http://*." + domain
		}
	}

	return "default-src 'self'; " +
		"script-src 'self' 'unsafe-inline' 'unsafe-eval' " + joinSources(scriptCDNs) + "; " +
		"style-src 'self' 'unsafe-inline' " + joinSources(styleCDNs) + "; " +
		"img-src 'self' data: blob: https:; " +
		"font-src 'self' data: " + joinSources(fontCDNs) + "; " +
		"connect-src 'self' " + joinSources(connectCDNs) + " " + connectDomain + "; " +
		"media-src 'self' blob: https:; " +
		"object-src 'none'; " +
		"frame-ancestors *"
}

// joinSources joins CDN sources with spaces
func joinSources(sources []string) string {
	result := ""
	for i, src := range sources {
		if i > 0 {
			result += " "
		}
		result += src
	}
	return result
}

// SecurityHeaders adds security-related HTTP headers
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg := config.Get()

		// Basic security headers
		// Note: X-Frame-Options removed to allow iframe embedding (CSP frame-ancestors handles this)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Content Security Policy with comprehensive CDN whitelist
		// In development, allow both HTTP and HTTPS for cross-subdomain requests
		isSecure := cfg.IsProduction() || cfg.HTTPS.Enabled
		w.Header().Set("Content-Security-Policy", buildCSP(cfg.Server.Domain, cfg.Server.Port, isSecure))

		// HSTS in production
		if cfg.IsProduction() {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Permissions Policy
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		next.ServeHTTP(w, r)
	})
}
