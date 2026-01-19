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

## Current Work: v0.10 - App Identity & Agent Debugging

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

**App Identity Spec** - `koder/ideas/specs/v0.10-app-identity/README.md`:
- UUID-based identity (`app_7f3k9x2m`) separate from label (`myapp`)
- Labels are mutable, swappable, optional
- Lineage tracking (`original_id`, `forked_from_id`)
- Fork/promote workflow for safe testing

### Next: Agent Debugging Endpoints

Reserved `/_fazt/` endpoints for CLI-based debugging:

| Endpoint | Purpose |
|----------|---------|
| `GET /_fazt/info` | App metadata, storage stats |
| `GET /_fazt/storage` | List all KV keys |
| `GET /_fazt/storage/:key` | Get specific value |
| `POST /_fazt/snapshot` | Create named snapshot |
| `POST /_fazt/restore/:name` | Restore snapshot |
| `GET /_fazt/logs` | Recent serverless execution logs |
| `GET /_fazt/errors` | Recent errors with traces |

**Agent workflow:**
```bash
# Test serverless function
curl -X POST http://app.192.168.64.3.nip.io:8080/api/action

# Check what happened
curl http://app.../\_fazt/logs

# Check errors if failed
curl http://app.../\_fazt/errors

# Verify storage state
curl http://app.../\_fazt/storage
```

### Implementation Order

1. Schema migration (UUID, label, lineage)
2. `fazt app rename/swap/fork/lineage` commands
3. `/_fazt/` endpoints (storage, logs, errors)
4. Update `/fazt-app` skill to use local-first workflow

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
