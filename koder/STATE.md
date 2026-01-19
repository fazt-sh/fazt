# Fazt Implementation State

**Last Updated**: 2026-01-20
**Current Version**: v0.10.3 (source = 0.10.3, local binary = 0.10.1, zyt = 0.10.3)

## Status

```
State: CLEAN
Next: Build more apps using the Vue + Vite pattern to validate it.
```

---

## Recent Session: Vue + Vite Refactor

### `/fazt-app` Skill Update

Made Vue + Vite **mandatory** for all fazt apps:

- Added `package.json` and `vite.config.js` templates
- Apps work in two modes:
  - **Static hosting**: Import maps resolve Vue to CDN
  - **Vite dev/build**: HMR in dev, bundled for production
- Updated component architecture: one component per `.js` file
- Added development workflow docs (Vite dev server + API proxy)

### Othelo Refactor

Refactored from single 1400-line `index.html` to modular Vue + Vite:

```
othelo/
├── package.json, vite.config.js
├── index.html (CSS + import maps)
├── main.js (3 lines - imports App)
├── components/
│   ├── App.js, StartScreen.js, GameScreen.js
│   ├── GameBoard.js, SettingsModal.js
│   ├── LeaderboardModal.js, GameOverScreen.js
├── lib/
│   ├── api.js, game.js, session.js
│   ├── settings.js, sounds.js, theme.js
└── api/main.js (unchanged - DS storage)
```

**Deploy workflow discovered**: Vite bundles JS, but fazt's serverless runtime
executes ALL `.js` files (not just `api/`). Solution: always build with Vite,
copy `api/` to `dist/`, then deploy `dist/`.

### Issue Found

Fazt serverless executes any `.js` request through Goja, even non-api files.
This prevents unbundled ES module apps. For now, Vite build is required for
production - static hosting only works for non-JS or bundled assets.

---

## Quick Reference

```bash
# Development with Vite (recommended for UI work)
cd servers/zyt/myapp && npm install && npm run dev
# http://localhost:5173 (API proxied to fazt)

# Production deploy (must build first)
npm run build
cp -r api dist/ && cp manifest.json dist/
mv dist /tmp/myapp && fazt app deploy /tmp/myapp --to local

# Local fazt server
fazt server start --domain 192.168.64.3 --port 8080 --db /tmp/fazt-local.db

# Debug endpoints
curl http://myapp.192.168.64.3.nip.io:8080/_fazt/info
curl http://myapp.192.168.64.3.nip.io:8080/_fazt/storage
```

---

## v0.10 Implementation (Complete)

| Component | File |
|-----------|------|
| Agent Endpoints | `internal/handlers/agent_handler.go` |
| Aliases | `internal/handlers/aliases_handler.go` |
| Command Gateway | `internal/handlers/cmd_gateway.go` |
| CLI v2 | `cmd/server/app_v2.go` |
| Release Script | `scripts/release.sh` |
| VFS (fixed) | `internal/hosting/vfs.go` |
| Apps Handler (fixed) | `internal/handlers/apps_handler.go` |
