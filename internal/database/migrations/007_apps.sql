-- Migration 007: Sites to Apps Migration
-- Creates proper apps table with metadata

-- 1. Create apps table
CREATE TABLE IF NOT EXISTS apps (
    id TEXT PRIMARY KEY,          -- app_xxxx or existing site_id for migration
    name TEXT NOT NULL UNIQUE,    -- subdomain/app name
    source TEXT DEFAULT 'deploy', -- 'deploy', 'git', 'template', 'system'
    manifest TEXT,                -- JSON manifest (app.json contents)
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_apps_name ON apps(name);
CREATE INDEX IF NOT EXISTS idx_apps_source ON apps(source);

-- 2. Migrate existing sites to apps table
-- Insert existing sites from files table as apps
INSERT OR IGNORE INTO apps (id, name, source, created_at)
SELECT DISTINCT
    site_id as id,
    site_id as name,
    CASE
        WHEN site_id IN ('root', '404', 'admin') THEN 'system'
        ELSE 'deploy'
    END as source,
    MIN(created_at) as created_at
FROM files
GROUP BY site_id;

-- 3. Create domains table for custom domain mapping
CREATE TABLE IF NOT EXISTS domains (
    id TEXT PRIMARY KEY,
    domain TEXT NOT NULL UNIQUE,  -- e.g., 'blog.example.com'
    app_id TEXT NOT NULL,         -- references apps(id)
    is_primary INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (app_id) REFERENCES apps(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_domains_app_id ON domains(app_id);
CREATE INDEX IF NOT EXISTS idx_domains_domain ON domains(domain);

-- 4. Add app_id column to related tables for proper FK (optional, site_id stays for now)
-- Note: SQLite doesn't support ALTER TABLE ADD CONSTRAINT, so we keep site_id as-is
-- In future migrations, we can rename site_id to app_id across all tables

-- 5. Update env_vars to use key/value naming for consistency
-- (key is a reserved word in some SQL dialects, using 'name' for now is fine)
