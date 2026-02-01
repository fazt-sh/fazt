# Fazt Implementation State

**Last Updated**: 2026-02-01
**Current Version**: v0.18.0

## Status

State: **CLEAN** - Verified completed features, updated thinking directions

---

## Last Session (2026-02-01)

**Feature Verification + Documentation Update**

### 1. Verified Completed Features

Confirmed three engineering tasks already implemented:
- **E4: Mock OAuth Provider** - Completed in v0.17.0 (`/auth/dev/login`)
- **E7: SPA Routing** - Completed in v0.12.0 (`--spa` flag)
- **E8: Private Directory** - Completed in v0.13.0 (`private/` + auth-gated access)

### 2. Updated Thinking Directions

- Marked E4, E7, E8 as completed in `koder/THINKING_DIRECTIONS.md`
- Added checkmarks and version numbers for tracking
- Cleaned up active engineering directions list

### Key Files Modified

- `koder/THINKING_DIRECTIONS.md` - Marked completed features

---

## Next Up

Choose any direction from `koder/THINKING_DIRECTIONS.md`:

### Product (P)

1. **P1: Google Sign-in Redirect** - Fix OAuth redirect to original page or improve landing page (mostly solved by SPA routing)
2. **P2: Nexus App** - Build comprehensive stress test app using ALL fazt capabilities
3. **P3: App Audit** - Verify docs sync, list deployed apps, add Google sign-in, track high scores
4. **P4: Target Apps** - Build small startup IT suite (Meet, Docs, Chat, Notes, Sign, Files)

### Engineering (E)

5. **E1: `fazt @peer` Pattern Audit** - Review all commands for peer support, ensure CLI ↔ API parity
6. **E2: Analytics Deep Dive** - Audit collection/storage, user tracking, visualization dashboard
7. **E3: Role-Based Access Control (RBAC)** - Granular permissions per app/resource beyond owner vs user
8. **E4: Mock OAuth Provider** (Plan 24) - Dev login form at `/auth/dev/login` for local testing
9. **E5: SQL Command** (Plan 25) - `fazt sql "SELECT..."` for local/remote DB queries
10. **E6: Qor Extraction Evaluation** - Review `~/Projects/qor` for reusable components/patterns
11. **E7: SPA Routing** (Plan 28) - Clean URLs via `--spa` flag (BFBB + optional enhancement)
12. **E8: Private Directory** (Plan 29) - Reserved `private/` for server-only data, blocked HTTP access

### Documentation (D)

13. **D1: Documentation Overhaul** - Comprehensive markdown docs as Claude skill, sync with API/CLI

### Admin UI (A)

14. **A1: Continue Page-by-Page Refactoring** - Aliases, System, Settings pages → design system (follow `refactoring-pages.md`)
15. **A2: Admin API Parity** - Build features to match CLI/API capabilities

### Strategy (S)

16. **S1: Capability Comparison** - Re-evaluate vs Supabase & Vercel (gaps, advantages)
17. **S2: "Break Hyperscaler Stack"** - Position fazt as sovereign compute unit replacing cloud lock-in
18. **S3: Vertical Scaling Evaluation** - Test performance at $50, $500 VPS tiers
19. **S4: External Integration Value Matrix** - Rank S3, Litestream, Turso, Cloudflare by value-add

### Business (B)

20. **B1: License Discussion** - Consider Fair Code License ($1000 per $1M revenue over $1M threshold)
21. **B2: Cloud Provider Partnerships** - Explore DigitalOcean, MS, Google one-click installer deals
22. **B3: Private Repo Feasibility** - Evaluate going private until license resolved

### Vision (V)

23. **V1: Philosophy Rewrite** - Update `koder/philosophy/` to reflect sovereign compute vision

---

---

## Ideas Available

See `koder/ideas/` for detailed specifications.

### Roadmap: Version Themes (v0.8-v0.20)

- **v0.8 - Kernel**: OS nomenclature, EDD, safeguards, events, pulse, devices, resilience primitives
- **v0.9 - Storage**: Unified storage API (kv/ds/rd/s3), shards, IPFS-lite content addressing
- **v0.10 - Runtime**: Serverless v2, standard library, JS-cron, sandbox, WASM primitives, SSG (Jekyll-lite)
- **v0.11 - Distribution**: App identity (UUIDs), Git-based marketplace, manifest permissions
- **v0.12 - Agentic**: AI shim (OpenAI/Anthropic/Gemini), MCP server, harness apps (self-evolving)
- **v0.13 - Network**: VPN gateway (WireGuard), multi-domain mapping, shadow apps
- **v0.14 - Security**: Notary kernel (hardware crypto), temporal identity (TOTP), kernel RLS
- **v0.15 - Identity**: Persona (hardware-bound owner), sovereign SSO, OAuth bridge provider
- **v0.16 - Mesh**: P2P sync, ActivityPub/Nostr protocols, threshold trust (Shamir)
- **v0.17 - Realtime**: WebSocket pub/sub channels, server push, presence tracking
- **v0.18 - Email**: SMTP sink, email routing by local part, serverless triggers
- **v0.19 - Workers**: Persistent job queue, retry/dead-letter, progress reporting, concurrency control
- **v0.20 - Services**: Forms, Image, PDF, Markdown, Search, QR, Comments, ShortURL, Captcha, Hooks, RAG, Notify, API Profiles

### Experimental Ideas (crazy.md)

- Multiple DBs per instance
- GPG support
- PHP/CGI server
- Sensor layer (camera/audio for VLM input)
- JSON-first data formats with validated types
- Wazero for multi-language runtimes (Ruby/Python)
- Fazt as MCP client (browser/computer/mobile control)
- Money as resource (virtual currency for activity cost modeling)
- Lite extractions: Bitcoin primitives, BitTorrent primitives, GPG-lite

### Lite Extractions (evaluated projects)

**USE-AS-IS** (ready to import):
- backlite: SQLite task queues with retry/backoff
- gortsplib: RTSP client/server for IP cameras
- mochi-mqtt: Embeddable MQTT v5 broker for IoT
- peerdiscovery: UDP multicast LAN peer discovery
- pion/webrtc: P2P NAT traversal, data channels
- bluemonday: HTML sanitization (security-critical)
- gjson: Fast JSON path queries without unmarshaling
- goldmark: CommonMark compliant markdown parser
- go-expr: Safe expression language for Go

**PATTERN** (architectural ideas):
- go-redka: Redis-on-SQLite schema patterns
- go-crush: Skills discovery, permissions flow, LSP routing
- go-adk: State prefix convention (app:/user:/temp:)
- go-chi: URL format extension middleware

**EXTRACT** (specific algorithms):
- go-lingoose: Recursive text splitter + cosine similarity

---

## Quick Reference

```bash
# Database location (single DB for everything)
~/.fazt/data.db

# Override if needed
fazt server start --db /custom/path.db
# or
export FAZT_DB_PATH=/custom/path.db

# Deploy admin UI
cd admin && npm run build
fazt app deploy dist --to local --name admin-ui

# Admin UI URLs
http://admin-ui.192.168.64.3.nip.io:8080?mock=true  # Local mock
https://admin.zyt.app                                # Production
```
