# Fazt Implementation State

**Last Updated**: 2026-01-17
**Current Version**: v0.9.23 (Plan 18 implemented, not yet released)

## Status

```
State: READY
Plan 18 complete. Uncommitted changes ready for review/commit.
```

---

## Uncommitted Work

Plan 18 (App Ecosystem) is fully implemented but not committed:

- `fazt app` CLI namespace (list, deploy, install, upgrade, pull, info, remove)
- Git integration via go-git for `fazt app install`
- Source tracking in DB for upgrade detection
- `/fazt-app` skill for Claude-driven app development
- API endpoints for source info and file content

Run `git status` to see changed files. Ready for commit or testing.

---

## Next: Vite Dev Enhancement

**Idea**: Optional Vite integration for better DX during app development.

### Goals

- Error catching during development
- HMR (Hot Module Replacement) for faster iteration
- Optional build step for performance (tree-shaking, minification)
- Apps must still work WITHOUT Vite (zero-build fallback)

### Approach

- If npm available locally, use Vite transparently
- Only adds `vite.config.js` to app folder
- If npm unavailable, config file sits inert
- Graceful degradation: always works without build tools

### Considerations

- HMR needs API proxy for serverless endpoints
- VM IP (192.168.64.3) - can we avoid hardcoding?
- Keep it elegant: detect environment, adapt automatically

### Questions to Resolve

1. How to detect npm availability and use Vite conditionally?
2. Proxy config for `/api/*` routes to fazt server?
3. Environment variable or auto-detect for VM IP?
4. vite.config.js template that works for fazt apps?

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
