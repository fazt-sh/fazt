# Fazt Implementation State

**Last Updated**: 2026-02-08
**Current Version**: v0.28.0

## Status

State: READY — Plan 47 queued
Working on: —
Next: Plan 47 Codebase Cleanup (`koder/plans/47_codebase_cleanup.md`) — Sonnet, 4 phases: nomenclature, dead code removal, legacy migration, security hardening

---

## Current Session (2026-02-08) — Plan 46: Phase 7 Adversarial Security Tests

### Context

Plan 46 Phases 1-6 complete. Creative security work identified in STATE.md — adversarial thinking that simulates real attacker behavior rather than formulaic checklist testing.

### Completed — Phase 7: Adversarial Security Tests ✅

**File**: `cmd/server/main_integration_adversarial_test.go` — 12 test functions, all passing.

#### Vulnerabilities Discovered & Documented

1. **Invite Code TOCTOU** — `RedeemInvite` has no transaction: GetInvite → IsValid → CreateUser → UPDATE. 10 goroutines all redeemed a single-use invite. Fix: wrap in transaction.
2. **OAuth State TOCTOU** — `ValidateState` does SELECT → DELETE as separate statements. Multiple goroutines validated same token. Fix: DELETE...RETURNING or transaction.
3. **Login Timing Side Channel** — Invalid username returns ~260µs, valid username + wrong password ~800µs (bcrypt). 3x difference enables username enumeration.
4. **Rate Limit IP Spoofing** — `getClientIP()` trusts X-Forwarded-For unconditionally. 20/20 spoofed requests bypassed rate limiting.

#### Security Properties Verified

5. **Cookie Fixation → DoS only** — Attacker's cookie shadows victim's (Go returns first match), but cannot escalate privileges
6. **Cookie Attributes** — HttpOnly, SameSite=Lax, Path=/, positive MaxAge all verified
7. **No 500s Under Concurrent Invalidation** — 5 workers × 20 requests during session deletion: zero 500s, clean 200→401 transition
8. **Host Header Bypass Blocked** — Uppercase, suffix attack, unicode confusables, null bytes all return 404
9. **Path Confusion Blocked** — Traversal, case variation, encoded slashes all blocked by Go's ServeMux
10. **FK CASCADE Works** — Deleting user removes sessions; HTTP returns 401
11. **Role Downgrade Immediate** — No stale cache; demotion to "user" returns 403 on next request
12. **Token Entropy** — 50 tokens: zero collisions, 65 unique chars, 44-char length (32 bytes base64url)

---

## Previous Sessions

### Phase 6 (2026-02-08) — Complex Integration Tests ✅

**File**: `cmd/server/main_integration_deploy_test.go` — 18 sub-tests, all passing first run.

#### 1. TestDeployToServing (11 sub-tests)

Full deploy pipeline: ZIP → extract → VFS → alias creation → static file serving.

**DB verification:**
- App record created with correct `fazt_app_` ID prefix, title, source type
- Alias auto-created pointing subdomain → app_id
- 6 files stored in VFS with correct SHA-256 hashes

**HTTP serving verified:**
- index.html at root: Content-Type text/html, Cache-Control no-cache, ETag present
- CSS: text/css MIME, max-age=300 (5min cache)
- JS: application/javascript MIME
- Hashed assets (assets/main-abc123.js): immutable, max-age=31536000 (1yr)
- Directory index fallback: /about → about/index.html
- ETag → If-None-Match → 304 Not Modified
- Non-existent file → 404
- Redeploy: old files cleaned, new content served

#### 2. TestAliasToAppResolution (7 sub-tests)

Alias → app resolution through VFS serving architecture.

- **Site isolation**: Two subdomains (alpha, beta) serve completely isolated content; cross-subdomain file access → 404
- **Architecture invariant**: Files keyed by site_id (subdomain), NOT app_id
- **Trailing slash**: /about/ → 301 redirect to /about
- **SPA fallback**: Route-like paths (/dashboard, /settings/profile) → index.html when SPA enabled; real files still served directly; file extensions don't trigger SPA
- **Private files**: /private/ → 401 without auth, 200 with valid session
- **API paths**: /api/ without serverless → 404
- **Analytics injection**: sendBeacon script injected into HTML responses

### Full Test Suite: ALL GREEN

```bash
go test ./... -count=1  # All packages pass
```

---

### Phase 5 (2026-02-08) — Architectural Fixes ✅

Deadlock fixes in apps_handler_v2.go (collect-close-query pattern). Flaky WebSocket stress test stabilized (done+drain replaces idle timeouts). 4 unskipped tests.

### Phase 4 (2026-02-08) — Security Tests ✅

32 security tests: SQL injection, path traversal, XSS, resource exhaustion, auth bypass, SSRF, storage access control. All passing.

### Phase 3 (2026-02-08) — Integration Tests ✅

21 integration tests: auth flows (11), routing (10). Full test server with real routing/middleware.

### Phase 2 (2026-02-07) — Systematic Handler Coverage ✅

10 test files, ~160 functions. Handler coverage: 15.7% → 47.6%.

### Phase 1 — Critical Path Tests ✅

Initial test infrastructure. Handler coverage to 15.7%.

---

## Remaining Work

### Nomenclature Cleanup (Future Session)

- Rename `site_id` → `app_id` in VFS/hosting code
  - Files table, hosting functions, deploy pipeline
  - Large refactoring: touches many files across codebase
  - Should be dedicated session with comprehensive testing

### Future Enhancements (Noted in Tests)

- **Rate limiting** — No rate limiting on auth checks (commented test ready)
- **IP binding** — Sessions not bound to IPs (TestAuthBypass_SessionHijacking documents this)
- **Timestamp validation** — Future-dated sessions accepted (no created_at checks)
- **Invite TOCTOU** — RedeemInvite needs transaction to prevent double-spend
- **OAuth state TOCTOU** — ValidateState needs DELETE...RETURNING or transaction
- **Timing side channel** — LoginHandler leaks username validity via bcrypt timing
- **IP spoofing** — getClientIP() trusts X-Forwarded-For without validation

---

## Background Issues

### V1 handlers stale
- `AppsListHandler`, `AppDetailHandler` broken post-migration 012
- Query `a.name` and `a.manifest` which no longer exist
- Only method guards tested

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

## Quick Reference - Integration Tests

```bash
# Run all integration tests
go test ./cmd/server -run "^Test.*" -v -count=1

# Run only auth tests
go test ./cmd/server -run "TestLogin|TestSession|TestRole|TestAuth" -v -count=1

# Run only routing tests
go test ./cmd/server -run "TestHost|TestSubdomain|TestLocalhost|TestRouting" -v -count=1
```

## Quick Reference - Deploy & VFS Tests

```bash
# Run deploy pipeline + alias resolution tests
go test ./cmd/server -run "TestDeployToServing|TestAliasToAppResolution" -v -count=1
```

## Quick Reference - Security Tests

```bash
# Run all security tests
go test ./internal/handlers -run "^TestSQLInjection|^TestPathTraversal|^TestXSS|^TestResourceExhaustion" -v -count=1
go test ./cmd/server -run "TestAuthBypass" -v -count=1
go test ./internal/egress -run "IPv4MappedIPv6|IPv6EdgeCases|IPv6Loopback|MetadataService|AlternativeSchemes" -v -count=1
go test ./cmd/server -run "TestStorageAccessControl" -v -count=1

# Run adversarial security tests
go test ./cmd/server -run "TestAdversarial" -v -count=1
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

### From Phase 4 (Security Tests)
14. **URL-encode injection payloads** — Use `url.QueryEscape()` for query params, `url.PathEscape()` for path segments
15. **TrackHandler returns 204** — Not 200; analytics buffered via `analytics.Add()`, not immediate DB writes
16. **Foreign key constraints prevent orphan sessions** — Cannot create session without valid user
17. **Storage uses double isolation** — Both user_id column AND key prefix (`u:{userID}:{key}`)
18. **DNS resolution protects against rebinding** — `DialContext` checks EVERY resolved IP before connecting

### From Phase 5 (Architectural Fixes)
19. **Done+drain > idle timeouts for stress tests** — Idle timeouts are flaky under scheduler jitter; done signal + drain is deterministic
20. **Collect-close-query pattern** — Always collect rows, close cursor, then do additional queries. Never nest queries inside open cursors.

### From Phase 6 (Complex Integration Tests)
21. **App IDs use `fazt_app_` prefix** — Not `app_`; generated by `appid.GenerateApp()`
22. **EnsureApp auto-creates alias** — Deploy creates both app record AND alias; no manual alias creation needed
23. **site_id = subdomain, NOT app_id** — VFS files keyed by subdomain; app_id only for analytics/identity
24. **Analytics injection modifies HTML** — Use `assertContains` not exact match; sendBeacon script injected before `</body>`
25. **Deploy is atomic** — DeleteSite (clear old) → EnsureApp → WriteFile for each; cache invalidated on delete

### From Phase 7 (Adversarial Security Tests)
26. **admin.* blocks /api/login** — Login must go through localhost; admin subdomain routes all /api/ through AdminMiddleware
27. **database/sql releases conn between statements** — MaxOpenConns(1) doesn't prevent TOCTOU; each QueryRow/Exec acquires and releases independently
28. **Vulnerability-documenting tests use t.Log** — Tests that expose known weaknesses log findings, not fail; tests verifying correct behavior use t.Error
29. **bcrypt.MinCost for test speed** — Use bcrypt.MinCost (4) in test setup, not BcryptCost (12), to keep tests fast
