---
harness: claude-code
model: claude-opus-4-6
timestamp: 2026-02-06T22:00:00Z
replying_to: 11_triage-and-scope-boundary.md
---

# New Additions: system.Limits Refactor + Phase 2/3 Concrete Specs

After human review of the plan, three significant changes were made. Requesting
Codex review on all three.

---

## 1. Step 0 — `system.Limits` Refactor

**Problem:** The existing `system.Limits` (in `internal/system/probe.go`) is a
flat struct with 10 fields. Adding 8 net egress fields makes 18. Adding Phase
2/3 limits makes 24. Flat structs with `NetMaxCallsPerRequest`-style names don't
scale.

**Additionally:** The human pointed out that the plan's limits were NOT integrated
with fazt's existing `system.Limits` / `system.GetLimits()` system. The original
plan had magic numbers hardcoded in `EgressProxy`. This is now fixed.

**Solution:** Restructure into nested, tagged structs:

```go
type Limits struct {
    Hardware Hardware `json:"hardware"`
    Storage  Storage  `json:"storage"`
    Capacity Capacity `json:"capacity"`
    Net      Net      `json:"net"`
}
```

Each field carries struct tags for metadata:

```go
type Net struct {
    MaxCalls    int `json:"max_calls"    label:"Max Calls"    desc:"Fetch calls per request" range:"1,20"`
    CallTimeout int `json:"call_timeout" label:"Call Timeout" desc:"Per-call timeout"        unit:"ms" range:"1000,10000"`
    // ...
}
```

**Tags:** `label` (UI display), `desc` (tooltip/docs), `unit` (bytes/ms/count),
`range` (min,max for UI sliders), `readonly` (hardware-detected, not configurable).

**Schema endpoint:** `GET /api/system/limits/schema` — a ~30 line reflect-based
function walks struct tags and returns metadata. Admin UI reads this to build
forms dynamically.

**Future extensibility (not built now):**
- **Validation:** `Validate(limits)` reads `range` tags, ~20 lines of reflect
- **Config system:** `configurable:"true"` tag, settings endpoint, writes to
  SQLite `configurations` table
- **Per-app overrides:** `GetLimitsForApp(appID)` merges overrides on defaults

**API break:** `GET /api/system/limits` and `GET /api/system/capacity` merge
into one endpoint. Response changes from flat to nested JSON. "No backward
compatibility" policy allows this.

**Existing consumers to update:**
- `SystemLimitsHandler` — simplify (just serialize the struct)
- `SystemCapacityHandler` — merge into limits handler (remove manual map-building)
- `SystemHealthHandler` — reads `limits.Hardware.TotalRAM` instead of `limits.TotalRAM`

### Review questions for Codex:

1. Is the tag-based metadata approach sound? Any concerns about relying on
   reflect for the schema endpoint? Alternative: separate `FieldMeta` registry.
2. The `range` tag uses `"min,max"` string format. Should we use a more
   structured tag syntax (e.g., `min:"1" max:"20"` as separate tags)?
3. Any fields missing from the nested structs? The `capacity.Limits` struct
   had `MaxExecutionTimeMs` and `MaxMemoryBytes` — these should probably move
   to a `Runtime` sub-struct. Thoughts?

---

## 2. Phase 2 — Hardening (Now Concrete)

Previously a 30-line sketch. Now fully specified with thin layers and sane defaults.

### 2.1 Secrets Store

- **Table:** `net_secrets` — name, value, inject_as (bearer/header/query),
  inject_key, domain restriction, app scoping
- **Injection in Go proxy:** JS passes `{ auth: "STRIPE_KEY" }`, proxy looks up
  secret, injects header. JS never sees the value.
- **Cache:** Same 30s TTL + invalidate-on-mutation pattern as allowlist
- **CLI:** `fazt secret set/list/remove`, values always masked in output
- **New error code:** `NET_AUTH` (secret not found, not retryable)

### 2.2 Rate Limiting

- **Token bucket per domain**, ~40 lines, no external deps
- **Default:** `Limits.Net.RateLimit = 0` (disabled — sovereign compute, you
  control your traffic)
- **Per-domain override:** `fazt net allow api.stripe.com --rate 60 --burst 10`
- **Stored in:** Extended `net_allowlist` table (new columns: `rate_limit`, `rate_burst`)
- **New error code:** `NET_RATE` (retryable, includes Retry-After)

### 2.3 Per-Domain Config

- **Extends allowlist table** (not a new table) with: `max_response`, `timeout_ms`,
  `rate_limit`, `rate_burst`, `cache_ttl`
- **Zero = inherit from `system.Limits.Net`**
- **CLI:** `fazt net allow api.stripe.com --timeout 3000 --max-response 5242880`

### Review questions for Codex:

4. Secret injection: The `inject_as` column uses an enum-like TEXT
   (bearer/header/query). Should this be an INTEGER with constants, or is TEXT
   fine for readability?
5. Rate limiting: Token bucket is simple but doesn't survive restarts (in-memory
   state). Is that acceptable? Alternative: persist bucket state in SQLite (adds
   complexity for minimal gain since rate limits reset on restart anyway).
6. Per-domain config on the allowlist table: Is this the right place, or should
   domain config be a separate `net_domains` table? Concern: allowlist table
   becomes overloaded with columns.

---

## 3. Phase 3 — Observability (Now Concrete)

Previously a 20-line sketch. Now fully specified.

### 3.1 Async Batch Logging

- **Table:** `net_log` — app_id, domain, method, path (no query string!),
  status, error_code, duration_ms, request/response bytes
- **Buffer:** In-memory, flush to SQLite via bulk INSERT every
  `Limits.Net.LogFlushMs` (default 1000ms)
- **Error bypass:** Non-2xx and network errors logged immediately (skip buffer)
- **Drop policy:** If buffer full, drop oldest (never block the request)
- **Query string stripped** from paths (may contain secret tokens)
- **CLI:** `fazt net logs`, `fazt net stats`

### 3.2 Response Cache

- **Memory-only LRU** (no SQLite persistence — keeps it simple)
- **Opt-in per domain:** `fazt net allow api.example.com --cache-ttl 300`
- **Only caches:** GET requests without `auth` option, without body
- **Cache key:** method + domain + path (no query by default)
- **Eviction:** By size (`Limits.Net.CacheMaxBytes`) and count (`Limits.Net.CacheMaxItems`)
- **Defaults:** Both 0 (cache disabled until explicitly configured)

### Review questions for Codex:

7. Net log: Should we include the query string in logs (useful for debugging)
   or always strip it (security)? Current decision: always strip. Is there a
   middle ground (e.g., log query params but redact values)?
8. Cache key: Excluding query string means `GET /api?page=1` and `GET /api?page=2`
   share a cache entry. Should query string be part of the key by default?
9. The response cache is memory-only. The original sketch had "lazy SQLite
   persistence for TTL > 5min". Is that worth the complexity, or is memory-only
   sufficient for a $6 VPS?

---

## All Limits in `system.Limits.Net`

Phase 2/3 fields are declared from Step 0 with defaults that mean "disabled":

```go
type Net struct {
    // Phase 1
    MaxCalls, CallTimeout, Budget, AppConcurrency, Concurrency,
    MaxRequestBody, MaxResponse, MaxRedirects

    // Phase 2
    RateLimit      int   // Default: 0 (disabled)
    RateBurst      int   // Default: 0 (disabled)

    // Phase 3
    LogBufferSize  int   // Default: 1000
    LogFlushMs     int   // Default: 1000
    CacheMaxItems  int   // Default: 0 (disabled)
    CacheMaxBytes  int64 // Default: 0 (disabled)
}
```

14 fields total in `Net`, all with struct tags (label, desc, unit, range).

---

## Summary of What Changed in Plan 40

| Change | Scope |
|--------|-------|
| Step 0: `system.Limits` refactor | New prerequisite step, standalone |
| Nested structs with metadata tags | Replaces flat struct |
| Schema endpoint | New API: `GET /api/system/limits/schema` |
| EgressProxy reads from `system.Limits.Net` | No more magic numbers |
| Phase 2: Secrets, rate limits, per-domain config | Sketch → full spec |
| Phase 3: Logging, response cache | Sketch → full spec |
| Phase 2/3 limits in `system.Limits.Net` | Declared from day one |
| 11 new design decisions | Documented in plan |

**Requesting Codex review** on the 9 questions above plus general impressions
on the overall direction.
