---
command: ""
description: "Sovereign compute - deploy static sites and serverless apps"
syntax: "fazt <command> [options]"
version: "0.24.7"
updated: "2026-02-04"

examples:
  - title: "Deploy an app"
    command: "fazt @zyt app deploy ./my-app"
    description: "Deploy a local directory to a remote peer"
  - title: "List apps"
    command: "fazt app list"
    description: "List deployed apps"
  - title: "List users"
    command: "fazt user list --limit 20"
    description: "List users with pagination"
  - title: "Run SQL query"
    command: "fazt sql \"SELECT * FROM apps\""
    description: "Query the local database"

related:
  - command: "app"
    description: "App management commands"
  - command: "user"
    description: "User management commands"
  - command: "alias"
    description: "Alias management commands"
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
- `fazt app list` - List deployed apps
- `fazt app status --alias <name>` - Show app status with user data
- `fazt @<peer> app <command>` - Execute app commands on a remote peer

### User Management
- `fazt user list` - List all users
- `fazt user status --email <email>` - Show user status with app data
- `fazt user set-role --email <email> --role <role>` - Set user role
- `fazt @<peer> user <command>` - Execute user commands on a remote peer

### Alias Management
- `fazt alias list` - List all aliases
- `fazt alias info --name <subdomain>` - Show alias details
- `fazt @<peer> alias <command>` - Execute alias commands on a remote peer

### Peer Management
- `fazt peer list` - List configured peers
- `fazt peer add` - Add a new peer
- `fazt @<peer> status` - Check peer health
- `fazt @<peer> upgrade` - Upgrade peer binary

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

## Pagination

List commands support pagination:

- `--offset <n>` - Skip first n results (default: 0)
- `--limit <n>` - Max results to return (default: 20, max: 100)

```bash
fazt user list --limit 50
fazt alias list --offset 20 --limit 20   # Page 2
```

## Remote Execution

Use the `@peer` prefix to execute commands on remote peers:

```bash
fazt @zyt app list           # List apps on 'zyt' peer
fazt @zyt user list          # List users on 'zyt' peer
fazt @local app deploy ./app # Deploy to 'local' peer
```
