-- Net log: async batch logging for outbound HTTP requests
CREATE TABLE IF NOT EXISTS net_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id TEXT NOT NULL,
    domain TEXT NOT NULL,
    method TEXT NOT NULL,
    path TEXT NOT NULL,
    status INTEGER,
    error_code TEXT,
    duration_ms INTEGER NOT NULL,
    request_bytes INTEGER,
    response_bytes INTEGER,
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE INDEX IF NOT EXISTS idx_net_log_app ON net_log(app_id, created_at);
CREATE INDEX IF NOT EXISTS idx_net_log_domain ON net_log(domain, created_at);
