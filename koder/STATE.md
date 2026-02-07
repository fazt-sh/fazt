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

### Also Done
- **Bundle script** — `packages/fazt-sdk/build.sh` → `dist/fazt-sdk.mjs` (32KB, unminified)
- **Plan 44** — Drop app idea documented (`koder/plans/44_drop_app.md`)
  - `fazt upload ./file.zip` → `https://drop.zyt.app/<hash>/file.zip`
  - Folder support, content-type detection
  - Simple short hashes for now; CID-compatible URLs for when v0.9 storage lands
- Admin build verified (538ms, 70 modules, passes clean)

---

## Next Session

### Important

- work on plan 45 before anything

### Priority
- **Test admin against real server** — Auth, apps, aliases all work after SDK move
- **Test admin with `?mock=true`** — All pages load, data renders
- **Migrate Preview app** — Use `createAppClient()` instead of hand-rolled `api.js`
- **Document media APIs in KB** — `fazt.app.media.{probe,transcode,serve,resize}`

### Direction
- **Plan 44: Drop app** — File/folder hosting via fazt (idea stage)
- **SDK external consumption** — Relative imports for now; Drop will host the bundle later

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
