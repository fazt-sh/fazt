# Fazt Implementation State

**Last Updated**: 2026-01-28
**Current Version**: v0.11.5

## Status

State: RELEASED - Auth working end-to-end, guestbook app functional

---

## Last Session

**Guestbook App + Auth Fixes (v0.11.1 - v0.11.5)**

1. **Remote Auth Commands** (v0.11.1)
   - `fazt @peer auth provider <name>` - configure providers via API
   - New API endpoints: `GET/PUT /api/auth/providers/{name}`

2. **Auth Migration Fix** (v0.11.2)
   - Migration 015_auth.sql wasn't in migrations array

3. **Auth Routes Public** (v0.11.3)
   - Added `/auth/` to public paths in middleware

4. **Auth on All Subdomains** (v0.11.4)
   - Auth routes work on any subdomain, not just admin
   - OAuth users always get 'user' role (owner separate)

5. **OAuth Subdomain Fix** (v0.11.5)
   - Callback URLs always use root domain (zyt.app)
   - Fixes redirect_uri_mismatch when logging in from subdomains

6. **Guestbook App** (`servers/zyt/guestbook/`)
   - Full auth-protected guestbook with Google sign-in
   - Uses `fazt.storage.ds.*` for message storage
   - Avatar-based auth UI with popup/dropdown
   - Messages only visible when signed in

## Key Learnings

- Serverless entry point must be `api/main.js` (not per-route files)
- Must call `handler(request)` at end of main.js (runtime doesn't auto-invoke)
- Storage API: `fazt.storage.ds.find()` / `fazt.storage.ds.insert()` (not query/put)
- `fazt.uuid()` doesn't exist - let ds.insert auto-generate IDs
- OAuth callbacks must use root domain for all subdomains

---

## Next Explorations

### Guestbook Improvements

- [ ] Remove "Welcome Sign in to leave messages" - just show sign-in button
- [ ] Pagination - show 20 messages per page
- [ ] FAB + modal for adding entries (better UX than inline form)
- [ ] Inner page version - guestbook as route on zyt.app vs subdomain

### Auth DX Improvements

- [ ] **Runtime helpers**: Fold common auth patterns into JS runtime
  - Auto-redirect to login?
  - `fazt.auth.requireUser()` that handles the flow?
  - Session refresh helpers?

- [ ] **Headless components**: Reusable patterns for auth flows
  - Sign-in button component
  - Auth state provider
  - Protected route wrapper

- [ ] **Update /fazt-app skill**: Capture patterns from guestbook
  - Document the working flow
  - Storage API usage patterns
  - Auth integration patterns
  - Common pitfalls (main.js, handler call, etc.)

---

## What Works

- [x] OAuth with Google (configured on zyt.app)
- [x] Domain-wide SSO cookies (`.zyt.app`)
- [x] Auth from any subdomain (redirects properly)
- [x] JS runtime bindings (`fazt.auth.getUser()`)
- [x] Document storage (`fazt.storage.ds.*`)
- [x] Guestbook app end-to-end

---

## Quick Reference

```bash
# Configure OAuth provider remotely
fazt @zyt auth provider google --client-id XXX --client-secret YYY
fazt @zyt auth provider google --enable

# Deploy app
fazt app deploy servers/zyt/guestbook --to zyt

# Check server status
fazt remote status zyt

# Upgrade server
fazt remote upgrade zyt
```
