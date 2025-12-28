# v0.16 - Mesh

**Theme**: Decentralization, federation, and P2P.

## Summary

v0.16 enables Fazt instances to sync with each other and participate in
decentralized protocols. Your data can exist on multiple nodes, and your
identity can federate with the wider internet.

## Documents

- `p2p.md` - Kernel mesh synchronization
- `protocols.md` - ActivityPub and Nostr support

## Key Capabilities

### Kernel Mesh

- Transparent sync between Fazt nodes
- Local + remote instances in sync
- Gossip protocol for changes

### Protocol Support

- **ActivityPub**: Federate with Mastodon
- **Nostr**: Sovereign social identity

## Dependencies

- v0.9 (Storage): For sync primitives
- v0.14 (Security): For signed messages
- v0.15 (Identity): For federated identity
