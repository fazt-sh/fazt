# Fazt.sh Assistant Guide

**Reference**: `koder/start.md` contains the detailed architectural map.

## 🧠 Core Philosophy
*   **Cartridge Architecture**: One Binary (`fazt`) + One SQLite DB (`data.db`).
*   **Zero Dependencies**: Pure Go + `modernc.org/sqlite`. NO CGO. No external runtimes.
*   **VFS**: User sites/assets live in the DB. No disk I/O for hosting.
*   **System Sites**: Reserved sites (`root`, `404`) MUST be seeded from `go:embed` assets on startup if missing. They live in VFS but originate from the binary.
*   **Safety**: `CGO_ENABLED=0` always.

## 🔨 Build & Test
*   **Build (Local)**: `go build -o fazt ./cmd/server`
*   **Test (All)**: `go test ./...`
*   **Run (Dev)**: `go run ./cmd/server server start --domain localhost --port 8080`

## 📦 Release Workflow
1.  **Code**: Implement feature/fix.
2.  **Test**: `go test ./...` (MUST PASS).
3.  **Changelog**: Update `CHANGELOG.md`.
4.  **Tag**: `git tag vX.Y.Z && git push origin master --tags`.
5.  **Build**: GitHub Action auto-builds release (Version injected via ldflags).

## 📂 Structure
*   `cmd/server/`: Main entry point.
*   `internal/provision/`: Systemd, Install, User management.
*   `internal/hosting/`: VFS, Deploy, CertMagic.
*   `internal/database/`: Migrations, Query logic.
*   `install.sh`: The "curl | bash" installer.

## 🛠️ Environment
*   **Git**: Container has credentials. Can commit & push to `origin/master`.
*   **Note**: `GEMINI.md` is symlinked to this file.