# Fazt Implementation State

**Last Updated**: 2026-01-31
**Monorepo Version**: v0.17.0 (unified versioning)

## Component Status

| Component | Status | Complete | Notes |
|-----------|--------|----------|-------|
| fazt-binary | stable | 100% | Core platform |
| admin | alpha | 15% | Dashboard complete, other pages skeleton |
| fazt-sdk | alpha | 20% | Basic API client |
| knowledge-base | stable | 80% | Comprehensive docs |

## Current State

State: **CLEAN** - Monorepo versioning implemented, admin UI with auth

---

## Last Session (2026-01-31)

**Monorepo Versioning + Admin UI Refinements**

Implemented unified versioning strategy and moved admin to tracked location.

1. **Auth integration** (leverages existing fazt auth system):
   - Connected to `/auth/session` and `/auth/logout` endpoints
   - User dropdown shows provider (Google, GitHub, Dev)
   - Role badge displays Owner/Admin status
   - Logged-out state redirects to `/auth/login`
   - Working sign out flow

2. **Footer improvements**:
   - Mock mode toggle as icon button (database icon)
   - Settings panel toggle as icon button (palette icon)
   - Both show active state (accent color + background)

3. **Empty states**:
   - Apps table shows "No apps yet" message
   - Call-to-action for first deployment

4. **SDK updates**:
   - Updated `fazt-sdk` to use correct auth endpoints
   - Mock adapter returns proper User structure with role/provider
   - Added `auth.session()` and `auth.signOut()` methods

4. **Monorepo versioning** (architectural decision):
   - Unified versioning: all components share v0.17.0
   - Status markers track maturity: stable, beta, alpha
   - Completeness % shows progress towards parity
   - Root `version.json` is source of truth
   - Moved admin from `servers/local/admin/` to `admin/` (now tracked)

5. **Documentation updates**:
   - Updated CLAUDE.md with monorepo structure
   - Updated /open skill for unified versioning workflow
   - Version files in root, admin/, knowledge-base/

**Key discovery**: Fazt already has built-in owner/admin roles! No new features needed - just wiring to existing `User.role`, `User.IsOwner()`, `User.IsAdmin()` methods.

**Deployed**: `http://admin-ui.192.168.64.3.nip.io:8080`

---

## Testing Auth

**Dev login** (local only):
```bash
# Visit dev login page
open http://admin.192.168.64.3.nip.io:8080/auth/dev/login

# Pick role: User, Admin, or Owner
# Admin UI will show role badge and provider
```

**Mock mode**: `?mock=true` shows mock user (kodeman@gmail.com, Owner, Google)

---

## Next Up

### Admin UI Pages (15% → Goal: 100%)

**Dashboard** ✅
- Complete with auth, stats, apps table, activity feed
- Empty states, theme system, command palette

**Apps Page** (Priority: High)
- List all apps with filtering/search
- App detail view (files, source, config)
- CRUD operations (create, deploy, delete)
- Real-time status updates
- Empty state for no apps

**Aliases Page** (Priority: High)
- List all aliases with subdomain management
- Create/edit/delete aliases
- Alias types (proxy, redirect, static)
- DNS configuration UI
- Empty state for no aliases

**System Page** (Priority: Medium)
- Health dashboard (CPU, memory, storage)
- Active workers/runtime stats
- Database metrics
- VFS cache stats
- Capacity monitoring

**Settings Page** (Priority: Medium)
- OAuth provider configuration
- Rate limits and quotas
- Domain settings
- User management (invite, roles)
- System preferences

### Additional Features

- App detail page with file browser
- Real-time updates (WebSocket/SSE)
- More command palette actions
- Admin-only UI elements
- Logs viewer
- Metrics/analytics

---

## Quick Reference

```bash
# Deploy admin UI
fazt app deploy admin --to local --name admin-ui

# Test with mock data
http://admin-ui.192.168.64.3.nip.io:8080?mock=true

# View source (BFBB - no build)
ls admin/packages/
ls admin/src/

# Check version/status
cat version.json
cat admin/version.json
```

---

## LEGACY_CODE Markers

```bash
grep -rn "LEGACY_CODE" internal/
```

- `internal/storage/bindings.go` - `fazt.storage.*` namespace
- `internal/appid/appid.go` - old `app_*` format
- `internal/auth/service.go` - `generateUUID()`

See `koder/LEGACY.md` for removal guide.
