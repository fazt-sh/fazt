package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/fazt-sh/fazt/internal/analytics"
	"github.com/fazt-sh/fazt/internal/api"
	"github.com/fazt-sh/fazt/internal/config"
	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/hosting"
	"github.com/fazt-sh/fazt/internal/system"
)

var startTime = time.Now()

// SystemHealthHandler returns the system health status and metrics
func SystemHealthHandler(w http.ResponseWriter, r *http.Request) {
	// Get Memory Stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Get DB Stats
	dbStats := database.GetDBStats()

	// Get Analytics Stats
	bufferStats := analytics.GetStats()

	// Get VFS Stats
	vfsStats := hosting.GetStats()

	// Get Limits
	limits := system.GetLimits()

	response := map[string]interface{}{
		"status":         "healthy",
		"uptime_seconds": time.Since(startTime).Seconds(),
		"version":        config.Version,
		"mode":           config.Get().Server.Env,
		"memory": map[string]interface{}{
			"used_mb":      float64(m.Alloc) / 1024 / 1024,
			"limit_mb":     float64(limits.TotalRAM) / 1024 / 1024,
			"vfs_cache_mb": float64(vfsStats.CacheSizeBytes) / 1024 / 1024,
		},
		"database": map[string]interface{}{
			"path":             config.Get().Database.Path,
			"open_connections": dbStats.OpenConnections,
			"in_use":           dbStats.InUse,
		},
		"runtime": map[string]interface{}{
			"queued_events": bufferStats.EventsQueued,
			"goroutines":    runtime.NumGoroutine(),
		},
	}

	api.Success(w, http.StatusOK, response)
}

// SystemLimitsHandler returns the resource limits
func SystemLimitsHandler(w http.ResponseWriter, r *http.Request) {
	limits := system.GetLimits()
	api.Success(w, http.StatusOK, limits)
}

// SystemCacheHandler returns VFS cache statistics
func SystemCacheHandler(w http.ResponseWriter, r *http.Request) {
	stats := hosting.GetStats()
	api.Success(w, http.StatusOK, stats)
}

// SystemDBHandler returns database statistics
func SystemDBHandler(w http.ResponseWriter, r *http.Request) {
	stats := database.GetDBStats()
	api.Success(w, http.StatusOK, stats)
}

// SystemConfigHandler returns the server configuration (sanitized)
func SystemConfigHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()
	// Create sanitized copy
	safeCfg := map[string]interface{}{
		"version": config.Version,
		"domain":  cfg.Server.Domain,
		"env":     cfg.Server.Env,
		"https":   cfg.HTTPS.Enabled,
		"ntfy":    cfg.Ntfy.URL != "",
	}
	api.Success(w, http.StatusOK, safeCfg)
}