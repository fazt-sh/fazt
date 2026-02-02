# fazt peer - Peer Management

**Updated**: 2026-02-02

Manage connections to fazt instances (peers).

## Commands

### fazt peer list

List all configured peers.

```bash
fazt peer list
fazt peer list --format json    # JSON output
fazt peer list --format markdown # Markdown table (default)
```

Output:
```
NAME     URL                          DEFAULT
local    http://192.168.64.3:8080
prod     https://admin.example.com    *
```

### fazt peer add

Add a new peer connection.

```bash
fazt peer add <name> --url <url> --token <api-key>
```

**Parameters:**
- `name` - Short name for the peer (e.g., `prod`, `local`, `staging`)
- `--url` - Admin URL of the fazt instance
- `--token` - API key for authentication

**Example:**
```bash
fazt peer add prod \
  --url https://admin.example.com \
  --token fzt_abc123...
```

### fazt peer remove

Remove a peer.

```bash
fazt peer remove <name>
```

### fazt peer default

Set the default peer (used when peer not specified).

```bash
fazt peer default <name>
```

### fazt peer status

Check peer health, version, and uptime.

```bash
fazt peer status [name]
```

If name omitted, checks default peer.

Output:
```
Peer: prod (https://admin.example.com)
Status: healthy
Version: 0.11.5
Uptime: 14d 3h 22m
Apps: 12
Storage: 234 MB
```

### fazt peer upgrade

Check for or perform upgrades on a peer.

```bash
# Check if upgrade available
fazt peer upgrade check

# Perform upgrade
fazt peer upgrade <name>
```

## Remote Execution Shorthand

Execute commands on a remote peer using `@<peer>` prefix:

```bash
# Run command on specific peer
fazt @prod app list
fazt @local auth providers

# Equivalent to
fazt app list --on prod
```

## Typical Setup

```bash
# 1. Create API key on the server
fazt server create-key --name my-laptop --db /path/to/data.db

# 2. Add peer on client
fazt peer add prod \
  --url https://admin.example.com \
  --token <token-from-step-1>

# 3. Verify connection
fazt peer status prod
```
