# Fazt Implementation State

**Last Updated**: 2026-02-07
**Current Version**: v0.27.0

## Status

State: CLEAN
Plan 43 (fazt-sdk evolution) implemented.

---

## Last Session (2026-02-07) — Plan 43: fazt-sdk Universal Client

### What Changed

Evolved fazt-sdk from admin-only to universal client serving both Admin UI and fazt-apps.

#### SDK Restructured
- **Moved** `admin/packages/fazt-sdk/` → `packages/fazt-sdk/` (repo root)
- **`errors.js`** (NEW) — `ApiError` class extending `Error` with status helpers
  (`.isAuth`, `.isForbidden`, `.isNotFound`, `.isRateLimit`, `.isServer`)
- **`admin.js`** (NEW) — Extracted admin namespace from old `index.js`
- **`app.js`** (NEW) — App namespace: `auth.me()`, `auth.login()`, `auth.logout()`,
  `http` (direct access), `upload()` (file + progress), `paginate()` (offset-based)
- **`client.js`** — FormData-aware `request()`, XHR `upload()` with progress tracking,
  `ApiError` instead of plain objects
- **`index.js`** — Thin re-export: `createClient()` (admin), `createAppClient()` (apps)
- **`mock.js`** — Added `GET /api/me` route for app auth
- **`types.js`** — Added `UploadOptions`, `PaginateOptions`, `PaginatedResult`, `AppUser`

#### Admin Import Updated
- **`admin/src/client.js`** — Import path updated to `../../packages/fazt-sdk/index.js`

#### Backward Compatible
- `createClient()` returns identical shape — all admin stores work unchanged
- `ApiError` extends `Error` so `.message` and `.code` access patterns compatible

### Key Files
```bash
packages/fazt-sdk/index.js    # Entry point
packages/fazt-sdk/client.js   # HTTP client + upload
packages/fazt-sdk/errors.js   # ApiError class
packages/fazt-sdk/admin.js    # Admin namespace
packages/fazt-sdk/app.js      # App namespace
packages/fazt-sdk/mock.js     # Mock adapter
packages/fazt-sdk/types.js    # JSDoc types
admin/src/client.js            # Updated import path
```

---

## Next Session

### Priority
- **Verify admin build** — `cd admin && npm run build` + test with `?mock=true`
- **Test against real server** — Auth, apps, aliases all work
- **Migrate Preview app** — Use `createAppClient()` instead of hand-rolled `api.js`
- **Document media APIs in KB** — `fazt.app.media.{probe,transcode,serve,resize}`

### Known Issues
- **`fazt @local app list`** — Returns empty error (pre-existing bug)

---

## Quick Reference

```bash
# Test admin build
cd admin && npm run build

# Test all Go
go test ./... -short -count=1

# Deploy preview
fazt @local app deploy ./servers/local/preview
```
