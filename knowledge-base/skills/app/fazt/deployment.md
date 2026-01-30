# Deployment Guide

How to deploy apps to fazt instances.

## App Versioning

Every fazt app should include version metadata for tracking deployments.

### Setup

1. **Add version script** (`scripts/version.js`):
   - Copy from templates or create per [templates/scripts/version.js](../templates/scripts/version.js)

2. **Update package.json**:
```json
{
  "version": "1.0.0",
  "scripts": {
    "version": "node scripts/version.js",
    "build": "node scripts/version.js && vite build && cp version.json dist/"
  }
}
```

3. **Bump version** before significant deploys:
```bash
npm version patch  # 1.0.0 → 1.0.1
npm version minor  # 1.0.0 → 1.1.0
npm version major  # 1.0.0 → 2.0.0
```

### Output

`version.json` is generated with:
```json
{
  "version": "1.0.0",
  "build": "abc1234",
  "branch": "main",
  "timestamp": "2026-01-29T12:00:00.000Z"
}
```

### Checking Deployed Version

```bash
curl https://<app>.<domain>/version.json
```

### Displaying in UI

Fetch `/version.json` and display in footer or settings:
```javascript
const res = await fetch('/version.json')
const { version, build } = await res.json()
// Display: v1.0.0 (abc1234)
```

---

## Built-in Build + Deploy

`fazt app deploy` handles building automatically. Point it at the **project
root**, not `dist/`.

```bash
fazt app deploy ./my-app --to local
fazt app deploy ./my-app --to <remote-peer>
```

### What Happens

1. Checks if `package.json` has a `"build"` script
2. Detects package manager (bun → pnpm → npm → yarn)
3. Installs dependencies if `node_modules` missing
4. Runs `<pm> run build`
5. Finds output dir (`dist/`, `build/`, `out/`, `.output/`)
6. Deploys that output to the peer

### Decision Tree

| Scenario | What fazt does |
|----------|----------------|
| Has `package.json` + build script | Build with detected PM, deploy output |
| Has `package.json` but no PM | Uses existing `dist/` or errors |
| No `package.json` | Deploys source directory directly |
| `--no-build` flag | Skips build, deploys directory as-is |

### Examples

```bash
# Vue/React app with build script - builds and deploys dist/
fazt app deploy ./my-vue-app --to local

# Simple static site (no package.json) - deploys as-is
fazt app deploy ./static-site --to local

# Skip build (deploy source or pre-built dist/)
fazt app deploy ./my-app --to local --no-build
fazt app deploy ./my-app/dist --to local --no-build
```

## Local vs Remote Deployment

| Target | URL Pattern | Use Case |
|--------|-------------|----------|
| Local | `http://<app>.<local-ip>.nip.io:<port>` | Development, testing |
| Remote | `https://<app>.<domain>` | Production |

### Local Deployment

```bash
fazt app deploy ./my-app --to local
```

**Advantages:**
- Fast iteration
- No internet required
- Debug endpoints available (`/_fazt/*`)

**Limitations:**
- No HTTPS (OAuth won't work)
- Not accessible externally
- nip.io domain (wildcard DNS)

### Remote Deployment

```bash
fazt app deploy ./my-app --to <remote-peer>
```

**Advantages:**
- Real domain with HTTPS
- OAuth authentication works
- Publicly accessible

### Deployment Workflow

```bash
# 1. Development with HMR
npm run dev                          # Vite dev server

# 2. Test on local fazt (builds automatically)
fazt app deploy ./my-app --to local  # Test full stack including API

# 3. Production (builds automatically)
fazt app deploy ./my-app --to <remote-peer>
```

## OAuth Requires Remote Deployment

**Real OAuth authentication only works on remote peers with real domains.**

Why:
- OAuth providers require HTTPS callback URLs
- Callback URL must match registered domain exactly
- `localhost` and IP addresses aren't valid OAuth callbacks

### Mock OAuth Provider (Proposed)

The ideal solution is a built-in `dev` provider that simulates OAuth locally:
- Same auth flow, same cookies, same `fazt.auth.getUser()` response
- Just shows a form instead of redirecting to Google
- **Zero code changes** when deploying to remote

See [auth-integration.md](../references/auth-integration.md) for details.

### Current Workarounds

Until mock OAuth is available, use these strategies:

**1. Test non-auth features locally, auth on remote:**
```bash
# Test UI and API logic locally
fazt app deploy ./my-app --to local

# Test auth flow on remote (builds automatically)
fazt app deploy ./my-app --to <remote-peer>
```

**2. Mock auth in serverless API:**
```javascript
// api/main.js - Development mock
var user = fazt.auth.getUser()

// If no user and running locally, use mock
if (!user && request.headers['x-mock-user']) {
  user = {
    id: 'mock_user_1',
    email: 'dev@example.com',
    name: 'Dev User',
    role: 'user'
  }
}
```

**3. Check if remote already has OAuth:**
```bash
fazt @<remote-peer> auth providers
```

If Google is already enabled, you can test auth immediately on remote.

## Private Directory

The `private/` directory stores server-only files with dual access:

| Access | Use Case | Behavior |
|--------|----------|----------|
| HTTP `GET /private/*` | Serve to users | Requires auth (401 if not logged in) |
| Serverless `fazt.private.*` | Process in code | Direct read for logic |

### Use Cases

- **Large files** (video, images): Stream to authenticated users via HTTP
- **Data files** (JSON, config): Process in serverless API
- **Seed data**: Bundle initial data with app

### Project Structure

```
my-app/
├── src/                 # Vue/React source
├── api/
│   └── main.js          # Serverless handler
├── private/             # Server-only files
│   ├── config.json      # App configuration
│   ├── seed-data.json   # Initial data
│   └── videos/          # Protected media
└── public/              # Static assets
```

### Deployment

By default, if `private/` is gitignored, it won't be deployed:

```bash
# Warning shown when private/ exists but is gitignored
$ fazt app deploy ./my-app --to zyt
Warning: private/ is gitignored but exists
  Use --include-private to deploy private files
  Skipping private/...

# Explicitly include gitignored private/
$ fazt app deploy ./my-app --to zyt --include-private
Including gitignored private/ (5 files)
```

### Accessing Private Files

**Serverless API** (for data processing):
```javascript
// api/main.js
var config = fazt.private.readJSON('config.json')
var users = fazt.private.readJSON('seed-data.json')

if (fazt.private.exists('feature-flags.json')) {
  var flags = fazt.private.readJSON('feature-flags.json')
}
```

**HTTP** (for authenticated users):
```
GET /private/video.mp4
→ 401 Unauthorized (not logged in)
→ Streams video (logged in)
```

See [serverless-api.md](../references/serverless-api.md) for full `fazt.private.*` API.

## Debug Endpoints (Local Only)

When deployed locally, these endpoints help debug:

| Endpoint | Description |
|----------|-------------|
| `/_fazt/info` | App metadata, app_id |
| `/_fazt/storage` | View storage contents |
| `/_fazt/logs` | Recent serverless execution logs |
| `/_fazt/errors` | Error logs |

Example:
```bash
curl http://my-app.<local-ip>.nip.io:<port>/_fazt/info
```

## Deployment Checklist

### Local (Development)

- [ ] `fazt remote list` shows `local` peer
- [ ] Local server running (`systemctl --user status fazt-local`)
- [ ] Deploy source: `fazt app deploy ./my-app --to local`
- [ ] Access via: `http://<app>.192.168.64.3.nip.io:8080`
- [ ] Check logs if issues: `fazt app logs <app> --on local`

### Remote (Production)

- [ ] `fazt remote list` shows remote peer
- [ ] Deploy: `fazt app deploy ./my-app --to <peer>` (builds automatically)
- [ ] Access via: `https://<app>.<domain>`
- [ ] If using auth, verify OAuth provider enabled

## Common Issues

### "App not found" after deploy

The app exists but alias isn't linked. Check:
```bash
fazt app list <peer> --aliases
```

### OAuth redirect fails

- Verify callback URL matches exactly in Google Console
- Ensure HTTPS (won't work on local)
- Check provider is enabled: `fazt @<peer> auth providers`

### Serverless API returns 500

Check logs:
```bash
fazt app logs <app> --on <peer> -f
```

Or locally:
```bash
curl http://<app>.192.168.64.3.nip.io:8080/_fazt/errors
```
