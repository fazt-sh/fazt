# Fazt Implementation State

**Last Updated**: 2026-01-27
**Current Version**: v0.11.0

## Status

State: RELEASED - Auth primitive deployed to zyt.app, ready for testing

---

## Last Session

**Auth Primitive Implementation (Plan 23) + fazt-app Skill Update**

1. **Implemented Complete Auth System** (`internal/auth/`)
   - `service.go` - Main AuthService orchestrator
   - `users.go` - User CRUD operations with roles (owner/admin/user)
   - `sessions_db.go` - SQLite-backed session management
   - `providers.go` - OAuth configs (Google, GitHub, Discord, Microsoft)
   - `oauth.go` - State management, CSRF protection
   - `handlers.go` - HTTP handlers for login, callback, logout
   - `invites.go` - Invite code system for password users
   - `service_test.go` - Unit tests (all passing)

2. **Database Migration**
   - `internal/database/migrations/015_auth.sql` - Schema for auth tables

3. **Runtime Integration**
   - `internal/runtime/auth_bindings.go` - JS API (`fazt.auth.*`)
   - Modified `handler.go` - AuthContext injection

4. **CLI Commands**
   - `cmd/server/auth.go` - `fazt auth provider|users|invite` commands
   - Modified `main.go` - Auth service init, routes registration

5. **Updated fazt-app Skill**
   - Added authentication patterns documentation
   - Vue `useAuth` composable pattern
   - Protected API and page patterns
   - Auth + Sessions combined pattern
   - Updated instructions for auth decisions

## Files Created

```
internal/auth/
  service.go         # Main auth service
  users.go           # User CRUD
  sessions_db.go     # SQLite session management
  providers.go       # OAuth provider configs
  oauth.go           # OAuth flow (state, callback)
  handlers.go        # HTTP handlers
  invites.go         # Invite code system
  bindings.go        # JS API bindings (unused - runtime one is used)
  service_test.go    # Tests

internal/runtime/
  auth_bindings.go   # Runtime auth injection

internal/database/migrations/
  015_auth.sql       # Database schema

cmd/server/
  auth.go            # CLI commands
```

## Files Modified

- `cmd/server/main.go` - Auth service init, routes, CLI command
- `internal/runtime/handler.go` - AuthContext, auth injector
- `.claude/skills/fazt-app/SKILL.md` - Auth patterns documentation

---

## What Works

- [x] Database schema and migrations
- [x] User creation (first user = owner)
- [x] OAuth provider configuration via CLI
- [x] OAuth state management (CSRF protection)
- [x] Session creation and validation
- [x] Domain-wide SSO cookies
- [x] JS runtime bindings (`fazt.auth.*`)
- [x] Invite code generation and redemption
- [x] Password user signup via invites
- [x] Login/logout HTTP handlers
- [x] CLI commands for user/invite management
- [x] All unit tests passing

---

## Next Up

1. **Configure OAuth on zyt.app** - Set up Google OAuth provider
2. **Test auth flow** - Visit https://zyt.app/auth/login
3. **Build test app** using `/fazt-app` with auth patterns

---

## Testing Notes

**Cannot be tested automatically (requires real OAuth credentials):**
- OAuth flow with real providers
- Subdomain SSO (requires real domain)
- Cookie behavior in browser

**To test on zyt.app:**
```bash
# 1. Configure a provider
fazt auth provider google \
  --client-id "xxx.apps.googleusercontent.com" \
  --client-secret "GOCSPX-xxx" \
  --db servers/zyt/data.db
fazt auth provider google --enable --db servers/zyt/data.db

# 2. Visit https://zyt.app/auth/login

# 3. Test JS API in an app
```

---

## Quick Reference

```bash
# Configure provider
fazt auth provider google --client-id XXX --client-secret YYY
fazt auth provider google --enable

# List users
fazt auth users

# Create invite
fazt auth invite --role user

# View invites
fazt auth invites
```
