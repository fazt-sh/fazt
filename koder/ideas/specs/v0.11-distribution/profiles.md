# Deployment Profiles

## Summary

Deployment profiles unify project detection, building, and serving into a
single coherent system. `fazt deploy` auto-detects your project type and
does the right thing.

```bash
cd my-vite-app
fazt deploy          # Detects Vite, builds via GitHub Actions, serves static
```

## Why

Users have different project types:
- Plain HTML (no build needed)
- Jekyll blogs (build with jekyll-lite)
- React/Vue SPAs (build with npm, serve static)
- Documentation sites (various formats)

Without profiles, users must manually configure build commands, output
directories, and serve modes. Profiles encode this knowledge.

## Profile Categories

| Category | Build | Runtime | Examples |
|----------|-------|---------|----------|
| **Static** | None | File serve | HTML, Docsify |
| **SSG Internal** | jekyll-lite | File serve | Jekyll, Markdown blogs |
| **SSG External** | npm (external) | File serve | Vite, Docusaurus, Astro |

**Note:** Fazt does not support SSR (Server-Side Rendering) frameworks like
Next.js with `getServerSideProps`. Those require a Node.js runtime. Fazt
serves static files only.

## Available Profiles

### Static (No Build)

| Profile | Detection | Output | Notes |
|---------|-----------|--------|-------|
| `static` | Default fallback | `.` | Serve directory as-is |
| `docsify` | `index.html` with docsify | `.` | Client-side docs |
| `spa` | `index.html` + `assets/` | `.` | Pre-built SPA |

### SSG Internal (Built by Fazt)

| Profile | Detection | Build | Output |
|---------|-----------|-------|--------|
| `jekyll` | `_config.yml` or `_posts/` | jekyll-lite | `_site/` |

### SSG External (Built by Provider)

| Profile | Detection | Build Command | Output |
|---------|-----------|---------------|--------|
| `vite` | `vite.config.*` | `npm run build` | `dist/` |
| `astro` | `astro.config.*` | `npm run build` | `dist/` |
| `docusaurus` | `docusaurus.config.js` | `npm run build` | `build/` |
| `vitepress` | `.vitepress/config.*` | `npm run build` | `.vitepress/dist/` |
| `nextjs-export` | `next.config.*` + `output: 'export'` | `npm run build` | `out/` |
| `hugo` | `hugo.toml` or `config.toml` | `hugo` | `public/` |
| `eleventy` | `.eleventy.js` or `eleventy.config.js` | `npm run build` | `_site/` |
| `gatsby` | `gatsby-config.js` | `npm run build` | `public/` |
| `sveltekit-static` | `svelte.config.js` + `adapter-static` | `npm run build` | `build/` |

## Detection Algorithm

```
1. Check app.json for explicit profile
   → If set, use it

2. Scan for config files (in priority order):
   - vite.config.* → vite
   - next.config.* → check for output:'export' → nextjs-export or UNSUPPORTED
   - astro.config.* → astro
   - docusaurus.config.js → docusaurus
   - .vitepress/config.* → vitepress
   - svelte.config.js → check for adapter-static → sveltekit-static or UNSUPPORTED
   - gatsby-config.js → gatsby
   - .eleventy.js → eleventy
   - hugo.toml or config.toml (with [params]) → hugo
   - _config.yml or _posts/ → jekyll
   - index.html with docsify script → docsify

3. Check for pre-built output:
   - dist/ or build/ or out/ exists with index.html → spa

4. Check for index.html:
   - index.html exists → static

5. Fallback:
   - Serve directory as static
```

## Unsupported Profiles

These frameworks require server-side Node.js runtime, which Fazt doesn't provide:

| Framework | Why Unsupported | Alternative |
|-----------|-----------------|-------------|
| Next.js (SSR) | Needs Node.js for `getServerSideProps` | Use `output: 'export'` for static |
| SvelteKit (SSR) | Needs Node.js adapter | Use `adapter-static` |
| Nuxt (SSR) | Needs Node.js runtime | Use `nuxt generate` for static |
| Remix | Needs Node.js runtime | No static option |

When detected, Fazt shows helpful error:

```
Profile detected: nextjs (SSR mode)

This project uses Next.js with server-side rendering, which requires
a Node.js runtime. Fazt serves static files only.

Options:
1. Add `output: 'export'` to next.config.js for static export
2. Use Vercel, Netlify, or Cloudflare Pages for SSR support

See: https://fazt.sh/docs/profiles#nextjs
```

## Configuration

### Explicit Profile (app.json)

```json
{
  "name": "my-app",
  "profile": "vite",
  "build": {
    "command": "npm run build",
    "output": "dist",
    "env": {
      "VITE_API_URL": "https://api.example.com"
    }
  }
}
```

### Profile Overrides

Override any profile setting:

```json
{
  "name": "my-app",
  "profile": "vite",
  "build": {
    "output": "build"    // Override default dist/ → build/
  }
}
```

### Skip Detection

```json
{
  "name": "my-app",
  "profile": "static",   // Don't try to build, serve as-is
  "root": "public"       // Serve this subdirectory
}
```

## Deploy Flow

```
fazt deploy ./my-project
         │
         ▼
┌─────────────────────────────────────────────────┐
│ 1. DETECT                                       │
│    Read app.json or scan for config files       │
│    Result: profile = "vite"                     │
└─────────────────────┬───────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────┐
│ 2. CHECK BUILD NEEDED                           │
│    Profile says: build with npm                 │
│    Output dir (dist/) exists? No                │
│    Result: build required                       │
└─────────────────────┬───────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────┐
│ 3. BUILD (if needed)                            │
│    Upload source to build provider              │
│    Trigger: npm run build                       │
│    Receive artifact (dist.tar.gz)               │
└─────────────────────┬───────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────┐
│ 4. DEPLOY                                       │
│    Extract artifact to VFS                      │
│    Configure routing (SPA fallback, etc.)       │
│    Provision SSL if needed                      │
└─────────────────────┬───────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────┐
│ 5. DONE                                         │
│    https://my-project.example.com               │
└─────────────────────────────────────────────────┘
```

## CLI

```bash
# Auto-detect and deploy
fazt deploy ./my-project

# Show detected profile without deploying
fazt deploy ./my-project --dry-run

# Force specific profile
fazt deploy ./my-project --profile vite

# Skip build (deploy pre-built or source)
fazt deploy ./my-project --no-build

# Force rebuild even if output exists
fazt deploy ./my-project --rebuild

# Deploy specific subdirectory
fazt deploy ./my-project --root dist
```

## Output

```
$ fazt deploy ./my-vite-app

Analyzing project...
  Profile: vite (detected from vite.config.ts)
  Build: npm run build → dist/
  Provider: GitHub Actions

Building...
  Triggered build: bld_abc123
  [████████████████████████████████] 100%
  Build complete: 34s, 2.1MB

Deploying...
  Uploading to VFS: 847 files
  Configuring SPA routing
  SSL: using existing certificate

Done! https://my-vite-app.example.com
```

## SPA Routing

For SPA profiles (vite, astro, nextjs-export, etc.), Fazt automatically
configures fallback routing:

```
GET /about           → Try /about.html, then /about/index.html
GET /dashboard/users → Try exact path, then /index.html (SPA fallback)
GET /assets/main.js  → Serve exact file (no fallback for assets)
```

This matches behavior users expect from Vercel/Netlify.

## Profile-Specific Features

### Jekyll

- Drafts: `fazt serve --drafts` includes `_drafts/` folder
- Future posts: Posts with future dates hidden by default
- Incremental: Only rebuild changed files (planned)

### Vite/React

- Environment: `VITE_*` env vars passed to build
- Preview: Source maps available in preview mode
- Chunking: Respects Vite's code splitting

### Docusaurus

- Versions: Multi-version docs supported
- Search: Algolia config passed through
- i18n: Multi-language builds supported

## Build Caching

To avoid unnecessary builds:

```
1. Hash: package-lock.json + source files
2. Check: Does cached artifact with this hash exist?
3. If yes: Skip build, use cached artifact
4. If no: Trigger build, cache result
```

Cache invalidation:
- `fazt deploy --rebuild` forces fresh build
- `package-lock.json` change invalidates cache
- Source file changes invalidate cache
- 7-day TTL on cached artifacts

## Error Handling

### Unsupported Framework

```
$ fazt deploy ./my-remix-app

Profile detected: remix

Remix requires a Node.js runtime for server-side rendering.
Fazt serves static files only.

This framework is not supported. Consider:
- Vercel: https://vercel.com
- Cloudflare Pages: https://pages.cloudflare.com
- Fly.io: https://fly.io
```

### Build Failed

```
$ fazt deploy ./my-vite-app

Building...
  Build failed after 12s

Error: npm run build exited with code 1

Logs:
  > vite build
  > error TS2304: Cannot find name 'foo'
  > src/App.tsx:15:3

Fix the error and try again, or deploy with --no-build
```

### Missing Build Provider

```
$ fazt deploy ./my-vite-app

Profile detected: vite (needs external build)

No build provider configured. This project needs npm to build.

Setup options:
  fazt dev config build.github --token <github-pat>
  fazt dev config build.modal --token <modal-key>

Or build locally first:
  npm run build
  fazt deploy ./my-vite-app --no-build
```

## Relationship to Other Specs

- `build.md`: Profiles use build service for external builds
- `static-site.md`: SSG profiles (jekyll, docusaurus, etc.) defined there
- `marketplace.md`: Installed apps can specify profiles

## Open Questions

1. **Monorepo support**: `fazt deploy ./apps/web` with root package.json?
2. **Custom profiles**: Let users define profiles in `~/.config/fazt/profiles/`?
3. **Profile plugins**: Community-contributed detection patterns?
