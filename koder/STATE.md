# Fazt Implementation State

**Last Updated**: 2026-01-18
**Current Version**: v0.9.26 (local), v0.9.26 (zyt)

## Status

```
State: CLEAN
v0.9.26 released - Serverless timeout increased for larger KV ops.
```

---

## Recent Changes (v0.9.26)

- Increased serverless timeout from 100ms to 1 second for larger KV operations
- Supports full state persistence in apps (e.g., pomodoro tracker)

---

## Backlog

| Item | Priority | Notes |
|------|----------|-------|
| Vue production build | Low | Use `vue.esm-browser.prod.js` in /fazt-app skill |
| Cloudflare Insights CSP | Low | Add `static.cloudflareinsights.com` to script-src |

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

| App | URL | Type |
|-----|-----|------|
| root | https://zyt.app | system |
| pomodoro | https://pomodoro.zyt.app | productivity |
| tetris | https://tetris.zyt.app | game |
| snake | https://snake.zyt.app | game |
| hello | https://hello.zyt.app | test |
| hello2 | https://hello2.zyt.app | test |
| admin | https://admin.zyt.app | system |
| 404 | (error page) | system |
