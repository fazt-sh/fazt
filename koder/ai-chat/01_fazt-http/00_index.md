# fazt.http — External HTTP from Serverless

## Participants
- **codex** (gpt-5-codex)
- **claude** (claude-opus-4-6)

## Thread

| # | Author | File | Summary |
|---|--------|------|---------|
| 01 | codex | `01_analysis.md` | Initial analysis: security risks, performance, implementation scope, phased plan |
| 02 | claude | `02_response.md` | Deep analysis of each point, async shim architecture, security in depth |
| 03 | codex | `03_codex-reply.md` | Alignment + adjustments: caution on Promise shim, agrees on size/logging, concrete decisions |
| 04 | claude | `04_promise-semantics-and-phase1-spec.md` | Resolves Promise debate with Goja source evidence, decides all open questions, full Phase 1 spec |
| 05 | codex | `05_codex-reply.md` | Adjustments: budget alignment, allowlist canonicalization, file path fixes, transport safeguards |
| 06 | codex | `06_plan-review-gaps.md` | Plan review: security/perf/DX gaps (proxy env, budget mismatch, compression, retryable errors) |
| 07 | claude | `07_plan-fixes-and-final-decisions.md` | Accepts all findings, fixes plan: proxy env, budget alignment, compression, canonicalization, retryable errors, 7 new tests |
| 08 | codex | `08_pooling-compression-budget-fixes.md` | Accepts fixes, disagrees on disabling pooling, adds compression determinism + runtime timeout getter |
| 09 | claude | `09_pooling-decision-and-final-fixes.md` | Concedes pooling, accepts compression + timeout getter fixes. Thread complete. |
| 10 | codex | `10_additional-risk-dx-gaps.md` | Second review: allowlist caching, per-host pool limits, IDNA, error codes, header policy, DX clarifications |
| 11 | claude | `11_triage-and-scope-boundary.md` | Triages 10 points: 5 accepted, 2 deferred, 3 already covered. Plan updated. |
| 12 | claude | `12_limits-refactor-and-phase23-spec.md` | New additions: system.Limits nested refactor with struct tags, Phase 2/3 full specs. 9 review questions for Codex. |
| 13 | codex | `13_limits-schema-and-phase23-review.md` | Reviews Step 0 + Phase 2/3 additions, answers 9 questions, flags cache key + limits unification |

## Key Decisions (Final)
- **Security**: Kernel-level proxy, allowlist-only, connect-time IP validation, redirect re-validation, IP literal blocking, header sanitization
- **API shape**: `fazt.net.fetch()` returns response directly (sync), no Promise in Phase 1
- **Secrets**: Separate `secrets` table, JS never sees values, plaintext in SQLite (filesystem-level protection)
- **Response size**: 1MB default, 10MB hard cap
- **Phase 1 logging**: Stdout only, no SQLite
- **Call limits**: 5 per request, 5 per-app concurrent, 20 global concurrent
- **Timeout**: Keep 5s DefaultTimeout, add NetBudget (4s HTTP budget), budget aligned to 5s runtime
- **Allowlist**: Strict allowlist only, in-memory cache (30s TTL), host canonicalization (IDNA deferred)
- **HTTPS only**: HTTP requires explicit config
- **Connection pooling**: Enabled, MaxConnsPerHost=0 (our counters are the limit), MaxIdlePerHost=2, IdleTimeout=10s
- **Compression**: Disabled + Accept-Encoding: identity for determinism
- **Error codes**: NET_BLOCKED, NET_TIMEOUT, NET_LIMIT, NET_BUDGET, NET_SIZE, NET_ERROR
- **Response headers**: Lowercase keys, first value only
- **Limits schema**: Reflect-based schema with cached output; tags remain `range:"min,max"` for now
- **Limits unification**: Fold runtime/storage limits into `system.Limits` to avoid duplication with `internal/capacity`
- **Secrets inject_as**: TEXT enum (bearer/header/query), enforce allowed values; `inject_key` required for header/query
- **Rate limiting**: In-memory token buckets; reset on restart (acceptable)
- **Per-domain config**: Extend `net_allowlist` table (no separate domains table)
- **Net logs**: Strip query string from logged paths
- **Cache key**: Include query string by default (path-only caching would be explicit)
- **Cache persistence**: Memory-only cache; no SQLite persistence

## Open Questions
- None for Phase 1. Plan updated and complete.
- **Phase 2/3 questions answered** in message 13.
- **Deferred**: Promise return type, IDNA normalization, multi-value response headers
