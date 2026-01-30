-- Migration 017: User-Scoped Storage
-- Adds user_id to storage tables for user isolation (Plan 30b)

-- Add user_id column to app_kv
-- NULL = shared app data, non-NULL = user's private data
ALTER TABLE app_kv ADD COLUMN user_id TEXT;

-- Add user_id column to app_docs
ALTER TABLE app_docs ADD COLUMN user_id TEXT;

-- Add user_id column to app_blobs
ALTER TABLE app_blobs ADD COLUMN user_id TEXT;

-- Indexes for efficient user-scoped queries
CREATE INDEX IF NOT EXISTS idx_app_kv_user ON app_kv(app_id, user_id);
CREATE INDEX IF NOT EXISTS idx_app_docs_user ON app_docs(app_id, user_id, collection);
CREATE INDEX IF NOT EXISTS idx_app_blobs_user ON app_blobs(app_id, user_id);

-- Add user_id and app_id to events for analytics (if events table exists)
-- These may already exist, so we use ALTER TABLE ADD COLUMN which will fail silently on SQLite
-- if column exists. Wrapped in separate statements.
