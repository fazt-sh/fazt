# Fazt Implementation State

**Last Updated**: 2026-02-06
**Current Version**: v0.25.4

## Status

State: PLANNING
Plan 40 (fazt.http) complete. Needs human understanding session before implementation.

---

## Last Session (2026-02-06) — fazt.http Design & Review

### What Was Done

#### 1. AI-Chat Thread: fazt.http (11 messages, Claude + Codex)
Full architectural discussion on adding external HTTP calls to serverless runtime.
Thread at `koder/ai-chat/01_fazt-http/` — 11 messages covering security (SSRF,
redirect, secrets), performance (budget, pooling, write pressure), and DX (API
shape, error codes, docs).

#### 2. Plan 40: fazt.http — Written and Reviewed
`koder/plans/40_fazt_http.md` — Full Phase 1 spec with Phases 2-3 sketched.
Reviewed twice by Codex. All findings addressed. Key decisions:
- `fazt.net.fetch()` sync API, returns response object
- Kernel-level egress proxy with allowlist-only security
- SSRF protection via DialContext IP validation
- Budget system aligned with 5s runtime timeout
- 28-test matrix covering security, limits, and DX

**No code was written this session.** Plan only.

---

## Next Session — DISCUSSION FIRST

**IMPORTANT**: Human wants to understand the plan deeply before implementation.
The plan covers complex security and runtime topics (SSRF, DNS rebinding, Goja
Promise semantics, timeout budgets, connection pooling). Human needs a
walkthrough session to feel technically confident before approving implementation.

### Suggested approach:
1. Open session, read this state
2. **Discussion mode** — walk through Plan 40 section by section
3. Human asks questions, Claude explains the "why" behind each decision
4. Focus areas likely: SSRF attacks, how DialContext works, budget system,
   allowlist mechanics, error flow from Go to JS
5. Close session after understanding is solid
6. Next session: implement

### Key resources for discussion:
- `koder/plans/40_fazt_http.md` — the full plan
- `koder/ai-chat/01_fazt-http/` — 11-message design thread (deep context)
- `internal/runtime/runtime.go` — current Goja runtime (what we're extending)
- `internal/timeout/budget.go` — current budget system (what we're adding to)
- `internal/runtime/handler.go` — request flow (where egress proxy plugs in)

---

## Quick Reference

```bash
# Plan
cat koder/plans/40_fazt_http.md

# AI-Chat thread
ls koder/ai-chat/01_fazt-http/*.md

# Key code files for understanding
cat internal/runtime/runtime.go
cat internal/timeout/budget.go
cat internal/runtime/handler.go
```
