# Plan: Admin Dashboard VFS Refactor ("Cartridge Admin")

## Objective
Move the Admin Dashboard (UI) from the binary's embedded `web/` directory into the SQLite VFS, treating it like any other hosted site. This aligns with the "One Binary + One DB" philosophy, enables "hot" UI updates via deployment, and simplifies the codebase.

## Core Changes
1.  **Storage**: Move assets from `web/{static,templates}` to `internal/assets/system/admin`.
2.  **Seeding**: Update `EnsureSystemSites` to seed the `admin` site into VFS on startup (if missing).
3.  **Routing**: Update `admin.<domain>` routing to use the generic VFS handler (with Auth Middleware).
4.  **Frontend**: Refactor HTML templates into a **Static SPA** (Single Page App). Remove server-side `{{.}}` rendering; use API calls.
5.  **CLI**: Add `fazt server reset-admin` to force-update the VFS admin site from the binary.

## Phases

### Phase 1: Frontend Refactor (Static Conversion)
Convert the existing Go templates into pure HTML/JS that fetches data from APIs.
- [ ] **Login Page**: Convert `login.html` to use AJAX/Fetch for login. Handle errors via JS.
- [ ] **Dashboard**: Convert `index.html` to remove `{{.Version}}`, `{{.Username}}`. Fetch from API.
- [ ] **Hosting**: Convert `hosting.html`.
- [ ] **API Gaps**: Create/Update endpoints for missing data (e.g., `/api/user/me` for username/version).

### Phase 2: Asset Migration & Seeding
Move files to the system assets folder and implement seeding.
- [ ] **Move Files**: `web/static/**` -> `internal/assets/system/admin/**`.
- [ ] **Move HTML**: `web/templates/*.html` -> `internal/assets/system/admin/*.html`.
- [ ] **Update Embeds**: Update `internal/assets/assets.go` to embed `system/admin`.
- [ ] **Update Seeder**: Modify `internal/hosting/manager.go` to seed `admin`.

### Phase 3: Routing & Backend
Switch the server to serve Admin from VFS.
- [ ] **Routing**: In `cmd/server/main.go`, change `admin.` route to use `siteHandler` (wrapped in Auth).
- [ ] **Cleanup**: Remove old `http.FileServer` logic and `dashboardMux` static handlers.
- [ ] **Security**: Ensure API handlers (`/api/*`) are still attached and protected.

### Phase 4: CLI & Polish
- [ ] **Reset Command**: Implement `fazt server reset-admin`.
- [ ] **CSP**: Fix Content Security Policy for the new structure (and CDN usage).
- [ ] **Verify**: Test Login, Dashboard, Hosting, and Deploy flows.

## Todo List
(Will be populated in the next step via `write_todos`)
