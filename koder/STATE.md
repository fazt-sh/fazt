# Fazt Implementation State

**Last Updated**: 2026-02-01
**Current Version**: v0.17.0

## Status

State: **CLEAN** - Database path fix deployed to local & remote, Apps page refactored

---

## Last Session (2026-02-01)

**Database Path Architecture + Apps Page Refactoring**

### 1. Refactored Apps Page to Design System

- **Apps List View**: Added `getUIState()`/`setUIState()` helpers, wrapped in `.design-system-page` structure
- **Apps Detail View**: Converted to panel-based layout with collapsible sections (Details, Aliases, Files)
- Stats displayed in responsive `.panel-grid.grid-3`
- Collapse states persist per-app using `apps.detail.{appId}.{section}.collapsed`

### 2. Created Refactoring Guide

- New file: `knowledge-base/workflows/admin-ui/refactoring-pages.md`
- Step-by-step guide for converting pages to design system
- Before/after examples from Apps page
- Common mistakes and testing checklist

### 3. Fixed Database Path Architecture

**Problem**: Default `./data.db` was CWD-relative, causing confusion with multiple databases.

**Solution**: Changed default to `~/.fazt/data.db` - consistent location regardless of CWD.

- Updated `internal/database/path.go`:
  - `DefaultDBPath = "~/.fazt/data.db"`
  - Fixed `expandPath()` to expand `~` in default
- Priority chain preserved: `--db` flag > `FAZT_DB_PATH` env > default

### 4. Deployed to Local & Remote

- **Local**: Moved DB to `~/.fazt/data.db`, updated systemd service (no `--db` needed)
- **Remote (zyt)**: SSH'd in, moved `/home/fazt/.config/fazt/data.db` → `/home/fazt/.fazt/data.db`, uploaded new binary
- Both servers now use identical default path

### Key Files Modified

- `internal/database/path.go` - New default path
- `admin/src/pages/apps.js` - Design system refactor
- `knowledge-base/workflows/admin-ui/refactoring-pages.md` - NEW
- `CLAUDE.md` - Updated database section
- `~/.config/systemd/user/fazt-local.service` - Removed `--db` flag

---

## Next Up

1. **Continue Page-by-Page Refactoring**
   - Aliases page (high priority - currently placeholder)
   - System page (panel groups for metrics)
   - Settings page (panel groups for config)
   - Follow `refactoring-pages.md` guide

2. **Admin API Parity**
   - Build features to match CLI/API capabilities

---

## Quick Reference

```bash
# Database location (single DB for everything)
~/.fazt/data.db

# Override if needed
fazt server start --db /custom/path.db
# or
export FAZT_DB_PATH=/custom/path.db

# Deploy admin UI
cd admin && npm run build
fazt app deploy dist --to local --name admin-ui

# Admin UI URLs
http://admin-ui.192.168.64.3.nip.io:8080?mock=true  # Local mock
https://admin.zyt.app                                # Production
```
