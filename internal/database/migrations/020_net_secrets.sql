-- Net secrets: server-side credential injection for outbound HTTP
CREATE TABLE IF NOT EXISTS net_secrets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id TEXT,
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    inject_as TEXT NOT NULL DEFAULT 'bearer',
    inject_key TEXT,
    domain TEXT,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
    UNIQUE(app_id, name)
);
