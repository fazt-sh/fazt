package egress

import (
	"testing"
	"time"

	"github.com/fazt-sh/fazt/internal/system"
)

func TestNetLoggerBufferAndFlush(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	system.ResetCachedLimits()
	logger := &NetLogger{
		buffer:     make([]NetLogEntry, 0, 100),
		db:         db,
		done:       make(chan struct{}),
		bufferSize: 100,
		flushMs:    100,
	}

	// Log some entries
	for i := 0; i < 5; i++ {
		logger.Log(NetLogEntry{
			AppID:      "test-app",
			Domain:     "api.example.com",
			Method:     "GET",
			Path:       "/data",
			Status:     200,
			DurationMs: 50,
		})
	}

	// Manual flush
	logger.flush()

	// Check DB
	var count int
	db.QueryRow("SELECT COUNT(*) FROM net_log").Scan(&count)
	if count != 5 {
		t.Errorf("expected 5 log entries, got %d", count)
	}
}

func TestNetLoggerErrorBypassesBuffer(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	system.ResetCachedLimits()
	logger := &NetLogger{
		buffer:     make([]NetLogEntry, 0, 100),
		db:         db,
		done:       make(chan struct{}),
		bufferSize: 100,
		flushMs:    60000, // Very long flush interval
	}

	// Log an error entry
	logger.Log(NetLogEntry{
		AppID:     "test-app",
		Domain:    "api.example.com",
		Method:    "GET",
		Path:      "/data",
		ErrorCode: "NET_TIMEOUT",
		DurationMs: 5000,
	})

	// Wait a bit for the async flush goroutine
	time.Sleep(50 * time.Millisecond)

	var count int
	db.QueryRow("SELECT COUNT(*) FROM net_log").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 error entry flushed immediately, got %d", count)
	}
}

func TestNetLoggerDropsWhenFull(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	system.ResetCachedLimits()
	logger := &NetLogger{
		buffer:     make([]NetLogEntry, 0, 3),
		db:         db,
		done:       make(chan struct{}),
		bufferSize: 3,
		flushMs:    60000,
	}

	// Fill buffer
	for i := 0; i < 3; i++ {
		logger.Log(NetLogEntry{
			AppID:      "test-app",
			Domain:     "example.com",
			Method:     "GET",
			Path:       "/data",
			Status:     200,
			DurationMs: 10,
		})
	}

	// This should be dropped (buffer full, not an error)
	logger.Log(NetLogEntry{
		AppID:      "test-app",
		Domain:     "example.com",
		Method:     "GET",
		Path:       "/dropped",
		Status:     200,
		DurationMs: 10,
	})

	logger.flush()

	var count int
	db.QueryRow("SELECT COUNT(*) FROM net_log").Scan(&count)
	if count != 3 {
		t.Errorf("expected 3 entries (4th dropped), got %d", count)
	}
}

func TestStripQueryString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/api/data?key=secret", "/api/data"},
		{"/api/data", "/api/data"},
		{"/search?q=hello&page=2", "/search"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := stripQueryString(tt.input); got != tt.want {
			t.Errorf("stripQueryString(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNetLoggerStartStop(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	system.ResetCachedLimits()
	logger := NewNetLogger(db)
	logger.Start()

	// Log some entries
	logger.Log(NetLogEntry{
		AppID:      "test-app",
		Domain:     "example.com",
		Method:     "GET",
		Path:       "/data",
		Status:     200,
		DurationMs: 10,
	})

	// Stop should flush
	logger.Stop()

	var count int
	db.QueryRow("SELECT COUNT(*) FROM net_log").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 entry after stop (final flush), got %d", count)
	}
}
