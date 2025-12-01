# Fazt.sh - Single Binary PaaS

**Core Goal:** A zero-dependency "Cartridge" application (One Binary + One SQLite DB).

## 🚀 Key Commands
*   **Build (Linux)**: `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o fazt ./cmd/server`
*   **Build (Mac)**: `CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o fazt ./cmd/server`
*   **Dev Run**: `go run ./cmd/server server start --port 8080 --domain localhost`
*   **Test**: `go test ./...`

## 🏗 Architecture (Must Read)
1.  **Pure Go**: We use `modernc.org/sqlite`. **NO CGO**. This allows static cross-compilation.
2.  **Embedded Assets**: `migrations/` and `web/` are embedded via `//go:embed`.
    *   *Rule*: Never use `os.ReadFile` for core assets; use `assets.WebFS` or `database.migrationFS`.
3.  **VFS**: User sites live in SQLite (`files` table). **NO DISK I/O** for user content.
4.  **CLI Structure**:
    *   `fazt server ...`: Runtime (start app, init config).
    *   `fazt service ...`: System Ops (install systemd, logs).
    *   `fazt client ...`: User tools (deploy site).

## 📂 Project Map
*   `cmd/server/`: Entry point & CLI parsing.
*   `internal/provision/`: Systemd/User/Install logic.
*   `internal/hosting/`: VFS & Deploy logic.
*   `internal/assets/`: Embedded `web/` assets.
*   `internal/database/`: SQLite conn & Embedded `migrations/`.

## ⚠️ Critical Constraints
*   **Do not re-introduce CGO**: Keep `CGO_ENABLED=0` capability.
*   **Do not write to disk**: Except `data.db` and `config.json`.
*   **Maintain Install Flow**: `service install` must remain idempotent and single-command.