# Fazt Implementation State

**Last Updated**: 2026-02-06
**Current Version**: v0.25.4

## Status

State: PLANNING
Plan 40 (fazt.http) fully specified — all 4 phases. Ready for implementation.

---

## Last Session (2026-02-06) — Human Review & Plan Expansion

### What Was Done

#### 1. Human Walkthrough of Plan 40
Explained security model (SSRF, DialContext, redirect attacks, IP literal
blocking) in plain terms. Human now understands the "why" behind each decision.

#### 2. system.Limits Refactor (Step 0 added to plan)
Human identified that plan's limits weren't integrated with fazt's existing
`system.Limits` / `system.GetLimits()`. Major addition:
- Flat `system.Limits` (10 fields) → nested structs: `Hardware`, `Storage`,
  `Runtime`, `Capacity`, `Net`
- Struct tags for metadata: `label`, `desc`, `unit`, `range`, `readonly`
- Schema endpoint: `GET /api/system/limits/schema` (reflect-based, `sync.Once` cached)
- Deletes `internal/capacity/` — absorbed into `system.Limits`
- Future-enables: validation, config system, per-app overrides, admin UI forms

#### 3. Phase 2 & 3 Expanded (Sketch → Full Spec)
Human wanted thin layers with sane defaults, not deferred sketches:
- **Phase 2**: Secrets store (Go-side injection, JS never sees values), rate
  limiting (token bucket per domain, default disabled), per-domain config
  (extends allowlist table, zero = inherit)
- **Phase 3**: Async batch logging (buffer + flush, errors bypass, query
  strings stripped), response cache (memory-only LRU, opt-in per domain)
- All Phase 2/3 limits declared in `system.Limits.Net` from Step 0

#### 4. AI-Chat Thread Extended (messages 12-13)
- Message 12 (Claude): Summarized all additions, posed 9 review questions
- Message 13 (Codex): Reviewed all 9, flagged cache key fix (applied) and
  limits unification with `internal/capacity` (folded into plan)

**No code was written this session.** Plan expansion and review only.

---

## Next Session — IMPLEMENT

Plan 40 is complete. All phases specified. Codex-reviewed. Human understands
the security model. Ready to write code.

### Implementation order:
1. **Step 0**: `system.Limits` refactor (standalone, no egress code)
   - Nested structs + tags in `internal/system/`
   - Schema endpoint
   - Delete `internal/capacity/`, rewire `activity/logger.go`
   - Update fazt-sdk, knowledge-base API docs
2. **Step 1-8**: Phase 1 egress (proxy, allowlist, budget, response, wiring, CLI, tests, docs)
3. **Phase 2**: Secrets, rate limits, per-domain config
4. **Phase 3**: Logging, cache

### Key resources:
- `koder/plans/40_fazt_http.md` — the full plan (all phases)
- `koder/ai-chat/01_fazt-http/` — 13-message design thread
- `internal/system/probe.go` — current flat Limits (what Step 0 refactors)
- `internal/capacity/capacity.go` — to be deleted (absorbed into system.Limits)
- `internal/runtime/runtime.go` — Goja runtime (what Phase 1 extends)

---

## Quick Reference

```bash
# Plan
cat koder/plans/40_fazt_http.md

# AI-Chat thread
ls koder/ai-chat/01_fazt-http/*.md

# Key code files
cat internal/system/probe.go        # Current limits (Step 0 target)
cat internal/capacity/capacity.go   # To be deleted
cat internal/runtime/runtime.go     # Goja runtime
cat internal/timeout/budget.go      # Budget system
cat internal/runtime/handler.go     # Request flow
```
