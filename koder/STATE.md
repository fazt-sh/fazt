# Fazt Implementation State

**Last Updated**: 2026-01-19
**Current Version**: v0.10.1 (local = zyt = source ✓)

## Status

```
State: CLEAN
v0.10 App Identity fully implemented and deployed.
```

---

## Completed: v0.10 - App Identity, Aliases & Remote Execution

**Released**: v0.10.0 + v0.10.1 (bug fix)

### Features Implemented

- **App Identity**: Permanent UUIDs (`app_xxxxxxxx`) independent of aliases
- **Alias System**: Subdomains as routing layer (proxy/redirect/reserved/split)
- **Lineage Tracking**: Fork relationships with `original_id` and `forked_from_id`
- **@peer Execution**: `fazt @zyt app list` runs commands on remote peers
- **Agent Endpoints**: `/_fazt/*` for LLM testing workflows
- **Traffic Splitting**: Weighted distribution with sticky sessions
- **Fast Release Script**: `./scripts/release.sh` for local builds + upload

### Key Files

| Component | File |
|-----------|------|
| Migration | `internal/database/migrations/012_app_identity.sql` |
| ID Generation | `internal/appid/appid.go` |
| Aliases Handler | `internal/handlers/aliases_handler.go` |
| Apps v2 Handler | `internal/handlers/apps_handler_v2.go` |
| Agent Handler | `internal/handlers/agent_handler.go` |
| Command Gateway | `internal/handlers/cmd_gateway.go` |
| CLI v2 | `cmd/server/app_v2.go` |
| Release Script | `scripts/release.sh` |

---

## Quick Reference

| Command | Purpose |
|---------|---------|
| `fazt app list zyt` | List apps |
| `fazt @zyt app list` | Remote execution |
| `fazt app fork --alias myapp --as myapp-v2 --to zyt` | Fork with lineage |
| `fazt app swap alias1 alias2 --on zyt` | Atomic swap |
| `./scripts/release.sh v0.x.y "Title"` | Fast local release |

---

## Next Up

No active plan. Potential next steps from roadmap:
- v0.11: Distribution (marketplace, manifest)
- v0.12: Agentic (AI harness, ai-shim)

See `koder/ideas/ROADMAP.md` for full roadmap.

---

## Apps on zyt.app

| App | ID | URL |
|-----|----|-----|
| tetris | app_98ed8539 | https://tetris.zyt.app |
| pomodoro | app_9c479319 | https://pomodoro.zyt.app |
| othelo | app_706bd43c | https://othelo.zyt.app |
| snake | app_1d28078d | https://snake.zyt.app |
