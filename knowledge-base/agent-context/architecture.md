# Fazt Architecture

## Core Model

**Cartridge Model**: One Binary (`fazt`) + One SQLite DB (`data.db`)

- **Pure Go**: `modernc.org/sqlite`, NO CGO, runs everywhere
- **Swarm Ready**: Multiple nodes mesh into personal cloud
- **AI Native**: Lowers floor (anyone can use), raises ceiling (agents)
- **Resilient**: Works when network is denied

## Uniform Peers

Every fazt instance is a first-class peer. There's no "dev" vs "production"
distinction - just peers that happen to run in different locations.

**Domain handling is automatic:**

| Domain Type | Behavior |
|-------------|----------|
| Real domain (`zyt.app`) | Trusted - never modified |
| Wildcard DNS (`*.nip.io`) | Auto-updates if IP changes |
| IP address | Auto-updates if machine changes |
| Empty | Auto-detects local IP |

This means:
- Same binary, same commands everywhere
- Copy `data.db` to another machine - domain auto-adjusts
- No environment variables to remember
- Real domains are always respected

## App Model

An **app** in fazt is a website with optional serverless capabilities.

- A static site is called an **app**
- A subdomain is called an **alias**

### App Structure

```
my-app/
├── manifest.json      # Required: { "name": "my-app" }
├── index.html         # Entry point
├── static/            # Assets (css, js, images)
├── private/           # Server-only files (auth-gated)
└── api/               # Serverless functions
    └── main.js        # Handles /api/* requests
```

### Where Apps Live

```
servers/                 # gitignored - NOT part of fazt source
└── zyt/                 # Apps for zyt.app
    ├── xray/            # An app
    └── my-new-app/      # Another app
```

Apps are instance-specific, not fazt source code.

## Serverless Runtime

JavaScript files in `api/` are executed server-side via Goja:

```javascript
// api/main.js
if (request.path === '/api/hello') {
  respond({ message: "Hello", time: Date.now() })
}
```

**Limitations**: ES5 syntax, no npm modules, no async/await.

## Current Capabilities (v0.13.x)

| Feature | Status |
|---------|--------|
| Static Hosting | VFS-backed, subdomain routing |
| Admin Dashboard | React SPA at admin.* |
| Serverless Runtime | JavaScript via Goja |
| OAuth (Google) | App-level user auth |
| Storage API | KV, Docs, Blobs |
| Private Files | Auth-gated HTTP + serverless access |
| Analytics | Event tracking |
| Remote Management | `fazt peer` CLI |
