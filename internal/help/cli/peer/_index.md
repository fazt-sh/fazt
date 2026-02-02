---
command: "peer"
description: "Manage remote fazt instances (peers)"
syntax: "fazt peer <command> [options]"
version: "0.20.0"
updated: "2026-02-02"

examples:
  - title: "List peers"
    command: "fazt peer list"
    description: "Show all configured peers with status"
  - title: "Add a peer"
    command: "fazt peer add zyt --url https://admin.zyt.app --token <TOKEN>"
    description: "Add a new remote peer"
  - title: "Check peer status"
    command: "fazt peer status zyt"
    description: "Show detailed status for a specific peer"
  - title: "Upgrade peer"
    command: "fazt peer upgrade zyt"
    description: "Upgrade fazt binary on remote peer"

related:
  - command: "app"
    description: "App management (use with @peer)"
  - command: "sql"
    description: "Execute SQL on peer"
---

# fazt peer

Manage remote fazt instances (peers). Peers are remote servers that you can deploy apps to.

## Commands

| Command | Description |
|---------|-------------|
| `list` | List all configured peers |
| `add` | Add a new peer |
| `remove` | Remove a peer |
| `default` | Set the default peer |
| `status` | Check peer health and version |
| `upgrade` | Upgrade fazt on remote peer |

## Adding a Peer

To add a new peer, you need:
1. The peer's URL (admin domain)
2. An API token (created with `fazt server create-key`)

```bash
# On the remote server
fazt server create-key --name my-laptop

# On your local machine
fazt peer add prod --url https://admin.example.com --token <TOKEN>
```

## Using Peers

Once configured, use the `@peer` prefix to execute commands:

```bash
fazt @prod app list
fazt @prod app deploy ./my-app
fazt @prod sql "SELECT * FROM apps"
```

## Default Peer

If you only have one peer, it becomes the default automatically.
With multiple peers, set the default:

```bash
fazt peer default prod
```

Then commands without `@peer` use the default:

```bash
fazt app list  # Uses default peer
```
