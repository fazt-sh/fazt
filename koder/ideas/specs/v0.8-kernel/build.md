# Build - External Build Service Abstraction

## Summary

Build (`dev.build.*`) provides a unified interface to external build services.
Fazt remains pure Go with no Node.js/npm toolchain, but can trigger builds on
external compute and receive artifacts.

This enables React, Next.js, Vite, and similar apps to be deployed without
requiring users to build locally.

## Why

Fazt's philosophy: single binary, pure Go, no CGO. This means no Node.js
runtime for building JavaScript apps.

But users want:
```bash
cd my-react-app
fazt deploy        # Just works, even without local npm
```

Solution: Outsource builds to external services, like we outsource LLM
inference to OpenAI/Anthropic.

```
dev.llm.*    → AI compute (OpenAI, Anthropic)
dev.build.*  → Build compute (GitHub Actions, Modal, Depot)
```

## Flow

```
┌─────────────────────────────────────────────────────────────────┐
│  User: fazt deploy ./my-react-app                               │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│  Fazt Kernel                                                    │
│                                                                 │
│  1. Detect: package.json exists, no dist/ folder                │
│  2. Check: app.json has build config? Or prompt user            │
│  3. Upload: Push source to build provider                       │
│  4. Trigger: Start build job                                    │
│  5. Wait: Poll for completion (or webhook callback)             │
│  6. Download: Fetch built artifacts                             │
│  7. Deploy: Store in VFS, serve                                 │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│  External Build Providers                                       │
│  GitHub Actions │ Modal │ Depot │ Fly Machines                  │
└─────────────────────────────────────────────────────────────────┘
```

## Providers

| Provider | Model | Free Tier | Best For |
|----------|-------|-----------|----------|
| `github` | Workflow dispatch | 2000 mins/month | Users with GitHub repos |
| `modal` | On-demand containers | $30/month credit | Flexible, any runtime |
| `depot` | Optimized builds | 30 builds/month | Fast npm/Docker builds |
| `fly` | On-demand VMs | Pay-per-use | Full control |

### Provider: GitHub Actions (Primary)

Most developers already have GitHub. Zero additional setup for them.

**How it works:**
1. User connects GitHub repo to Fazt
2. Fazt creates/updates `.github/workflows/fazt-build.yml`
3. On deploy, Fazt triggers workflow via GitHub API
4. Workflow builds and POSTs artifacts back to Fazt

**Workflow template:**
```yaml
# .github/workflows/fazt-build.yml (auto-generated)
name: Fazt Build
on:
  workflow_dispatch:
    inputs:
      callback_url:
        description: 'Fazt callback URL'
        required: true
      build_id:
        description: 'Build ID'
        required: true

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'

      - run: npm ci

      - run: npm run build

      - name: Upload to Fazt
        run: |
          tar -czf dist.tar.gz -C dist .
          curl -X POST "${{ inputs.callback_url }}" \
            -H "X-Build-ID: ${{ inputs.build_id }}" \
            -F "artifact=@dist.tar.gz"
```

### Provider: Modal

On-demand containers for users without GitHub or wanting more control.

**How it works:**
1. Fazt uploads source as tarball
2. Modal spins up container with Node.js
3. Runs build commands
4. Returns artifacts
5. Container terminates

### Provider: Depot

Optimized for fast npm builds with intelligent caching.

**How it works:**
1. Fazt triggers build via Depot API
2. Depot pulls from Git or receives tarball
3. Builds with aggressive caching
4. Returns artifacts

## Configuration

### System-level (provider credentials)

```bash
# GitHub (uses existing GitHub App or PAT)
fazt dev config build.github --token ghp_xxx

# Modal
fazt dev config build.modal --token ak-xxx

# Depot
fazt dev config build.depot --token dpt_xxx

# Set default provider
fazt dev config build.default github
```

### App-level (build settings)

```json
// app.json
{
  "name": "my-react-app",
  "build": {
    "provider": "github",
    "repo": "user/my-react-app",
    "command": "npm run build",
    "output": "dist",
    "node": "20",
    "env": {
      "VITE_API_URL": "https://api.example.com"
    }
  }
}
```

### Auto-detection

If no `build` config, Fazt detects common patterns:

| Files Present | Framework | Build Command | Output |
|---------------|-----------|---------------|--------|
| `vite.config.*` | Vite | `npm run build` | `dist` |
| `next.config.*` | Next.js | `npm run build` | `.next` or `out` |
| `astro.config.*` | Astro | `npm run build` | `dist` |
| `svelte.config.*` | SvelteKit | `npm run build` | `build` |
| `package.json` (generic) | Unknown | `npm run build` | `dist` |

## Interface

### JavaScript API

```javascript
// Trigger a build manually
const build = await fazt.dev.build.trigger({
  provider: 'github',              // Optional if default set
  repo: 'user/my-react-app',       // For github provider
  ref: 'main',                     // Branch/tag/commit
  command: 'npm run build',        // Override default
  env: {                           // Build-time env vars
    NODE_ENV: 'production'
  }
});
// Returns: {
//   id: 'bld_abc123',
//   status: 'pending',
//   provider: 'github',
//   startedAt: '2025-01-04T...'
// }

// Check build status
const status = await fazt.dev.build.status('bld_abc123');
// Returns: {
//   id: 'bld_abc123',
//   status: 'running' | 'success' | 'failed',
//   logs: '...',
//   duration: 45,              // seconds
//   artifactSize: 2500000      // bytes
// }

// List recent builds
const builds = await fazt.dev.build.list({ limit: 10 });

// Get build logs
const logs = await fazt.dev.build.logs('bld_abc123');

// Cancel a running build
await fazt.dev.build.cancel('bld_abc123');
```

### CLI

```bash
# Trigger build for current directory
fazt build

# Trigger build for specific app
fazt build --app my-react-app

# Watch build progress
fazt build --watch

# View build logs
fazt build logs bld_abc123

# List recent builds
fazt build list

# Cancel build
fazt build cancel bld_abc123
```

### Deploy with auto-build

```bash
# If source needs building, triggers build first
fazt deploy ./my-react-app

# Force rebuild even if artifacts exist
fazt deploy ./my-react-app --rebuild

# Skip build (deploy source as-is)
fazt deploy ./my-react-app --no-build
```

## Callback Endpoint

Fazt exposes a callback endpoint for build providers:

```
POST /_build/callback
Headers:
  X-Build-ID: bld_abc123
  X-Build-Status: success | failed
Body:
  multipart/form-data with artifact file
```

The endpoint:
1. Validates build ID exists and is pending
2. Extracts artifact tarball
3. Stores files in VFS
4. Marks build complete
5. Triggers deployment if auto-deploy enabled

## Build Caching

To speed up builds, providers cache:

| Provider | Caching |
|----------|---------|
| GitHub Actions | `actions/cache` for node_modules |
| Modal | Layer caching in container |
| Depot | Intelligent npm cache across builds |

Fazt also caches artifacts locally:
- Hash of `package-lock.json` + source files
- Skip rebuild if hash matches cached artifact
- `fazt deploy --rebuild` to force fresh build

## Error Handling

```javascript
try {
  await fazt.dev.build.trigger({ ... });
} catch (e) {
  switch (e.code) {
    case 'BUILD_NOT_CONFIGURED':
      // No provider configured
      console.log('Run: fazt dev config build.github --token xxx');
      break;
    case 'BUILD_TIMEOUT':
      // Build took too long
      console.log('Build exceeded', e.timeout, 'seconds');
      break;
    case 'BUILD_FAILED':
      // Build command failed
      console.log('Build failed:', e.logs);
      break;
    case 'BUILD_QUOTA_EXCEEDED':
      // Provider quota hit
      console.log('Build quota exceeded for', e.provider);
      break;
    case 'ARTIFACT_TOO_LARGE':
      // Output exceeds limit
      console.log('Artifact size', e.size, 'exceeds limit');
      break;
  }
}
```

## Limits

| Limit | Value | Notes |
|-------|-------|-------|
| Build timeout | 10 minutes | Configurable per provider |
| Artifact size | 100MB | Compressed tarball |
| Source size | 50MB | Uploaded to provider |
| Concurrent builds | 3 | Per Fazt instance |

## Security

### Credential Storage

Build provider tokens are:
- Stored encrypted in kernel config
- Never exposed to JS runtime
- Only used by kernel for API calls

### Source Upload

When uploading source to build providers:
- `.env` files are excluded
- `.git` directory is excluded
- Respects `.gitignore` patterns
- Sensitive files in `app.json.exclude` skipped

### Artifact Validation

Downloaded artifacts are:
- Checked against expected build ID
- Scanned for executable files (warning only)
- Size-limited to prevent abuse

## Dashboard UI

```
┌─────────────────────────────────────────────────────────────────┐
│  BUILDS                                                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  BUILD PROVIDERS                                    [+ Add]     │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐            │
│  │ 🐙 GitHub    │ │ ⚡ Modal      │ │ 🏗️ Depot     │            │
│  │              │ │              │ │              │            │
│  │ ✓ Connected  │ │ ○ Not setup  │ │ ○ Not setup  │            │
│  │ 45 builds    │ │              │ │              │            │
│  └──────────────┘ └──────────────┘ └──────────────┘            │
│                                                                 │
│  RECENT BUILDS                                                  │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ App          │ Status  │ Duration │ Provider │ Time     │   │
│  ├─────────────────────────────────────────────────────────┤   │
│  │ my-react-app │ ● success │ 47s     │ GitHub   │ 2m ago  │   │
│  │ docs-site    │ ● success │ 23s     │ GitHub   │ 1h ago  │   │
│  │ dashboard    │ ○ failed  │ 12s     │ GitHub   │ 2h ago  │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  [View All Builds]                                              │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## MCP Tools

When MCP is enabled (v0.12), build operations are exposed:

```
Tools:
  build.trigger       Trigger a build
  build.status        Check build status
  build.logs          Get build logs
  build.cancel        Cancel running build

Resources:
  fazt://builds                   List of recent builds
  fazt://builds/{id}              Build details and logs
  fazt://builds/{id}/artifact     Download build artifact
```

## Example: Full React Deploy Flow

```bash
# First time setup
fazt dev config build.github --token ghp_xxx

# Create React app (user does this)
npm create vite@latest my-app -- --template react
cd my-app

# Deploy (Fazt handles everything)
fazt deploy

# Output:
# Detected: Vite (React)
# No dist/ found, triggering build...
# Build provider: GitHub Actions
# Creating workflow in user/my-app...
# Triggering build...
# Build started: bld_abc123
# [====================================] 100%
# Build complete: 47s, 2.3MB artifact
# Deploying to my-app.example.com...
# Done! https://my-app.example.com
```

## Implementation Notes

- Providers implemented as Go packages under `pkg/kernel/dev/build/`
- Each provider implements `BuildProvider` interface
- Artifact storage reuses existing VFS infrastructure
- Build logs stored in events table for querying

## Dependencies

| Dependency | Purpose |
|------------|---------|
| `google/go-github/v57` | GitHub API client |
| `modal-labs/modal-client-go` | Modal API (if exists, else HTTP) |

## Binary Size Impact

| Component | Size |
|-----------|------|
| go-github | ~300KB |
| HTTP clients | minimal |

Total: ~300-400KB additional.

## Open Questions

1. **Monorepo support**: How to handle `apps/web` subfolder builds?
2. **Private npm packages**: Pass npm tokens to build environment?
3. **Build preview**: Deploy build output before promoting to production?
