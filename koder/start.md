# Fazt.sh - Session Bootstrap

**Fazt**: Personal PaaS in one Go binary. Static sites + serverless on your own VPS.

## Initialize
1. Read `koder/NEXT_SESSION.md` for current task
2. Load the plan referenced there
3. Execute

## Architecture (Reference)
- **Binary**: `fazt` (Go, CGO_ENABLED=0)
- **Database**: `data.db` (SQLite via modernc.org/sqlite)
- **VFS**: Sites stored in `files` table
- **Admin**: React SPA at `admin/`, embedded in binary
- **Routing**: Host-based (`admin.`, `*.domain`)

## Commands
```bash
# Go backend
go build -o fazt ./cmd/server
go test ./...

# Admin SPA
cd admin && npm run dev -- --port 37180
cd admin && npm run build
```
