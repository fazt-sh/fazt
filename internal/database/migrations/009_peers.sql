-- Migration 009: Peers table for remote fazt nodes
-- Enables fazt-to-fazt communication with all config in SQLite

-- Known remote fazt nodes
CREATE TABLE IF NOT EXISTS peers (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(8)))),
    name TEXT UNIQUE NOT NULL,           -- Human name: "zyt", "home", "pi"
    url TEXT NOT NULL,                   -- Admin URL: https://admin.zyt.app
    token TEXT,                          -- API key for authentication

    -- Metadata
    description TEXT,                    -- "Personal server", "Raspberry Pi"
    is_default INTEGER DEFAULT 0,        -- Only one can be default

    -- Connection state (updated on use)
    last_seen_at TEXT,                   -- Last successful contact
    last_version TEXT,                   -- Last known fazt version
    last_status TEXT,                    -- "healthy", "unreachable", etc.

    -- Future: Mesh identity (v0.16)
    node_id TEXT,                        -- Unique node identifier
    public_key TEXT,                     -- For encrypted mesh communication

    -- Timestamps
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

-- Only one default peer allowed
CREATE UNIQUE INDEX IF NOT EXISTS idx_peers_default
    ON peers(is_default) WHERE is_default = 1;

-- Fast lookup by name
CREATE INDEX IF NOT EXISTS idx_peers_name ON peers(name);
