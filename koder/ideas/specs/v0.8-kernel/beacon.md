# Beacon (Local Discovery)

## Summary

Zero-config local network discovery via mDNS. Find other Fazt nodes on your
LAN when internet/DNS is unavailable. Runs silently in background, provides
peer information to Mesh and other kernel subsystems.

## Why Kernel-Level

Beacon is infrastructure, not a feature:
- Mesh needs to find peers somehow - currently assumes internet
- When internet dies, DNS dies, peer discovery dies
- Beacon provides the fallback: broadcast "I exist" on local network
- Apps don't call Beacon directly - Mesh uses it internally

## The Resilience Contract

```
Internet available:
  Mesh uses internet discovery → Beacon ignored

Internet unavailable, LAN available:
  Mesh asks Beacon → finds local peers → works

Both unavailable:
  Beacon has nothing to find → graceful degradation
```

**Apps don't change.** `fazt.mesh.sync()` just works in more conditions.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        MESH                                 │
│                          │                                  │
│                    getPeers()                               │
│                          │                                  │
│              ┌───────────┴───────────┐                      │
│              ▼                       ▼                      │
│     Internet Discovery          Beacon                      │
│     (DNS, hardcoded)         (mDNS local)                   │
│              │                       │                      │
│              └───────────┬───────────┘                      │
│                          ▼                                  │
│                   Merged peer list                          │
└─────────────────────────────────────────────────────────────┘
```

## Protocol

Uses standard mDNS (RFC 6762) with service type `_fazt._tcp.local`:

```
Announcement:
  _fazt._tcp.local → 192.168.1.10:8080
  TXT: version=0.8, mesh=true, name=my-pi

Discovery:
  Query _fazt._tcp.local → all responding Fazt nodes
```

## Node Record

```javascript
{
  id: 'node_abc123',           // Stable node ID (from kernel identity)
  name: 'my-pi',               // Human-friendly name
  ip: '192.168.1.10',
  port: 8080,
  version: '0.8.0',
  services: ['mesh', 'storage'],  // Available services
  lastSeen: 1705312200000,
  meta: {}                     // Optional metadata
}
```

## Kernel Integration

Beacon hooks into Mesh peer discovery:

```go
// In mesh/peers.go (pseudo-code)
func (m *Mesh) GetPeers() []Peer {
    // Try internet discovery first
    peers := m.internetDiscovery()
    if len(peers) > 0 {
        return peers
    }

    // Fallback to Beacon
    return m.beacon.Discover()
}
```

**Single integration point.** Beacon is otherwise standalone.

## Explicit Usage (Optional)

Apps can query Beacon directly if needed:

```javascript
// Usually not needed - Mesh handles this
const neighbors = await fazt.beacon.discover();
// [{ id, name, ip, port, services, lastSeen }]

// Check if specific service is nearby
const storageNodes = neighbors.filter(n =>
  n.services.includes('storage')
);
```

## CLI

```bash
# See what Beacon is announcing
fazt beacon status
# Announcing: my-pi @ 192.168.1.10:8080
# Services: mesh, storage

# Scan for nearby nodes
fazt beacon scan
# Found 2 nodes:
#   neighbor-a (192.168.1.15) - mesh, storage
#   neighbor-b (192.168.1.22) - mesh

# Set friendly name
fazt beacon set-name "kitchen-pi"
```

## JS API (Explicit, Optional)

```javascript
fazt.beacon.discover(options?)
// options: { timeout: 5000, service: 'mesh' }
// Returns: Promise<Node[]>

fazt.beacon.on('found', handler)
// Live discovery - handler called when new node appears

fazt.beacon.on('lost', handler)
// Handler called when node disappears

fazt.beacon.announce(services)
// Override announced services (usually automatic)
```

## Configuration

```bash
# Beacon is enabled by default
fazt config set beacon.enabled true

# Set announcement name
fazt config set beacon.name "my-pi"

# Disable if you don't want local discovery
fazt config set beacon.enabled false
```

## Storage

Beacon caches discovered nodes for fast startup:

```sql
CREATE TABLE beacon_nodes (
    id TEXT PRIMARY KEY,
    name TEXT,
    ip TEXT NOT NULL,
    port INTEGER NOT NULL,
    version TEXT,
    services TEXT,           -- JSON array
    meta TEXT,               -- JSON
    last_seen INTEGER,
    created_at INTEGER
);

CREATE INDEX idx_beacon_lastseen ON beacon_nodes(last_seen);
```

Cache is advisory - stale entries are pruned, fresh discovery takes precedence.

## Implementation Notes

- Pure Go using `golang.org/x/net/ipv4` for multicast
- ~200 lines of code
- Zero external dependencies
- Runs as goroutine, minimal CPU when idle
- Announces every 30s, listens continuously

## Limits

| Limit              | Default                               |
| ------------------ | ------------------------------------- |
| `maxNodes`         | 100 (cached)                          |
| `announceInterval` | 30s                                   |
| `discoveryTimeout` | 5s                                    |
| `nodeExpiry`       | 5 minutes (no announcement = removed) |

## Security Considerations

- Beacon only announces presence, not data
- No authentication at discovery layer (Mesh handles auth)
- Nodes verify identity via kernel signatures before trusting
- Beacon can be disabled for high-security environments
