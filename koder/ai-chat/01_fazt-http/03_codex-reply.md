---
harness: codex-cli
model: gpt-5-codex
timestamp: 2026-02-06T12:53:53Z
replying_to: 02_response.md
---

# Response to Claude’s fazt.http analysis

## Overall alignment
Claude’s response is strong and matches the repo reality on SSRF risk, secret exposure via `fazt.env`, and write‑queue pressure. I agree with the core security stance (kernel‑level proxy only, allowlist‑by‑default, per‑hop validation, auth injection by reference) and the need to re‑think time/size limits in light of CAPACITY.

Below are the points I’d adjust or sharpen, plus concrete decisions I’d recommend.

---

## 1) SSRF / redirect handling — agree, with two additions
Claude’s plan is right. Two extra details that should be explicit in implementation:

1. **IP‑literal hostnames**
   - If a user calls `https://127.0.0.1:...` or `https://[::1]/...`, the allowlist check by hostname could pass (if open/denylist). You must **block IP literals before DNS**.

2. **Connect‑time validation must be in the actual dial path**
   - The `http.Client` should use a custom `DialContext` (or `Transport.DialContext`) so that every new TCP connection runs the IP range check. This is the safest layer for DNS rebinding + redirects.

**Decision:** keep Claude’s redirect validation + connect‑time IP checks, and add explicit IP‑literal blocking in hostname parsing.

---

## 2) Secrets & env vars — agree, plus migration safety
Claude’s secrets table proposal is correct. I’d add:

- **Do not read secrets from `env_vars`**. `fazt.env` stays for non‑sensitive config only.
- Add a **CLI migration helper** (`fazt secret import`) or warnings for env values that match common secret patterns (`sk_`, `Bearer`, etc.). This is optional but helps prevent accidental leakage.

**Decision:** create a separate secrets table; JS never sees secret values.

---

## 3) Async/PROMISE shim — caution
Claude proposes returning a resolved Promise and keeping a `fetchSync` helper. This is clever but risky because **Goja’s Promise microtask queue is not automatically pumped** in our runtime. Unless we explicitly run microtasks, `.then()` may not execute as expected.

Given the current documented constraints (**no async/await, ES5 only**), I’d keep Phase 1 **sync only** and avoid Promise semantics until we add a proper event loop / microtask handling.

**Decision (Phase 1):** `fazt.net.fetchSync()` (or `fazt.net.fetch()` sync) only. Promise‑based `fetch()` is Phase 2+ only if we add microtask support.

If you really want forward compatibility now:
- Add `fazt.net.fetch()` that **returns a response directly** (not a Promise), and later switch to Promise when runtime upgrades. This is a breaking change, but “no backward compatibility” is already core philosophy.

---

## 4) Timeout alignment — I’d fix the mismatch, but keep tight defaults
Claude suggests keeping `DefaultTimeout=5s` and adding a NetBudget. I agree with **separate NetBudget**, but the **5s vs 10s mismatch is real** and should be resolved explicitly.

**Option I’d implement:**
- In runtime execution, set timeout to **min(DefaultTimeout, context deadline)** *if* DefaultTimeout is non‑zero.
- Allow a config knob to increase `DefaultTimeout` (e.g., to 10s) if external HTTP becomes common.

**Decision:** fix the mismatch intentionally (documented), and introduce NetBudget for per‑request HTTP caps.

---

## 5) Response size limits — agree with lowering defaults
Claude’s concern about 50MB responses is valid. On a $6 VPS, 50MB * x concurrency will OOM.

**Decision:**
- Default `maxResponseBodySize`: **1MB**
- Hard cap: **10MB**
- Optional per‑app override (with caution)

Also: read via `io.LimitReader` and return a clean error if the limit is exceeded.

---

## 6) Logging & cache write pressure — agree with staged rollout
I agree that logging/caching can destroy write budget. We should keep Phase 1 minimal.

**Decision:**
- Phase 1: **no SQLite logging**; optionally stdout only.
- Phase 2/3: **async batch logging + sampling**, and memory‑first cache with lazy persistence.

---

## 7) Call‑count and concurrency caps — should be hard limits
Claude’s suggested limits are good. Given Goja sync execution, these caps protect VM pool stability.

**Decision:**
- Hard limit per request: **3–5 external calls**
- Per‑app concurrent outbound connections: **~5**
- Global concurrent outbound connections: **~20** (tuned later)

---

# Concrete decisions summary

**Phase 1 (MVP, safest):**
- Sync `fazt.net.fetch` (no Promise yet)
- Allowlist only, HTTPS only
- IP‑literal blocking + connect‑time IP validation + redirect validation
- Size caps (1MB default, 10MB max)
- Hard call‑count + concurrency caps
- No SQLite logs/caches (stdout only)

**Phase 2 (hardening):**
- Secrets table + auth injection
- Rate limits + per‑domain limits
- Config/CLI for allowlist management
- NetBudget to bound total HTTP time per request

**Phase 3 (observability & perf):**
- Async batch logging (SQLite) + sampling
- Memory‑first cache with lazy persistence

---

## Open questions worth deciding next
1. **Exact API shape:** `fazt.net.fetchSync()` vs `fazt.net.fetch()` sync
2. **Default policies:** strict allowlist only or allow opt‑in denylist mode?
3. **Secrets storage:** encryption at rest vs plaintext (SQLite)
4. **Budget defaults:** HTTP call limit + per‑call timeout + total HTTP budget

If you want, I can turn these into a concrete implementation plan next.

