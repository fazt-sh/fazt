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

### Immediate Action Required

```go
// cmd/server/server.go - Add these timeouts
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

1. Add HTTP server timeouts (fixes slowloris + slow headers)
2. Verify/implement rate limiting
3. Improve runtime recovery under load

See: `koder/plans/27_harness_findings.md`

### Plan 24: Mock OAuth Provider

Enable full auth flow testing locally without code changes.

```
Local:  "Sign in" → Dev form → Session → fazt.auth.getUser() ✓
Remote: "Sign in" → Google   → Session → fazt.auth.getUser() ✓
```

Same code. Same API. Different provider.

See: `koder/plans/24_mock_oauth.md`

### Plan 25: SQL Command

Debug local and remote fazt instances with direct SQL queries.

```bash
fazt sql "SELECT * FROM apps"              # Local
fazt @zyt sql "SELECT * FROM auth_users"   # Remote
```

See: `koder/plans/25_sql_command.md`

---

## Last Session

**Plan 26: Harness Refactor (COMPLETED)**

Converted test harness from embedded binary code to proper Go integration
tests using `_test.go` files with `//go:build integration` tag.

### Changes Made

1. **Created integration test files**
   - `testutil.go` - Shared test helpers (getTarget, HTTP client, assertions)
   - `requests_test.go` - Static, API read/write, auth, serverless tests
   - `resilience_test.go` - Memory, timeout, queue tests
   - `baseline_test.go` - Latency, throughput, resource tests
   - `security_test.go` - Rate limit, payload, slowloris tests

2. **Created test-harness app** (`servers/local/test-harness/`)
   - Serverless endpoints: `/api/health`, `/api/hello`, `/api/echo`, `/api/slow`, `/api/timeout`
   - Uses nip.io wildcard DNS for subdomain routing
   - Deploy: `fazt app deploy servers/local/test-harness --to local`

3. **Removed dead code**
   - `cmd/server/harness.go` - CLI handler
   - `internal/harness/harness.go` - Orchestrator
   - `internal/harness/baseline/`, `requests/`, `resilience/`, `security/` directories

4. **Kept as library code**
   - `config.go` - Types (Expected, Actual, TestResult, etc.)
   - `report.go` - Report generation
   - `gaps/` - Gap tracking

### Usage

```bash
# Deploy test app first
fazt app deploy servers/local/test-harness --to local

# Run integration tests
FAZT_TARGET="http://test-harness.192.168.64.3.nip.io:8080" \
FAZT_TOKEN="<api-key>" \
FAZT_TEST_APP="test-harness.192.168.64.3.nip.io" \
go test -v -tags=integration ./internal/harness/...
```

### Test Categories

| Category | Tests | Purpose |
|----------|-------|---------|
| Baseline | HealthLatency, Throughput, MixedWorkload | Performance baselines |
| Static | Health, DotfileBlocked, PathTraversal | Static file security |
| API | Health, NonexistentEndpoint, HealthLatency | API behavior |
| Serverless | BasicExecution, RequestBody, Timeout | Runtime validation |
| Resilience | HighConcurrency, SustainedLoad, GCPressure | Stress testing |
| Security | RateLimit, Slowloris, ConnectionFlood | Attack resilience |

Tests that require admin endpoints or KV storage skip gracefully when
running against a test app.

### Benefits

- **No binary bloat** - `_test.go` never compiled into release
- **Standard tooling** - Works with `go test`, CI, coverage
- **Selective execution** - Build tags separate unit/integration
- **Better reporting** - `go test -json` for structured output

---

## Previous Session

**Test Harness + Timeout Budget System**

1. **Timeout/Budget System** (`internal/timeout/`)
   - `budget.go` - Budget tracking for serverless execution
   - `budget_test.go` - Unit tests (passing)
   - Storage bindings updated to use budget for admission control

2. **Test Harness** (original implementation, now refactored)
   - Performance baselines: 44k RPS static, 40k RPS API @ 100 concurrent
   - P50 latency: 176-304µs
   - P99 latency: 500µs-1ms

---

## Quick Reference

```bash
# Session commands (this repo)
/open                    # Start session
/close                   # End session
/release                 # Release workflow

# Global skills (any repo)
/fazt-app               # Build fazt apps

# Integration tests
FAZT_TARGET=http://localhost:8080 go test -tags=integration ./internal/harness/...

# Check OAuth status
fazt @<peer> auth providers
```
