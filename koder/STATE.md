# Fazt Implementation State

**Last Updated**: 2026-01-29
**Current Version**: v0.11.5

## Status

State: PENDING REFACTOR - harness needs conversion to integration tests

---

## Next Up

### Plan 26: Harness Refactor (PRIORITY)

Test harness was incorrectly embedded in fazt binary. Refactor to proper Go
integration tests using `_test.go` files with `//go:build integration` tag.

```bash
# After refactor
FAZT_TARGET=http://localhost:8080 go test -tags=integration ./internal/harness/...
```

See: `koder/plans/26_harness_refactor.md`

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

**Test Harness + Timeout Budget System**

### Committed Work

1. **Timeout/Budget System** (`internal/timeout/`)
   - `budget.go` - Budget tracking for serverless execution
   - `budget_test.go` - Unit tests (passing)
   - Storage bindings updated to use budget for admission control

2. **Test Harness** (`internal/harness/`)
   - Full implementation committed
   - Has design flaw: embedded in binary (should be `_test.go` files)
   - Plan 26 documents the refactor needed
   - Next session: execute Plan 26 to convert

### Harness Performance Results (localhost:8080)

Before refactoring, the harness validated these baselines:

- Static read: 44,119 RPS @ 100 concurrency
- API read: 39,649 RPS @ 100 concurrency
- P50 latency: 176-304µs
- P99 latency: 500µs-1ms

### Harness Structure (to be refactored)

```
internal/harness/
├── harness.go           # Main orchestrator → merge into tests
├── config.go            # Configuration → keep as library
├── report.go            # Report generation → keep as library
├── baseline/            # → baseline_test.go
├── requests/            # → requests_test.go
├── resilience/          # → resilience_test.go
├── security/            # → security_test.go
└── gaps/                # Gap tracker → keep as library
```

---

## Previous Session

**fazt-app Skill Documentation Fixes**

1. **Fixed deployment docs** - `fazt app deploy` has built-in build
2. **Fixed OAuth redirect bug in docs** - Must use absolute URLs
3. **Removed hardcoded zyt.app references**
4. **Created Plan 25: SQL Command**

---

## Quick Reference

```bash
# Session commands (this repo)
/open                    # Start session
/close                   # End session
/release                 # Release workflow

# Global skills (any repo)
/fazt-app               # Build fazt apps

# Check OAuth status
fazt @<peer> auth providers
```
