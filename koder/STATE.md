# Fazt Implementation State

**Last Updated**: 2026-02-07
**Current Version**: v0.28.0

## Status

State: CLEAN
Released v0.28.0 with `fazt @peer app list` fix, API key auth, and integration tests. All peers upgraded.

---

## Last Session (2026-02-07) — Bug Fix + Security + Tests

### What Was Done

1. **Tracked down `fazt @local app list` bug**
   - Symptom: Empty "Error:" message when running `fazt @local app list`
   - Root cause: `/api/cmd` endpoint caught by `AdminMiddleware` (requires session auth), but CLI sends Bearer token (API key auth)
   - Fix: Added `/api/cmd` to auth bypass list in `main.go:1524`

2. **Added API key authentication to CmdGatewayHandler**
   - Endpoints that bypass `AdminMiddleware` must validate API keys themselves
   - Followed pattern from `DeployHandler` (Bearer token validation)
   - Added `database.SetDB()` for test injection

3. **Wrote 6 integration tests** (`internal/handlers/cmd_gateway_test.go`)
   - `TestCmdGateway_RejectsUnauthenticated` — No auth → 401
   - `TestCmdGateway_RejectsInvalidToken` — Bad token → 401
   - `TestCmdGateway_AcceptsValidAPIKey` — Valid key → 200
   - `TestCmdGateway_RejectsInvalidMethod` — GET → 405
   - `TestCmdGateway_AppListReturnsApps` — Full execution
   - `TestCmdGateway_UnknownCommand` — Error handling
   - All tests passing ✅

4. **Released v0.28.0**
   - Bug fix + security fix + tests
   - SDK evolution (Plan 43) + Admin rebuild (Plan 45) from previous session
   - Upgraded: local v0.28.0, zyt v0.28.0

### Commits

```
f8319f2 add API key auth to /api/cmd + integration tests
ec5f5a6 release: v0.28.0
7a18dbe fix: add /api/cmd to auth bypass list
```

### Why Tests Didn't Catch This Initially

This was an **integration/routing bug** - not a unit test bug:
- Unit tests for `CmdGatewayHandler` would pass ✓
- Unit tests for CLI would pass ✓
- Unit tests for middleware would pass ✓
- Bug was in **routing configuration** (how components connect)
- Requires end-to-end tests: real server + real HTTP requests + different auth methods

---

## Next Session

### Priority

1. **Test admin against real server** — Verify auth, apps, aliases work after rebuild
2. **Test admin with `?mock=true`** — All pages load, data renders correctly
3. **Deploy admin to local + zyt** — `fazt @local app deploy ./admin`

### Direction

- **Migrate Preview app** — Use `createAppClient()` instead of hand-rolled `api.js`
- **Document media APIs in KB** — `fazt.app.media.{probe,transcode,serve,resize}`
- **Plan 44: Drop app** — File/folder hosting via fazt (idea stage)

---

## Quick Reference

```bash
# Test admin build
cd admin && npm run build

# Test all Go
go test ./... -short -count=1

# Deploy admin
fazt @local app deploy ./admin
fazt @zyt app deploy ./admin

# Test mock mode
# Open admin URL with ?mock=true

# Peer status
fazt peer list
fazt @local status
fazt @zyt status
```
