-- Net allowlist: domain-level access control for serverless outbound HTTP
CREATE TABLE IF NOT EXISTS net_allowlist (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    domain TEXT NOT NULL,
    app_id TEXT,
    https_only INTEGER NOT NULL DEFAULT 1,
    rate_limit INTEGER DEFAULT 0,
    rate_burst INTEGER DEFAULT 0,
    max_response INTEGER DEFAULT 0,
    timeout_ms INTEGER DEFAULT 0,
    cache_ttl INTEGER DEFAULT 0,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    UNIQUE(domain, app_id)
);
