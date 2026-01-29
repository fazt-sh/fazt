# Plan 27: Harness Test Findings & Remediation

**Created**: 2026-01-29
**Status**: NEEDS ACTION
**Priority**: HIGH (security vulnerabilities identified)

---

## Executive Summary

Integration test harness (Plan 26) revealed security and resilience gaps that
need addressing before fazt can be considered production-ready.

| Severity | Finding | Status |
|----------|---------|--------|
| **CRITICAL** | Slowloris vulnerability | Unmitigated |
| **HIGH** | No rate limiting detected | Unmitigated |
| **MEDIUM** | Slow recovery after load | Unmitigated |
| **LOW** | Slow header timeout | Unmitigated |

---

## 1. Slowloris Vulnerability (CRITICAL)

### Finding

```
TestSecurity_ServiceDuringSlowloris:
  Service during slowloris: 0/10 legitimate requests succeeded
  Warning: server appears vulnerable to slowloris attack
```

20 slow connections completely block all legitimate traffic. This is a
denial-of-service vulnerability.

### Root Cause

Go's `net/http` server has no built-in protection against slow clients.
The server holds connections open indefinitely waiting for complete requests.

### Attack Vector

```
1. Attacker opens many connections
2. Sends partial HTTP request (headers incomplete)
3. Trickles data slowly to keep connection alive
4. Server exhausts connection pool
5. Legitimate users cannot connect
```

### Possible Fixes

**Option A: ReadHeaderTimeout + ReadTimeout (Recommended)**
```go
server := &http.Server{
    ReadHeaderTimeout: 5 * time.Second,  // Time to read headers
    ReadTimeout:       10 * time.Second, // Time to read entire request
    WriteTimeout:      30 * time.Second, // Time to write response
    IdleTimeout:       60 * time.Second, // Keep-alive timeout
}
```
- Pros: Simple, built into stdlib
- Cons: May affect legitimate slow clients

**Option B: Connection limits per IP**
```go
// Track connections per IP, reject if > threshold
type connLimiter struct {
    mu    sync.Mutex
    conns map[string]int
    max   int
}
```
- Pros: Targeted protection
- Cons: More complex, needs cleanup goroutine

**Option C: Reverse proxy (nginx/caddy)**
```nginx
# nginx.conf
client_header_timeout 5s;
client_body_timeout 10s;
limit_conn_zone $binary_remote_addr zone=conn:10m;
limit_conn conn 10;
```
- Pros: Battle-tested, offloads work
- Cons: Additional dependency

### Recommendation

Implement Option A first (server timeouts), then consider Option B for
defense in depth. Option C is for production deployments behind a proxy.

### Test to Verify Fix

```bash
go test -v -tags=integration -run TestSecurity_ServiceDuringSlowloris ./internal/harness/...
# Should show: Service during slowloris: 8/10 legitimate requests succeeded
```

---

## 2. No Rate Limiting Detected (HIGH)

### Finding

```
TestSecurity_BasicRateLimit:
  Rate limit test: ok=50, rate_limited=0
  Warning: no rate limiting detected (this may be expected)
```

50 rapid requests all succeeded. No 429 responses.

### Root Cause

Rate limiting was designed but may not be enabled or configured.

### Attack Vector

```
1. Attacker sends thousands of requests
2. Exhausts server resources
3. Legitimate users experience degraded service
```

### Possible Fixes

**Option A: Token bucket per IP**
```go
type rateLimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex
    rate     rate.Limit  // requests per second
    burst    int         // burst size
}

func (rl *rateLimiter) Allow(ip string) bool {
    rl.mu.Lock()
    limiter, exists := rl.limiters[ip]
    if !exists {
        limiter = rate.NewLimiter(rl.rate, rl.burst)
        rl.limiters[ip] = limiter
    }
    rl.mu.Unlock()
    return limiter.Allow()
}
```

**Option B: Sliding window counter**
- Track request counts per IP per time window
- More memory efficient for many IPs

**Option C: Middleware integration**
- Use existing middleware if available
- Check if rate limiting exists but isn't configured

### Research Needed

1. Check if `internal/middleware/` has rate limiting code
2. Check server configuration for rate limit settings
3. Determine appropriate limits (100 req/s? 1000 req/s?)

### Test to Verify Fix

```bash
go test -v -tags=integration -run TestSecurity_BasicRateLimit ./internal/harness/...
# Should show: Rate limit test: ok=30, rate_limited=20
```

---

## 3. Slow Recovery After Heavy Load (MEDIUM)

### Finding

```
Multiple tests needed 3-5 retries after stress tests
Server returns 500 TimeoutError after heavy load
Takes 3-5 seconds to recover
```

### Root Cause

Serverless runtime (Goja) may have:
- Memory pressure from many concurrent executions
- Goroutine exhaustion
- No request queuing/backpressure

### Impact

- Cascading failures during traffic spikes
- Poor user experience during recovery

### Possible Fixes

**Option A: Runtime pooling**
```go
type RuntimePool struct {
    pool    chan *goja.Runtime
    maxSize int
}

func (p *RuntimePool) Get() *goja.Runtime {
    select {
    case rt := <-p.pool:
        return rt
    default:
        return goja.New()
    }
}
```

**Option B: Request queuing with backpressure**
```go
type RequestQueue struct {
    queue chan *Request
    size  int
}

// Return 503 Service Unavailable when queue full
```

**Option C: Circuit breaker**
```go
// Trip circuit when error rate > threshold
// Fail fast instead of queuing
```

### Research Needed

1. Profile runtime under load (pprof)
2. Check current Goja pooling implementation
3. Measure memory usage during stress

### Test to Verify Fix

```bash
go test -v -tags=integration -run "TestResilience_SustainedLoad|TestResilience_GCPressure" ./internal/harness/...
# Should not need retries in subsequent tests
```

---

## 4. Slow Header Timeout (LOW)

### Finding

```
TestSecurity_SlowHeaders:
  Warning: server did not timeout slow header sending
```

Server waited 6+ seconds for slow headers instead of timing out.

### Root Cause

No `ReadHeaderTimeout` configured on HTTP server.

### Fix

Same as Slowloris fix - add server timeouts.

---

## 5. Skipped Tests

These tests were skipped because they require admin endpoints, not a deployed
app. They should be run separately against the admin interface.

### Tests Requiring Admin Target

| Test | Endpoint | Reason |
|------|----------|--------|
| TestAPIRead_AppsListUnauthorized | /api/apps | Admin API |
| TestAuth_ProtectedEndpointNoSession | /api/me | Admin API |
| TestAuth_APIKeyAuth | /api/apps | Admin API |
| TestAuth_LoginLogoutCycle | /api/login | Admin API |
| TestAuth_SessionCookieSecurity | /api/login | Admin API |

**How to run:**
```bash
# Against admin interface (not app subdomain)
FAZT_TARGET="http://192.168.64.3:8080" \
FAZT_TOKEN="<admin-token>" \
FAZT_USERNAME="dev" \
FAZT_PASSWORD="dev" \
go test -v -tags=integration -run "TestAuth|TestAPIRead_AppsList" ./internal/harness/...
```

### Tests Requiring KV Storage

| Test | Endpoint | Reason |
|------|----------|--------|
| TestAPIWrite_KVSet | /api/storage/kv/... | Needs storage enabled |
| TestAPIWrite_ConcurrentWrites | /api/storage/kv/... | Needs storage enabled |
| TestResilience_QueueNormalLoad | /api/storage/kv/... | Needs storage enabled |
| TestResilience_QueueHeavyLoad | /api/storage/kv/... | Needs storage enabled |
| TestResilience_QueueBurst | /api/storage/kv/... | Needs storage enabled |
| TestResilience_QueueRecovery | /api/storage/kv/... | Needs storage enabled |
| TestSecurity_SmallPayload | /api/storage/kv/... | Needs storage enabled |
| TestSecurity_MalformedJSON | /api/storage/kv/... | Needs storage enabled |

**How to run:**
```bash
# Requires app with storage access
FAZT_TARGET="http://app-with-storage.192.168.64.3.nip.io:8080" \
FAZT_TOKEN="<token>" \
go test -v -tags=integration -run "TestAPIWrite|TestResilience_Queue|TestSecurity_.*Payload" ./internal/harness/...
```

### Tests Requiring Deploy Endpoint

| Test | Endpoint | Reason |
|------|----------|--------|
| TestSecurity_LargePayloadRejected | /api/deploy | Admin API |

---

## 6. Performance Baselines (Reference)

Current performance on local dev server (192.168.64.3):

| Metric | Value | Threshold | Status |
|--------|-------|-----------|--------|
| P50 latency | 400µs | <5ms | ✅ |
| P95 latency | 700µs | <20ms | ✅ |
| P99 latency | 1ms | <100ms | ✅ |
| RPS @1 conn | 430 | >300 | ✅ |
| RPS @10 conn | 2,300 | >1,500 | ✅ |
| RPS @50 conn | 2,900 | >2,500 | ✅ |
| RPS @100 conn | 3,000 | >2,500 | ✅ |
| High concurrency (500) | 46% success | >40% | ✅ |
| Max latency under load | 2s | <3s | ✅ |

These are baselines for serverless execution. Static file serving would be
significantly faster (~10-40k RPS based on earlier tests).

---

## 7. Action Plan

### Phase 1: Critical Security (This Week)

1. **Add HTTP server timeouts**
   - ReadHeaderTimeout: 5s
   - ReadTimeout: 10s
   - WriteTimeout: 30s
   - IdleTimeout: 60s

2. **Verify rate limiting**
   - Check if exists
   - Enable/configure if disabled
   - Implement if missing

### Phase 2: Resilience (Next Week)

3. **Runtime pooling**
   - Profile current behavior
   - Implement pool if beneficial

4. **Request backpressure**
   - Add queue with max size
   - Return 503 when full

### Phase 3: Validation

5. **Re-run full test suite**
   - All security tests should pass
   - No warnings about vulnerabilities

6. **Run admin tests separately**
   - Validate auth flows
   - Validate KV storage

---

## 8. Test Commands Reference

```bash
# Full suite against test-harness app
FAZT_TARGET="http://test-harness.192.168.64.3.nip.io:8080" \
FAZT_TOKEN="gvfg2rynqizdwilw" \
FAZT_TEST_APP="test-harness.192.168.64.3.nip.io" \
go test -v -tags=integration ./internal/harness/...

# Security tests only
go test -v -tags=integration -run "TestSecurity" ./internal/harness/...

# Resilience tests only
go test -v -tags=integration -run "TestResilience" ./internal/harness/...

# Performance baselines only
go test -v -tags=integration -run "TestBaseline" ./internal/harness/...

# Single test with verbose output
go test -v -tags=integration -run "TestSecurity_ServiceDuringSlowloris" ./internal/harness/...
```

---

## References

- [Slowloris Attack](https://en.wikipedia.org/wiki/Slowloris_(computer_security))
- [Go HTTP Server Timeouts](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/)
- [Rate Limiting Patterns](https://pkg.go.dev/golang.org/x/time/rate)
