# Fazt - Session Bootstrap

**Fazt**: Sovereign compute infrastructure in one Go binary. Runs on anything
from phones to servers—your personal swarm of AI-native compute nodes.

## Initialize
1. Read `koder/philosophy/VISION.md` for strategic context (if new to Fazt)
2. Read `koder/NEXT_SESSION.md` for current task
3. Load the plan referenced there
4. Execute

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
