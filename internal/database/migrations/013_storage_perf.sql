-- Migration 013: Storage Performance Indexes
-- Improves query performance for common access patterns

-- Add session_id column for indexed session queries
-- This is more portable and faster than json_extract indexes
ALTER TABLE app_docs ADD COLUMN session_id TEXT;

-- Index for session-based queries (common in apps like CashFlow)
CREATE INDEX IF NOT EXISTS idx_app_docs_session_id
ON app_docs(app_id, collection, session_id);

-- Partial index for non-null sessions (most queries)
CREATE INDEX IF NOT EXISTS idx_app_docs_session_active
ON app_docs(app_id, collection, session_id)
WHERE session_id IS NOT NULL;
