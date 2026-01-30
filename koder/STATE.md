# Fazt Implementation State

**Last Updated**: 2026-01-30
**Current Version**: v0.13.0

## Status

State: **COMPLETE** - Plan 30a done, ready for v0.14.0 release

---

## Active Directions

See `koder/THINKING_DIRECTIONS.md` for full list.

**Current**:
- **Plan 30a: Config Consolidation** (v0.14.0) - ✅ Complete

**Next up**:
- v0.14.0 Release
- Plan 30b: User Data Foundation (v0.15.0)
- Plan 30c: Access Control (v0.16.0)
- P2: Nexus App (stress test all capabilities)
- E4: Plan 24 - Mock OAuth (local auth testing)

---

## Pending Plans

| Plan | Status | Purpose | Target |
|------|--------|---------|--------|
| 30a: Config Consolidation | ✅ Complete | Single DB philosophy | v0.14.0 |
| 30b: User Data Foundation | Ready | User isolation, IDs, analytics | v0.15.0 |
| 30c: Access Control | Ready | RBAC, domain gating | v0.16.0 |
| 29: Private Directory | ✅ Released | `private/` with dual access | v0.13.0 |
| 24: Mock OAuth | Not started | Local auth testing | - |
| 25: SQL Command | Not started | Remote DB debugging | - |

---

## Current Session (2026-01-30)

**Plan 30a: Config Consolidation - Implementation**

### Completed

1. **Unified DB Path Resolution** (`internal/database/path.go`)
   - `database.ResolvePath(explicit)` - single source of truth
   - Priority: explicit flag > `FAZT_DB_PATH` env > `./data.db` default
   - Added tests in `path_test.go`

2. **Instance Config Migration** (`internal/config/migrate.go`)
   - `config.MigrateFromFile(db)` - migrates `~/.config/fazt/config.json` to DB
   - `config.LoadFromDB(db)` - loads config from DB with defaults
   - `config.SaveToDB(db, cfg)` - persists config to DB
   - Auto-renames migrated files to `.bak`

3. **Client DB Consolidation** (`cmd/server/main.go`)
   - Updated `getClientDB()` to use unified path resolution
   - Added `migrateLegacyClientDB()` to migrate peers from `~/.config/fazt/data.db`
   - Server startup now calls `config.MigrateFromFile()` after DB init

4. **Removed Legacy Commands & Packages**
   - Deleted `internal/mcp/` - skills replace MCP paradigm
   - Deleted `internal/clientconfig/` - config in DB now
   - Removed `fazt servers` handlers → error redirects to `fazt remote`
   - Removed `fazt deploy` / `fazt client deploy` → error redirects to `fazt app deploy`
   - Removed `fazt client apps` → error redirects to `fazt app list`
   - Removed MCP HTTP routes (`/mcp/*`) from server

### Files Changed

- `internal/database/path.go` - NEW: unified path resolution
- `internal/database/path_test.go` - NEW: tests
- `internal/config/migrate.go` - NEW: JSON→DB migration
- `internal/mcp/` - DELETED
- `internal/clientconfig/` - DELETED
- `cmd/server/main.go` - Removed MCP, clientconfig, legacy commands

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
