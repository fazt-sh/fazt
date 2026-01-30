# Fazt Overview

Fazt is a **single-owner compute node that can support multiple users**. One Go
binary + one SQLite database = a complete platform to launch apps and services.

## Philosophy

- **Single Owner**: One person owns and operates the node. They can build apps
  that serve many users - a chat app, a SaaS product, a startup MVP.
- **Cartridge Model**: Binary + DB file. Copy to any machine, it just works.
- **Pure Go**: No CGO, no external dependencies. Cross-compiles to everything.
- **Static First**: Primary goal is self-hostable Surge alternative. Zero build
  steps for deployment.
- **Uniform Peers**: Every fazt instance is a first-class peer. No "dev" vs
  "production" - just peers in different locations.
- **Comparable to Supabase/Vercel**: In value delivered, but fully self-contained.
  No vendor dependencies, no external services required.

## Core Capabilities

| Capability | Description |
|------------|-------------|
| Static Hosting | VFS-backed file serving, any static site |
| Multi-site | Subdomain routing (app.domain.com) |
| Serverless Runtime | JavaScript execution via Goja |
| Storage | Document store, key-value, blob storage |
| Authentication | OAuth providers (Google, GitHub, etc.) |
| Admin Dashboard | React SPA for management |
| Peer-to-Peer | Native node-to-node communication |
| Remote Management | Deploy, upgrade, monitor from CLI |

## Architecture

```
┌─────────────────────────────────────────┐
│              fazt binary                │
├─────────────────────────────────────────┤
│  HTTP Server (net/http)                 │
│  ├── Static file serving (VFS)          │
│  ├── Subdomain routing                  │
│  ├── Serverless runtime (Goja)          │
│  ├── Auth (OAuth + cookies)             │
│  └── Admin API                          │
├─────────────────────────────────────────┤
│  SQLite (modernc.org/sqlite)            │
│  ├── Apps & files (VFS)                 │
│  ├── Storage (ds, kv, s3)               │
│  ├── Users & sessions                   │
│  └── Configuration                      │
└─────────────────────────────────────────┘
```

## Capacity (Typical $6 VPS)

| Metric | Limit | Notes |
|--------|-------|-------|
| Read throughput | ~20,000/s | Static files, cached |
| Write throughput | ~800/s | SQLite single-writer |
| Mixed workload | ~2,300/s | 30% writes typical |
| WebSocket connections | 5,000-10,000 | Real-time features |
| Broadcast messages/s | 10,000+ | Ephemeral, never hits disk |

**Key insight**: Broadcasts (cursors, typing indicators, presence) are unlimited
because they never touch SQLite. Only persist what matters.

## Terminology

| Term | Meaning |
|------|---------|
| **App** | A website/application hosted on fazt |
| **Alias** | A subdomain pointing to an app |
| **Peer** | A configured fazt instance (local or remote) |
| **Session** | URL-based anonymous workspace (`?s=word-word-word`) |

## Version

Current: **0.11.5**

Check with: `fazt version`
