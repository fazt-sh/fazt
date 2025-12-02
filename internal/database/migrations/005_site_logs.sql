-- This migration is redundant because site_logs was accidentally included in 004_vfs.sql
-- We keep this file to maintain the migration version sequence (5).

-- CREATE TABLE site_logs (
--     id INTEGER PRIMARY KEY AUTOINCREMENT,
--     site_id TEXT NOT NULL,
--     level TEXT NOT NULL,
--     message TEXT NOT NULL,
--     created_at DATETIME DEFAULT CURRENT_TIMESTAMP
-- );

-- CREATE INDEX idx_site_logs_site_id ON site_logs(site_id);
-- CREATE INDEX idx_site_logs_created_at ON site_logs(created_at);
