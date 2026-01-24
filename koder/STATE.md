# Fazt Implementation State

**Last Updated**: 2026-01-25
**Current Version**: v0.10.10

## Status

State: CLEAN - WebSocket pub/sub system implemented and tested

---

## Last Session

**WebSocket Pub/Sub Implementation**

1. **Enhanced `internal/hosting/ws.go`**
   - Added `Client` struct with ID, channels, send buffer
   - Added `channels` map to `SiteHub` for pub/sub
   - JSON protocol: subscribe, unsubscribe, message, ping, pong
   - Heartbeat: ping every 30s, 10s timeout
   - New methods: `BroadcastToChannel`, `BroadcastAll`, `GetSubscribers`, `ChannelCount`, `KickClient`

2. **Created `internal/hosting/realtime.go`** (Goja bindings)
   - `fazt.realtime.broadcast(channel, data)` - Push to channel subscribers
   - `fazt.realtime.broadcastAll(data)` - Push to all app connections
   - `fazt.realtime.subscribers(channel)` - List client IDs
   - `fazt.realtime.count(channel?)` - Count subscribers or total
   - `fazt.realtime.kick(clientId, reason)` - Disconnect client

3. **Wired realtime injector** in `internal/runtime/handler.go`
   - Added alongside storage injector in `executeWithFazt()`

4. **Changed WebSocket route** from `/ws` to `/_ws`
   - Line 1491 in `cmd/server/main.go`

5. **Added 21 new tests**
   - `ws_test.go`: Channel subscription, isolation, broadcast, protocol
   - `realtime_test.go`: Goja binding tests

6. **Fixed pre-existing test failures**
   - Updated `hosting_test.go` schema to match migration 012
   - Added `title`, `aliases` table, `app_id` column

7. **Updated spec** (`koder/ideas/specs/v0.17-realtime/websocket.md`)
   - Added "Future: Yjs Protocol Support (v0.17.1)" section
   - Added "Future: Push Notifications (v0.18+)" section with FCM/APNs

## Client Protocol

```json
// Subscribe
→ {"type":"subscribe","channel":"chat"}
← {"type":"subscribed","channel":"chat"}

// Receive message
← {"type":"message","channel":"chat","data":{...},"timestamp":1706000000000}

// Heartbeat (server sends every 30s)
← {"type":"ping"}
→ {"type":"pong"}
```

## Next Up

1. **Stress test WebSocket implementation**
   - Concurrent connections, message throughput
   - Channel fanout performance

2. **Brainstorm /fazt-app ideas** showcasing storage + WebSockets
   - Collaborative apps (Yjs-style)
   - Real-time dashboards
   - Chat/presence demos

---

## Quick Reference

```bash
# Rebuild and restart local server
go build -o ~/.local/bin/fazt ./cmd/server && \
  systemctl --user restart fazt-local

# Test WebSocket (websocat)
websocat ws://test.192.168.64.3:8080/_ws
{"type":"subscribe","channel":"chat"}

# Run WebSocket tests
go test -v ./internal/hosting/... -run "TestChannel|TestBroadcast|TestRealtime"
```
