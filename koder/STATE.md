# Fazt Implementation State

**Last Updated**: 2026-01-30
**Current Version**: v0.11.9

## Status

State: **APP BUILDING** - Testing capabilities through real apps

---

## Active Directions

See `koder/THINKING_DIRECTIONS.md` for full list.

**Currently pursuing**:
- E7: Plan 28 - SPA Routing (clean URLs for BFBB apps)
- E8: Plan 29 - Private Directory (server-only data files)

**Next up**:
- P2: Nexus App (stress test all capabilities)
- E4: Plan 24 - Mock OAuth (local auth testing)

**On hold**:
- P1: Google Sign-in redirect → mostly solved by Plan 28

---

## Pending Plans

| Plan | Status | Purpose |
|------|--------|---------|
| 28: SPA Routing | Ready | Clean URLs via `--spa` flag, BFBB-compatible |
| 29: Private Directory | Ready | `private/` for server-only data files |
| 24: Mock OAuth | Not started | Local auth testing |
| 25: SQL Command | Not started | Remote DB debugging |

---

## Last Session (2026-01-30)

**BFBB Routing & Private Directory Planning**

1. **Clarified BFBB philosophy**
   - Build-Free But Buildable: source static-hostable, built output optimized
   - Fazt "quasi-builds" (copies to dist/) even without nodejs

2. **Plan 28: SPA Routing** (`koder/plans/28_spa_routing.md`)
   - `--spa` flag enables clean URLs at deploy time
   - Build-time switch: `import.meta.env.VITE_SPA_ROUTING`
   - Server serves index.html for non-file routes when `spa: true`
   - BFBB preserved: source uses hash routing (works anywhere)

3. **Plan 29: Private Directory** (`koder/plans/29_private_directory.md`)
   - `private/` directory blocked from HTTP (403)
   - Serverless can read via `fazt.private.read()`
   - Use cases: seed data, config, mock data for PoCs

4. **Updated /fazt-app skill spec** (in Plan 28)
   - Router template with env-aware history mode
   - Deploy commands with `--spa` flag
   - Updated hosting-quirks.md, auth-integration.md specs

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
