# Fazt Implementation State

**Last Updated**: 2026-01-30
**Current Version**: v0.11.9

## Status

State: **APP BUILDING** - Testing capabilities through real apps

---

## Active Directions

See `koder/THINKING_DIRECTIONS.md` for full list.

**Currently pursuing**:
- E8: Plan 29 - Private Directory (server-only data files)

**Next up**:
- P2: Nexus App (stress test all capabilities)
- E4: Plan 24 - Mock OAuth (local auth testing)

**On hold**:
- P1: Google Sign-in redirect → solved by Plan 28

---

## Pending Plans

| Plan | Status | Purpose |
|------|--------|---------|
| 28: SPA Routing | **Done** | Clean URLs via `--spa` flag |
| 29: Private Directory | Ready | `private/` for server-only data files |
| 24: Mock OAuth | Not started | Local auth testing |
| 25: SQL Command | Not started | Remote DB debugging |

---

## Last Session (2026-01-30)

**Plan 28: SPA Routing Implemented**

1. **Database**: Migration 016 adds `spa` column to apps table
2. **VFS**: `GetAppSPA`/`SetAppSPA` methods for reading/writing spa flag
3. **Handler**: SPA fallback in `ServeVFS` - serves index.html for non-file routes
4. **Build**: `EnvVars` option sets `VITE_SPA_ROUTING=true` during build
5. **CLI**: `--spa` flag on `fazt app deploy` command
6. **Remote**: `DeployWithOptions` passes spa flag to server

**Usage**:
```bash
fazt app deploy ./my-spa --to zyt --spa
```

The app's router should check `import.meta.env.VITE_SPA_ROUTING` to switch
between hash routing (default) and history routing (SPA mode).

**Tested**: Clean URLs work (e.g., `/dashboard` returns index.html).
Without `--spa`, same route returns 404.

---

## Recent Completions

### v0.11.9 - Security Hardening

Protection stack complete:
```
TCP_DEFER_ACCEPT → ConnLimiter → TLS → ReadHeaderTimeout → Rate Limit
```

| Issue | Status |
|-------|--------|
| Slowloris | ✅ Fixed |
| Rate limiting | ✅ 500 req/s/IP |
| Connection exhaustion | ✅ 50 conns/IP |
| Header timeout | ✅ 5s |

---

## Quick Reference

See `CLAUDE.md` for full command reference.

```bash
fazt remote status zyt          # Check production
ssh root@165.227.11.46          # Emergency SSH
journalctl --user -u fazt-local -f  # Local logs
```
