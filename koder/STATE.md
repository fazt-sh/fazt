# Fazt Implementation State

**Last Updated**: 2026-01-20
**Current Version**: v0.10.5

## Status

```
State: IN_PROGRESS
Next: Rebuild binary to include template fixes
```

---

## Completed: Simplify /fazt-app

Reduced `/fazt-app` skill from 1004 lines to 127 lines (87% reduction).

### Changes Made

1. **Skill rewritten** (`.claude/commands/fazt-app.md`)
   - Workflow-focused instead of code-heavy
   - References CLI commands and template
   - Quick storage reference
   - 127 lines total

2. **Template fixes** (`internal/assets/templates/vue-api/`)
   - Renamed `api/items.js` → `api/main.js` (serverless requires main.js)
   - Added `handler(request)` call at end (was missing execution)
   - Added `genId()` helper (fazt.uuid() doesn't exist)
   - Fixed vite.config.js to externalize Vue for build

### Template Files Changed

| File | Change |
|------|--------|
| `api/main.js` | Renamed from items.js, added handler call, added genId |
| `vite.config.js` | Added Vue externalization for Vite build |

---

## Action Required: Rebuild Binary

The template fixes are in source but not in the installed binary. To apply:

```bash
go build -o ~/.local/bin/fazt ./cmd/server
```

Until rebuilt, `fazt app create` uses old templates. Workaround: use `--no-build`
when deploying to skip Vite build step.

---

## Verified Working

Tested full workflow:
```bash
fazt app create myapp --template vue-api
fazt app validate ./myapp
fazt app deploy ./myapp --to local --no-build
curl http://myapp.192.168.64.3.nip.io:8080/api/items  # ✓ returns JSON
```

---

## Quick Reference

```bash
# Create app
fazt app create myapp --template vue-api

# Validate
fazt app validate ./myapp

# Local testing
fazt app deploy ./myapp --to local --no-build
fazt app logs myapp --peer local -f

# Production
fazt app deploy ./myapp --to zyt
```
