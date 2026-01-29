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
