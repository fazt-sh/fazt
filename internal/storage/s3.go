package storage

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// SQLBlobStore implements BlobStore using SQLite.
type SQLBlobStore struct {
	db     *sql.DB
	writer *WriteQueue
}

// NewSQLBlobStore creates a new SQLite-backed blob store.
func NewSQLBlobStore(db *sql.DB) *SQLBlobStore {
	return NewSQLBlobStoreWithWriter(db, nil)
}

// NewSQLBlobStoreWithWriter creates a blob store with an optional write queue.
func NewSQLBlobStoreWithWriter(db *sql.DB, writer *WriteQueue) *SQLBlobStore {
	return &SQLBlobStore{db: db, writer: writer}
}

// Put stores a blob.
func (s *SQLBlobStore) Put(ctx context.Context, appID, path string, data []byte, mimeType string) error {
	// Normalize path
	path = normalizePath(path)

	// Calculate hash
	hash := sha256Hash(data)

	query := `
		INSERT INTO app_blobs (app_id, path, data, mime_type, size_bytes, hash, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, strftime('%s', 'now'))
		ON CONFLICT(app_id, path) DO UPDATE SET
			data = excluded.data,
			mime_type = excluded.mime_type,
			size_bytes = excluded.size_bytes,
			hash = excluded.hash,
			updated_at = strftime('%s', 'now')
	`

	writeOp := func() error {
		return withRetry(ctx, func() error {
			_, err := s.db.ExecContext(ctx, query, appID, path, data, mimeType, len(data), hash)
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
func (s *SQLBlobStore) Get(ctx context.Context, appID, path string) (*Blob, error) {
	path = normalizePath(path)

	query := `
		SELECT data, mime_type, size_bytes, hash FROM app_blobs
		WHERE app_id = ? AND path = ?
	`
	var data []byte
	var mimeType, hash string
	var size int64

	err := withRetry(ctx, func() error {
		return s.db.QueryRowContext(ctx, query, appID, path).Scan(&data, &mimeType, &size, &hash)
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
func (s *SQLBlobStore) Delete(ctx context.Context, appID, path string) error {
	path = normalizePath(path)

	query := `DELETE FROM app_blobs WHERE app_id = ? AND path = ?`

	writeOp := func() error {
		return withRetry(ctx, func() error {
			_, err := s.db.ExecContext(ctx, query, appID, path)
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
func (s *SQLBlobStore) List(ctx context.Context, appID, prefix string) ([]BlobMeta, error) {
	prefix = normalizePath(prefix)

	query := `
		SELECT path, mime_type, size_bytes, updated_at FROM app_blobs
		WHERE app_id = ? AND path LIKE ?
		ORDER BY path
	`
	var rows *sql.Rows
	err := withRetry(ctx, func() error {
		var err error
		rows, err = s.db.QueryContext(ctx, query, appID, prefix+"%")
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list blobs: %w", err)
	}
	defer rows.Close()

	var blobs []BlobMeta
	for rows.Next() {
		var path, mimeType string
		var size, updatedAt int64
		if err := rows.Scan(&path, &mimeType, &size, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		blobs = append(blobs, BlobMeta{
			Path:      path,
			MimeType:  mimeType,
			Size:      size,
			UpdatedAt: time.Unix(updatedAt, 0),
		})
	}

	return blobs, nil
}

// Exists checks if a blob exists.
func (s *SQLBlobStore) Exists(ctx context.Context, appID, path string) (bool, error) {
	path = normalizePath(path)

	query := `SELECT 1 FROM app_blobs WHERE app_id = ? AND path = ? LIMIT 1`
	var exists int
	err := withRetry(ctx, func() error {
		return s.db.QueryRowContext(ctx, query, appID, path).Scan(&exists)
	})
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// GetMeta retrieves blob metadata without the data.
func (s *SQLBlobStore) GetMeta(ctx context.Context, appID, path string) (*BlobMeta, error) {
	path = normalizePath(path)

	query := `
		SELECT path, mime_type, size_bytes, updated_at FROM app_blobs
		WHERE app_id = ? AND path = ?
	`
	var blobPath, mimeType string
	var size, updatedAt int64

	err := withRetry(ctx, func() error {
		return s.db.QueryRowContext(ctx, query, appID, path).Scan(&blobPath, &mimeType, &size, &updatedAt)
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get blob metadata: %w", err)
	}

	return &BlobMeta{
		Path:      blobPath,
		MimeType:  mimeType,
		Size:      size,
		UpdatedAt: time.Unix(updatedAt, 0),
	}, nil
}

// Copy copies a blob to a new path.
func (s *SQLBlobStore) Copy(ctx context.Context, appID, srcPath, dstPath string) error {
	srcPath = normalizePath(srcPath)
	dstPath = normalizePath(dstPath)

	query := `
		INSERT INTO app_blobs (app_id, path, data, mime_type, size_bytes, hash, created_at, updated_at)
		SELECT app_id, ?, data, mime_type, size_bytes, hash, strftime('%s', 'now'), strftime('%s', 'now')
		FROM app_blobs
		WHERE app_id = ? AND path = ?
		ON CONFLICT(app_id, path) DO UPDATE SET
			data = excluded.data,
			mime_type = excluded.mime_type,
			size_bytes = excluded.size_bytes,
			hash = excluded.hash,
			updated_at = strftime('%s', 'now')
	`
	var result sql.Result
	writeOp := func() error {
		return withRetry(ctx, func() error {
			var err error
			result, err = s.db.ExecContext(ctx, query, dstPath, appID, srcPath)
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
		return fmt.Errorf("failed to copy blob: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("source blob not found: %s", srcPath)
	}

	return nil
}

// Move moves a blob to a new path.
func (s *SQLBlobStore) Move(ctx context.Context, appID, srcPath, dstPath string) error {
	if err := s.Copy(ctx, appID, srcPath, dstPath); err != nil {
		return err
	}
	return s.Delete(ctx, appID, srcPath)
}

// TotalSize returns the total size of all blobs for an app.
func (s *SQLBlobStore) TotalSize(ctx context.Context, appID string) (int64, error) {
	query := `SELECT COALESCE(SUM(size_bytes), 0) FROM app_blobs WHERE app_id = ?`
	var total int64
	err := withRetry(ctx, func() error {
		return s.db.QueryRowContext(ctx, query, appID).Scan(&total)
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get total size: %w", err)
	}
	return total, nil
}

// normalizePath normalizes a blob path.
func normalizePath(path string) string {
	// Remove leading slash
	path = strings.TrimPrefix(path, "/")
	// Remove duplicate slashes
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}
	return path
}

// sha256Hash computes the SHA256 hash of data.
func sha256Hash(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
