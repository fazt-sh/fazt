---
command: "app list"
description: "List deployed applications on a peer"
syntax: "fazt [@peer] app list [peer] [--aliases]"
version: "0.20.0"
updated: "2026-02-02"
category: "deployment"

arguments:
  - name: "peer"
    type: "string"
    required: false
    description: "Peer name (alternative to @peer prefix)"

flags:
  - name: "--aliases"
    type: "bool"
    default: "false"
    description: "Show alias list instead of apps"
  - name: "--format"
    type: "string"
    default: "markdown"
    description: "Output format: markdown or json"

examples:
  - title: "List apps on default peer"
    command: "fazt app list"
    description: "Shows apps on the default configured peer"
  - title: "List apps on specific peer"
    command: "fazt @zyt app list"
    description: "Shows apps on the 'zyt' peer"
  - title: "List aliases"
    command: "fazt @zyt app list --aliases"
    description: "Shows subdomain aliases instead of apps"
  - title: "JSON output"
    command: "fazt app list --format json"
    description: "Machine-readable JSON output for scripting"

related:
  - command: "app info"
    description: "Show details about a specific app"
  - command: "app deploy"
    description: "Deploy a new app"
  - command: "app remove"
    description: "Remove a deployed app"
  - command: "peer list"
    description: "List configured peers"

peer:
  supported: true
  local: true
  remote: true
  syntax: "@peer prefix or positional argument"
---

# Output

The default output shows a table with:
- **ID** - App identifier (truncated)
- **Title** - App display name
- **Visibility** - public or private
- **Aliases** - Subdomain aliases

# Aliases Mode

With `--aliases`, shows:
- **Subdomain** - The alias subdomain
- **Type** - Alias type (app, redirect, etc.)
- **Target** - Target app ID

# Peer Selection

Two ways to specify the peer:

```bash
# @peer prefix (recommended)
fazt @zyt app list

# Positional argument
fazt app list zyt
```

If neither is specified, uses the default peer.
