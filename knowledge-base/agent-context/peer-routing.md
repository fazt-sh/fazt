---
title: "@Peer Routing Pattern"
updated: 2026-02-02
---

# @Peer Routing Pattern

## Overview

The `@peer` pattern is the **universal** way to execute commands on Fazt peers. ALL commands use the same syntax:

```bash
fazt @<target> <command> [args...]
```

**Examples:**
```bash
fazt @zyt app list        # List apps on zyt peer
fazt @local app deploy    # Deploy to local peer
fazt @zyt status          # Check peer health
fazt @zyt upgrade         # Upgrade peer binary
```

## Universal Syntax

Every command follows the same pattern - no exceptions to remember:

| Command | Description |
|---------|-------------|
| `fazt @<peer> app <cmd>` | App management |
| `fazt @<peer> status` | Check peer health and version |
| `fazt @<peer> upgrade` | Upgrade fazt binary on peer |
| `fazt @<peer> sql <query>` | Execute SQL |
| `fazt @<peer> auth <cmd>` | Auth operations |
| `fazt @<peer> server <cmd>` | Server operations |
| `fazt @<peer> service <cmd>` | Service operations |
| `fazt @<peer> config <cmd>` | Config operations |

## Remote-Capable Commands

These commands execute successfully on remote peers:

### App Commands
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

### Top-Level Commands
- `status` - Check peer health and version
- `upgrade` - Upgrade fazt binary on peer

### Auth Commands
- `provider` - Show OAuth provider config
- `providers` - List OAuth providers

### SQL Commands
- Any query execution (`fazt @peer sql <query>`)

## Commands with Helpful Errors

These commands accept @peer syntax but show helpful guidance since they require local access:

### App Commands (Local Operations)
- `create` - Creates local files
- `validate` - Validates local directory

```
$ fazt @zyt app create my-app
Error: 'app create' is a local operation
This command operates on local files, not remote peers.
Usage: fazt app create ...
```

### Auth Commands (Security-Restricted)
- `users` - List users (requires direct DB access)
- `user` - Show user details
- `invite` - Create invite code
- `invites` - List invites

```
$ fazt @zyt auth users
Error: 'auth users' is not available remotely

User management requires direct database access for security.
To manage users on a remote peer, SSH into the server:
  ssh user@zyt.app
  fazt auth users
```

### Service Commands (System Access Required)
- `install` - Install systemd service
- `uninstall` - Remove systemd service
- `start` / `stop` / `restart` - Manage service

```
$ fazt @zyt service install
Error: 'service' requires local system access (systemd/sudo).

To manage the service on zyt, SSH into the machine:
  ssh user@zyt.app
  fazt service install
```

### Server Commands (Some Local-Only)
- `init` - Initialize server (requires local access)
- `start` - Can work remotely via API
- `create-key` - Works remotely

## How @Peer Routing Works

### 1. CLI Parsing
The CLI parses `@peer` prefix as the first argument:
```bash
fazt @zyt app list
     ^^^^ parsed first, sets target peer
          ^^^^^^^^ then command routing
```

### 2. Universal Routing
All commands route through `handleAtPeerRouting()`:
```go
func handleAtPeerRouting(peerName string, args []string) {
    command := args[0]  // "app", "status", "upgrade", etc.

    switch command {
    case "app":
        handleAppCommandV2(args[1:])
    case "status":
        handlePeerStatusDirect(peerName)
    case "upgrade":
        handlePeerUpgradeDirect(peerName)
    case "service":
        // Show helpful SSH guidance
        showLocalOnlyError("service", peerName)
    // ... all commands handled
    }
}
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
fazt @zyt status
fazt @zyt upgrade
```

### Default Peer
If no `@peer` specified, uses configured default:
```bash
fazt app list              # Uses default peer
fazt peer set-default zyt  # Set default peer
```

### Local Operations (no peer)
```bash
fazt app create my-app     # Always local
fazt app validate ./myapp  # Always local
fazt peer list             # Local peer config
fazt peer add prod ...     # Local peer config
```

## Error Handling

### Helpful Guidance for Local-Only Commands
Commands that can't work remotely show guidance with SSH instructions:

**Local file operations:**
```
$ fazt @zyt app create my-app
Error: 'app create' is a local operation
This command operates on local files, not remote peers.
Usage: fazt app create ...
```

**Security-restricted operations:**
```
$ fazt @zyt auth users
Error: 'auth users' is not available remotely

User management requires direct database access for security.
To manage users on a remote peer, SSH into the server:
  ssh user@zyt.app
  fazt auth users
```

**System access required:**
```
$ fazt @zyt service install
Error: 'service' requires local system access (systemd/sudo).

To manage the service on zyt, SSH into the machine:
  ssh user@zyt.app
  fazt service install
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
// Parse @peer prefix - ALWAYS first
if len(args) > 0 && strings.HasPrefix(args[0], "@") {
    peerName := args[0][1:]  // Strip @
    handleAtPeerRouting(peerName, args[1:])
    return
}
```

### Universal @Peer Router
File: `cmd/server/main.go`

```go
func handleAtPeerRouting(peerName string, args []string) {
    if len(args) == 0 {
        fmt.Fprintf(os.Stderr, "Usage: fazt @%s <command> [args...]\n", peerName)
        os.Exit(1)
    }

    command := args[0]
    cmdArgs := args[1:]

    switch command {
    case "app":
        targetPeerName = peerName
        handleAppCommandV2(cmdArgs)
    case "status":
        handlePeerStatusDirect(peerName)
    case "upgrade":
        handlePeerUpgradeDirect(peerName)
    case "sql":
        targetPeerName = peerName
        handleSQLCommand(cmdArgs)
    case "service":
        // Helpful error with SSH guidance
        fmt.Fprintf(os.Stderr, "Error: 'service' requires local system access (systemd/sudo).\n\n")
        fmt.Fprintf(os.Stderr, "To manage the service on %s, SSH into the machine:\n", peerName)
        fmt.Fprintf(os.Stderr, "  ssh user@%s.app\n", peerName)
        fmt.Fprintf(os.Stderr, "  fazt service %s\n", strings.Join(cmdArgs, " "))
        os.Exit(1)
    // ... other commands
    }
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
fazt @zyt status

# Should upgrade peer
fazt @zyt upgrade
```

### Verify Local Commands Error
```bash
# Should error with helpful message
fazt @zyt app create test
fazt @local app validate ./test

# Should explain SSH requirement
fazt @zyt server init
fazt @zyt service install
```

### Verify JSON Output
```bash
# All commands support --format json
fazt @zyt app list --format json
fazt @zyt app files admin-ui --format json
fazt @zyt status --format json
```

## Best Practices

1. **Use @peer for clarity** - Even with default peer set, `@peer` makes intent explicit
2. **Set default peer** - Reduces typing for common target: `fazt peer set-default prod`
3. **Universal syntax** - Use `fazt @peer <command>` for everything - errors guide you when needed
4. **SSH for local-only** - Commands that need local access will tell you exactly what SSH command to run
5. **JSON scripting** - Use `--format json` for programmatic access
6. **Debugging** - Use `--verbose` flag to see detailed output: `fazt --verbose @zyt app list`

## Peer Configuration (Local Commands)

These commands manage local peer configuration (don't use @peer prefix):

```bash
fazt peer list             # List configured peers
fazt peer add prod ...     # Add a peer
fazt peer remove prod      # Remove a peer
fazt peer set-default prod # Set default peer
```

## Related Documentation

- **CLI Reference**: `knowledge-base/agent-context/api.md`
- **Remote Client**: `internal/remote/client.go`
- **Command Gateway**: `internal/handlers/command_handler.go`
- **Help Text**: `fazt app help`, `fazt peer help`
