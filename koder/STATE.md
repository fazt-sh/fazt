# Fazt Implementation State

**Last Updated**: 2026-02-07
**Current Version**: v0.28.0

## Status

State: BLOCKED — Issue 05 (Test DB Connection Hang)
Working on: Plan 46 Phase 1 mostly complete, blocked by database connection hang in CMD gateway tests
Next: Fix Issue 05, then finish Phase 1.4 edge cases

---

## Current Session (2026-02-07) — Plan 46: Test Coverage Overhaul

### Context

Two production bugs escaped because tests didn't catch them:
1. `/api/cmd` routing bug (auth bypass misconfiguration) — v0.28.0
2. `ResolveAlias` 'app' type bug (schema mismatch) — admin deployment

Deep coverage audit revealed **31% overall coverage** with critical gaps:
- **Routing**: 1.4% (140+ untested lines in `createRootHandler`)
- **Auth Middleware**: 11.9% (146 lines, zero coverage on core functions)
- **Handlers**: 7.3% (17 handlers completely untested, 2,900+ lines)
- **Database**: 4.4%

Created **Plan 46** (`koder/plans/46_test_coverage_overhaul.md`):
- **Goal**: 31% → 85% coverage over 5 weeks
- **Priority**: Security > Reliability > Performance > Features
- **4 Phases**: Week 1 (critical), Weeks 2-3 (systematic), Week 4 (integration), Week 5 (security)

### What Was Done — Phase 1.1: Routing Tests (COMPLETE ✅)

**Created**: `cmd/server/main_routing_test.go` (1,045 lines)

**20 comprehensive test functions** covering all critical routing paths:

1. **Host Routing Tests**
   - `TestRouting_AdminDomain_APIBypass` — 9 bypass endpoints (deploy, cmd, sql, upgrade, health, logs, users, aliases, apps/*/status)
   - `TestRouting_AdminDomain_AdminMiddleware` — Auth + role enforcement (no auth → 401, user role → 403, admin role → 200)
   - `TestRouting_AdminDomain_TrackEndpoint` — Public endpoint (no auth required)
   - `TestRouting_LocalhostSpecialCase` — Localhost always serves dashboard
   - `TestRouting_RootDomain` — Both `root.domain` and bare `domain` serve root site
   - `TestRouting_404Domain` — `404.domain` serves 404 site
   - `TestRouting_SubdomainRouting` — `app.domain` routing

2. **Local-Only Routes** (`/_app/<id>/`)
   - `TestRouting_LocalOnlyRoutes_FromLocal` — Local IPs allowed (127.0.0.1, ::1, 192.168.*, 10.*)
   - `TestRouting_LocalOnlyRoutes_FromPublic` — Public IPs blocked (404, not 401 to avoid revealing route)

3. **Auth Routes** (`/auth/*`)
   - `TestRouting_AuthRoutes_AvailableEverywhere` — Available on all hosts (admin, root, subdomains, localhost)
   - `TestRouting_LoginRoute_PostOnly` — POST /auth/login routing

4. **Middleware Order**
   - `TestRouting_MiddlewareOrder_AuthBeforeAdmin` — AdminMiddleware checks auth first

5. **Port Stripping**
   - `TestRouting_PortStripping` — Handles :8080, :443, :3000
   - `TestRouting_IPv6_PortStripping` — IPv6 [::1]:8080 doesn't crash

6. **Edge Cases**
   - `TestRouting_EmptyHost` — Doesn't crash
   - `TestRouting_UnknownSubdomain_Fallback` — Serves 404 site
   - `TestRouting_CaseSensitivity` — Documents case behavior
   - `TestRouting_PathPrecedence_BypassBeforeAdmin` — /api/deploy bypasses AdminMiddleware
   - `TestRouting_PathPrecedence_AppsStatusBypass` — /api/apps/*/status bypasses, /api/apps/* requires admin
   - `TestRouting_AdminDomain_Fallthrough` — Non-API paths on admin.* fall through to app serving

**Test Statistics**:
- 20 parent test functions
- 60+ subtests (via t.Run)
- ~200 test cases total (as planned in Plan 46)
- All tests passing ✅

**Infrastructure Created**:
- `setupRoutingTestDB()` — In-memory SQLite with full auth schema (auth_users, auth_sessions, files)
- `setupRoutingTestConfig()` — Test config with domain "test.local"
- `setupTestHandlers()` — Initializes handler globals (auth service, rate limiter)
- `createTestUser()` — Uses `authService.CreateUser()` (proper UUID generation)
- `createTestSession()` — Uses `authService.CreateSession()` (proper token hashing)

**Key Fixes During Implementation**:
1. **Auth schema mismatch** — Test schema had `id` but production uses `token_hash` + `user_id TEXT`
2. **Session cookie name** — Changed from "session" to "fazt_session"
3. **User ID type** — Changed from `INTEGER` to `TEXT` (UUIDs)
4. **Session creation** — Use `authService.CreateSession()` not manual INSERT (handles token hashing)
5. **User creation** — Use `authService.CreateUser()` not manual INSERT (generates UUIDs)
6. **Hosting init** — Call `hosting.Init(db)` to initialize VFS (files table queries)
7. **Handler init** — Call `handlers.InitAuth()` to initialize rate limiter

**Coverage Impact**: Routing coverage 1.4% → ~90% (estimated, will measure)

---

### What Was Done — Phase 1.2: Auth Middleware Tests (COMPLETE ✅)

**Created**: `internal/middleware/auth_test.go` (529 lines)

**Coverage Added**:
- `AuthMiddleware()`:
  - Public path bypass, valid/invalid bearer token, empty/malformed bearer
  - Valid/invalid/expired session cookies
  - No auth (HTML redirect), API vs HTML response split
- `AdminMiddleware()`:
  - No session, invalid session, user role → 403, admin/owner → 200
- `requiresAuth()`:
  - Exact public paths, public prefixes, protected paths
  - Case sensitivity, trailing slash behavior, edge cases, path traversal

---

### What Was Done — Phase 1.3: Schema Sync (COMPLETE ✅)

**Goal**: Ensure tests use production migrations instead of handwritten schema.

**Changes**:
- Added `database.RunMigrations(db)` helper (exposed in `internal/database/db.go`).
- `internal/handlers/handlers_test.go` now uses embedded migrations for setup.
- `internal/handlers/auth_test.go` now reuses `setupTestDB()`.
- Added `internal/handlers/schema_test.go`:
  - FK enforcement (auth_sessions → auth_users)
  - UNIQUE constraints (auth_users.email)
  - DEFAULT values (apps.visibility/source)

**Note**: Full schema equality diff test still not implemented (future work if needed).

---

### What Was Done — Phase 1.4: Critical Handlers (PARTIAL ✅)

**Added Tests**:
- `internal/handlers/auth_admin_test.go`:
  - `requireAPIKeyAuth` (missing/invalid/valid)
  - `requireAdminAuth` (no session, user/admin/owner roles, API key)
- `internal/handlers/deploy_handler_test.go`:
  - Method not allowed, missing auth, bad auth format, invalid API key
  - Missing site_name, invalid file type, success path
- `internal/handlers/sql_handler_test.go`:
  - Method not allowed, invalid JSON, empty query
  - Write requires `write: true`, select success, write success
- `internal/handlers/agent_handler_test.go`:
  - info missing header + success
  - storage get not found + JSON value
  - snapshot + restore end-to-end

**Fix Applied**:
- `internal/hosting/deploy.go`: close API key rows before updating `last_used_at` to avoid SQLite locks in tests.

---

## What Was Done — Rate Limiter Goroutine Leak Fix (COMPLETE ✅)

**Problem**: Full test suite was timing out after ~60s due to goroutine leaks
- Each `auth.NewRateLimiter()` started a background cleanup goroutine
- Tests never stopped these goroutines, causing accumulation
- Eventually exhausted resources and hung

**Solution** (Commit 82329bc):
- Added `done` channel to `RateLimiter` and `DeployLimiter` structs
- Added `Stop()` methods to gracefully shutdown cleanup goroutines
- Updated cleanup loops to `select` on ticker and done channels
- Updated 11+ test files to call `limiter.Stop()` in `t.Cleanup()`

**Results**:
- ✅ Individual tests pass cleanly
- ✅ Middleware tests pass (0.211s)
- ✅ Routing tests pass (0.049s)
- ❌ Full suite still hangs (different issue - see below)

---

## BLOCKER — Issue 05: Test Database Connection Hang

**Discovered during rate limiter fix verification**.

**Symptoms**:
- `go test ./internal/handlers -count=1` times out after 28s
- Hangs on `TestCmdGateway_AcceptsValidAPIKey`
- Stuck waiting for database connection in `getAliasesForApp()`

**Not related to rate limiters** - This is a separate database connection leak issue.

**Impact**:
- ❌ Cannot run full handler test suite
- ❌ Blocks Plan 46 Phase 1.4 completion
- ❌ Cannot measure coverage improvements

**Details**: See `koder/issues/05_test-db-connection-hang.md`

**Assignment**: **OPUS** - Needs deep investigation

---

## What's Next — FOR OPUS

### Priority 1: Fix Issue 05 (Database Connection Hang)

**File**: `koder/issues/05_test-db-connection-hang.md`

**Quick summary**:
- Tests that fail auth early → PASS
- Tests that execute commands → HANG on database query
- Likely: connection not returned to pool after `ValidateAPIKey()` or cmd execution
- Setup: In-memory SQLite with `SetMaxOpenConns(1)`

**Suggested investigation**:
1. Add debug logging around `db.Query()` / `rows.Close()` lifecycle
2. Check if `ValidateAPIKey()` truly releases connection (despite commit 4836058 fix)
3. Look for unclosed transactions in cmd gateway flow
4. Consider increasing `SetMaxOpenConns` for tests (workaround)

### Priority 2: Finish Phase 1.4 Edge Cases (after Issue 05 fixed)

Once full suite can run:
- Expand `auth_handlers.go` tests (rate limiter exhaustion, remember_me TTL)
- Add deploy handler ZIP/path traversal/oversize tests
- Add SQL handler error-path tests
- Measure actual coverage (not just estimates)

---

## Quick Reference

```bash
# Routing
go test -v ./cmd/server -run TestRouting

# Middleware
go test ./internal/middleware -count=1

# Handlers (targeted)
go test ./internal/handlers -run TestRequireAPIKeyAuth -count=1 -v
go test ./internal/handlers -run TestRequireAdminAuth -count=1 -v
go test ./internal/handlers -run TestDeployHandler -count=1 -v
go test ./internal/handlers -run TestHandleSQL -count=1 -v
go test ./internal/handlers -run TestAgent -count=1 -v

# Full handlers suite
go test ./internal/handlers -count=1
```

---

## Session Summary (2026-02-07)

### Codex Session (Phase 1.2-1.4)
**Tests Added** (all committed):
- ✅ Middleware auth tests (`internal/middleware/auth_test.go`) - 529 lines
- ✅ Schema sync + constraint tests - migrations in tests
- ✅ Admin/API key gate tests (`internal/handlers/auth_admin_test.go`) - 209 lines
- ✅ Deploy handler tests (`internal/handlers/deploy_handler_test.go`) - 184 lines
- ✅ SQL handler tests (`internal/handlers/sql_handler_test.go`) - 148 lines
- ✅ Agent handler tests (`internal/handlers/agent_handler_test.go`) - 162 lines

**Issues Found**:
- ❌ Goroutine leaks (rate limiters not stopped) - **FIXED by Claude**
- ❌ Database connection hang in CMD tests - **OPEN (Issue 05)**

### Claude Session (Review + Fix)
**Reviewed Codex's work**:
- Code quality: 7/10 - Good patterns, comprehensive coverage
- Integration testing: Missing - didn't verify full suite runs
- Found critical goroutine leak preventing full suite execution

**Fixed**:
- ✅ Added `Stop()` methods to rate limiters (commit 82329bc)
- ✅ Updated all test files with proper cleanup
- ✅ Documented DB hang issue (koder/issues/05)

**Handoff to Opus**:
- 🎯 **Primary task**: Fix Issue 05 (database connection hang)
- 🎯 **Secondary**: Complete Phase 1.4 edge cases after unblocking
- 📊 **Goal**: Get full test suite passing, measure coverage

---

## Key Learnings

1. **Always run full test suite** - Individual tests passing ≠ suite passing
2. **Goroutine cleanup is critical** - Background goroutines need explicit Stop()
3. **Database connection limits** - `SetMaxOpenConns(1)` + leaks = deadlock
4. **Test isolation matters** - Global state (`database.SetDB()`) can cause races
5. **Rate limiters in tests** - Need cleanup or tests accumulate goroutines

---

## For Opus: Quick Start

```bash
# Reproduce the hang:
go test ./internal/handlers -run TestCmdGateway_AcceptsValidAPIKey -timeout=10s

# Read the issue:
cat koder/issues/05_test-db-connection-hang.md

# Check Plan 46:
cat koder/plans/46_test_coverage_overhaul.md

# When fixed, measure coverage:
go test ./internal/handlers -coverprofile=coverage.out
go tool cover -func=coverage.out | grep "total:"
```
