# Issue: Embedded Admin Dashboard Auth Mismatch

**Date:** 2026-02-03
**Status:** RESOLVED
**Priority:** High
**Resolution:** 2026-02-04

## Problem

After successful dev login, the embedded React admin dashboard (served at `admin.192.168.64.3.nip.io:8080`) redirects to `/#/login` instead of showing the authenticated dashboard.

## Root Cause Analysis

There are **two separate authentication systems** that aren't properly integrated:

### 1. Backend Session Systems

| System | Storage | Cookie | Used By |
|--------|---------|--------|---------|
| `SessionStore` | In-memory map | `fazt_session` | Simple password login (`/api/login`) |
| `auth.Service` | `auth_sessions` table | `fazt_session` | OAuth/Dev login (`/auth/dev/callback`) |

Both use the same cookie name `fazt_session`, but validate against different stores.

### 2. Auth Middleware

The `AuthMiddleware` in `internal/middleware/auth.go` was updated to check both:
```go
// 3. Try in-memory session store first (simple password login)
valid, err := sessionStore.ValidateSession(sessionID)
if err == nil && valid {
    next.ServeHTTP(w, r)
    return
}

// 4. Try database-backed sessions (OAuth/dev login)
if dbAuthService != nil {
    user, err := dbAuthService.ValidateSession(sessionID)
    if err == nil && user != nil {
        next.ServeHTTP(w, r)
        return
    }
}
```

This works - API calls succeed after dev login (verified with curl).

### 3. The Real Problem: Embedded React Admin

The embedded React admin at `/internal/assets/system/admin/` is a **bundled React SPA** that has its own client-side auth logic:

1. On load, it calls some auth endpoint to check if user is logged in
2. If not authenticated (by its own check), it redirects to `/#/login`
3. This happens **before** any API call that would be protected by our middleware

The React admin likely:
- Calls `/auth/session` or similar endpoint
- Expects a specific response format
- Uses its own auth context/state management

## Evidence

1. **Dev login works** - Creates user in `auth_users`, creates session in `auth_sessions`, sets cookie
2. **API access works** - After dev login, `curl -b cookies /api/aliases` returns data (25 items)
3. **Middleware works** - Updated to check both session stores
4. **But React admin still redirects** - Client-side JS redirects to `/#/login` before API calls

## Key Files

- `/internal/middleware/auth.go` - Backend auth middleware (updated)
- `/internal/auth/session.go` - In-memory SessionStore
- `/internal/auth/sessions_db.go` - Database-backed sessions (Service.CreateSession/ValidateSession)
- `/internal/auth/handlers.go` - OAuth/dev login handlers
- `/internal/handlers/auth_handlers.go` - Simple password login handler
- `/internal/assets/system/admin/` - Embedded React admin (bundled, no source)

## What We Don't Know

1. **What endpoint does the React admin call to check auth?**
   - Possibly `/auth/session`?
   - Possibly a custom endpoint?

2. **What response format does it expect?**
   - The `/auth/session` endpoint returns user data for database sessions
   - Does the React admin call a different endpoint?

3. **Where is the React admin source code?**
   - Only bundled JS exists in `/internal/assets/system/admin/`
   - Source may be in a separate repo or was compiled elsewhere

## Potential Solutions

### Option A: Find React Admin's Auth Check
- Reverse-engineer the bundled JS to find what endpoint it calls
- Ensure that endpoint works with database-backed sessions

### Option B: Unify Session Systems
- Make simple password login also create database sessions
- Remove in-memory SessionStore entirely
- Single source of truth for all sessions

### Option C: Add Compatibility Endpoint
- If React admin calls `/api/user/me` or similar
- Make that endpoint also check database sessions

### Option D: Replace Embedded Admin
- The `admin-ui` we're building could replace the embedded admin
- Deploy it to the `admin` alias instead of `admin-ui`
- But `admin` is reserved...

### Option E: Skip Embedded Admin
- Use `admin-ui` subdomain for the new admin UI
- Requires cross-subdomain cookie (already working)
- Users login at `admin.*` then use `admin-ui.*`

## Verification Commands

```bash
# 1. Dev login creates session
curl -c /tmp/cookies.txt -X POST http://admin.192.168.64.3.nip.io:8080/auth/dev/callback \
  -d "email=test@example.com&name=Test&role=admin&redirect=/"

# 2. Check session endpoint
curl -b /tmp/cookies.txt http://admin.192.168.64.3.nip.io:8080/auth/session

# 3. API access works
curl -b /tmp/cookies.txt http://admin.192.168.64.3.nip.io:8080/api/aliases

# 4. Check what React admin might call
curl -b /tmp/cookies.txt http://admin.192.168.64.3.nip.io:8080/api/user/me
```

## Related Work Done

1. Unified cookie name to `fazt_session`
2. Added cross-subdomain cookie support (domain `.192.168.64.3.nip.io`)
3. Added `POST /auth/login` for embedded admin's simple login
4. Updated middleware to check both session stores
5. Updated `admin-ui` to make cross-origin API calls

## Resolution

**Root cause confirmed:** The embedded React admin calls `/api/user/me` to check authentication. This endpoint (`UserMeHandler` in `internal/handlers/auth_handlers.go`) only checked the in-memory `sessionStore`, but dev/OAuth logins create sessions in the database via `auth.Service`.

**Final fix: Unified session system**

Removed the dual session store architecture entirely. All sessions now use the database-backed `auth.Service`:

1. **Removed in-memory SessionStore usage** - No more dual-store confusion
2. **Updated `InitAuth()`** - Now only takes `*auth.Service` and `*RateLimiter`
3. **Updated `LoginHandler`** - Password login now creates database sessions via `GetOrCreateLocalAdmin()`
4. **Updated `AuthMiddleware`** - Only checks database sessions
5. **Updated all handlers** - Use `authService.GetSessionFromRequest()`

**Files changed:**
- `internal/handlers/auth_handlers.go` - Unified on auth.Service
- `internal/middleware/auth.go` - Simplified to use only auth.Service
- `cmd/server/main.go` - Removed sessionStore, updated init
- `internal/auth/users.go` - Added `GetOrCreateLocalAdmin()` for password login
- `internal/handlers/auth_test.go` - Rewritten to use auth.Service with test DB

**Benefits:**
- Single source of truth for sessions (database)
- Sessions persist across server restarts
- Cleaner code, no dual-store checks
- Password login and OAuth/dev login use the same session system

**Verification:**
```bash
# Dev login + /api/user/me now works
curl -c cookies -X POST ".../auth/dev/callback" -d "email=test@example.com..."
curl -b cookies ".../api/user/me"
# Returns: {"data":{"email":"test@example.com","role":"admin","username":"Test","version":"0.22.0"}}

# /auth/session also works
curl -b cookies ".../auth/session"
# Returns: {"data":{"authenticated":true,"user":{...}}}
```

**Legacy code:** `internal/auth/session.go` and `session_test.go` are now unused but kept for reference. Can be removed in future cleanup.

## Follow-up: CSP Violation (2026-02-04)

After the unified session system was implemented, browser testing revealed a **Content Security Policy (CSP) violation** that prevented the admin-ui from making API calls.

**Problem:** The admin-ui Vue app (at `admin-ui.192.168.64.3.nip.io:8080`) was blocked from calling APIs on `admin.192.168.64.3.nip.io:8080` because the CSP only allowed `https://*.192.168.64.3.nip.io`, but we were using HTTP on port 8080.

**Console errors:**
```
Connecting to 'http://admin.192.168.64.3.nip.io:8080/auth/session' violates the following
Content Security Policy directive: "connect-src 'self' ... https://*.192.168.64.3.nip.io"
```

**Root cause:** CSP wildcard patterns don't match ports. `https://*.192.168.64.3.nip.io` means port 443, not 8080.

**Fix:** Updated `internal/middleware/security.go`:
1. Modified `buildCSP()` to accept `port` and `isSecure` parameters
2. In development (non-HTTPS), CSP now includes `http://*.192.168.64.3.nip.io:8080`
3. Updated `SecurityHeaders` middleware to pass port and HTTPS status to `buildCSP()`

**Result:**
- ✅ CSP now allows cross-subdomain API calls in development
- ✅ Admin-ui successfully loads auth state and displays dashboard
- ✅ All API endpoints accessible (apps, aliases, events, logs, etc.)
- ✅ Full authentication flow works: login → redirect → authenticated dashboard

**Verification:**
```bash
# CSP header now includes HTTP with port
curl -I http://admin-ui.192.168.64.3.nip.io:8080/
# Returns: connect-src 'self' ... https://*.192.168.64.3.nip.io http://*.192.168.64.3.nip.io:8080

# Login with owner role required for admin-ui
# (admin-ui checks: user.role === 'owner' || user.role === 'admin')

# Dashboard loads successfully with system metrics, navigation, and full functionality
```

**Status:** ✅ **FULLY RESOLVED** - Authentication flow works end-to-end in browser.
