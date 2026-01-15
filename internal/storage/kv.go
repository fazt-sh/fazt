package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// SQLKVStore implements KVStore using SQLite.
type SQLKVStore struct {
	db    *sql.DB
	cache map[string]kvCacheEntry
	mu    sync.RWMutex
	done  chan struct{}
}

type kvCacheEntry struct {
	value     interface{}
	expiresAt *time.Time
	cachedAt  time.Time
}

const (
	kvCacheMaxSize = 1000
	kvCleanupInterval = 5 * time.Minute
)

// NewSQLKVStore creates a new SQLite-backed KV store.
func NewSQLKVStore(db *sql.DB) *SQLKVStore {
	store := &SQLKVStore{
		db:    db,
		cache: make(map[string]kvCacheEntry),
		done:  make(chan struct{}),
	}
	go store.cleanupLoop()
	return store
}

// Set stores a value with optional TTL.
func (s *SQLKVStore) Set(ctx context.Context, appID, key string, value interface{}, ttl *time.Duration) error {
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	var expiresAt *int64
	if ttl != nil {
		exp := time.Now().Add(*ttl).Unix()
		expiresAt = &exp
	}

	query := `
		INSERT INTO app_kv (app_id, key, value, expires_at, updated_at)
		VALUES (?, ?, ?, ?, strftime('%s', 'now'))
		ON CONFLICT(app_id, key) DO UPDATE SET
			value = excluded.value,
			expires_at = excluded.expires_at,
			updated_at = strftime('%s', 'now')
	`
	_, err = s.db.ExecContext(ctx, query, appID, key, string(valueJSON), expiresAt)
	if err != nil {
		return fmt.Errorf("failed to set key: %w", err)
	}

	// Invalidate cache
	s.mu.Lock()
	delete(s.cache, s.cacheKey(appID, key))
	s.mu.Unlock()

	return nil
}

// Get retrieves a value by key.
func (s *SQLKVStore) Get(ctx context.Context, appID, key string) (interface{}, error) {
	cacheKey := s.cacheKey(appID, key)

	// Check cache
	s.mu.RLock()
	if entry, ok := s.cache[cacheKey]; ok {
		if entry.expiresAt == nil || entry.expiresAt.After(time.Now()) {
			s.mu.RUnlock()
			return entry.value, nil
		}
	}
	s.mu.RUnlock()

	// Query database (exclude expired)
	query := `
		SELECT value, expires_at FROM app_kv
		WHERE app_id = ? AND key = ?
		AND (expires_at IS NULL OR expires_at > strftime('%s', 'now'))
	`
	var valueJSON string
	var expiresAt sql.NullInt64
	err := s.db.QueryRowContext(ctx, query, appID, key).Scan(&valueJSON, &expiresAt)
	if err == sql.ErrNoRows {
		return nil, nil // Key not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	var value interface{}
	if err := json.Unmarshal([]byte(valueJSON), &value); err != nil {
		return nil, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	// Update cache
	s.mu.Lock()
	if len(s.cache) >= kvCacheMaxSize {
		s.evictOldest()
	}
	var expTime *time.Time
	if expiresAt.Valid {
		t := time.Unix(expiresAt.Int64, 0)
		expTime = &t
	}
	s.cache[cacheKey] = kvCacheEntry{
		value:     value,
		expiresAt: expTime,
		cachedAt:  time.Now(),
	}
	s.mu.Unlock()

	return value, nil
}

// Delete removes a key.
func (s *SQLKVStore) Delete(ctx context.Context, appID, key string) error {
	query := `DELETE FROM app_kv WHERE app_id = ? AND key = ?`
	_, err := s.db.ExecContext(ctx, query, appID, key)
	if err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}

	// Invalidate cache
	s.mu.Lock()
	delete(s.cache, s.cacheKey(appID, key))
	s.mu.Unlock()

	return nil
}

// List returns all keys matching a prefix.
func (s *SQLKVStore) List(ctx context.Context, appID, prefix string) ([]KVEntry, error) {
	query := `
		SELECT key, value, expires_at FROM app_kv
		WHERE app_id = ? AND key LIKE ?
		AND (expires_at IS NULL OR expires_at > strftime('%s', 'now'))
		ORDER BY key
	`
	rows, err := s.db.QueryContext(ctx, query, appID, prefix+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}
	defer rows.Close()

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

		entry := KVEntry{
			Key:   key,
			Value: value,
		}
		if expiresAt.Valid {
			t := time.Unix(expiresAt.Int64, 0)
			entry.ExpiresAt = &t
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (s *SQLKVStore) cacheKey(appID, key string) string {
	return appID + ":" + key
}

func (s *SQLKVStore) evictOldest() {
	var oldestKey string
	var oldestTime time.Time
	first := true

	for k, entry := range s.cache {
		if first || entry.cachedAt.Before(oldestTime) {
			oldestKey = k
			oldestTime = entry.cachedAt
			first = false
		}
	}

	if oldestKey != "" {
		delete(s.cache, oldestKey)
	}
}

func (s *SQLKVStore) cleanupLoop() {
	ticker := time.NewTicker(kvCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanupExpired()
		case <-s.done:
			return
		}
	}
}

func (s *SQLKVStore) cleanupExpired() {
	// Delete expired keys from database
	_, _ = s.db.Exec(`
		DELETE FROM app_kv
		WHERE expires_at IS NOT NULL
		AND expires_at <= strftime('%s', 'now')
	`)

	// Clean cache
	s.mu.Lock()
	now := time.Now()
	for k, entry := range s.cache {
		if entry.expiresAt != nil && entry.expiresAt.Before(now) {
			delete(s.cache, k)
		}
	}
	s.mu.Unlock()
}

// Close stops the cleanup goroutine.
func (s *SQLKVStore) Close() {
	close(s.done)
}

// HasKey checks if a key exists (not exported, for internal use).
func (s *SQLKVStore) HasKey(ctx context.Context, appID, key string) (bool, error) {
	query := `
		SELECT 1 FROM app_kv
		WHERE app_id = ? AND key = ?
		AND (expires_at IS NULL OR expires_at > strftime('%s', 'now'))
		LIMIT 1
	`
	var exists int
	err := s.db.QueryRowContext(ctx, query, appID, key).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Keys returns all keys matching a prefix (just the keys, not values).
func (s *SQLKVStore) Keys(ctx context.Context, appID, prefix string) ([]string, error) {
	query := `
		SELECT key FROM app_kv
		WHERE app_id = ? AND key LIKE ?
		AND (expires_at IS NULL OR expires_at > strftime('%s', 'now'))
		ORDER BY key
	`
	// Escape SQL LIKE special characters in prefix
	escapedPrefix := strings.ReplaceAll(prefix, "%", "\\%")
	escapedPrefix = strings.ReplaceAll(escapedPrefix, "_", "\\_")

	rows, err := s.db.QueryContext(ctx, query, appID, escapedPrefix+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		keys = append(keys, key)
	}

	return keys, nil
}
