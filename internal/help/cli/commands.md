# Fazt CLI Commands Reference

**Version**: 0.24.13
**Updated**: 2026-02-04

This document catalogs all CLI commands, their arguments, flags, and patterns for consistency analysis.

---

## Command Structure

```
fazt [global-flags] <command> [subcommand] [flags] [arguments]
fazt [global-flags] @<peer> <command> [subcommand] [flags] [arguments]
```

---

## Global Flags

These flags work with any command and can be placed anywhere in the command:

| Flag | Description | Example |
|------|-------------|---------|
| `--verbose` | Show detailed output (migrations, debug info) | `fazt --verbose app list` |
| `--format <fmt>` | Output format: `markdown` (default) or `json` | `fazt peer list --format json` |

**Note**: Global flags are position-independent and can appear anywhere:
```bash
fazt --verbose @local app list
fazt @local app list --verbose
fazt @local --verbose app list
```

---

## Top-Level Commands

### 1. `fazt app`

**Purpose**: App management (list, deploy, info, remove)

#### Subcommands

##### `app list`
- **Args**: None
- **Flags**:
  - `--aliases` - Show alias list instead of apps
- **Pattern**: Local by default, remote via `@peer` prefix
- **Remote**: `fazt @peer app list`

##### `app info [identifier]`
- **Args**: `[identifier]` - Optional app identifier
- **Flags**:
  - `--alias <name>` - Reference by alias
  - `--id <app_id>` - Reference by ID
- **Pattern**: Local by default, remote via `@peer` prefix

##### `app deploy <dir>`
- **Args**: `<dir>` - Required directory path
- **Flags**:
  - `--name <name>` - Optional app name
  - `--spa` - Enable SPA routing
  - `--no-build` - Skip build step
  - `--include-private` - Include private/ directory
- **Pattern**: Local by default, remote via `@peer` prefix

##### `app validate <dir>`
- **Args**: `<dir>` - Required directory path
- **Flags**: None
- **Pattern**: Local only, no peer support

##### `app logs <app>`
- **Args**: `<app>` - Required app identifier
- **Flags**:
  - `-f` - Follow logs (tail)
- **Pattern**: Local by default, remote via `@peer` prefix

##### `app install <url>`
- **Args**: `<url>` - Git repository URL
- **Flags**: None
- **Pattern**: Local by default, remote via `@peer` prefix

##### `app create <name>`
- **Args**: `<name>` - App name
- **Flags**:
  - `--template <type>` - Template type (static, vue, vue-api)
- **Pattern**: Local only, creates directory

##### `app remove [identifier]`
- **Args**: `[identifier]` - Optional app identifier
- **Flags**:
  - `--alias <name>` - Reference by alias
  - `--id <app_id>` - Reference by ID
  - `--with-forks` - Delete app and all forks
- **Pattern**: Local by default, remote via `@peer` prefix

##### `app upgrade <app>`
- **Args**: `<app>` - App identifier
- **Flags**: None
- **Pattern**: Local by default, remote via `@peer` prefix

##### `app link <subdomain>`
- **Args**: `<subdomain>` - Subdomain to link
- **Flags**:
  - `--id <app_id>` - REQUIRED app ID
- **Pattern**: Local by default, remote via `@peer` prefix

##### `app unlink <subdomain>`
- **Args**: `<subdomain>` - Subdomain to unlink
- **Flags**: None
- **Pattern**: Local by default, remote via `@peer` prefix

##### `app reserve <subdomain>`
- **Args**: `<subdomain>` - Subdomain to reserve
- **Flags**: None
- **Pattern**: Local by default, remote via `@peer` prefix

##### `app swap <a1> <a2>`
- **Args**: `<a1> <a2>` - Two aliases to swap
- **Flags**: None
- **Pattern**: Local by default, remote via `@peer` prefix

##### `app split <subdomain>`
- **Args**: `<subdomain>` - Subdomain for traffic splitting
- **Flags**:
  - `--ids <list>` - Comma-separated app_id:weight pairs
- **Pattern**: Local by default, remote via `@peer` prefix

##### `app fork`
- **Args**: None
- **Flags**:
  - `--alias <name>` OR `--id <app_id>` - Source app
  - `--as <name>` - New alias name
  - `--no-storage` - Don't clone storage
- **Pattern**: Local by default, remote via `@peer` prefix

##### `app lineage`
- **Args**: None
- **Flags**:
  - `--alias <name>` OR `--id <app_id>` - App to trace
- **Pattern**: Local by default, remote via `@peer` prefix

---

### 2. `fazt peer`

**Purpose**: Peer management (add, list, status, upgrade)

#### Subcommands

##### `peer add <name>`
- **Args**: `<name>` - Required peer name
- **Flags**:
  - `--url <url>` - REQUIRED peer URL
  - `--token <token>` - REQUIRED auth token
- **Pattern**: Name as positional, connection via flags

##### `peer list`
- **Args**: None
- **Flags**:
  - `--format <format>` - Output format (markdown, json)
- **Pattern**: Simple list, no args

##### `peer remove <name>`
- **Args**: `<name>` - Required peer name
- **Flags**: None
- **Pattern**: Name as positional

##### `peer default <name>`
- **Args**: `<name>` - Required peer name
- **Flags**: None
- **Pattern**: Name as positional

##### `peer status [name]`
- **Args**: `[name]` - Optional peer name
- **Flags**: None
- **Pattern**: Name as positional, uses default if omitted

##### `peer upgrade [name]`
- **Args**: `[name]` - Optional peer name
- **Flags**: None
- **Pattern**: Name as positional, checks all if omitted

##### ~~`peer apps [name]`~~ (DEPRECATED)
- **Replacement**: `fazt app list` or `fazt @peer app list`

##### ~~`peer deploy <dir>`~~ (DEPRECATED)
- **Replacement**: `fazt app deploy <dir>` or `fazt @peer app deploy <dir>`

---

### 3. `fazt service`

**Purpose**: System service management (systemd integration)

#### Subcommands

##### `service install`
- **Args**: None
- **Flags**:
  - `--domain <domain>` - Server domain
  - `--email <email>` - Email for Let's Encrypt
  - `--https` - Enable HTTPS
- **Pattern**: Local only, requires sudo

##### `service start`
- **Args**: None
- **Flags**: None
- **Pattern**: Controls systemd service

##### `service stop`
- **Args**: None
- **Flags**: None
- **Pattern**: Controls systemd service

##### `service status`
- **Args**: None
- **Flags**: None
- **Pattern**: Queries systemd service

##### `service logs`
- **Args**: None
- **Flags**: None
- **Pattern**: Follows journalctl logs

---

### 4. `fazt server`

**Purpose**: Server control (init, start, config)

#### Subcommands

##### `server init`
- **Args**: None
- **Flags**:
  - `--username <user>` - Admin username
  - `--password <pass>` - Admin password
  - `--domain <domain>` - Server domain
  - `--db <path>` - Database path
- **Pattern**: Local only, first-run setup

##### `server start`
- **Args**: None
- **Flags**:
  - `--domain <domain>` - Override domain
  - `--port <port>` - Override port
  - `--db <path>` - Override database path
- **Pattern**: Local only, manual server start

##### `server status`
- **Args**: None
- **Flags**: None
- **Pattern**: Shows local config and status

##### `server set-credentials`
- **Args**: None
- **Flags**:
  - `--username <user>` - New username
  - `--password <pass>` - New password
- **Pattern**: Local only, updates DB

##### `server set-config`
- **Args**: None
- **Flags**:
  - `--domain <domain>` - Update domain
  - `--port <port>` - Update port
  - `--env <env>` - Update environment
- **Pattern**: Local only, updates DB

##### `server create-key`
- **Args**: None
- **Flags**:
  - `--name <name>` - Key name/description
- **Pattern**: Local only, generates API key

##### `server reset-admin`
- **Args**: None
- **Flags**: None
- **Pattern**: Local only, resets admin UI

---

### 5. `fazt client` (LEGACY)

**Purpose**: Legacy client commands (pre-peer architecture)

#### Subcommands

##### `client set-auth-token`
- **Args**: None
- **Flags**:
  - `--token <token>` - Auth token
- **Pattern**: Local config, deprecated

##### `client deploy`
- **Args**: None
- **Flags**:
  - `--path <path>` - Directory to deploy
  - `--domain <domain>` - Target domain
- **Pattern**: Legacy deploy, use `app deploy` instead

##### `client sites`
- **Args**: None
- **Flags**: None
- **Pattern**: Legacy list, use `app list` instead

##### `client logs`
- **Args**: None
- **Flags**:
  - `--site <name>` - Site name
- **Pattern**: Legacy logs, use `app logs` instead

##### `client delete`
- **Args**: None
- **Flags**:
  - `--site <name>` - Site name
- **Pattern**: Legacy delete, use `app remove` instead

---

### 6. `fazt version`

**Purpose**: Show version info

- **Args**: None
- **Flags**: None
- **Pattern**: Simple info display

---

### 7. `fazt help`

**Purpose**: Show help message

- **Args**: None
- **Flags**: None
- **Pattern**: Simple help display

---

## Special Patterns

### @peer Prefix

```bash
fazt @<peer> <command> [args...]
```

**Supported Commands**:
- `fazt @peer app <subcommand>` - Execute app commands on remote peer
- `fazt @peer server info` - Get remote server info

**Not Supported**:
- `peer` commands (already about peers)
- `service` commands (local systemd only)
- `server` commands (except info)
- `client` commands (legacy)

---

## Recent Command Additions

### `fazt sql` (v0.18.0)

**Purpose**: Execute SQL queries on local or remote databases

##### `sql <query>`
- **Args**: `<query>` - SQL query string
- **Flags**:
  - `--format <format>` - Output format (markdown, json)
- **Pattern**: Local by default, remote via `@peer` prefix

**Examples**:
```bash
fazt sql "SELECT * FROM apps"                 # Query local database
fazt @zyt sql "SELECT * FROM apps"            # Query remote database
fazt sql "SELECT * FROM apps" --format json   # JSON output
```

### `fazt user` (v0.24.7)

**Purpose**: User management - list users, view status, set roles

See detailed documentation in `knowledge-base/cli/user/`

### `fazt alias` (v0.24.7)

**Purpose**: Alias management - list and inspect routing aliases

See detailed documentation in `knowledge-base/cli/alias/`

### `fazt logs` (v0.24.8 - v0.24.13)

**Purpose**: Activity log management - unified logging with weight-based prioritization

**Features**:
- Query logs with extensive filtering (weight, app, user, action, time)
- Statistics and analysis
- Cleanup with dry-run safety
- Export to JSON/CSV
- Permissive URL parsing for alias filters (v0.24.13)
- Flag aliases: `--url` and `--link` work alongside `--alias` (v0.24.13)

See detailed documentation in `knowledge-base/cli/logs/`

---

### `fazt logs`

**Purpose**: Activity log management - query, analyze, and clean up activity logs

#### Subcommands

##### `logs list`
- **Args**: None
- **Flags**:
  - `--min-weight N` - Minimum weight (0-9)
  - `--max-weight N` - Maximum weight (0-9)
  - `--app ID` - Filter by app ID
  - `--alias URL` - Filter by subdomain/alias (permissive URL parsing)
  - `--url URL` - Alias for `--alias`
  - `--link URL` - Alias for `--alias`
  - `--user ID` - Filter by user ID
  - `--actor-type T` - Filter by actor type (user/system/api_key/anonymous)
  - `--type T` - Filter by resource type (app/user/session/kv/doc/page/config)
  - `--resource ID` - Filter by resource ID
  - `--action A` - Filter by action (pageview, deploy, login, etc.)
  - `--result R` - Filter by result (success/failure)
  - `--since TIME` - Show entries since (e.g., '24h', '7d', '2024-01-15')
  - `--until TIME` - Show entries until
  - `--limit N` - Number of entries (default: 20)
  - `--offset N` - Skip first n results
- **Pattern**: Local by default, remote via `@peer` prefix

##### `logs stats`
- **Args**: None
- **Flags**: Same filters as `logs list`
- **Pattern**: Local by default, remote via `@peer` prefix

##### `logs cleanup`
- **Args**: None
- **Flags**: Same filters as `logs list`, plus:
  - `--force` - Actually delete (default: preview only)
- **Pattern**: Local by default, remote via `@peer` prefix
- **Safety**: Always runs dry-run unless `--force` specified

##### `logs export`
- **Args**: None
- **Flags**: Same filters as `logs list`, plus:
  - `-f FORMAT` - Output format: json or csv (default: json)
  - `-o FILE` - Output file (default: stdout)
- **Pattern**: Local by default, remote via `@peer` prefix

**Weight Scale** (0-9):
- 9: Security (API key changes, role changes)
- 8: Auth (login/logout, sessions)
- 7: Config (alias/redirect changes)
- 6: Deployment (app deploy/delete)
- 5: Data Mutation (KV/doc writes)
- 4: User Action (form submissions)
- 3: Navigation (page navigation)
- 2: Analytics (pageviews, clicks)
- 1: System (health checks, server events)
- 0: Debug (timing, cache hits)

**Permissive URL Parsing**: The `--alias`, `--url`, and `--link` flags accept URLs in any format and automatically extract the subdomain:
```bash
fazt logs list --alias tetris                        # Just alias
fazt logs list --url https://tetris.zyt.app/         # Full URL
fazt logs list --link tetris.zyt.app/game            # Domain with path
```

**Examples**:
```bash
fazt logs list                                       # Recent activity
fazt logs list --url https://fun-game.zyt.app/path   # Pageviews for subdomain
fazt logs list --app my-app --since 24h              # App activity (24h)
fazt logs list --min-weight 5                        # Important events only
fazt logs stats --app my-app                         # Statistics
fazt logs cleanup --max-weight 2 --until 7d          # Preview cleanup
fazt logs cleanup --max-weight 2 --until 7d --force  # Execute cleanup
fazt logs export --app my-app -f csv -o logs.csv     # Export to CSV
fazt @zyt logs list --link tetris.zyt.app            # Remote peer
```

---

## Deprecated Commands

### `fazt remote` (Renamed in v0.18.0)
- **Replacement**: `fazt peer`
- **Reason**: Naming consistency and clarity

### `fazt peer apps [name]` (Removed)
- **Replacement**: `fazt app list` or `fazt @peer app list`
- **Reason**: App commands should be under `app` namespace

### `fazt peer deploy <dir>` (Removed)
- **Replacement**: `fazt app deploy <dir>` or `fazt @peer app deploy <dir>`
- **Reason**: App commands should be under `app` namespace

### App Command Peer Flags (Removed in v0.18.0)
- **`--to`, `--from`, `--on` flags removed from all app commands**
- **Replacement**: Use `@peer` prefix for remote operations
- **Reason**: Simplified, consistent syntax

**Migration examples**:
```bash
# OLD (removed)
fazt app deploy ./app --to zyt
fazt app remove myapp --from zyt
fazt app info myapp --on zyt

# NEW (current)
fazt @zyt app deploy ./app
fazt @zyt app remove myapp
fazt @zyt app info myapp
```

### `fazt client` (Legacy)
- **Replacement**: Use `fazt app` and `fazt peer` commands
- **Reason**: Pre-peer architecture, superseded by modern commands

---

## Design Philosophy

The v0.18.0 CLI follows these principles:

1. **Local-first**: Commands operate locally by default
2. **Explicit remote**: Use `@peer` prefix for remote operations
3. **Consistent syntax**: Same command works locally and remotely
4. **No directional flags**: Removed `--to`, `--from`, `--on` confusion
5. **Visual clarity**: `@peer` is always in the same position
