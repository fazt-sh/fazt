# Fazt Implementation State

**Last Updated**: 2026-02-07
**Current Version**: v0.28.0

## Status

State: IN PROGRESS — Plan 46 Phase 1 (Test Coverage Overhaul)
Working on: Phase 1.1 COMPLETE ✅, Phase 1.2 next

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
2. **Session cookie name** — Changed from `"session"` to `"fazt_session"`
3. **User ID type** — Changed from `INTEGER` to `TEXT` (UUIDs)
4. **Session creation** — Use `authService.CreateSession()` not manual INSERT (handles token hashing)
5. **User creation** — Use `authService.CreateUser()` not manual INSERT (generates UUIDs)
6. **Hosting init** — Call `hosting.Init(db)` to initialize VFS (files table queries)
7. **Handler init** — Call `handlers.InitAuth()` to initialize rate limiter

**What This Achieves**:
- ✅ Routing config changes will break tests
- ✅ Auth bypass changes will break tests
- ✅ Middleware order changes will break tests
- ✅ Path precedence validated
- ✅ Host routing validated

**Coverage Impact**: Routing coverage 1.4% → ~90% (estimated, will measure)

---

## What's Next — Phase 1.2: Middleware Tests

**Goal**: Test `internal/middleware/auth.go` (146 lines, currently 11.9% coverage)

**Target File**: `internal/middleware/auth_test.go` (expand existing)

**Functions to Test** (currently 0% coverage):

1. **`AuthMiddleware()`** — Core auth enforcement
   ```go
   TestAuthMiddleware_NoAuthRequired()     // Public paths pass through
   TestAuthMiddleware_BearerToken_Valid()  // API key auth succeeds
   TestAuthMiddleware_BearerToken_Invalid() // Bad token → redirect/401
   TestAuthMiddleware_Session_Valid()      // Session cookie auth succeeds
   TestAuthMiddleware_Session_Expired()    // Expired → redirect/401
   TestAuthMiddleware_Session_Invalid()    // Bad session → redirect/401
   TestAuthMiddleware_NoAuth()             // No auth → redirect/401
   TestAuthMiddleware_APIvsHTML()          // /api/* → 401 JSON, HTML → redirect
   ```

2. **`AdminMiddleware()`** — Role-based access control
   ```go
   TestAdminMiddleware_NoSession()         // No auth → 401 JSON
   TestAdminMiddleware_UserRole()          // User role → 403 JSON
   TestAdminMiddleware_AdminRole()         // Admin role → 200
   TestAdminMiddleware_OwnerRole()         // Owner role → 200
   TestAdminMiddleware_InvalidSession()    // Invalid session → 401
   ```

3. **`requiresAuth()`** — Path whitelist logic
   ```go
   TestRequiresAuth_PublicPaths()          // /, /login.html, /health, etc.
   TestRequiresAuth_PublicPrefixes()       // /track, /r/, /webhook/, /static/, /assets/
   TestRequiresAuth_AuthPaths()            // /auth/*, /api/login, /api/deploy
   TestRequiresAuth_ProtectedPaths()       // /api/*, /dashboard/*, etc.
   TestRequiresAuth_CaseSensitivity()      // Path matching is case-sensitive
   TestRequiresAuth_TrailingSlash()        // /static vs /static/
   ```

4. **Edge Cases**
   ```go
   TestAuthMiddleware_EmptyBearerHeader()  // "Bearer " with no token
   TestAuthMiddleware_MalformedBearer()    // "Bearer" without space, "Token xyz", etc.
   TestAuthMiddleware_PathTraversal()      // /../api/login, /static/../api/apps
   TestRequiresAuth_EdgeCases()            // Empty path, ., .., etc.
   ```

**Implementation Pattern** (follow routing tests):
```go
func TestAuthMiddleware_BearerToken_Valid(t *testing.T) {
    db := setupTestDB(t)
    authService := auth.NewService(db, "test.local", false)

    // Create API key
    apiKey := createTestAPIKey(t, db)

    // Create middleware
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(200)
        w.Write([]byte("success"))
    })
    authMW := middleware.AuthMiddleware(authService)

    // Test request
    req := httptest.NewRequest("GET", "/api/apps", nil)
    req.Header.Set("Authorization", "Bearer "+apiKey)

    rr := httptest.NewRecorder()
    authMW(handler).ServeHTTP(rr, req)

    if rr.Code != 200 {
        t.Errorf("Expected 200, got %d", rr.Code)
    }
}
```

**Estimated Time**: 4-6 hours (as per Plan 46)
**Expected Coverage**: Middleware 11.9% → ~95%

---

## What's After That — Phase 1.3 & 1.4

### Phase 1.3: Schema Sync (~4-6 hours)

**Goal**: Test DB schema === Production DB schema

**Current Problem**: Tests use handwritten schema, missing tables:
- `apps`, `aliases`, `peers`, `auth_users`, `auth_sessions`
- `storage_keys`, `storage_objects`, `activity_log`
- `net_allowlist`, `net_secrets`, `net_log`, `workers`

**Solution**: Use production migrations in tests
```go
// OLD approach (handlers_test.go)
schema := `CREATE TABLE ...`  // Handwritten, drifts over time

// NEW approach
func setupTestDB(t *testing.T) *sql.DB {
    db := createMemoryDB()
    // Run all migrations in order
    RunMigrations(db, "../../database/migrations")
    return db
}
```

**Tests to Add**:
- `TestSchemaEquality()` — Test DB columns === Production DB columns
- `TestForeignKeyConstraints()` — All FKs enforced
- `TestUniqueConstraints()` — All UNIQUE constraints work
- `TestDefaultValues()` — DEFAULT values match production

**File**: `internal/handlers/handlers_test.go` (modify setup)

### Phase 1.4: Critical Handlers (~16-20 hours)

Test 4 highest-risk handlers (by security + untested lines):

1. **`auth_handlers.go`** (462 lines)
   - Login rate limiting, password verification, session creation
   - Invalid credentials, SQL injection attempts
   - Session token validation, expiry handling

2. **`deploy.go`** (170 lines)
   - API key validation, rate limiting (5/min)
   - File upload size limits, path traversal
   - ZIP bomb protection, malicious archives

3. **`sql.go`** (169 lines) — ADMIN SQL (!!)
   - Admin-only access (non-admin rejected)
   - SQL injection prevention (parameterized queries)
   - Read-only enforcement (if applicable)

4. **`agent_handler.go`** (474 lines)
   - Agent authentication, authorization
   - API key scoping, resource access control

**Pattern**: Follow existing handler tests (`cmd_gateway_test.go`, `aliases_test.go`)

---

## Files Changed (Not Committed Yet)

```
cmd/server/main_routing_test.go          (NEW, 1,045 lines)
```

---

## Commit Plan

```bash
# Commit routing tests
git add cmd/server/main_routing_test.go
git commit -m "test: add comprehensive routing tests (Plan 46 Phase 1.1)

- 20 test functions, 60+ subtests, ~200 test cases
- Host routing: admin, root, 404, subdomains, localhost
- Auth bypass: 9 endpoints tested
- AdminMiddleware: role enforcement validated
- Local-only routes: /_app/<id>/ IP filtering
- Edge cases: port stripping, IPv6, case sensitivity
- All tests passing ✅

Coverage: Routing 1.4% → ~90% (est)

Part of Plan 46: Test Coverage Overhaul (31% → 85%)
Ref: koder/plans/46_test_coverage_overhaul.md"
```

---

## Quick Reference

```bash
# Run routing tests
go test -v ./cmd/server -run TestRouting

# Run all tests
go test ./... -short -count=1

# Check coverage
go test ./cmd/server -coverprofile=coverage.out
go tool cover -html=coverage.out

# Next: Implement middleware tests
# File: internal/middleware/auth_test.go
# Pattern: Follow routing tests approach
```

---

## Session Notes

**Time Spent**: ~3 hours (routing tests)
**Tests Written**: 20 functions, ~200 cases
**All Passing**: ✅

**Next Agent**:
1. Read this STATE.md
2. Read Plan 46 Phase 1.2 section (lines 98-122)
3. Implement middleware tests in `internal/middleware/auth_test.go`
4. Follow pattern from `cmd/server/main_routing_test.go`
5. Aim for ~95% middleware coverage
6. Run tests, commit, update STATE.md

**Key Learnings**:
- Always use `authService.CreateUser()` and `authService.CreateSession()` instead of manual DB inserts
- Auth schema uses `token_hash` (not `id`), `user_id TEXT` (not INTEGER)
- Session cookie name is `"fazt_session"`
- Initialize hosting (`hosting.Init(db)`) and handlers (`handlers.InitAuth()`) in test setup
- Test routing logic, not handler implementation (routing tests shouldn't need to mock everything)
