---
command: "peer"
description: "Configure peer connections (add, remove, list)"
syntax: "fazt peer <command> [options]"
version: "0.21.0"
updated: "2026-02-02"

examples:
  - title: "List peers"
    command: "fazt peer list"
    description: "Show all configured peers"
  - title: "Add a peer"
    command: "fazt peer add zyt --url https://admin.zyt.app --token <TOKEN>"
    description: "Add a new remote peer"
  - title: "Check peer status"
    command: "fazt @zyt status"
    description: "Check health of a specific peer"
  - title: "Upgrade peer"
    command: "fazt @zyt upgrade"
    description: "Upgrade fazt binary on remote peer"

related:
  - command: "@peer status"
    description: "Check peer health"
  - command: "@peer upgrade"
    description: "Upgrade a peer"
  - command: "app"
    description: "App management"
---

# fazt peer

Configure peer connections. Peers are remote fazt instances you can manage.

## Configuration Commands

| Command | Description |
|---------|-------------|
| `list` | List all configured peers |
| `add` | Add a new peer |
| `remove` | Remove a peer |
| `default` | Set the default peer |

## @peer Commands

Operations on a specific peer use the `@peer` prefix:

| Command | Description |
|---------|-------------|
| `fazt @<peer> status` | Check peer health and version |
| `fazt @<peer> upgrade` | Upgrade fazt on remote peer |
| `fazt @<peer> app list` | List apps on peer |
| `fazt @<peer> sql "..."` | Execute SQL on peer |

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
fazt @prod status
fazt @prod upgrade
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
