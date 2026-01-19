# Fazt Implementation State

**Last Updated**: 2026-01-19
**Current Version**: v0.10.1 (local), v0.10.0 (zyt - needs update)

## Status

```
State: PENDING RELEASE UPLOAD
v0.10.1 committed and pushed, GitHub release created but assets need manual upload.
GitHub PAT expired during release process.
```

## Action Required

1. Generate new GitHub PAT with `repo` scope
2. Upload release assets to v0.10.1:
   - `/tmp/fazt-release/fazt-v0.10.1-linux-amd64.tar.gz`
   - `/tmp/fazt-release/fazt-v0.10.1-linux-arm64.tar.gz`
3. Run `fazt remote upgrade zyt`

Or manually upload via GitHub web UI: https://github.com/fazt-sh/fazt/releases/tag/v0.10.1

---

## Completed: v0.10 - App Identity, Aliases & Remote Execution

**Implemented**: 2026-01-19

### What's New

- **App Identity**: Permanent UUIDs (`app_xxxxxxxx`) independent of aliases
- **Alias System**: Subdomains as routing layer (proxy/redirect/reserved/split)
- **Lineage Tracking**: Fork relationships with `original_id` and `forked_from_id`
- **@peer Execution**: `fazt @zyt app list` runs commands on remote peers
- **Agent Endpoints**: `/_fazt/*` for LLM testing workflows
- **Traffic Splitting**: Weighted distribution with sticky sessions

### Bug Fix in v0.10.1

After v0.10 migration, files have both `site_id` (old subdomain) and `app_id`
(new UUID). The VFS layer was querying by site_id when given app_id, causing
404s for migrated apps. Fixed by using `app_id` column for lookups.

### Implementation Files

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
| `fazt app deploy ./dir --to zyt` | Deploy to production |
| `fazt @zyt app list` | Remote command execution |
| `fazt app fork --alias myapp --as myapp-v2 --to zyt` | Fork with lineage |
| `fazt app link myalias --id app_xxx --to zyt` | Attach alias |
| `fazt app swap alias1 alias2 --on zyt` | Atomic swap |

### Local Development (Wildcard DNS)

```bash
fazt server start --domain 192.168.64.3 --port 8080 --db /tmp/fazt-local.db
# Access at http://myapp.192.168.64.3.nip.io:8080
```

---

## Apps on zyt.app

| App | ID | URL |
|-----|----|-----|
| pomodoro | app_9c479319 | https://pomodoro.zyt.app |
| tetris | app_98ed8539 | https://tetris.zyt.app |
| othelo | app_706bd43c | https://othelo.zyt.app |
| snake | app_1d28078d | https://snake.zyt.app |

---

## Related Specs

- `koder/ideas/specs/v0.10-app-identity/` - App identity, lineage, agent endpoints
