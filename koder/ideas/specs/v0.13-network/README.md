# v0.13 - Network

**Theme**: Advanced networking primitives.

## Summary

v0.13 adds a built-in VPN gateway (WireGuard) and multi-domain support. Apps
can be made "shadow"—only accessible via VPN—and external domains can map to
internal apps.

## Documents

- `vpn.md` - WireGuard gateway
- `domains.md` - Custom domain mapping

## Key Capabilities

### VPN Gateway

- Userspace WireGuard via `wireguard-go`
- QR code peer provisioning
- Zero-config for owner
- "Shadow apps" only visible via tunnel

### Multi-Domain

- Map any domain to any app
- On-demand HTTPS via CertMagic
- White-label support

## Dependencies

- v0.8 (Kernel): Network module
