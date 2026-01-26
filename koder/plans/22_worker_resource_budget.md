# Plan 22: Worker System with Resource Budget

Implementation of background workers with memory pool limits and daemon mode.

## Summary

Add `internal/worker/` package providing:
- **Resource Budget**: 256MB shared pool, soft limits via MemStats
- **Daemon Mode**: Long-running workers with crash restart + checkpoint recovery
- **JS API**: `fazt.worker.spawn()`, `job.progress()`, `job.checkpoint()`

## Files to Create

```
internal/worker/
  pool.go        # WorkerPool - job lifecycle, concurrency limits
  job.go         # Job struct, execution, VM integration
  budget.go      # ResourceBudget - memory tracking
  supervisor.go  # Daemon restart logic with backoff
  bindings.go    # JS API (fazt.worker.*)
  stats.go       # Monitoring endpoints
```

## Files to Modify

| File | Change |
|------|--------|
| `internal/database/migrations/014_workers.sql` | New table |
| `internal/runtime/handler.go` | Add worker injector |
| `cmd/server/main.go` | Init pool, shutdown, restore daemons |

## Database Schema

```sql
CREATE TABLE worker_jobs (
    id TEXT PRIMARY KEY,
    app_id TEXT NOT NULL,
    code_path TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    config TEXT DEFAULT '{}',      -- {daemon, timeout, memoryLimit}
    progress REAL DEFAULT 0.0,
    result TEXT,
    error TEXT,
    logs TEXT,
    checkpoint TEXT,               -- JSON for recovery
    restart_count INTEGER DEFAULT 0,
    created_at INTEGER,
    started_at INTEGER,
    completed_at INTEGER
);
```

## Implementation Phases

### Phase 1: Core Infrastructure
**Goal**: Basic job spawn/cancel/complete

1. Create `internal/worker/pool.go`:
   - `WorkerPool` struct with job map, queue channel
   - `Spawn()` - validates limits, queues job
   - `Cancel()` - sets cancelled flag, interrupts after grace
   - `Shutdown()` - graceful with WaitGroup (pattern from `storage/writer.go`)

2. Create `internal/worker/job.go`:
   - `Job` struct with status, progress, checkpoint
   - `Run()` - gets VM from pool, executes code, handles timeout
   - Pattern from `runtime/runtime.go` lines 115-139 (interrupt)

3. Add migration `014_workers.sql`

4. Wire into `cmd/server/main.go`:
   - Init after `storage.InitWriter()` (line ~2719)
   - Shutdown before `analytics.Shutdown()` (line ~2956)

### Phase 2: JS API
**Goal**: Spawn workers from serverless handlers

1. Create `internal/worker/bindings.go`:
   - `InjectWorkerNamespace()` - adds `fazt.worker.*`
   - `InjectJobContext()` - adds `job.*` inside worker
   - Pattern from `storage/bindings.go` lines 61-83

2. Update `internal/runtime/handler.go`:
   - Add worker injector to `ExecuteWithInjectors()` call

**JS API**:
```javascript
// From serverless handler
fazt.worker.spawn('workers/sync.js', {
    daemon: true,
    memory: '64MB',
    timeout: null
});

// Inside worker
job.progress(0.5);
job.checkpoint({ cursor: 100 });
while (!job.cancelled) { ... }
```

### Phase 3: Resource Budget
**Goal**: Memory pool with soft limits

1. Create `internal/worker/budget.go`:
   - Track allocated bytes per job
   - `Request(jobID, bytes)` - returns false if pool full
   - `Release(jobID)` - frees on completion
   - Monitor goroutine reads `runtime.MemStats` every 100ms
   - Pattern from `capacity/capacity.go` lines 163-182

2. Integrate with pool:
   - Check budget before starting job
   - Queue if pool exhausted (don't error)
   - 500ms grace period before interrupt

### Phase 4: Daemon Mode
**Goal**: Long-running workers with restart

1. Create `internal/worker/supervisor.go`:
   - `OnJobCrash()` - schedule restart with backoff
   - Backoff: 1s, 2s, 4s, 8s... max 60s
   - Reset after 5 minutes healthy

2. Checkpoint recovery:
   - Save to `worker_jobs.checkpoint` column
   - Restore on restart via `job.getCheckpoint()`

3. Add `RestoreDaemons()`:
   - Called at startup
   - Queries `WHERE daemon=1 AND status='running'`
   - Restarts with saved checkpoint

### Phase 5: Polish
**Goal**: Monitoring and tests

1. Create `internal/worker/stats.go`:
   - `PoolStats` struct (active, queued, memory usage)
   - Wire into `/health` endpoint

2. Add tests:
   - `pool_test.go` - spawn, cancel, limits
   - `budget_test.go` - allocation, exhaustion
   - `supervisor_test.go` - restart, backoff

## Key Patterns to Reuse

| Pattern | Source | Reuse For |
|---------|--------|-----------|
| Goroutine pool | `storage/writer.go:46-67` | WorkerPool |
| Graceful shutdown | `storage/writer.go:135-150` | Pool.Shutdown() |
| JS bindings factory | `storage/bindings.go:61-83` | Worker bindings |
| VM injector | `runtime/handler.go:153-176` | Worker context |
| MemStats tracking | `capacity/capacity.go:163-182` | Budget monitor |
| Singleton init | `analytics/buffer.go:54-71` | Pool initialization |

## Verification

1. **Unit tests**: `go test ./internal/worker/...`

2. **Manual test - basic job**:
```javascript
// api/test-worker.js
const jobId = fazt.worker.spawn('workers/hello.js', {});
return { jobId };

// workers/hello.js
job.progress(0.5);
return { message: 'done' };
```

3. **Manual test - daemon**:
```javascript
// Start daemon
fazt.worker.spawn('workers/ticker.js', {
    daemon: true,
    timeout: null
});

// workers/ticker.js
let count = job.getCheckpoint()?.count || 0;
while (!job.cancelled) {
    count++;
    fazt.realtime.broadcast('tick', { count });
    job.checkpoint({ count });
    job.sleep(1000);
}
```

4. **Verify limits**:
   - Spawn 6 jobs for same app (5th should queue)
   - Spawn 21 total jobs (21st should queue)
   - Request 300MB job (should wait for pool)

5. **Verify daemon restart**:
   - Start daemon, kill server, restart
   - Daemon should resume with checkpoint

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Memory spike before kill | 500ms grace + MemStats warning at 80% |
| Checkpoint too large | Limit to 1MB, validate before save |
| Daemon restart storm | Exponential backoff, max 60s |
| SQLite contention | Route writes through WriteQueue |

## Out of Scope

- Hard memory limits (requires cgroups/WASM)
- Operation counting (requires Goja fork)
- Cron scheduling (separate spec, not needed for daemons)
