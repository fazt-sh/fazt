package hosting

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// FileSystem defines the interface for site storage
type FileSystem interface {
	WriteFile(siteID, path string, content io.Reader, size int64, mimeType string) error
	ReadFile(siteID, path string) (*File, error)
	DeleteSite(siteID string) error
	Exists(siteID, path string) (bool, error)
}

// File represents a file in the VFS
type File struct {
	Content  io.ReadCloser
	Size     int64
	MimeType string
	Hash     string
	ModTime  time.Time
}

// CachedFile holds file data in memory
type CachedFile struct {
	Data     []byte
	Size     int64
	MimeType string
	Hash     string
	ModTime  time.Time
}

// SQLFileSystem implements FileSystem using SQLite with in-memory caching
type SQLFileSystem struct {
	db      *sql.DB
	cache   map[string]CachedFile
	cacheMu sync.RWMutex
}

// NewSQLFileSystem creates a new SQL-backed file system
func NewSQLFileSystem(db *sql.DB) *SQLFileSystem {
	return &SQLFileSystem{
		db:    db,
		cache: make(map[string]CachedFile),
	}
}

// cacheKey generates a unique key for the cache
func cacheKey(siteID, path string) string {
	return siteID + ":" + path
}

// WriteFile writes a file to the database
func (fs *SQLFileSystem) WriteFile(siteID, path string, content io.Reader, size int64, mimeType string) error {
	// Read content to calculate hash and prepare for blob
	data, err := io.ReadAll(content)
	if err != nil {
		return fmt.Errorf("failed to read content: %w", err)
	}

	// Calculate SHA256 hash
	hash := sha256.Sum256(data)
	hashStr := hex.EncodeToString(hash[:])

	// Insert or Replace
	query := `
		INSERT INTO files (site_id, path, content, size_bytes, mime_type, hash, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(site_id, path) DO UPDATE SET
			content = excluded.content,
			size_bytes = excluded.size_bytes,
			mime_type = excluded.mime_type,
			hash = excluded.hash,
			updated_at = CURRENT_TIMESTAMP
	`
	
	_, err = fs.db.Exec(query, siteID, path, data, size, mimeType, hashStr)
	if err != nil {
		return fmt.Errorf("failed to write file to DB: %w", err)
	}

	// Invalidate cache
	fs.cacheMu.Lock()
	delete(fs.cache, cacheKey(siteID, path))
	fs.cacheMu.Unlock()

	return nil
}

// ReadFile reads a file from the database (or cache)
func (fs *SQLFileSystem) ReadFile(siteID, path string) (*File, error) {
	key := cacheKey(siteID, path)

	// Check cache
	fs.cacheMu.RLock()
	if cached, ok := fs.cache[key]; ok {
		fs.cacheMu.RUnlock()
		return &File{
			Content:  io.NopCloser(newByteReader(cached.Data)),
			Size:     cached.Size,
			MimeType: cached.MimeType,
			Hash:     cached.Hash,
			ModTime:  cached.ModTime,
		}, nil
	}
	fs.cacheMu.RUnlock()

	// Query DB
	query := `
		SELECT content, size_bytes, mime_type, hash, updated_at
		FROM files WHERE site_id = ? AND path = ?
	`
	
	var data []byte
	var size int64
	var mimeType, hash string
	var modTime time.Time

	err := fs.db.QueryRow(query, siteID, path).Scan(&data, &size, &mimeType, &hash, &modTime)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("file not found")
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Update cache
	fs.cacheMu.Lock()
	// Simple eviction policy: if cache too big, clear it
	if len(fs.cache) > 1000 {
		fs.cache = make(map[string]CachedFile)
	}
	fs.cache[key] = CachedFile{
		Data:     data,
		Size:     size,
		MimeType: mimeType,
		Hash:     hash,
		ModTime:  modTime,
	}
	fs.cacheMu.Unlock()

	return &File{
		Content:  io.NopCloser(newByteReader(data)),
		Size:     size,
		MimeType: mimeType,
		Hash:     hash,
		ModTime:  modTime,
	}, nil
}

// DeleteSite deletes all files for a site
func (fs *SQLFileSystem) DeleteSite(siteID string) error {
	_, err := fs.db.Exec("DELETE FROM files WHERE site_id = ?", siteID)
	
	// Invalidate all files for this site in cache
	fs.cacheMu.Lock()
	// Since we can't efficiently search by prefix in map, we iterate
	// optimization: if cache is huge, maybe just clear it all?
	// For now, iterate is okay for <1000 items
	for k := range fs.cache {
		if strings.HasPrefix(k, siteID+":") {
			delete(fs.cache, k)
		}
	}
	fs.cacheMu.Unlock()
	
	return err
}

// Exists checks if a file exists
func (fs *SQLFileSystem) Exists(siteID, path string) (bool, error) {
	// Check cache first
	key := cacheKey(siteID, path)
	fs.cacheMu.RLock()
	if _, ok := fs.cache[key]; ok {
		fs.cacheMu.RUnlock()
		return true, nil
	}
	fs.cacheMu.RUnlock()

	var count int
	err := fs.db.QueryRow("SELECT COUNT(*) FROM files WHERE site_id = ? AND path = ?", siteID, path).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Helper for byte reader
type byteReader struct {
	data []byte
	pos  int
}

func newByteReader(data []byte) *byteReader {
	return &byteReader{data: data}
}

func (r *byteReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
