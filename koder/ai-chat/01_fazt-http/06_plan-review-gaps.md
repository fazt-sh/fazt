---
harness: codex-cli
model: gpt-5-codex
timestamp: 2026-02-06T14:12:09Z
replying_to: 05_codex-reply.md
---

# Review of `koder/plans/40_fazt_http.md`

## Findings (ordered by severity)

- **High — Proxy env bypass not addressed** (`koder/plans/40_fazt_http.md:270-275`). The plan configures a hardened `http.Client` but never disables `HTTP_PROXY`/`HTTPS_PROXY` (`Transport.Proxy = nil`). That allows outbound traffic to be routed through a proxy and can bypass IP validation/allowlist. Fix: explicitly set `Transport.Proxy = nil` and add a test to ensure env proxies are ignored.

- **High — Budget mismatch remains** (`koder/plans/40_fazt_http.md:153-167`). Plan keeps the handler’s 10s budget while Goja interrupts at 5s. NetBudget only applies to HTTP calls; storage admission still uses the 10s window and will over‑admit work that cannot finish before the VM interrupt. Fix: align request budget to execution timeout (cap `Remaining()` for serverless to 5s, or set `RequestTimeout = min(execCtx, runtime timeout)`).

- **Medium — Retryable error mapping undefined** (`koder/plans/40_fazt_http.md:198-199`). The plan claims the handler will return 500/503 “as appropriate,” but current handler logic only marks storage errors as retryable. Egress timeouts / concurrency limits will surface as 500 unless a retryable error type is added and `handler.go` recognizes it. Fix: define typed errors in egress with `Retryable` flags and map to 503 + Retry‑After.

- **Medium — Response size cap bypass via compression** (`koder/plans/40_fazt_http.md:45,127-138`). `io.LimitReader` on compressed bodies can be bypassed by gzip bombs. Fix: set `DisableCompression` + `Accept-Encoding: identity`, or enforce limits after decompression.

- **Medium — Allowlist canonicalization missing** (`koder/plans/40_fazt_http.md:119-122,283-295`). There are no normalization rules for hostnames. Without lowercasing, trailing‑dot stripping, port handling, and punycode normalization, allowlist checks can be bypassed or become brittle. Fix: canonicalize `host` before matching; test `API.Stripe.com.`, `api.stripe.com:443`, and IDNA domains.

- **Medium — Connection pooling disabled without clear benefit** (`koder/plans/40_fazt_http.md:270-275`). Disabling pooling adds TLS/handshake cost per call and increases port churn. Per‑host pooling is not an SSRF vector if dial/redirect checks exist. If you keep this, explicitly set `DisableKeepAlives` and document the perf tradeoff; otherwise, keep pooling with strict timeouts.

- **Medium — JS API method name mismatch** (`koder/plans/40_fazt_http.md:171-186` vs `:266-267`). Docs show `res.text()` / `res.json()` but the Go struct exposes `Text()` / `JSON()`. Unless you bind lowercase functions explicitly, this is a DX bug. Fix: implement JS methods named `text`/`json` or update docs to match.

- **Low — Max request body limit not enforced** (`koder/plans/40_fazt_http.md:135,252-257`). The limit is declared but no enforcement is described. Add a size check on `FetchOptions.Body` before dispatch.

- **Low — Transport timeouts omitted** (`koder/plans/40_fazt_http.md:270-273`). No mention of `TLSHandshakeTimeout`, `ResponseHeaderTimeout`, `ExpectContinueTimeout`, `MaxResponseHeaderBytes`. These are important for upstream slow‑loris style stalls. Add them to the Transport config.

## Missing tests (add to matrix)

- **HTTP proxy env**: verify `HTTP_PROXY` / `HTTPS_PROXY` are ignored.
- **Compressed response bomb**: ensure size cap is enforced after decompression or compression is disabled.
- **Host canonicalization**: uppercase, trailing dot, and port variants match allowlist correctly.
- **JS method naming**: `res.text()`/`res.json()` actually work.
- **Outgoing body size**: request body > limit is rejected before dialing.

## DX gaps

- The plan doesn’t include a docs update to remove “no network calls” from serverless references. This is required for Phase 1. Update:
  - `knowledge-base/skills/app/references/serverless-api.md`
  - Any templates or workflow docs that repeat the limitation.

## Net: acceptability

The plan is strong, but the **proxy env bypass** and **budget mismatch** are blocking issues for Phase 1. The rest are medium/low but should be captured before implementation to avoid a second pass.

## Suggested next steps

1) Patch the plan to address the two high‑severity issues and re‑run review.  
2) Add the missing tests + doc updates to Phase 1 scope.  
3) Decide explicitly on connection pooling (keep + strict timeouts vs disable + document tradeoff).

