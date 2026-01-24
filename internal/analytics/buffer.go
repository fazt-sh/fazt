package analytics

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/storage"
)

// Event represents a single tracking event
type Event struct {
	Domain      string
	Tags        string
	SourceType  string
	EventType   string
	Path        string
	Referrer    string
	UserAgent   string
	IPAddress   string
	QueryParams string
	CreatedAt   time.Time
}

// Config holds the buffer configuration
type Config struct {
	FlushInterval time.Duration
	BatchSize     int
	MaxRetries    int
}

// DefaultConfig returns safe defaults
func DefaultConfig() Config {
	return Config{
		FlushInterval: 30 * time.Second,
		BatchSize:     1000,
		MaxRetries:    1,
	}
}

// Buffer aggregates events and flushes them to the DB
type Buffer struct {
	mu         sync.Mutex
	events     []Event
	config     Config
	stopChan   chan struct{}
	wg         sync.WaitGroup
	isShutdown bool
}

var (
	// globalBuffer is the singleton instance
	globalBuffer *Buffer
	initOnce     sync.Once
)

// Init initializes the global analytics buffer
func Init() {
	initOnce.Do(func() {
		globalBuffer = &Buffer{
			events:   make([]Event, 0, DefaultConfig().BatchSize),
			config:   DefaultConfig(),
			stopChan: make(chan struct{}),
		}
		globalBuffer.startFlusher()
		log.Println("Analytics: Write buffer initialized")
	})
}

// Add queues an event for writing
func Add(e Event) {
	if globalBuffer == nil {
		// Fallback if not initialized (shouldn't happen in prod)
		log.Println("Warning: Analytics buffer not initialized, dropping event")
		return
	}

	// Capture timestamp if not set
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now()
	}

	globalBuffer.mu.Lock()
	defer globalBuffer.mu.Unlock()

	if globalBuffer.isShutdown {
		return
	}

	globalBuffer.events = append(globalBuffer.events, e)

	// Trigger flush if batch size reached
	if len(globalBuffer.events) >= globalBuffer.config.BatchSize {
		go globalBuffer.flush()
	}
}

// BufferStats holds current buffer metrics
type BufferStats struct {
	EventsQueued int
	BatchSize    int
}

// GetStats returns the current buffer statistics
func GetStats() BufferStats {
	if globalBuffer == nil {
		return BufferStats{}
	}
	globalBuffer.mu.Lock()
	defer globalBuffer.mu.Unlock()
	return BufferStats{
		EventsQueued: len(globalBuffer.events),
		BatchSize:    globalBuffer.config.BatchSize,
	}
}

// Shutdown flushes remaining events and stops the background worker
func Shutdown() {
	if globalBuffer == nil {
		return
	}

	globalBuffer.mu.Lock()
	if globalBuffer.isShutdown {
		globalBuffer.mu.Unlock()
		return
	}
	globalBuffer.isShutdown = true
	globalBuffer.mu.Unlock()

	close(globalBuffer.stopChan)
	globalBuffer.wg.Wait()
	
	// Final flush
	globalBuffer.flush()
	log.Println("Analytics: Buffer shutdown complete")
}

func (b *Buffer) startFlusher() {
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		ticker := time.NewTicker(b.config.FlushInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				b.flush()
			case <-b.stopChan:
				return
			}
		}
	}()
}

// flush writes the current buffer to the database
func (b *Buffer) flush() {
	b.mu.Lock()
	if len(b.events) == 0 {
		b.mu.Unlock()
		return
	}

	// Swap buffers
	batch := b.events
	b.events = make([]Event, 0, b.config.BatchSize)
	b.mu.Unlock()

	if err := b.writeBatch(batch); err != nil {
		log.Printf("Analytics Error: Failed to flush batch of %d events: %v", len(batch), err)
		// Note: In a production system, we might retry or persist to disk.
		// For now, per plan, we drop if failing to preserve system stability.
	} else {
		if len(batch) > 100 {
			log.Printf("Analytics: Flushed %d events", len(batch))
		}
	}
}

func (b *Buffer) writeBatch(batch []Event) error {
	db := database.GetDB()
	if db == nil {
		return fmt.Errorf("database not available")
	}

	// Route through WriteQueue to prevent SQLITE_BUSY errors
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return storage.QueueWrite(ctx, func() error {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		stmt, err := tx.Prepare(`
			INSERT INTO events (domain, tags, source_type, event_type, path, referrer, user_agent, ip_address, query_params, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, e := range batch {
			_, err := stmt.Exec(
				e.Domain,
				e.Tags,
				e.SourceType,
				e.EventType,
				e.Path,
				e.Referrer,
				e.UserAgent,
				e.IPAddress,
				e.QueryParams,
				e.CreatedAt,
			)
			if err != nil {
				return err
			}
		}

		return tx.Commit()
	})
}
