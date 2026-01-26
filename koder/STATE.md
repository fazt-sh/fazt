# Fazt Implementation State

**Last Updated**: 2026-01-26
**Current Version**: v0.10.13

## Status

State: PLANNING - Worker system with resource budget spec'd, ready to implement

---

## Last Session

**Worker Resource Budget Spec**

Designed and spec'd background workers with memory pool limits and daemon mode.

### Spec Updates

**`koder/ideas/specs/v0.8-kernel/limits.md`**:
- Added resource budget model (256MB shared pool)
- Memory-based limits instead of time-based
- Daemon mode configuration

**`koder/ideas/specs/v0.19-workers/jobs.md`**:
- `memory: '64MB'` - request from pool
- `timeout: null` - run indefinitely
- `daemon: true` - restart on crash with backoff
- `job.cancelled` - graceful shutdown flag
- `job.checkpoint()` - crash/restart recovery
- Traffic simulator example

### Key Design Decisions

| Before | After |
|--------|-------|
| 30 min hard timeout | Memory pool (256MB) as primary limit |
| Reject on limit | Queue until memory available |
| Crash = failed | Daemon mode: auto-restart with backoff |

### Implementation Plan

Created `koder/plans/22_worker_resource_budget.md` with 5 phases:
1. Core Infrastructure - WorkerPool, Job, DB migration
2. JS API - fazt.worker.spawn(), job.* bindings
3. Resource Budget - soft limits via MemStats
4. Daemon Mode - restart + checkpoint recovery
5. Polish - tests, monitoring

---

## Previous Session

**Storage API Performance**

Added efficient query operations to prevent memory issues:
- `ds.find(collection, query, { limit, offset, order })`
- `ds.count(collection, query)`
- `ds.deleteOldest(collection, keep)`

---

## Next Up

**Implement Worker System** (Plan 22)

Phase 1: Core Infrastructure
- Create `internal/worker/` package
- `WorkerPool` with job lifecycle
- Migration `014_workers.sql`
- Wire into server startup/shutdown

Start with:
```bash
# Create package structure
mkdir -p internal/worker

# Run existing tests first
go test ./...
```

---

## Quick Reference

```bash
# Build and install
go build -o ~/.local/bin/fazt ./cmd/server

# Run tests
go test ./...

# Local server logs
journalctl --user -u fazt-local -f

# Deploy to local
fazt app deploy <dir> --to local
```

## Key Files for Worker Implementation

| File | Purpose |
|------|---------|
| `internal/storage/writer.go` | Goroutine pool pattern |
| `internal/storage/bindings.go` | JS binding factory pattern |
| `internal/runtime/handler.go` | VM injector integration |
| `internal/capacity/capacity.go` | MemStats tracking |

---

## Backlog

### NEXUS Dashboard Refinement

After worker implementation, return to NEXUS (`servers/zyt/nexus/`):

- **UI/UX polish** - Better styling, animations
- **More widgets** - Table, line chart, heatmap, progress bars
- **Mobile responsiveness**
- **Real API integration** - Connect to fazt analytics
- **Widget library expansion**

Current state: Multi-layout system working (Flight Tracker, Web Analytics, Mall).
MapWidget, DataManager, LayoutSwitcher all functional.
