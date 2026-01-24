# WebSocket & Pub/Sub

## Summary

Native WebSocket support with channel-based pub/sub. Clients connect to
`/_ws`, subscribe to channels, and receive broadcasts. Server-side code
can push messages via `fazt.realtime`.

## Connection Lifecycle

```
Client                          Fazt Kernel
   │                                 │
   ├─── WSS connect ────────────────►│
   │                                 │ Authenticate (optional)
   │◄─── connection.open ────────────┤
   │                                 │
   ├─── subscribe("chat") ──────────►│
   │◄─── subscribed("chat") ─────────┤
   │                                 │
   │◄─── message(channel, data) ─────┤ (from server broadcast)
   │                                 │
   ├─── unsubscribe("chat") ────────►│
   │◄─── unsubscribed("chat") ───────┤
   │                                 │
   ├─── close ──────────────────────►│
   │                                 │
```

## Endpoint

```
wss://app-slug.domain.com/_ws
wss://app-slug.domain.com/_ws?token=<auth-token>
```

The `/_ws` path is reserved by the kernel. Requests are routed to the
realtime subsystem, not the app's VFS.

## Client Protocol

### Message Format

All messages are JSON:

```json
{
  "type": "subscribe|unsubscribe|message|ping|pong",
  "channel": "channel-name",
  "data": { ... }
}
```

### Subscribe

```json
{ "type": "subscribe", "channel": "chat" }
```

Response:
```json
{ "type": "subscribed", "channel": "chat" }
```

### Receive Message

```json
{
  "type": "message",
  "channel": "chat",
  "data": { "user": "alice", "text": "hello" },
  "timestamp": 1704067200000
}
```

### Unsubscribe

```json
{ "type": "unsubscribe", "channel": "chat" }
```

### Ping/Pong (Heartbeat)

```json
{ "type": "ping" }
```

Response:
```json
{ "type": "pong" }
```

Kernel sends ping every 30s. Client must respond within 10s or connection
is terminated.

## Server-Side API (fazt.realtime)

### Broadcast to Channel

```javascript
// api/chat.js - HTTP handler that broadcasts
module.exports = async (req) => {
    const { channel, message } = req.json;

    await fazt.realtime.broadcast(channel, {
        user: req.user,
        text: message,
        timestamp: Date.now()
    });

    return { json: { sent: true } };
};
```

### Get Channel Subscribers

```javascript
const subscribers = await fazt.realtime.subscribers('chat');
// [{ clientId: 'abc', connectedAt: ..., metadata: {...} }, ...]
```

### Get Connection Count

```javascript
const count = await fazt.realtime.count('chat');
// 42
```

### Kick Client

```javascript
await fazt.realtime.kick(clientId, 'reason');
```

### Broadcast to All (App-wide)

```javascript
await fazt.realtime.broadcastAll({
    type: 'system',
    message: 'Server restarting in 5 minutes'
});
```

## Channel Types

### Public Channels

Any client can subscribe. No authentication required.

```javascript
// Client subscribes to "updates"
{ "type": "subscribe", "channel": "updates" }
```

### Private Channels (prefix: `private-`)

Require authentication token. Kernel validates before allowing subscribe.

```javascript
// Client must connect with token
wss://app.domain.com/_ws?token=<jwt>

// Then subscribe to private channel
{ "type": "subscribe", "channel": "private-user-123" }
```

### Presence Channels (prefix: `presence-`)

Track who's online. Subscribers receive join/leave events.

```javascript
// Subscribe
{ "type": "subscribe", "channel": "presence-room-1" }

// Receive presence events
{ "type": "presence", "channel": "presence-room-1",
  "event": "join", "client": { "id": "abc", "name": "Alice" } }

{ "type": "presence", "channel": "presence-room-1",
  "event": "leave", "client": { "id": "abc" } }

// Get current members
{ "type": "members", "channel": "presence-room-1" }
// Response:
{ "type": "members", "channel": "presence-room-1",
  "members": [{ "id": "abc", "name": "Alice" }, ...] }
```

## App Isolation

Channels are automatically namespaced by `app_uuid`:

```
Internal: app_x9z2k:chat
Client sees: chat
```

- App A's "chat" channel is completely separate from App B's "chat"
- `fazt.realtime.broadcast('chat', ...)` only reaches App A's subscribers
- No cross-app channel access unless explicit kernel permission

## Authentication Integration

### Anonymous Connections

Allowed by default. Client gets a random `clientId`.

### Authenticated Connections

Pass token in query string or first message:

```javascript
// Option 1: Query string
wss://app.domain.com/_ws?token=<jwt>

// Option 2: First message
{ "type": "auth", "token": "<jwt>" }
```

Token is validated against v0.15 Persona/SSO system.

### Per-Channel Auth

Apps can require auth for specific channels via `app.json`:

```json
{
  "realtime": {
    "channels": {
      "public-*": { "auth": false },
      "private-*": { "auth": true },
      "admin": { "auth": true, "role": "admin" }
    }
  }
}
```

## Limits

Configured in kernel limits system (v0.8):

| Limit                       | Default | Description               |
| --------------------------- | ------- | ------------------------- |
| `maxConnectionsTotal`       | 5000    | All apps combined         |
| `maxConnectionsPerApp`      | 500     | Per app                   |
| `maxChannelsPerApp`         | 100     | Channels per app          |
| `maxSubscriptionsPerClient` | 20      | Channels per connection   |
| `maxMessageSizeKB`          | 64      | Single message            |
| `maxMessagesPerSecond`      | 100     | Per connection            |
| `idleTimeoutSeconds`        | 300     | Disconnect if no activity |

### Limit Behavior

- **At 80%**: Warning logged, owner notified
- **At 100%**: New connections rejected with `503`
- **Rate exceeded**: Message dropped, client warned

## Client SDK (Optional)

Fazt provides a browser SDK for convenience:

```html
<script src="https://fazt.sh/realtime.min.js"></script>
<script>
const rt = new FaztRealtime('wss://app.domain.com/_ws');

rt.on('connected', () => {
    rt.subscribe('chat');
});

rt.on('message', (channel, data) => {
    console.log(`${channel}: ${data}`);
});

rt.subscribe('presence-room-1', {
    onJoin: (member) => console.log('joined:', member),
    onLeave: (member) => console.log('left:', member)
});
</script>
```

## Implementation Notes

### Go Libraries

- `nhooyr.io/websocket` - Modern, well-maintained
- Alternative: `gorilla/websocket` - Battle-tested

### Hub Pattern

Central hub manages all connections:

```go
type Hub struct {
    apps       map[string]*AppHub  // Per-app hubs
    register   chan *Client
    unregister chan *Client
    broadcast  chan *Message
}

type AppHub struct {
    appUUID    string
    clients    map[string]*Client
    channels   map[string]map[*Client]bool
}
```

### Memory Management

- Per connection: ~10-50KB
- Per channel subscription: ~100 bytes
- Message buffer: configurable, default 256 messages

### Graceful Shutdown

On kernel restart:
1. Send "server_restart" to all clients
2. Wait 5s for clean disconnects
3. Force close remaining

## Future: Yjs Protocol Support (v0.17.1)

Yjs is a CRDT library for collaborative editing (Google Docs, Notion, Figma).
The y-websocket provider expects a specific binary protocol.

### Why Built-In

- Most collaboration tools use Yjs
- Currently requires separate y-websocket server
- Fazt should be a complete platform for collaborative apps

### Protocol Overview

y-websocket uses binary WebSocket frames:

```
[messageType: 1 byte][...payload]

Message types:
0 = sync step 1 (state vector)
1 = sync step 2 (diff)
2 = update
3 = awareness (cursor positions, selections)
```

### Implementation

Add binary frame support to existing WebSocket hub:

```go
// Detect Yjs client by URL or first message
if isYjsClient(conn) {
    handleYjsProtocol(conn, room)
} else {
    handleJsonProtocol(conn) // existing behavior
}
```

### API Surface

```javascript
// Server-side: get Yjs document state
const doc = await fazt.realtime.yjsDoc('doc-123')

// Apply update programmatically
await fazt.realtime.yjsUpdate('doc-123', updateBytes)
```

### Persistence

Yjs updates stored in SQLite for durability:

```sql
CREATE TABLE yjs_docs (
    app_id TEXT,
    doc_id TEXT,
    state BLOB,        -- Encoded Y.Doc
    updated_at DATETIME,
    PRIMARY KEY (app_id, doc_id)
);
```

## Future: Push Notifications (v0.18+)

For offline/mobile push, integrate FCM/APNs.

### Difficulty Assessment

| Component | Effort | Notes |
|-----------|--------|-------|
| FCM integration | Medium | HTTP API, need Google Cloud project |
| APNs integration | Medium | HTTP/2, need Apple Developer cert |
| Token storage | Low | New table for device tokens |
| Subscription management | Low | Link tokens to channels |
| App configuration | Low | Apps provide FCM/APNs keys |

### Total Estimate

~500-800 lines of Go. Main complexity is managing device tokens and
handling delivery failures (token expiration, uninstalls).

### API Surface

```javascript
// Register device token (from app's client code)
await fazt.push.register(token, platform) // 'fcm' | 'apns'

// Server-side: send push to offline users
await fazt.push.send(userId, {
    title: 'New message',
    body: 'Alice: hey!',
    data: { chatId: '123' }
})

// Or broadcast to channel (only offline subscribers)
await fazt.push.broadcastOffline('chat:room1', notification)
```

### Architecture

```
Browser online:  WebSocket (existing)
Browser offline: Skip (no push to desktop browsers)
Mobile online:   WebSocket (existing)
Mobile offline:  FCM/APNs push
```

### Not In Scope

- Web Push API (requires service worker, complex)
- SMS notifications (different beast entirely)
