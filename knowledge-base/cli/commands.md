# Fazt CLI Commands Reference

**Version**: 0.18.0
**Generated**: 2026-02-01

This document catalogs all CLI commands, their arguments, flags, and patterns for consistency analysis.

---

## Command Structure

```
fazt <command> [subcommand] [flags] [arguments]
fazt @<peer> <command> [subcommand] [flags] [arguments]
```

---

## Top-Level Commands

### 1. `fazt app`

**Purpose**: App management (list, deploy, info, remove)

#### Subcommands

##### `app list [peer]`
- **Args**: `[peer]` - Optional peer name
- **Flags**:
  - `--aliases` - Show alias list instead of apps
- **Pattern**: Peer as positional arg OR via flag
- **Remote**: `fazt @peer app list`

##### `app info [identifier]`
- **Args**: `[identifier]` - Optional app identifier
- **Flags**:
  - `--alias <name>` - Reference by alias
  - `--id <app_id>` - Reference by ID
  - `--on <peer>` - Target peer
- **Pattern**: Either positional OR flags, peer via `--on`

##### `app deploy <dir>`
- **Args**: `<dir>` - Required directory path
- **Flags**:
  - `--to <peer>` - Target peer (REQUIRED)
  - `--name <name>` - Optional app name
- **Pattern**: Source as positional, destination via flag

##### `app validate <dir>`
- **Args**: `<dir>` - Required directory path
- **Flags**: None
- **Pattern**: Local only, no peer support

##### `app logs <app>`
- **Args**: `<app>` - Required app identifier
- **Flags**:
  - `-f` - Follow logs (tail)
  - `--on <peer>` - Target peer
- **Pattern**: App as positional, peer via flag

##### `app install <url>`
- **Args**: `<url>` - Git repository URL
- **Flags**:
  - `--to <peer>` - Target peer
- **Pattern**: Source as positional, destination via flag

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
  - `--on <peer>` - Target peer (implied context?)
  - `--from <peer>` - Source peer
  - `--with-forks` - Delete app and all forks
- **Pattern**: Multiple peer flags (`--on`, `--from`)

##### `app upgrade <app>`
- **Args**: `<app>` - App identifier
- **Flags**:
  - `--on <peer>` - Target peer
- **Pattern**: App as positional, peer via flag

##### `app link <subdomain>`
- **Args**: `<subdomain>` - Subdomain to link
- **Flags**:
  - `--id <app_id>` - REQUIRED app ID
  - `--to <peer>` - Target peer
- **Pattern**: Alias as positional, app via flag

##### `app unlink <subdomain>`
- **Args**: `<subdomain>` - Subdomain to unlink
- **Flags**:
  - `--from <peer>` - Source peer
- **Pattern**: Alias as positional, peer via `--from`

##### `app reserve <subdomain>`
- **Args**: `<subdomain>` - Subdomain to reserve
- **Flags**:
  - `--on <peer>` - Target peer
- **Pattern**: Alias as positional, peer via `--on`

##### `app swap <a1> <a2>`
- **Args**: `<a1> <a2>` - Two aliases to swap
- **Flags**:
  - `--on <peer>` - Target peer
- **Pattern**: Two aliases as positional, peer via flag

##### `app split <subdomain>`
- **Args**: `<subdomain>` - Subdomain for traffic splitting
- **Flags**:
  - `--ids <list>` - Comma-separated app_id:weight pairs
  - `--on <peer>` - Target peer
- **Pattern**: Alias as positional, targets via flag

##### `app fork`
- **Args**: None
- **Flags**:
  - `--alias <name>` OR `--id <app_id>` - Source app
  - `--as <name>` - New alias name
  - `--to <peer>` - Target peer
  - `--no-storage` - Don't clone storage
- **Pattern**: All flags, no positional args

##### `app lineage`
- **Args**: None
- **Flags**:
  - `--alias <name>` OR `--id <app_id>` - App to trace
  - `--on <peer>` - Target peer
- **Pattern**: All flags, no positional args

---

### 2. `fazt remote`

**Purpose**: Peer management (add, list, status, upgrade)

#### Subcommands

##### `remote add <name>`
- **Args**: `<name>` - Required peer name
- **Flags**:
  - `--url <url>` - REQUIRED peer URL
  - `--token <token>` - REQUIRED auth token
- **Pattern**: Name as positional, connection via flags

##### `remote list`
- **Args**: None
- **Flags**: None
- **Pattern**: Simple list, no args

##### `remote remove <name>`
- **Args**: `<name>` - Required peer name
- **Flags**: None
- **Pattern**: Name as positional

##### `remote default <name>`
- **Args**: `<name>` - Required peer name
- **Flags**: None
- **Pattern**: Name as positional

##### `remote status [name]`
- **Args**: `[name]` - Optional peer name
- **Flags**: None
- **Pattern**: Name as positional, uses default if omitted

##### `remote upgrade [name]`
- **Args**: `[name]` - Optional peer name
- **Flags**: None
- **Pattern**: Name as positional, checks all if omitted

##### ~~`remote apps [name]`~~ (DEPRECATED)
- **Replacement**: `fazt app list [peer]`

##### ~~`remote deploy <dir>`~~ (DEPRECATED)
- **Replacement**: `fazt app deploy <dir> --to <peer>`

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
- `remote` commands (already about peers)
- `service` commands (local systemd only)
- `server` commands (except info)
- `client` commands (legacy)

---

## Peer Flag Patterns

Different commands use different flag names for peers:

| Command | Source Flag | Destination Flag | Context Flag |
|---------|-------------|------------------|--------------|
| `app deploy` | - | `--to` | - |
| `app list` | - | - | positional |
| `app info` | - | - | `--on` |
| `app remove` | `--from` | - | `--on` |
| `app link` | - | `--to` | - |
| `app unlink` | `--from` | - | - |
| `app reserve` | - | - | `--on` |
| `app logs` | - | - | `--on` |
| `app fork` | - | `--to` | - |
| `app lineage` | - | - | `--on` |

**Inconsistencies**:
- Mix of `--to`, `--from`, `--on` for peer targeting
- Some commands accept peer as positional arg
- Some commands use `--on` for context
- `app remove` has both `--on` and `--from`

---

## Deprecated Commands

### `fazt servers` (Removed in v0.18.0)
- **Replacement**: `fazt remote`
- **Reason**: Naming consistency (server vs remote)

### `fazt remote apps [name]` (Deprecated)
- **Replacement**: `fazt app list [peer]`
- **Reason**: App commands should be under `app` namespace

### `fazt remote deploy <dir>` (Deprecated)
- **Replacement**: `fazt app deploy <dir> --to <peer>`
- **Reason**: App commands should be under `app` namespace

### `fazt client` (Legacy)
- **Replacement**: Use `fazt app` and `fazt remote` commands
- **Reason**: Pre-peer architecture, superseded by modern commands

---

## Analysis Questions

1. **Peer Flag Consistency**: Should we standardize on one flag pattern?
   - `--to` for destinations (deploy, link, fork)
   - `--from` for sources (unlink, remove?)
   - `--on` for context/target (info, logs, status)
   - Or use positional peer args everywhere?

2. **Default Peer**: Should local operations omit peer flags?
   - `fazt app deploy ./my-app` → defaults to `local`?
   - `fazt app list` → defaults to `local`?

3. **Positional vs Flags**: Inconsistent patterns
   - `app list [peer]` - peer as positional
   - `app info [identifier]` - identifier as positional but also `--alias`/`--id`
   - `app fork` - everything as flags

4. **Remote Execution**: `@peer` prefix vs flags
   - When to use `fazt @zyt app list` vs `fazt app list zyt`?
   - Both work, but which is preferred?

5. **Legacy Cleanup**: Should `client` commands be fully removed?

6. **Verb Consistency**:
   - `list` vs `info` (both query)
   - `remove` vs `delete` vs `unlink`
   - `add` vs `create` vs `install`

---

## POLS (Principle of Least Surprise) Issues

1. **Deploy requires --to**: Even for local deploys?
2. **Multiple ways to specify peer**: Positional, `--to`, `--on`, `--from`, `@peer` prefix
3. **Identifier ambiguity**: `app info myapp` vs `app info --alias myapp` vs `app info --id app_123`
4. **Inconsistent defaults**: Some commands default to local, others require explicit peer
5. **Flag naming**: Why `--on` vs `--to` vs `--from`?
6. **Remote prefix syntax**: `@peer` is elegant but discoverable?
