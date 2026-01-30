// Package storage provides user-scoped storage wrappers.
// These wrap the base storage implementations to add automatic user isolation.
package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// UserScopedKV wraps SQLKVStore to provide user-isolated key-value storage.
// All operations are scoped to (app_id, user_id).
type UserScopedKV struct {
	db     *sql.DB
	writer *WriteQueue
	appID  string
	userID string
}

// NewUserScopedKV creates a user-scoped KV store.
func NewUserScopedKV(db *sql.DB, writer *WriteQueue, appID, userID string) *UserScopedKV {
	return &UserScopedKV{
		db:     db,
		writer: writer,
		appID:  appID,
		userID: userID,
	}
}

// Set stores a value with optional TTL.
// Uses prefixed key to ensure user isolation within existing schema.
func (s *UserScopedKV) Set(ctx context.Context, key string, value interface{}, ttl *time.Duration) error {
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	var expiresAt *int64
	if ttl != nil {
		exp := time.Now().Add(*ttl).Unix()
		expiresAt = &exp
	}

	// Prefix key with user_id for isolation within existing schema
	scopedKey := s.scopeKey(key)

	query := `
		INSERT INTO app_kv (app_id, user_id, key, value, expires_at, updated_at)
		VALUES (?, ?, ?, ?, ?, strftime('%s', 'now'))
		ON CONFLICT(app_id, key) DO UPDATE SET
			value = excluded.value,
			expires_at = excluded.expires_at,
			updated_at = strftime('%s', 'now')
	`

	writeOp := func() error {
		return withRetry(ctx, func() error {
			_, err := s.db.ExecContext(ctx, query, s.appID, s.userID, scopedKey, string(valueJSON), expiresAt)
			return err
		})
	}

	if s.writer != nil {
		err = s.writer.Write(ctx, writeOp)
	} else {
		err = writeOp()
	}
	if err != nil {
		return fmt.Errorf("failed to set key: %w", err)
	}

	return nil
}

// scopeKey prefixes a key with user_id for isolation
func (s *UserScopedKV) scopeKey(key string) string {
	return "u:" + s.userID + ":" + key
}

// Get retrieves a value by key.
func (s *UserScopedKV) Get(ctx context.Context, key string) (interface{}, error) {
	scopedKey := s.scopeKey(key)
	query := `
		SELECT value FROM app_kv
		WHERE app_id = ? AND key = ?
		AND (expires_at IS NULL OR expires_at > strftime('%s', 'now'))
	`
	var valueJSON string
	err := withRetry(ctx, func() error {
		return s.db.QueryRowContext(ctx, query, s.appID, scopedKey).Scan(&valueJSON)
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	var value interface{}
	if err := json.Unmarshal([]byte(valueJSON), &value); err != nil {
		return nil, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return value, nil
}

// Delete removes a key.
func (s *UserScopedKV) Delete(ctx context.Context, key string) error {
	scopedKey := s.scopeKey(key)
	query := `DELETE FROM app_kv WHERE app_id = ? AND key = ?`

	writeOp := func() error {
		return withRetry(ctx, func() error {
			_, err := s.db.ExecContext(ctx, query, s.appID, scopedKey)
			return err
		})
	}

	var err error
	if s.writer != nil {
		err = s.writer.Write(ctx, writeOp)
	} else {
		err = writeOp()
	}
	if err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}

	return nil
}

// List returns all keys matching a prefix.
func (s *UserScopedKV) List(ctx context.Context, prefix string) ([]KVEntry, error) {
	scopedPrefix := s.scopeKey(prefix)
	query := `
		SELECT key, value, expires_at FROM app_kv
		WHERE app_id = ? AND key LIKE ?
		AND (expires_at IS NULL OR expires_at > strftime('%s', 'now'))
		ORDER BY key
	`
	var rows *sql.Rows
	err := withRetry(ctx, func() error {
		var err error
		rows, err = s.db.QueryContext(ctx, query, s.appID, scopedPrefix+"%")
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}
	defer rows.Close()

	// Prefix to strip from returned keys
	keyPrefix := "u:" + s.userID + ":"

	var entries []KVEntry
	for rows.Next() {
		var key, valueJSON string
		var expiresAt sql.NullInt64
		if err := rows.Scan(&key, &valueJSON, &expiresAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		var value interface{}
		if err := json.Unmarshal([]byte(valueJSON), &value); err != nil {
			return nil, fmt.Errorf("failed to unmarshal value: %w", err)
		}

		// Strip user prefix from key
		displayKey := key
		if len(key) > len(keyPrefix) {
			displayKey = key[len(keyPrefix):]
		}

		entry := KVEntry{Key: displayKey, Value: value}
		if expiresAt.Valid {
			t := time.Unix(expiresAt.Int64, 0)
			entry.ExpiresAt = &t
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// UserScopedDocs wraps SQLDocStore to provide user-isolated document storage.
type UserScopedDocs struct {
	db     *sql.DB
	writer *WriteQueue
	appID  string
	userID string
}

// NewUserScopedDocs creates a user-scoped document store.
func NewUserScopedDocs(db *sql.DB, writer *WriteQueue, appID, userID string) *UserScopedDocs {
	return &UserScopedDocs{
		db:     db,
		writer: writer,
		appID:  appID,
		userID: userID,
	}
}

// scopeCollection prefixes collection with user_id for isolation
func (s *UserScopedDocs) scopeCollection(collection string) string {
	return "u:" + s.userID + ":" + collection
}

// Insert adds a new document to a collection.
func (s *UserScopedDocs) Insert(ctx context.Context, collection string, doc map[string]interface{}) (string, error) {
	id, ok := doc["id"].(string)
	if !ok || id == "" {
		id = uuid.New().String()
	}

	docCopy := make(map[string]interface{})
	for k, v := range doc {
		if k != "id" {
			docCopy[k] = v
		}
	}

	dataJSON, err := json.Marshal(docCopy)
	if err != nil {
		return "", fmt.Errorf("failed to marshal document: %w", err)
	}

	scopedCollection := s.scopeCollection(collection)
	query := `
		INSERT INTO app_docs (app_id, user_id, collection, id, data, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, strftime('%s', 'now'), strftime('%s', 'now'))
	`

	writeOp := func() error {
		return withRetry(ctx, func() error {
			_, err := s.db.ExecContext(ctx, query, s.appID, s.userID, scopedCollection, id, string(dataJSON))
			return err
		})
	}

	if s.writer != nil {
		err = s.writer.Write(ctx, writeOp)
	} else {
		err = writeOp()
	}
	if err != nil {
		return "", fmt.Errorf("failed to insert document: %w", err)
	}

	return id, nil
}

// Find retrieves documents matching a query.
func (s *UserScopedDocs) Find(ctx context.Context, collection string, query map[string]interface{}) ([]Document, error) {
	return s.FindWithOptions(ctx, collection, query, nil)
}

// FindWithOptions retrieves documents with pagination.
func (s *UserScopedDocs) FindWithOptions(ctx context.Context, collection string, query map[string]interface{}, opts *FindOptions) ([]Document, error) {
	qb := NewQueryBuilder()
	whereClause, args, err := qb.Build(query)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	scopedCollection := s.scopeCollection(collection)
	fullArgs := make([]interface{}, 0, len(args)+2)
	fullArgs = append(fullArgs, s.appID, scopedCollection)
	fullArgs = append(fullArgs, args...)

	order := "DESC"
	if opts != nil && opts.Order == "asc" {
		order = "ASC"
	}

	sqlQuery := fmt.Sprintf(`
		SELECT id, data, created_at, updated_at FROM app_docs
		WHERE app_id = ? AND collection = ? AND %s
		ORDER BY created_at %s
	`, whereClause, order)

	if opts != nil && opts.Limit > 0 {
		sqlQuery += fmt.Sprintf(" LIMIT %d", opts.Limit)
		if opts.Offset > 0 {
			sqlQuery += fmt.Sprintf(" OFFSET %d", opts.Offset)
		}
	}

	var rows *sql.Rows
	err = withRetry(ctx, func() error {
		var err error
		rows, err = s.db.QueryContext(ctx, sqlQuery, fullArgs...)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var id, dataJSON string
		var createdAt, updatedAt int64
		if err := rows.Scan(&id, &dataJSON, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(dataJSON), &data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal document: %w", err)
		}

		docs = append(docs, Document{
			ID:        id,
			Data:      data,
			CreatedAt: time.Unix(createdAt, 0),
			UpdatedAt: time.Unix(updatedAt, 0),
		})
	}

	return docs, nil
}

// FindOne retrieves a single document by query.
func (s *UserScopedDocs) FindOne(ctx context.Context, collection string, query map[string]interface{}) (*Document, error) {
	docs, err := s.FindWithOptions(ctx, collection, query, &FindOptions{Limit: 1})
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return nil, nil
	}
	return &docs[0], nil
}

// Update modifies documents matching a query.
func (s *UserScopedDocs) Update(ctx context.Context, collection string, query, changes map[string]interface{}) (int64, error) {
	qb := NewQueryBuilder()
	whereClause, whereArgs, err := qb.Build(query)
	if err != nil {
		return 0, fmt.Errorf("failed to build query: %w", err)
	}

	ub := NewUpdateBuilder()
	updateExpr, updateArgs, err := ub.Build("data", changes)
	if err != nil {
		return 0, fmt.Errorf("failed to build update: %w", err)
	}

	scopedCollection := s.scopeCollection(collection)
	allArgs := make([]interface{}, 0, len(updateArgs)+len(whereArgs)+2)
	allArgs = append(allArgs, updateArgs...)
	allArgs = append(allArgs, s.appID, scopedCollection)
	allArgs = append(allArgs, whereArgs...)

	sqlQuery := fmt.Sprintf(`
		UPDATE app_docs
		SET data = %s, updated_at = strftime('%%s', 'now')
		WHERE app_id = ? AND collection = ? AND %s
	`, updateExpr, whereClause)

	var result sql.Result
	writeOp := func() error {
		return withRetry(ctx, func() error {
			var err error
			result, err = s.db.ExecContext(ctx, sqlQuery, allArgs...)
			return err
		})
	}

	if s.writer != nil {
		err = s.writer.Write(ctx, writeOp)
	} else {
		err = writeOp()
	}
	if err != nil {
		return 0, fmt.Errorf("failed to update documents: %w", err)
	}

	return result.RowsAffected()
}

// Delete removes documents matching a query.
func (s *UserScopedDocs) Delete(ctx context.Context, collection string, query map[string]interface{}) (int64, error) {
	qb := NewQueryBuilder()
	whereClause, args, err := qb.Build(query)
	if err != nil {
		return 0, fmt.Errorf("failed to build query: %w", err)
	}

	scopedCollection := s.scopeCollection(collection)
	fullArgs := make([]interface{}, 0, len(args)+2)
	fullArgs = append(fullArgs, s.appID, scopedCollection)
	fullArgs = append(fullArgs, args...)

	sqlQuery := fmt.Sprintf(`
		DELETE FROM app_docs
		WHERE app_id = ? AND collection = ? AND %s
	`, whereClause)

	var result sql.Result
	writeOp := func() error {
		return withRetry(ctx, func() error {
			var err error
			result, err = s.db.ExecContext(ctx, sqlQuery, fullArgs...)
			return err
		})
	}

	if s.writer != nil {
		err = s.writer.Write(ctx, writeOp)
	} else {
		err = writeOp()
	}
	if err != nil {
		return 0, fmt.Errorf("failed to delete documents: %w", err)
	}

	return result.RowsAffected()
}

// Count returns the number of documents matching a query.
func (s *UserScopedDocs) Count(ctx context.Context, collection string, query map[string]interface{}) (int64, error) {
	qb := NewQueryBuilder()
	whereClause, args, err := qb.Build(query)
	if err != nil {
		return 0, fmt.Errorf("failed to build query: %w", err)
	}

	scopedCollection := s.scopeCollection(collection)
	fullArgs := make([]interface{}, 0, len(args)+2)
	fullArgs = append(fullArgs, s.appID, scopedCollection)
	fullArgs = append(fullArgs, args...)

	sqlQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM app_docs
		WHERE app_id = ? AND collection = ? AND %s
	`, whereClause)

	var count int64
	err = withRetry(ctx, func() error {
		return s.db.QueryRowContext(ctx, sqlQuery, fullArgs...).Scan(&count)
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	return count, nil
}

// UserScopedBlobs wraps SQLBlobStore to provide user-isolated blob storage.
type UserScopedBlobs struct {
	db     *sql.DB
	writer *WriteQueue
	appID  string
	userID string
}

// NewUserScopedBlobs creates a user-scoped blob store.
func NewUserScopedBlobs(db *sql.DB, writer *WriteQueue, appID, userID string) *UserScopedBlobs {
	return &UserScopedBlobs{
		db:     db,
		writer: writer,
		appID:  appID,
		userID: userID,
	}
}

// scopePath prefixes path with user_id for isolation
func (s *UserScopedBlobs) scopePath(path string) string {
	return "u/" + s.userID + "/" + normalizePath(path)
}

// Put stores a blob.
func (s *UserScopedBlobs) Put(ctx context.Context, path string, data []byte, mimeType string) error {
	scopedPath := s.scopePath(path)
	hash := sha256Hash(data)

	query := `
		INSERT INTO app_blobs (app_id, user_id, path, data, mime_type, size_bytes, hash, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, strftime('%s', 'now'))
		ON CONFLICT(app_id, path) DO UPDATE SET
			data = excluded.data,
			mime_type = excluded.mime_type,
			size_bytes = excluded.size_bytes,
			hash = excluded.hash,
			updated_at = strftime('%s', 'now')
	`

	writeOp := func() error {
		return withRetry(ctx, func() error {
			_, err := s.db.ExecContext(ctx, query, s.appID, s.userID, scopedPath, data, mimeType, len(data), hash)
			return err
		})
	}

	var err error
	if s.writer != nil {
		err = s.writer.Write(ctx, writeOp)
	} else {
		err = writeOp()
	}
	if err != nil {
		return fmt.Errorf("failed to store blob: %w", err)
	}

	return nil
}

// Get retrieves a blob by path.
func (s *UserScopedBlobs) Get(ctx context.Context, path string) (*Blob, error) {
	scopedPath := s.scopePath(path)

	query := `
		SELECT data, mime_type, size_bytes, hash FROM app_blobs
		WHERE app_id = ? AND path = ?
	`
	var data []byte
	var mimeType, hash string
	var size int64

	err := withRetry(ctx, func() error {
		return s.db.QueryRowContext(ctx, query, s.appID, scopedPath).Scan(&data, &mimeType, &size, &hash)
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get blob: %w", err)
	}

	return &Blob{
		Data:     data,
		MimeType: mimeType,
		Size:     size,
		Hash:     hash,
	}, nil
}

// Delete removes a blob.
func (s *UserScopedBlobs) Delete(ctx context.Context, path string) error {
	scopedPath := s.scopePath(path)

	query := `DELETE FROM app_blobs WHERE app_id = ? AND path = ?`

	writeOp := func() error {
		return withRetry(ctx, func() error {
			_, err := s.db.ExecContext(ctx, query, s.appID, scopedPath)
			return err
		})
	}

	var err error
	if s.writer != nil {
		err = s.writer.Write(ctx, writeOp)
	} else {
		err = writeOp()
	}
	if err != nil {
		return fmt.Errorf("failed to delete blob: %w", err)
	}

	return nil
}

// List returns metadata for blobs matching a prefix.
func (s *UserScopedBlobs) List(ctx context.Context, prefix string) ([]BlobMeta, error) {
	scopedPrefix := s.scopePath(prefix)

	query := `
		SELECT path, mime_type, size_bytes, updated_at FROM app_blobs
		WHERE app_id = ? AND path LIKE ?
		ORDER BY path
	`
	var rows *sql.Rows
	err := withRetry(ctx, func() error {
		var err error
		rows, err = s.db.QueryContext(ctx, query, s.appID, scopedPrefix+"%")
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list blobs: %w", err)
	}
	defer rows.Close()

	// Prefix to strip from returned paths
	pathPrefix := "u/" + s.userID + "/"

	var blobs []BlobMeta
	for rows.Next() {
		var path, mimeType string
		var size, updatedAt int64
		if err := rows.Scan(&path, &mimeType, &size, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Strip user prefix from path
		displayPath := path
		if len(path) > len(pathPrefix) {
			displayPath = path[len(pathPrefix):]
		}

		blobs = append(blobs, BlobMeta{
			Path:      displayPath,
			MimeType:  mimeType,
			Size:      size,
			UpdatedAt: time.Unix(updatedAt, 0),
		})
	}

	return blobs, nil
}
