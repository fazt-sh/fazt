---
description: List and manage apps on a fazt server
model: haiku
allowed-tools: Read, Glob, Bash, WebFetch
---

# Fazt Apps

List, view, and manage apps on a fazt instance.

## Arguments

`$ARGUMENTS` format: `[command] [app_name] [--server=<name>]`

Commands:
- (none) - List all apps
- `list` - List all apps
- `show <name>` - Show app details
- `files <name>` - List app files
- `delete <name>` - Delete an app

Examples:
- `/fazt-apps` - List apps on default server
- `/fazt-apps --server=zyt` - List apps on zyt
- `/fazt-apps show blog` - Show details for "blog" app
- `/fazt-apps delete old-site` - Delete "old-site" app

## Server Selection

1. List available servers: `ls servers/`
2. Parse `--server=<name>` if provided
3. If one server, use it; if multiple and no flag, ask user

## Read Server Config

```
servers/<name>/config.json
```

Token lookup: config.json `token` field, then `$FAZT_TOKEN_<NAME>` env var.

## Commands

### List Apps (default)

```bash
curl -s "<url>/api/apps" -H "Authorization: Bearer <token>"
```

Display as table:
```
Apps on <server_name> (<domain>)

Name          Files    Size      Updated
----          -----    ----      -------
blog          42       128KB     2024-01-15
docs          18       64KB      2024-01-14
admin         12       256KB     2024-01-10
```

### Show App Details

```bash
curl -s "<url>/api/apps/<name>" -H "Authorization: Bearer <token>"
```

Display:
```
App: <name>
Source: <source>
Files: <count>
Size: <bytes>
Created: <date>
Updated: <date>

Manifest:
<json manifest if present>
```

### List App Files

```bash
curl -s "<url>/api/apps/<name>/files" -H "Authorization: Bearer <token>"
```

Display file tree.

### Delete App

**Confirm before deletion**: "Delete app '<name>' from <server>? (y/n)"

```bash
curl -s -X DELETE "<url>/api/apps/<name>" -H "Authorization: Bearer <token>"
```

Report success/failure.

## Error Handling

- No token: "No API token for server '<name>'"
- App not found: "App '<name>' not found on <server>"
- Cannot delete system apps (root, 404, admin)
