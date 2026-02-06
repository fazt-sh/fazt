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
| **0 — Limits refactor** | Nested `system.Limits` with struct tags + schema | Specified below |
| **1 — Safe MVP** | fetch + allowlist + IP blocking + budget | Specified below |
| **2 — Hardening** | Secrets, rate limits, per-domain config | Specified below |
| **3 — Observability** | SQLite logging, response cache | Specified below |

Each phase is independently shippable. All phases are fully specified with thin
layers and sane defaults. Limits for all phases are declared in `system.Limits.Net`
from Step 0 — Phase 2/3 fields ship with defaults that mean "disabled" until
those features activate.

---

## Step 0 — Refactor `system.Limits` (prerequisite)

Before egress work, restructure the flat `system.Limits` into nested, tagged
structs. This is a standalone refactor that benefits the whole system.

### Why now

- Current: 10 flat fields, about to add 8 more → 18 flat `LongPrefixedNames`
- `SystemCapacityHandler` manually builds nested maps from flat struct
- Limits/capacity endpoints not documented in knowledge-base
- Adding struct tags now means validation + config + admin UI come free later

### New structure

```go
// internal/system/limits.go (rename from probe.go — it's more than probing now)

// Limits describes all system resource limits.
// Struct tags provide metadata for API schema, admin UI, and future validation.
//
// Tags:
//   json     — JSON field name
//   label    — Human-readable label (for UI)
//   desc     — Short description (for UI tooltips / API docs)
//   unit     — Value unit: "bytes", "ms", "count" (for UI formatting)
//   range    — "min,max" accepted values (for UI sliders, future validation)
//   readonly — "true" if hardware-detected, not configurable
type Limits struct {
    Hardware Hardware `json:"hardware"`
    Storage  Storage  `json:"storage"`
    Runtime  Runtime  `json:"runtime"`
    Capacity Capacity `json:"capacity"`
    Net      Net      `json:"net"`
}

type Hardware struct {
    TotalRAM     int64 `json:"total_ram"      label:"Total RAM"      desc:"Detected system memory"   unit:"bytes" readonly:"true"`
    AvailableRAM int64 `json:"available_ram"   label:"Available RAM"  desc:"Estimated available memory" unit:"bytes" readonly:"true"`
    CPUCores     int   `json:"cpu_cores"       label:"CPU Cores"      desc:"Detected CPU cores"       readonly:"true"`
}

type Storage struct {
    MaxVFS      int64 `json:"max_vfs"       label:"VFS Cache"     desc:"Max VFS cache size"       unit:"bytes" range:"10485760,1073741824"`
    MaxUpload   int64 `json:"max_upload"     label:"Max Upload"    desc:"Max upload size"          unit:"bytes" range:"1048576,104857600"`
    WriteQueue  int   `json:"write_queue"    label:"Write Queue"   desc:"Max pending writes"       range:"100,10000"`
    MaxFileSize int64 `json:"max_file_size"  label:"Max File"      desc:"Max single file size"     unit:"bytes" range:"1048576,1073741824"`
    MaxSiteSize int64 `json:"max_site_size"  label:"Max Site"      desc:"Max total site size"      unit:"bytes" range:"10485760,5368709120"`
    MaxLogRows  int   `json:"max_log_rows"   label:"Max Log Rows"  desc:"Activity log row limit"   range:"10000,5000000"`
}

type Runtime struct {
    ExecTimeout int   `json:"exec_timeout" label:"Exec Timeout" desc:"Serverless execution timeout" unit:"ms" range:"100,10000"`
    MaxMemory   int64 `json:"max_memory"   label:"Max Memory"   desc:"Per-execution memory limit"   unit:"bytes" range:"1048576,268435456"`
}

type Capacity struct {
    Users       int `json:"users"        label:"Concurrent Users" desc:"Conservative estimate"    readonly:"true"`
    UsersMax    int `json:"users_max"    label:"Max Users"        desc:"Best-case estimate"       readonly:"true"`
    Reads       int `json:"reads"        label:"Read Throughput"  desc:"Requests/sec"             unit:"req/s" readonly:"true"`
    Writes      int `json:"writes"       label:"Write Throughput" desc:"Requests/sec"             unit:"req/s" readonly:"true"`
    Mixed       int `json:"mixed"        label:"Mixed Throughput" desc:"80% read / 20% write"     unit:"req/s" readonly:"true"`
    MaxRequests int `json:"max_requests" label:"Max Concurrent"   desc:"Max concurrent requests"  range:"50,5000"`
    Timeout     int `json:"timeout_ms"   label:"Request Timeout"  desc:"Request timeout"          unit:"ms" range:"1000,30000"`
}

type Net struct {
    // Phase 1 — Egress core
    MaxCalls       int   `json:"max_calls"       label:"Max Calls"       desc:"Fetch calls per request"    range:"1,20"`
    CallTimeout    int   `json:"call_timeout"    label:"Call Timeout"    desc:"Per-call timeout"           unit:"ms" range:"1000,10000"`
    Budget         int   `json:"budget"          label:"HTTP Budget"     desc:"Total HTTP time per request" unit:"ms" range:"1000,10000"`
    AppConcurrency int   `json:"app_concurrency" label:"App Concurrency" desc:"Per-app concurrent outbound" range:"1,20"`
    Concurrency    int   `json:"concurrency"     label:"Concurrency"     desc:"Global concurrent outbound"  range:"5,100"`
    MaxRequestBody int64 `json:"max_req_body"    label:"Max Request"     desc:"Outgoing body size limit"   unit:"bytes" range:"1024,10485760"`
    MaxResponse    int64 `json:"max_response"    label:"Max Response"    desc:"Response body size limit"   unit:"bytes" range:"1024,10485760"`
    MaxRedirects   int   `json:"max_redirects"   label:"Max Redirects"   desc:"Redirect hop limit"         range:"0,10"`

    // Phase 2 — Rate limiting
    RateLimit      int   `json:"rate_limit"      label:"Rate Limit"      desc:"Default requests/min per domain" range:"0,1000"`
    RateBurst      int   `json:"rate_burst"       label:"Rate Burst"      desc:"Burst allowance above rate"     range:"0,100"`

    // Phase 3 — Observability
    LogBufferSize  int   `json:"log_buffer"      label:"Log Buffer"      desc:"In-memory log entries before flush" range:"100,10000"`
    LogFlushMs     int   `json:"log_flush"       label:"Log Flush"       desc:"Flush interval"              unit:"ms" range:"500,10000"`
    CacheMaxItems  int   `json:"cache_max_items" label:"Cache Items"     desc:"Max cached responses"        range:"0,10000"`
    CacheMaxBytes  int64 `json:"cache_max_bytes" label:"Cache Size"      desc:"Max cache memory"            unit:"bytes" range:"0,104857600"`
}
```

### Schema endpoint

A small reflect-based function walks the struct tags and returns metadata for
the admin UI. ~30 lines of code. Schema is built once via `sync.Once` and
cached — subsequent requests return the pre-built JSON. Skip unexported fields
and `json:"-"` tags; always key by JSON tag name when present. Two endpoints:

```
GET /api/system/limits          → values (clean JSON, for programmatic use)
GET /api/system/limits/schema   → labels, descriptions, ranges, units (for UI)
```

Schema response shape:
```json
{
  "hardware": {
    "total_ram": {
      "label": "Total RAM",
      "desc": "Detected system memory",
      "unit": "bytes",
      "read_only": true
    }
  },
  "net": {
    "max_calls": {
      "label": "Max Calls",
      "desc": "Fetch calls per request",
      "min": 1, "max": 20,
      "read_only": false
    }
  }
}
```

### Future extensibility (not built now, but enabled)

- **Validation**: `Validate(limits)` reads `range` tags, checks values. ~20 lines.
- **Config system**: Add `configurable:"true"` tag. Settings endpoint accepts
  updates for configurable fields, validates against `range`, writes to
  `configurations` table. Read-only fields stay hardware-detected.
- **Per-app overrides**: `GetLimitsForApp(appID)` merges app overrides on top
  of system defaults. `range` tags constrain what overrides are allowed.

### Files changed

```
internal/system/probe.go        → Refactor into nested structs with tags
                                  Add Runtime sub-struct (ExecTimeout, MaxMemory)
internal/system/schema.go       → New: reflect-based schema extractor (sync.Once cached)
internal/system/schema_test.go  → New: test tag extraction
internal/handlers/system.go     → Simplify: remove manual map-building,
                                  merge limits+capacity into one handler,
                                  add /schema endpoint
internal/capacity/              → DELETE: all fields absorbed into system.Limits
                                  (Storage.WriteQueue, Runtime.*, Capacity.*)
internal/activity/logger.go     → Rewire: capacity.DefaultLimits() → system.GetLimits()
                                  (line 66: only consumer of capacity package)
admin/packages/fazt-sdk/index.js → Update: /api/system/capacity → /api/system/limits
                                   (line 127: only SDK consumer)
knowledge-base/agent-context/api.md → Document system endpoints, note API break
koder/CAPACITY.md               → Update references to removed endpoint
```

### Tests

| # | Test | Validates |
|---|------|-----------|
| S1 | Schema extracts all tags from nested struct | Reflection works |
| S2 | Range tags parse into min/max ints | UI gets valid numbers |
| S3 | ReadOnly fields marked correctly | UI knows what to gray out |
| S4 | `GET /api/system/limits` returns nested JSON | API shape correct |
| S5 | `GET /api/system/limits/schema` returns metadata | UI can build forms |

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

**Integration with `system.Limits`:** All net egress limits live in `system.Limits`
(not hardcoded in `egress/proxy.go`). This means they show up in
`GET /api/system/limits`, scale with hardware profile, and have a single source
of truth.

Net limits live in `system.Limits.Net` (see Step 0). Defaults set in
`GetLimits()`, global concurrency scales with CPU cores.

**`EgressProxy` reads from `system.Limits.Net` at construction time** — no
magic numbers in egress code:

```go
func NewEgressProxy(allowlist *Allowlist) *EgressProxy {
    net := system.GetLimits().Net
    return &EgressProxy{
        callLimit:    net.MaxCalls,
        maxReqBody:   net.MaxRequestBody,
        maxRespBody:  net.MaxResponse,
        maxRedirects: net.MaxRedirects,
        perAppLimit:  int32(net.AppConcurrency),
        globalLimit:  int32(net.Concurrency),
        // ...
    }
}
```

| Limit | Default | Scales with | Rationale |
|-------|---------|-------------|-----------|
| Max calls per request | 5 | — | auth + API + webhook patterns |
| Per-call timeout | 4s | — | Leaves 1s for storage in 5s JS budget |
| Total HTTP budget per request | 4s | — | Shared across all calls, not per-call |
| Per-app concurrent outbound | 5 | — | Limits blast radius of one app |
| Global concurrent outbound | 20 | CPU cores | 10× scaleFactor, bounded |
| Max request body (outgoing) | 1MB | — | API payloads are small |
| Max response body | 1MB | — | 99% of API responses |
| Max redirects | 3 | — | Standard practice |

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

#### New files (Step 0 — see above)

```
internal/system/schema.go      # Reflect-based struct tag schema extractor
internal/system/schema_test.go # Schema extraction tests
```

#### Modified files

```
internal/system/probe.go       # Refactored in Step 0 (nested structs + tags)
internal/handlers/system.go    # Simplified in Step 0 (merged endpoints, schema)
internal/runtime/fazt.go       # Add fazt.net namespace with fetch binding
internal/runtime/handler.go    # Create EgressProxy, add net injector to chain
internal/timeout/budget.go     # Add NetBudget fields to Config + NetContext()
internal/database/db.go        # Register migration 019
cmd/server/main.go             # Add "net" case to command switch
knowledge-base/agent-context/api.md  # Document system endpoints + API break
```

---

### Implementation Detail

#### 1. `internal/egress/proxy.go`

Core type that owns the `http.Client` with security hardening:

```go
type EgressProxy struct {
    client      *http.Client
    allowlist   *Allowlist
    callLimit   int           // From Limits.Net.MaxCalls
    maxReqBody  int64         // From Limits.Net.MaxRequestBody
    maxRespBody int64         // From Limits.Net.MaxResponse
    maxRedirects int          // From Limits.Net.MaxRedirects
    perAppLimit int32         // From Limits.Net.AppConcurrency
    globalLimit int32         // From Limits.Net.Concurrency
    mu          sync.Mutex    // Protects concurrent counters
    appConns    map[string]int32  // Per-app concurrent count
    globalConns int32         // Global concurrent count
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
Step 0: system.Limits refactor (prerequisite, standalone)
        Restructure flat Limits into nested structs with tags.
        Add schema extractor + /schema endpoint.
        Simplify handlers/system.go (merge limits+capacity).
        Update knowledge-base/agent-context/api.md.
        Tests S1-S5.

Step 1: internal/egress/proxy.go + proxy_test.go
        Core proxy with IP validation, redirect checking, size limits.
        Reads limits from system.GetLimits().Net.
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

## Phase 2 — Hardening

Ships as thin layers with sane defaults. Each feature is independently useful
and works with zero configuration out of the box.

### 2.1 Secrets store

JS never sees secret values. The Go proxy injects auth headers at dispatch time.

**Migration 020:**

```sql
CREATE TABLE IF NOT EXISTS net_secrets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id TEXT,                -- NULL = global
    name TEXT NOT NULL,         -- e.g. "STRIPE_KEY"
    value TEXT NOT NULL,        -- Plaintext (DB file = security boundary)
    inject_as TEXT NOT NULL DEFAULT 'bearer',  -- bearer | header | query
    inject_key TEXT,            -- Header name or query param (if not bearer)
    domain TEXT,                -- Only inject for requests to this domain
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
    UNIQUE(app_id, name)
);
```

**How injection works:**

```
JS:   fazt.net.fetch(url, { auth: "STRIPE_KEY" })
                                  │
Go proxy.Fetch():                 ▼
  1. Look up "STRIPE_KEY" in net_secrets (app-scoped first, then global)
  2. If not found → error NET_AUTH ("secret not found")
  3. Check domain match (if secret has domain restriction)
  4. Inject based on inject_as:
     - bearer: Authorization: Bearer <value>
     - header: <inject_key>: <value>  (e.g. X-API-Key)
     - query:  append ?<inject_key>=<value> to URL
  5. Dispatch request — JS never sees the value
```

**In-memory cache:** Secrets loaded once per request (not per-call). Same 30s
TTL + invalidate-on-mutation pattern as the allowlist.

**CLI:**

```bash
fazt secret set STRIPE_KEY sk_live_xxx              # Global, bearer (default)
fazt secret set OPENAI_KEY sk-xxx --as header --key "Authorization" --domain api.openai.com
fazt secret set WEBHOOK_TOKEN abc --as query --key "token" --app my-app
fazt secret list                                     # Shows names + domains, NOT values
fazt secret list --app my-app                        # App-scoped
fazt secret remove STRIPE_KEY                        # Remove
```

Output always masks values: `STRIPE_KEY = sk_l****xxx (bearer, global)`.

**New error code:** `NET_AUTH` — secret not found or domain mismatch (not retryable).

**Files:**

```
internal/egress/secrets.go          # SecretsStore: load, cache, inject
internal/egress/secrets_test.go     # Injection tests
internal/database/migrations/020_net_secrets.sql
cmd/server/secret.go                # CLI commands
```

### 2.2 Rate limiting

Token bucket per domain. Defaults from `system.Limits.Net.RateLimit` (0 = no
limit). Per-domain overrides in the allowlist table.

**Extend migration 019 (or add migration 021):**

```sql
ALTER TABLE net_allowlist ADD COLUMN rate_limit INTEGER DEFAULT 0;  -- req/min, 0 = use system default
ALTER TABLE net_allowlist ADD COLUMN rate_burst INTEGER DEFAULT 0;  -- 0 = use system default
```

**Implementation:**

```go
// internal/egress/ratelimit.go
type RateLimiter struct {
    buckets map[string]*tokenBucket  // keyed by domain
    mu      sync.RWMutex
    defaults struct {
        rate  int  // From Limits.Net.RateLimit
        burst int  // From Limits.Net.RateBurst
    }
}

func (rl *RateLimiter) Allow(domain string) bool
```

Token bucket is a simple struct: `tokens float64`, `lastRefill time.Time`,
`rate float64` (tokens/sec), `burst int`. No external dependencies. ~40 lines.

**Sane defaults:**
- `Limits.Net.RateLimit = 0` → no global rate limit (sovereign compute, you
  control your own traffic)
- Per-domain: set via `fazt net allow api.stripe.com --rate 60 --burst 10`
- Rate limiter only activates when rate > 0

**New error code:** `NET_RATE` — rate limited (retryable, includes Retry-After).

**Files:**

```
internal/egress/ratelimit.go       # Token bucket implementation
internal/egress/ratelimit_test.go  # Rate limiting tests
```

### 2.3 Per-domain config

Extend the allowlist table to hold domain-specific overrides. No new table —
the allowlist already scopes by domain + app.

**Extend migration 019 (or migration 021):**

```sql
ALTER TABLE net_allowlist ADD COLUMN max_response INTEGER DEFAULT 0;  -- 0 = use system default
ALTER TABLE net_allowlist ADD COLUMN timeout_ms INTEGER DEFAULT 0;    -- 0 = use system default
-- rate_limit and rate_burst added above
```

**How it works:** When `proxy.Fetch()` checks the allowlist, it also reads the
per-domain config. Domain-specific values override system defaults. Zero means
"inherit from `system.Limits.Net`".

```go
type DomainConfig struct {
    Domain      string
    HTTPSOnly   bool
    RateLimit   int    // 0 = inherit
    RateBurst   int    // 0 = inherit
    MaxResponse int64  // 0 = inherit
    Timeout     int    // 0 = inherit
}

func (a *Allowlist) ConfigFor(domain, appID string) *DomainConfig
```

**CLI extension:**

```bash
fazt net allow api.stripe.com --rate 60 --burst 10 --timeout 3000 --max-response 5242880
fazt net list  # Shows all config columns
```

---

## Phase 3 — Observability

Ships as thin layers. Logging is zero-config (just works). Cache is opt-in
(explicit CLI command to enable for a domain).

### 3.1 Async batch logging

Separate buffer from the WriteQueue. Zero impact on request latency.

**Migration 022:**

```sql
CREATE TABLE IF NOT EXISTS net_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id TEXT NOT NULL,
    domain TEXT NOT NULL,
    method TEXT NOT NULL,
    path TEXT NOT NULL,           -- URL path (no query string — may contain secrets)
    status INTEGER,               -- Response status, NULL if error
    error_code TEXT,              -- NET_TIMEOUT etc., NULL if success
    duration_ms INTEGER NOT NULL,
    request_bytes INTEGER,
    response_bytes INTEGER,
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE INDEX idx_net_log_app ON net_log(app_id, created_at);
CREATE INDEX idx_net_log_domain ON net_log(domain, created_at);
```

**Implementation:**

```go
// internal/egress/logger.go
type NetLogger struct {
    buffer []NetLogEntry
    mu     sync.Mutex
    db     *sql.DB
    done   chan struct{}
    // Config from Limits.Net
    bufferSize int  // Default: 1000
    flushMs    int  // Default: 1000
}

type NetLogEntry struct {
    AppID         string
    Domain        string
    Method        string
    Path          string
    Status        int
    ErrorCode     string
    DurationMs    int
    RequestBytes  int64
    ResponseBytes int64
    Timestamp     int64
}

func (l *NetLogger) Log(entry NetLogEntry)       // Non-blocking, drops if buffer full
func (l *NetLogger) Start()                       // Starts flush ticker
func (l *NetLogger) Stop()                        // Flushes remaining, stops ticker
```

**Behavior:**
- Buffer in memory, flush to SQLite via bulk INSERT every `Limits.Net.LogFlushMs`
- Errors (non-2xx, network failures) always logged immediately (bypass buffer)
- If buffer full, drop oldest entries (never block the request)
- Query string stripped from path (may contain secret tokens)
- Log retention: same as activity logs (`Limits.Storage.MaxLogRows` shared, or
  separate `net_log_max_rows` if needed)

**Sane defaults:**
- `Limits.Net.LogBufferSize = 1000`
- `Limits.Net.LogFlushMs = 1000`
- Logging enabled by default for all fetch calls (zero config)

**CLI:**

```bash
fazt net logs                              # Recent fetch logs
fazt net logs --app my-app                 # Per-app
fazt net logs --domain api.stripe.com      # Per-domain
fazt net logs --errors                     # Only failures
fazt net stats                             # Aggregate: calls/domain, avg latency, error rate
```

**Files:**

```
internal/egress/logger.go          # NetLogger buffer + flush
internal/egress/logger_test.go     # Buffer, flush, drop tests
internal/database/migrations/022_net_log.sql
cmd/server/net.go                  # Add logs/stats subcommands
```

### 3.2 Response cache

Memory-first LRU. Opt-in per domain. Never caches authenticated responses
unless explicitly configured.

**How it activates:** Only for domains that have `cache_ttl > 0` in their
allowlist config.

**Extend allowlist table (migration 021 or 023):**

```sql
ALTER TABLE net_allowlist ADD COLUMN cache_ttl INTEGER DEFAULT 0;  -- seconds, 0 = no cache
```

**Implementation:**

```go
// internal/egress/cache.go
type NetCache struct {
    items   map[string]*cacheEntry  // key = method + domain + path + query
    order   []string                // LRU order
    mu      sync.RWMutex
    maxItems int   // From Limits.Net.CacheMaxItems
    maxBytes int64 // From Limits.Net.CacheMaxBytes
    curBytes int64 // Current memory usage
}

type cacheEntry struct {
    response  *FetchResponse
    expiresAt time.Time
    size      int64  // Approximate memory footprint
}

func (c *NetCache) Get(key string) (*FetchResponse, bool)
func (c *NetCache) Put(key string, resp *FetchResponse, ttl time.Duration)
```

**Cache key rules:**
- Key = `method + domain + path + query` (query string included by default)
- Requests with `auth` option → NOT cached (prevents leaking between apps)
- Requests with `body` → NOT cached (POST/PUT are mutations)
- Only GET requests cached
- Per-domain TTL from allowlist config

**Fetch flow with cache:**

```
proxy.Fetch():
  1. Build cache key
  2. If cacheable (GET, no auth, domain has cache_ttl > 0):
     a. Check memory cache → hit? return cached response
     b. Miss → make request → store in cache with TTL
  3. If not cacheable → make request directly
```

**Eviction:** When `curBytes > maxBytes` or `len(items) > maxItems`, evict
oldest entries until under limit. Simple LRU, no background goroutine needed.

**Sane defaults:**
- `Limits.Net.CacheMaxItems = 0` → cache disabled by default
- `Limits.Net.CacheMaxBytes = 0` → cache disabled by default
- Enable per-domain: `fazt net allow api.example.com --cache-ttl 300`
- Enable globally: set `CacheMaxItems` and `CacheMaxBytes` in system limits

**CLI:**

```bash
fazt net allow api.example.com --cache-ttl 300   # Cache responses for 5 minutes
fazt net cache                                    # Show cache stats (items, size, hit rate)
fazt net cache clear                              # Clear all cached responses
```

**Files:**

```
internal/egress/cache.go           # LRU cache implementation
internal/egress/cache_test.go      # Cache hit/miss/eviction/TTL tests
```

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
| Limits integration | Nested `system.Limits` with struct tags | Single source of truth. Tags enable validation, config, admin UI without code changes. |
| Limits refactor | Step 0 prerequisite, standalone | Flat 18-field struct unacceptable. Nested + tagged is the foundation for all future subsystems. |
| Schema endpoint | `/api/system/limits/schema` via reflect | Admin UI builds forms from metadata. No hardcoded labels in frontend. |
| Limits API break | Merge limits+capacity, nested JSON | "No backward compatibility" — cleaner API, documented in knowledge-base. |
| Secrets storage | Plaintext in `net_secrets` table | DB file is security boundary. KMS overkill for $6 VPS. |
| Secret injection | Go proxy injects, JS never sees values | Defense in depth. Even XSS in serverless can't leak API keys. |
| Rate limit default | 0 (disabled) | Sovereign compute — user controls traffic. Opt-in per domain. |
| Rate limit impl | Token bucket per domain, ~40 lines | No external deps. Simple, correct, proven pattern. |
| Per-domain config | Extend allowlist table, not new table | Domain is already the natural key. Zero = inherit from system defaults. |
| Net logging | Async buffer, separate from WriteQueue | Zero request latency impact. Errors bypass buffer for immediate logging. |
| Log path stripping | Strip query string from logged paths | Query strings may contain secret tokens. |
| Cache activation | Opt-in per domain (cache_ttl > 0) | Caching API responses by default would be wrong. Explicit is better. |
| Cache key | method + domain + path + query, no auth requests | Prevents leaking authenticated responses between apps; avoids cache collisions. |
| Phase 2/3 limits | Declared in Step 0, defaults = disabled | Struct fields exist from day one. Features activate when defaults change. |
| Limits unification | Delete `internal/capacity/`, absorb into `system.Limits` | One source of truth. `capacity.Limits` duplicated fields that belong in system. |
| Runtime sub-struct | `Limits.Runtime` (ExecTimeout, MaxMemory) | Fields from `capacity.Limits` need a home. Runtime is the natural grouping. |
| Schema caching | `sync.Once` — build schema once, serve cached JSON | Reflection is fine but shouldn't run on every request. |
| Rate limit persistence | In-memory, reset on restart | Persisting bucket state adds write pressure for marginal gain. Sovereign compute — acceptable. |
| Cache persistence | Memory-only, no SQLite | SQLite persistence adds complexity. Users can build persistent caching in app code if needed. |

---

## Codex Review Notes (2026-02-06)

Comments on the Step 0 limits refactor and Phase 2/3 additions (ai-chat msg 13).

- **Schema endpoint**: Build once and cache (sync.Once). Skip unexported fields and `json:"-"` tags; always key by JSON tag when present.
- **Range tags**: Keep `range:"min,max"` for now; parse defensively and allow empty/missing values.
- **Limits unification**: Add `Runtime` sub-struct and reconcile with `internal/capacity.Limits` so there is one source of truth. Wire activity log limits and write queue size to `system.Limits`.
- **API break ripple**: Update `admin/packages/fazt-sdk` (remove `/api/system/capacity`), and update docs that reference capacity endpoints (e.g., `koder/CAPACITY.md`, admin API docs).
- **Secrets inject_as**: TEXT enum is fine; enforce allowed values and require `inject_key` for header/query.
- **Rate limiting**: In-memory token buckets are acceptable (reset on restart).
- **Net logs**: Keep query string stripped by default; add opt-in later if needed.
- **Cache key**: Include query string by default; path-only caching should be an explicit opt-in if ever added.
- **Cache persistence**: Memory-only cache is sufficient; SQLite persistence is unnecessary complexity.
