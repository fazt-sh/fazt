# fazt app - App Management

Deploy, manage, and monitor apps on fazt instances.

## Core Commands

### fazt app list

List apps on a peer.

```bash
fazt app list <peer>
fazt app list <peer> --aliases  # Include alias mappings
```

### fazt app info

Show app details.

```bash
fazt app info --alias <name> --on <peer>
fazt app info --id <app_id> --on <peer>
```

### fazt app deploy

Deploy a directory to a peer. **Builds automatically** if `package.json` has
a build script.

```bash
fazt app deploy <directory> [--to <peer>] [--no-build] [--spa] [--include-private]
```

**How it works:**
1. Detects if `package.json` has a `"build"` script
2. Finds package manager (bun/pnpm/npm/yarn) and runs build
3. Deploys output (`dist/`, `build/`, `out/`, or `.output/`)
4. If no build script, deploys directory as-is

**Flags:**
| Flag | Description |
|------|-------------|
| `--no-build` | Skip build step, deploy directory as-is |
| `--spa` | Enable SPA routing (clean URLs, serves index.html for unknown routes) |
| `--include-private` | Include gitignored `private/` directory in deployment |

**Examples:**
```bash
# Standard deploy (builds automatically)
fazt app deploy ./my-app --to local
fazt app deploy ./my-app --to <remote-peer>

# Skip build (for pre-built or static sites)
fazt app deploy ./my-app --to local --no-build

# SPA routing for clean URLs (/dashboard instead of /#/dashboard)
fazt app deploy ./my-app --to <remote-peer> --spa

# Include gitignored private/ directory
fazt app deploy ./my-app --to <remote-peer> --include-private
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
fazt app logs <app> --on <peer>
fazt app logs <app> --on <peer> -f  # Follow (tail)
```

### fazt app remove

Remove an app.

```bash
fazt app remove --alias <name> --from <peer>
fazt app remove --id <app_id> --from <peer>
fazt app remove --alias <name> --from <peer> --with-forks  # Remove app and forks
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
fazt app install <github-url> --to <peer>
```

### fazt app upgrade

Upgrade git-sourced app to latest.

```bash
fazt app upgrade <app> --on <peer>
```

## Alias Management

Apps have IDs (`app_abc123`) and aliases (subdomains like `tetris`).

### fazt app link

Link a subdomain to an app.

```bash
fazt app link <subdomain> --id <app_id> --to <peer>
```

### fazt app unlink

Remove an alias.

```bash
fazt app unlink <subdomain> --from <peer>
```

### fazt app reserve

Reserve/block a subdomain (prevents use).

```bash
fazt app reserve <subdomain> --on <peer>
```

### fazt app swap

Atomically swap two aliases (blue-green deployment).

```bash
fazt app swap <alias1> <alias2> --on <peer>
```

**Example - Zero-downtime deployment:**
```bash
# 1. Fork current app
fazt app fork --alias tetris --as tetris-v2 --to <peer>

# 2. Deploy update to fork (builds automatically)
fazt app deploy ./tetris-updated --to <peer> --alias tetris-v2

# 3. Test tetris-v2.<domain>

# 4. Swap aliases (instant cutover)
fazt app swap tetris tetris-v2 --on <peer>
```

### fazt app split

Configure traffic splitting between app versions.

```bash
fazt app split <alias> --ids <app1>:<percent>,<app2>:<percent> --on <peer>
```

**Example - Canary deployment:**
```bash
fazt app split tetris --ids app_old:90,app_new:10 --on <peer>
```

## Lineage (Forking)

### fazt app fork

Create a copy of an app (optionally without storage).

```bash
fazt app fork --alias <source> --as <new-alias> --to <peer>
fazt app fork --alias <source> --as <new-alias> --to <peer> --no-storage
```

### fazt app lineage

Show fork tree for an app.

```bash
fazt app lineage --id <app_id> --on <peer>
```

## Reference Flags

| Flag | Description |
|------|-------------|
| `--alias <name>` | Reference app by subdomain |
| `--id <app_id>` | Reference app by ID |
| `--to <peer>` | Target peer for deployment |
| `--on <peer>` | Target peer for queries |
| `--from <peer>` | Source peer for removal |
| `--with-forks` | Include forked apps in operation |
| `--spa` | Enable SPA routing (deploy only) |
| `--no-build` | Skip automatic build (deploy only) |
| `--include-private` | Include gitignored private/ (deploy only) |
