# Command Center - Phase 3: Personal PaaS Upgrade

## Project Overview
- **Current Version**: v0.2.0 (Auth + Analytics)
- **Target Version**: v0.3.0
- **Goal**: Transform CC into a "Personal Cloud" (Surge clone + Serverless Functions)
- **Constraint**: Single binary, SQLite backed, Local filesystem storage.

---

## Architecture Changes
1.  **Storage**: New directory structure `~/.config/cc/sites/{site_id}/`.
2.  **Routing**: `main.go` must switch from simple Mux to a Host-based router to handle `*.domain.com`.
3.  **Database**: New tables for `api_keys` and `kv_store`.
4.  **Runtime**: Embed `goja` (JS VM) for serverless functions.

---

## Phase 1: Database Expansion (Commit #1)

**Task**: Prepare the database for multi-tenancy and hosting.

1.  Create `migrations/002_paas.sql`:
    ```sql
    -- API Keys for deploying sites (for friends/CLI)
    CREATE TABLE api_keys (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL, -- e.g. "Friend Bob"
        key_hash TEXT NOT NULL, -- bcrypt hash of the token
        scopes TEXT, -- JSON array e.g. ["deploy:blog", "read:stats"]
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        last_used_at DATETIME
    );

    -- Key-Value store for Serverless Apps
    CREATE TABLE kv_store (
        site_id TEXT NOT NULL,
        key TEXT NOT NULL,
        value TEXT, -- JSON or text
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (site_id, key)
    );
    
    -- Track deployments
    CREATE TABLE deployments (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        site_id TEXT NOT NULL,
        size_bytes INTEGER,
        file_count INTEGER,
        deployed_by TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    ```
2.  Update `internal/database/db.go`:
    - Add `002_paas.sql` to the migration list.
    - Ensure migrations run on startup.

**Verification**: Run `./cc-server`, check `cc.db` schema using `sqlite3`.

---

## Phase 2: Host-Based Routing (Commit #2)

**Task**: Refactor `main.go` to handle Subdomains vs Dashboard.

1.  Refactor `cmd/server/main.go`:
    - Create `rootHandler(w, r)` function.
    - Logic:
      - If `Host == cfg.Server.Domain` (or localhost port): Serve Dashboard (existing logic).
      - If `Host == *.cfg.Server.Domain` (or *.localhost): Serve User Site.
2.  Extract existing dashboard routes into a `dashboardMux`.
3.  Create a placeholder `siteHandler` that returns "404 Site Not Found" for now.

**Verification**:
- `curl localhost:4698` -> Returns Dashboard HTML.
- `curl foo.localhost:4698` -> Returns "404 Site Not Found" (but proves routing works).

---

## Phase 3: Site Storage & Static Serving (Commit #3)

**Task**: Serve static files from the filesystem based on subdomain.

1.  Create `internal/hosting/manager.go`:
    - Define data dir: `~/.config/cc/sites/`.
    - Helper `GetSiteDir(subdomain string)`.
2.  Update `siteHandler` in `main.go`:
    - Extract subdomain (e.g., `blog` from `blog.x8.sh`).
    - Check if directory exists in `sites/`.
    - If yes: Use `http.FileServer` to serve the directory.
    - If no: Return 404.
3.  **Analytics Integration**:
    - Add middleware to `siteHandler` that inserts a row into `events` table with `source_type='hosting'` and `domain={subdomain}`.

**Verification**:
- Manually create `~/.config/cc/sites/test/index.html`.
- Access `http://test.localhost:4698`.
- Verify the page loads AND an event appears in the DB.

---

## Phase 4: Deploy API (Commit #4)

**Task**: Allow uploading sites via API (ZIP file).

1.  Create handler `POST /api/deploy` (protected).
2.  Input: Multipart form (file: `site.zip`, field: `site_name`, header: `Authorization: Bearer <token>`).
3.  Logic:
    - Validate Token (check `api_keys` table).
    - Validate `site_name` (alphanumeric).
    - Save ZIP to temp.
    - Unzip to `~/.config/cc/sites/{site_name}/` (clean old files first).
    - Record in `deployments` table.
4.  Security: Ensure unzip doesn't allow path traversal (`../../`).

**Verification**: Use `curl` to upload a zip file. Check if files appear in the config directory.

---

## Phase 5: CLI Deploy Command (Commit #5)

**Task**: Add a subcommand to the single binary to push sites.

1.  Update `cmd/server/main.go` or create `cmd/cli/`:
    - Check `os.Args`. If `cc-server deploy <site-name>`:
    - Zip current directory in memory.
    - Read token from `~/.cc-token` or flag.
    - POST to the server.
    - Stream response logs to stdout.

**Verification**:
- `cd my-website/`
- `../../cc-server deploy my-site`
- Output should show "Deployment successful".

---

## Phase 6: API Key Management UI (Commit #6)

**Task**: Dashboard interface to generate keys for friends.

1.  New Dashboard Page: **"Hosting"**.
2.  List active Sites (folders in `sites/`).
3.  Section "Deploy Keys":
    - Button "Generate New Key".
    - Input: Name (e.g., "Mike").
    - Action: Generate random string, hash it, store in DB, show *once* to user.
    - List active keys with "Revoke" button.

**Verification**: Generate a key in UI, use it with the CLI from Phase 5.

---

## Phase 7: Serverless Engine - Goja Integration (Commit #7)

**Task**: Embed JS runtime.

1.  Add dependency: `github.com/dop251/goja`.
2.  Create `internal/hosting/runtime.go`.
3.  Function `RunServerless(w, r, siteDir)`:
    - Check if `main.js` exists in `siteDir`.
    - If yes, create Goja VM.
    - Inject `req` object (method, headers, body).
    - Inject `res` object (send, json, status).
    - Run script.

**Verification**:
- Create `main.js`: `res.send("Hello from JS!");`
- Deploy.
- Visit site.

---

## Phase 8: Serverless KV Store (Commit #8)

**Task**: Give the JS runtime a database.

1.  Update `internal/hosting/runtime.go`:
    - Inject `db` object into Goja.
    - `db.get(key)` -> SELECT from `kv_store`.
    - `db.set(key, value)` -> INSERT/UPDATE `kv_store`.
    - Enforce `site_id` scope (Site A cannot read Site B's data).
2.  Update `db.set` to be atomic/safe.

**Verification**:
- Create `main.js`:
  ```js
  let count = db.get("visits") || 0;
  count++;
  db.set("visits", count);
  res.json({ visits: count });
  ```
- Refresh page multiple times. Count should go up.

---

## Phase 9: Env Vars & Secrets (Commit #9)

**Task**: Allow sites to have secrets (API Keys for OpenAI, etc).

1.  Add `env_vars` table to DB or use `kv_store` with a prefix.
2.  UI: Add "Environment Variables" to Hosting page.
3.  Goja: Inject `process.env` populated from DB.

**Verification**: Set `API_KEY` in UI. Access `process.env.API_KEY` in JS.

---

## Phase 10: WebSocket Support (Commit #10)

**Task**: Realtime support.

1.  This is tricky in Goja, but for "Standard" role, use Go's native WS.
2.  Add a generic WS hub in `internal/hosting/ws.go`.
3.  If `main.js` defines `socket.on('message')`, bridge it?
4.  *Simplification*: For v0.3.0, allow a simple "Broadcast" channel.
    - JS: `socket.broadcast("hello")`.
    - Clients connected to `wss://site.domain/ws` receive it.

**Verification**: Simple chat app demo.

---

## Phase 11: Security Hardening (Commit #11)

**Task**: Sandboxing.

1.  **Time Limit**: Use Goja's `Interrupt` to kill scripts taking > 100ms.
2.  **Memory Limit**: Goja doesn't strictly limit memory easily, but we can monitor.
3.  **Path Traversal**: Ensure static file server never leaves `site/{id}`.

---

## Phase 12: Documentation & Final Polish (Commit #12)

1.  Update `README.md` with "Personal Cloud" instructions.
2.  Create `examples/` folder with:
    - `static-site/`
    - `counter-app-js/`
    - `chat-app/`
3.  Final full system test.

---
