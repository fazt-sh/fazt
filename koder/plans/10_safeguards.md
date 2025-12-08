# Plan: Unsinkable Safeguards & Stability Architecture

## Objective
Transform Fazt from a "fragile" single-binary PaaS into an "Unsinkable" platform.
**Primary Strategy**: Eliminate high-frequency DB writes and enforce strict resource limits based on the container environment.

## Core Philosophy
1.  **Write Efficiency**: Never thrash the WAL with high-frequency counter updates. Buffered Stats is the #1 stability requirement.
2.  **Container Awareness**: Respect Cgroup limits (Docker/K8s) over Host limits.
3.  **Admin Isolation**: Admin assets are pinned in RAM; Admin traffic has reserved DB connections.
4.  **Fail Fast**: Return 503s immediately when resources (CPU/RAM/Disk) are exhausted rather than crashing.

## Architecture Changes

### 1. Buffered IO (The Anti-Thrash)
*   **Accumulator Pattern**: Request stats (visits, bandwidth) are aggregated in a `sync.Map` in memory.
*   **Batch Flusher**: A background goroutine flushes aggregated stats to the DB in a single transaction every 30 seconds (or when buffer > 1000 events).
    *   **Failure Mode**: Retry once after 5s; if still failing, drop oldest events (preserve RAM) and log warning.
*   **Graceful Shutdown**: Flush buffer on `SIGTERM`/`SIGINT`.

### 2. Resource-Aware Runtime
*   **Cgroup Probe**: Detect memory limits via `/sys/fs/cgroup/...` with fallback to `/proc/meminfo`.
*   **Smart Limits**:
    *   `MaxVFSCache`: 25% of Available RAM (tracked by **Bytes**, not count).
    *   `MaxUploadRAM`: 10% of RAM (buffer size before streaming to disk).
    *   `TempDir`: Configurable via `FAZT_TEMP_DIR` (default: `/tmp`), with auto-cleanup.

### 3. VFS Modernization
*   **Pinned Cache**: `admin/*` assets are loaded into a separate, non-evictable memory map on boot.
*   **Byte-Weighted LRU**: The main VFS cache evicts based on total byte size.
    *   **Cache Floor**: Never evict `index.html` for sites accessed in the last 60s (ensure landing pages load).
*   **Stream-Switch**:
    *   Small Uploads (< `MaxUploadRAM`) -> RAM -> DB.
    *   Large Uploads (> `MaxUploadRAM`) -> TempFile -> DB.

## Phased Implementation

### Phase 1: Write Optimization (Critical Path)
*Goal: Stop the DB from dying under normal traffic.*
- [ ] **Stats Buffer**: Create `internal/analytics/buffer.go` with `sync.Map`.
- [ ] **Flusher**: Implement the background flush goroutine (30s interval) with drop-oldest failure logic.
- [ ] **Shutdown Hook**: Ensure `main.go` waits for buffer flush on exit.
- [ ] **Integration**: Update `handlers/track.go`, `handlers/pixel.go`, and `middleware/logging.go` to use the buffer.

### Phase 2: Resource Awareness
*Goal: Know our limits.*
- [ ] **Probe**: Create `internal/system/probe.go` (Cgroup v1/v2 support).
- [ ] **Config**: Add `MaxVFSCacheBytes`, `MaxUploadRAM`, `TempDir` to `config.Config`.
- [ ] **Temp Manager**: Implement secure temp file creation with auto-cleanup.

### Phase 3: VFS Modernization
*Goal: Protect the Admin Panel and RAM.*
- [ ] **Connection Pools**: Split `sql.DB` into `AdminPool` (2 conns) and `HostingPool` (Max - 2).
- [ ] **Pinned Assets**: Update `hosting.Init` to load `admin` site into a `pinned` map.
- [ ] **Byte-LRU**: Refactor `internal/hosting/vfs.go` to track usage in bytes with Cache Floor logic.
- [ ] **Streaming**: Update `handlers/deploy.go` to use the Stream-Switch logic.

### Phase 4: Circuit Breakers (Runtime Safety)
*Goal: Reject work when full.*
- [ ] **CPU Semaphore**: Add a buffered channel to `hosting.RunServerless`. Return 503 if full.
- [ ] **Disk Guard**: Check `syscall.Statfs` before `WriteFile`. Reject if usage > 90%.
- [ ] **Metrics API**: `GET /api/system/metrics` (cache size, buffer depth, semaphore usage).
- [ ] **Health API**: Update `/health` to report 503 if breakers are tripped.

### Phase 5: Recovery Tools (CLI)
*Goal: Fix it when it breaks.*
- [ ] **DB Tuning**: Set `PRAGMA auto_vacuum = INCREMENTAL` in `database/db.go`.
- [ ] **Vacuum Command**: `fazt server vacuum` (runs `PRAGMA incremental_vacuum`).
- [ ] **Prune Command**: `fazt server prune` (cleans logs/cache).

## Roadmap Dependencies
*   Phase 1 is blocking for ANY high-traffic deployment.
*   Phase 3 is blocking for `core-backup`.
