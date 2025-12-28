# The Kernel Pivot

## Summary

Fazt v0.8 reframes the project from a "static site host" to a "Personal OS".
This is not just a naming change—it's a fundamental shift in how we think
about the system's responsibilities and boundaries.

## Rationale

### Why Not Stay as a Web Host?

The static hosting market is commoditized:
- GitHub Pages: Free, easy
- Cloudflare Pages: Free, global CDN
- Vercel: Free tier, excellent DX

Competing on "hosting" is a losing game.

### The Opportunity

There is **no tool** that lets you deploy a full-stack app (frontend + backend
+ database) to a fresh Ubuntu server by copying **one binary**.

- Coolify requires Docker
- Vercel requires cloud
- PocketBase requires a separate frontend host

### The New Framing

| Web Server Thinking | OS Thinking |
|---------------------|-------------|
| "Host my website" | "Run my applications" |
| Sites are content | Apps are processes |
| Files on disk | Virtual filesystem |
| Configuration files | System calls |
| Admin panel | OS Shell |

## What Changes

### Mental Model

Before: "Fazt is like Nginx + SQLite"
After: "Fazt is like a tiny Linux in a binary"

### User Story

Before: "I want to deploy my portfolio site"
After: "I want a personal cloud that runs my apps"

### Target User

Before: Developer deploying a static site
After:
- Indie hacker building micro-SaaS
- AI agent needing a persistent body
- Privacy-conscious user wanting self-hosted tools

## Implementation Impact

### Package Structure

```
Before (v0.7)              After (v0.8)
─────────────              ───────────
internal/                  pkg/kernel/
├── handlers/              ├── proc/      # Process lifecycle
├── hosting/               ├── fs/        # Virtual filesystem
├── config/                ├── net/       # SSL, routing
├── database/              ├── storage/   # SQLite, EDD
├── auth/                  ├── security/  # Auth, permissions
└── middleware/            └── syscall/   # API bridge

                           internal/      # App-specific logic
                           ├── handlers/
                           └── models/
```

### CLI Structure

```
Before (v0.7)              After (v0.8)
─────────────              ───────────
fazt server start          fazt proc start
fazt server init           fazt proc init
fazt deploy                fazt app deploy
                           fazt fs ls
                           fazt net route add
                           fazt storage backup
```

## Success Criteria

1. Developer can reason about Fazt using OS concepts
2. AI agents can interact using familiar system abstractions
3. New features naturally fit into the kernel structure
4. No increase in resource usage from the refactor
