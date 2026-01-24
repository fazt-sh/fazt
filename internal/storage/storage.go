// Package storage provides app-scoped storage primitives for fazt serverless functions.
// It implements key-value, document, and blob storage backed by SQLite.
package storage

import (
	"context"
	"database/sql"
	"sync"
	"time"
)

var (
	globalWriter *WriteQueue
	writerOnce   sync.Once
)

// InitWriter initializes the global write queue. Call once at server startup.
func InitWriter() {
	writerOnce.Do(func() {
		globalWriter = NewWriteQueue(DefaultWriteQueueConfig())
	})
}

// GetWriter returns the global write queue for serializing all DB writes.
// Returns nil if InitWriter hasn't been called.
func GetWriter() *WriteQueue {
	return globalWriter
}

// QueueWrite submits a write operation to the global queue.
// This is the preferred way for non-storage packages (like analytics) to do writes.
func QueueWrite(ctx context.Context, fn func() error) error {
	if globalWriter == nil {
		// Fallback: execute directly (shouldn't happen in production)
		return fn()
	}
	return globalWriter.Write(ctx, fn)
}

// KVStore provides key-value storage operations.
type KVStore interface {
	Set(ctx context.Context, appID, key string, value interface{}, ttl *time.Duration) error
	Get(ctx context.Context, appID, key string) (interface{}, error)
	Delete(ctx context.Context, appID, key string) error
	List(ctx context.Context, appID, prefix string) ([]KVEntry, error)
}

// KVEntry represents a key-value pair.
type KVEntry struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	ExpiresAt *time.Time  `json:"expires_at,omitempty"`
}

// DocStore provides document storage operations.
type DocStore interface {
	Insert(ctx context.Context, appID, collection string, doc map[string]interface{}) (string, error)
	Find(ctx context.Context, appID, collection string, query map[string]interface{}) ([]Document, error)
	FindOne(ctx context.Context, appID, collection, id string) (*Document, error)
	Update(ctx context.Context, appID, collection string, query, changes map[string]interface{}) (int64, error)
	Delete(ctx context.Context, appID, collection string, query map[string]interface{}) (int64, error)
}

// Document represents a stored document.
type Document struct {
	ID        string                 `json:"id"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// BlobStore provides blob storage operations.
type BlobStore interface {
	Put(ctx context.Context, appID, path string, data []byte, mimeType string) error
	Get(ctx context.Context, appID, path string) (*Blob, error)
	Delete(ctx context.Context, appID, path string) error
	List(ctx context.Context, appID, prefix string) ([]BlobMeta, error)
}

// Blob represents a stored blob.
type Blob struct {
	Data     []byte `json:"data"`
	MimeType string `json:"mime"`
	Size     int64  `json:"size"`
	Hash     string `json:"hash"`
}

// BlobMeta represents blob metadata for listings.
type BlobMeta struct {
	Path      string    `json:"path"`
	MimeType  string    `json:"mime"`
	Size      int64     `json:"size"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Storage combines all storage primitives.
type Storage struct {
	KV     KVStore
	Docs   DocStore
	Blobs  BlobStore
	db     *sql.DB
	writer *WriteQueue
}

// New creates a new Storage instance with all primitives.
// Uses the global WriteQueue (call InitWriter first).
func New(db *sql.DB) *Storage {
	// Use global writer if available, otherwise create local one
	writer := globalWriter
	if writer == nil {
		writer = NewWriteQueue(DefaultWriteQueueConfig())
	}
	return &Storage{
		KV:     NewSQLKVStoreWithWriter(db, writer),
		Docs:   NewSQLDocStoreWithWriter(db, writer),
		Blobs:  NewSQLBlobStoreWithWriter(db, writer),
		db:     db,
		writer: writer,
	}
}

// WriteStats returns current write queue statistics.
func (s *Storage) WriteStats() WriteStats {
	if s.writer == nil {
		return WriteStats{}
	}
	return s.writer.Stats()
}

// Close releases storage resources.
func (s *Storage) Close() {
	if s.writer != nil {
		s.writer.Close()
	}
	if kv, ok := s.KV.(*SQLKVStore); ok {
		kv.Close()
	}
}
