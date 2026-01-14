# Fazt Implementation State

**Last Updated**: 2026-01-14
**Current Version**: v0.9.3 (local), v0.9.1 (zyt remote)

## Status

```
State: READY TO RELEASE
Next: v0.9.3 - Self-upgrade fix
```

## v0.9.3 Changes

### Binary Ownership Fix

**Problem**: Remote upgrade failed with permission denied.

```bash
fazt remote upgrade zyt
# Error: REPLACE_ERROR: Failed to replace binary: open /usr/local/bin/fazt: permission denied
```

**Root Cause**: Binary owned by root, service runs as user `fazt`. Unix file permissions
prevent overwriting files you don't own, regardless of systemd settings.

**Solution**: Binary is now chowned to service user during install/upgrade.

**Files Changed**:
- `install.sh`: Added chown after extracting SERVICE_USER
- `internal/provision/manager.go`: Added chown after SetCapabilities

**One-time fix for existing installs** (like zyt):
```bash
ssh user@server 'sudo chown fazt:fazt /usr/local/bin/fazt'
```

After this, remote upgrades work forever.

---

## Pending: zyt.app

After v0.9.3 release:
1. User runs SSH one-liner on zyt to fix permissions
2. `fazt remote upgrade zyt` deploys v0.9.3
3. Future upgrades work without SSH

### Workaround Still Active
Homepage uses `apps.json` instead of manifest fetching (CSP fix in v0.9.2 not deployed).

---

## Completed Today

### Tetris Game
- **Location**: `servers/zyt/tetris/`
- **Live**: https://tetris.zyt.app
- 3D Three.js, Pixar-style colors, zen UI
- Light/dark mode, 5 themes, ghost toggle
- Touch controls, responsive, settings modal

### zyt.app Homepage
- **Location**: `servers/zyt/home/`
- **Live**: https://zyt.app
- Editorial design, Georgia serif
- Hero, About (powered by fazt), Apps gallery
- Light/dark mode

---

## Quick Reference

- **Primary context**: `CLAUDE.md` (root)
- **Deep implementation**: `koder/start.md`
- **Future specs**: `koder/ideas/specs/`

## Completed Plans

| Plan | Version | Date | Summary |
|------|---------|------|---------|
| 16 | v0.8.0 | 2026-01-13 | MCP, Serverless, Apps migration |
| 17 | v0.9.0 | 2026-01-14 | Peers table, `fazt remote` commands |
| 18 | v0.9.2 | 2026-01-14 | CSP subdomain fix |
| 19 | v0.9.3 | 2026-01-14 | Binary ownership fix for self-upgrade |
