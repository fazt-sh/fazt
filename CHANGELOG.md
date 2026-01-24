# Changelog

All notable changes to fazt.sh will be documented in this file.

## [Unreleased]

### Added
- **Capacity Module**: New `/api/system/capacity` endpoint with VPS tier profiles
- **Environment Detection**: Server auto-detects if stored domain matches current machine
  - New `internal/provision/detect.go` with IP detection and domain matching
  - On mismatch, warns and falls back to detected local IP
  - Enables portable database between machines
- **User Systemd Service**: Local development server runs as user service
  - `~/.config/systemd/user/fazt-local.service` - no sudo required
  - Persists across reboots via linger
  - Managed with `systemctl --user` commands

### Changed
- **Install Script**: Unified `install.sh` with three modes:
  - Production Server (system service, real domain, HTTPS)
  - Local Development (user service, auto-start, IP-based)
  - CLI Only (just binary, connect to remotes)
- **Session Skills**: `/fazt-start` and `/fazt-stop` now remote-agnostic
  - Dynamically detect all configured remotes
  - No hardcoded server names
  - Reports concurrent user estimates based on detected hardware
  - Profiles for $6, $15, $40 VPS tiers
- **Capacity Guide**: `koder/CAPACITY.md` documents performance limits and real-time scenarios
  - Models for collaborative docs, presence, chat, cursor sharing, AI monitoring
  - Key insight: broadcasts unlimited, only writes hit 800/s limit

### Changed
- **Storage Write Serialization**: All writes now go through single-writer WriteQueue
  - Eliminates SQLITE_BUSY errors under high concurrency
  - Tested: 100% write success at 2000 concurrent users
- **Analytics WriteQueue Integration**: Analytics batch writes now route through global WriteQueue
  - Fixes SQLITE_BUSY errors during high-traffic analytics collection
  - Added `storage.InitWriter()` and `storage.QueueWrite()` for cross-package use
- **Connection Pool Tuning**: MaxOpen=10, MaxIdle=10, Lifetime=5min, busy_timeout=2s
- **Retry Logic**: Storage operations retry 5x with exponential backoff (20-320ms)
- **VM Pool Size**: Increased from 10 to 100 for better concurrency

### Performance
- **100% success rate** at 2000 concurrent users (was 80%)
- Pure reads: 19,536 req/sec | Pure writes: 832 req/sec | Mixed (30% writes): 2,282 req/sec
- RAM under load: ~60MB stable

## [0.10.10] - 2026-01-24

### Fixed
- **Runtime Timeout**: Increased serverless execution timeout from 1s to 5s
  - Storage write operations were timing out in production
  - Affects `ds.insert`, `ds.update`, `ds.delete`, `kv.set`, `s3.put`

### Added
- **Force Restart**: `/api/upgrade?force=true` restarts service without version change
  - Useful when server needs restart but already on latest version

## [0.10.9] - 2026-01-24

### Added
- **Force Restart Option**: Added `?force=true` query param to upgrade endpoint

## [0.10.8] - 2026-01-21

### Added
- **Debug Mode**: New `FAZT_DEBUG=1` environment variable for development observability
  - Enabled by default in development mode (when `ENV=development`)
  - Logs storage operations with timing: `[DEBUG storage] find app/col query={} rows=3 took=2ms`
  - Logs runtime requests with tracing: `[DEBUG runtime] req=a1b2c3 app=myapp path=/api/hello status=200 took=45ms`
  - Logs VM pool state for performance monitoring
  - Warns on common mistakes (e.g., setting `id` field in insert)
  - All output goes to stderr - no storage, just realtime streaming

## [0.10.7] - 2026-01-20

### Fixed
- **Storage `findOne` API**: Now accepts query object `{ id, session }` instead of
  just string ID, matching expected MongoDB-style usage
- **Storage `id` field queries**: The `id` field is now queryable in `find()`,
  `update()`, and `delete()` operations - previously queries like `{ id: 'x' }`
  silently returned empty results
- **Better type validation**: Storage bindings now throw descriptive errors when
  wrong argument types are passed (e.g., "got object, expected string")

## [0.10.6] - 2026-01-20

### Fixed
- **SQLite Busy Timeout**: Added `PRAGMA busy_timeout=5000` to database initialization
  - Fixes ~10% intermittent timeout errors under concurrent load
  - SQLite now waits up to 5 seconds for write locks instead of failing immediately
  - Root cause: concurrent requests were getting SQLITE_BUSY errors instantly

## [0.10.5] - 2026-01-20

### Added

#### LLM-Friendly CLI Improvements
- **New Templates**: Added `static`, `vue`, and `vue-api` templates
  - `fazt app create myapp --template static` - Basic HTML/CSS/JS
  - `fazt app create myapp --template vue` - Vue 3 + Vite
  - `fazt app create myapp --template vue-api` - Vue + serverless + storage helpers
- **App Validation**: `fazt app validate <dir>` checks apps before deployment
  - Validates manifest.json schema
  - Checks required files exist
  - Parses JavaScript for syntax errors
  - `--json` flag for machine-readable output
- **App Logs**: `fazt app logs <app>` streams serverless execution logs
  - `-f` flag to follow/stream logs in real-time
  - `-n <count>` to show N recent logs
  - SSE endpoint at `/api/logs/stream`

#### Developer Experience
- **Respect .gitignore**: Deploy now skips files listed in `.gitignore`
  - Also skips `node_modules/`, `.git/`, `.DS_Store`, `*.log` by default
- **Better JS Errors**: Improved Goja error messages with line numbers and context
  - Extracts error type, line number, and source context
  - Helps LLMs debug serverless code

### Changed
- Updated help text to reflect new templates and commands
- Simplified `/fazt-app` skill from 1004 to 127 lines (orchestrates CLI instead of
  duplicating knowledge)

### Fixed
- **vue-api Template**: Fixed serverless API execution
  - Renamed `api/items.js` to `api/main.js` (serverless runtime requires main.js)
  - Added `handler(request)` call at end of file (was missing execution)
  - Added `genId()` helper function (`fazt.uuid()` doesn't exist)
  - Fixed Vite config to externalize Vue for proper builds

## [0.10.1] - 2026-01-19

### Fixed
- **VFS File Serving**: Fixed app routing to use `app_id` column after migration.
  After v0.10 migration, files are stored with both `site_id` (old subdomain)
  and `app_id` (new UUID). The VFS layer now correctly uses app_id for lookups
  when serving migrated apps.

## [0.10.0] - 2026-01-19

### Added

#### App Identity Model
- **Permanent App IDs**: Apps now have stable UUIDs (`app_xxxxxxxx`) independent
  of their routing aliases. IDs persist across renames and can be used for
  programmatic access.
- **Alias System**: Subdomains are now routing aliases that map to app IDs
  - `proxy` - Standard serving (default)
  - `redirect` - HTTP 301/302 to target URL
  - `reserved` - Holds subdomain without content
  - `split` - Weighted traffic distribution

#### Lineage & Forking
- **Fork with Lineage**: `fazt app fork <app-id> --alias new-name` creates a copy
  with full lineage tracking
  - `original_id` - Root ancestor of fork chain
  - `forked_from_id` - Immediate parent
- **Lineage Tree**: `fazt app lineage <app-id>` shows fork relationships

#### CLI Enhancements
- **@peer Remote Execution**: Run commands on remote peers directly
  - `fazt @zyt app list` - List apps on zyt peer
  - `fazt @zyt app info myapp` - Get app info from remote
- **Alias Management**:
  - `fazt app link <app-id> <alias>` - Attach alias to app
  - `fazt app unlink <alias>` - Remove alias
  - `fazt app reserve <alias>` - Reserve subdomain
  - `fazt app swap <alias1> <alias2>` - Atomic swap
  - `fazt app split <alias> <app1>:<weight> <app2>:<weight>` - Traffic splitting

#### API Endpoints
- `GET /api/v2/apps` - List apps with visibility filter
- `GET /api/v2/apps/:id` - Get app by ID
- `POST /api/v2/apps/:id/fork` - Fork an app
- `GET /api/v2/apps/:id/lineage` - Get lineage tree
- `GET /api/aliases` - List all aliases
- `POST /api/aliases` - Create alias
- `PUT /api/aliases/:subdomain` - Update alias
- `DELETE /api/aliases/:subdomain` - Delete alias
- `POST /api/aliases/swap` - Atomic alias swap
- `POST /api/aliases/split` - Configure traffic split
- `POST /api/cmd` - Command gateway for @peer execution

#### Agent Endpoints (/_fazt/*)
For LLM testing workflows:
- `GET /_fazt/info` - App metadata
- `GET /_fazt/storage` - List storage keys
- `GET /_fazt/storage/:key` - Get storage value
- `POST /_fazt/snapshot` - Create named snapshot
- `POST /_fazt/restore/:name` - Restore snapshot
- `GET /_fazt/snapshots` - List snapshots
- `GET /_fazt/logs` - Recent execution logs
- `GET /_fazt/errors` - Recent errors

#### Traffic Splitting
- Weighted distribution across multiple app versions
- Sticky sessions via `X-Fazt-Variant` cookie
- Useful for A/B testing and gradual rollouts

### Changed
- **Database Schema**: Migration 012 adds `apps_new` table with identity fields
  and `aliases` table. Existing apps are migrated automatically.
- **Routing**: Subdomain resolution now checks aliases table first, then falls
  back to direct app ID lookup for backwards compatibility.

### Migration Notes
- Existing apps receive auto-generated IDs during migration
- Existing subdomains become `proxy` type aliases pointing to their app IDs
- All existing functionality remains backward compatible

## [0.9.27] - 2026-01-18

### Changed
- **CSP**: Added Cloudflare Web Analytics (`static.cloudflareinsights.com`)

## [0.9.26] - 2026-01-18

### Changed
- **Serverless Timeout**: Increased default timeout from 100ms to 1 second to
  support larger KV operations (e.g., full state persistence in apps)

## [0.9.25] - 2026-01-18

### Changed
- **Comprehensive CSP Whitelist**: Expanded Content-Security-Policy to allow
  common CDNs for easier app development
  - Script CDNs: jsdelivr, unpkg, cdnjs, tailwindcss, esm.sh, skypack, jspm,
    googleapis, jquery, bootstrapcdn, fontawesome, statically, githubusercontent
  - Style CDNs: jsdelivr, unpkg, cdnjs, Google Fonts, Bunny Fonts, Fontshare,
    Adobe Fonts (Typekit), bootstrapcdn
  - Font hosts: Google Fonts, Bunny Fonts, Fontshare, Adobe Fonts, FontAwesome
  - Connect sources: jsdelivr, unpkg, esm.sh, skypack, jspm, GitHub API
  - Added media-src for audio/video support
  - Added object-src 'none' and frame-ancestors 'none' for security

## [0.9.24] - 2026-01-17

### Added
- **App Templates**: `fazt app create` scaffolds apps from embedded templates
  - `fazt app create myapp` - Creates minimal HTML app
  - `fazt app create myapp --template vite` - Creates Vite-ready app with HMR support
  - `fazt app create --list-templates` - Shows available templates
- **Unified Build Model**: Deploy automatically builds when needed
  - Detects package.json with build script
  - Supports bun, pnpm, yarn, npm (in priority order)
  - Respects lockfiles to match project's preferred package manager
  - Falls back to existing dist/ if no package manager available
  - `--no-build` flag to skip build step
- **Pre-built Branch Detection**: For apps requiring build without local npm
  - Checks for fazt-dist, dist, release, gh-pages branches
  - Automatically uses pre-built branch when build fails
- **API Endpoints**: For GUI and LLM harness integration
  - `POST /api/apps/install` - Install from GitHub URL
  - `POST /api/apps/create` - Create from template
  - `GET /api/templates` - List available templates

### Templates
- `minimal` - Basic HTML app with Tailwind CDN
- `vite` - Full Vite setup with HMR, Tailwind, importmaps, serverless API

### Build Behavior
| Has npm | Has build script | Result |
|---------|-----------------|--------|
| Yes | Yes | Runs build, deploys dist/ |
| Yes | No | Deploys source |
| No | Yes + dist/ exists | Deploys existing dist/ |
| No | Yes + no dist/ | **Error** (clear message) |
| No | No | Deploys source |

## [0.9.23] - 2026-01-16

### Changed
- Verification release to test auto-restart fix from v0.9.22

## [0.9.22] - 2026-01-16

### Fixed
- **Remote Upgrade Auto-Restart**: Service now restarts automatically after upgrade
  - Root cause: `exec.Command()` within systemd cgroup doesn't spawn independent processes
  - Solution: Use `systemd-run --scope` to escape cgroup, then `os.Exit(0)` lets
    systemd's `Restart=always` bring service back with new binary
  - Removed redundant `setcap` call (AmbientCapabilities in systemd unit handles this)
  - No more SSH required for remote upgrades

## [0.9.8] - 2026-01-16

### Fixed
- **Upgrade Auto-Restart Timing**: Added 500ms delay before restart command to ensure
  HTTP response is fully sent to client before service restarts

## [0.9.7] - 2026-01-16

### Fixed
- **API Endpoint Routing**: Fixed `/api` and `/api/*` paths not reaching serverless
  handler with storage support
  - Site handler now routes API paths to the new `ServerlessHandler`
  - Apps can now use `fazt.storage.kv`, `fazt.storage.ds`, `fazt.storage.s3`
  - `api/main.js` is the entry point for API handlers (not root `main.js`)

## [0.9.6] - 2026-01-16

### Fixed
- **Remote Upgrade Self-Restart**: Fixed upgrade handler not restarting the service
  - Added sudoers rule to allow fazt user to run `systemctl restart fazt`
  - Upgrade handler now uses `sudo systemctl restart fazt`
  - Install script creates `/etc/sudoers.d/fazt` with NOPASSWD rules
- **Remote Upgrade Directory Permissions**: Fixed "permission denied" on binary staging
  - `/usr/local/bin` now has group write permission for fazt user
  - `ReadWritePaths` in systemd only affects namespace, not filesystem permissions
  - Install script sets directory group to fazt with g+w permission

## [0.9.5] - 2026-01-16

### Added
- **Storage Primitives**: App-scoped storage via `fazt.storage.*` JS API
  - `fazt.storage.kv`: Key-value store with TTL support
    - `set(key, value, ttlMs?)`, `get(key)`, `delete(key)`, `list(prefix)`
  - `fazt.storage.ds`: Document store with MongoDB-style queries
    - `insert(collection, doc)`, `find(collection, query)`, `findOne(collection, id)`
    - `update(collection, query, update)`, `delete(collection, query)`
    - Query operators: `$eq`, `$ne`, `$gt`, `$lt`, `$gte`, `$lte`, `$in`, `$contains`
  - `fazt.storage.s3`: Blob storage with SHA256 hashing
    - `put(path, data, mimeType)`, `get(path)`, `delete(path)`, `list(prefix)`
  - Migration 010: `app_kv`, `app_docs`, `app_blobs` tables
- **Stdlib Embedding**: CommonJS modules available via `require()`
  - lodash, uuid, dayjs, validator, marked, zod, cheerio
  - Vite-built IIFE bundles embedded in binary
  - `require('lodash')` resolves to embedded stdlib before local files

### Changed
- Serverless runtime now supports `require()` for code splitting
- Apps isolated by `app_id` - cannot access other apps' storage

## [0.9.4] - 2026-01-14

### Fixed
- **Atomic Binary Replacement**: Remote upgrade now uses atomic rename instead of
  copy, fixing "text file busy" error when upgrading a running binary
  - New binary staged to `.fazt.new` in same directory
  - `os.Rename` atomically replaces running binary (works because it replaces
    directory entry, not file content)

## [0.9.3] - 2026-01-14

### Fixed
- **Binary Ownership for Self-Upgrade**: Binary now owned by service user, not root
  - Enables `fazt remote upgrade` to work without sudo
  - Install script chowns binary to SERVICE_USER during upgrades
  - Fresh installs chown binary to service user after SetCapabilities
  - **Existing installs**: Run `sudo chown fazt:fazt /usr/local/bin/fazt` once

## [0.9.2] - 2026-01-14

### Fixed
- **CSP Subdomain Communication**: Apps on same fazt server can now fetch from
  each other (e.g., root site fetching manifests from app subdomains)
  - `connect-src` now includes `https://*.{domain}` dynamically
  - Enables cross-app data sharing without CSP violations

## [0.9.1] - 2026-01-14

### Fixed
- **Install Script Upgrades**: Service file now updated during upgrades
  - Extracts existing user from service file
  - Regenerates with latest template (includes ProtectSystem fixes)
  - Runs `systemctl daemon-reload` automatically
  - Future upgrades via `fazt remote upgrade` will work without manual SSH

### Added
- **Release Skill**: `.claude/commands/fazt-release.md` for consistent releases

## [0.9.0] - 2026-01-14

### Added
- **Peers Table**: All peer configuration stored in SQLite (`peers` table)
  - Migration 009 creates the peers table
  - No external config files - move DB, everything works
  - Future-ready: `node_id`, `public_key` fields for mesh (v0.16)
- **`fazt remote` Commands**: Native fazt-to-fazt communication
  - `fazt remote add <name> --url <url> --token <token>` - Add peer
  - `fazt remote list` - List configured peers
  - `fazt remote remove <name>` - Remove peer
  - `fazt remote default <name>` - Set default peer
  - `fazt remote status [name]` - Check peer health and version
  - `fazt remote apps [name]` - List apps on peer
  - `fazt remote upgrade [name]` - Check/perform upgrades
  - `fazt remote deploy <dir>` - Deploy to peer
- **Auto-Migration**: Imports `~/.fazt/config.json` into peers table on first run
- **Claude Skills**: Management commands via Claude Code
  - `/fazt-status`, `/fazt-deploy`, `/fazt-apps`, `/fazt-upgrade`
- **CLAUDE.md Enhancements**: Environment context, capability matrix, MCP vs Skills

### Changed
- **Client DB Location**: `~/.config/fazt/data.db` (XDG compliant)
- Peer status (last_seen, version) tracked in database

### Migration
- Old `~/.fazt/config.json` automatically imported to peers table
- File renamed to `config.json.migrated` after import

## [0.8.4] - 2026-01-14

### Fixed
- **Remote Upgrade Compatibility**: Fixed upgrade API failing due to `ProtectSystem=full`
  - Changed systemd service to use `ProtectSystem=strict` with `ReadWritePaths=/usr/local/bin`
  - Backup now uses temp directory instead of alongside binary
  - Existing installations need to update service file manually or reinstall

## [0.8.3] - 2026-01-14

### Fixed
- **Auto-detect Service Database**: `fazt server create-key` now automatically finds the correct database
  - Reads systemd service file to find `WorkingDirectory`
  - No more `--db` flag needed when service is installed
  - Shows which database is being used for transparency
  - Priority: explicit `--db` flag > `FAZT_DB_PATH` env > service path > `./data.db`

## [0.8.2] - 2026-01-14

### Added
- **Remote Upgrade API**: `POST /api/upgrade` endpoint for remote server upgrades
  - Check for updates: `POST /api/upgrade?check=true`
  - Perform upgrade: `POST /api/upgrade` (requires API key auth)
  - Auto-restarts service after successful upgrade

## [0.8.1] - 2026-01-14

### Added
- **MCP Routes**: Wired up Model Context Protocol HTTP endpoints
  - `POST /mcp/initialize` - MCP handshake
  - `POST /mcp/tools/list` - List available tools
  - `POST /mcp/tools/call` - Execute MCP tools

### Fixed
- **Install Script**: Fixed version detection regex to use portable `-oE` instead of `-oP`
  - Fixes "vunknown" display on systems without Perl regex support

## [0.8.0] - 2026-01-13

### Added
- **Multi-Server Client Config**: New `~/.fazt/config.json` for client-side configuration
  - `fazt servers add/list/default/remove/ping` commands
  - Smart defaults: single server auto-selects, multiple requires `--to`
- **MCP Server Package**: Model Context Protocol implementation (`internal/mcp/`)
  - Tools: `fazt_servers_list`, `fazt_apps_list`, `fazt_deploy`, `fazt_app_delete`, `fazt_system_status`
  - `fazt server create-key` for headless API key creation
- **Serverless Runtime**: JavaScript execution via Goja (`internal/runtime/`)
  - VM pooling for performance
  - Request/response injection for `/api/*` routes
  - `require()` shim with caching for module loading
  - `fazt.*` namespace: `fazt.app`, `fazt.env`, `fazt.log`
- **Analytics Dashboard**: Built into admin panel with real-time stats
  - Stats cards (today/week/month/all-time)
  - Timeline chart, top domains/tags, source type breakdown
- **Apps API**: New `/api/apps` endpoints (replaces `/api/sites`)
  - `GET /api/apps` - List all apps with metadata
  - `GET /api/apps/{id}` - App details
  - `DELETE /api/apps/{id}` - Delete app
  - `GET /api/apps/{id}/files` - App file tree
- **CLI**: `fazt client apps` command using new config system

### Changed
- **Sites to Apps Migration**: Database migration (`007_apps.sql`)
  - New `apps` table with id, name, source, manifest
  - New `domains` table for custom domain mapping
  - Backwards compatible: `/api/sites` still works

## [0.7.2] - 2025-12-08

### Added
- **Analytics Buffering**: RAM-based event buffering system to prevent database write storms (`internal/analytics/buffer.go`)
- **System Observability API**: New endpoints for monitoring and resource awareness
  - `GET /api/system/health` - Server status, uptime, version, memory, database stats
  - `GET /api/system/limits` - Resource thresholds (RAM, VFS cache, upload limits)
  - `GET /api/system/cache` - VFS cache statistics (hits, misses, size)
  - `GET /api/system/db` - Database connection statistics
  - `GET /api/system/config` - Server configuration (sanitized)
- **Site Management API**: New endpoints for detailed site operations
  - `GET /api/sites/{id}` - Get single site details
  - `GET /api/sites/{id}/files` - List files in tree format
  - `GET /api/sites/{id}/files/{path}` - Download/view specific file content
- **Traffic Configuration API**: Complete CRUD operations
  - `DELETE /api/redirects/{id}` - Delete a redirect
  - `DELETE /api/webhooks/{id}` - Delete a webhook
  - `PUT /api/webhooks/{id}` - Update webhook (name, endpoint, secret, active status)
- **Response Standardization**: New `internal/api/response.go` package with standard envelope format `{data, meta, error}`
- **Resource Awareness**: New `internal/system/probe.go` detects container/host RAM limits via cgroup v1/v2
- **Test Coverage**: Comprehensive tests for analytics buffer and system probe (16 tests total)

### Changed
- **Authentication**: Fixed middleware to accept Bearer tokens for API access (was blocking CLI deployments)
- **Handler Migration**: Updated 6 handlers to use new standardized response envelope:
  - SitesHandler, SystemHealthHandler, SystemLimitsHandler, SystemCacheHandler, SystemDBHandler, SystemConfigHandler
- **Event Tracking**: Updated PixelHandler, TrackHandler, RedirectHandler, WebhookHandler to use analytics buffer

### Fixed
- Bearer token authentication now works correctly for CLI/API clients
- Analytics events no longer cause DB write contention under high load

### Documentation
- Added `koder/analysis/implementation-review.md` - Implementation status analysis
- Updated `CLAUDE.md` with server stability testing guidance
- Updated `.gitignore` for build artifacts (*.pid, cookies.txt)

## [0.7.1] - 2025-12-07

### Fixed
- **Seeding**: Fixed `EnsureSystemSites` to recursively seed directories. This ensures that assets in subdirectories (like `static/css`, `static/js`) are correctly copied to the VFS during startup, fixing 404 errors on the Admin Dashboard.

## [0.7.0] - 2025-12-07

### Added
- **Admin VFS**: The Admin Dashboard is now a standard "Cartridge" site hosted in the VFS (`system/admin`). This unifies the architecture—everything is served from the database.
- **CLI**: Added `fazt server reset-admin` command to force-update the dashboard assets from the binary to the VFS (useful for upgrades).
- **Frontend**: Dashboard is now a Single Page Application (SPA). All data is fetched via JSON APIs.
- **API**: Added `/api/user/me` endpoint for session info.

### Fixed
- **Install**: Improved upgrade reliability by stopping the service before checking port availability.
- **CSP**: Content Security Policy now allows loading source maps from `cdn.jsdelivr.net`.
- **Assets**: Fixed 404s for static assets by moving them into the VFS structure.

## [0.6.5] - 2025-12-07

### Fixed
- **Assets**: Fixed missing CSS/JS/Image files in production builds. The server now correctly serves static assets from the embedded binary when running in production mode, instead of looking for a non-existent local directory.

## [0.6.4] - 2025-12-07

### Fixed
- **Install**: Fixed HTTPS configuration to use production Let's Encrypt certificates by default. Previously, it defaulted to Staging certificates (untrusted by browsers) due to a missing configuration override.

## [0.6.3] - 2025-12-07

### Fixed
- **Startup**: Fixed crash when running as systemd service. Configuration validation is now deferred until after credentials are loaded from the database, preventing "auth username is required" errors on startup.

## [0.6.2] - 2025-12-07

### Fixed
- **Install**: Fixed `ufw` syntax error by passing arguments correctly to `exec.Command`. This ensures the firewall configuration step works as intended on fresh installs.

## [0.6.1] - 2025-12-07

### Fixed
- **Install**: Fixed "text file busy" error during installation/upgrade on Linux. The installer now detects if the source and target binaries are the same file and skips the copy operation, preventing self-overwrite failures.

## [0.6.0] - 2025-12-03

### Added
- **Architecture**: "Cartridge" Philosophy fully implemented.
    - **One Binary + One DB**: Configuration is now stored in SQLite (`configurations` table).
    - **Config-less**: `config.json` is removed.
    - **Portable**: Run `fazt` anywhere, it uses `./data.db` by default.
- **Install**: Interactive `install.sh` with "Headless Server" vs "Command Line Tool" modes.
- **CLI**:
    - Interactive `fazt server init`.
    - Persistent Client Config (`client.server_url` stored in DB).
    - `fazt deploy` supports full domains (e.g., `my-site.fazt.sh` -> `my-site`).
- **UX**: New gradient ASCII banner.

### Changed
- **Breaking**: `config.json` is no longer supported. Use `fazt server set-config` or `init`.

## [0.5.13] - 2025-12-02

### Fixed
- **Build**: Fixed release artifact naming. Binaries inside tarballs are now consistently named `fazt` (instead of `fazt-linux-amd64`, etc.) to ensure the install script finds them correctly.

## [0.5.12] - 2025-12-02

### Added
- **Performance**: In-memory VFS cache to reduce SQLite reads for frequently accessed files.
- **Routing**: Implemented reserved domains:
    - `admin.<domain>`: Routes to Dashboard.
    - `root.<domain>`: Routes to the "root" site (Welcome Page).
    - `404.<domain>`: Routes to the "404" site (Universal 404).
- **Content**: Embedded "Welcome" and "Universal 404" sites that are automatically seeded on startup.
- **Docs**: Added `docs/index.html` (Landing Page) and `docs/install.sh` (One-line installer).

### Changed
- **Build**: Refactored versioning to use linker flags (`-ldflags`). `cmd/server/main.go` no longer contains the hardcoded version.
- **Architecture**: `localhost` now serves the Dashboard by default, but allows subdomain routing for testing.

## [0.5.11] - 2025-12-02

### Added
- **Install**: Automatically configure UFW firewall (if present) to allow ports 80, 443, and SSH.

## [0.5.10] - 2025-12-02

### Fixed
- **Install**: Fixed "permission denied" error by ensuring the parent `.config` directory is owned by the `fazt` user when created.

## [0.5.9] - 2025-12-02

### Fixed
- **Systemd**: Explicitly set `--config` path in systemd service file to prevent "config not found" errors when `$HOME` is not set correctly by the service runner.

## [0.5.8] - 2025-12-02

### Fixed
- **Build**: Fixed compilation error due to unused import in provision manager.

## [0.5.7] - 2025-12-02

### Added
- **Install**: Added pre-flight check for ports 80 and 443 to detect conflicts with existing web servers (like Nginx/Apache) before installation.

## [0.5.6] - 2025-12-02

### Changed
- **UX**: Completely redesigned the installation experience with beautiful colors, banners, and clearer instructions.
- **CLI**: Improved output formatting for `fazt service install` with a highlighted credentials box.

## [0.5.5] - 2025-12-02

### Fixed
- **Systemd**: Added `AmbientCapabilities` to systemd service definition to ensure the `fazt` process can bind to ports 80/443 even if file capabilities are lost or ignored.

## [0.5.4] - 2025-12-02

### Fixed
- **Migrations**: Fix duplicate `site_logs` table creation in migration 005 (accidental duplicate of 004).

## [0.5.3] - 2025-11-28

### Fixed
- **Embeds**: Fix migration file embedding path to ensure migrations run correctly in production binary.

## [0.5.2] - 2025-11-27

### Added
- **Build**: Added `linux/arm64` build target for Raspberry Pi / ARM servers.

### Fixed
- **Install**: Fixed install script URL construction.

## [0.5.0] - 2025-11-24

### Added

**CLI Improvements**
- `fazt server init` command for first-time server initialization with required credentials and domain
- `fazt server status` command to view current configuration and server state
- `fazt server set-config` command for updating server settings (domain, port, environment)
- `fazt deploy` alias as a top-level shortcut for `fazt client deploy`

**Security Enhancements**
- Authentication now always required (removed optional `auth.enabled` flag)
- Secure by default: auth credentials mandatory in all configurations
- Config files always created with 0600 permissions
- Config directories always created with 0700 permissions

**Configuration Management**
- Removed `auth.enabled` field from config structure
- Improved config validation to always require auth credentials
- Better error messages for missing or invalid configurations
- Unified command structure for better discoverability

### Changed
- `fazt server set-credentials` now focuses only on updating credentials (no longer creates config)
- All server initialization must go through `fazt server init` command
- Config structure simplified: auth section no longer has "enabled" field

### Fixed
- CLI argument parsing for `fazt deploy` alias now works correctly with --help flag
- Integration test suite bugs (arithmetic expressions with set -e, bcrypt pattern matching)
- Consistent error handling across all commands

## [0.2.0] - 2024-11-12

### Added

**Authentication & Security**
- Username/password authentication with bcrypt hashing
- Session management with secure cookies
- Rate limiting (5 attempts per 15 min per IP)
- Audit logging for all security events
- Brute-force protection with automatic lockout
- Login page with modern UI
- Session expiry and refresh
- Remember me functionality (7-day sessions)

**Configuration System**
- JSON-based configuration files
- CLI flags: --config, --db, --port, --username, --password, --env
- Environment-specific configs (development/production)
- Simple credential setup: `./cc-server --username admin --password pass`
- Automatic config directory creation
- Config validation on startup
- Backward compatible with environment variables

**Security Enhancements**
- Security headers (CSP, HSTS, X-Frame-Options, etc.)
- CSRF protection via SameSite cookies
- Automatic file permission enforcement (0600 for config/db)
- Session hijacking prevention
- IP-based rate limiting
- Audit trail for all authentication events

**Database Improvements**
- Migration tracking system
- Automatic backups (keeps last 5)
- New default location: ~/.config/cc/data.db
- Audit logs table
- Migration versioning

**CLI Improvements**
- --version flag with build info
- --help flag with comprehensive documentation
- --verbose and --quiet modes
- Beautiful startup banner
- Production warnings
- Better error messages

**Documentation**
- SECURITY.md - Complete security guide
- CONFIGURATION.md - Configuration reference
- UPGRADE.md - v0.1.0 to v0.2.0 migration guide
- Updated README with v0.2.0 features

**Testing**
- Comprehensive authentication flow test script
- End-to-end testing capabilities

### Changed
- Default config location: ~/.config/cc/config.json
- Default database location: ~/.config/cc/data.db
- Dashboard now protected by authentication (when enabled)
- Improved startup messages with visual banner
- Enhanced Makefile with new targets

### Security
- Dashboard requires authentication (configurable)
- Tracking endpoints remain public
- Protected routes: /, /api/* (except /api/login)
- Public routes: /track, /pixel.gif, /r/*, /webhook/*, /static/*, /login, /health
- File permissions automatically enforced
- Secure cookie defaults in production

### Deprecated
- Environment variables (still work but use JSON config instead)

## [0.1.0] - 2025-11-11

### Added
- Initial release of Command Center
- Universal tracking endpoint with domain auto-detection
- 1x1 transparent GIF pixel tracking
- URL redirect service with click tracking
- Webhook receiver with HMAC SHA256 validation
- Real-time dashboard with interactive charts
- Analytics page with filtering (domain, source, search)
- Redirects management interface
- Webhooks configuration interface
- Settings page with integration snippets
- PWA support with service worker
- Client-side tracking script (track.min.js)
- Light/dark theme toggle with persistence
- SQLite database with WAL mode
- ntfy.sh integration for notifications
- RESTful API with 8 endpoints
- Comprehensive test scripts
- Production-ready deployment configuration

### Features
- **Backend**: Go + SQLite with proper indexing
- **Frontend**: Tabler UI with Chart.js visualizations
- **Database**: 4 tables (events, redirects, webhooks, notifications)
- **API**: Complete CRUD operations for all resources
- **Security**: HMAC validation, input sanitization, prepared statements
- **Performance**: Database indexing, service worker caching, auto-refresh

### Documentation
- Complete README with installation instructions
- API endpoint documentation
- Deployment guide (systemd, nginx)
- Usage examples for all tracking methods
- Troubleshooting section

### Testing
- 4 comprehensive test scripts
- All endpoints tested and validated
- Mock data generator for development

---

**Total Commits**: 13
**Lines of Code**: ~5000+
**Build Time**: ~8 hours (autonomous session)
