// Package debug provides debug logging utilities for fazt.
// Debug mode is enabled via FAZT_DEBUG=1 or automatically in development mode.
package debug

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	enabled     bool
	enabledOnce sync.Once
)

// IsEnabled returns true if debug mode is active.
// Checks FAZT_DEBUG env var on first call and caches the result.
func IsEnabled() bool {
	enabledOnce.Do(func() {
		debug := os.Getenv("FAZT_DEBUG")
		if debug != "" {
			enabled = debug == "1" || debug == "true"
		} else {
			// Check if running in development mode
			env := os.Getenv("ENV")
			enabled = env == "" || env == "development"
		}
		if enabled {
			log.Printf("[DEBUG] Debug mode enabled")
		}
	})
	return enabled
}

// Log logs a debug message if debug mode is enabled.
func Log(category, format string, args ...interface{}) {
	if !IsEnabled() {
		return
	}
	msg := fmt.Sprintf(format, args...)
	log.Printf("[DEBUG %s] %s", category, msg)
}

// Warn logs a warning message if debug mode is enabled.
func Warn(category, format string, args ...interface{}) {
	if !IsEnabled() {
		return
	}
	msg := fmt.Sprintf(format, args...)
	log.Printf("[WARN  %s] %s", category, msg)
}

// StorageOp logs a storage operation with timing.
func StorageOp(op, app, collection string, query interface{}, rows int64, duration time.Duration) {
	if !IsEnabled() {
		return
	}
	queryStr := formatQuery(query)
	log.Printf("[DEBUG storage] %s %s/%s query=%s rows=%d took=%s",
		op, app, collection, queryStr, rows, duration.Round(time.Microsecond))
}

// RuntimeReq logs a runtime request with timing.
func RuntimeReq(reqID, app, path string, status int, duration time.Duration) {
	if !IsEnabled() {
		return
	}
	log.Printf("[DEBUG runtime] req=%s app=%s path=%s status=%d took=%s",
		reqID, app, path, status, duration.Round(time.Microsecond))
}

// RuntimePool logs VM pool state.
func RuntimePool(poolSize, available int) {
	if !IsEnabled() {
		return
	}
	log.Printf("[DEBUG runtime] pool: size=%d available=%d", poolSize, available)
}

// SQL logs a SQL query if debug is enabled.
func SQL(query string, args []interface{}) {
	if !IsEnabled() {
		return
	}
	// Truncate long queries
	q := strings.TrimSpace(query)
	q = strings.ReplaceAll(q, "\n", " ")
	q = strings.ReplaceAll(q, "\t", " ")
	// Collapse multiple spaces
	for strings.Contains(q, "  ") {
		q = strings.ReplaceAll(q, "  ", " ")
	}
	if len(q) > 200 {
		q = q[:197] + "..."
	}
	log.Printf("[DEBUG sql] %s args=%v", q, args)
}

// formatQuery formats a query object for logging.
func formatQuery(query interface{}) string {
	if query == nil {
		return "{}"
	}
	b, err := json.Marshal(query)
	if err != nil {
		return fmt.Sprintf("%v", query)
	}
	s := string(b)
	if len(s) > 100 {
		s = s[:97] + "..."
	}
	return s
}
