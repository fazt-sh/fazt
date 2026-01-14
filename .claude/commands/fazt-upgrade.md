---
description: Check for and upgrade fazt server to latest version
model: haiku
allowed-tools: Read, Glob, Bash, WebFetch
---

# Fazt Upgrade

Check for updates and upgrade a remote fazt instance.

## Arguments

`$ARGUMENTS` format: `[check] [--server=<name>]`

Options:
- (none) - Check and prompt to upgrade
- `check` - Only check for updates, don't upgrade

Examples:
- `/fazt-upgrade` - Check and offer upgrade
- `/fazt-upgrade check` - Check only
- `/fazt-upgrade --server=zyt` - Check specific server

## Server Selection

1. List available servers: `ls servers/`
2. Parse `--server=<name>` if provided
3. If one server, use it; if multiple and no flag, ask user

## Read Server Config

```
servers/<name>/config.json
```

Token lookup: config.json `token` field, then `$FAZT_TOKEN_<NAME>` env var.

## Check for Updates

```bash
curl -s -X POST "<url>/api/upgrade?check=true" \
  -H "Authorization: Bearer <token>"
```

Response fields:
- `current_version`: Running version
- `new_version`: Latest available
- `action`: "already_latest" or "check_only"

Display:
```
Server: <name> (<domain>)
Current: v<current_version>
Latest:  v<new_version>

[Already on latest / Update available]
```

## Perform Upgrade

If update available and not `check` mode:

1. **Confirm**: "Upgrade <name> from v<current> to v<new>? (y/n)"

2. **Execute**:
   ```bash
   curl -s -X POST "<url>/api/upgrade" \
     -H "Authorization: Bearer <token>"
   ```

3. **Report result**:
   - Success: "Upgraded! Server will restart."
   - Failure: Report error message

## Notes

- Upgrade causes server restart (~2-5 seconds downtime)
- Backup is created automatically before upgrade
- Rollback happens automatically if upgrade fails
- Only works if server has GitHub access

## Error Handling

- No token: "No API token for server '<name>'"
- GitHub unreachable: "Failed to check latest release"
- No matching binary: "No release for this platform"
