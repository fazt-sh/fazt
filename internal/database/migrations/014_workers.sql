-- Migration 014: Worker Jobs
-- Background job queue with resource limits and daemon mode

CREATE TABLE IF NOT EXISTS worker_jobs (
    id TEXT PRIMARY KEY,
    app_id TEXT NOT NULL,
    handler TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    config TEXT DEFAULT '{}',
    progress REAL DEFAULT 0.0,
    result TEXT,
    error TEXT,
    logs TEXT DEFAULT '[]',
    checkpoint TEXT,
    attempt INTEGER DEFAULT 1,
    restart_count INTEGER DEFAULT 0,
    daemon_backoff_ms INTEGER DEFAULT 0,
    created_at INTEGER,
    started_at INTEGER,
    done_at INTEGER,
    last_healthy_at INTEGER
);

-- Index for querying jobs by app and status
CREATE INDEX IF NOT EXISTS idx_worker_jobs_app_status
ON worker_jobs(app_id, status);

-- Index for job queue ordering (pending jobs by priority and creation time)
CREATE INDEX IF NOT EXISTS idx_worker_jobs_queue
ON worker_jobs(status, created_at)
WHERE status = 'pending';

-- Index for daemon jobs (for RestoreDaemons)
CREATE INDEX IF NOT EXISTS idx_worker_jobs_daemon
ON worker_jobs(status)
WHERE json_extract(config, '$.daemon') = 1;

-- Index for cleanup of old completed jobs
CREATE INDEX IF NOT EXISTS idx_worker_jobs_cleanup
ON worker_jobs(done_at)
WHERE status IN ('done', 'failed', 'cancelled');
