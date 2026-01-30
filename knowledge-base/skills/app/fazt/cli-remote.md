# fazt remote - Peer Management

Manage connections to fazt instances (peers).

## Commands

### fazt remote list

List all configured peers.

```bash
fazt remote list
```

Output:
```
NAME     URL                          DEFAULT
local    http://192.168.64.3:8080
prod     https://admin.example.com    *
```

### fazt remote add

Add a new peer connection.

```bash
fazt remote add <name> --url <url> --token <api-key>
```

**Parameters:**
- `name` - Short name for the peer (e.g., `prod`, `local`, `staging`)
- `--url` - Admin URL of the fazt instance
- `--token` - API key for authentication

**Example:**
```bash
fazt remote add prod \
  --url https://admin.example.com \
  --token fzt_abc123...
```

### fazt remote remove

Remove a peer.

```bash
fazt remote remove <name>
```

### fazt remote default

Set the default peer (used when peer not specified).

```bash
fazt remote default <name>
```

### fazt remote status

Check peer health, version, and uptime.

```bash
fazt remote status [name]
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

### fazt remote upgrade

Check for or perform upgrades on a peer.

```bash
# Check if upgrade available
fazt remote upgrade check

# Perform upgrade
fazt remote upgrade <name>
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
fazt remote add prod \
  --url https://admin.example.com \
  --token <token-from-step-1>

# 3. Verify connection
fazt remote status prod
```
