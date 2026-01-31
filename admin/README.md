# Fazt Admin UI

**Version**: 0.17.0
**Architecture**: Build-Free But Buildable (BFBB)

The official web-based admin interface for Fazt. Provides comprehensive management of apps, aliases, system health, and configuration.

## Architecture

Pure ESM modules with no build step required for development:

```
admin/
├── packages/           # Core libraries (versioned with admin)
│   ├── zap/           # State management + router + commands
│   ├── fazt-sdk/      # API client with mock adapter
│   └── fazt-ui/       # CSS design system
├── src/
│   ├── pages/         # Page components (dashboard, apps, etc.)
│   ├── stores/        # Reactive state stores
│   └── routes.js      # SPA routes
├── index.html         # Main entry point
└── manifest.json      # Fazt app manifest
```

## Features

- **Real Auth Integration** - `/auth/session`, `/auth/logout` endpoints
- **Role-Based Access** - Owner/Admin only (enforced)
- **Mock Mode** - Testing with `?mock=true` parameter
- **Theme System** - 5 palettes (Stone, Slate, Oxide, Forest, Violet) × Light/Dark
- **Command Palette** - Cmd+K for quick navigation
- **SPA Routing** - Clean URLs with history API
- **Empty States** - Graceful handling of no data

## Development

**No build step needed:**
```bash
fazt app deploy admin --to local --name admin-ui
open http://admin-ui.192.168.64.3.nip.io:8080
```

**With mock data:**
```bash
open http://admin-ui.192.168.64.3.nip.io:8080?mock=true
```

**Optional build (for optimization):**
```bash
npm install
npm run build   # Outputs to dist/
```

## Version Sync

The admin UI version **must match** the fazt binary version. Use `/open` skill to verify:

```bash
# Check all component versions
cat version.json
cat ../internal/config/config.go | grep Version
cat ../knowledge-base/version.json
```

## Components

| Component | Path | Purpose |
|-----------|------|---------|
| zap | `packages/zap/` | Reactive state, router, commands |
| fazt-sdk | `packages/fazt-sdk/` | API client + mock adapter |
| fazt-ui | `packages/fazt-ui/` | CSS tokens + design system |

## Deployment

**To local:**
```bash
fazt app deploy admin --to local --name admin-ui
```

**To production:**
```bash
fazt app deploy admin --to zyt --name admin
```

## Access Control

- **Requirement**: User must have `role = "owner"` or `role = "admin"`
- **Unauthorized**: Redirects to `/unauthorized.html`
- **Login**: Redirects to `/auth/dev/login` (local) or `/auth/login` (production)

## Mock Mode

For testing without real auth:

```bash
# Always shows: kodeman@gmail.com, Owner, Google provider
?mock=true
```

Mock data defined in `packages/fazt-sdk/fixtures/*.json`
