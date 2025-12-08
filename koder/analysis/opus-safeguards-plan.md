Critical Review: Resource & Resilience Plan
Overall Assessment: Mixed - Good Ideas, Flawed Priorities
The plan identifies real problems but proposes some over-engineered solutions and has inverted priorities.

🔴 Critical Issues
1. Phase 1's Goal is Unrealistic
"Admin panel works even if the DB is physically removed."

Problem: If the DB is removed, the entire application is useless. What would an admin even DO with access? They can't:

View stats (DB is gone)
Manage sites (stored in DB)
View logs (stored in DB)
Deploy anything (VFS is in DB)
The Real Problem to Solve: Admin panel should remain responsive during:

High traffic / DB read contention
WAL checkpoint delays
Large file uploads locking tables
Recommendation: Reframe Phase 1 as "Admin Isolation" not "Admin Independence." Goal should be: Admin never competes with user traffic for DB connections.

2. Emergency Token is Over-Engineered Security Surface
The proposed "Signed JWT/Cookie" fallback auth:

Adds cryptographic complexity
Creates a potential bypass vector if key derivation is weak
Solves an edge case that doesn't justify the risk
Current auth is in-memory anyway (session.go uses sync.Map). If the DB is locked, session validation still works because sessions aren't stored in DB.

Recommendation: Remove the Emergency Token. Instead, ensure sessionStore never touches DB for validation (it already doesn't).

3. Priority Inversion: Phase 4 Should Be Phase 1
Looking at current code (internal/handlers/track.go):

// Every single tracking event writes directly to DB
_, err := db.ExecContext(ctx, `INSERT INTO events ...`, ...)

Under traffic, this will kill performance BEFORE any VFS caching issues matter. Stats buffering is the highest-impact change and should be first.

Recommended Order:

Buffered Stats (current Phase 4)
Resource-Aware Runtime (current Phase 2)
Circuit Breakers (current Phase 3)
Admin Isolation (current Phase 1, simplified)
Emergency CLI Tools (current Phase 5)
🟡 Design Concerns
4. Hardware Probe Must Be Cgroup-Aware
From CLAUDE.md: Target is Digital Ocean droplet. But what about future containerization?

// Bad: Uses host RAM
total := runtime.MemStats.Sys

// Good: Respects cgroup limits
// /sys/fs/cgroup/memory/memory.limit_in_bytes

Recommendation: Read /sys/fs/cgroup/memory.max (cgroup v2) first, fallback to /proc/meminfo.

5. VFS Cache Eviction is Currently "All or Nothing"
Current code in internal/hosting/vfs.go:

if len(fs.cache) >= 1000 {
    fs.cache = make(map[string]*cachedFile)  // Wipes everything
}

The plan's LRU approach is good, but tracking bytes instead of entries is correct. However:

Missing Detail: How do you evict when a single 50MB file exceeds the limit? Need minimum-viable-cache rules (e.g., always keep index.html for active sites).

6. Streaming Uploads - Where Do Temp Files Go?
"Safe Path: Size > MaxRAMBuffer → Stream to TempFile → DB"

Questions not addressed:

What if /tmp is a ramfs (common in containers)?
Cleanup on crash? Need atomic rename or defer cleanup.
What if temp partition is smaller than upload?
Recommendation: Add to Phase 2:

 Temp file directory config (FAZT_TEMP_DIR env var)
 Atomic upload: write to temp, then DB transaction
 Cleanup goroutine for orphaned temps
7. CPU Semaphore Failure Mode Unclear
"Add a Semaphore (buffered channel) to hosting.RunServerless. Drop requests if channel full."

What does "drop" mean?

Return 503 Service Unavailable?
Queue and wait with timeout?
Silent failure?
Recommendation: Specify behavior explicitly:

select {
case sem <- struct{}{}:
    // proceed
default:
    http.Error(w, "Service busy", http.StatusServiceUnavailable)
    return
}

🟢 Good Ideas Worth Keeping
Item	Why It's Good
Buffered Stats	Direct DB writes per event is the real bottleneck
Disk Space Check	SQLite doesn't handle disk full gracefully
PRAGMA auto_vacuum = INCREMENTAL	Prevents DB bloat
Pinned Admin Assets	Reduces DB reads for frequently-accessed files
LRU Cache with Byte Tracking	Current "nuke all" eviction is bad
🔵 Missing Items
A. Graceful Shutdown for Buffered Stats
If you buffer writes and the process crashes/restarts, you lose data. Need:

// In main.go
signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
<-sigChan
statsBuffer.Flush() // Ensure writes before exit

B. DB Connection Pool Isolation
Current code uses single pool for everything. Consider:

type DBPools struct {
    Admin    *sql.DB  // 2 connections, for dashboard
    Analytics *sql.DB // 3 connections, for stats writes
    Hosting  *sql.DB  // 20 connections, for VFS reads
}

This is more impactful than "pinned VFS" for keeping admin responsive.

C. Health Check Endpoint Behavior
What happens to /health under load? Should it:

Always return 200 (liveness)?
Return 503 if circuit breakers are tripped (readiness)?
D. Metrics for Operators
How does admin know breakers tripped? Need:

Current cache size / max
Requests dropped by semaphore
Buffer flush stats
Revised Plan Suggestion
## Phase 1: Stats Buffering (Critical Path)
- [ ] Create `internal/analytics/buffer.go` with `sync.Map`
- [ ] Flush goroutine (every 30s OR 1000 events)
- [ ] Graceful shutdown hook to flush before exit
- [ ] Update `/track` and `/pixel.gif` to use buffer

## Phase 2: Resource Awareness
- [ ] Cgroup-aware probe (`/sys/fs/cgroup/` → `/proc/meminfo` fallback)
- [ ] Config: `MaxVFSCacheBytes`, `MaxUploadRAM`, `TempDir`
- [ ] Streaming uploads with temp file + cleanup

## Phase 3: VFS Cache Improvements  
- [ ] LRU eviction by bytes (not entry count)
- [ ] Pin admin assets in non-evictable map on boot
- [ ] Metrics: cache hit rate, size

## Phase 4: Circuit Breakers
- [ ] Request semaphore with 503 on overflow
- [ ] Disk space check before VFS writes (reject at 90%)
- [ ] Expose breaker status at `/api/system/health`

## Phase 5: CLI Recovery Tools
- [ ] `fazt db vacuum` 
- [ ] `fazt db prune --older-than 30d`
- [ ] Set `auto_vacuum = INCREMENTAL` in migrations

Summary
Original Phase	Verdict
Phase 1 (Cockpit)	🔴 Over-engineered. Simplify to "pin admin assets" only
Phase 2 (Probe)	🟡 Good, but needs cgroup awareness
Phase 3 (Breakers)	🟢 Solid, needs explicit failure modes
Phase 4 (Buffering)	🟢 Critical - should be Phase 1
Phase 5 (CLI)	🟢 Good, straightforward
Key Recommendation: Flip the priority. Stats buffering is the biggest production risk right now. The "indestructible cockpit" scenario is an edge case that doesn't justify the complexity of emergency tokens.
