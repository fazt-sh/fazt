# Plan 21: Storage Layer Performance & Concurrency Fixes

**Status**: Planning
**Priority**: P1 (Critical)
**Goal**: Enable 100+ concurrent users on a $6 VPS

## Problem Statement

The CashFlow app stress testing revealed fundamental issues with fazt's storage layer:

| Test | Success Rate | Root Cause |
|------|--------------|------------|
| Sequential writes (20) | 70% | TimeoutError |
| Concurrent writes (10) | 60% | SQLITE_BUSY |
| Sequential reads (50) | 48% | TimeoutError |
| Throttled writes (100ms delays) | 90% | Works with backpressure |

These aren't app-specific issues - they affect ALL apps using `fazt.storage.*`.

## Root Cause Analysis

### Issue 1: Context Not Propagated to Storage

**Location**: `internal/storage/bindings.go`

Every storage binding uses `context.Background()`:
```go
ctx := context.Background()  // ← WRONG
if err := kv.Set(ctx, appID, key, value, ttl); err != nil {
```

This means:
- Storage operations don't respect the runtime's 5s timeout
- A slow query can block indefinitely while the runtime times out
- The operation continues in the background after timeout

### Issue 2: Timeout Race Condition

**Locations**: `internal/database/db.go`, `internal/runtime/runtime.go`

```
Runtime timeout:     5 seconds (DefaultTimeout)
SQLite busy_timeout: 5 seconds (PRAGMA busy_timeout=5000)
```

Scenario:
1. Request starts, runtime has 5s budget
2. First query waits 3s for lock (busy_timeout)
3. Query executes in 0.5s
4. Second query starts, waits 2s for lock...
5. Runtime times out at T=5s, but query is mid-flight

The busy_timeout should be SHORTER than runtime timeout to allow graceful
failure and potential retry.

### Issue 3: No Retry Logic for Transient Errors

SQLITE_BUSY is a transient error that often resolves in milliseconds. Currently,
it's returned immediately as a hard failure. With retry + backoff, most of these
would succeed.

### Issue 4: Missing JSON Index

**Location**: `internal/database/migrations/010_storage.sql`

CashFlow queries documents by session:
```javascript
ds.find('transactions', { session: session })
```

This translates to:
```sql
SELECT * FROM app_docs
WHERE app_id = ? AND collection = ?
  AND json_extract(data, '$.session') = ?
```

Without an index, this scans ALL documents in the collection.

### Issue 5: Global Connection Pool Contention

Single `*sql.DB` shared across all requests. Under load:
- Readers block writers (despite WAL)
- Connection pool exhaustion (MaxOpenConns=25)
- No prioritization of short vs long queries

## Proposed Fixes

### Fix 1: Propagate Context to Storage Operations

**Files**: `internal/storage/bindings.go`, `internal/runtime/handler.go`

Pass the execution context to storage:
```go
// In handler.go - create context with storage deadline
storageCtx, cancel := context.WithTimeout(ctx, 4*time.Second)

// In bindings.go - accept context parameter
func InjectStorageNamespace(vm *goja.Runtime, storage *Storage, appID string, ctx context.Context)

// In each binding
func makeKVSet(...) func(goja.FunctionCall) goja.Value {
    return func(call goja.FunctionCall) goja.Value {
        // Use passed context, not background
        if err := kv.Set(ctx, appID, key, value, ttl); err != nil {
```

### Fix 2: Adjust Timeout Hierarchy

**File**: `internal/database/db.go`

```go
// Before
db.Exec("PRAGMA busy_timeout=5000")  // 5s - same as runtime

// After
db.Exec("PRAGMA busy_timeout=2000")  // 2s - leaves 3s for actual work
```

**File**: `internal/runtime/runtime.go`

Consider increasing runtime timeout for storage-heavy apps:
```go
const (
    DefaultTimeout = 10 * time.Second  // Was 5s
    MaxPoolSize    = 10
)
```

### Fix 3: Add Retry Logic with Backoff

**File**: `internal/storage/storage.go` (new helper)

```go
func withRetry(ctx context.Context, op func() error) error {
    backoff := 10 * time.Millisecond
    maxRetries := 3

    for i := 0; i < maxRetries; i++ {
        err := op()
        if err == nil {
            return nil
        }

        // Only retry on transient errors
        if !isRetryable(err) {
            return err
        }

        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(backoff):
            backoff *= 2
        }
    }
    return lastErr
}

func isRetryable(err error) bool {
    return strings.Contains(err.Error(), "SQLITE_BUSY") ||
           strings.Contains(err.Error(), "database is locked")
}
```

### Fix 4: Add JSON Extraction Index

**File**: `internal/database/migrations/013_storage_perf.sql`

```sql
-- Functional index on common query patterns
-- Note: SQLite 3.38+ supports this syntax
CREATE INDEX IF NOT EXISTS idx_app_docs_session
ON app_docs(app_id, collection, json_extract(data, '$.session'));

-- Alternative: Add explicit session column for indexed queries
ALTER TABLE app_docs ADD COLUMN session_id TEXT;
CREATE INDEX IF NOT EXISTS idx_app_docs_session_id ON app_docs(app_id, collection, session_id);
```

The explicit column approach is more portable and faster.

### Fix 5: Connection Pool Tuning

**File**: `internal/database/db.go`

```go
// Before
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)

// After
db.SetMaxOpenConns(10)     // Fewer connections = less contention
db.SetMaxIdleConns(10)     // Keep all connections warm
db.SetConnMaxLifetime(5 * time.Minute)
db.SetConnMaxIdleTime(1 * time.Minute)
```

SQLite performs better with fewer concurrent connections due to its
locking model. WAL mode helps, but writes are still serialized.

### Fix 6: Write Serialization (Optional)

For write-heavy workloads, serialize all writes through a channel:

```go
type Storage struct {
    writeChan chan writeOp
    // ...
}

func (s *Storage) serializedWrite(ctx context.Context, fn func() error) error {
    done := make(chan error, 1)
    s.writeChan <- writeOp{fn: fn, done: done}
    select {
    case err := <-done:
        return err
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

This eliminates SQLITE_BUSY entirely but adds latency.

## Implementation Order

1. **Fix 1 (Context)** - Critical, unlocks timeout behavior
2. **Fix 2 (Timeouts)** - Simple config change, immediate impact
3. **Fix 3 (Retry)** - Handles remaining transient failures
4. **Fix 5 (Pool)** - Quick config change
5. **Fix 4 (Index)** - Requires migration
6. **Fix 6 (Serialize)** - Only if still seeing issues

## Testing Plan

After fixes, run same stress tests:

```bash
# Target: 95%+ success rate
# Test 1: 100 sequential writes in <30s
# Test 2: 20 concurrent writes, 90%+ success
# Test 3: 100 reads in <5s, 95%+ success
# Test 4: Mixed read/write workload
```

## Expected Outcomes

| Metric | Before | Target |
|--------|--------|--------|
| Sequential write success | 70% | 98%+ |
| Concurrent write success | 60% | 90%+ |
| Read success | 48% | 98%+ |
| Avg response time | 320ms | <100ms |

## Notes

SQLite CAN handle 100+ concurrent users when properly configured:
- WAL mode is already enabled ✓
- Need busy_timeout < runtime_timeout
- Need fewer connections, not more
- Need retry logic for transient failures
- Need proper context propagation

The $6 VPS is not the bottleneck - the configuration is.
