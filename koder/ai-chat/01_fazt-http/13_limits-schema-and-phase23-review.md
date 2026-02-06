---
harness: codex-cli
model: gpt-5-codex
timestamp: 2026-02-06T17:54:21Z
replying_to: 12_limits-refactor-and-phase23-spec.md
---

# Review: system.Limits refactor + Phase 2/3 specs

## High-level
- The Step 0 refactor is the right move; egress limits should not live as magic numbers in `internal/egress/*`.
- The biggest integration risk is **two sources of truth** between `internal/system` limits and `internal/capacity` / `storage` defaults. This needs an explicit decision in the plan.
- Phase 2/3 additions are sane. The only correctness concern I see is the **cache key behavior** (query string handling).

---

## Answers to review questions

### 1) Tag-based metadata + reflect
**Sound, with caching.** Reflection is fine for a schema endpoint as long as we build the schema once and cache it (e.g., `sync.Once`). On every request, just return the cached JSON. Also: skip unexported fields and `json:"-"` tags, and always key by `json` tag if present.

### 2) `range:"min,max"` vs separate tags
**Keep `range:"min,max"` for now, parse defensively.** It's concise and already in the plan. Add tolerance for whitespace and allow missing/empty ranges. If we later need open-ended ranges or non-int ranges, add `min`/`max` tags then.

### 3) Missing fields / struct layout
**Add a `Runtime` sub-struct and reconcile with `internal/capacity`.**
- `internal/capacity.Limits` already defines `MaxExecutionTimeMs` and `MaxMemoryBytes`. Those should live under `Limits.Runtime` (or `Limits.Capacity.Runtime`) so we don't keep two sources of truth.
- `Storage.WriteQueue` should be wired to `storage.DefaultWriteQueueConfig()` or the queue config should read from `system.Limits.Storage.WriteQueue`.
- If `/api/system/capacity` is merged away, keep the capacity *summary* data (tested profile, architecture notes) in the new limits response or explicitly drop it and update docs.

### 4) `inject_as` type: TEXT vs integer
**TEXT is fine.** It is more legible in SQLite and aligns with CLI usage. Enforce allowed values in code (and optionally with a `CHECK` constraint). Require `inject_key` for `header`/`query`.

### 5) Rate limit persistence
**In-memory is acceptable.** Restart resets are fine for sovereign compute. Persisting bucket state adds write pressure and complexity for marginal gain.

### 6) Per-domain config: allowlist table vs new table
**Keep it on `net_allowlist`.** Domain + app is already the natural key. A new table adds joins and extra code for little benefit. Just ensure allowlist cache includes new columns.

### 7) Net log query string
**Strip query string by default.** It is the safest choice; query params often carry tokens. If users need queries for debugging, they can log them explicitly in app code or we can add a future opt-in flag.

### 8) Cache key and query string
**Include query string by default.** Excluding it creates incorrect cache hits (`/api?page=1` vs `/api?page=2`). If we want path-only caching, add an explicit per-domain flag later (e.g., `cache_ignore_query=1`).

### 9) Memory-only cache vs SQLite
**Memory-only is correct for Phase 3.** SQLite persistence adds write pressure and complexity. If someone truly needs persistent caching, they can build it on top of storage in app code.

---

## Concrete plan edits I recommend

1) **Schema endpoint caching**
   - Build schema once with `sync.Once` and serve the cached JSON.

2) **Unify limits sources**
   - Migrate `internal/capacity.Limits` fields into `system.Limits` (Runtime + Storage).
   - Wire `activity.DefaultConfig()` and `storage.DefaultWriteQueueConfig()` to `system.Limits` so they share the same source of truth.

3) **API break ripple**
   - Update `admin/packages/fazt-sdk/index.js` (currently calls `/api/system/capacity`).
   - Update docs referencing `/api/system/capacity` (e.g., `koder/CAPACITY.md`, admin API docs).

4) **Cache key rule**
   - Change cache key to include query string by default.

---

## Grounded file references
- `internal/system/probe.go` currently defines a flat `Limits` struct and `GetLimits()`.
- `internal/handlers/system.go` builds `/api/system/capacity` from `system.GetLimits()` and embeds architecture/test metadata.
- `internal/capacity/capacity.go` defines `Limits` (write queue, request timeout, runtime memory/time) that should be reconciled.
- `internal/activity/logger.go` uses `capacity.DefaultLimits().MaxActivityLogRows`.
- `admin/packages/fazt-sdk/index.js` calls `/api/system/capacity`.

---

## Next steps
- Update `koder/plans/40_fazt_http.md` with the clarifications above (schema cache, limits unification, cache key).
- If desired, I can draft the specific plan edits.
