# Fazt Implementation State

**Last Updated**: 2026-01-14
**Current Version**: v0.9.4 (local and zyt)

## Status

```
State: CLEAN
All systems operational. Remote upgrades fully working.
```

## Session Summary (2026-01-14)

### Completed

1. **Tetris Game** - https://tetris.zyt.app
   - 3D Three.js, Pixar-style colors, zen UI
   - 5 themes, light/dark mode, touch controls

2. **zyt.app Homepage** - https://zyt.app
   - Editorial design, Georgia serif
   - JSON-driven apps gallery (`servers/zyt/home/apps.json`)

3. **Remote Upgrade Fixes**
   - v0.9.2: CSP allows subdomain communication
   - v0.9.3: Binary owned by service user
   - v0.9.4: Atomic rename (fixes "text file busy")

4. **Claude Skills Cleanup**
   - Added: `/fazt-start`, `/fazt-stop`
   - Updated: `/fazt-release` (idempotent)
   - Removed: `/fazt-upgrade`, `/fazt-status`, `/fazt-apps`, `/fazt-deploy`
     (simple CLI commands, documented in CLAUDE.md)

### Current Skills

| Skill | Purpose |
|-------|---------|
| `/fazt-start` | Session context loading |
| `/fazt-stop` | Session documentation |
| `/fazt-release` | Full release workflow (idempotent) |
| `/fazt-ideate` | Brainstorm ideas |
| `/fazt-lite-extract` | Evaluate library extraction |

### CLI Commands (no skill needed)

```bash
fazt remote status zyt      # Health, version
fazt remote apps zyt        # List apps
fazt remote deploy <dir> zyt # Deploy app
fazt remote upgrade zyt     # Upgrade server
```

## Apps on zyt.app

| App | URL |
|-----|-----|
| home | https://zyt.app |
| tetris | https://tetris.zyt.app |
| xray | https://xray.zyt.app |

## Quick Reference

| Doc | Purpose |
|-----|---------|
| `CLAUDE.md` | Primary context |
| `koder/STATE.md` | Current state (this file) |
| `koder/start.md` | Deep implementation protocol |
| `koder/ideas/specs/` | Future feature specs |

## Next Session

Run `/fazt-start` or read this file to get context.
