-- Activity Log: Unified logging system with weight-based prioritization
CREATE TABLE IF NOT EXISTS activity_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp INTEGER NOT NULL DEFAULT (unixepoch()),

    -- Actor
    actor_type TEXT NOT NULL DEFAULT 'system',  -- user/system/api_key/anonymous
    actor_id TEXT,
    actor_ip TEXT,
    actor_ua TEXT,

    -- Resource
    resource_type TEXT NOT NULL,  -- app/user/session/kv/doc/page/config
    resource_id TEXT,

    -- Action
    action TEXT NOT NULL,
    result TEXT DEFAULT 'success',

    -- Weight (0-9, higher = more important)
    weight INTEGER NOT NULL DEFAULT 2,

    -- Details (JSON)
    details TEXT
);

-- Primary lookup: recent activity
CREATE INDEX idx_activity_log_timestamp ON activity_log(timestamp);

-- Filter by importance
CREATE INDEX idx_activity_log_weight ON activity_log(weight);

-- Cleanup: delete low-weight old entries first
CREATE INDEX idx_activity_log_cleanup ON activity_log(weight, timestamp);

-- Resource-specific queries
CREATE INDEX idx_activity_log_resource ON activity_log(resource_type, resource_id);
