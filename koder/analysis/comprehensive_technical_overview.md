# Fazt.sh: Comprehensive Technical Overview

**Date**: December 7, 2025
**Version**: 0.4.0 (Dev)
**Target Audience**: Senior Engineers / Architectural Reviewers

## 1. Executive Summary

**Fazt.sh** is a "Personal PaaS" (Platform as a Service) designed around a radical philosophy of simplicity and portability. Unlike traditional hosting platforms that rely on container orchestration (Kubernetes, Docker) or complex microservices, Fazt collapses the entire stack into a **Single Binary + Single Database** architecture.

It allows a user to deploy static sites and serverless functions to a single low-cost VPS (e.g., $6/mo DigitalOcean Droplet) with zero external dependencies.

## 2. Core Philosophy & Architectural Constraints

### 2.1 The "Cartridge" Concept
The architecture mimics a game console cartridge:
*   **The Console**: The `fazt` binary (stateless logic, runtime).
*   **The Cartridge**: The `data.db` SQLite file (state, filesystem, config).
*   **Result**: To migrate a server, you simply copy `data.db` to a new machine.

### 2.2 Zero External Dependencies
*   **No CGO**: The project is built with `CGO_ENABLED=0` to ensure static linking and cross-platform compatibility (Linux/amd64, Linux/arm64, macOS).
*   **No Runtime Dependencies**: It does not require Node.js, Python, Nginx, or Docker to be installed on the host.
*   **Pure Go**: All functionality, including the JS runtime and SQLite driver, is implemented in pure Go.

## 3. System Architecture

### 3.1 High-Level Blocks
```mermaid
graph TD
    User[User / Internet] --> |HTTPS :443| Server[Fazt Server (Go)]
    
    subgraph "Fazt Server Process"
        Router[Host Router]
        Dashboard[Admin Dashboard]
        Hosting[VFS Site Handler]
        Runtime[JS Runtime (Goja)]
        CertMagic[Auto-TLS (CertMagic)]
    end
    
    subgraph "Storage (SQLite)"
        TableFiles[table: files (VFS)]
        TableKV[table: kv_store]
        TableConfig[table: config]
        TableLogs[table: site_logs]
    end
    
    User --> Router
    Router --> |admin.domain| Dashboard
    Router --> |*.domain| Hosting
    
    Hosting --> |Static File| TableFiles
    Hosting --> |main.js| Runtime
    
    Runtime --> |db.get/set| TableKV
    Runtime --> |fetch| ExternalAPI[External APIs]
    
    Dashboard --> TableConfig
    CertMagic --> |Cert Storage| TableFiles
```

### 3.2 Key Components

#### A. Virtual Filesystem (VFS)
*   **Implementation**: Instead of storing user sites on the host's ext4/xfs filesystem, Fazt stores files as BLOBs in the `files` table in SQLite.
*   **Schema**: `files(site_id, path, content, size, hash, mime_type, updated_at)`.
*   **Performance**: Uses an in-memory LRU cache (`internal/hosting/vfs.go`) to avoid hitting the DB for hot assets (CSS/JS/HTML).
*   **Rationale**: Enables atomic backups (snapshotting the DB snapshots the filesystem) and prevents inode exhaustion on small VPSs.

#### B. Serverless Runtime
*   **Engine**: `dop251/goja` (ECMAScript 5.1(+) implementation in pure Go).
*   **Trigger**: If a site contains `main.js`, it is executed for every request.
*   **Capabilities**:
    *   `req` / `res`: Express-like API.
    *   `db`: Key-Value store access (`kv_store` table).
    *   `fetch`: HTTP client with SSRF protection.
    *   `socket`: WebSocket broadcasting.
*   **Constraints**: 100ms execution timeout, 1MB body limit.

#### C. Routing & Domain Management
*   **Library**: `caddyserver/certmagic`.
*   **Function**: Automatically manages Let's Encrypt / ZeroSSL certificates.
*   **Storage**: Certificates are stored in the SQLite DB (custom implementation of CertMagic storage interface).
*   **Logic**:
    *   `admin.<domain>` -> Dashboard (Authentication required).
    *   `root.<domain>` / `<domain>` -> System Landing Page.
    *   `*.<domain>` -> User Sites (VFS).

#### D. Database & State
*   **Library**: `modernc.org/sqlite` (CGO-free SQLite).
*   **Migrations**: SQL files embedded in binary, applied on startup (`internal/database/migrations/`).
*   **Tables**:
    *   `files`: Hosting content.
    *   `kv_store`: User-land persistence.
    *   `events`: Analytics (page views).
    *   `site_logs`: `console.log` capture from JS runtime.
    *   `config`: Server configuration (Port, Auth, Env).

## 4. Technical Specifications

| Component | Technology | Reasoning |
| :--- | :--- | :--- |
| **Language** | Go 1.24+ | Concurrency, Static Binary, Tooling. |
| **Database** | SQLite (modernc) | Zero config, file-based, CGO-free. |
| **HTTP Server** | `net/http` | Standard library is robust enough. |
| **HTTPS** | `certmagic` | Best-in-class ACME implementation. |
| **JS Runtime** | `goja` | Pure Go, safe sandboxing (unlike V8/cgo). |
| **WebSockets** | `gorilla/websocket` | Standard for Go WS. |
| **Password Hashing** | `bcrypt` | Standard security practice. |

## 5. Deployment & Operations

### 5.1 Installation
A single shell script (`install.sh`) handles:
1.  Downloading the binary.
2.  Creating a system user (`fazt`).
3.  Setting up `systemd` service.
4.  Configuring `setcap` for binding to port 443 without root.

### 5.2 Deployment Protocol (Client -> Server)
1.  **Client**: Zips the target directory.
2.  **Client**: `POST /api/deploy` with Bearer token.
3.  **Server**: Unzips in memory.
4.  **Server**: Calculates SHA256 hashes of files.
5.  **Server**: Upserts into `files` table (deduplication at file path level).
6.  **Server**: Invalidates VFS cache for that site.

## 6. Current Limitations & Known Issues

1.  **No "Real" Multi-Tenancy**: Designed for a single "owner" hosting many sites.
2.  **JS Runtime Version**: ES5.1 (Goja) lacks modern JS features (async/await, ES6 modules) without transpilation.
3.  **Database Locking**: SQLite writer lock can bottleneck if high write concurrency (WAL mode enabled to mitigate).
4.  **Memory Usage**: VFS Cache and Goja VMs can consume RAM; strict limits are enforced but tuning is manual.

## 7. Future Directions (Roadmap Alignment)

*   **Core Config DB**: Finalize migration of all config to DB (removing `config.json` entirely).
*   **JS Cron**: Scheduled serverless tasks.
*   **Backup/Restore**: CLI commands to safely snapshot the running DB.
*   **App Store**: Git-based "Click to Install" for standardized apps (e.g., specific `fazt` compatible repos).
