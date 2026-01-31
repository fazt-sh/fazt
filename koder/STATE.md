# Fazt Implementation State

**Last Updated**: 2026-01-31
**Current Version**: v0.17.0

## Status

State: **CLEAN** - Admin UI foundation complete

---

## Last Session (2026-01-31)

**Fazt Admin UI Foundation - BFBB Architecture**

Built a complete admin UI using Build-Free But Buildable (BFBB) pattern.

1. **Created packages** (`servers/local/admin/packages/`):
   - `zap/` - State management (atoms, maps) + router + command palette
   - `fazt-sdk/` - API client with mock adapter
   - `fazt-ui/` - CSS design system (tokens, utilities)

2. **Built admin app** (`servers/local/admin/`):
   - Dashboard with stats, apps table, activity feed
   - Apps, Aliases, System, Settings pages
   - Theme/palette switcher (5 palettes, light/dark)
   - Command palette (Cmd+K)
   - SPA mode with clean URLs (history routing)

3. **Key features**:
   - Pure ESM modules (no build required)
   - Tailwind CDN for utilities
   - Mock data mode (`?mock=true`)
   - Footer toggles for mock mode and settings panel
   - AI agent interface (`window.__fazt_agent`)

**Deployed**: `http://admin-ui.192.168.64.3.nip.io:8080`

---

## Next Up

1. **Refine other pages** (Apps, Aliases, System, Settings)
   - Fix layout issues
   - Match dashboard polish level

2. **Wire to real API**
   - Connect fazt-sdk to actual endpoints
   - Test with live data

3. **Add features**:
   - App detail page
   - Real-time updates
   - More command palette actions

---

## Quick Reference

```bash
# Deploy admin UI
fazt app deploy servers/local/admin --to local --name admin-ui

# Test with mock data
http://admin-ui.192.168.64.3.nip.io:8080?mock=true

# View source (BFBB - no build)
ls servers/local/admin/packages/
ls servers/local/admin/src/
```

---

## LEGACY_CODE Markers

```bash
grep -rn "LEGACY_CODE" internal/
```

- `internal/storage/bindings.go` - `fazt.storage.*` namespace
- `internal/appid/appid.go` - old `app_*` format
- `internal/auth/service.go` - `generateUUID()`

See `koder/LEGACY.md` for removal guide.
