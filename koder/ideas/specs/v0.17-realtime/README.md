# v0.17 - Realtime

**Theme**: WebSocket-based real-time communication.

## Summary

v0.17 adds native WebSocket support with pub/sub channels. Apps can push
live updates to connected clients, track presence, and build collaborative
features—all with automatic app isolation.

## Goals

1. **Pub/Sub Channels**: Broadcast messages to subscribers
2. **Presence**: Know who's connected
3. **Server Push**: Send from serverless to connected clients
4. **Isolation**: Apps can't see each other's connections

## Key Capabilities

| Capability         | Description                   |
| ------------------ | ----------------------------- |
| WebSocket endpoint | `wss://app.domain.com/_ws`    |
| Channels           | Namespaced pub/sub rooms      |
| Presence           | Track connected clients       |
| Server broadcast   | Push from serverless handlers |
| Rate limiting      | Configurable per-app limits   |

## Documents

- `websocket.md` - Connection lifecycle, channels, and API

## Dependencies

- v0.10 (Runtime): Serverless handlers for broadcast triggers
- v0.8 (Kernel): Limits system for connection caps

## Performance Target

On 1GB RAM VPS:
- **5,000 concurrent connections** (safe baseline)
- **10,000 connections** (with tuning)
- ~50KB per connection overhead
