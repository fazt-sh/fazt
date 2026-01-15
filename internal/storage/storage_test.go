package storage

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// setupTestDB creates a temporary SQLite database for testing.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// Create temp file for database
	tmpFile, err := os.CreateTemp("", "fazt_storage_test_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFile.Close()

	// Register cleanup
	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})

	// Open database
	db, err := sql.Open("sqlite", tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	// Create tables
	schema := `
		CREATE TABLE IF NOT EXISTS app_kv (
			app_id TEXT NOT NULL,
			key TEXT NOT NULL,
			value TEXT,
			expires_at INTEGER,
			created_at INTEGER DEFAULT (strftime('%s', 'now')),
			updated_at INTEGER DEFAULT (strftime('%s', 'now')),
			PRIMARY KEY (app_id, key)
		);
		CREATE TABLE IF NOT EXISTS app_docs (
			app_id TEXT NOT NULL,
			collection TEXT NOT NULL,
			id TEXT NOT NULL,
			data TEXT NOT NULL,
			created_at INTEGER DEFAULT (strftime('%s', 'now')),
			updated_at INTEGER DEFAULT (strftime('%s', 'now')),
			PRIMARY KEY (app_id, collection, id)
		);
		CREATE TABLE IF NOT EXISTS app_blobs (
			app_id TEXT NOT NULL,
			path TEXT NOT NULL,
			data BLOB NOT NULL,
			mime_type TEXT NOT NULL,
			size_bytes INTEGER NOT NULL,
			hash TEXT NOT NULL,
			created_at INTEGER DEFAULT (strftime('%s', 'now')),
			updated_at INTEGER DEFAULT (strftime('%s', 'now')),
			PRIMARY KEY (app_id, path)
		);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

// TestKVStore tests the key-value store.
func TestKVStore(t *testing.T) {
	db := setupTestDB(t)
	kv := NewSQLKVStore(db)
	defer kv.Close()
	ctx := context.Background()
	appID := "test-app"

	t.Run("SetAndGet", func(t *testing.T) {
		err := kv.Set(ctx, appID, "key1", "value1", nil)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		val, err := kv.Get(ctx, appID, "key1")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if val != "value1" {
			t.Errorf("expected 'value1', got %v", val)
		}
	})

	t.Run("SetAndGetJSON", func(t *testing.T) {
		data := map[string]interface{}{
			"name": "Alice",
			"age":  float64(30),
		}
		err := kv.Set(ctx, appID, "json-key", data, nil)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		val, err := kv.Get(ctx, appID, "json-key")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		m, ok := val.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map, got %T", val)
		}
		if m["name"] != "Alice" {
			t.Errorf("expected 'Alice', got %v", m["name"])
		}
	})

	t.Run("GetNonExistent", func(t *testing.T) {
		val, err := kv.Get(ctx, appID, "nonexistent")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if val != nil {
			t.Errorf("expected nil, got %v", val)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		err := kv.Set(ctx, appID, "to-delete", "value", nil)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		err = kv.Delete(ctx, appID, "to-delete")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		val, err := kv.Get(ctx, appID, "to-delete")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if val != nil {
			t.Errorf("expected nil after delete, got %v", val)
		}
	})

	t.Run("List", func(t *testing.T) {
		// Set some keys with prefix
		kv.Set(ctx, appID, "prefix:a", "1", nil)
		kv.Set(ctx, appID, "prefix:b", "2", nil)
		kv.Set(ctx, appID, "other:c", "3", nil)

		entries, err := kv.List(ctx, appID, "prefix:")
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(entries) != 2 {
			t.Errorf("expected 2 entries, got %d", len(entries))
		}
	})

	t.Run("TTL", func(t *testing.T) {
		// TTL has second-level granularity due to Unix timestamp storage
		ttl := 2 * time.Second
		err := kv.Set(ctx, appID, "ttl-key", "value", &ttl)
		if err != nil {
			t.Fatalf("Set with TTL failed: %v", err)
		}

		// Should exist immediately
		val, err := kv.Get(ctx, appID, "ttl-key")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if val == nil {
			t.Error("expected value, got nil")
		}

		// Wait for expiration (2s + buffer)
		time.Sleep(3 * time.Second)

		// Should be expired now
		val, err = kv.Get(ctx, appID, "ttl-key")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if val != nil {
			t.Errorf("expected nil after TTL, got %v", val)
		}
	})

	t.Run("AppIsolation", func(t *testing.T) {
		err := kv.Set(ctx, "app1", "shared-key", "app1-value", nil)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}
		err = kv.Set(ctx, "app2", "shared-key", "app2-value", nil)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		val1, _ := kv.Get(ctx, "app1", "shared-key")
		val2, _ := kv.Get(ctx, "app2", "shared-key")

		if val1 != "app1-value" {
			t.Errorf("app1 expected 'app1-value', got %v", val1)
		}
		if val2 != "app2-value" {
			t.Errorf("app2 expected 'app2-value', got %v", val2)
		}
	})
}

// TestDocStore tests the document store.
func TestDocStore(t *testing.T) {
	db := setupTestDB(t)
	ds := NewSQLDocStore(db)
	ctx := context.Background()
	appID := "test-app"
	collection := "users"

	t.Run("Insert", func(t *testing.T) {
		doc := map[string]interface{}{
			"email": "alice@example.com",
			"name":  "Alice",
		}
		id, err := ds.Insert(ctx, appID, collection, doc)
		if err != nil {
			t.Fatalf("Insert failed: %v", err)
		}
		if id == "" {
			t.Error("expected non-empty ID")
		}
	})

	t.Run("FindOne", func(t *testing.T) {
		doc := map[string]interface{}{
			"id":    "user-123",
			"email": "bob@example.com",
			"name":  "Bob",
		}
		_, err := ds.Insert(ctx, appID, collection, doc)
		if err != nil {
			t.Fatalf("Insert failed: %v", err)
		}

		found, err := ds.FindOne(ctx, appID, collection, "user-123")
		if err != nil {
			t.Fatalf("FindOne failed: %v", err)
		}
		if found == nil {
			t.Fatal("expected document, got nil")
		}
		if found.Data["name"] != "Bob" {
			t.Errorf("expected 'Bob', got %v", found.Data["name"])
		}
	})

	t.Run("Find", func(t *testing.T) {
		// Insert some documents
		ds.Insert(ctx, appID, "products", map[string]interface{}{"name": "Widget", "price": float64(10)})
		ds.Insert(ctx, appID, "products", map[string]interface{}{"name": "Gadget", "price": float64(20)})
		ds.Insert(ctx, appID, "products", map[string]interface{}{"name": "Doohickey", "price": float64(30)})

		// Find all
		docs, err := ds.Find(ctx, appID, "products", map[string]interface{}{})
		if err != nil {
			t.Fatalf("Find failed: %v", err)
		}
		if len(docs) != 3 {
			t.Errorf("expected 3 documents, got %d", len(docs))
		}
	})

	t.Run("FindWithQuery", func(t *testing.T) {
		// Find with equality
		docs, err := ds.Find(ctx, appID, "products", map[string]interface{}{
			"name": "Widget",
		})
		if err != nil {
			t.Fatalf("Find failed: %v", err)
		}
		if len(docs) != 1 {
			t.Errorf("expected 1 document, got %d", len(docs))
		}
	})

	t.Run("FindWithOperator", func(t *testing.T) {
		// Find with $gt operator
		docs, err := ds.Find(ctx, appID, "products", map[string]interface{}{
			"price": map[string]interface{}{
				"$gt": float64(15),
			},
		})
		if err != nil {
			t.Fatalf("Find failed: %v", err)
		}
		if len(docs) != 2 {
			t.Errorf("expected 2 documents (price > 15), got %d", len(docs))
		}
	})

	t.Run("Delete", func(t *testing.T) {
		ds.Insert(ctx, appID, "temp", map[string]interface{}{"name": "to-delete", "type": "temp"})

		count, err := ds.Delete(ctx, appID, "temp", map[string]interface{}{"name": "to-delete"})
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
		if count != 1 {
			t.Errorf("expected 1 deleted, got %d", count)
		}
	})
}

// TestBlobStore tests the blob store.
func TestBlobStore(t *testing.T) {
	db := setupTestDB(t)
	blobs := NewSQLBlobStore(db)
	ctx := context.Background()
	appID := "test-app"

	t.Run("PutAndGet", func(t *testing.T) {
		data := []byte("Hello, World!")
		err := blobs.Put(ctx, appID, "test.txt", data, "text/plain")
		if err != nil {
			t.Fatalf("Put failed: %v", err)
		}

		blob, err := blobs.Get(ctx, appID, "test.txt")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if blob == nil {
			t.Fatal("expected blob, got nil")
		}
		if string(blob.Data) != "Hello, World!" {
			t.Errorf("expected 'Hello, World!', got %s", string(blob.Data))
		}
		if blob.MimeType != "text/plain" {
			t.Errorf("expected 'text/plain', got %s", blob.MimeType)
		}
	})

	t.Run("GetNonExistent", func(t *testing.T) {
		blob, err := blobs.Get(ctx, appID, "nonexistent.txt")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if blob != nil {
			t.Error("expected nil for nonexistent blob")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		blobs.Put(ctx, appID, "to-delete.txt", []byte("data"), "text/plain")

		err := blobs.Delete(ctx, appID, "to-delete.txt")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		blob, _ := blobs.Get(ctx, appID, "to-delete.txt")
		if blob != nil {
			t.Error("expected nil after delete")
		}
	})

	t.Run("List", func(t *testing.T) {
		blobs.Put(ctx, appID, "uploads/a.txt", []byte("a"), "text/plain")
		blobs.Put(ctx, appID, "uploads/b.txt", []byte("b"), "text/plain")
		blobs.Put(ctx, appID, "other/c.txt", []byte("c"), "text/plain")

		items, err := blobs.List(ctx, appID, "uploads/")
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(items) != 2 {
			t.Errorf("expected 2 items, got %d", len(items))
		}
	})

	t.Run("Hash", func(t *testing.T) {
		data := []byte("test data for hashing")
		blobs.Put(ctx, appID, "hash-test.txt", data, "text/plain")

		blob, _ := blobs.Get(ctx, appID, "hash-test.txt")
		if blob.Hash == "" {
			t.Error("expected non-empty hash")
		}
	})
}
