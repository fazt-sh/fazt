-- Migration 012: App Identity, Aliases, and Lineage (v0.10)
-- This migration separates app identity from routing (aliases)

-- 1. Create new apps table with identity model
CREATE TABLE apps_new (
    -- Identity (immutable)
    id TEXT PRIMARY KEY,              -- "app_7f3k9x2m" (UUID, never changes)

    -- Lineage
    original_id TEXT,                 -- Root ancestor (self if original)
    forked_from_id TEXT,              -- Immediate parent (NULL if original)

    -- Metadata (inherited on fork)
    title TEXT,                       -- "Tetris" - what it is
    description TEXT,                 -- "Classic block-stacking game"
    tags TEXT,                        -- JSON array: ["game", "arcade"]
    visibility TEXT DEFAULT 'unlisted', -- public|unlisted|private

    -- Source tracking
    source TEXT DEFAULT 'deploy',     -- 'deploy', 'git', 'fork', 'system'
    source_url TEXT,
    source_ref TEXT,
    source_commit TEXT,

    -- Timestamps
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_apps_new_original ON apps_new(original_id);
CREATE INDEX idx_apps_new_forked_from ON apps_new(forked_from_id);
CREATE INDEX idx_apps_new_visibility ON apps_new(visibility);

-- 2. Create aliases table (routing layer)
CREATE TABLE aliases (
    -- Routing key
    subdomain TEXT PRIMARY KEY,       -- "tetris" → tetris.zyt.app

    -- Routing behavior
    type TEXT DEFAULT 'proxy',        -- proxy|redirect|reserved|split

    -- Target(s)
    targets TEXT,                     -- JSON (structure depends on type)

    -- Timestamps
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 3. Migrate existing apps to new schema
-- Generate app IDs based on existing name, create aliases
INSERT INTO apps_new (id, original_id, title, source, source_url, source_ref, source_commit, created_at, updated_at)
SELECT
    'app_' || lower(substr(hex(randomblob(4)), 1, 8)) as id,
    'app_' || lower(substr(hex(randomblob(4)), 1, 8)) as original_id,
    name as title,
    source,
    source_url,
    source_ref,
    source_commit,
    created_at,
    updated_at
FROM apps;

-- Fix original_id to match id (they're all originals)
UPDATE apps_new SET original_id = id;

-- 4. Create temp mapping table for migration
CREATE TEMP TABLE app_id_mapping AS
SELECT
    apps.name as old_name,
    apps_new.id as new_id
FROM apps
JOIN apps_new ON apps.name = apps_new.title;

-- 5. Create aliases for each migrated app
INSERT INTO aliases (subdomain, type, targets, created_at, updated_at)
SELECT
    m.old_name as subdomain,
    'proxy' as type,
    json_object('app_id', m.new_id) as targets,
    a.created_at,
    a.updated_at
FROM app_id_mapping m
JOIN apps a ON a.name = m.old_name;

-- 6. Add app_id column to files table and migrate
ALTER TABLE files ADD COLUMN app_id TEXT;

UPDATE files SET app_id = (
    SELECT new_id FROM app_id_mapping WHERE old_name = files.site_id
);

-- 7. Create reserved system aliases
INSERT OR IGNORE INTO aliases (subdomain, type, targets, created_at)
VALUES
    ('admin', 'reserved', NULL, CURRENT_TIMESTAMP),
    ('api', 'reserved', NULL, CURRENT_TIMESTAMP);

-- 8. Drop old apps table and rename new one
DROP TABLE apps;
ALTER TABLE apps_new RENAME TO apps;

-- 9. Re-create indexes on apps table
CREATE INDEX idx_apps_original ON apps(original_id);
CREATE INDEX idx_apps_forked_from ON apps(forked_from_id);
CREATE INDEX idx_apps_visibility ON apps(visibility);
CREATE INDEX idx_apps_title ON apps(title);

-- 10. Create index on files.app_id
CREATE INDEX IF NOT EXISTS idx_files_app_id ON files(app_id);
