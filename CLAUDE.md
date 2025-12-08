# Fazt.sh Assistant Guide

**Reference**: `koder/start.md` contains the detailed architectural map.

## 🧠 Core Philosophy
*   **Cartridge Architecture**: One Binary (`fazt`) + One SQLite DB (`data.db`).
*   **Zero Dependencies**: Pure Go + `modernc.org/sqlite`. NO CGO. No external runtimes.
*   **VFS**: User sites/assets live in the DB. No disk I/O for hosting.
*   **System Sites**: Reserved sites (`root`, `404`) MUST be seeded from `go:embed` assets on startup if missing. They live in VFS but originate from the binary.
*   **Safety**: `CGO_ENABLED=0` always.

## 🔨 Build & Test
*   **Test First Methadology**: When build a feature, implement all tests first
    and build to pass the tests
*   **Build (Local)**: `go build -o fazt ./cmd/server`
*   **Test (All)**: `go test ./...`
*   **Run (Dev)**: `go run ./cmd/server server start --domain localhost --port 8080`
*   **Env Differences**: Recognise what can be tested in the coding environment
    and build tests likewise

### ⚠️ Testing Server Stability
When testing if the server runs without crashing:
*   **DO NOT** use bash backgrounding (`&`) to test server stability
    - Backgrounded processes appear to "work" but you cannot detect silent exits
*   **PREFERRED METHODS**:
    1. Use `timeout 10 go run ./cmd/server server start` to test for N seconds
    2. Run in foreground and check logs for panics/errors
    3. Use `ps aux | grep server` AFTER a delay to verify process persistence
*   **CORRECT PATTERN**:
    ```bash
    # Start server
    go run ./cmd/server server start --port 8080 &

    # Wait for startup (3-5 seconds)
    sleep 3

    # Verify it's still running
    ps aux | grep "server start" | grep -v grep

    # Test endpoints
    curl http://localhost:8080/health

    # Clean up
    pkill -f "server start"
    ```

## 📦 Release Workflow
**Detailed Guide**: `koder/workflows/ON_NEW_VERSION.md`

1.  **Code**: Implement feature/fix.
2.  **Test**: `go test ./...` (MUST PASS).
3.  **Changelog**: Update `CHANGELOG.md` AND `docs/changelog.json`.
4.  **Tag**: `git tag vX.Y.Z && git push origin master --tags`.
5.  **Build**: GitHub Action auto-builds release (Version injected via ldflags).

## 📂 Structure
*   `cmd/server/`: Main entry point.
*   `internal/provision/`: Systemd, Install, User management.
*   `internal/hosting/`: VFS, Deploy, CertMagic.
*   `internal/database/`: Migrations, Query logic.
*   `install.sh`: The "curl | bash" installer.
*   `koder/docs/admin-api/request-response.md`: **Ground Truth** API documentation (Real request/response examples).

## 🛠️ Environment
*   **Git**: Container has credentials. Can commit & push to `origin/master`.
*   **Container Setup**: You are running inside a podman container, with no
    systemd;
*   **Host**: Host machine is a mac m4
*   **Deplayment Target**: We are deploying to Digital Ocean Droplet, x86/Ubuntu
    24.04, LTS, $6 base instance
*   **Note**: `GEMINI.md` is symlinked to this file.

## Important

- fazt.sh is just a proposed url; its NOT LIVE, don't use it anywhere
