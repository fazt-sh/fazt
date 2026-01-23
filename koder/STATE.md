# Fazt Implementation State

**Last Updated**: 2026-01-24
**Current Version**: v0.10.8

## Status

State: PLANNING COMPLETE - Ready to rebuild CashFlow

---

## Last Session (2026-01-24)

### DevTools Planning & App Architecture

Major planning session for fazt app development patterns and devtools infrastructure.

**Key Decisions:**

1. **"Build-free, but buildable"** - Apps work served raw AND built with Vite/Bun
2. **Pinia for state management** - Battle-tested, works without build
3. **Import maps** - Clean imports (`from 'vue'`) that work raw and built
4. **Multi-page structure** - Router, stores, pages, components hierarchy
5. **DevTools in fazt** - Not browser extension, server-native streaming

**Files Created/Updated:**

Plan:
- `koder/plans/20_devtools.md` - Comprehensive devtools plan

Skill templates (`.claude/skills/fazt-app/templates/`):
- `index.html` - PWA setup with import maps
- `api/main.js` - Session-scoped API template
- `src/main.js` - App initialization with Pinia/Router
- `src/App.js` - Root component
- `src/router.js` - Client-side routing
- `src/stores/app.js` - Pinia store template
- `src/pages/Home.js`, `Settings.js` - Page templates
- `src/components/ui/Button.js`, `Modal.js`, `Card.js` - UI components
- `src/lib/api.js`, `session.js`, `settings.js` - Utilities

Skill updates:
- Removed hardcoded "zyt" references
- Updated template references
- Added note about app_id being auto-generated UUID

Infrastructure:
- `internal/middleware/security.go` - Changed CSP to allow iframe embedding
- Installed `agent-browser` for headless browser testing

**What Exists in Fazt (Researched):**
- `FAZT_DEBUG=1` with structured logging
- `/_fazt/info`, `/_fazt/logs`, `/_fazt/errors` endpoints
- `/_fazt/storage`, `/_fazt/snapshot`, `/_fazt/restore` endpoints
- `fazt.storage.kv/ds/s3` with MongoDB-style queries
- `site_logs` table for execution log persistence

**What Needs Building:**
- `/_fazt/stream` - SSE endpoint for real-time events
- `/_fazt/events` - POST endpoint for client-side events
- `/_fazt/scripts` - Injectable JS library
- Shared component library (deferred)

---

## Next Session

**Task**: Rebuild CashFlow app using new templates and patterns

Goals:
1. Use proper multi-page structure (router, stores, pages, components)
2. Use Pinia for state management
3. Use import maps
4. Validate "build-free, but buildable" works
5. Test thoroughly with agent-browser
6. Make it a reference implementation

---

## Quick Reference

```bash
# Local server
FAZT_DEBUG=1 fazt server start --port 8080 --domain 192.168.64.3 --db servers/local/data.db

# Check peers
fazt remote list

# Deploy locally first
fazt app deploy servers/zyt/<name> --to local

# Deploy to production (after approval)
fazt app deploy servers/zyt/<name> --to zyt

# Agent browser testing
agent-browser open http://<app>.192.168.64.3.nip.io:8080
agent-browser snapshot
agent-browser eval "..."

# Release
source .env && ./scripts/release.sh vX.Y.Z
```

## Architecture Notes

**App Structure (template):**
```
app-name/
├── index.html          # Import maps, PWA meta
├── api/main.js         # Serverless API
└── src/
    ├── main.js         # App init (Pinia, Router)
    ├── App.js          # Root component
    ├── router.js       # Routes
    ├── stores/         # Pinia stores
    ├── pages/          # Page components
    ├── components/ui/  # Reusable UI
    └── lib/            # Utilities
```

**app_id vs alias:**
- `app_id` = UUID like `app_sd7w8rvt` (auto-generated)
- `alias` = subdomain like `cashflow` (from manifest name)
