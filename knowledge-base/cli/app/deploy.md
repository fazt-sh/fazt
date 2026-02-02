---
# Command Definition
command: "app deploy"
version: "0.18.0"
category: "deployment"

# Synopsis
syntax: "fazt [@peer] app deploy <directory> [flags]"
description: "Deploy a local directory to a fazt instance"

# Arguments
arguments:
  - name: "directory"
    type: "path"
    required: true
    description: "Path to the directory to deploy"
    default: null

# Flags
flags:
  - name: "--name"
    short: "-n"
    type: "string"
    default: "directory name"
    description: "App name (overrides directory name and manifest.json name)"
  - name: "--spa"
    type: "bool"
    default: false
    description: "Enable SPA routing for clean URLs (e.g., /dashboard instead of /dashboard.html)"
  - name: "--no-build"
    type: "bool"
    default: false
    description: "Skip automatic build step (npm/pnpm/yarn/bun)"
  - name: "--include-private"
    type: "bool"
    default: false
    description: "Include gitignored private/ directory in deployment"

# Peer Support
peer:
  supported: true
  local: true
  remote: true
  deprecated_flags:
    - name: "--to"
      migration: "Use @peer prefix instead: fazt @peer app deploy ..."

# Examples
examples:
  - title: "Deploy to local fazt"
    command: "fazt app deploy ./my-app"
    description: "Deploy the ./my-app directory to the local fazt instance"
    expects_error: false

  - title: "Deploy to remote peer"
    command: "fazt @zyt app deploy ./my-app"
    description: "Deploy the ./my-app directory to the remote peer named 'zyt'"
    expects_error: false

  - title: "Deploy with custom name"
    command: "fazt @zyt app deploy ./my-app --name blog"
    description: "Deploy with app name 'blog' instead of 'my-app'"
    expects_error: false

  - title: "Deploy SPA application"
    command: "fazt @zyt app deploy ./my-spa --spa"
    description: "Deploy a single-page application with clean URL routing"
    expects_error: false

  - title: "Deploy without building"
    command: "fazt app deploy ./dist --no-build"
    description: "Deploy pre-built files, skip automatic build detection"
    expects_error: false

  - title: "Deploy with private files"
    command: "fazt @zyt app deploy ./my-app --include-private"
    description: "Include gitignored private/ directory (for server-only secrets)"
    expects_error: false

# Related Commands
related:
  - command: "app list"
    description: "List deployed apps"
  - command: "app remove"
    description: "Remove a deployed app"
  - command: "app validate"
    description: "Validate app structure before deploying"
  - command: "app info"
    description: "Show details about a deployed app"
  - command: "app pull"
    description: "Download app files from a peer"

# Common Errors
errors:
  - code: "ENOENT"
    message: "Error: directory 'path' does not exist"
    solution: "Verify the path exists and is accessible. Use absolute or relative path."

  - code: "ENOPEER"
    message: "Error: No local fazt server running."
    solution: |
      Start the local server with 'fazt server start', or target a remote peer:
        fazt @<peer> app deploy ./app

  - code: "ENOMANIFEST"
    message: "Warning: No manifest.json found"
    solution: |
      Create a manifest.json in the app root:
        { "name": "my-app" }
      Or use --name flag to specify the app name.

  - code: "EBUILD"
    message: "Error: build failed"
    solution: |
      Check build output for errors. Options:
        1. Fix the build errors
        2. Use --no-build to skip building
        3. Build locally first, then deploy the output directory

  - code: "EAUTH"
    message: "Error: unauthorized"
    solution: |
      The peer rejected the request. Check:
        1. Token is valid: fazt remote status <peer>
        2. Token has deploy permissions
---

# fazt app deploy

Deploy a local directory to a fazt instance.

## Synopsis

```
fazt app deploy <directory> [--name <name>] [--spa] [--no-build]
fazt @<peer> app deploy <directory> [--name <name>] [--spa] [--no-build]
```

## Description

The `deploy` command uploads a directory to a fazt instance. By default,
it targets the local fazt server. Use the `@peer` prefix to deploy to a
remote peer.

### Build Detection

If the directory contains a `package.json`, fazt automatically detects
and runs the appropriate build tool:

| File | Build Command |
|------|---------------|
| `pnpm-lock.yaml` | `pnpm install && pnpm build` |
| `bun.lockb` | `bun install && bun run build` |
| `yarn.lock` | `yarn install && yarn build` |
| `package-lock.json` | `npm install && npm run build` |

The build output (typically `dist/` or `build/`) is then deployed.
Use `--no-build` to skip this step.

### App Naming

The app name is determined in this order:

1. `--name` flag (highest priority)
2. `name` field in `manifest.json`
3. Directory name (lowest priority)

### SPA Routing

For single-page applications (React, Vue, Svelte, etc.), use the `--spa`
flag to enable clean URL routing. This allows URLs like `/dashboard` to
work without the `.html` extension.

When SPA routing is enabled:
- `/dashboard` serves `index.html` (not 404)
- Assets (`/js/`, `/css/`) are served normally
- The app handles routing client-side

You can also enable SPA routing via `manifest.json`:

```json
{
  "name": "my-spa",
  "spa": true
}
```

### Private Directory

The `private/` directory is special:

- **Gitignored by default**: Won't be committed to git
- **Excluded from deploy**: Won't be uploaded unless `--include-private`
- **Auth-gated on server**: Only accessible to authenticated requests

Use `private/` for:
- Server-side secrets
- API keys
- Configuration that shouldn't be public

## Arguments

**`<directory>`** (required)

Path to the directory to deploy. Can be absolute or relative.

## Flags

**`--name, -n <name>`**

Override the app name. By default, uses the directory name or
`manifest.json` name field.

**`--spa`**

Enable SPA (single-page application) routing. Serves `index.html` for
all paths that don't match a file, allowing client-side routing.

**`--no-build`**

Skip automatic build detection and execution. Deploy files as-is.
Useful when:
- Deploying pre-built assets
- Build is handled externally
- No build step needed

**`--include-private`**

Include the gitignored `private/` directory in the deployment.
By default, gitignored `private/` is excluded even if it exists.

## Examples

### Deploy to local fazt

```bash
fazt app deploy ./my-app
```

Output:
```
Building with npm...
Build: dist (42 files via npm)
Deploying 'dist' to local as 'my-app'...
Zipped 42 files (156 KB)

Deployed: my-app
Files:    42
Size:     156 KB
```

### Deploy to remote peer

```bash
fazt @zyt app deploy ./my-app
```

The `@zyt` prefix targets the peer named "zyt" (configured via
`fazt remote add`).

### Deploy SPA with clean URLs

```bash
fazt @zyt app deploy ./my-spa --spa
```

Output includes:
```
SPA:      enabled (clean URLs)
```

Now `/dashboard` works without `.html` extension.

### Deploy with custom name

```bash
fazt @zyt app deploy ./2024-blog-redesign --name blog
```

Deploys as "blog" instead of "2024-blog-redesign".

### Deploy pre-built files

```bash
# Build externally
npm run build

# Deploy the dist folder directly
fazt app deploy ./dist --no-build
```

### Include private configuration

```bash
# private/ contains secrets.json
fazt @zyt app deploy ./my-app --include-private
```

Output:
```
Including gitignored private/ (3 files)
```

## Workflow Examples

### Blue-Green Deployment

```bash
# Deploy new version as separate app
fazt @zyt app deploy ./my-app --name my-app-v2

# Test the new version at my-app-v2.zyt.app

# When ready, swap aliases
fazt @zyt app swap my-app my-app-v2

# Old app is now at my-app-v2, new is at my-app
```

### Development to Production

```bash
# Deploy to local for testing
fazt app deploy ./my-app

# Test at my-app.localhost:8080

# Deploy to production
fazt @zyt app deploy ./my-app
```

## Common Errors

### No local server running

```
Error: No local fazt server running.

To deploy to a remote peer:
  fazt @<peer> app deploy ./app

Available peers:
  - zyt (https://admin.zyt.app)
  - local (http://192.168.64.3:8080)

To start local server:
  fazt server start
```

### Directory not found

```
Error: directory './my-ap' does not exist
```

Check the path and try again.

### Build failed

```
Error: build failed: npm ERR! Missing script: "build"
```

Either add a build script to `package.json` or use `--no-build`.

### Private directory warning

```
Warning: private/ is gitignored but exists
  Use --include-private to deploy private files
  Skipping private/...
```

This is informational. Use `--include-private` if you need those files.

## Deprecated Patterns

**Old (flag-based):**
```bash
fazt app deploy ./my-app --to zyt
```

**New (@peer-based):**
```bash
fazt @zyt app deploy ./my-app
```

The `--to` flag is deprecated and will be removed in v0.20.0.
It currently works but shows a deprecation notice.

## See Also

- **`fazt app list`** - List deployed apps on a peer
- **`fazt app remove`** - Remove a deployed app
- **`fazt app validate`** - Check app structure before deploying
- **`fazt app info`** - Show details about a deployed app
- **`fazt app pull`** - Download app files from a peer
- **`fazt app install`** - Install app from git repository
