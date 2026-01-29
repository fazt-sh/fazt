# Fazt Implementation State

**Last Updated**: 2026-01-29
**Current Version**: v0.11.9

## Status

State: **APP BUILDING** - Testing capabilities through real apps

---

## Current Focus

Building apps to stress-test fazt's capabilities and discover edge cases.
This validates the security hardening (Plan 27) under real workloads.

**Target apps** (small startup IT suite):
- Meet (scheduling)
- Docs (collaborative editing)
- Chat (real-time messaging)
- Notes (Notion-like)
- Sign (document signing)
- Files (file sharing)

Each app exercises different fazt capabilities and helps find gaps.

---

## Tomorrow's Directions

### 1. Google Sign-in Redirect
- Currently always redirects to root after OAuth
- Either fix the redirect to original page, or make the landing page pretty

### 2. Nexus App (Stress Test All Capabilities)
- Find/build an app idea that uses ALL fazt capabilities
- Consider making Nexus do this
- Goal: comprehensive stress test of the platform

### 3. `fazt @peer` Pattern Audit
- Review all commands for `@peer` support
- `fazt app list` seems incomplete - should show better results
- Ensure CLI ↔ API 1:1 parity (all CLI commands have API equivalents)

### 4. Analytics Deep Dive
- Audit: are all analytics properly collected & stored?
- Can analytics track users for comprehensive data flow view?
- Evaluate: is every state change captured? Should it be? (perf tradeoffs)
- Consider config options to disable some analytics for efficiency
- Need visualization/dashboard

### 5. Role-Based Access Control (RBAC)
- Can owner also Google sign-in and system recognizes them by email?
- Granular permissions per app/resource
- Current: owner vs user (OAuth) - needs refinement

### 6. App Audit
- Verify all docs are synced
- List all deployed apps
- Enhance apps with Google sign-in
- Games should have high score tracking

### 7. Documentation Overhaul
- Build comprehensive markdown-based, multi-file fazt documentation
- Organize as a Claude skill (usable directly in Claude Code)
- Structure should:
  - Sync with API/CLI changes
  - Generate documentation site
  - Drive development vision
- README.md is outdated - needs refresh

### 8. License Discussion
- Reconsider MIT license
- Proposed model: **Fair Code License**
  - MIT/free for everyone doing <$1M revenue
  - $1000 per $1M revenue above threshold
- Questions to resolve:
  - Is $1000/$1M too high? Too low?
  - How to enforce/verify?
  - Precedents (Elastic, MongoDB, etc.)?
- Cloud provider consideration:
  - DigitalOcean, MS, Google may want one-click installer
  - Don't scare them off - perhaps volume/partnership deals?
- **Evaluate making repo private** until license is figured out
  - Will any dependencies break?
  - Impact on current users?

### 9. Strategic Positioning
- **Capability comparison**: Re-evaluate vs Supabase & Vercel
  - What do they have that we don't?
  - What do we have that they don't?
  - Where's the gap?

- **Concept: "Break hyperscaler stack to independent nodes"**
  - Can fazt be the unit of compute that replaces cloud lock-in?
  - Each fazt = sovereign, portable, interconnectable

- **Vertical scaling evaluation**
  - Current: $6 VPS handles 2,300 req/s
  - What about $50 VPS? $500 VPS?
  - At what point does horizontal > vertical?

- **External integration value matrix**
  - Rank by value-add: S3, Litestream, Turso, Cloudflare, others
  - Which integration unlocks most capability?
  - Which has best effort/reward ratio?

---

## Pending Plans

### Plan 24: Mock OAuth Provider

**Status**: NOT IMPLEMENTED
**Purpose**: Enable full auth flow testing locally without code changes

Key features:
- Dev login form at `/auth/dev/login` (local only)
- Creates real sessions (same as production OAuth)
- Role selection for testing admin/owner flows
- Zero code changes when deploying to production

**Why needed**: Currently can't test auth flows locally without HTTPS.

### Plan 25: SQL Command

**Status**: NOT IMPLEMENTED
**Purpose**: Debug and inspect databases without SSH

```bash
fazt sql "SELECT * FROM apps"              # Local
fazt @zyt sql "SELECT * FROM auth_users"   # Remote
```

Key features:
- Read-only by default, `--write` flag for mutations
- Output formats: table, json, csv
- Works locally and remotely via API

**Why needed**: Currently requires SSH + sqlite3 for remote debugging.

---

## Completed Plans

### Plan 27: Security Hardening (v0.11.9)

All critical security issues from harness testing are fixed:

| Issue | Fix | Status |
|-------|-----|--------|
| Slowloris vulnerability | TCP-level ConnLimiter + TCP_DEFER_ACCEPT | ✅ |
| No rate limiting | Per-IP token bucket (500 req/s) | ✅ |
| Connection exhaustion | 50 conns/IP limit at TCP Accept | ✅ |
| No header timeout | 5s ReadHeaderTimeout | ✅ |
| HTTPS protection | CertMagic + ManageAsync() integration | ✅ |

Protection stack:
```
TCP_DEFER_ACCEPT → ConnLimiter → TLS → ReadHeaderTimeout → Rate Limit
```

### Plan 26: Harness Refactor

Converted test harness from CLI command to proper Go integration tests:
- `internal/harness/*_test.go` with `//go:build integration`
- Removed `cmd/server/harness.go` from binary
- Standard `go test` tooling

---

## Security Architecture

### Protection Stack (HTTP + HTTPS)

```
┌─────────────────────────────────────────────────────────────┐
│ Layer 1: TCP_DEFER_ACCEPT (Linux kernel)                    │
│          Connections that never send data → kernel drops    │
├─────────────────────────────────────────────────────────────┤
│ Layer 2: ConnLimiter (TCP Accept level)                     │
│          >50 conns from same IP → rejected before goroutine │
├─────────────────────────────────────────────────────────────┤
│ Layer 3: TLS (HTTPS mode only, via CertMagic)               │
│          TLS wraps protected listener, not raw TCP          │
├─────────────────────────────────────────────────────────────┤
│ Layer 4: ReadHeaderTimeout (net/http)                       │
│          Slow header senders → killed in 5 seconds          │
├─────────────────────────────────────────────────────────────┤
│ Layer 5: Rate Limiting Middleware (Handler)                 │
│          >500 req/s from same IP → 429 response             │
└─────────────────────────────────────────────────────────────┘
```

### Implementation Files

| File | Purpose |
|------|---------|
| `internal/listener/connlimit.go` | TCP-level per-IP connection limiter |
| `internal/listener/tcp.go` | TCP_DEFER_ACCEPT wrapper (Linux) |
| `internal/middleware/ratelimit.go` | Request-level rate limiting |
| `cmd/server/main.go:2955-3070` | Server startup with protection stack |

---

## Capacity Verified (v0.11.9)

| Metric | Value | Notes |
|--------|-------|-------|
| Local throughput | ~40,000 req/s | Read-only health checks |
| Mixed workload | ~2,300 req/s | 30% writes |
| 500 concurrent users | 90%+ success | 1 req/s each (rate-limited by single IP) |
| Production (zyt.app) | 100% success | Network-bound at 70 req/s |
| Slowloris protection | PASS | Connections killed after 5s |

**Startup capacity**: <1000 employees using fair-use IT apps = ~1% of capacity.

---

## Quick Reference

```bash
# Run integration tests
FAZT_TARGET="http://test-harness.192.168.64.3.nip.io:8080" \
FAZT_TOKEN="gvfg2rynqizdwilw" \
go test -v -tags=integration ./internal/harness/...

# Check zyt status
fazt remote status zyt

# Upgrade zyt
fazt remote upgrade zyt

# SSH to zyt (when remote upgrade fails)
ssh root@165.227.11.46

# Local server logs
journalctl --user -u fazt-local -f
```
