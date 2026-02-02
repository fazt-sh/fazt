# fazt app - App Management

**Updated**: 2026-02-02 (added app files command)

Deploy, manage, and monitor apps on fazt instances.

## Core Commands

### fazt app list

List apps on a peer.

```bash
fazt app list                   # Local apps
fazt @zyt app list              # Remote apps
fazt app list --aliases         # Include alias mappings
fazt @zyt app list --aliases    # Remote with aliases
```

### fazt app info

Show app details.

```bash
fazt app info --alias <name>        # Local app
fazt app info --id <app_id>         # Local app by ID
fazt @zyt app info --alias <name>   # Remote app
fazt @zyt app info --id <app_id>    # Remote app by ID
```

### fazt app files

List all files in a deployed app with sizes and timestamps.

```bash
fazt app files <app>                # Local app (by alias or ID)
fazt app files --alias <name>       # Local app by alias
fazt app files --id <app_id>        # Local app by ID
fazt @zyt app files <app>           # Remote app
fazt @zyt app files --id <app_id>   # Remote app by ID
fazt app files <app> --format json  # JSON output
```

**Output:**
```
# Files in admin-ui

| Path | Size | Modified |
|---|---|---|
| assets/index-ClxHNpIk.js | 82.4 KB | 2026-02-01T14:14:34Z |
| favicon.png | 39.8 KB | 2026-02-01T14:14:34Z |
| index.html | 41.3 KB | 2026-02-01T14:14:34Z |
| manifest.json | 25 B | 2026-02-01T14:14:34Z |

10 files
```

**Use cases:**
- Verify files deployed correctly
- Check file sizes and timestamps
- Debug deployment issues
- Audit deployed content

### fazt app deploy

Deploy a directory to a peer. **Builds automatically** if `package.json` has
a build script.

```bash
fazt app deploy <directory> [--no-build] [--spa] [--include-private] [--name <name>]
fazt @peer app deploy <directory> [--no-build] [--spa] [--include-private] [--name <name>]
```

**How it works:**
1. Detects if `package.json` has a `"build"` script
2. Finds package manager (bun/pnpm/npm/yarn) and runs build
3. Deploys output (`dist/`, `build/`, `out/`, or `.output/`)
4. If no build script, deploys directory as-is

**Flags:**
| Flag | Description |
|------|-------------|
| `--name <name>` | Override app name |
| `--no-build` | Skip build step, deploy directory as-is |
| `--spa` | Enable SPA routing (clean URLs, serves index.html for unknown routes) |
| `--include-private` | Include gitignored `private/` directory in deployment |

**Examples:**
```bash
# Standard deploy (builds automatically)
fazt app deploy ./my-app              # Local
fazt @zyt app deploy ./my-app         # Remote

# Skip build (for pre-built or static sites)
fazt app deploy ./my-app --no-build

# SPA routing for clean URLs (/dashboard instead of /#/dashboard)
fazt @zyt app deploy ./my-app --spa

# Include gitignored private/ directory
fazt @zyt app deploy ./my-app --include-private

# Custom app name
fazt @zyt app deploy ./my-app --name blog
```

See [deployment.md](deployment.md) and [hosting-quirks.md](hosting-quirks.md)
for routing details.

### fazt app validate

Check app structure before deployment.

```bash
fazt app validate <directory>
```

Validates:
- manifest.json exists and is valid
- Required files present
- No obvious issues

### fazt app logs

View serverless execution logs.

```bash
fazt app logs <app>           # Local logs
fazt @zyt app logs <app>      # Remote logs
fazt app logs <app> -f        # Follow (tail)
fazt @zyt app logs <app> -f   # Follow remote logs
```

### fazt app remove

Remove an app.

```bash
fazt app remove --alias <name>               # Local
fazt app remove --id <app_id>                # Local by ID
fazt @zyt app remove --alias <name>          # Remote
fazt app remove --alias <name> --with-forks  # Remove app and forks
```

## App Creation

### fazt app create

Create new app from template.

```bash
fazt app create <name> --template <type>
```

Templates:
- `static` - Basic HTML/CSS/JS
- `vue` - Vue 3 with Vite
- `vue-api` - Vue 3 + serverless API

### fazt app install

Install app from git repository.

```bash
fazt app install <github-url>         # Local
fazt @zyt app install <github-url>    # Remote
```

### fazt app upgrade

Upgrade git-sourced app to latest.

```bash
fazt app upgrade <app>         # Local
fazt @zyt app upgrade <app>    # Remote
```

## Alias Management

Apps have IDs (`app_abc123`) and aliases (subdomains like `tetris`).

### fazt app link

Link a subdomain to an app.

```bash
fazt app link <subdomain> --id <app_id>         # Local
fazt @zyt app link <subdomain> --id <app_id>    # Remote
```

### fazt app unlink

Remove an alias.

```bash
fazt app unlink <subdomain>         # Local
fazt @zyt app unlink <subdomain>    # Remote
```

### fazt app reserve

Reserve/block a subdomain (prevents use).

```bash
fazt app reserve <subdomain>         # Local
fazt @zyt app reserve <subdomain>    # Remote
```

### fazt app swap

Atomically swap two aliases (blue-green deployment).

```bash
fazt app swap <alias1> <alias2>         # Local
fazt @zyt app swap <alias1> <alias2>    # Remote
```

**Example - Zero-downtime deployment:**
```bash
# 1. Fork current app
fazt @zyt app fork --alias tetris --as tetris-v2

# 2. Deploy update to fork (builds automatically)
fazt @zyt app deploy ./tetris-updated --name tetris-v2

# 3. Test tetris-v2.<domain>

# 4. Swap aliases (instant cutover)
fazt @zyt app swap tetris tetris-v2
```

### fazt app split

Configure traffic splitting between app versions.

```bash
fazt app split <alias> --ids <app1>:<percent>,<app2>:<percent>         # Local
fazt @zyt app split <alias> --ids <app1>:<percent>,<app2>:<percent>    # Remote
```

**Example - Canary deployment:**
```bash
fazt @zyt app split tetris --ids app_old:90,app_new:10
```

## Lineage (Forking)

### fazt app fork

Create a copy of an app (optionally without storage).

```bash
fazt app fork --alias <source> --as <new-alias>                # Local
fazt @zyt app fork --alias <source> --as <new-alias>           # Remote
fazt app fork --alias <source> --as <new-alias> --no-storage   # Without storage
```

### fazt app lineage

Show fork tree for an app.

```bash
fazt app lineage --id <app_id>         # Local
fazt @zyt app lineage --id <app_id>    # Remote
```

## Reference Flags

| Flag | Description |
|------|-------------|
| `--alias <name>` | Reference app by subdomain |
| `--id <app_id>` | Reference app by ID |
| `--name <name>` | Override app name (deploy only) |
| `--with-forks` | Include forked apps in operation |
| `--spa` | Enable SPA routing (deploy only) |
| `--no-build` | Skip automatic build (deploy only) |
| `--include-private` | Include gitignored private/ (deploy only) |
| `--as <name>` | New alias name (fork only) |
| `--no-storage` | Don't clone storage (fork only) |

## Removed Flags (v0.18.0)

These flags have been removed. Use `@peer` prefix instead:

| Removed Flag | Replacement |
|--------------|-------------|
| `--to <peer>` | `fazt @peer app deploy ...` |
| `--from <peer>` | `fazt @peer app remove ...` |
| `--on <peer>` | `fazt @peer app info ...` |
