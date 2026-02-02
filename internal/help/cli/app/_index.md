---
command: "app"
description: "Manage applications - deploy, list, remove, and configure"
syntax: "fazt [@peer] app <command> [options]"
version: "0.20.0"
updated: "2026-02-02"

examples:
  - title: "List apps on default peer"
    command: "fazt app list"
    description: "Shows apps on the default configured peer"
  - title: "List apps on specific peer"
    command: "fazt @zyt app list"
    description: "Shows apps on the 'zyt' peer"
  - title: "Deploy an app"
    command: "fazt @zyt app deploy ./my-app"
    description: "Deploy a local directory to a remote peer"
  - title: "Get app info"
    command: "fazt @zyt app info --alias tetris"
    description: "Show details about a deployed app"

related:
  - command: "app deploy"
    description: "Deploy a directory to a peer"
  - command: "app list"
    description: "List deployed apps"
  - command: "app remove"
    description: "Remove a deployed app"
  - command: "peer"
    description: "Manage remote peers"
---

# fazt app

Manage applications on fazt instances.

## Remote Commands (support @peer)

These commands can be executed on remote peers using the `@peer` prefix:

| Command | Description |
|---------|-------------|
| `list` | List apps (--aliases for alias list) |
| `info` | Show app details (--alias or --id) |
| `files` | List files in a deployed app |
| `deploy` | Deploy directory to peer |
| `logs` | View serverless execution logs |
| `install` | Install app from git repository |
| `remove` | Remove app |
| `upgrade` | Upgrade git-sourced app |
| `pull` | Download app files from peer |

## Alias Management

| Command | Description |
|---------|-------------|
| `link` | Link subdomain to app (--id required) |
| `unlink` | Remove alias |
| `reserve` | Reserve/block subdomain |
| `swap` | Atomically swap two aliases |
| `split` | Configure traffic splitting |

## Local Commands (no @peer support)

| Command | Description |
|---------|-------------|
| `create` | Create local app from template |
| `validate` | Validate local directory |

## Identification Options

Apps can be identified by:
- `--alias <name>` - Reference by subdomain alias
- `--id <app_id>` - Reference by app ID

## Peer Selection

Use `@<peer>` prefix for remote operations:
```bash
fazt @zyt app list              # List apps on zyt peer
fazt @local app deploy ./myapp  # Deploy to local peer
```
