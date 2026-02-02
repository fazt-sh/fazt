---
title: "@Peer Routing Pattern"
updated: 2026-02-02
---

# @Peer Routing Pattern

## Overview

The `@peer` pattern enables remote command execution on Fazt peers. Commands prefixed with `@<peer>` are routed to the specified peer instead of executing locally.

**Example:**
```bash
fazt @zyt app list        # List apps on zyt peer
fazt @local app deploy    # Deploy to local peer
```

## Remote-Capable Commands

### App Commands
Commands that support `@peer` prefix:
- `list` - List apps
- `info` - Show app details
- `files` - List files in deployed app
- `deploy` - Deploy directory to peer
- `logs` - View serverless execution logs
- `install` - Install app from git repository
- `remove` - Remove app
- `upgrade` - Upgrade git-sourced app
- `link` - Link subdomain to app
- `unlink` - Remove alias
- `reserve` - Reserve/block subdomain
- `swap` - Atomically swap two aliases
- `split` - Configure traffic splitting
- `fork` - Fork an app
- `lineage` - Show fork tree
- `pull` - Pull app from git

### Auth Commands
- `provider` - Show OAuth provider config
- `providers` - List OAuth providers

### Peer Commands
- `list` - List configured peers
- `status` - Check peer health and status
- `add` - Add new peer
- `remove` - Remove peer
- `set-default` - Set default peer
- `upgrade` - Upgrade remote peer

### SQL Commands
- Any query execution (`fazt @peer sql <query>`)

## Local-Only Commands

These commands **do not** support `@peer` and will error if used with it:

### App Commands (Local Operations)
- `create` - Creates local files, not remote operation
- `validate` - Validates local directory before deployment

**Error message:**
```
Error: 'app create' is a local operation
This command operates on local files, not remote peers.
Usage: fazt app create ...
```

### Auth Commands (Security-Restricted)
- `users` - List users (requires direct DB access)
- `user` - Show user details
- `invite` - Create invite code
- `invites` - List invites

**Error message:**
```
Error: 'auth users' is not available remotely

User management requires direct database access for security.
To manage users on a remote peer, SSH into the server:
  ssh user@host
  fazt auth users
```

## How @Peer Routing Works

### 1. CLI Parsing
The CLI parses `@peer` prefix before subcommands:
```bash
fazt @zyt app list
     ^^^^ parsed first
          ^^^^^^^^ then subcommand routing
```

### 2. Global Context
Sets `targetPeerName` global variable that handlers check:
```go
var targetPeerName string  // Set by CLI parser
```

### 3. Peer Resolution
Commands resolve the peer using `remote.ResolvePeer()`:
```go
peer, err := remote.ResolvePeer(db, targetPeerName)
```

### 4. Remote Execution
Commands route through either:
- **Command Gateway**: `executeRemoteCmd(peer, "app", args)`
- **Direct API**: `remote.NewClient(peer).MethodName()`

## Usage Patterns

### Explicit Remote (with @peer)
```bash
fazt @zyt app list
fazt @local app deploy ./myapp
fazt @prod app files tetris
```

### Default Peer
If no `@peer` specified, uses configured default:
```bash
fazt app list              # Uses default peer
fazt peer set-default zyt  # Set default peer
```

### Positional Peer (legacy)
Some commands support positional peer argument:
```bash
fazt app list zyt
fazt app info myapp zyt
```

### Local Operations (no peer)
```bash
fazt app create my-app     # Always local
fazt app validate ./myapp  # Always local
```

## Error Handling

### Clear @Peer Errors
When `@peer` is used incorrectly, commands provide helpful guidance:

**Local-only command:**
```
$ fazt @zyt app create my-app
Error: 'app create' is a local operation
This command operates on local files, not remote peers.
Usage: fazt app create ...
```

**Remote user management:**
```
$ fazt @zyt auth users
Error: 'auth users' is not available remotely

User management requires direct database access for security.
To manage users on a remote peer, SSH into the server:
  ssh user@host
  fazt auth users
```

### Peer Resolution Errors

**No peers configured:**
```
No peers configured.
Run: fazt peer add <name> --url <url> --token <token>
```

**No default peer:**
```
Multiple peers configured. Specify which peer:
  fazt app list <name>
  fazt @<peer> app list
```

**Peer not found:**
```
Error: peer 'unknown' not found
```

## Implementation Details

### CLI Entry Point
File: `cmd/server/main.go`

```go
// Parse @peer prefix
if len(args) > 0 && strings.HasPrefix(args[0], "@") {
    targetPeerName = args[0][1:]  // Strip @
    args = args[1:]               // Remove from args
}
```

### Command Handlers
File: `cmd/server/app_v2.go`, `cmd/server/auth.go`, etc.

**Remote-capable handler pattern:**
```go
func handleAppList(args []string) {
    db := getClientDB()
    defer database.Close()

    // Resolve peer (uses targetPeerName if set)
    peer, err := remote.ResolvePeer(db, targetPeerName)

    // Execute remotely
    client := remote.NewClient(peer)
    apps, err := client.Apps()
}
```

**Local-only handler pattern:**
```go
func handleAppCreate(args []string) {
    // Guard against @peer usage
    if targetPeerName != "" {
        fmt.Fprintf(os.Stderr, "Error: 'app create' is a local operation\n")
        fmt.Fprintf(os.Stderr, "This command operates on local files, not remote peers.\n")
        os.Exit(1)
    }

    // ... local file operations
}
```

### Remote Client
File: `internal/remote/client.go`

```go
type Client struct {
    peer   *Peer
    client *http.Client
}

func (c *Client) GetAppFiles(name string) ([]FileEntry, error) {
    resp, err := c.doRequest("GET", "/api/apps/"+name+"/files", nil)
    // ... parse response
}
```

## Testing @Peer Pattern

### Verify Remote Commands Work
```bash
# Should list apps on zyt
fazt @zyt app list

# Should show files
fazt @zyt app files admin-ui

# Should check status
fazt peer status zyt
```

### Verify Local Commands Error
```bash
# Should error with helpful message
fazt @zyt app create test
fazt @local app validate ./test

# Should explain SSH requirement
fazt @zyt auth users
```

### Verify JSON Output
```bash
# All commands support --format json
fazt @zyt app list --format json
fazt @zyt app files admin-ui --format json
fazt peer status --format json
```

## Best Practices

1. **Use @peer for clarity** - Even with default peer set, `@peer` makes intent explicit
2. **Set default peer** - Reduces typing for common target: `fazt peer set-default prod`
3. **Local operations** - Never use `@peer` with `create` or `validate`
4. **User management** - SSH into server for auth user operations
5. **JSON scripting** - Use `--format json` for programmatic access
6. **Debugging** - Use `--verbose` flag to see detailed output: `fazt --verbose @zyt app list`

## Related Documentation

- **CLI Reference**: `knowledge-base/agent-context/api.md`
- **Remote Client**: `internal/remote/client.go`
- **Command Gateway**: `internal/handlers/command_handler.go`
- **Help Text**: `fazt app help`, `fazt auth help`, `fazt peer help`
