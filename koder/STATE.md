# Fazt Implementation State

**Last Updated**: 2026-01-31
**Current Version**: v0.16.0

## Status

State: **IN_PROGRESS** - Plan 24 implementation complete, uncommitted

---

## Last Session (2026-01-31)

**Implemented Plan 24: Mock OAuth Provider**

1. **Dev login routes** (`internal/auth/dev_provider.go`):
   - `GET /auth/dev/login` - dev login form
   - `POST /auth/dev/callback` - process dev login
   - `IsLocalMode(r)` - detects local/HTTP mode

2. **Login page update** (`internal/auth/handlers.go`):
   - Shows "Dev Login" button when in local mode
   - Separated "Development" section with divider

3. **Full session creation**:
   - Creates real user with `provider: "dev"`
   - Creates valid session cookie
   - Supports role selection (user/admin/owner)

4. **Verified user storage works**:
   - `fazt.app.user.kv.*` works with dev-authenticated user
   - Test app: `test-user-storage` on local

---

## Testing Notes

**Tested:**
- Dev login form renders at `/auth/dev/login`
- Dev login blocked on HTTPS (403 Forbidden)
- Session cookie set correctly
- `fazt.auth.getUser()` returns dev user
- `fazt.app.user.kv.set/get` works with authenticated user
- All unit tests pass (including new dev provider tests)

**Not tested:**
- User isolation between different dev users (needs second session)

---

## Next Up

1. **Commit Plan 24 changes**
   - Review and commit dev provider implementation
   - Consider: Should this be v0.16.1 or v0.17.0?

2. **Plan 30c: Access Control** (v0.17.0)
   - RBAC with hierarchical roles
   - Email domain gating

---

## LEGACY_CODE Markers

```bash
grep -rn "LEGACY_CODE" internal/
```

- `internal/storage/bindings.go` - `fazt.storage.*` namespace
- `internal/appid/appid.go` - old `app_*` format
- `internal/auth/service.go` - `generateUUID()`

See `koder/LEGACY.md` for removal guide.
