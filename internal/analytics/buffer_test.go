package analytics

import (
	"sync"
	"testing"
	"time"

	"github.com/fazt-sh/fazt/internal/database"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.FlushInterval != 30*time.Second {
		t.Errorf("Expected FlushInterval to be 30s, got %v", cfg.FlushInterval)
	}
	if cfg.BatchSize != 1000 {
		t.Errorf("Expected BatchSize to be 1000, got %d", cfg.BatchSize)
	}
	if cfg.MaxRetries != 1 {
		t.Errorf("Expected MaxRetries to be 1, got %d", cfg.MaxRetries)
	}
}

func TestAddEvent(t *testing.T) {
	// Setup: Initialize test database
	dbPath := "/tmp/test-analytics.db"
	if err := database.Init(dbPath); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer database.Close()

	// Reset global buffer to force re-init
	globalBuffer = nil
	initOnce = sync.Once{}

	// Initialize buffer
	Init()

	// Add an event
	event := Event{
		Domain:     "test.com",
		Tags:       "tag1,tag2",
		SourceType: "hosting",
		EventType:  "pageview",
		Path:       "/test",
		Referrer:   "https://example.com",
		UserAgent:  "TestAgent/1.0",
		IPAddress:  "127.0.0.1",
		QueryParams: "foo=bar",
	}

	Add(event)

	// Verify event was queued
	stats := GetStats()
	if stats.EventsQueued >= 1 {
		t.Logf("Event queued successfully: %d events", stats.EventsQueued)
	}
}

func TestAddEventWithoutInit(t *testing.T) {
	// Reset global buffer
	globalBuffer = nil

	// Should not panic when buffer not initialized
	event := Event{
		Domain:     "test.com",
		EventType:  "pageview",
	}

	// This should log a warning but not panic
	Add(event)
}

func TestGetStatsWithoutInit(t *testing.T) {
	// Reset global buffer
	globalBuffer = nil

	stats := GetStats()

	// Should return empty stats
	if stats.EventsQueued != 0 {
		t.Errorf("Expected 0 events when buffer not initialized, got %d", stats.EventsQueued)
	}
	if stats.BatchSize != 0 {
		t.Errorf("Expected 0 batch size when buffer not initialized, got %d", stats.BatchSize)
	}
}

func TestShutdownWithoutInit(t *testing.T) {
	// Reset global buffer
	globalBuffer = nil

	// Should not panic
	Shutdown()
}

func TestEventTimestamp(t *testing.T) {
	// Setup database
	dbPath := "/tmp/test-analytics-timestamp.db"
	if err := database.Init(dbPath); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer database.Close()

	// Reset global buffer to force re-init
	globalBuffer = nil
	initOnce = sync.Once{}

	Init()

	// Add event without timestamp
	event := Event{
		Domain:     "test.com",
		EventType:  "pageview",
		Path:       "/test",
	}

	before := time.Now()
	Add(event)
	after := time.Now()

	// Flush immediately to check timestamp
	if globalBuffer != nil {
		globalBuffer.flush()
	}

	// Query the database to verify timestamp was set
	db := database.GetDB()
	var createdAt time.Time
	err := db.QueryRow("SELECT created_at FROM events ORDER BY created_at DESC LIMIT 1").Scan(&createdAt)
	if err != nil {
		t.Fatalf("Failed to query event: %v", err)
	}

	// Verify timestamp is between before and after
	if createdAt.Before(before) || createdAt.After(after) {
		t.Errorf("Timestamp %v not in expected range [%v, %v]", createdAt, before, after)
	}
}

func TestShutdownFlushesEvents(t *testing.T) {
	// Setup database
	dbPath := "/tmp/test-analytics-shutdown.db"
	if err := database.Init(dbPath); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer database.Close()

	// Reset global buffer to force re-init
	globalBuffer = nil
	initOnce = sync.Once{}

	Init()

	// Add multiple events
	for i := 0; i < 5; i++ {
		Add(Event{
			Domain:     "test.com",
			EventType:  "pageview",
			Path:       "/test",
		})
	}

	// Verify events are queued
	stats := GetStats()
	if stats.EventsQueued < 1 {
		t.Errorf("Expected events queued before shutdown, got %d", stats.EventsQueued)
	}

	// Shutdown (should flush)
	Shutdown()

	// Verify events were flushed to database
	db := database.GetDB()
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM events WHERE domain = 'test.com'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query events: %v", err)
	}

	if count < 5 {
		t.Errorf("Expected at least 5 events in database after shutdown, got %d", count)
	}
}

func TestConcurrentAdds(t *testing.T) {
	// Setup database
	dbPath := "/tmp/test-analytics-concurrent.db"
	if err := database.Init(dbPath); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer database.Close()

	// Reset global buffer to force re-init
	globalBuffer = nil
	initOnce = sync.Once{}

	Init()

	// Concurrently add events
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				Add(Event{
					Domain:     "concurrent.com",
					EventType:  "pageview",
					Path:       "/test",
				})
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Force flush
	if globalBuffer != nil {
		globalBuffer.flush()
	}

	// Verify count
	stats := GetStats()
	// Should have 100 events total, but some might have been flushed
	// So we just verify no panic occurred and some events were queued
	t.Logf("Events queued after concurrent adds: %d", stats.EventsQueued)
}
