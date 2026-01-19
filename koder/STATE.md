# Fazt Implementation State

**Last Updated**: 2026-01-19
**Current Version**: v0.9.27 (local), v0.9.27 (zyt)

## Status

```
State: IN_PROGRESS
Working on: Agent-friendly API surface for CLI-based debugging
Spec: koder/ideas/specs/v0.10-app-identity/README.md
```

---

## Current Work: v0.10 - App Identity, Aliases & Agent Debugging

### Goal

Enable fully agentic/CLI-based app development and debugging:
1. Build locally (Vite, optimized)
2. Deploy to local fazt (test via API, no browser)
3. Iterate until approved
4. Deploy to production

### Completed This Session

**Wildcard DNS** - Zero-config local development:
- `--domain 192.168.64.3` auto-wraps to `192.168.64.3.nip.io`
- Apps accessible at `http://app.192.168.64.3.nip.io:8080`
- No host config needed (works across VM/host boundaries)
- Files changed: `internal/config/config.go`, `cmd/server/main.go`

**App Identity + Aliases Spec** - Major design iteration:
- **Separated apps from routing**: Apps have identity, aliases handle subdomains
- **Apps table**: id, original_id, title, description, tags, visibility
- **Aliases table**: subdomain → app_id (proxy, redirect, reserved, split)
- Multiple aliases can point to same app (solves multi-label question)
- Traffic splitting with sticky sessions (A/B testing)
- Logs use app_id (stable across renames)
- CLI uses `--alias` and `--id` flags for flexibility

**Key Design Decisions:**
- No `subdomain` field on apps - aliases ARE the routing layer
- `visibility`: public (listed), unlisted (accessible but not discoverable), private
- Metadata (title, description, tags) inherited on fork
- Reserved subdomains via `reserved` alias type

### Next: Implementation

1. Schema migration (apps + aliases tables)
2. ID generation (nanoid)
3. Routing update (check aliases first)
4. CLI `--alias/--id` flags
5. `fazt app link/unlink/split/swap` commands
6. `/_fazt/` debug endpoints
7. Traffic splitting with sticky sessions

### Spec Location

`koder/ideas/specs/v0.10-app-identity/README.md` - fully updated

---

## Quick Reference

| Command | Purpose |
|---------|---------|
| `fazt app list zyt` | List apps |
| `fazt app deploy ./dir --to local` | Deploy to local |
| `fazt app deploy ./dir --to zyt` | Deploy to production |
| `fazt remote status zyt` | Check health/version |

### Local Development (Wildcard DNS)

```bash
# Start local server (IP auto-wrapped with nip.io)
fazt server start --domain 192.168.64.3 --port 8080 --db /tmp/fazt-local.db

# Access apps
http://admin.192.168.64.3.nip.io:8080    # Dashboard
http://myapp.192.168.64.3.nip.io:8080    # Your app
```

---

## Apps

### Local (192.168.64.3.nip.io:8080)

| App | URL |
|-----|-----|
| othelo | http://othelo.192.168.64.3.nip.io:8080 |
| tetris | http://tetris.192.168.64.3.nip.io:8080 |
| snake | http://snake.192.168.64.3.nip.io:8080 |

### Production (zyt.app)

| App | URL |
|-----|-----|
| pomodoro | https://pomodoro.zyt.app |
| tetris | https://tetris.zyt.app |
| othelo | https://othelo.zyt.app |
| snake | https://snake.zyt.app |

---

## Related Specs

- `koder/ideas/specs/v0.10-app-identity/` - App identity, lineage, agent endpoints
