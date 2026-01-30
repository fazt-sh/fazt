# Fazt Implementation State

**Last Updated**: 2026-01-30
**Current Version**: v0.12.0

## Status

State: **CLEAN** - Plan 28 complete, released v0.12.0

---

## Active Directions

See `koder/THINKING_DIRECTIONS.md` for full list.

**Next up**:
- E8: Plan 29 - Private Directory (server-only data files)
- P2: Nexus App (stress test all capabilities)
- E4: Plan 24 - Mock OAuth (local auth testing)

**Resolved**:
- P1: Google Sign-in redirect → solved by Plan 28 SPA routing

---

## Pending Plans

| Plan | Status | Purpose |
|------|--------|---------|
| 29: Private Directory | Ready | `private/` for server-only data files |
| 24: Mock OAuth | Not started | Local auth testing |
| 25: SQL Command | Not started | Remote DB debugging |

---

## Last Session (2026-01-30)

**Plan 28: SPA Routing - Released v0.12.0**

Implemented clean URL support for SPAs:

1. **`--spa` flag**: `fazt app deploy ./my-app --to zyt --spa`
2. **Server fallback**: Returns `index.html` for routes without file extensions
3. **Trailing slash redirect**: 301 from `/path/` to `/path` for SEO
4. **Build env var**: `VITE_SPA_ROUTING=true` during `--spa` builds
5. **BFBB pattern**: Hash routing in dev, history routing in prod

**Files changed**:
- `internal/database/migrations/016_spa_routing.sql` - spa column
- `internal/hosting/vfs.go` - GetAppSPA/SetAppSPA methods
- `internal/hosting/handler.go` - SPA fallback + trailing slash redirect
- `cmd/server/app.go` - --spa flag on deploy command
- `internal/build/build.go` - EnvVars option for build
- `internal/remote/client.go` - DeployWithOptions

**Skill docs updated**: `~/.claude/skills/fazt-app/`
- `fazt/hosting-quirks.md` - SPA routing section + BFBB pattern
- `fazt/cli-app.md` - --spa flag documentation
- `SKILL.md` - Quick reference with --spa

---

## Quick Reference

```bash
fazt remote status zyt          # Check production
fazt app deploy ./app --to zyt --spa  # Deploy with clean URLs
journalctl --user -u fazt-local -f    # Local logs
```
