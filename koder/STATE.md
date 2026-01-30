# Fazt Implementation State

**Last Updated**: 2026-01-30
**Current Version**: v0.13.0

## Status

State: **READY** - Plan 30 split into 30a/30b/30c, ready to implement

---

## Active Directions

See `koder/THINKING_DIRECTIONS.md` for full list.

**Next up**:
- **Plan 30a: Config Consolidation** (v0.14.0) - Start here
- Plan 30b: User Data Foundation (v0.15.0)
- Plan 30c: Access Control (v0.16.0)
- P2: Nexus App (stress test all capabilities)
- E4: Plan 24 - Mock OAuth (local auth testing)

---

## Pending Plans

| Plan | Status | Purpose | Target |
|------|--------|---------|--------|
| 30a: Config Consolidation | **Ready** | Single DB philosophy | v0.14.0 |
| 30b: User Data Foundation | Ready | User isolation, IDs, analytics | v0.15.0 |
| 30c: Access Control | Ready | RBAC, domain gating | v0.16.0 |
| 29: Private Directory | ✅ Released | `private/` with dual access | v0.13.0 |
| 24: Mock OAuth | Not started | Local auth testing | - |
| 25: SQL Command | Not started | Remote DB debugging | - |

---

## Current Session (2026-01-30)

**Plan 30 Expansion + Split**

### RBAC Design

Designed lightweight RBAC with:
- Hierarchical roles (`teacher/chemistry/*`)
- Expiry support for temporary access
- `fazt.auth.hasRole(pattern)` / `requireRole(pattern)`
- `fazt.admin.users.role.(add|remove|list|find)`

### Domain Gating

Designed email domain restrictions:
- `fazt.admin.config.auth.domains.(allow|block|list)`
- Whitelist (employees only) and blacklist (no disposable emails)
- Enforced at OAuth callback
- API-configurable (no code changes)

### Config Consolidation

**Critical fix**: External config files violate single-DB philosophy.

- `~/.config/fazt/config.json` → Move to DB
- `~/.fazt/config.json` → Move to DB
- Single `config` table with hierarchical keys
- Bootstrap only needs DB path (flag/env/default)

### Plan Split

Split Plan 30 into three focused plans:

```
30a: Config in DB (foundation)
 ↓
30b: User isolation, IDs, analytics
 ↓
30c: RBAC, domain gating
```

### Files Created/Changed

- `koder/plans/30a_config_consolidation.md` - Config in DB
- `koder/plans/30b_user_data_foundation.md` - User isolation, IDs
- `koder/plans/30c_access_control.md` - RBAC, domain gating
- `koder/plans/30_user_isolation_analytics.md` - Updated as index

---

## Quick Reference

```bash
# Deploy with private files
fazt app deploy ./my-app --to zyt --include-private

# Plan 30a - config in DB
# No more ~/.config/fazt/config.json or ~/.fazt/config.json
# Everything in data.db

# Plan 30b - new API examples
fazt.app.user.ds.insert('settings', { theme: 'dark' })
fazt.admin.users.delete('fazt_usr_Nf4rFeUfNV2H')

# Plan 30c - RBAC examples
fazt.auth.requireRole('dept/engineering')
fazt.admin.config.auth.domains.allow(['storybrain.com'])
```
