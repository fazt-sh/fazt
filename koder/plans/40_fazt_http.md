# Plan 40: fazt.http — External HTTP from Serverless

**Created**: 2026-02-06
**Status**: DRAFT
**Goal**: Allow serverless JS to make outbound HTTP requests through a secure egress proxy
**Context**: ai-chat thread `01_fazt-http` (messages 01-07, Claude + Codex)
**Reviewed by**: Codex (messages 05-06), fixes applied from review

---

## Overview

Today serverless functions can't make network calls. `fazt.net.fetch()` adds a
secure egress proxy that lets JS call external APIs (Stripe, OpenAI, webhooks)
while blocking SSRF, limiting resource usage, and keeping secrets out of JS.

This unlocks fazt as a "glue" platform — not just static hosting + storage, but
a real application backend that can integrate with the outside world.

---

## Phases

| Phase | Scope | Status |
|-------|-------|--------|
| **1 — Safe MVP** | fetch + allowlist + IP blocking + budget | This plan |
| **2 — Hardening** | Secrets, rate limits, per-domain config | Sketched below |
| **3 — Observability** | SQLite logging, response cache | Sketched below |

Each phase is independently shippable. Phase 1 is fully specified. Phases 2-3
are documented enough to ensure Phase 1 design doesn't block them.

---

## Phase 1 — Safe MVP

### What ships

1. `fazt.net.fetch(url, options)` — sync, returns response object
2. Allowlist enforcement — domain-level, wildcard support
3. IP blocking — connect-time validation via custom `DialContext`
4. IP literal blocking — reject URLs with raw IPs before DNS
5. Redirect validation — re-check host + scheme at each hop
6. HTTPS only — HTTP rejected unless domain explicitly allows it
7. NetBudget — call count + time limits
8. Response size cap — `io.LimitReader`, 1MB default
9. Stdout logging only — zero SQLite write pressure
10. CLI: `fazt net allow`, `fazt net list`, `fazt net remove`

### What does NOT ship

- Secrets table / auth injection (Phase 2)
- Rate limiting per-domain (Phase 2)
- SQLite logging (Phase 3)
- Response cache (Phase 3)

---

### Security Model

#### SSRF Protection

The Goja VM runs in the same Go process as fazt. A naive `http.Get()` binding
would let JS reach `127.0.0.1:8080` (fazt itself), `192.168.*.*` (LAN),
`169.254.169.254` (cloud metadata), etc.

**Defense: custom `DialContext` with IP validation at connect time.**

After DNS resolves but before TCP connect, check the resolved IP against
blocked ranges. This runs on every new connection — no TOCTOU race.

```go
// Blocked IP ranges
var blockedNets = []net.IPNet{
    parseCIDR("127.0.0.0/8"),       // Loopback
    parseCIDR("10.0.0.0/8"),        // Private (A)
    parseCIDR("172.16.0.0/12"),     // Private (B)
    parseCIDR("192.168.0.0/16"),    // Private (C)
    parseCIDR("169.254.0.0/16"),    // Link-local / metadata
    parseCIDR("100.64.0.0/10"),     // CGNAT
    parseCIDR("0.0.0.0/8"),         // "This network"
    parseCIDR("::1/128"),           // IPv6 loopback
    parseCIDR("fc00::/7"),          // IPv6 unique-local
    parseCIDR("fe80::/10"),         // IPv6 link-local
}
```

**IP literal blocking**: If the URL hostname is a raw IP (`https://127.0.0.1/`,
`https://[::1]/`), reject before DNS — the allowlist operates on domain names.

#### Redirect Protection

Go's `http.Client` follows redirects automatically. A `302` from an allowed
host to `http://127.0.0.1:8080/api/secret` would bypass the allowlist.

**Defense: `CheckRedirect` function that re-validates at each hop.**

- Is the new hostname in the allowlist?
- Is it still HTTPS?
- Is the redirect count within limits?
- IP validation still fires at `DialContext` (defense in depth)

#### Allowlist

Strict allowlist-only. No denylist, no open mode. Single-user sovereign compute
means you know exactly which APIs you're calling.

```sql
-- Migration 019
CREATE TABLE net_allowlist (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    domain TEXT NOT NULL,         -- "api.stripe.com" or "*.googleapis.com"
    app_id TEXT,                  -- NULL = global, else scoped to app
    https_only INTEGER DEFAULT 1, -- 1 = HTTPS required, 0 = HTTP allowed
    created_at INTEGER DEFAULT (unixepoch()),
    UNIQUE(domain, app_id)
);
```

Matching rules:
- Exact: `api.stripe.com` matches only `api.stripe.com`
- Wildcard: `*.googleapis.com` matches `maps.googleapis.com`, `oauth2.googleapis.com`
- No bare wildcards (`*`) — that's denylist mode in disguise

**Host canonicalization** (applied to both URL and allowlist entries):
- Lowercase: `API.Stripe.com` → `api.stripe.com`
- Strip trailing dot: `api.stripe.com.` → `api.stripe.com`
- Strip port: `api.stripe.com:443` → `api.stripe.com`

```go
func canonicalizeHost(raw string) string {
    host := strings.ToLower(raw)
    host = strings.TrimSuffix(host, ".")
    if h, _, err := net.SplitHostPort(host); err == nil {
        host = h
    }
    return host
}
```

---

### Budget & Limits

| Limit | Default | Rationale |
|-------|---------|-----------|
| Max calls per request | 5 | auth + API + webhook patterns |
| Per-call timeout | 4s | Leaves 1s for storage in 5s JS budget |
| Total HTTP budget per request | 4s | Shared across all calls, not per-call |
| Per-app concurrent outbound | 5 | Limits blast radius of one app |
| Global concurrent outbound | 20 | 20% of 100 VM pool = bounded |
| Max request body (outgoing) | 1MB | API payloads are small |
| Max response body | 1MB | 99% of API responses |
| Max response body (hard cap) | 10MB | Per-app configurable |
| Max redirects | 3 | Standard practice |

**NetBudget integrates with existing timeout system:**

The existing `timeout.Budget` tracks remaining time and provides `StorageContext()`
for storage operations. NetBudget follows the same pattern:

```
Existing:  budget.StorageContext(ctx) → scoped context for one storage op
New:       budget.NetContext(ctx)     → scoped context for one HTTP call
```

Both use admission control: check remaining time, allocate a fraction, return
`ErrInsufficientTime` if not enough.

**Timeout architecture — budget aligned with runtime (fix from review):**

Currently `handler.go:108` creates a 10s budget but Goja interrupts at 5s
(`runtime.go:498`). This means storage/net admission control sees 10s of runway
when only 5s exists. This is a pre-existing bug we fix here.

```
0s ─────── JS starts ──────── 5s (Goja interrupt)
           │                    │
           ├─ kv.set (5ms)      │  ← budget.Remaining() now tracks THIS window
           ├─ fetch (2s)        │
           ├─ kv.set (5ms)      │
           ├─ fetch (1.5s)      │
           └─ respond()         │
                                └── post-exec (logs, response) ── 10s

Budget sees 5s (correct).  Handler context is still 10s (for post-exec work).
```

**Fix in `handler.go`:**
```go
cfg := timeout.DefaultConfig()
execCtx, cancel := context.WithTimeout(ctx, cfg.RequestTimeout)  // 10s for full request
defer cancel()

// Budget tracks the JS execution window (5s), not the full request (10s)
budgetCtx, _ := context.WithTimeout(ctx, h.runtime.Timeout())  // 5s
budget := timeout.NewBudget(budgetCtx, cfg)
```

**Prerequisite:** Add `Timeout()` getter to `internal/runtime/runtime.go`:
```go
func (r *Runtime) Timeout() time.Duration {
    return r.timeout
}
```
(The `timeout` field is currently private with no accessor.)

`DefaultTimeout` stays at 5s. `RequestTimeout` stays at 10s. Budget now
correctly reflects the 5s VM lifetime.

---

### JS API

```javascript
var res = fazt.net.fetch("https://api.example.com/data", {
    method: "GET",                        // default: GET
    headers: { "Accept": "application/json" },
    body: '{"key": "value"}',             // string only in Phase 1
    timeout: 3000                         // ms, optional, capped by NetBudget
});

res.status      // 200
res.ok          // true (status 200-299)
res.headers     // { "content-type": "application/json", ... }
                // Keys lowercased, first value only (multi-value → first)
res.text()      // response body as string
res.json()      // parsed JSON (throws on invalid JSON)
```

**Error handling:**

```javascript
var res = fazt.net.fetch("https://api.example.com/data");
if (!res.ok) {
    respond({ status: 502, body: { error: "upstream returned " + res.status } });
    return;
}
```

Errors that prevent a response (network failure, timeout, blocked) throw a JS
error with structured error types:

**Retryable errors** (→ 503 + Retry-After):
- Concurrency limit hit (per-app or global)
- NetBudget insufficient time

**Non-retryable errors** (→ 500):
- Upstream timeout (the external service is slow)
- Allowlist rejected (config issue)
- IP blocked (security block)
- Request body too large

```go
// internal/egress/errors.go
type EgressError struct {
    Code      string  // Stable code for JS error handling
    Message   string
    Retryable bool
}

// Error codes:
//   NET_BLOCKED  — allowlist/IP rejected (not retryable)
//   NET_TIMEOUT  — upstream timeout (not retryable)
//   NET_LIMIT    — concurrency limit hit (retryable)
//   NET_BUDGET   — insufficient time budget (retryable)
//   NET_SIZE     — request/response body too large (not retryable)
//   NET_ERROR    — other network error (not retryable)

func IsRetryableError(err error) bool { ... }
```

JS-side error handling:
```javascript
try {
    var res = fazt.net.fetch(url);
} catch (e) {
    if (e.code === "NET_TIMEOUT") { /* upstream slow */ }
    if (e.code === "NET_BLOCKED") { /* check allowlist config */ }
}
```

Handler integration (`handler.go`): add `egress.IsRetryableError(result.Error)`
alongside existing `storage.IsRetryableError()` check.

**Return type**: Direct response object (sync). Not a Promise. "No backward
compatibility" means we can change the return type later when we add a real
async runtime.

---

### File Changes

#### New files

```
internal/egress/proxy.go       # EgressProxy: http.Client, DialContext, IP validation
internal/egress/allowlist.go   # Allowlist: load, match, wildcard, canonicalization
internal/egress/response.go    # JSResponse: the object returned to JS
internal/egress/errors.go      # EgressError with Retryable flag
internal/egress/proxy_test.go  # SSRF, redirect, size, timeout tests
internal/egress/allowlist_test.go

internal/database/migrations/019_net_allowlist.sql

cmd/server/net.go              # CLI: fazt net allow/list/remove
```

#### Modified files

```
internal/runtime/fazt.go       # Add fazt.net namespace with fetch binding
internal/runtime/handler.go    # Create EgressProxy, add net injector to chain
internal/timeout/budget.go     # Add NetBudget fields to Config + NetContext()
internal/database/db.go        # Register migration 019
cmd/server/main.go             # Add "net" case to command switch
```

---

### Implementation Detail

#### 1. `internal/egress/proxy.go`

Core type that owns the `http.Client` with security hardening:

```go
type EgressProxy struct {
    client     *http.Client
    allowlist  *Allowlist
    callLimit  int           // Max calls per request (default 5)
    maxBody    int64         // Max response body (default 1MB)
    mu         sync.Mutex    // Protects concurrent counters
    appConns   map[string]int32  // Per-app concurrent count
    globalConns int32        // Global concurrent count
}

type FetchOptions struct {
    Method  string
    Headers map[string]string
    Body    string
    Timeout time.Duration
}

type FetchResponse struct {
    Status  int
    OK      bool
    Headers map[string]string
    body    []byte  // Read once, stored
}

func (r *FetchResponse) Text() string { ... }
func (r *FetchResponse) JSON() (interface{}, error) { ... }
```

**JS binding** (explicit lowercase method names on Goja object):

```go
func responseToJS(vm *goja.Runtime, resp *FetchResponse) goja.Value {
    obj := vm.NewObject()
    obj.Set("status", resp.Status)
    obj.Set("ok", resp.OK)
    obj.Set("headers", resp.Headers)
    obj.Set("text", func(call goja.FunctionCall) goja.Value {
        return vm.ToValue(resp.Text())
    })
    obj.Set("json", func(call goja.FunctionCall) goja.Value {
        data, err := resp.JSON()
        if err != nil {
            panic(vm.NewGoError(err))
        }
        return vm.ToValue(data)
    })
    return obj
}
```

The `http.Client` is configured with a hardened `Transport`:

```go
transport := &http.Transport{
    Proxy:                  nil,   // CRITICAL: ignore HTTP_PROXY/HTTPS_PROXY env
    DialContext:            safeDialer.DialContext,
    DisableCompression:     true,  // Raw bodies so LimitReader is accurate
    TLSHandshakeTimeout:   5 * time.Second,
    ResponseHeaderTimeout: 5 * time.Second,
    ExpectContinueTimeout: 1 * time.Second,
    MaxResponseHeaderBytes: 1 << 20, // 1MB header limit
    // Connection pooling: enabled with strict limits
    MaxIdleConns:          20,    // Match global concurrent limit
    MaxIdleConnsPerHost:   2,     // Low — we don't hammer single hosts
    IdleConnTimeout:       10 * time.Second,  // Short idle life
    MaxConnsPerHost:       0,     // Unlimited — our per-app/global counters are the real limit
}
```

Plus `CheckRedirect` for redirect validation.

**Compression determinism:** In addition to `DisableCompression: true`, set
`Accept-Encoding: identity` on all outbound requests in `proxy.Fetch()` before
dispatch. This prevents servers that ignore the transport setting from sending
compressed responses. User-set `Accept-Encoding` headers are overridden — we
control the transport layer.

**Connection pooling is safe** because allowlist + IP validation run on every
request (not per-connection). A reused connection to `api.stripe.com` was already
allowlisted. Strict pool limits prevent resource accumulation.

**Request body limit:** Check `len(opts.Body) > maxRequestBody` before dispatch.
Reject with `NET_SIZE` error if exceeded.

**Header sanitization:** Strip unsafe headers from user-provided options before
dispatch. Blocked headers: `Host` (set from URL), `Connection`,
`Proxy-Authorization`, `Proxy-Connection`, `Transfer-Encoding`,
`Accept-Encoding` (forced to `identity`).

**Concurrency control:**
- Per-app: `sync/atomic` counter, checked before each call
- Global: same pattern, separate counter
- Both reject with retryable error if at limit

#### 2. `internal/egress/allowlist.go`

```go
type Allowlist struct {
    db       *sql.DB
    cache    map[string][]AllowlistEntry  // keyed by appID ("" = global)
    mu       sync.RWMutex
    loadedAt time.Time
    ttl      time.Duration  // 30s default
}

func (a *Allowlist) IsAllowed(domain string, appID string) bool
func (a *Allowlist) Add(domain string, appID string, httpsOnly bool) error
func (a *Allowlist) Remove(domain string, appID string) error
func (a *Allowlist) List(appID string) ([]AllowlistEntry, error)
```

`IsAllowed` checks in-memory cache first (30s TTL), falls back to DB reload.
Cache invalidated on `Add`/`Remove` mutations. Checks both app-scoped and
global entries. Wildcard matching: `*.example.com` matches `sub.example.com`
but not `example.com` itself.

#### 3. Runtime injection

Follows the existing injector pattern from `handler.go:202-263`:

```go
// In handler.go executeWithFazt():
netInjector := func(vm *goja.Runtime) error {
    if egressProxy != nil {
        return egress.InjectNetNamespace(vm, egressProxy, app.ID, execCtx, budget)
    }
    return nil
}

// Added to ExecuteWithInjectors call
return h.runtime.ExecuteWithInjectors(ctx, code, req, loader,
    faztInjector, storageInjector, appStorageInjector,
    realtimeInjector, workerInjector, authInjector, privateInjector,
    netInjector)  // ← new
```

The injection function in `egress/` (not in `runtime/fazt.go` — keeps egress
self-contained):

```go
func InjectNetNamespace(vm *goja.Runtime, proxy *EgressProxy, appID string,
    ctx context.Context, budget *timeout.Budget) error {

    netObj := vm.NewObject()
    callCount := 0

    netObj.Set("fetch", func(call goja.FunctionCall) goja.Value {
        callCount++
        if callCount > proxy.callLimit {
            panic(vm.NewGoError(fmt.Errorf("fetch limit exceeded (%d calls)", proxy.callLimit)))
        }

        url := call.Argument(0).String()
        opts := parseOptions(vm, call)

        // Get net context from budget
        netCtx, cancel, err := budget.NetContext(ctx)
        if err != nil {
            panic(vm.NewGoError(err))
        }
        defer cancel()

        resp, err := proxy.Fetch(netCtx, appID, url, opts)
        if err != nil {
            panic(vm.NewGoError(err))
        }

        return responseToJS(vm, resp)
    })

    // Get existing fazt object and add net namespace
    fazt := vm.Get("fazt").ToObject(vm)
    fazt.Set("net", netObj)
    return nil
}
```

#### 4. Budget extension

Add to `timeout/budget.go`:

```go
// In Config:
NetCallTimeout   time.Duration // Max time for single HTTP call (default: 4s)
MinNetTime       time.Duration // Min time to start a net call (default: 1s)

// In DefaultConfig():
NetCallTimeout:   4 * time.Second,
MinNetTime:       1 * time.Second,

// New method on Budget:
func (b *Budget) NetContext(parent context.Context) (context.Context, context.CancelFunc, error) {
    if b == nil {
        // Workers don't have budgets — no net calls allowed
        return nil, nil, fmt.Errorf("net calls not available in this context")
    }
    remaining := b.Remaining()
    if remaining < b.config.MinNetTime {
        return nil, nil, ErrInsufficientTime
    }
    opTimeout := remaining - 500*time.Millisecond // Reserve 500ms for post-call work
    if opTimeout > b.config.NetCallTimeout {
        opTimeout = b.config.NetCallTimeout
    }
    ctx, cancel := context.WithTimeout(parent, opTimeout)
    return ctx, cancel, nil
}
```

#### 5. Migration 019

```sql
-- 019_net_allowlist.sql
CREATE TABLE IF NOT EXISTS net_allowlist (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    domain TEXT NOT NULL,
    app_id TEXT,
    https_only INTEGER NOT NULL DEFAULT 1,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    UNIQUE(domain, app_id)
);
```

#### 6. CLI (`cmd/server/net.go`)

```
fazt net allow api.stripe.com              # Global allowlist
fazt net allow api.openai.com --app myapp  # Per-app
fazt net allow "*.googleapis.com"          # Wildcard
fazt net list                              # Show all entries
fazt net list --app myapp                  # Show app-scoped entries
fazt net remove api.old-service.com        # Remove entry
```

Follows existing CLI pattern: `handleNetCommand(args)` → switch on subcommand
→ `flag.NewFlagSet` for flags → `getClientDB()` for database access.

---

### Test Matrix

| # | Test | Layer | Validates |
|---|------|-------|-----------|
| 1 | Fetch public HTTPS URL | Integration | Happy path |
| 2 | Fetch `http://127.0.0.1:8080` | Proxy | IP blocked at DialContext |
| 3 | Fetch `http://192.168.1.1` | Proxy | Private IP blocked |
| 4 | Fetch `http://169.254.169.254` | Proxy | Metadata IP blocked |
| 5 | Fetch `https://[::1]/` | Proxy | IPv6 loopback blocked |
| 6 | URL with IP literal (no DNS) | Proxy | Rejected before dial |
| 7 | Non-allowlisted domain | Allowlist | Domain check works |
| 8 | Wildcard allowlist match | Allowlist | `*.example.com` matches `sub.example.com` |
| 9 | Wildcard doesn't match bare | Allowlist | `*.example.com` rejects `example.com` |
| 10 | Redirect to private IP | Proxy | CheckRedirect + DialContext blocks |
| 11 | Redirect to HTTP from HTTPS | Proxy | Scheme downgrade blocked |
| 12 | Redirect chain > 3 hops | Proxy | Max redirects enforced |
| 13 | Response > 1MB | Proxy | Truncated by LimitReader, error returned |
| 14 | 6th fetch in one request | Budget | Call limit enforced |
| 15 | Fetch with <1s budget remaining | Budget | ErrInsufficientTime |
| 16 | Per-app concurrency at limit | Proxy | 6th concurrent call rejected |
| 17 | Global concurrency at limit | Proxy | 21st concurrent call rejected |
| 18 | DNS resolves to private IP | Proxy | Blocked at connect time (not at DNS) |
| 19 | Timeout mid-transfer | Proxy | Context cancellation cleans up |
| 20 | Valid JSON response + `.json()` | Response | Parse works |
| 21 | Invalid JSON + `.json()` | Response | Throws JS error |
| 22 | `HTTP_PROXY` env set | Proxy | Proxy env ignored |
| 23 | Compressed response (gzip) | Proxy | Raw body received (compression disabled) |
| 24 | `API.Stripe.com` uppercase | Allowlist | Canonicalization matches |
| 25 | `api.stripe.com.` trailing dot | Allowlist | Canonicalization matches |
| 26 | `api.stripe.com:443` with port | Allowlist | Canonicalization matches |
| 27 | `res.text()` / `res.json()` names | Response | JS lowercase methods work |
| 28 | Request body > 1MB | Proxy | Rejected before dial |

---

### Implementation Order

```
Step 1: internal/egress/proxy.go + proxy_test.go
        Core proxy with IP validation, redirect checking, size limits.
        Tests 2-6, 10-13, 18-19.

Step 2: internal/egress/allowlist.go + allowlist_test.go + migration 019
        Domain matching, wildcard support, database CRUD.
        Tests 7-9.

Step 3: internal/timeout/budget.go
        Add NetContext(), NetCallTimeout, MinNetTime to Config.
        Tests 14-15.

Step 4: internal/egress/response.go
        JSResponse with .text(), .json(), .ok, .status, .headers.
        Tests 20-21.

Step 5: internal/runtime/fazt.go + internal/runtime/handler.go
        Wire egress proxy into fazt.net namespace via injector chain.
        Test 1 (integration).

Step 6: cmd/server/net.go + cmd/server/main.go
        CLI commands for allowlist management.

Step 7: Integration test
        Full end-to-end: deploy app with fetch, run against real server.
        Tests 16-17 (concurrency).

Step 8: Docs update
        Remove "no network calls" from serverless references:
        - knowledge-base/skills/app/references/serverless-api.md
        - Any workflow docs that repeat the limitation
```

---

## Phase 2 — Hardening (Sketch)

### Secrets store

New table, separate from `env_vars`:

```sql
CREATE TABLE secrets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id TEXT,                -- NULL = global
    name TEXT NOT NULL,
    value TEXT NOT NULL,         -- Plaintext (DB file is security boundary)
    inject_as TEXT DEFAULT 'bearer', -- bearer, header, query
    inject_header TEXT,          -- Custom header name (if inject_as = 'header')
    domain TEXT,                 -- Only inject for this domain
    created_at INTEGER DEFAULT (unixepoch()),
    updated_at INTEGER DEFAULT (unixepoch()),
    UNIQUE(app_id, name)
);
```

Auth injection happens in the Go proxy — JS passes `auth: "STRIPE_KEY"`, the
proxy looks up the secret and injects the header. JS never sees the value.

CLI: `fazt secret set STRIPE_KEY sk_live_...` (value masked in output).

### Rate limits

Per-domain rate limits stored in a config table or new `net_rate_limits` table.
Enforced in the proxy with a token bucket per domain.

### Per-domain config

Extend `net_allowlist` or add `net_domains` table for domain-specific settings:
max response size, HTTP allowed, custom timeout, rate limit.

---

## Phase 3 — Observability (Sketch)

### Async batch logging

Separate from the WriteQueue to avoid storage pressure:

```go
type NetLogBuffer struct {
    entries []NetLogEntry
    mu      sync.Mutex
    ticker  *time.Ticker  // Flush every 1s
}
```

Buffer in memory, flush to SQLite in bulk INSERT every second. Errors always
logged immediately. High-volume domains sampled (log 10% after first 10).

### Response cache

Memory-first LRU with lazy SQLite persistence:

```
fetch() → check memory LRU → check SQLite cache → make request
response → store in memory LRU → async persist (if TTL > 5min)
```

Short-TTL responses (<5min) never touch SQLite. Cache key includes domain +
path + auth reference (so different apps don't share cached authenticated responses).

---

## Design Decisions Log

All decisions from ai-chat thread `01_fazt-http`:

| Decision | Choice | Why |
|----------|--------|-----|
| API shape | `fazt.net.fetch()` sync return | Simplest. No backward compat needed. |
| Promise in Phase 1 | No | Sync runtime. Goja drains microtasks but adds complexity for no gain. |
| Timeout | Keep 5s, add NetBudget | Don't give rogue scripts more VM time. |
| Allowlist mode | Strict only | Sovereign compute = you know your APIs. |
| Secrets encryption | Plaintext in SQLite | DB file is the security boundary. KMS is overkill for $6 VPS. |
| Logging Phase 1 | Stdout only | Zero SQLite write pressure. |
| Cache Phase 1 | None | Understand real usage patterns first. |
| Package location | `internal/egress/` | Self-contained, doesn't pollute runtime package. |
| Connection pooling | Enabled, unlimited per-host | Our per-app/global counters are the real limit. MaxConnsPerHost=0, MaxIdlePerHost=2, IdleTimeout=10s. |
| Worker access | No net calls | Workers don't have budgets. Net calls are request-scoped only. |
| Proxy env | `Transport.Proxy = nil` | Ignore HTTP_PROXY/HTTPS_PROXY — would bypass IP validation. |
| Compression | `DisableCompression: true` | Receive raw bodies so `io.LimitReader` is accurate. No gzip bombs. |
| Budget alignment | Budget uses 5s (runtime), not 10s (request) | Pre-existing bug fixed. Admission control now matches actual VM lifetime. |
| Host canonicalization | Lowercase + strip dot + strip port | Prevents allowlist bypass via case/formatting tricks. IDNA deferred. |
| Allowlist caching | In-memory, 30s TTL, invalidate on mutation | Avoids SQLite read on every fetch call. |
| Error codes | `NET_BLOCKED`, `NET_TIMEOUT`, `NET_LIMIT`, etc. | Stable codes for JS error branching. |
| Header sanitization | Strip Host, Connection, Proxy-*, Transfer-Encoding | Prevents subtle upstream bypass/breakage. |
| Response headers | Lowercase keys, first value only | Matches existing request header pattern in handler.go. |
