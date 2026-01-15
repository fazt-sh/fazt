// Package storage provides app-scoped storage primitives for fazt serverless functions.
// It implements key-value, document, and blob storage backed by SQLite.
package storage

import (
	"context"
	"database/sql"
	"time"
)

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
	KV    KVStore
	Docs  DocStore
	Blobs BlobStore
	db    *sql.DB
}

// New creates a new Storage instance with all primitives.
func New(db *sql.DB) *Storage {
	return &Storage{
		KV:    NewSQLKVStore(db),
		Docs:  NewSQLDocStore(db),
		Blobs: NewSQLBlobStore(db),
		db:    db,
	}
}
