# Fazt Implementation State

**Last Updated**: 2026-01-19
**Current Version**: v0.9.27 (local), v0.9.27 (zyt)

## Status

```
State: CLEAN
v0.10 spec complete, ready for implementation.
Spec: koder/ideas/specs/v0.10-app-identity/README.md
```

---

## Next Up: v0.10 - App Identity, Aliases & Remote Execution

**Spec**: `koder/ideas/specs/v0.10-app-identity/README.md`

### Core Design

- **Apps** = content + identity (id, title, description, tags, visibility, lineage)
- **Aliases** = routing layer (subdomain → app, supports proxy/redirect/reserved/split)
- **`@peer`** = remote execution (`fazt @zyt app list`)

### Key Decisions

- No subdomain field on apps - aliases ARE the routing layer
- Multiple aliases can point to same app
- Logs use app_id (stable across alias changes)
- CLI uses `--alias` and `--id` flags
- 1:1 CLI to API mapping via command gateway (`POST /api/cmd`)

### Implementation Phases

1. **Core Data Model**: Schema migration, ID generation, routing
2. **CLI + API**: All commands with 1:1 API mapping
3. **Remote Execution**: `@peer` parsing, command gateway
4. **Advanced**: Traffic splitting, visibility, agent endpoints

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
