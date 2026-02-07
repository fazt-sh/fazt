package media

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
)

const mediaCachePrefix = "_media/"

// MediaCache stores and retrieves processed image variants in app_blobs.
// Variants are stored under _media/{hash}/{transform_key} paths,
// optionally scoped to a user: u/{userId}/_media/{hash}/{transform_key}.
type MediaCache struct {
	db     *sql.DB
	userID string // empty = shared/app-level cache
}

// NewMediaCache creates a media cache for shared (app-level) blobs.
func NewMediaCache(db *sql.DB) *MediaCache {
	return &MediaCache{db: db}
}

// NewUserMediaCache creates a media cache scoped to a user.
// Cache entries live under u/{userId}/_media/... so deleting all user
// blobs (path LIKE 'u/{userId}/%') also removes cached variants.
func NewUserMediaCache(db *sql.DB, userID string) *MediaCache {
	return &MediaCache{db: db, userID: userID}
}

// Get retrieves a cached variant. Checks in-memory LRU first, then DB.
func (c *MediaCache) Get(ctx context.Context, appID, blobPath string, opts TransformOpts) ([]byte, string, error) {
	key := c.cacheKey(blobPath, opts)

	// Check in-memory cache first
	mc := getMemCache()
	if data, mime := mc.get(appID, key); data != nil {
		return data, mime, nil
	}

	// Fall through to DB
	var data []byte
	var mimeType string
	err := c.db.QueryRowContext(ctx,
		`SELECT data, mime_type FROM app_blobs WHERE app_id = ? AND path = ?`,
		appID, key,
	).Scan(&data, &mimeType)
	if err == sql.ErrNoRows {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}

	// Promote to in-memory cache
	mc.put(appID, key, data, mimeType)
	return data, mimeType, nil
}

// Put stores a processed variant in both DB and in-memory cache.
func (c *MediaCache) Put(ctx context.Context, appID, blobPath string, opts TransformOpts, data []byte, mimeType string) error {
	key := c.cacheKey(blobPath, opts)
	_, err := c.db.ExecContext(ctx,
		`INSERT INTO app_blobs (app_id, path, data, mime_type, size_bytes, hash, updated_at)
		 VALUES (?, ?, ?, ?, ?, '', strftime('%s', 'now'))
		 ON CONFLICT(app_id, path) DO UPDATE SET
		   data = excluded.data, mime_type = excluded.mime_type,
		   size_bytes = excluded.size_bytes, updated_at = strftime('%s', 'now')`,
		appID, key, data, mimeType, len(data),
	)
	if err == nil {
		getMemCache().put(appID, key, data, mimeType)
	}
	return err
}

// Invalidate deletes all cached variants for a given blob path (DB + memory).
func (c *MediaCache) Invalidate(ctx context.Context, appID, blobPath string) error {
	prefix := c.prefix() + pathHash(blobPath) + "/"
	getMemCache().invalidatePrefix(appID, prefix)
	_, err := c.db.ExecContext(ctx,
		`DELETE FROM app_blobs WHERE app_id = ? AND path LIKE ?`,
		appID, prefix+"%",
	)
	return err
}

// cacheKey builds the full blob path for a cached variant.
func (c *MediaCache) cacheKey(blobPath string, opts TransformOpts) string {
	return c.prefix() + pathHash(blobPath) + "/" + opts.CacheKey()
}

// prefix returns the path prefix for cache entries.
func (c *MediaCache) prefix() string {
	if c.userID != "" {
		return "u/" + c.userID + "/" + mediaCachePrefix
	}
	return mediaCachePrefix
}

// pathHash returns a short hex hash of a blob path.
func pathHash(path string) string {
	h := sha256.Sum256([]byte(path))
	return hex.EncodeToString(h[:8])
}

// InvalidateForPath is a standalone helper for use in s3 bindings.
// It invalidates cached variants for a blob path under the given scope (DB + memory).
func InvalidateForPath(db *sql.DB, appID, blobPath, userID string) {
	var prefix string
	if userID != "" {
		prefix = "u/" + userID + "/" + mediaCachePrefix + pathHash(blobPath) + "/"
	} else {
		prefix = mediaCachePrefix + pathHash(blobPath) + "/"
	}
	getMemCache().invalidatePrefix(appID, prefix)
	db.Exec(`DELETE FROM app_blobs WHERE app_id = ? AND path LIKE ?`, appID, prefix+"%")
}
