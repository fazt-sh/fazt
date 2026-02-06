package egress

import (
	"database/sql"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/fazt-sh/fazt/internal/system"
)

// NetLogEntry represents a single outbound HTTP log entry.
type NetLogEntry struct {
	AppID         string
	Domain        string
	Method        string
	Path          string
	Status        int
	ErrorCode     string
	DurationMs    int
	RequestBytes  int64
	ResponseBytes int64
	Timestamp     int64
}

// NetLogger buffers outbound HTTP log entries and flushes them to SQLite.
type NetLogger struct {
	buffer     []NetLogEntry
	mu         sync.Mutex
	db         *sql.DB
	done       chan struct{}
	wg         sync.WaitGroup
	bufferSize int
	flushMs    int
}

// NewNetLogger creates a NetLogger with settings from system.Limits.Net.
func NewNetLogger(db *sql.DB) *NetLogger {
	netLimits := system.GetLimits().Net
	return &NetLogger{
		buffer:     make([]NetLogEntry, 0, netLimits.LogBufferSize),
		db:         db,
		done:       make(chan struct{}),
		bufferSize: netLimits.LogBufferSize,
		flushMs:    netLimits.LogFlushMs,
	}
}

// Start begins the background flush ticker.
func (l *NetLogger) Start() {
	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		ticker := time.NewTicker(time.Duration(l.flushMs) * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				l.flush()
			case <-l.done:
				return
			}
		}
	}()
}

// Stop flushes remaining entries and stops the background ticker.
func (l *NetLogger) Stop() {
	close(l.done)
	l.wg.Wait()
	l.flush() // Final flush
}

// Log adds an entry to the buffer. Non-blocking; drops if buffer full.
// Errors (non-2xx or network failures) are flushed immediately.
func (l *NetLogger) Log(entry NetLogEntry) {
	if entry.Timestamp == 0 {
		entry.Timestamp = time.Now().Unix()
	}

	// Strip query string from path (may contain secret tokens)
	entry.Path = stripQueryString(entry.Path)

	l.mu.Lock()

	// Errors bypass buffer for immediate logging
	isError := entry.ErrorCode != "" || entry.Status >= 400
	if isError {
		l.buffer = append(l.buffer, entry)
		l.mu.Unlock()
		go l.flush()
		return
	}

	// Drop if buffer full (never block the request)
	if len(l.buffer) >= l.bufferSize {
		l.mu.Unlock()
		return
	}

	l.buffer = append(l.buffer, entry)
	l.mu.Unlock()
}

// LogFromFetch creates a log entry from a fetch result.
func (l *NetLogger) LogFromFetch(appID, rawURL, method string, resp *FetchResponse,
	fetchErr error, duration time.Duration, reqBytes int64) {

	entry := NetLogEntry{
		AppID:        appID,
		Method:       method,
		DurationMs:   int(duration.Milliseconds()),
		RequestBytes: reqBytes,
	}

	// Parse URL for domain and path
	if parsed, err := url.Parse(rawURL); err == nil {
		entry.Domain = canonicalizeHost(parsed.Hostname())
		entry.Path = parsed.Path
	}

	if resp != nil {
		entry.Status = resp.Status
		entry.ResponseBytes = int64(len(resp.body))
	}

	if fetchErr != nil {
		if ee, ok := fetchErr.(*EgressError); ok {
			entry.ErrorCode = ee.Code
		} else {
			entry.ErrorCode = CodeError
		}
	}

	l.Log(entry)
}

func (l *NetLogger) flush() {
	l.mu.Lock()
	if len(l.buffer) == 0 {
		l.mu.Unlock()
		return
	}

	batch := l.buffer
	l.buffer = make([]NetLogEntry, 0, l.bufferSize)
	l.mu.Unlock()

	if err := l.writeBatch(batch); err != nil {
		log.Printf("egress: failed to flush %d log entries: %v", len(batch), err)
	}
}

func (l *NetLogger) writeBatch(batch []NetLogEntry) error {
	tx, err := l.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO net_log (app_id, domain, method, path, status, error_code,
		                     duration_ms, request_bytes, response_bytes, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range batch {
		var status, errCode interface{}
		if e.Status > 0 {
			status = e.Status
		}
		if e.ErrorCode != "" {
			errCode = e.ErrorCode
		}

		_, err := stmt.Exec(e.AppID, e.Domain, e.Method, e.Path,
			status, errCode, e.DurationMs,
			e.RequestBytes, e.ResponseBytes, e.Timestamp)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func stripQueryString(rawPath string) string {
	if idx := strings.Index(rawPath, "?"); idx != -1 {
		return rawPath[:idx]
	}
	return rawPath
}
