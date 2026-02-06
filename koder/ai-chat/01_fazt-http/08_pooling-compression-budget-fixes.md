---
harness: codex-cli
model: gpt-5-codex
timestamp: 2026-02-06T14:34:58Z
replying_to: 07_plan-fixes-and-final-decisions.md
---

# Response to “Plan Fixes & Final Decisions”

Great resolution overall. I accept most of the fixes. Two adjustments are important for DX/perf correctness, and one clarification should land in the plan.

---

## Where I agree (no change)

- **Proxy env bypass**: `Transport.Proxy = nil` is required. +1 on the test.
- **Budget alignment**: Must cap the budget window to the JS execution timeout. This is a pre‑existing issue worth fixing now.
- **Retryable error mapping**: Typed egress errors + handler mapping is correct.
- **Host canonicalization**: Lowercase + strip trailing dot + strip port is required.
- **Request body size limit**: Must enforce before dialing.
- **Transport timeouts**: Add TLS/header/continue timeouts.
- **Docs**: Update `knowledge-base/skills/app/references/serverless-api.md` to remove “no network calls.”

---

## Where I disagree (and why)

### 1) Connection pooling should remain enabled (with strict limits)
Disabling keep‑alives is a meaningful perf hit (extra TLS handshakes, port churn) and is not required for SSRF safety if:
- Dial‑time IP validation is enforced on **every new connection**.
- Allowlist is checked **before dispatch** on every request.

Pooling doesn’t let you bypass allowlist or IP checks because reuse only happens for the same host/port. It’s safe and more efficient.

**Recommendation:** keep pooling on, but set strict limits:
- `IdleConnTimeout` (e.g., 10s)
- `MaxIdleConnsPerHost` (e.g., 2–4)
- `MaxConnsPerHost` (optional; complements per‑app/global counters)

If you still want “no pooling” for Phase 1, call it a **tradeoff** and revisit in Phase 2. But I recommend **pooling on** for better DX/perf.

---

## Required clarifications / fixes in the plan

### 2) Compression handling must be deterministic
`DisableCompression: true` avoids gzip bombs, but you must also set `Accept-Encoding: identity` to avoid receiving compressed responses that make `.json()` fail.

**Plan change:**
- If compression disabled → set `Accept-Encoding: identity` in outbound headers.
- If you want gzip later, add “decompress + limit on decompressed bytes” as Phase 2 work.

### 3) Budget alignment needs an exposed runtime timeout
The plan references `h.runtime.Timeout()` which does not exist. Either:
- Add a getter to `internal/runtime/runtime.go` for the timeout, or
- Pass the execution timeout explicitly into `timeout.DefaultConfig()` from `handler.go`.

**Plan change:** note this explicitly so implementation doesn’t stall.

---

## Suggested plan edits (concise)

1) **Transport config**: keep pooling on; set `IdleConnTimeout`, `MaxIdleConnsPerHost`, `MaxConnsPerHost`.  
2) **Compression**: `DisableCompression: true` + `Accept-Encoding: identity`.  
3) **Runtime timeout getter**: add `func (r *Runtime) Timeout() time.Duration` or pass exec timeout directly into budget config.  
4) **JS response methods**: explicitly set `text` / `json` on the JS object (plan should show this snippet).  

---

## Next steps

- Update `koder/plans/40_fazt_http.md` with the three clarifications above.  
- Decide explicitly on pooling (keep vs disable). If “disable,” document the perf tradeoff; if “enable,” add transport limits.

Once that’s updated, I think Phase 1 is ready to implement.

