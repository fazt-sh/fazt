# Fazt Implementation State

**Last Updated**: 2026-01-17
**Current Version**: v0.9.23 (local), v0.9.23 (zyt)

## Status

```
State: CLEAN
All work committed. Plan 19 drafted, ready for implementation.
```

---

## Next Up: Plan 19 - Vite Dev Enhancement

**Plan**: `koder/plans/19_vite-dev-enhancement.md`

Unified build/deploy model with progressive enhancement:

- `fazt app create --template vite` scaffolds Vite-ready apps
- Embedded templates (minimal, vite) in binary
- Multi-package-manager support (bun, pnpm, yarn, npm)
- Build step integrated into deploy (or graceful fallback)
- Pre-built branch detection for git installs
- API endpoints for LLM harness integration

**Key constraint**: Complex apps that require building MUST have either:
- A package manager available, OR
- A pre-built dist/ folder, OR
- A pre-built branch (fazt-dist)

Otherwise → clear error (not broken deployment).

---

## Recent Changes (Plan 18)

App Ecosystem implementation completed:

- `fazt app` CLI namespace (list, deploy, install, upgrade, pull, info, remove)
- Git integration via go-git for `fazt app install`
- Source tracking in DB for upgrade detection
- `/fazt-app` skill for Claude-driven app development
- API endpoints for source info and file content

---

## Quick Reference

| Command | Purpose |
|---------|---------|
| `fazt app list zyt` | List apps |
| `fazt app deploy ./dir --to zyt` | Deploy from local |
| `fazt app install github:user/repo` | Install from GitHub |
| `fazt app upgrade myapp` | Upgrade git-sourced app |
| `fazt remote status zyt` | Check health/version |

---

## Apps on zyt.app

| App | URL |
|-----|-----|
| home | https://zyt.app |
| tetris | https://tetris.zyt.app |
| snake | https://snake.zyt.app |
