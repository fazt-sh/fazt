---
name: fazt-app
description: Build and deploy polished Vue+API apps to fazt instances. Creates PWA-ready apps with advanced storage, session management, authentication, and production-quality UX. Use when building new fazt apps.
context: fork
---

# /fazt-app - Build Fazt Apps

Build and deploy polished, PWA-ready apps to fazt instances with Claude.

## Documentation Index

### Fazt Platform

- **[fazt/overview.md](fazt/overview.md)** - What is fazt, capabilities, architecture
- **[fazt/deployment.md](fazt/deployment.md)** - Local vs remote, auto-build deploy, versioning
- **[fazt/hosting-quirks.md](fazt/hosting-quirks.md)** - Static files vs API, hash routing, serverless limits
- **[fazt/cli-app.md](fazt/cli-app.md)** - `fazt app` commands (deploy, list, fork, swap)
- **[fazt/cli-remote.md](fazt/cli-remote.md)** - `fazt remote` commands (peers, status)
- **[fazt/cli-auth.md](fazt/cli-auth.md)** - `fazt auth` commands (providers, users)
- **[fazt/cli-server.md](fazt/cli-server.md)** - `fazt server` commands (init, start)
- **[fazt/mock-oauth-spec.md](fazt/mock-oauth-spec.md)** - Proposed mock OAuth for local dev

### App Development

- **[references/serverless-api.md](references/serverless-api.md)** - fazt.storage.*, fazt.auth.* APIs
- **[references/frontend-patterns.md](references/frontend-patterns.md)** - Vue setup, session, settings
- **[references/design-system.md](references/design-system.md)** - Colors, typography, components
- **[references/auth-integration.md](references/auth-integration.md)** - OAuth, local testing strategies

### Patterns & Examples

- **[patterns/layout.md](patterns/layout.md)** - Fixed-height layouts, responsive
- **[patterns/ui-patterns.md](patterns/ui-patterns.md)** - Layout shift prevention, click-outside, mobile navbar
- **[patterns/modals.md](patterns/modals.md)** - Modal with click-outside close
- **[patterns/testing.md](patterns/testing.md)** - Testing data flow and state
- **[patterns/google-oauth.md](patterns/google-oauth.md)** - Google OAuth setup guide
- **[examples/cashflow.md](examples/cashflow.md)** - Full reference app

---

## Prerequisites

### 1. Verify fazt Binary

```bash
fazt version
```

If not installed, the fazt binary must be built and added to PATH.

### 2. Check Configured Peers

```bash
fazt remote list
```

You should see at least one peer configured. Common setup:
- `local` - Local dev server (for testing)
- A remote peer - Production server with real domain

### 3. Check OAuth Status (If Auth Needed)

```bash
fazt @<remote-peer> auth providers
```

If `google enabled` appears, OAuth is ready. No setup needed.

---

## Build Philosophy: Build-Free, but Buildable

**Always default to this paradigm** unless impossible.

Fazt apps are **static-hostable** - the final output is just HTML, CSS, and JS
files. No Node.js runtime, no SSR server. The app runs entirely in the browser
plus fazt's serverless API.

**If Vite/npm is available**, use it during development:
- Hot Module Replacement (instant feedback)
- Better error messages and stack traces
- Optimized, minified production bundles
- TypeScript support if needed

**The output remains static-hostable:**
```bash
fazt app deploy ./my-app --to <peer>  # Builds and deploys automatically
```

Fazt detects `package.json` with a build script, runs the build, and deploys
the output (`dist/`, `build/`, etc.).

---

## Workflow

### Step 1: Understand Requirements

Parse the user's description to determine:
- App name and purpose
- Does it need authentication?
- What data needs to be stored?

**Evaluate auth need from description:**

| Description implies... | Auth needed? |
|------------------------|--------------|
| "personal", "my data", "private" | Yes |
| "user accounts", "login", "profile" | Yes |
| "multi-user", "team", "collaboration with accounts" | Yes |
| "shareable", "public link", "anonymous" | No (use sessions) |
| "tool", "calculator", "converter" | No |
| Unclear | **Ask user** |

**If unclear, ask:**
- "Does this app need user accounts/login?"
- "Should data be private per-user, or shareable via links?"

| Answer | Approach |
|--------|----------|
| No login needed | URL sessions (`?s=word-word-word`) |
| User accounts needed | Fazt auth |
| Both | Combine auth + sessions |

**If auth needed, check remote peer:**
```bash
fazt @<remote-peer> auth providers
```
If Google is already enabled, proceed. If not, ask user if they want to set it up.

### Step 2: Scaffold

**CRITICAL: Apps MUST be created in `servers/<peer>/` directory!**

```bash
mkdir -p servers/<peer>/<name>
cd servers/<peer>/<name>
npm init -y
npm install vue
npm install -D vite @vitejs/plugin-vue tailwindcss autoprefixer postcss
```

The `servers/` directory is gitignored - apps are instance-specific.

**Copy standard assets:**
```bash
cp ~/.claude/fazt-assets/favicon.png .
cp ~/.claude/fazt-assets/apple-touch-icon.png .
```

### Step 3: Build the App

Create the following structure:

```
<name>/
├── index.html           # PWA meta, imports
├── vite.config.js
├── package.json
├── tailwind.config.js
├── postcss.config.js
├── favicon.png
├── apple-touch-icon.png
├── scripts/
│   └── version.js       # Version generator (run at build)
├── src/
│   ├── main.js          # Vue app init
│   ├── App.vue          # Root component
│   ├── pages/           # Page components
│   ├── components/      # UI components
│   └── lib/             # Utilities (api, session, settings)
├── api/
│   └── main.js          # Serverless API
├── version.json         # Generated build metadata
└── dist/                # Built output
```

**Versioning**: Include `scripts/version.js` and configure build to generate
`version.json`. See [deployment.md](fazt/deployment.md#app-versioning) for setup.

Reference the documentation files for patterns and code.

### Step 4: Test Locally

```bash
# Development with HMR
npm run dev

# Test serverless API on local fazt (builds automatically)
fazt app deploy . --to local
```

Access at: `http://<name>.192.168.64.3.nip.io:8080`

**Debug endpoints (local only):**
- `/_fazt/info` - App metadata
- `/_fazt/storage` - Storage contents
- `/_fazt/logs` - Execution logs
- `/_fazt/errors` - Error logs

**STOP** - Present URL to user and wait for approval.

### Step 5: Deploy to Production

After user approval (builds automatically):

```bash
fazt app deploy . --to <remote-peer>
```

Report production URL: `https://<name>.<domain>`

---

## Key Reminders

### Deploy from Project Root

Fazt builds automatically when `package.json` has a build script:

```bash
# CORRECT - point at project root, fazt handles build
fazt app deploy ./my-app --to <remote-peer>

# Also valid - skip build for pre-built or static sites
fazt app deploy ./my-app --to <peer> --no-build
```

### OAuth Requires Remote

Real OAuth only works on remote peers with real domains (HTTPS required).

**Critical**: OAuth flows through the root domain. When redirecting to login,
use **absolute URLs** (including origin), not relative paths:

```javascript
// CORRECT - returns to myapp.<domain> after OAuth
login(window.location.href)

// WRONG - returns to root domain (loses subdomain)
login(window.location.pathname)
```

**Local testing options:**
1. Mock OAuth provider (proposed fazt feature) - same flow, zero code changes
2. Test non-auth features locally, auth on remote
3. Mock user header workaround (see [auth-integration.md](references/auth-integration.md))

**Note**: Remote peer may already have OAuth configured - check first.

### Check Remote OAuth First

```bash
fazt @<remote-peer> auth providers
```

If Google is enabled, proceed with auth. If not, ask user if they want to
set it up (see [patterns/google-oauth.md](patterns/google-oauth.md)).

---

## Quick Reference

### App Identifiers

- **app_id**: System ID (e.g., `app_7f3k9x2m`)
- **alias**: Subdomain (e.g., `tetris` → `tetris.<domain>`)

### Essential Commands

```bash
fazt remote list                              # List peers
fazt @<peer> auth providers                   # Check OAuth status
fazt app deploy ./my-app --to local           # Deploy to local (builds auto)
fazt app deploy ./my-app --to <remote-peer>   # Deploy to prod (builds auto)
fazt app deploy ./my-app --to <peer> --spa    # Deploy with SPA routing (clean URLs)
fazt app logs <app> --on <peer> -f            # View logs
```

### Serverless Globals

```javascript
request          // { method, path, query, body, headers }
respond(data)    // Send response
respond(201, d)  // With status code
fazt.storage.ds  // Document store
fazt.storage.kv  // Key-value store
fazt.storage.s3  // Blob storage
fazt.auth.*      // Authentication APIs
```
