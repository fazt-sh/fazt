# Fazt Implementation State

**Last Updated**: 2026-01-31
**Current Version**: v0.17.0

## Status

State: **CLEAN** - v0.17.0 released

---

## Last Session (2026-01-31)

**Released v0.17.0: Dev OAuth Provider**

1. **Implemented Plan 24**:
   - Dev login routes (`internal/auth/dev_provider.go`)
   - Login page integration with dev mode detection
   - Full session creation with role selection
   - Unit tests and initial manual testing

2. **Released v0.17.0**:
   - Committed dev provider implementation
   - Built and uploaded release assets
   - Upgraded both remotes (local, zyt)

---

## Testing Needed (v0.17.0)

**To verify on local server** (`http://admin.192.168.64.3.nip.io:8080`):

1. **Dev login flow**:
   - Visit `/auth/dev/login` directly
   - Test form submission with different roles
   - Verify session persistence across requests

2. **Login page integration**:
   - Check if "Dev Login" button appears in local mode
   - Verify visual separation (divider)

3. **Production blocking**:
   - Confirm dev login blocked on zyt (HTTPS)
   - Should return 403 Forbidden

4. **User isolation**:
   - Create two different dev users (different emails)
   - Verify storage isolation with `fazt.app.user.kv.*`

---

## Next Up

1. **Test v0.17.0 dev provider** (see above)

2. **Plan 30c: Access Control** (v0.18.0)
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
