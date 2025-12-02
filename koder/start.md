# Fazt.sh - AI Bootloader

This document provides context to bootstrap a new coding session.
Read carefully.

## 1. Core Philosophy 🧠
- **One Binary**: Single executable (`fazt`) + Single DB (`data.db`).
- **Zero Dependencies**: No CGO. No external runtimes. Pure Go + ModernC SQLite.
- **Cartridge Architecture**: The DB is the filesystem. Sites live in SQL.
- **Safety**: `CGO_ENABLED=0` always. Secure by default.

## 2. Architecture Pillars 🏛️
- **VFS (Virtual Filesystem)**:
  - All site assets stored in `files` table (blob).
  - In-memory LRU cache for performance (`internal/hosting/vfs.go`).
- **System Sites**:
  - Reserved domains (`root`, `404`) are seeded from binary.
  - Source: `internal/assets/system/`.
  - Logic: `internal/hosting/manager.go:EnsureSystemSites`.
- **Runtime**:
  - `goja` JS runtime for serverless functions (`main.js`).
  - Supports `fetch`, `db.get/set` (KV), `socket` (WS).
- **Routing**:
  - `admin.<domain>` -> Dashboard.
  - `root.<domain>` / `<domain>` -> Welcome Site.
  - `404.<domain>` -> Universal 404.
  - `*.<domain>` -> User Sites.

## 3. Codebase Map 🗺️
- `cmd/server/`       : Main entry point. Versioning, Routing logic.
- `internal/hosting/` : VFS, Deployment (Zip->DB), Runtime, CertMagic.
- `internal/database/`: SQLite Init, Migrations (`001_...sql`).
- `internal/config/`  : JSON/Env config loader.
- `internal/assets/`  : Embedded static files (Dashboard + System Sites).
- `docs/`             : Public documentation (Lander, Install Script).
- `koder/`            : Meta-docs for AI (Plans, Ideas, Analysis).

## 4. Initialization Protocol 🚀
To fully understand the current state, perform these actions:

1. **Read Core Logic**:
   - `read_file internal/hosting/vfs.go` (Storage engine)
   - `read_file cmd/server/main.go` (Routing & Entry)
   - `read_file internal/hosting/manager.go` (Seeding & Ops)

2. **Read Meta-Context**:
   - `read_file GEMINI.md` (Operational Guidelines)
   - `read_file koder/ideas/01_random.md` (Roadmap & Backlog)

3. **Verify State**:
   - `run_shell_command "make test"` (Ensure stability)

## 5. Session Goal 🎯
After reading the above, output a concise summary of:
1. The architectural state you found.
2. The next item on the roadmap (`<next-to-implement>`).
3. Confirmation that you are ready to code.
