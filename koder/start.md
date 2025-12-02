# Bootstrap: Fazt.sh Engineering

**Welcome.** This is a single-binary PaaS designed to be self-contained, robust, and zero-dependency.

## 1. The "Cartridge" Model
Think of `fazt` like a game cartridge.
*   **The Console**: The `fazt` binary. Contains all logic, UI templates, and migrations.
*   **The Save File**: `data.db` (SQLite). Contains all user data, configuration, SSL certs, and hosted files.
*   **Upgrade**: Replace the binary, keep the DB. Downgrade? Replace the binary, keep the DB.

## 2. Critical Rules
1.  **NO CGO**: We use `modernc.org/sqlite` to ensure static binaries work on any Linux dist (Alpine, Scratch, Debian).
2.  **No Disk I/O**: User sites are uploaded via ZIP, extracted into RAM, and stored in SQLite blobs. We do NOT write user files to disk.
3.  **Idempotency**: The `install.sh` and `fazt service install` commands must be safe to run multiple times.

## 3. Key Workflows

### A. Developing
```bash
# Run locally with hot-ish reload (go run)
go run ./cmd/server server init --username admin --password secret --domain localhost
go run ./cmd/server server start
```

### B. Testing
We rely on standard Go testing.
```bash
go test ./...
```
*Always* run this before pushing.

### C. Releasing
We use semantic versioning (vX.Y.Z).
1.  Update `const Version` in `cmd/server/main.go`.
2.  Update `CHANGELOG.md`.
3.  Commit: `git commit -m "chore(release): vX.Y.Z"`
4.  Tag: `git tag vX.Y.Z`
5.  Push: `git push origin master --tags`

## 4. Where is...?
*   **Systemd/Install Logic**: `internal/provision/`
*   **Web Server/VFS**: `internal/hosting/`
*   **Database/Migrations**: `internal/database/`
*   **CLI Commands**: `cmd/server/main.go`