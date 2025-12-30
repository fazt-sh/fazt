# Fazt Roadmap: v0.7 → v0.16

## Overview

Fazt evolves from a **static site host** to a **Personal Cloud OS**.

Each version builds on the previous, adding capabilities while maintaining:
- Single binary (no external dependencies)
- Single database (`data.db`)
- Resource efficiency (1GB RAM target)

---

## v0.7 - Cartridge PaaS (Current)

**Status**: Implemented

**Theme**: Static hosting with the "Cartridge" philosophy.

**Capabilities**:
- VFS: Files stored as BLOBs in SQLite
- Auto HTTPS via CertMagic
- Reserved domains: `admin.*`, `root.*`, `404.*`
- Basic serverless JS runtime (Goja)
- Analytics buffering, redirects, webhooks
- CLI: deploy, server management

**Philosophy Established**:
- Binary is disposable
- Database is precious
- Backup = copy one file

---

## v0.8 - Kernel

**Status**: Planned

**Theme**: Paradigm shift from "web server" to "operating system".

**Key Changes**:
- **OS Nomenclature**: Internal packages renamed to OS concepts
  - `proc/` - Process lifecycle, systemd
  - `fs/` - Virtual filesystem
  - `net/` - SSL, routing, domains, egress proxy
  - `storage/` - SQLite engine
  - `security/` - Auth, permissions
  - `syscall/` - API bridge
- **EDD**: Evolutionary Database Design (append-only schema)
- **Safeguards**: Resource limits, circuit breakers
- **"Everything is an App"**: Sites become Apps with UUIDs
- **Events**: Internal event bus for IPC
- **Proxy**: Network egress control with caching
- **Pulse**: Cognitive observability (system self-awareness)
  - Periodic health collection and LLM analysis
  - Natural language queries about system state
  - Proactive notifications for critical issues
- **Devices**: External service abstraction (`/dev/*`)
  - `dev.billing` - Stripe, Paddle, LemonSqueezy
  - `dev.sms` - Twilio, MessageBird
  - `dev.email` - SendGrid, Postmark
  - `dev.oauth` - Google, GitHub, Apple
- **Infrastructure**: Cloud substrate abstraction (`dev.infra.*`)
  - `dev.infra.vps` - Hetzner, DigitalOcean, Vultr
  - `dev.infra.dns` - Cloudflare, Hetzner DNS
  - `dev.infra.domain` - Cloudflare Registrar
- **Resilience Primitives**: Infrastructure-layer fallbacks
  - `beacon` - Local network discovery (mDNS) when DNS unavailable
  - `timekeeper` - Local time consensus when NTP unavailable
  - `chirp` - Audio data transfer when no network exists
  - `mnemonic` - Human-channel exchange (voice, paper, radio)

**New Surface**:
- `fazt proc start|stop|upgrade`
- `fazt fs mount|ls`
- `fazt net route add`
- `fazt storage migrate`
- `fazt events emit|watch`
- `fazt net logs|allow|limits`
- `fazt pulse status|ask|history`
- `fazt dev list|test|logs|config`
- `fazt dev infra vps list|create|destroy|status`
- `fazt dev infra dns zones|records|set|delete`
- `fazt dev infra domain check|register|list`
- `fazt beacon status|scan`
- `fazt time status|sync`
- `fazt chirp send|listen`
- `fazt mnemonic encode|decode`

**Specs**: `specs/v0.8-kernel/`

---

## v0.9 - Storage

**Status**: Planned

**Theme**: Advanced storage patterns for scale.

**Key Changes**:
- **Unified Storage API**: `fazt.storage` namespace
  - `.kv` - Key-value store
  - `.ds` - Document store (JSON)
  - `.rd` - Relational (virtual tables)
  - `.s3` - Blob storage
- **Shards**: Micro-document pattern for high-volume data
- **IPFS-Lite**: Content-addressable storage
  - Files indexed by CID (hash)
  - `/ipfs/<CID>` gateway
  - Automatic de-duplication

**New Surface**:
- `fazt.storage.kv.get|set|delete`
- `fazt.storage.ds.find|insert|update`
- `fazt.fs.cid(path)`
- `fazt storage upload <file>`

**Specs**: `specs/v0.9-storage/`

---

## v0.10 - Runtime

**Status**: Planned

**Theme**: Intelligent serverless execution.

**Key Changes**:
- **Serverless v2**: `api/` folder convention
  - Entry point: `api/main.js`
  - `require()` shim for local imports
- **Standard Library**: Bundled ES5 builds
  - `lodash`, `cheerio`, `uuid`, `zod`, `marked`
  - `require('lodash')` works without npm
- **JS-Cron**: Scheduled function execution
  - Defined in `app.json`
  - Hibernate architecture (zero RAM when idle)
- **Sandbox**: Safe code execution for agents
  - Isolated environment with resource limits
  - Optional capability grants
- **WASM Primitive**: Internal wazero runtime
  - Not exposed to JS apps
  - Used by kernel services for performance-critical ops
  - Embedded modules: libimage, libpdf, libxlsx

**New Surface**:
- `require('./db.js')` - Local imports
- `require('lodash')` - Stdlib access
- `fazt.schedule(delay, state)`
- `app.json` cron definitions
- `fazt.sandbox.exec(options)` - Safe code execution

**Specs**: `specs/v0.10-runtime/`

---

## v0.11 - Distribution

**Status**: Planned

**Theme**: App ecosystem and marketplace.

**Key Changes**:
- **App Identity**: Stable UUIDs for all apps
  - Decouples identity from subdomain
  - Enables rename without data loss
- **Marketplace**: Git-based app stores
  - Repos as package sources
  - `fazt app install <name>`
  - Personal vs marketplace apps
- **Manifest**: `app.json` for metadata, permissions

**New Surface**:
- `fazt app install|update|remove <name>`
- `fazt app list --source marketplace`
- `app.json` permission declarations

**Specs**: `specs/v0.11-distribution/`

---

## v0.12 - Agentic

**Status**: Planned

**Theme**: AI-native platform capabilities.

**Key Changes**:
- **AI Shim**: `fazt.ai` namespace
  - Unified interface for OpenAI/Anthropic/Gemini
  - Auto credential injection from env
  - Streaming support
- **MCP Server**: Fazt as Model Context Protocol server
  - Agents can deploy apps, read logs, query data
- **Harness Apps**: Self-evolving agentic applications
  - Apps that modify their own code
  - Git-like versioning in VFS

**New Surface**:
- `fazt.ai.complete(prompt, options)`
- `fazt.ai.stream(prompt, handler)`
- `fazt.git.commit|diff|rollback`
- `fazt.kernel.deploy|status`

**Specs**: `specs/v0.12-agentic/`

---

## v0.13 - Network

**Status**: Planned

**Theme**: Advanced networking primitives.

**Key Changes**:
- **VPN Gateway**: Built-in WireGuard (userspace)
  - Zero-config peer provisioning
  - QR code setup
  - "Shadow apps" only visible via VPN
- **Multi-Domain**: Custom domain mapping
  - Map external domains to apps
  - On-demand HTTPS for any domain

**New Surface**:
- `fazt net vpn add-peer`
- `fazt net domain map <domain> <app>`
- `fazt.net.vpn.status()`
- `fazt.net.vpn.peer_info()`

**Specs**: `specs/v0.13-network/`

---

## v0.14 - Security

**Status**: Planned

**Theme**: Cryptographic primitives and trust.

**Key Changes**:
- **Notary Kernel**: Hardware-bound cryptography
  - Persona: Process-level keypairs
  - Attestation: Code integrity verification
  - Vaulting: Sealed memory for secrets
- **Temporal Identity**: TOTP integration
  - 2FA for sensitive syscalls
- **Kernel RLS**: Row-level security at Go layer
  - Auto-filtering by `app_id` and `user_id`

**New Surface**:
- `fazt.security.sign(data)`
- `fazt.security.verify(data, sig)`
- `fazt.security.vault.store|retrieve`
- `fazt.halt(reason, data)` - Human-in-the-loop

**Specs**: `specs/v0.14-security/`

---

## v0.15 - Identity

**Status**: Planned

**Theme**: Sovereign identity and authentication.

**Key Changes**:
- **Persona**: Hardware-bound owner identity
  - Cryptographic proof, not passwords
  - Kernel-managed keypair
- **Sovereign SSO**: Zero-handshake auth across subdomains
  - Session inheritance from OS shell
  - No OAuth dance for internal apps
- **OAuth Bridge**: Extend identity to external apps
  - "Sign in with Fazt" provider

**New Surface**:
- `fazt.security.getPersona()`
- `fazt.security.isOwner()`
- `fazt.security.signAssertion()`
- `fazt.security.requireAuth()`

**Specs**: `specs/v0.15-identity/`

---

## v0.16 - Mesh

**Status**: Planned

**Theme**: Decentralization and federation.

**Key Changes**:
- **Kernel Mesh**: P2P synchronization between nodes
  - Gossip protocol for data sync
  - Local + remote instances
- **Protocol Support**:
  - ActivityPub: Federate with Mastodon
  - Nostr: Sovereign social identity
- **Threshold Trust**: Shamir secret sharing
  - Split secrets across nodes
  - Multi-device key recovery

**New Surface**:
- `fazt mesh join|status|sync`
- `fazt.mesh.broadcast(data)`
- `fazt.protocols.activitypub.*`
- `fazt.protocols.nostr.*`

**Specs**: `specs/v0.16-mesh/`

---

## v0.17 - Realtime

**Status**: Planned

**Theme**: WebSocket-based real-time communication.

**Key Changes**:
- **WebSocket Support**: Native pub/sub channels
  - Endpoint: `wss://app.domain.com/_ws`
  - Public, private, and presence channels
- **Server Push**: Broadcast from serverless handlers
- **Presence**: Track connected clients

**New Surface**:
- `fazt.realtime.broadcast(channel, data)`
- `fazt.realtime.subscribers(channel)`
- `fazt.realtime.count(channel)`

**Specs**: `specs/v0.17-realtime/`

---

## v0.18 - Email

**Status**: Planned

**Theme**: Inbound email processing.

**Key Changes**:
- **SMTP Sink**: Receive emails at your domain
  - Port 25 listener
  - Route by local part: `support@` → support app
- **Serverless Trigger**: Process with JS handlers
- **Email Storage**: Query inbox via `fazt.email`

**New Surface**:
- `fazt.email.list(options)`
- `fazt.email.get(id)`
- `fazt.email.attachment(id)`
- `fazt.email.markProcessed(id)`

**Specs**: `specs/v0.18-email/`

---

## v0.19 - Workers

**Status**: Planned

**Theme**: Long-running background jobs.

**Key Changes**:
- **Job Queue**: Persistent, survives restarts
  - Spawn jobs that run for minutes
  - Progress reporting
- **Retry & Dead-Letter**: Handle failures gracefully
- **Concurrency Control**: Priority, unique keys

**New Surface**:
- `fazt.worker.spawn(handler, options)`
- `fazt.worker.get(id)`
- `fazt.worker.list(options)`
- `fazt.worker.cancel(id)`
- `job.progress(n)` / `job.log(msg)`

**Specs**: `specs/v0.19-workers/`

---

## v0.20 - Services

**Status**: Planned

**Theme**: Common patterns as platform primitives.

**Key Changes**:
- **Services Layer**: Go libraries between kernel and apps
- **Forms**: Dumb bucket for form submissions
- **Media**: Image resize, optimize, thumbnails, blurhash, QR, barcode, mimetype
- **PDF**: HTML/CSS to PDF generation (WASM-powered)
- **Markdown**: Compile .md to HTML, shortcodes, classless CSS
- **Search**: Full-text indexing with Bleve
- **QR**: Generate QR codes from text/URL
- **Comments**: User feedback on any entity, threading, moderation
- **Short URL**: Shareable links with click tracking
- **Captcha**: Math/text challenges for spam protection
- **Hooks**: Bidirectional webhooks (inbound + outbound)
  - Inbound: Signature verification for Stripe, GitHub, Shopify
  - Outbound: Event delivery with retry and logging
- **Sanitize**: HTML/text sanitization (XSS protection)
- **Money**: Decimal arithmetic for currency (integer cents)
- **Humanize**: Human-readable formatting (bytes, time, numbers)
- **Timezone**: IANA timezone handling (embedded tzdata)
- **Password**: Secure Argon2id hashing for credentials
- **Geo**: Geographic primitives, IP geolocation (embedded geodata)

**Architecture**:
```
Apps (JS)
    ↓
Services (Go)  ← forms, media, pdf, markdown, search, qr, hooks,
                  sanitize, money, humanize, timezone, password, geo
    ↓
Kernel (Go)    ← proc, fs, net, storage, security, wasm, pulse, dev
```

**New Surface**:
- `fazt.services.forms.list|get|delete|count|clear`
- `fazt.services.media.resize|thumbnail|crop|optimize|convert`
- `fazt.services.media.blurhash|qr|barcode|mimetype`
- `fazt.services.pdf.fromHtml|fromFile|fromUrl|merge|info`
- `fazt.services.markdown.render|renderFile`
- `fazt.services.search.index|query|reindex|dropIndex`
- `fazt.services.qr.generate|dataUrl`
- `fazt.services.comments.add|list|get|hide|approve|delete`
- `fazt.services.shorturl.create|get|stats|delete`
- `fazt.services.captcha.create|verify`
- `fazt.services.hooks.events|replay|register|emit`
- `fazt.services.sanitize.html|text|markdown|url`
- `fazt.services.money.add|subtract|multiply|divide|format|parse`
- `fazt.services.humanize.bytes|time|duration|number|ordinal`
- `fazt.services.timezone.now|convert|format|isDST|info`
- `fazt.services.password.hash|verify|needsRehash`
- `fazt.services.geo.distance|fromIP|contains|nearby`
- `/_services/forms/{name}` - POST endpoint
- `/_services/media/{path}` - On-the-fly processing
- `/_services/pdf/render` - HTML to PDF
- `/_services/qr?data=...` - QR generation
- `/_services/barcode?data=&format=...` - Barcode generation
- `/_services/geo/ip` - IP geolocation endpoint
- `/_services/comments/{target}` - Comments endpoint
- `/_s/{code}` - Short URL redirect
- `/_hooks/{provider}` - Inbound webhook receiver

**Specs**: `specs/v0.20-services/`

---

## Beyond v0.20

Ideas not yet assigned to versions:

- **Telegram Bot Server**: Native bot hosting
- **Hardware Attestation**: TPM-sealed databases
- **Vibe-to-Cartridge**: Natural language app generation
- **Digital Executor**: Post-mortem automation

---

## Dependency Graph

```
v0.7 (Current)
  │
  ▼
v0.8 (Kernel) ──────────────────────────────────┐
  │                                              │
  ▼                                              │
v0.9 (Storage)                                   │
  │                                              │
  ▼                                              │
v0.10 (Runtime)                                  │
  │                                              │
  ├───────────────┬──────────────┐               │
  ▼               ▼              ▼               │
v0.11           v0.12          v0.13             │
(Distribution)  (Agentic)      (Network)         │
  │               │              │               │
  └───────────────┴──────────────┘               │
                  │                              │
                  ▼                              │
                v0.14 (Security) ◄───────────────┘
                  │
                  ▼
                v0.15 (Identity)
                  │
                  ▼
                v0.16 (Mesh)
                  │
  ┌───────────────┼───────────────┬───────────────┐
  ▼               ▼               ▼               ▼
v0.17           v0.18           v0.19           v0.20
(Realtime)      (Email)         (Workers)       (Services)
```

**Critical Path**: v0.8 → v0.9 → v0.10 must be sequential.

**Parallel Work**:
- v0.11, v0.12, v0.13 can be developed concurrently after v0.10
- v0.17, v0.18, v0.19, v0.20 can be developed concurrently after v0.10
  (they only depend on Runtime, not on each other or v0.16)
