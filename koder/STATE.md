# Fazt Implementation State

**Last Updated**: 2026-01-17
**Current Version**: v0.9.24 (local), pending release to zyt

## Status

```
State: IN_PROGRESS
Plan 19 (Vite Dev Enhancement) implemented, pending release.
```

---

## Just Completed: Plan 19 - Vite Dev Enhancement

**Plan**: `koder/plans/19_vite-dev-enhancement.md`

Implemented unified build/deploy model with:

### Features Added

1. **Embedded Templates** (`internal/assets/templates/`)
   - `minimal` - Basic HTML app with Tailwind
   - `vite` - Full Vite setup with HMR, serverless API

2. **CLI Create Command** (`fazt app create`)
   - `fazt app create myapp` - Uses minimal template
   - `fazt app create myapp --template vite` - Uses vite template
   - `fazt app create --list-templates` - Shows available templates

3. **Build Package** (`internal/build/`)
   - Package manager detection (bun, pnpm, yarn, npm)
   - Lockfile-aware selection
   - Build execution with install step
   - Fallback to existing dist/ or source

4. **Updated Deploy** (uses build package)
   - Auto-builds when package.json has build script
   - `--no-build` flag to skip
   - Clear error messages when build required but impossible

5. **Pre-built Branch Detection**
   - Checks: fazt-dist, dist, release, gh-pages
   - Falls back automatically during `fazt app install`

6. **API Endpoints**
   - `POST /api/apps/install` - Install from GitHub
   - `POST /api/apps/create` - Create from template
   - `GET /api/templates` - List templates

### Files Created/Modified

| File | Status |
|------|--------|
| `internal/assets/templates/minimal/*` | Created |
| `internal/assets/templates/vite/*` | Created |
| `internal/assets/templates.go` | Created |
| `internal/assets/templates_test.go` | Created |
| `internal/build/build.go` | Created |
| `internal/build/pkgmgr.go` | Created |
| `internal/build/build_test.go` | Created |
| `cmd/server/app.go` | Modified (added create, build integration) |
| `cmd/server/app_create.go` | Created |
| `internal/git/git.go` | Modified (FindPrebuiltBranch) |
| `internal/handlers/apps_handler.go` | Modified (install, create endpoints) |
| `cmd/server/main.go` | Modified (new routes) |

---

## Quick Reference

| Command | Purpose |
|---------|---------|
| `fazt app create myapp` | Create minimal app |
| `fazt app create myapp --template vite` | Create Vite app |
| `fazt app list zyt` | List apps |
| `fazt app deploy ./dir --to zyt` | Deploy (with auto-build) |
| `fazt app deploy ./dir --to zyt --no-build` | Deploy without build |
| `fazt app install github:user/repo` | Install from GitHub |
| `fazt remote status zyt` | Check health/version |

---

## Apps on zyt.app

| App | URL |
|-----|-----|
| home | https://zyt.app |
| tetris | https://tetris.zyt.app |
| snake | https://snake.zyt.app |
