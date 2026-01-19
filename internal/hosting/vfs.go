package hosting

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
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
	ListFiles(siteID string) ([]FileEntry, error)
	EnsureApp(name string, source *SourceInfo) error
	GetAppSource(name string) (*SourceInfo, error)
}

// File represents a file in the VFS
type File struct {
	Content  io.ReadCloser
	Size     int64
	MimeType string
	Hash     string
	ModTime  time.Time
}

// FileEntry represents a file in a listing
type FileEntry struct {
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
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

// VFSStats holds file system statistics
type VFSStats struct {
	CachedFiles    int
	CacheSizeBytes int64
}

// GetStats returns VFS statistics
func (fs *SQLFileSystem) GetStats() VFSStats {
	fs.cacheMu.RLock()
	defer fs.cacheMu.RUnlock()

	var size int64
	for _, f := range fs.cache {
		size += f.Size
	}

	return VFSStats{
		CachedFiles:    len(fs.cache),
		CacheSizeBytes: size,
	}
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

// ListFiles returns a list of files for a site
func (fs *SQLFileSystem) ListFiles(siteID string) ([]FileEntry, error) {
	query := `
		SELECT path, size_bytes, updated_at
		FROM files WHERE site_id = ?
		ORDER BY path
	`
	
	rows, err := fs.db.Query(query, siteID)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var files []FileEntry
	for rows.Next() {
		var f FileEntry
		if err := rows.Scan(&f.Path, &f.Size, &f.ModTime); err != nil {
			return nil, err
		}
		files = append(files, f)
	}

	return files, nil
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

// SourceInfo contains source tracking information for an app
type SourceInfo struct {
	Type   string // "deploy" or "git"
	URL    string // git URL (for git-sourced apps)
	Ref    string // git ref (tag, branch, or commit)
	Commit string // resolved commit SHA
}

// EnsureApp creates or updates an app entry in the apps table
func (fs *SQLFileSystem) EnsureApp(name string, source *SourceInfo) error {
	sourceType := "deploy"
	var sourceURL, sourceRef, sourceCommit *string

	if source != nil {
		sourceType = source.Type
		if source.URL != "" {
			sourceURL = &source.URL
		}
		if source.Ref != "" {
			sourceRef = &source.Ref
		}
		if source.Commit != "" {
			sourceCommit = &source.Commit
		}
	}

	// Check if app exists by title (via alias lookup)
	var existingID string
	err := fs.db.QueryRow(`SELECT id FROM apps WHERE title = ?`, name).Scan(&existingID)

	if err == sql.ErrNoRows {
		// Create new app with generated ID
		appID := generateAppID()
		query := `
			INSERT INTO apps (id, title, source, source_url, source_ref, source_commit, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`
		_, err = fs.db.Exec(query, appID, name, sourceType, sourceURL, sourceRef, sourceCommit)
		if err != nil {
			return fmt.Errorf("failed to create app entry: %w", err)
		}

		// Create alias pointing to this app
		aliasQuery := `
			INSERT OR IGNORE INTO aliases (subdomain, type, targets, created_at, updated_at)
			VALUES (?, 'app', json_object('app_id', ?), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`
		fs.db.Exec(aliasQuery, name, appID)
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to lookup app: %w", err)
	}

	// Update existing app
	query := `
		UPDATE apps SET
			source = ?,
			source_url = COALESCE(?, source_url),
			source_ref = COALESCE(?, source_ref),
			source_commit = COALESCE(?, source_commit),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err = fs.db.Exec(query, sourceType, sourceURL, sourceRef, sourceCommit, existingID)
	return err
}

// generateAppID creates a unique app ID like "app_7f3k9x2m"
func generateAppID() string {
	const charset = "0123456789abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[randInt(len(charset))]
	}
	return "app_" + string(b)
}

// randInt returns a random int in [0, max)
func randInt(max int) int {
	var b [1]byte
	rand.Read(b[:])
	return int(b[0]) % max
}

// GetAppSource returns source tracking info for an app
func (fs *SQLFileSystem) GetAppSource(name string) (*SourceInfo, error) {
	query := `
		SELECT source, source_url, source_ref, source_commit
		FROM apps WHERE id = ? OR title = ?
	`

	var sourceType string
	var sourceURL, sourceRef, sourceCommit sql.NullString

	err := fs.db.QueryRow(query, name, name).Scan(&sourceType, &sourceURL, &sourceRef, &sourceCommit)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("app not found: %s", name)
	}
	if err != nil {
		return nil, err
	}

	return &SourceInfo{
		Type:   sourceType,
		URL:    sourceURL.String,
		Ref:    sourceRef.String,
		Commit: sourceCommit.String,
	}, nil
}

// ReadFileByAppID reads a file using app_id instead of site_id
func (fs *SQLFileSystem) ReadFileByAppID(appID, path string) (*File, error) {
	key := cacheKey(appID, path)

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

	// Query DB using app_id
	query := `
		SELECT content, size_bytes, mime_type, hash, updated_at
		FROM files WHERE app_id = ? AND path = ?
	`

	var data []byte
	var size int64
	var mimeType, hash string
	var modTime time.Time

	err := fs.db.QueryRow(query, appID, path).Scan(&data, &size, &mimeType, &hash, &modTime)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("file not found")
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Update cache
	fs.cacheMu.Lock()
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

// ListFilesByAppID returns files for an app using app_id
func (fs *SQLFileSystem) ListFilesByAppID(appID string) ([]FileEntry, error) {
	query := `
		SELECT path, size_bytes, updated_at
		FROM files WHERE app_id = ?
		ORDER BY path
	`

	rows, err := fs.db.Query(query, appID)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var files []FileEntry
	for rows.Next() {
		var f FileEntry
		if err := rows.Scan(&f.Path, &f.Size, &f.ModTime); err != nil {
			return nil, err
		}
		files = append(files, f)
	}

	return files, nil
}

// DeleteAppFiles deletes all files for an app by app_id
func (fs *SQLFileSystem) DeleteAppFiles(appID string) error {
	_, err := fs.db.Exec("DELETE FROM files WHERE app_id = ?", appID)

	// Invalidate cache
	fs.cacheMu.Lock()
	for k := range fs.cache {
		if strings.HasPrefix(k, appID+":") {
			delete(fs.cache, k)
		}
	}
	fs.cacheMu.Unlock()

	return err
}

// ExistsByAppID checks if a file exists using app_id
func (fs *SQLFileSystem) ExistsByAppID(appID, path string) (bool, error) {
	// Check cache first
	key := cacheKey(appID, path)
	fs.cacheMu.RLock()
	if _, ok := fs.cache[key]; ok {
		fs.cacheMu.RUnlock()
		return true, nil
	}
	fs.cacheMu.RUnlock()

	// Query DB using app_id
	var count int
	err := fs.db.QueryRow("SELECT COUNT(*) FROM files WHERE app_id = ? AND path = ?", appID, path).Scan(&count)
	return count > 0, err
}

// GetMimeType returns the MIME type for a file path
func GetMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	mimeTypes := map[string]string{
		".html": "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".json": "application/json",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".svg":  "image/svg+xml",
		".ico":  "image/x-icon",
		".woff": "font/woff",
		".woff2": "font/woff2",
		".ttf":  "font/ttf",
		".eot":  "application/vnd.ms-fontobject",
		".txt":  "text/plain",
		".xml":  "application/xml",
		".pdf":  "application/pdf",
		".zip":  "application/zip",
		".wasm": "application/wasm",
	}
	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}
	return "application/octet-stream"
}
