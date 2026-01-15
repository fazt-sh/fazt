-- Migration 010: Storage Primitives
-- Provides kv, document store, and blob storage for apps

-- Key-Value Store (app-scoped with TTL)
CREATE TABLE IF NOT EXISTS app_kv (
    app_id TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT,
    expires_at INTEGER,
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER DEFAULT (strftime('%s', 'now')),
    PRIMARY KEY (app_id, key)
);

CREATE INDEX IF NOT EXISTS idx_app_kv_expires ON app_kv(expires_at)
    WHERE expires_at IS NOT NULL;

-- Document Store (JSON documents with collections)
CREATE TABLE IF NOT EXISTS app_docs (
    app_id TEXT NOT NULL,
    collection TEXT NOT NULL,
    id TEXT NOT NULL,
    data TEXT NOT NULL,
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER DEFAULT (strftime('%s', 'now')),
    PRIMARY KEY (app_id, collection, id)
);

CREATE INDEX IF NOT EXISTS idx_app_docs_collection ON app_docs(app_id, collection);

-- Blob Storage (S3-like)
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

CREATE INDEX IF NOT EXISTS idx_app_blobs_prefix ON app_blobs(app_id, path);
