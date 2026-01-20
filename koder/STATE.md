# Fazt Implementation State

**Last Updated**: 2026-01-20
**Current Version**: v0.10.5

## Status

State: INVESTIGATING
**Task: Runtime Reliability & Observability**

---

## Last Session

- Built momentum todo app, discovered intermittent ~10% timeout failures
- Investigated runtime architecture, identified likely root cause: missing SQLite busy_timeout
- Documented three-track plan: Reliability, Capacity, Defensive Systems
- Fixed /fazt-app skill to enforce `servers/zyt/<name>` path (was creating in repo root)

---

## Problem Statement

Intermittent timeout errors (~10% of requests) in the Goja JS runtime. Observed
pattern: approximately 1 in 10 API requests fail with `TimeoutError: execution
timeout` even when executing simple, fast operations.

**Reproduction:**
```bash
for i in {1..10}; do
  curl -s -X POST "http://192.168.64.3:8080/api/tasks" \
    -H "Host: momentum.192.168.64.3.nip.io" \
    -H "Content-Type: application/json" \
    -d '{"session":"test","text":"Task '$i'"}'
done
# ~1 request will timeout
```

**Impact:**
- Affects all fazt apps with serverless APIs
- Non-deterministic failures frustrate users
- Hard to debug without proper observability

---

## Current Architecture

### Runtime Overview (`internal/runtime/`)

```
┌─────────────────────────────────────────────────────────┐
│                   ServerlessHandler                      │
│  - Loads api/main.js from VFS                           │
│  - Builds Request object from http.Request              │
│  - Executes with 5s outer timeout                       │
└─────────────────────┬───────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────┐
│                      Runtime                             │
│  - VM Pool (10 Goja VMs, pre-warmed)                    │
│  - DefaultTimeout: 1 second                             │
│  - Interrupt mechanism for timeout                      │
└─────────────────────┬───────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────┐
│                   Goja VM                                │
│  Injected globals:                                      │
│  - request (method, path, query, headers, body)         │
│  - respond(status?, body)                               │
│  - console.{log,warn,error,debug}                       │
│  - require() for module loading                         │
│  - fazt.{app, env}                                      │
│  - fazt.storage.{kv, ds, s3}                            │
└─────────────────────────────────────────────────────────┘
```

### Key Files

| File | Lines | Purpose |
|------|-------|---------|
| runtime.go | 667 | Core execution, VM pool, timeout handling |
| handler.go | 232 | HTTP handling, file loading, log persistence |
| bindings.go | 419 | Storage namespace injection (kv, ds, s3) |
| fazt.go | 83 | App context and env var injection |
| stdlib.go | 59 | Standard library modules |

### Timeout Mechanism (runtime.go:112-136)

```go
// Current implementation
timeoutCtx, cancel := context.WithTimeout(ctx, r.timeout)
defer cancel()

go func() {
    select {
    case <-timeoutCtx.Done():
        vm.Interrupt("execution timeout")
    case <-done:
    }
}()
```

**Potential Issues:**
1. Race between timeout goroutine and execution completion
2. VM pool reuse - interrupted VMs may retain bad state
3. No visibility into why timeouts occur

### Storage Bindings Issue (bindings.go)

All storage operations create `context.Background()` instead of using the
request context:
```go
ctx := context.Background()  // ← Should use request context
```

This means storage operations ignore the timeout and could hang indefinitely.

---

## Investigation Areas

### 1. VM Pool Behavior

**Questions:**
- Are interrupted VMs properly reset before reuse?
- Does `vm.ClearInterrupt()` fully restore VM state?
- Should interrupted VMs be discarded instead of returned to pool?

**Test Strategy:**
- Add test that interrupts a VM then reuses it
- Verify subsequent executions work correctly

### 2. Timeout Race Conditions

**Questions:**
- Is there a race between timeout interrupt and normal completion?
- What happens if interrupt fires during VM return to pool?

**Test Strategy:**
- Add tests with very short timeouts (1-10ms)
- Verify no panics or corruption

### 3. Storage Context Propagation

**Questions:**
- Should storage ops respect request timeout?
- Can a slow DB query cause the timeout pattern we see?

**Action:**
- Change `context.Background()` to use request context
- Add timeout to individual DB operations

### 4. Observability Gaps

**Current State:**
- Console logs captured and persisted
- Execution duration tracked
- Errors formatted with line numbers

**Missing:**
- Request-level tracing (unique ID per request)
- Timing breakdown (parse, inject, execute, storage)
- VM pool metrics (active, waiting, interrupted)
- Storage operation timing

---

## Proposed Improvements

### Phase 1: Better Debugging (Immediate)

1. **Request Tracing**
   - Add unique request ID to each execution
   - Include in all log entries
   - Return in error responses

2. **Timing Breakdown**
   ```go
   type ExecuteMetrics struct {
       RequestID   string
       ParseTime   time.Duration
       InjectTime  time.Duration
       ExecuteTime time.Duration
       TotalTime   time.Duration
   }
   ```

3. **Debug Mode Flag**
   - Environment variable: `FAZT_RUNTIME_DEBUG=1`
   - Logs detailed timing for every request
   - Shows VM pool state

### Phase 2: Storage Context Fix (Quick Win)

1. Propagate request context to all storage operations
2. Add per-operation timeout (e.g., 500ms per DB call)
3. Log slow operations (>100ms)

### Phase 3: VM Pool Hardening

1. **Option A: Discard interrupted VMs**
   - Safer but more GC pressure
   - `returnVM()` checks interrupt state

2. **Option B: Better VM reset**
   - Clear all globals after each request
   - Verify clean state before reuse

3. **Pool metrics**
   - Track: pool size, active VMs, interrupted count
   - Expose via `/debug/runtime` endpoint (admin only)

### Phase 4: Comprehensive Testing

1. **Stress tests**
   - 100 concurrent requests
   - Various timeout configurations
   - Verify no corruption

2. **Chaos tests**
   - Random delays in storage operations
   - Verify graceful timeout handling

3. **Regression tests**
   - Known timeout scenarios
   - Verify fixes don't regress

---

## Debug Tools to Build

### 1. Runtime REPL / Test Endpoint

```
POST /api/_debug/execute
{
  "code": "fazt.storage.ds.find('tasks', {})",
  "timeout": 5000
}

Response:
{
  "result": [...],
  "metrics": { parseMs: 1, executeMs: 45 },
  "logs": []
}
```

### 2. VM Pool Status

```
GET /api/_debug/runtime

Response:
{
  "poolSize": 10,
  "available": 8,
  "inUse": 2,
  "totalExecutions": 1547,
  "timeouts": 12,
  "avgExecutionMs": 23
}
```

### 3. Request Trace Log

```
GET /api/_debug/traces?limit=100

Response:
[
  {
    "id": "req_abc123",
    "app": "momentum",
    "path": "/api/tasks",
    "duration": 45,
    "status": 200,
    "breakdown": { inject: 2, execute: 40, storage: 38 }
  }
]
```

---

## Storage Benchmarking

### Why Benchmark?

Before optimizing or adding defensive measures, we need to understand:
- What are the actual limits?
- Where are the bottlenecks?
- What's "normal" vs "degraded" performance?

### Metrics to Capture

| Metric | KV | DS | S3 |
|--------|----|----|-----|
| Read latency (p50, p95, p99) | ? | ? | ? |
| Write latency (p50, p95, p99) | ? | ? | ? |
| Max throughput (ops/sec) | ? | ? | ? |
| Max data size per op | ? | ? | ? |
| Concurrent request limit | ? | ? | ? |
| Memory under load | ? | ? | ? |

### Test Scenarios

1. **Baseline Performance**
   - Single request latency for various data sizes
   - 1KB, 10KB, 100KB, 1MB documents

2. **Throughput Under Load**
   - 10, 50, 100, 500 concurrent requests
   - Mix of reads/writes (80/20, 50/50)

3. **Sustained Load**
   - Constant request rate for 5 minutes
   - Watch for degradation over time

4. **Stress Test**
   - Ramp up until failure
   - Find the breaking point

5. **Recovery**
   - After overload, how quickly does it recover?
   - Are there lasting effects?

### Benchmark Tool Design

```bash
# Proposed CLI
fazt bench storage --type ds --duration 60s --concurrency 50
fazt bench storage --type kv --ops 10000 --report bench-results.json
fazt bench runtime --script test.js --concurrency 100
```

Output:
```
Storage Benchmark: ds (document store)
Duration: 60s | Concurrency: 50

Operations:
  Total:     12,847
  Succeeded: 12,834 (99.9%)
  Failed:    13 (0.1%)

Latency (ms):
  p50:  23
  p95:  67
  p99:  142
  max:  892

Throughput:
  Avg: 214 ops/sec
  Peak: 287 ops/sec

Errors:
  TimeoutError: 8
  DBLocked: 5
```

### Defensive Systems to Consider

Based on benchmark results, we can implement:

1. **Rate Limiting**
   - Per-app request limits
   - Per-operation limits (writes more expensive)
   - Graceful degradation (429 Too Many Requests)

2. **Request Queuing**
   - Queue requests when at capacity
   - Configurable queue depth
   - Timeout for queued requests

3. **Caching**
   - Read-through cache for hot data
   - Cache invalidation on writes
   - Configurable TTL per collection

4. **Circuit Breaker**
   - Fail fast when system is degraded
   - Auto-recovery after cooldown
   - Prevent cascade failures

5. **Resource Limits**
   - Max document size
   - Max documents per collection
   - Max storage per app

6. **DoS Protection**
   - Identify abusive patterns
   - Temporary bans
   - Slow-start for new apps

### SQLite-Specific Considerations

Fazt uses SQLite (via modernc.org/sqlite, pure Go). Key characteristics:
- Single-writer, multiple-reader
- WAL mode for better concurrency
- File-based locking

**Current Configuration** (internal/database/db.go):
```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
PRAGMA journal_mode=WAL  ✓
PRAGMA foreign_keys=ON   ✓
PRAGMA busy_timeout=???  ✗ NOT SET!
```

**Missing: busy_timeout!**
Without busy_timeout, SQLite returns SQLITE_BUSY immediately if the database
is locked. This could explain the ~10% timeout rate - concurrent writes are
failing instantly instead of waiting.

**Recommended fix:**
```go
// Add after WAL mode
db.Exec("PRAGMA busy_timeout=5000")  // Wait up to 5 seconds for lock
```

**Known SQLite limits:**
- Write serialization (one write at a time)
- Database lock contention under high write load
- SQLITE_BUSY errors when lock can't be acquired

**Questions answered:**
- Is WAL mode enabled? **YES**
- What's the busy timeout? **NOT SET (0ms = fail immediately)**
- Are we using connection pooling correctly? **25 max, 5 idle - reasonable**

---

## Action Items

### Track A: Reliability (Timeout Investigation)

1. [ ] **Add SQLite busy_timeout (LIKELY ROOT CAUSE)**
2. [ ] Fix storage context propagation
3. [ ] Add request ID tracing
4. [ ] Add timing breakdown to ExecuteResult
5. [ ] Add debug logging flag
6. [ ] Test VM pool with interrupted VMs
7. [ ] Decide on VM discard vs reset strategy

### Track B: Capacity (Benchmarking)

1. [ ] Build benchmark tool for storage
2. [ ] Run baseline benchmarks (kv, ds, s3)
3. [ ] Test concurrent load scenarios
4. [ ] Find breaking points
5. [ ] Document limits in CLAUDE.md
6. [ ] Check SQLite configuration (WAL, busy timeout)

### Track C: Defensive Systems (After A & B)

1. [ ] Design rate limiting strategy
2. [ ] Implement request queuing (if needed)
3. [ ] Add circuit breaker pattern
4. [ ] Set resource limits per app
5. [ ] Build monitoring dashboard

---

## Quick Reference

```bash
# Run runtime tests
go test -v ./internal/runtime/

# Run with race detector
go test -race ./internal/runtime/

# Build and test locally
go build -o fazt ./cmd/server
./fazt server start --port 8080 --db /tmp/fazt-dev.db
```

## Related Files

- `internal/runtime/` - Core runtime package
- `internal/storage/` - Storage bindings
- `internal/handlers/serverless.go` - HTTP integration
- `koder/ideas/specs/v0.10-runtime/` - Runtime spec docs
