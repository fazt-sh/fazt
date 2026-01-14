# Plan 17: Peers Table & Remote Commands

**Goal**: Enable fazt-to-fazt communication with all config in SQLite.
**Principle**: Move DB to another system → everything works.

## Schema

### Migration: `009_peers.sql`

```sql
-- Known remote fazt nodes
CREATE TABLE IF NOT EXISTS peers (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(8)))),
    name TEXT UNIQUE NOT NULL,           -- Human name: "zyt", "home", "pi"
    url TEXT NOT NULL,                   -- Admin URL: https://admin.zyt.app
    token TEXT,                          -- API key (encrypted at rest later)

    -- Metadata
    description TEXT,                    -- "Personal server", "Raspberry Pi"
    is_default INTEGER DEFAULT 0,        -- Only one can be default

    -- Connection state (updated on use)
    last_seen_at TEXT,                   -- Last successful contact
    last_version TEXT,                   -- Last known fazt version
    last_status TEXT,                    -- "healthy", "unreachable", etc.

    -- Future: Mesh identity
    node_id TEXT,                        -- Unique node identifier
    public_key TEXT,                     -- For encrypted mesh communication

    -- Timestamps
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

-- Only one default peer
CREATE UNIQUE INDEX IF NOT EXISTS idx_peers_default
    ON peers(is_default) WHERE is_default = 1;

-- Fast lookup by name
CREATE INDEX IF NOT EXISTS idx_peers_name ON peers(name);
```

## CLI Commands

### `fazt remote` subcommand

```
fazt remote                     # List known peers
fazt remote add <name> <url>    # Add peer
fazt remote remove <name>       # Remove peer
fazt remote default <name>      # Set default peer
fazt remote status [name]       # Check peer health
fazt remote apps [name]         # List apps on peer
fazt remote upgrade [name]      # Upgrade peer
fazt remote deploy <dir> [name] # Deploy to peer
```

### Examples

```bash
# Add a peer
fazt remote add zyt https://admin.zyt.app --token xxx

# Set as default
fazt remote default zyt

# Check status (uses default)
fazt remote status

# List apps on specific peer
fazt remote apps zyt

# Deploy to peer
fazt remote deploy ./my-site zyt
```

## Internal Package

### `internal/remote/`

```go
// peer.go - Peer management
type Peer struct {
    ID          string
    Name        string
    URL         string
    Token       string
    Description string
    IsDefault   bool
    LastSeenAt  *time.Time
    LastVersion string
    LastStatus  string
}

func AddPeer(db *sql.DB, name, url, token string) error
func RemovePeer(db *sql.DB, name string) error
func GetPeer(db *sql.DB, name string) (*Peer, error)
func GetDefaultPeer(db *sql.DB) (*Peer, error)
func ListPeers(db *sql.DB) ([]Peer, error)
func SetDefaultPeer(db *sql.DB, name string) error

// client.go - HTTP client for peer communication
type Client struct {
    peer *Peer
    http *http.Client
}

func NewClient(peer *Peer) *Client
func (c *Client) Status() (*StatusResponse, error)
func (c *Client) Apps() ([]App, error)
func (c *Client) Deploy(zipPath, siteName string) error
func (c *Client) Upgrade(checkOnly bool) (*UpgradeResponse, error)

// Later: transport abstraction for mesh
type Transport interface {
    Do(req *Request) (*Response, error)
}
type HTTPTransport struct { ... }
type MeshTransport struct { ... }  // v0.16
```

## Migration Path

### From `~/.fazt/config.json`

```bash
# On first run after upgrade, if old config exists:
# 1. Read ~/.fazt/config.json
# 2. Insert each server into peers table
# 3. Rename to ~/.fazt/config.json.migrated
# 4. Print notice to user
```

```go
func MigrateOldConfig(db *sql.DB) error {
    oldPath := filepath.Join(os.Getenv("HOME"), ".fazt", "config.json")
    if _, err := os.Stat(oldPath); os.IsNotExist(err) {
        return nil // Nothing to migrate
    }

    // Read old config
    data, _ := os.ReadFile(oldPath)
    var old OldConfig
    json.Unmarshal(data, &old)

    // Insert peers
    for name, server := range old.Servers {
        AddPeer(db, name, server.URL, server.Token)
        if name == old.DefaultServer {
            SetDefaultPeer(db, name)
        }
    }

    // Rename old file
    os.Rename(oldPath, oldPath+".migrated")
    log.Printf("Migrated %d servers to peers table", len(old.Servers))
    return nil
}
```

## Database Location

**Question**: Which database for client-only usage?

**Options**:

A. **Require local fazt instance**: Must run `fazt server init` first
   - Pro: Consistent model, one DB
   - Con: Heavy for just client usage

B. **Auto-create minimal client.db**: `~/.config/fazt/data.db`
   - Pro: Works without full server setup
   - Con: Two possible DB locations

C. **Unified approach**:
   - `fazt` always has a local DB (even if not running as server)
   - Default: `~/.config/fazt/data.db`
   - Server mode: `./data.db` or specified path
   - Same schema, same tables

**Recommendation**: Option C - unified approach

```go
func GetDefaultDBPath() string {
    // Server mode (has WorkingDirectory set)
    if isServerMode() {
        return "./data.db"
    }
    // Client mode
    return filepath.Join(xdg.ConfigHome, "fazt", "data.db")
}
```

## Portability

**Copy DB → Everything works:**

```bash
# On machine A
fazt remote add zyt https://admin.zyt.app --token xxx
fazt remote add home https://home.local:8080 --token yyy

# Copy to machine B
scp ~/.config/fazt/data.db machineB:~/.config/fazt/

# On machine B - just works
fazt remote status zyt   # ✓
fazt remote apps home    # ✓
```

## Implementation Order

1. **Migration 009**: Create peers table
2. **internal/remote/peer.go**: CRUD operations
3. **internal/remote/client.go**: HTTP client
4. **cmd/server/remote.go**: CLI commands
5. **Migration logic**: Old config → peers table
6. **Update clientconfig**: Deprecate, use peers table
7. **Tests**: Peer operations, client mocking

## Future: Mesh Integration (v0.16)

The peers table is ready for mesh:

```sql
-- Already in schema:
node_id TEXT,      -- Populated when mesh handshake occurs
public_key TEXT,   -- For encrypted P2P communication
```

```go
// Transport abstraction
type MeshTransport struct {
    nodeID    string
    publicKey []byte
}

func (t *MeshTransport) Do(req *Request) (*Response, error) {
    // Use WireGuard/libp2p instead of HTTP
}
```

Same `fazt remote status zyt` command, different transport underneath.

## Files to Create/Modify

| File | Action |
|------|--------|
| `internal/database/migrations/009_peers.sql` | Create |
| `internal/remote/peer.go` | Create |
| `internal/remote/client.go` | Create |
| `internal/remote/migrate.go` | Create |
| `cmd/server/remote.go` | Create |
| `internal/database/db.go` | Add migration 9 |
| `internal/clientconfig/` | Deprecate |

## Version

This would be **v0.9.0** - significant new capability.
