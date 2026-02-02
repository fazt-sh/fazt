---
command: ""
description: "Sovereign compute - deploy static sites and serverless apps"
syntax: "fazt <command> [options]"
version: "0.20.0"
updated: "2026-02-02"

examples:
  - title: "Deploy an app"
    command: "fazt @zyt app deploy ./my-app"
    description: "Deploy a local directory to a remote peer"
  - title: "List apps"
    command: "fazt app list"
    description: "List deployed apps"
  - title: "Run SQL query"
    command: "fazt sql \"SELECT * FROM apps\""
    description: "Query the local database"

related:
  - command: "app"
    description: "App management commands"
  - command: "peer"
    description: "Peer management commands"
  - command: "server"
    description: "Server management commands"
---

# fazt

Sovereign compute - single Go binary + SQLite database that runs anywhere.

## Commands

### App Management
- `fazt app <command>` - Deploy, manage, and monitor applications
- `fazt @<peer> app <command>` - Execute app commands on a remote peer

### Peer Management
- `fazt peer list` - List configured peers
- `fazt peer add` - Add a new peer
- `fazt peer status` - Check peer health

### Server Management
- `fazt server init` - Initialize a new server
- `fazt server start` - Start the server
- `fazt server status` - Show server status

### Utilities
- `fazt sql <query>` - Execute SQL queries
- `fazt upgrade` - Upgrade fazt binary

## Global Flags

- `--verbose` - Show detailed output (migrations, debug info)
- `--format <fmt>` - Output format: markdown (default) or json
- `--help, -h` - Show help for any command

## Remote Execution

Use the `@peer` prefix to execute commands on remote peers:

```bash
fazt @zyt app list           # List apps on 'zyt' peer
fazt @local app deploy ./app # Deploy to 'local' peer
```
