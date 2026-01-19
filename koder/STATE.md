# Fazt Implementation State

**Last Updated**: 2026-01-19
**Current Version**: v0.10.0 (local), v0.9.27 (zyt - pending update)

## Status

```
State: CLEAN
v0.10 implementation complete, pending zyt deployment.
```

---

## Completed: v0.10 - App Identity, Aliases & Remote Execution

**Implemented**: 2026-01-19
**Spec**: `koder/ideas/specs/v0.10-app-identity/README.md`

### What's New

- **App Identity**: Permanent UUIDs (`app_xxxxxxxx`) independent of aliases
- **Alias System**: Subdomains as routing layer (proxy/redirect/reserved/split)
- **Lineage Tracking**: Fork relationships with `original_id` and `forked_from_id`
- **@peer Execution**: `fazt @zyt app list` runs commands on remote peers
- **Agent Endpoints**: `/_fazt/*` for LLM testing workflows
- **Traffic Splitting**: Weighted distribution with sticky sessions

### Implementation Details

| Component | File |
|-----------|------|
| Migration | `internal/database/migrations/012_app_identity.sql` |
| ID Generation | `internal/appid/appid.go` |
| Aliases Handler | `internal/handlers/aliases_handler.go` |
| Apps v2 Handler | `internal/handlers/apps_handler_v2.go` |
| Agent Handler | `internal/handlers/agent_handler.go` |
| Command Gateway | `internal/handlers/cmd_gateway.go` |
| CLI v2 | `cmd/server/app_v2.go` |

---

## Quick Reference

| Command | Purpose |
|---------|---------|
| `fazt app list zyt` | List apps |
| `fazt app deploy ./dir --to local` | Deploy to local |
| `fazt app deploy ./dir --to zyt` | Deploy to production |
| `fazt @zyt app list` | Remote command execution |
| `fazt app fork <id> --alias new-name` | Fork with lineage |
| `fazt app link <id> <alias>` | Attach alias |
| `fazt app swap alias1 alias2` | Atomic swap |

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
