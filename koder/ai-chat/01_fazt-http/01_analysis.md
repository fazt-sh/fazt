---
harness: codex-cli
model: gpt-5-codex
timestamp: 2026-02-06T10:58:18Z
---

# fazt.http (external HTTP from serverless) — analysis

## Context snapshot (repo-grounded)
- Serverless runtime is Goja, **sync-only**, ES5, **no async/await**, and docs explicitly say **no network calls** today. (`internal/runtime/runtime.go`, `knowledge-base/skills/app/references/serverless-api.md`)
- Runtime default timeout is **5s**; handler budget is **10s** with per-op storage sub-budgets. (`internal/runtime/runtime.go`, `internal/runtime/handler.go`, `internal/timeout/budget.go`)
- All writes are serialized via a **single WriteQueue** (SQLite single-writer). (`internal/storage/writer.go`)
- Capacity targets are tuned for a $6 VPS: ~800 writes/s, ~2,300 mixed req/s. (`koder/CAPACITY.md`)
- A kernel-level **egress proxy** is already specced (allow/deny, caching, auth injection, logs). (`koder/ideas/specs/v0.8-kernel/proxy.md`)

## Security analysis
**Primary risks**
- **SSRF / internal network access**: Without strict egress control, serverless can hit localhost, RFC1918, metadata IPs, internal services.
- **DNS rebinding / redirect bypass**: Allowlist by hostname alone can be bypassed via resolution changes or redirects.
- **Secret leakage**: If app code sees secrets, it can exfiltrate via logs or outgoing requests.
- **Abuse / cost**: Unbounded external calls can exhaust upstream rate limits, costs, and local capacity.

**Non-negotiable safeguards (MVP)**
- **Kernel-level proxy only** (no direct outbound in runtime).
- **Allowlist by default**; denylist and open modes are opt-in.
- **IP range blocking** at connect time (loopback, link-local, private, CGNAT, ULA, etc.).
- **Redirect validation** at each hop (host + resolved IP check).
- **HTTPS by default**; HTTP requires explicit config.
- **Secrets injected by reference** (e.g., `auth: 'STRIPE_KEY'`) so secrets never enter JS.
- **Per-app and global rate limits** and bounded retries.
- **Size limits** for request/response bodies and headers.

**Logging hygiene**
- Log **metadata only** by default (method, domain, status, latency, cached). Any headers or bodies must be redacted or disabled.

## Performance & capacity analysis
**Blocking call pressure**
- Goja is synchronous. A blocking HTTP call ties up a serverless VM and the request goroutine until completion. With **MaxPoolSize = 100**, 100 slow calls can stall serverless concurrency. (`internal/runtime/runtime.go`)

**Timeout mismatch**
- Runtime interrupts after **5s**, but request budget allows **10s**. This will cut off external HTTP early unless aligned. (`internal/runtime/runtime.go`, `internal/runtime/handler.go`, `internal/timeout/budget.go`)

**Write queue pressure**
- Proxy logs + cache writes will consume the SQLite write budget. If every external call logs (or caches), a modest volume can push the write queue into admission failures and cause `503` retryable responses. (`internal/storage/writer.go`, `koder/CAPACITY.md`)

**Memory risk**
- If responses are read fully into memory, a few large responses can spike RAM and GC. Hard caps + streaming/partial reads are required.

## Implementation complexity (realistic scope)
This is not just a `fetch()` binding; it’s a **kernel egress subsystem**.

**Core components**
1. **Proxy core (Go)**
   - Strict `http.Client` timeouts (connect/TLS/header/body)
   - Dialer with IP range validation
   - Redirect handler with re-validation
   - Allowlist/denylist + wildcard matching
   - Rate limits + concurrency pool (per app + global)
   - Size caps + response buffering rules

2. **JS binding**
   - `fazt.net.fetch(url, options)` must be **sync** for Goja
   - Response object: `status`, `ok`, `headers`, `text()`, `json()`
   - Structured errors: code + retryable flag

3. **Secrets store**
   - DB-backed secrets (not exposed via `fazt.env`)
   - Auth injection by reference (`auth: 'OPENAI_KEY'`)

4. **Storage schema + retention**
   - `kernel_net_logs`, `kernel_net_cache` tables (already spec’d)
   - Retention cleanup / cache eviction with write budget awareness

5. **Config & CLI**
   - DB-backed config (no files) to match Fazt philosophy
   - Commands: allowlist, logs, cache clear, limits

6. **Docs & tests**
   - Update serverless docs (remove “no network calls”)
   - SSRF tests, redirect tests, size/timeouts, rate limit tests

## Limits to tweak (likely required)
**Timeout alignment**
- Align runtime timeout with request budget. Options:
  - Raise `DefaultTimeout` to 10s, or
  - Make runtime use the context deadline from the handler only.

**Net-specific limits**
- `maxRequestBodySize`: 1–10MB
- `maxResponseBodySize`: 1–5MB default, configurable
- `maxTimeout`: 30–60s default; higher only by config
- `maxRedirects`: 3–5
- Per-domain + per-app rate limits

**Write queue mitigation**
- Sampling for logs, or async batch writes
- Cache writes capped by size + TTL; avoid cache stampedes

## What external API access unlocks
- Payments (Stripe), email/SMS (SendGrid/Twilio)
- AI APIs (OpenAI/Anthropic)
- Third-party data enrichment and aggregation
- Webhook relay & workflow automation
- Centralized caching of upstream APIs

This is a major unlock: Fazt becomes a true “glue” platform rather than purely local storage + static hosting.

## Recommended implementation path
**Phase 1 — Safe MVP (closed by default)**
- `fazt.net.fetch` sync API
- Allowlist-only mode + IP range blocking
- Strict timeouts + size caps
- Minimal logging metadata only
- No cache initially

**Phase 2 — Hardening**
- Rate limiting + concurrency pool
- Auth injection with secrets store
- Redirect validation + HTTPS enforcement
- Structured errors and retry guidance

**Phase 3 — Cache + Logs**
- TTL cache with size limit + eviction
- Logs table with retention and sampling

**Phase 4 — Sandbox parity**
- `net:fetch` capability uses same proxy path (`koder/ideas/specs/v0.10-runtime/sandbox.md`)

## Key decision still open
**Sync vs async API**
- Goja is sync; current docs explicitly forbid async/await. The lowest-risk path is a **sync `fazt.net.fetch`** with immediate response + `.json()` helper. Promise support would be a large scope increase.

## Open questions
- Default policy: allowlist-only vs denylist vs open?
- Secret storage design: table schema, rotation, access control?
- How much logging is acceptable without harming write budget?
- Should cache be per-app or global with app namespace?

