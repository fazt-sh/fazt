# Fazt Implementation State

**Last Updated**: 2026-01-20
**Current Version**: v0.10.4

## Status

```
State: CLEAN
Next: Build apps with the improved static hosting.
```

---

## Current Session: Static Hosting + Headless Apps

### 1. Legacy Serverless Removal

Removed legacy root `main.js` serverless pattern. Static files (including `.js`)
now served correctly without Goja interception.

**Changes:**
- Deleted `internal/hosting/runtime.go` (-428 lines)
- Removed legacy handler call from `cmd/server/main.go`

### 2. Headless Apps Support

Apps can now be API-only (no `index.html`, just `api/main.js`).

**Changes:**
- `SiteExists()` now checks for `index.html` OR `api/main.js`
- Test added for headless API apps

### 3. Local-Only App ID Routes

Added `/_app/<id>/` escape hatch for direct app access by ID.
Only works from local/private IPs - returns 404 from public IPs.

**Use cases:**
- Local development without DNS/subdomain setup
- LLM agent testing
- Debugging

**Changes:**
- Added `IsLocalRequest()` and `ParseAppPath()` to `internal/hosting/manager.go`
- Added routing in `cmd/server/main.go`
- Tests for both functions

### 4. Development Philosophy (CLAUDE.md)

- No backward compatibility - break things and evolve
- Static hosting first - zero build steps required
- All other features are progressive enhancements

---

## Quick Reference

```bash
# Deploy static site directly (no build needed!)
fazt app deploy servers/zyt/myapp --to local

# App types
myapp/
├── index.html, *.js, *.css    # Static - served as files
└── api/main.js                # Serverless - executed by Goja

# Headless API (no index.html)
myapi/
└── api/main.js                # Just the API

# Local-only escape hatch (dev/testing)
curl http://localhost:8080/_app/myapp/api/hello

# Local fazt server
fazt server start --domain 192.168.64.3 --port 8080 --db /tmp/fazt-local.db
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
