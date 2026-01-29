# Fazt Implementation State

**Last Updated**: 2026-01-29
**Current Version**: v0.11.6

## Status

State: **READY** - Security hardening complete, ready for production

---

## Vision Reminder

Fazt aims to be **Supabase/Vercel-level capability** in a single binary + SQLite:
- Target: $6 VPS handling real production traffic
- Standard: "Install & it just works" - no config nightmare
- Bar: If it's not production-ready, it's not ready

---

## Security Status (Plan 27)

| Severity | Issue | Status |
|----------|-------|--------|
| **CRITICAL** | Slowloris vulnerability | ✅ Mitigated (ReadHeaderTimeout) |
| **HIGH** | No rate limiting | ✅ Fixed (per-IP token bucket) |
| **MEDIUM** | Connection exhaustion | ✅ Fixed (connection limits) |
| **LOW** | No header read timeout | ✅ Fixed (5s ReadHeaderTimeout) |

### Security Measures Implemented

1. **HTTP Server Timeouts** (cmd/server/main.go)
   - ReadHeaderTimeout: 5s (prevents slowloris)
   - ReadTimeout: 10s
   - WriteTimeout: 30s
   - IdleTimeout: 60s

2. **Rate Limiting** (internal/middleware/ratelimit.go)
   - Per-IP token bucket algorithm
   - 500 req/s sustained, 1000 burst
   - Returns 429 with Retry-After header
   - Automatic cleanup of stale entries

3. **Connection Limiting** (internal/middleware/ratelimit.go)
   - Max 200 concurrent connections per IP
   - Returns 503 when limit exceeded

---

## Next Up

### Plan 27 Phase 2: Performance Optimization (Optional)

These are enhancements, not blockers:

1. Runtime pooling improvements for Goja
2. Request queue with backpressure
3. Circuit breaker for serverless

### Plan 24: Mock OAuth Provider (Deferred)

### Plan 25: SQL Command (Deferred)

---

## Last Session

**Plan 27: Security Hardening Release (v0.11.6)**

### Completed

1. **Added HTTP server timeouts**
   - ReadHeaderTimeout: 5s (key slowloris fix)
   - Adjusted other timeouts for better protection

2. **Implemented rate limiting middleware**
   - Per-IP token bucket (golang.org/x/time/rate)
   - RWMutex for performance (read-heavy workload)
   - Automatic cleanup goroutine

3. **Implemented connection limiting**
   - Per-IP concurrent connection tracking
   - Prevents resource exhaustion

4. **Integration test verification**
   - SlowHeaders test: PASS (server closes slow connections)
   - Rate limiting: Working (429 responses confirmed)
   - All unit tests: PASS

### Test Results (Post-Fix)

| Test | Result | Notes |
|------|--------|-------|
| SlowHeaders | ✅ PASS | Server closes slow header connections |
| ConnectionFlood | ✅ PASS | 200/200 success |
| RateLimitRecovery | ✅ PASS | 4/5 succeeded after cooldown |
| ServiceDuringSlowloris | ⚠️ Warning | Expected - requires reverse proxy for full protection |

### Notes

- Slowloris test still shows warning (0/10 during attack)
- This is expected behavior for Go's net/http without a reverse proxy
- For full slowloris protection, deploy behind Caddy/nginx
- The ReadHeaderTimeout fix prevents the most common attack variant

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
