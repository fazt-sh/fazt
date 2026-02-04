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
	// Require API key auth (bypasses AdminMiddleware for remote peer access)
	if !requireAPIKeyAuth(w, r) {
		return
	}

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

// SystemCapacityHandler returns capacity estimates and limits
func SystemCapacityHandler(w http.ResponseWriter, r *http.Request) {
	limits := system.GetLimits()

	capacity := map[string]interface{}{
		"system": map[string]interface{}{
			"total_ram_mb": float64(limits.TotalRAM) / 1024 / 1024,
			"cpu_cores":    limits.CPUCount,
		},
		"capacity": map[string]interface{}{
			"concurrent_users":     limits.ConcurrentUsers,
			"concurrent_users_max": limits.ConcurrentUsersMax,
			"read_throughput":      limits.ReadThroughput,
			"write_throughput":     limits.WriteThroughput,
			"mixed_throughput":     limits.MixedThroughput,
		},
		"limits": map[string]interface{}{
			"max_vfs_mb":    float64(limits.MaxVFSBytes) / 1024 / 1024,
			"max_upload_mb": float64(limits.MaxUploadBytes) / 1024 / 1024,
		},
		"architecture": map[string]interface{}{
			"storage":           "SQLite + WAL mode",
			"write_strategy":    "Single-writer serialization (WriteQueue)",
			"overload_behavior": "HTTP 503 with Retry-After header",
		},
		"tested": map[string]interface{}{
			"version":     "0.10.10",
			"date":        "2026-01-24",
			"environment": "Stress test with concurrent HTTP clients",
		},
	}

	api.Success(w, http.StatusOK, capacity)
}