# VPN Gateway

## Summary

Built-in WireGuard VPN transforms Fazt into a privacy gateway. Route device
traffic through your VPS, access "shadow apps" privately, and create secure
tunnels without external software.

## Implementation

### Userspace WireGuard

Uses `wireguard-go` to maintain zero-dependency promise:
- No kernel modules required
- No `wg-quick` or system tools
- Binds to UDP port (default: 51820)

### Latent Activation

VPN module is dormant by default:
- UDP port only binds if peers exist
- Zero resource usage when unused
- Attack surface: zero when disabled

## CLI

```bash
# Initialize VPN (generates server keypair)
fazt net vpn init

# Add peer (generates config)
fazt net vpn add-peer --name "iPhone"
# Output: Peer added. Scan QR or download config.

# Show QR code for mobile setup
fazt net vpn qr peer_abc123

# List peers
fazt net vpn list

# Remove peer
fazt net vpn remove-peer peer_abc123
```

## Storage

```sql
CREATE TABLE vpn_peers (
    id TEXT PRIMARY KEY,
    name TEXT,
    public_key TEXT,
    allowed_ips TEXT,
    config TEXT,          -- JSON metadata
    last_handshake INTEGER,
    created_at INTEGER
);
```

## Shadow Apps

Apps can be VPN-only:

```json
{
  "name": "private-dashboard",
  "visibility": "vpn-only"
}
```

These apps:
- Don't respond to public requests
- Only accessible via VPN tunnel
- Effectively invisible on internet

## JS Runtime

```javascript
// Check if request came via VPN
if (fazt.net.vpn.status()) {
    // Show admin features
}

// Get peer info
const peer = fazt.net.vpn.peer();
console.log(peer.name);  // "iPhone"

// Require elevated trust
await fazt.net.vpn.authorize();  // Triggers TOTP
```

## Use Cases

1. **Privacy Gateway**: Route all phone traffic through VPS
2. **Private Apps**: Admin tools invisible to internet
3. **Secure Access**: SSH alternative for management
4. **Mobile Development**: Test apps from phone via tunnel
