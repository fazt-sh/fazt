package activity

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/storage"
	"github.com/fazt-sh/fazt/internal/system"
)

// Weight constants for activity importance (0-9, higher = more important)
const (
	WeightDebug        = 0 // Request timing, cache hits
	WeightSystem       = 1 // Health checks, backups
	WeightAnalytics    = 2 // Pageview, click (default)
	WeightNavigation   = 3 // Internal page views
	WeightUserAction   = 4 // Form submissions
	WeightDataMutation = 5 // KV set/delete, doc CRUD
	WeightDeployment   = 6 // Site deploy/delete
	WeightConfig       = 7 // Alias CRUD, redirect CRUD
	WeightAuth         = 8 // Login/logout, session create/delete
	WeightSecurity     = 9 // API key create/delete, role changes
)

// Actor types
const (
	ActorUser      = "user"
	ActorSystem    = "system"
	ActorAPIKey    = "api_key"
	ActorAnonymous = "anonymous"
)

// Entry represents a single activity log entry
type Entry struct {
	Timestamp    time.Time
	ActorType    string
	ActorID      string
	ActorIP      string
	ActorUA      string
	ResourceType string
	ResourceID   string
	Action       string
	Result       string
	Weight       int
	Details      map[string]interface{}
}

// Config holds the buffer configuration
type Config struct {
	FlushInterval time.Duration
	BatchSize     int
	MaxRows       int
	CleanupBatch  int
}

// DefaultConfig returns safe defaults
func DefaultConfig() Config {
	limits := system.GetLimits()
	return Config{
		FlushInterval: 10 * time.Second,
		BatchSize:     500,
		MaxRows:       limits.Storage.MaxLogRows,
		CleanupBatch:  10000,
	}
}

// Logger aggregates activity entries and flushes them to the DB
type Logger struct {
	mu         sync.Mutex
	entries    []Entry
	config     Config
	stopChan   chan struct{}
	wg         sync.WaitGroup
	isShutdown bool
}

var (
	globalLogger *Logger
	initOnce     sync.Once
)

// Init initializes the global activity logger
func Init() {
	initOnce.Do(func() {
		globalLogger = &Logger{
			entries:  make([]Entry, 0, DefaultConfig().BatchSize),
			config:   DefaultConfig(),
			stopChan: make(chan struct{}),
		}
		globalLogger.startFlusher()
		log.Println("Activity: Logger initialized")
	})
}

// Log queues an entry for writing
func Log(e Entry) {
	if globalLogger == nil {
		return
	}

	// Set timestamp if not provided
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}

	// Default result
	if e.Result == "" {
		e.Result = "success"
	}

	// Default actor type
	if e.ActorType == "" {
		e.ActorType = ActorSystem
	}

	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()

	if globalLogger.isShutdown {
		return
	}

	globalLogger.entries = append(globalLogger.entries, e)

	// Trigger flush if batch size reached
	if len(globalLogger.entries) >= globalLogger.config.BatchSize {
		go globalLogger.flush()
	}
}

// LogSuccess logs a successful action
func LogSuccess(actorType, actorID, ip, resourceType, resourceID, action string, weight int, details map[string]interface{}) {
	Log(Entry{
		ActorType:    actorType,
		ActorID:      actorID,
		ActorIP:      ip,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Action:       action,
		Result:       "success",
		Weight:       weight,
		Details:      details,
	})
}

// LogFailure logs a failed action
func LogFailure(actorType, actorID, ip, resourceType, resourceID, action, reason string, weight int) {
	Log(Entry{
		ActorType:    actorType,
		ActorID:      actorID,
		ActorIP:      ip,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Action:       action,
		Result:       "failure",
		Weight:       weight,
		Details:      map[string]interface{}{"reason": reason},
	})
}

// LogFromRequest extracts context from HTTP request and logs an entry
func LogFromRequest(r *http.Request, userID, resourceType, resourceID, action string, weight int, details map[string]interface{}) {
	actorType := ActorAnonymous
	actorID := ""

	if userID != "" {
		actorType = ActorUser
		actorID = userID
	}

	// Check for API key auth
	if authHeader := r.Header.Get("Authorization"); authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		actorType = ActorAPIKey
	}

	Log(Entry{
		ActorType:    actorType,
		ActorID:      actorID,
		ActorIP:      ExtractIP(r),
		ActorUA:      r.UserAgent(),
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Action:       action,
		Result:       "success",
		Weight:       weight,
		Details:      details,
	})
}

// ExtractIP gets the client's IP address from the request
func ExtractIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies/load balancers)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
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

// GetBufferStats returns the current buffer statistics
func GetBufferStats() (queued int, batchSize int) {
	if globalLogger == nil {
		return 0, 0
	}
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	return len(globalLogger.entries), globalLogger.config.BatchSize
}

// Shutdown flushes remaining entries and stops the background worker
func Shutdown() {
	if globalLogger == nil {
		return
	}

	globalLogger.mu.Lock()
	if globalLogger.isShutdown {
		globalLogger.mu.Unlock()
		return
	}
	globalLogger.isShutdown = true
	globalLogger.mu.Unlock()

	close(globalLogger.stopChan)
	globalLogger.wg.Wait()

	// Final flush
	globalLogger.flush()
	log.Println("Activity: Logger shutdown complete")
}

func (l *Logger) startFlusher() {
	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		ticker := time.NewTicker(l.config.FlushInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				l.flush()
				l.checkAndCleanup()
			case <-l.stopChan:
				return
			}
		}
	}()
}

// flush writes the current buffer to the database
func (l *Logger) flush() {
	l.mu.Lock()
	if len(l.entries) == 0 {
		l.mu.Unlock()
		return
	}

	// Swap buffers
	batch := l.entries
	l.entries = make([]Entry, 0, l.config.BatchSize)
	l.mu.Unlock()

	if err := l.writeBatch(batch); err != nil {
		log.Printf("Activity Error: Failed to flush batch of %d entries: %v", len(batch), err)
	} else if len(batch) > 50 {
		log.Printf("Activity: Flushed %d entries", len(batch))
	}
}

func (l *Logger) writeBatch(batch []Entry) error {
	db := database.GetDB()
	if db == nil {
		return fmt.Errorf("database not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return storage.QueueWrite(ctx, func() error {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		stmt, err := tx.Prepare(`
			INSERT INTO activity_log (timestamp, actor_type, actor_id, actor_ip, actor_ua, resource_type, resource_id, action, result, weight, details)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, e := range batch {
			var detailsJSON *string
			if e.Details != nil && len(e.Details) > 0 {
				bytes, err := json.Marshal(e.Details)
				if err == nil {
					s := string(bytes)
					detailsJSON = &s
				}
			}

			_, err := stmt.Exec(
				e.Timestamp.Unix(),
				e.ActorType,
				e.ActorID,
				e.ActorIP,
				e.ActorUA,
				e.ResourceType,
				e.ResourceID,
				e.Action,
				e.Result,
				e.Weight,
				detailsJSON,
			)
			if err != nil {
				return err
			}
		}

		return tx.Commit()
	})
}

// checkAndCleanup checks row count and cleans up if necessary
func (l *Logger) checkAndCleanup() {
	db := database.GetDB()
	if db == nil {
		return
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM activity_log").Scan(&count)
	if err != nil {
		return
	}

	if count > l.config.MaxRows {
		// Delete oldest low-weight entries first
		_, err := db.Exec(`
			DELETE FROM activity_log
			WHERE id IN (
				SELECT id FROM activity_log
				ORDER BY weight ASC, timestamp ASC
				LIMIT ?
			)
		`, l.config.CleanupBatch)
		if err != nil {
			log.Printf("Activity: Cleanup failed: %v", err)
		} else {
			log.Printf("Activity: Auto-cleaned %d old entries", l.config.CleanupBatch)
		}
	}
}
