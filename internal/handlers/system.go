package handlers

import (
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/fazt-sh/fazt/internal/activity"
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

// SystemLogsHandler returns activity log entries with filtering
func SystemLogsHandler(w http.ResponseWriter, r *http.Request) {
	if !requireAPIKeyAuth(w, r) {
		return
	}

	params := parseLogQueryParams(r)
	entries, total, err := activity.Query(database.GetDB(), params)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Failed to query logs", err.Error(), nil)
		return
	}

	api.Success(w, http.StatusOK, map[string]interface{}{
		"entries": entries,
		"total":   total,
		"showing": len(entries),
		"offset":  params.Offset,
		"limit":   params.Limit,
	})
}

// SystemLogsStatsHandler returns activity log statistics
func SystemLogsStatsHandler(w http.ResponseWriter, r *http.Request) {
	if !requireAPIKeyAuth(w, r) {
		return
	}

	params := parseLogQueryParams(r)
	stats, err := activity.GetStatsFiltered(database.GetDB(), params)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Failed to get stats", err.Error(), nil)
		return
	}

	api.Success(w, http.StatusOK, stats)
}

// SystemLogsCleanupHandler deletes activity logs matching filters
func SystemLogsCleanupHandler(w http.ResponseWriter, r *http.Request) {
	if !requireAPIKeyAuth(w, r) {
		return
	}

	params := parseLogQueryParams(r)
	dryRun := r.URL.Query().Get("force") != "true"

	count, err := activity.Cleanup(database.GetDB(), params, dryRun)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Cleanup failed", err.Error(), nil)
		return
	}

	api.Success(w, http.StatusOK, map[string]interface{}{
		"deleted": count,
		"dry_run": dryRun,
		"filters": activity.DescribeFilters(params),
	})
}

// parseLogQueryParams extracts activity log query params from request
func parseLogQueryParams(r *http.Request) activity.QueryParams {
	q := r.URL.Query()
	params := activity.QueryParams{
		Limit:  activity.DefaultLimit,
		Offset: 0,
	}

	if v := q.Get("min_weight"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			params.MinWeight = &n
		}
	}
	if v := q.Get("max_weight"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			params.MaxWeight = &n
		}
	}
	if v := q.Get("type"); v != "" {
		params.ResourceType = v
	}
	if v := q.Get("resource"); v != "" {
		params.ResourceID = v
	}
	if v := q.Get("app"); v != "" {
		params.AppID = v
	}
	if v := q.Get("user"); v != "" {
		params.UserID = v
	}
	if v := q.Get("actor_type"); v != "" {
		params.ActorType = v
	}
	if v := q.Get("action"); v != "" {
		params.Action = v
	}
	if v := q.Get("result"); v != "" {
		params.Result = v
	}
	if v := q.Get("since"); v != "" {
		if t, err := parseTimeParam(v); err == nil {
			params.Since = &t
		}
	}
	if v := q.Get("until"); v != "" {
		if t, err := parseTimeParam(v); err == nil {
			params.Until = &t
		}
	}
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			params.Limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			params.Offset = n
		}
	}

	return params
}

// parseTimeParam parses duration strings (24h, 7d) or dates (2024-01-15)
func parseTimeParam(s string) (time.Time, error) {
	// Try duration first (e.g., "24h", "7d")
	if len(s) > 1 {
		unit := s[len(s)-1]
		numStr := s[:len(s)-1]
		if num, err := strconv.Atoi(numStr); err == nil {
			switch unit {
			case 'h':
				return time.Now().Add(-time.Duration(num) * time.Hour), nil
			case 'd':
				return time.Now().Add(-time.Duration(num) * 24 * time.Hour), nil
			case 'm':
				return time.Now().Add(-time.Duration(num) * time.Minute), nil
			}
		}
	}
	// Try date format
	return time.Parse("2006-01-02", s)
}