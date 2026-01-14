# Fazt Implementation State

**Last Updated**: 2026-01-14
**Current Version**: v0.9.4 (local and zyt)

## Status

```
State: CLEAN
All systems operational. Remote upgrades fully working.
```

## Recent Session (2026-01-14)

### Completed

1. **Tetris Game** - 3D Three.js game at https://tetris.zyt.app
   - Pixar-style soft colors, zen UI
   - 5 color themes, light/dark mode
   - Touch controls, settings modal

2. **zyt.app Homepage** - https://zyt.app
   - Editorial design with Georgia serif
   - Apps gallery (JSON-driven via `apps.json`)
   - Light/dark mode

3. **CSP Fix (v0.9.2)** - Apps can fetch from sibling subdomains
   - `connect-src` includes `https://*.{domain}`

4. **Self-Upgrade Fix (v0.9.3)** - Binary owned by service user
   - Enables remote upgrade without sudo

5. **Atomic Upgrade (v0.9.4)** - Uses `os.Rename` instead of copy
   - Fixes "text file busy" on running binary

### Result

`fazt remote upgrade zyt` now works without SSH. Full dev lifecycle via Claude.

## Apps on zyt.app

| App | URL | Description |
|-----|-----|-------------|
| tetris | https://tetris.zyt.app | 3D Tetris game |
| xray | https://xray.zyt.app | Fazt internals visualizer |
| home | https://zyt.app | Homepage |

To update app list on homepage: edit `servers/zyt/home/apps.json`, redeploy.

## Quick Reference

| Doc | Purpose |
|-----|---------|
| `CLAUDE.md` | Primary context for Claude |
| `koder/STATE.md` | Current implementation state (this file) |
| `koder/start.md` | Deep implementation protocol |
| `koder/ideas/specs/` | Future feature specifications |

## Version History (Recent)

| Version | Date | Summary |
|---------|------|---------|
| v0.9.4 | 2026-01-14 | Atomic binary replacement |
| v0.9.3 | 2026-01-14 | Binary ownership fix |
| v0.9.2 | 2026-01-14 | CSP subdomain fix |
| v0.9.1 | 2026-01-14 | Install script service file updates |
| v0.9.0 | 2026-01-14 | Peers table, `fazt remote` commands |
