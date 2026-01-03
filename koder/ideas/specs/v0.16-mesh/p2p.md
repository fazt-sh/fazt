# Kernel Mesh

## Summary

Transparent synchronization between Fazt nodes. Run a local instance on your
laptop and a remote on your VPS—changes sync automatically.

## Use Case

```
┌─────────────────┐         ┌─────────────────┐
│  Laptop Fazt    │◄───────►│   VPS Fazt      │
│  (local dev)    │  sync   │  (production)   │
└─────────────────┘         └─────────────────┘
```

Edit locally with zero latency. Changes appear remotely within seconds.

## CLI

```bash
# Initialize mesh identity
fazt mesh init

# Join another node
fazt mesh join https://vps.example.com:8443 --token xxx

# Check sync status
fazt mesh status

# Force sync
fazt mesh sync

# Leave mesh
fazt mesh leave
```

## Sync Protocol

### Change Detection

Uses SQLite triggers to detect changes:

```sql
CREATE TRIGGER track_changes AFTER INSERT ON files
BEGIN
    INSERT INTO sync_log (table_name, row_id, operation, timestamp)
    VALUES ('files', NEW.id, 'INSERT', unixepoch());
END;
```

### Gossip

Changes propagate via gossip protocol:

1. Node A detects local change
2. A sends change summary to known peers
3. Peers request full change if needed
4. Changes merge (CRDTs for conflict resolution)

## JS Runtime

```javascript
// List mesh peers
const peers = await fazt.mesh.peers();

// Broadcast data to all peers
await fazt.mesh.broadcast({
    type: 'notification',
    message: 'Hello from laptop'
});

// Force sync
await fazt.mesh.sync();
```

## Conflict Resolution

Uses CRDTs (Conflict-free Replicated Data Types):
- Last-write-wins for simple values
- Merge for lists/sets
- Custom resolvers for complex data

## Security

- All sync traffic is encrypted
- Peers authenticate via Persona keypairs
- Changes are signed by origin node

## Go Implementation

### LAN Discovery

Uses [peerdiscovery](https://github.com/schollz/peerdiscovery) for UDP
multicast peer discovery on local networks:

```go
import "github.com/schollz/peerdiscovery"

pd, _ := peerdiscovery.NewPeerDiscovery(peerdiscovery.Settings{
    Payload: []byte(`{"node":"fazt-abc123","caps":["hosting","ai"]}`),
    Notify: func(d peerdiscovery.Discovered) {
        // Peer found on LAN
        fazt.events.Emit("beacon.peer.found", d.Address, d.Payload)
    },
    NotifyLost: func(lp peerdiscovery.LostPeer) {
        // Peer disappeared
        fazt.events.Emit("beacon.peer.lost", lp.Address)
    },
})
```

**Why peerdiscovery**: Pure Go (~450 lines), dual-stack IPv4/IPv6, peer GC.

### WAN P2P Connectivity

Uses [pion/webrtc](https://github.com/pion/webrtc) for NAT traversal and
encrypted P2P connections across the internet:

```go
import "github.com/pion/webrtc/v4"

// Create peer connection with ICE (NAT traversal)
pc, _ := webrtc.NewPeerConnection(webrtc.Configuration{
    ICEServers: []webrtc.ICEServer{
        {URLs: []string{"stun:stun.l.google.com:19302"}},
    },
})

// Create data channel for mesh sync
dc, _ := pc.CreateDataChannel("mesh-sync", nil)
dc.OnMessage(func(msg webrtc.DataChannelMessage) {
    // Handle sync data from peer
})
```

**Pion modules used**:
- `pion/webrtc` - Full WebRTC stack (~3MB)
- `pion/ice` - Just NAT traversal (~500KB)
- `pion/datachannel` - Just data channels

**Why pion**: Pure Go, MIT license, NAT traversal (ICE), encrypted channels,
browser-compatible (Admin SPA can connect directly to nodes).
