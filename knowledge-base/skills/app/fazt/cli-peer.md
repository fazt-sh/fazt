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

## @peer Commands

Operations on a specific peer use the universal `@<peer>` prefix:

### fazt @<peer> status

Check peer health, version, and uptime.

```bash
fazt @prod status
fazt @local status
```

Output:
```
Peer: prod (https://admin.example.com)
Status: healthy
Version: 0.11.5
Uptime: 14d 3h 22m
Apps: 12
Storage: 234 MB
```

### fazt @<peer> upgrade

Upgrade fazt binary on a peer.

```bash
fazt @prod upgrade
fazt @local upgrade
```

### Other @peer Commands

```bash
fazt @prod app list           # List apps on prod
fazt @prod app deploy ./myapp # Deploy to prod
fazt @prod sql "SELECT ..."   # Execute SQL on prod
fazt @prod auth providers     # List OAuth providers
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
fazt @prod status
```
