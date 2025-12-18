# Vision Document: Fazt v0.8 - The AI-Ready Personal Cloud OS

**Status**: Draft / Request for Comment (RFC)
**Target**: v0.8 Architecture
**Scope**: High-Level Strategic Pivot

---
# Evolutionary Database Design (EDD) & The Cartridge Kernel

## The Architect's Manifesto

### 1. What is EDD?

EDD is a **non-destructive** schema philosophy.
It treats the database as an **append-only** ledger.
Schemas evolve by adding, never by subtracting or renaming.

### 2. Significance

EDD decouples **Persistence** from **Application Logic**.
It eliminates the "Migration Trap" where old binaries crash on new schemas.
It ensures the DB is a **Universal Recipient** for all code versions.

### 3. Advantages vs. Disadvantages

**Advantages:**

* **Perfect Rollbacks**: Old binaries always work with the current DB.
* **Zero Downtime**: No table locks for complex drops or renames.
* **Solo-Friendly**: Reduces cognitive load and deployment risk.

**Disadvantages:**

* **Schema Bloat**: Unused columns persist indefinitely.
* **Data Integrity**: Must be managed in Go code, not SQL constraints.
* **Messy Tables**: Requires documentation or DB views to clarify.

### 4. Relations in EDD

Relations are **associations**, not **hard enforcements**.
Banish `FOREIGN KEY` constraints at the SQL level.
Use the **Go binary** to validate link integrity during logic cycles.

### 5. How to Enforce It

* **Nullable Only**: Every new field must be optional or have a default.
* **No Drops**: Strictly disable `DROP` and `RENAME` commands.
* **Soft Relations**: Use IDs without hard SQL-level constraints.
* **Version Mapping**: Map code structs to active columns only.

---

## The Cartridge Kernel SDK

### 1. Vision

Extract all "plumbing" from Fazt into a reusable SDK.
Build a "Cloud Kernel" for all future vertical projects.
Separate Infrastructure (Kernel) from specific Business Logic (Verticals).

### 2. Core Kernel Components

* **EDD Persistence**: SQLite wrapper with append-only enforcement.
* **Edge Engine**: CertMagic integration for zero-touch SSL.
* **VFS Manager**: Single-file UI storage and hot-swapping logic.
* **Lifecycle Engine**: Idempotent install and systemd management.
* **Backup Agent**: Embedded Litestream for real-time WAL streaming.

### 3. Proposed Go Package Structure

```text
/pkg/kernel/          # The SDK (Target for GitHub: fazt-sh/kernel)
  ├── edge/           # CertMagic, HTTPS, and host-based routing
  ├── db/             # SQLite Init, WAL tuning, and EDD migrations
  ├── vfs/            # DB-backed file storage and UI serving logic
  ├── provision/      # Systemd, service install, and auto-upgrading
  └── backup/         # Embedded Litestream library and config

/internal/            # Vertical Logic (Specific to Fazt or Cowork)
  ├── models/         # App-specific data structures
  ├── handlers/       # App-specific API endpoints
  └── analytics/      # Specific tracking logic

/cmd/app/             # The Main Entrypoint
  └── main.go         # Imports kernel and registers vertical logic

```

### 4. Architect Strategy for AI Agents

Agents follow the Kernel's rigid architectural safety rails.
This prevents over-engineering of basic infrastructure.
Agents focus entirely on unique business logic within the Vertical.

---

### Advice for the Solo Architect

* **DB is a Bucket**: It holds data; it doesn't judge validity.
* **Logic is the Brain**: Go knows what's valid; SQL does not.
* **Storage is Cheap**: Uptime and your own sanity are expensive.
* **Accept the Mess**: A messy working DB beats a clean dead one.

If you want to move forward, I can generate a **GitHub Action** template to automate the release of this new Kernel SDK. Do you want to see that?
