package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/fazt-sh/fazt/internal/api"
	"github.com/fazt-sh/fazt/internal/auth"
	"github.com/fazt-sh/fazt/internal/config"
	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/hosting"
)

// DeployHandler handles site deployments via ZIP upload
// POST /api/deploy
// - Multipart form with "file" (ZIP) and "site_name" field
// - Authorization: Bearer <token> header required
func DeployHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.BadRequest(w, "Method not allowed")
		return
	}

	// Rate limit: 5 deploys per minute per IP
	clientIP := r.RemoteAddr
	if fwdIP := r.Header.Get("X-Forwarded-For"); fwdIP != "" {
		clientIP = strings.Split(fwdIP, ",")[0]
	}
	limiter := auth.GetDeployLimiter()
	if !limiter.AllowDeploy(clientIP) {
		api.RateLimitExceeded(w, "Rate limit exceeded: max 5 deploys per minute")
		return
	}

	// Validate API key
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		api.Unauthorized(w, "Missing Authorization header")
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		api.Unauthorized(w, "Invalid Authorization format, use: Bearer <token>")
		return
	}

	db := database.GetDB()
	keyID, keyName, err := hosting.ValidateAPIKey(db, token)
	if err != nil {
		api.InvalidAPIKey(w)
		return
	}

	// Parse multipart form (max 100MB)
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		api.BadRequest(w, "Failed to parse form: "+err.Error())
		return
	}

	// Get site name
	siteName := r.FormValue("site_name")
	if siteName == "" {
		api.BadRequest(w, "Missing site_name field")
		return
	}

	// Smart Domain Handling: Strip root domain if present
	// Allows "my-site.fazt.sh" -> "my-site"
	cfg := config.Get()
	rootDomain := cfg.Server.Domain
	// Strip scheme if present
	if idx := strings.Index(rootDomain, "://"); idx != -1 {
		rootDomain = rootDomain[idx+3:]
	}
	// Strip suffix
	suffix := "." + rootDomain
	if strings.HasSuffix(strings.ToLower(siteName), suffix) {
		siteName = siteName[:len(siteName)-len(suffix)]
	}

	// Validate site name
	if err := hosting.ValidateSubdomain(siteName); err != nil {
		api.BadRequest(w, "Invalid site_name: "+err.Error())
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		api.BadRequest(w, "Missing or invalid file")
		return
	}
	defer file.Close()

	// Verify it's a ZIP file
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".zip") {
		api.BadRequest(w, "File must be a ZIP archive")
		return
	}

	// Read file into memory (we need to seek for zip.Reader)
	var buf bytes.Buffer
	size, err := io.Copy(&buf, file)
	if err != nil {
		api.InternalError(w, err)
		return
	}

	// Create zip reader
	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), size)
	if err != nil {
		api.BadRequest(w, "Invalid ZIP file: "+err.Error())
		return
	}

	// Deploy the site
	result, err := hosting.DeploySite(zipReader, siteName)
	if err != nil {
		api.InternalError(w, err)
		return
	}

	// Record deployment
	deployedBy := keyName
	if err := hosting.RecordDeployment(db, result.SiteID, result.SizeBytes, result.FileCount, deployedBy); err != nil {
		log.Printf("Failed to record deployment: %v", err)
	}

	// Record rate limit
	limiter.RecordDeploy(clientIP)

	log.Printf("Site deployed: %s by %s (key_id=%d), %d files, %d bytes",
		siteName, keyName, keyID, result.FileCount, result.SizeBytes)

	// Return success response
	api.Success(w, http.StatusOK, map[string]interface{}{
		"site":       siteName,
		"file_count": result.FileCount,
		"size_bytes": result.SizeBytes,
		"message":    "Deployment successful",
	})
}

// jsonError sends a JSON error response
// Note: This is kept for backward compatibility with other handlers that still use it
// New code should use api.* helpers instead
func jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}
