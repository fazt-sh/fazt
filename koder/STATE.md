# Fazt Implementation State

**Last Updated**: 2026-01-29
**Current Version**: v0.11.5

## Status

State: **ACTION REQUIRED** - security vulnerabilities block production readiness

---

## Vision Reminder

Fazt aims to be **Supabase/Vercel-level capability** in a single binary + SQLite:
- Target: $6 VPS handling real production traffic
- Standard: "Install & it just works" - no config nightmare
- Bar: If it's not production-ready, it's not ready

These findings are **BLOCKERS**, not nice-to-haves.

---

## Critical Findings (Plan 27)

| Severity | Issue | Status |
|----------|-------|--------|
| **CRITICAL** | Slowloris vulnerability - 20 slow connections block all traffic | ❌ Unmitigated |
| **HIGH** | No rate limiting detected | ❌ Unmitigated |
| **MEDIUM** | Slow recovery after load (3-5s) | ❌ Unmitigated |
| **LOW** | No header read timeout | ❌ Unmitigated |

**See**: `koder/plans/27_harness_findings.md` for full analysis and remediation plan.

### Immediate Fix Required

```go
// Find where http.Server is created and add:
server := &http.Server{
    ReadHeaderTimeout: 5 * time.Second,
    ReadTimeout:       10 * time.Second,
    WriteTimeout:      30 * time.Second,
    IdleTimeout:       60 * time.Second,
}
```

---

## Next Up

### Plan 27: Harness Findings Remediation (PRIORITY)

Fix security and resilience gaps identified by test harness.

**Phase 1 (Immediate - blocks release):**
1. Add HTTP server timeouts (fixes slowloris + slow headers)
2. Implement rate limiting (per-IP token bucket)
3. Add connection limits per IP

**Phase 2 (This week):**
4. Runtime pooling for Goja
5. Request queue with backpressure
6. Graceful degradation

See: `koder/plans/27_harness_findings.md`

### Plan 24: Mock OAuth Provider (Deferred)

### Plan 25: SQL Command (Deferred)

---

## Last Session

**Plan 26: Harness Refactor + Plan 27: Findings Documentation**

### Completed

1. **Converted test harness to Go integration tests**
   - Created `testutil.go`, `baseline_test.go`, `requests_test.go`, `resilience_test.go`, `security_test.go`
   - Removed old embedded harness code (~4,000 lines removed)
   - Tests use `//go:build integration` tag

2. **Created test-harness app**
   - `servers/local/test-harness/api/main.js`
   - Endpoints: `/api/health`, `/api/hello`, `/api/echo`, `/api/slow`, `/api/timeout`
   - Uses nip.io wildcard DNS for subdomain routing

3. **Documented critical findings**
   - Created `koder/plans/27_harness_findings.md`
   - Slowloris vulnerability (CRITICAL)
   - No rate limiting (HIGH)
   - Slow recovery (MEDIUM)
   - Full remediation plan with 4 phases

### Test Results Summary

| Category | Pass | Skip | Notes |
|----------|------|------|-------|
| Baseline | 7/7 | 0 | Performance solid |
| Static | 5/5 | 0 | Security rules working |
| API | 3/3 | 1 | Admin endpoints skipped |
| Serverless | 3/3 | 0 | Runtime working |
| Resilience | 7/7 | 4 | KV tests skipped |
| Security | 6/6 | 3 | **Slowloris warning logged** |

### Performance Baselines

| Metric | Value |
|--------|-------|
| P50 latency | 400µs |
| P95 latency | 700µs |
| P99 latency | 1ms |
| RPS @100 concurrent | 3,000 |

---

## Quick Reference

```bash
# Run integration tests
FAZT_TARGET="http://test-harness.192.168.64.3.nip.io:8080" \
FAZT_TOKEN="gvfg2rynqizdwilw" \
FAZT_TEST_APP="test-harness.192.168.64.3.nip.io" \
go test -v -tags=integration ./internal/harness/...

# Security tests only
go test -v -tags=integration -run "TestSecurity" ./internal/harness/...

# Deploy test app
fazt app deploy servers/local/test-harness --to local

# Check remote status
fazt remote list
```
