# Fazt Implementation State

**Last Updated**: 2026-01-29
**Current Version**: v0.11.7

## Status

State: **READY** - Defense-in-depth slowloris protection complete

---

## Security Architecture (Plan 27 Complete)

### Protection Stack

```
┌─────────────────────────────────────────────────────────────┐
│ Layer 1: TCP_DEFER_ACCEPT (Linux kernel)                    │
│          Connections that never send data → kernel drops    │
├─────────────────────────────────────────────────────────────┤
│ Layer 2: ConnLimiter (TCP Accept level)                     │
│          >50 conns from same IP → rejected before goroutine │
├─────────────────────────────────────────────────────────────┤
│ Layer 3: ReadHeaderTimeout (net/http)                       │
│          Slow header senders → killed in 5 seconds          │
├─────────────────────────────────────────────────────────────┤
│ Layer 4: Rate Limiting Middleware (Handler)                 │
│          >500 req/s from same IP → 429 response             │
└─────────────────────────────────────────────────────────────┘
```

### Security Status

| Severity | Issue | Status |
|----------|-------|--------|
| **CRITICAL** | Slowloris vulnerability | ✅ Defense-in-depth |
| **HIGH** | No rate limiting | ✅ Fixed (per-IP token bucket) |
| **MEDIUM** | Connection exhaustion | ✅ Fixed (TCP-level limits) |
| **LOW** | No header read timeout | ✅ Fixed (5s ReadHeaderTimeout) |

### Implementation Files

| File | Purpose |
|------|---------|
| `internal/listener/connlimit.go` | TCP-level per-IP connection limiter |
| `internal/listener/tcp.go` | TCP_DEFER_ACCEPT wrapper (Linux) |
| `internal/middleware/ratelimit.go` | Request-level rate limiting |
| `cmd/server/main.go:2955-3015` | Server startup with protection stack |

---

## Last Session

**Plan 27: TCP-Level Slowloris Protection (v0.11.7)**

### Completed

1. **Deep research on Go slowloris protection**
   - Confirmed: niche use case, most use reverse proxies
   - Found: `hashicorp/go-connlimit`, `valyala/tcplisten`
   - Decision: Custom implementation (~70 lines) + tcplisten

2. **Implemented TCP-level connection limiter**
   - `internal/listener/connlimit.go` - custom `net.Listener` wrapper
   - Per-IP tracking with `map[string]int` + `sync.Mutex`
   - Atomic counters for total connections
   - Connection rejected at Accept() before goroutine spawns

3. **Added TCP_DEFER_ACCEPT for Linux**
   - `internal/listener/tcp.go` using `valyala/tcplisten`
   - Kernel filters connections that connect but never send
   - Graceful fallback to `net.Listen` on non-Linux

4. **Integrated into server startup**
   - Both HTTP and HTTPS modes use protected listener
   - HTTPS: TLS wraps protected listener (CertMagic for certs)
   - Removed redundant middleware-level connection limiter

### Test Results

- TCP-level rejection working (logs show `per_ip_limit` rejections)
- SlowHeaders test: PASS (server closes slow connections)
- SlowBody test: PASS in 0.10s (was 7s before TCP_DEFER_ACCEPT)
- All unit tests: PASS

---

## Research Artifacts

Created research query framework:
```
koder/researches/
├── queries/
│   ├── 01_go-slowloris-protection.md   # Pure technical query
│   └── 02_fazt-slowloris-integration.md # Implementation-focused
└── reports/
    ├── 01_go-slowloris-protection/     # Research results
    └── 02_fazt-slowloris-integration/
```

Key findings documented in research reports.

---

## Next Up

### Plan 27 Phase 2: Performance Optimization (Optional)

- Runtime pooling improvements for Goja
- Request queue with backpressure
- Circuit breaker for serverless

### Plan 24: Mock OAuth Provider (Deferred)

### Plan 25: SQL Command (Deferred)

---

## Quick Reference

```bash
# Run integration tests
FAZT_TARGET="http://test-harness.192.168.64.3.nip.io:8080" \
FAZT_TOKEN="gvfg2rynqizdwilw" \
FAZT_TEST_APP="test-harness.192.168.64.3.nip.io" \
go test -v -tags=integration ./internal/harness/...

# Check connection limiter logs
journalctl --user -u fazt-local | grep "reject"

# Test per-IP limit (open 60 conns, limit is 50)
for i in {1..60}; do (nc -q 10 192.168.64.3 8080 &); done
```
