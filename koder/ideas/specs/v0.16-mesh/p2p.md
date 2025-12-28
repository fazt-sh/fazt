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
