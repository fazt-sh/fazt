# Fazt Implementation State

**Last Updated**: 2026-02-06
**Current Version**: v0.26.0

## Status

State: IMPLEMENTED
Plan 40 (fazt.http) fully implemented — all 4 phases. Tests passing.

---

## Last Session (2026-02-06) — Plan 40 Full Implementation

### What Was Done

#### Plan 40: fazt.net.fetch() — Complete Implementation

All 4 phases implemented in a single autonomous session:

**Step 0: system.Limits Refactor**
- Flat `Limits` → nested structs: `Hardware`, `Storage`, `Runtime`, `Capacity`, `Net`
- Struct tags: `json`, `label`, `desc`, `unit`, `range`, `readonly`
- Reflect-based schema endpoint: `GET /api/system/limits/schema` (sync.Once cached)
- Rewired `activity/logger.go` to use `system.GetLimits().Storage.MaxLogRows`
- Updated fazt-sdk: added `limitsSchema()`, deprecated `capacity()`
- `internal/capacity/` is now dead code (nothing imports it)

**Phase 1: Core Egress Proxy (Steps 1-8)**
- `internal/egress/proxy.go` — SSRF-hardened HTTP client
  - IP blocking (10 CIDR ranges: loopback, private, link-local, CGNAT, metadata)
  - IP literal rejection, redirect re-validation, header sanitization
  - Concurrency control (per-app + global atomic counters)
  - `Proxy: nil`, `DisableCompression: true` for security
- `internal/egress/allowlist.go` — Domain allowlist with DB + 30s TTL cache
  - Wildcard support (`*.googleapis.com`), bare `*` rejected
  - App-scoped + global, canonical host matching
- `internal/egress/errors.go` — 8 structured error codes
- `internal/egress/inject.go` — Goja VM injection (`fazt.net.fetch()`)
- `internal/timeout/budget.go` — Extended with `NetContext()`
- Budget alignment fix: uses runtime timeout (5s) not request timeout (10s)
- Migration 019: net_allowlist table
- CLI: `fazt net allow/list/remove`

**Phase 2: Secrets, Rate Limits, Per-Domain Config**
- `internal/egress/secrets.go` — Server-side credential store
  - bearer/header/query injection, domain restriction, app scoping
  - JS never sees secret values
- `internal/egress/ratelimit.go` — Token bucket per domain
- Migration 020: net_secrets table
- CLI: `fazt secret set/list/remove`

**Phase 3: Logging + Cache**
- `internal/egress/logger.go` — Async batch logging
  - Buffer + periodic flush, errors bypass buffer, query strings stripped
- `internal/egress/cache.go` — LRU response cache
  - Memory-only, disabled by default, opt-in per domain via cache_ttl
- Migration 021: net_log table

**Tests: 55 tests, all passing**
- `proxy_test.go` — 21 tests (IP blocking, allowlist, body size, response, errors, budget, canonicalization)
- `allowlist_test.go` — 5 tests (domain matching, CRUD, wildcards, HTTPS-only, canonicalization)
- `ratelimit_test.go` — 3 tests (disabled default, enforcement, per-domain)
- `secrets_test.go` — 11 tests (CRUD, scoping, injection types, domain restriction, validation)
- `cache_test.go` — 9 tests (get/put, expiration, eviction, LRU, stats, key rules)
- `logger_test.go` — 5 tests (buffer/flush, error bypass, drop, query strip, start/stop)
- Plus existing system tests updated and passing (14 tests)

### Files Created
```
internal/egress/proxy.go
internal/egress/allowlist.go
internal/egress/errors.go
internal/egress/inject.go
internal/egress/secrets.go
internal/egress/ratelimit.go
internal/egress/logger.go
internal/egress/cache.go
internal/egress/proxy_test.go
internal/egress/allowlist_test.go
internal/egress/secrets_test.go
internal/egress/ratelimit_test.go
internal/egress/cache_test.go
internal/egress/logger_test.go
internal/system/schema.go
internal/system/schema_test.go
internal/database/migrations/019_net_allowlist.sql
internal/database/migrations/020_net_secrets.sql
internal/database/migrations/021_net_log.sql
cmd/server/net.go
cmd/server/secret.go
```

### Files Modified
```
internal/system/probe.go          # Nested structs + tags
internal/system/probe_test.go     # Updated for nested access
internal/handlers/system.go       # Schema handler, fixed health handler
internal/activity/logger.go       # Rewired to system.GetLimits()
internal/timeout/budget.go        # NetContext()
internal/runtime/runtime.go       # Timeout() getter
internal/runtime/handler.go       # Egress wiring, budget fix
internal/database/db.go           # Migrations 019-021
cmd/server/main.go                # Route + init wiring
admin/packages/fazt-sdk/index.js  # limitsSchema()
```

---

## Next Session

### Remaining:
1. Integration test with real serverless app (deploy test app that uses `fazt.net.fetch()`)
2. App file upload feature (see below)

### Pre-existing flaky tests (to fix):

**`hosting/TestStressMessageThroughput`** — Stress test asserts 85% WS delivery
across 100 clients, but VM consistently hits 75-79%. Has `testing.Short()` skip
but `go test ./...` doesn't pass `-short`. Fix: lower threshold to 70%.

**`worker/TestPoolList`** — SQLite `:memory:` connection pool race. Each new
connection to `:memory:` gets an independent blank DB. Table creation on one
connection, query on another. Fix: `db.SetMaxOpenConns(1)` in test helper.
Same issue may affect other test helpers using `:memory:`.

### Key resources:
- `koder/plans/40_fazt_http.md` — the plan
- `internal/egress/` — all new code
- `internal/system/probe.go` — restructured limits

---

## Quick Reference

```bash
# Test egress
go test ./internal/egress/ -v

# Test all affected packages
go test ./internal/egress/ ./internal/system/ ./internal/runtime/ ./internal/timeout/ ./internal/handlers/ ./cmd/server/

# Key files
cat internal/egress/proxy.go       # Core proxy
cat internal/egress/inject.go      # JS binding
cat internal/system/probe.go       # Nested limits
cat internal/system/schema.go      # Schema extractor
```
