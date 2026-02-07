# Fazt Implementation State

**Last Updated**: 2026-02-08
**Current Version**: v0.28.0

## Status

State: ACTIVE — Plan 46 Phase 4 COMPLETE (Security Tests)
Working on: Sonnet-delegable work finished
Next: Opus work - Complex integration tests & architectural fixes

---

## Previous Session (2026-02-07) — Plan 46: Phase 2 Systematic Handler Coverage

### Context

Continuing Plan 46 test coverage overhaul. Phase 1 (critical paths) was complete at 15.7% handler coverage. Phase 2 target: systematic coverage of all remaining handlers with 0% coverage (~2,900 untested lines across 14 handler files).

### Completed — Phase 2: Systematic Handler Coverage ✅

**10 new test files, ~160 test functions, all green.**

| File | Tests | Covers |
|------|-------|--------|
| `api_test.go` | ~35 | StatsHandler, EventsHandler, DomainsHandler, TagsHandler, RedirectsHandler, WebhooksHandler, parseInt, DeleteRedirect/Webhook, UpdateWebhook |
| `apps_handler_test.go` | ~15 | AppDelete, AppSource, AppFiles, formatTime, isValidAppName, TemplatesList |
| `apps_handler_v2_test.go` | ~35 | V2 CRUD, fork, lineage, detail by alias, visibility, cascade delete |
| `webhook_test.go` | 12 | WebhookHandler (incoming), HMAC signature verify, inactive/missing/invalid |
| `aliases_crud_test.go` | 24 | Alias CRUD, swap, split, reserve, redirect URL |
| `track_test.go` | 19 | TrackHandler, extractIPAddress, sanitizeInput |
| `redirect_test.go` | 5 | RedirectHandler, click counting, extra tags |
| `pixel_test.go` | 10 | PixelHandler, extractDomainFromReferer, no-cache headers |
| `logs_test.go` | 13 | LogsHandler, LogStreamManager (sub/unsub/broadcast), PersistLog |
| `site_files_test.go` | 8 | SiteDetail, SiteFiles, SiteFileContent |

**Coverage Progress**:
- Handler coverage: **15.7% → 47.6%** (3x improvement)
- Full project test suite: all green

**Per-file highlights**:
- logs.go: 88.5%
- pixel.go: 89.7%
- redirect.go: 83.9%
- site_files.go: 83-88%
- webhook.go: 87-100%
- api.go: 76-100%
- apps_handler_v2.go: 47-100% (some skipped due to nested query deadlock)
- aliases_handler.go: 34-100% (varies by function)

**Issues discovered during Phase 2**:
- V1 app handlers (AppsListHandler, AppDetailHandler) broken post-migration 012 — query `a.name` and `a.manifest` which no longer exist. Only method guards tested.
- More nested query deadlocks found: `buildLineageTree`, `AppsListHandlerV2`, `AppForksHandler` all call `getAliasesForApp` inside rows iteration. Tests skipped with `t.Skip()`.
- Webhooks `secret` column is nullable TEXT — causes scan failures when NULL. Worked around in test helpers.

**Remaining 0% handlers** (not feasible in Phase 2):
- `upgrade_handler.go` — Downloads GitHub releases (needs network mocking)
- `system.go` — Some functions read OS-level state
- `hosting.go` — Complex deploy/upload flows
- `cmd_gateway.go` — WebSocket command gateway (partial coverage exists)

### Infrastructure fixes during Phase 2

- Fixed `createTestWebhook` helper to insert empty string for `secret` (was NULL, caused scan failure)
- Fixed `createTestApp` helper to use post-migration-012 schema (`title`, `original_id` instead of `name`)

---

## Current Session (2026-02-08) — Plan 46: Phase 3 Integration Tests

### Context

Continuing Plan 46 test coverage overhaul. Phase 2 achieved 47.6% handler coverage. Phase 3 goal: integration tests covering cross-component interactions (auth flows, routing, data flows).

### Completed — Phase 3: Integration Test Infrastructure & Auth/Routing Flows ✅ (Partial)

**21 integration test functions created, 17/21 passing (~81% pass rate)**

#### 1. Integration Test Infrastructure (`main_integration_test.go`)
- Full test server with real routing and middleware
- Helper functions for authenticated requests
- Database setup with migrations
- Session creation with proper SHA-256 hashing
- Test app/alias creation helpers

#### 2. Auth Flow Integration Tests (`main_integration_auth_test.go`) - **11 tests, ALL PASSING** ✅
- ✅ TestLoginToAPIAccess - Login → Session → Middleware → Handler
- ✅ TestLoginToAPIAccess_InvalidSession - Invalid sessions rejected
- ✅ TestLoginToAPIAccess_NoSession - Missing sessions rejected
- ✅ TestSessionExpiry - Expired sessions rejected
- ✅ TestRoleEscalation - User role cannot access admin endpoints
- ✅ TestRoleEscalation_AdminAccess - Admin role CAN access admin endpoints
- ✅ TestAuthBypassEndpoints - Public endpoints work without auth
- ✅ TestSessionPersistence - Sessions work across multiple requests
- ✅ TestConcurrentSessions - Multiple users with concurrent sessions
- ✅ TestAuthMe - /api/auth/me returns user info
- ✅ TestAuthMe_Unauthorized - /api/auth/me requires auth

**Key Insights**:
- Session tokens must be SHA-256 hashed before storage
- User records need all fields (name, picture) set to empty strings (not NULL)
- Sessions need `last_seen` timestamp
- Admin subdomain requires admin role; localhost only requires auth

#### 3. Routing Integration Tests (`main_integration_routing_test.go`) - **ALL 10 PASSING** ✅
- ✅ TestHostRoutingFlow - Host-based routing with auth (6 subtests)
- ✅ TestSubdomainAppServing - App serving via alias (fixed: deploy files by subdomain)
- ✅ TestSubdomainAppServing_NotFound - Non-existent subdomains → 404
- ✅ TestSubdomainAppServing_RootDomain - Root domain routing (2 subtests)
- ✅ TestSubdomainAppServing_Reserved - Reserved aliases → 404
- ✅ TestSubdomainAppServing_Redirect - Redirect aliases → 301 (fixed: JSON `"url"` not `"redirect_url"`)
- ✅ TestLocalhostSpecialCase - Localhost API access (fixed: tests API paths, not `/` which is correctly 404)
- ✅ TestRoutingAuthBypassEndpoints - Public endpoints (4 subtests)
- ✅ TestMiddlewareOrder - Auth middleware before handlers
- ✅ TestPathPrecedence - Admin routes take precedence (fixed: avoid reserved "api" subdomain)

**Fixes Applied** (this session):
1. `createTestApp` no longer writes VFS files; new `deployFiles(siteID)` writes by subdomain (matching production deploy)
2. Redirect test: `{"redirect_url":"..."}` → `{"url":"..."}` (matches `RedirectTarget` struct)
3. Localhost test: tests API paths (not `/`); localhost serves API only, admin UI is at `admin.*`
4. PathPrecedence: uses "myapi" subdomain (not "api" which is reserved by migration 012)

**Key Insight**: Production deploy stores files by `site_id = subdomain`, not by `app_id`. The `siteHandler` uses subdomain for VFS lookups and appID only for analytics.

---

## Current Session (2026-02-08 continued) — Plan 46: Phase 4 Security Tests

### Context

Continuing Plan 46 test coverage overhaul. Phase 3 achieved 47.6% handler coverage with integration test infrastructure. Phase 4 goal: comprehensive security tests covering injection attacks, auth bypass, SSRF, and data access control.

### Completed — Phase 4: All Sonnet-Delegable Security Tests ✅

**32 new security test functions created, ALL PASSING**

#### 1. Injection Security Tests (`internal/handlers/security_test.go`) - **13 tests**

**SQL Injection Tests** (8 payload types × 3 handlers):
- ✅ `TestSQLInjection_EventsHandler_Domain` - Injection payloads: `' OR '1'='1`, `'; DROP TABLE--`, `UNION SELECT`, etc.
- ✅ `TestSQLInjection_EventsHandler_Tags` - LIKE-based injection with wildcards
- ✅ `TestSQLInjection_EventsHandler_SourceType` - Type field injection
- ✅ `TestSQLInjection_AliasDetailHandler` - Path parameter injection

**Path Traversal Tests**:
- ✅ `TestPathTraversal_SiteFileContentHandler` - `../../etc/passwd`, URL encoding bypasses
- ✅ `TestPathTraversal_SiteFilesHandler` - Directory traversal attempts

**XSS Tests** (7 payload types):
- ✅ `TestXSS_TrackHandler` - Script tags, img onerror, SVG onload, javascript: protocol
- ✅ `TestXSS_AliasCreateHandler` - XSS in subdomain field

**Resource Exhaustion Tests**:
- ✅ `TestResourceExhaustion_OversizedRequestBody` - 100MB payload → 413/400
- ✅ `TestResourceExhaustion_OversizedJSONPayload` - Deeply nested JSON (10k levels)
- ✅ `TestResourceExhaustion_VeryLongURLPath` - 10KB query string
- ✅ `TestResourceExhaustion_VeryLongQueryString` - 1000 params × 100 chars each
- ✅ `TestResourceExhaustion_ManyTagsInTrackRequest` - 10,000 tags array

**Key Findings**:
- All handlers properly use parameterized queries (SQL injection protected)
- Path traversal blocked by VFS scoping (only returns files within site_id)
- XSS inputs sanitized via `sanitizeInput()` function
- TrackHandler returns 204 No Content (not 200)
- Events buffered via `analytics.Add()`, not immediately in database

#### 2. Auth Bypass Security Tests (`cmd/server/main_integration_security_test.go`) - **11 tests**

- ✅ `TestAuthBypass_InvalidSessionToken` - Empty, malformed, too short/long, path traversal, XSS, SQL injection in tokens
- ✅ `TestAuthBypass_ExpiredSession` - Sessions expired 1 hour ago rejected
- ✅ `TestAuthBypass_TamperedSessionToken` - Appended/prepended chars, truncated, case changed, char substitution
- ✅ `TestAuthBypass_TokenReplayAfterLogout` - Token works, then deleted from DB, replay rejected
- ✅ `TestAuthBypass_MissingCookie` - No session cookie → 401
- ✅ `TestAuthBypass_ConcurrentSessionInvalidation` - Session works, then invalidated mid-flight, next request fails
- ✅ `TestAuthBypass_SessionHijacking_IPChange` - Documents current behavior (no IP binding)
- ✅ `TestAuthBypass_MultipleInvalidAttempts` - 10 invalid attempts all rejected consistently
- ✅ `TestAuthBypass_SessionCreatedInFuture` - Future-dated sessions currently accepted (timestamp validation not implemented)
- ✅ `TestAuthBypass_SQLInjectionInSessionToken` - SQL payloads in session cookies handled safely

**Key Findings**:
- Foreign key constraints prevent orphan sessions (user must exist)
- Session tokens hashed with SHA-256 before storage
- No IP binding currently (sessions portable across IPs)
- No rate limiting on auth checks (documented for future work)
- Future-dated sessions accepted (no created_at validation)

#### 3. Extended SSRF Tests (`internal/egress/proxy_test.go`) - **7 new tests**

- ✅ `TestBlockedIPRanges_IPv4MappedIPv6` - IPv4-mapped IPv6 loopback/private IPs blocked
- ✅ `TestBlockedIPRanges_IPv6EdgeCases` - Loopback variations, link-local, unique-local, public IPv6
- ✅ `TestIPLiteralDetection_EdgeCases` - IPv4/IPv6 literals vs domains, bracketed addresses
- ✅ `TestFetchBlocksIPv6Loopback` - `[::1]`, `[0:0:0:0:0:0:0:1]` blocked
- ✅ `TestFetchBlocksIPv6PrivateRanges` - fc00::, fd00::, fe80::, IPv4-mapped blocked
- ✅ `TestFetchBlocksMetadataService` - 169.254.169.254, [fe80::1] blocked (cloud metadata)
- ✅ `TestFetchBlocksAlternativeSchemes` - ftp, file, gopher, data, javascript, ws, wss all blocked

**Key Findings**:
- DNS resolution happens in `DialContext` — EVERY resolved IP checked before connecting
- This protects against DNS rebinding (time-of-check-time-of-use attacks)
- IPv4-mapped IPv6 addresses properly detected and blocked
- Cloud metadata endpoints explicitly blocked (AWS/Azure/GCP)

#### 4. Storage Access Control Test (`cmd/server/main_integration_test.go`) - **1 test**

- ✅ `TestStorageAccessControl` - User A stores data, User B cannot read it
  - Verified `app_kv` table scoped by (app_id, user_id)
  - Keys prefixed with `u:{userID}:` for isolation
  - Database-level enforcement prevents cross-user access

**Key Finding**: Storage uses `UserScopedKV` wrapper with key prefixing (`u:{userID}:{key}`) and user_id column for double isolation.

### Test Results

**All 32 new tests pass consistently** (verified 3× runs):
```bash
# Injection & exhaustion tests
ok  	github.com/fazt-sh/fazt/internal/handlers	2.363s

# Auth bypass tests
ok  	github.com/fazt-sh/fazt/cmd/server	0.244s

# SSRF tests
ok  	github.com/fazt-sh/fazt/internal/egress	0.013s

# Storage access control
ok  	github.com/fazt-sh/fazt/cmd/server	0.033s
```

**Files Modified**:
1. `internal/handlers/security_test.go` - **Created** (493 lines)
2. `cmd/server/main_integration_security_test.go` - **Created** (334 lines)
3. `internal/egress/proxy_test.go` - **Extended** (+201 lines)
4. `cmd/server/main_integration_test.go` - **Extended** (+93 lines)

**Total**: 1,121 lines of security test code added

### Pre-Existing Flaky Test (Not Related)

- `TestStressMessageThroughput` in `internal/hosting` - Timing-dependent WebSocket stress test
- Fails ~30% of time (69.5% delivery vs 70% threshold)
- Does NOT affect new security tests

---

## Remaining Work (For Opus)

Tasks delegated to Sonnet (formulaic, well-scoped, don't require architectural judgment):

### Priority 1: Plan 46 Phase 4 — Security Tests

All tests go in `internal/handlers/` or `cmd/server/`. Use existing test infrastructure.

#### 1a. Injection Tests (new file: `internal/handlers/security_test.go`)
- SQL injection on all handlers accepting string params (`/api/stats?domain='; DROP TABLE--`)
- Path traversal on file endpoints (`/api/sites/{id}/files/../../etc/passwd`)
- XSS in user input fields (event tracking, webhook URLs, alias names)
- **Pattern**: Table-driven tests with known injection payloads, assert no 500s and no data leakage

#### 1b. Auth Bypass Tests (new file: `cmd/server/main_integration_security_test.go`)
- Invalid/expired/tampered session tokens
- Missing cookies with valid headers (and vice versa)
- Token replay after logout
- Brute force rate limiting on `/api/login`
- **Pattern**: Use existing `setupIntegrationTest()` infrastructure

#### 1c. Resource Exhaustion Tests (in `internal/handlers/security_test.go`)
- Oversized request bodies (exceeding BodySizeLimit middleware)
- Oversized JSON payloads
- Very long URL paths / query strings
- **Pattern**: Send large payloads, assert 413 or graceful rejection

#### 1d. SSRF Tests (in `internal/egress/`)
- Private IP blocking (127.0.0.1, 10.0.0.0/8, 192.168.0.0/16, 172.16.0.0/12)
- IPv6 bypass attempts (::1, IPv4-mapped IPv6)
- DNS rebinding scenarios
- **Pattern**: Tests already exist in `egress/`, extend coverage

### Priority 2: Data Flow Integration Tests

#### 2a. TestStorageAccessControl (`cmd/server/main_integration_test.go`)
- Create two users with sessions
- User A stores data, User B tries to access → rejected
- **Pattern**: Use existing `createSession()`, test storage API endpoints

### Nomenclature Cleanup (Future Session)

- Rename `site_id` → `app_id` in VFS/hosting code
  - Files table, hosting functions, deploy pipeline
  - Large refactoring: touches many files across codebase
  - Should be dedicated session with comprehensive testing

### Complex Integration Tests (Requires Architectural Understanding)

- **TestDeployToServing** - Multi-system flow: zip upload → extract → VFS → alias → serve
  - Touches: hosting, VFS, aliases, file system, routing
  - Requires understanding full deploy pipeline

- **TestAliasToAppResolution** - VFS serving architecture
  - How aliases map to apps, how site_id relates to subdomain
  - Overlaps with VFS/hosting architecture decisions

### Architectural Fixes

- **Nested query deadlock fixes** (Issue 05 pattern)
  - `buildLineageTree`, `AppsListHandlerV2`, `AppForksHandler`
  - All call `getAliasesForApp` inside rows iteration
  - Requires refactoring to collect rows first, then query

- **TestStressMessageThroughput stabilization**
  - WebSocket stress test with timing issues
  - May need architectural changes to message buffering

### Creative Security Work

- **Non-formulaic auth bypass scenarios**
  - Adversarial thinking: session fixation, CSRF, timing attacks
  - Race conditions in session validation
  - Edge cases in middleware ordering

### Future Enhancements (Noted in Tests)

- **Rate limiting** - No rate limiting on auth checks (commented test ready)
- **IP binding** - Sessions not bound to IPs (TestAuthBypass_SessionHijacking documents this)
- **Timestamp validation** - Future-dated sessions accepted (no created_at checks)

---

## Background Issues

### Flaky test: TestStressMessageThroughput
- In `internal/hosting` — WebSocket stress test
- Fails ~1/3 runs, passes 2/3 — timing-dependent
- Not blocking, but should be stabilized eventually

### Nested query deadlocks (Issue 05)
Multiple handlers have `db.Query()` inside rows iteration:
- `buildLineageTree` in apps_handler_v2.go
- `AppsListHandlerV2` in apps_handler_v2.go
- `AppForksHandler` in apps_handler_v2.go
- All call `getAliasesForApp` while iterating open cursor

---

## Quick Reference

```bash
# Full suite
go test ./... -count=1

# Handlers only
go test ./internal/handlers -count=1

# Coverage report
go test ./internal/handlers -coverprofile=coverage.out
go tool cover -func=coverage.out | grep "total:"

# Specific test groups
go test ./internal/handlers -run "TestTrackHandler" -v
go test ./internal/handlers -run "TestLogsHandler" -v
```

---

## Quick Reference - Integration Tests

```bash
# Run all integration tests
go test ./cmd/server -run "^Test.*" -v -count=1

# Run only auth tests
go test ./cmd/server -run "TestLogin|TestSession|TestRole|TestAuth" -v -count=1

# Run only routing tests
go test ./cmd/server -run "TestHost|TestSubdomain|TestLocalhost|TestRouting" -v -count=1

# Check which tests are failing
go test ./cmd/server -run "^Test.*" -count=1 2>&1 | grep "FAIL:"
```

## Quick Reference - Security Tests

```bash
# Run all security tests
go test ./internal/handlers -run "^TestSQLInjection|^TestPathTraversal|^TestXSS|^TestResourceExhaustion" -v -count=1
go test ./cmd/server -run "TestAuthBypass" -v -count=1
go test ./internal/egress -run "IPv4MappedIPv6|IPv6EdgeCases|IPv6Loopback|MetadataService|AlternativeSchemes" -v -count=1
go test ./cmd/server -run "TestStorageAccessControl" -v -count=1

# Run injection tests only
go test ./internal/handlers -run "Injection" -v -count=1

# Run auth bypass tests only
go test ./cmd/server -run "AuthBypass" -v -count=1

# Run SSRF tests (all)
go test ./internal/egress -v -count=1
```

## Key Learnings

### From Phase 2 (Handler Tests)
1. **Always run full test suite** — Individual tests passing ≠ suite passing
2. **Goroutine cleanup is critical** — Background goroutines need explicit Stop()
3. **Never nest queries with open cursors** — Collect rows first, close, then query more
4. **Test isolation matters** — Global state (`database.SetDB()`) can cause races
5. **SetMaxOpenConns(1) in tests is valuable** — Exposes connection lifecycle bugs that hide in production
6. **V1 handlers are stale** — Migration 012 broke v1 app handlers; they reference removed columns
7. **Nullable columns bite** — Always insert empty strings in test helpers, not NULL, for columns scanned into `string`

### From Phase 3 (Integration Tests)
8. **Session tokens require SHA-256 hashing** — Use `sha256.Sum256()` before storing in `auth_sessions.token_hash`
9. **User records need non-NULL fields** — Set `name` and `picture` to empty strings, not NULL (scan fails otherwise)
10. **Sessions need last_seen** — Insert with `last_seen` timestamp, not NULL
11. **Schema matters in tests** — Files table uses `mime_type`/`size_bytes`/`hash`, Aliases uses `subdomain`/`targets` JSON
12. **Targets JSON format** — `{"app_id":"..."}` for proxy/app aliases, `{"redirect_url":"..."}` for redirects
13. **Host-based routing differs** — `admin.*` requires role checks, `localhost` only requires auth
14. **Test helpers must match schema** — Helper functions broke when schema changed in migration 012

### From Phase 4 (Security Tests)
15. **URL-encode injection payloads** — Use `url.QueryEscape()` for query params, `url.PathEscape()` for path segments in test URLs
16. **TrackHandler returns 204** — Not 200; analytics buffered via `analytics.Add()`, not immediate DB writes
17. **Foreign key constraints prevent orphan sessions** — Cannot create session without valid user (good security property)
18. **Storage uses double isolation** — Both user_id column AND key prefix (`u:{userID}:{key}`)
19. **DNS resolution protects against rebinding** — `DialContext` checks EVERY resolved IP before connecting
20. **IPv4-mapped IPv6 blocked** — `::ffff:127.0.0.1` properly detected as loopback and blocked
