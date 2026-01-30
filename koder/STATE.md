# Fazt Implementation State

**Last Updated**: 2026-01-31
**Current Version**: v0.16.0

## Status

State: **CLEAN** - Released v0.16.0

---

## Last Session (2026-01-31)

**Released v0.16.0: User Data Foundation**

1. **New fazt ID format** (`internal/appid/appid.go`):
   - Format: `fazt_<type>_<12 base62 chars>` (e.g., `fazt_usr_Nf4rFeUfNV2H`)
   - Types: `usr`, `app`, `tok`, `ses`, `inv`

2. **User-scoped storage** (`internal/storage/scoped.go`):
   - `UserScopedKV`, `UserScopedDocs`, `UserScopedBlobs`
   - Automatic isolation via key/collection/path prefixing

3. **New `fazt.app.*` namespace** (`internal/storage/app_bindings.go`):
   - `fazt.app.ds/kv/s3.*` - shared app storage
   - `fazt.app.user.ds/kv/s3.*` - user's private storage

4. **Migration 017** - adds `user_id` column to storage tables

5. **Fixed runtime race condition** (`internal/runtime/runtime.go`):
   - Added goroutine exit synchronization before returning VM to pool

---

## Testing Notes

**Tested:**
- `fazt.app.kv.set/get` (shared) - works
- `fazt.app.user.*` without login - correctly errors
- Migration 017 - applied
- All unit tests pass

**Needs testing** (requires logged-in user):
- `fazt.app.user.*` with authenticated user
- User isolation between users

Test app: `test-user-storage` on local

---

## Next Up

1. **Plan 30c: Access Control** (v0.17.0)
   - RBAC with hierarchical roles
   - Email domain gating

2. **Plan 24: Mock OAuth** - enables local auth testing

---

## LEGACY_CODE Markers

```bash
grep -rn "LEGACY_CODE" internal/
```

- `internal/storage/bindings.go` - `fazt.storage.*` namespace
- `internal/appid/appid.go` - old `app_*` format
- `internal/auth/service.go` - `generateUUID()`

See `koder/LEGACY.md` for removal guide.
