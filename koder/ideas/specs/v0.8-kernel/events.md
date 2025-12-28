# Internal Events (IPC)

## Summary

Kernel-level event bus for inter-process communication. Apps and agents
can emit events, subscribe to events, and react to system happenings.
Foundation for decoupled, event-driven architectures.

## Why Kernel-Level

Events are a fundamental primitive:
- Apps need to communicate without tight coupling
- Agents need to observe and react to system state
- System events (deploy, error, limit) need broadcast mechanism
- Foundation for automation patterns

## Architecture

```
┌─────────┐     emit      ┌─────────────┐     deliver     ┌─────────┐
│  App A  │ ────────────► │   Event     │ ──────────────► │  App B  │
└─────────┘               │    Bus      │                 └─────────┘
                          │  (kernel)   │
┌─────────┐     emit      │             │     deliver     ┌─────────┐
│  Agent  │ ────────────► │             │ ──────────────► │  Worker │
└─────────┘               └─────────────┘                 └─────────┘
                                │
                                ▼
                          System Events
                          (deploy, error, etc.)
```

## Event Structure

```javascript
{
  id: 'evt_abc123',
  type: 'order.completed',      // Namespaced event type
  source: 'app:shop',           // Emitter (app, agent, system)
  data: { orderId: '123' },     // Payload
  timestamp: 1705312200000,
  correlationId: 'req_xyz'      // Optional, for tracing
}
```

## Usage

### Emit Event

```javascript
await fazt.events.emit('order.completed', {
  orderId: '123',
  total: 99.99,
  customer: 'user@example.com'
});
```

### Subscribe to Events

```javascript
// In app initialization or handler
fazt.events.on('order.completed', async (event) => {
  console.log(`Order ${event.data.orderId} completed`);
  await sendConfirmationEmail(event.data.customer);
});

// Pattern matching
fazt.events.on('order.*', async (event) => {
  // Matches order.completed, order.cancelled, order.refunded, etc.
});
```

### Unsubscribe

```javascript
const handler = async (event) => { ... };
fazt.events.on('order.completed', handler);

// Later
fazt.events.off('order.completed', handler);
```

## Event Namespacing

```
{source}.{entity}.{action}

Examples:
  app:shop.order.completed
  app:blog.post.published
  system.app.deployed
  system.limit.warning
  agent:monitor.alert.triggered
```

### Reserved Namespaces

| Prefix | Owner |
|--------|-------|
| `system.*` | Kernel only |
| `app:{name}.*` | Specific app |
| `agent:{name}.*` | Specific agent |

Apps can only emit events in their own namespace.

## System Events

Kernel emits events for system happenings:

| Event | When |
|-------|------|
| `system.app.deployed` | App deployment complete |
| `system.app.started` | App started |
| `system.app.stopped` | App stopped |
| `system.app.error` | App runtime error |
| `system.limit.warning` | Resource at 80% |
| `system.limit.exceeded` | Resource at 100% |
| `system.backup.completed` | Backup finished |
| `system.email.received` | Email received (v0.18) |
| `system.worker.completed` | Background job done (v0.19) |

```javascript
// Agent subscribing to system events
fazt.events.on('system.limit.warning', async (event) => {
  await fazt.services.notify.send({
    title: 'Resource Warning',
    body: `${event.data.resource} at ${event.data.percent}%`
  });
});
```

## Cross-App Events

By default, apps only see their own events. To receive events from
other apps, explicit permission is needed:

```json
// app.json
{
  "permissions": [
    "events:subscribe:app:shop.*",
    "events:subscribe:system.*"
  ]
}
```

## Persistence

Events are ephemeral by default (not stored). For persistence:

```javascript
await fazt.events.emit('order.completed', data, {
  persist: true,
  ttl: '7d'
});

// Query historical events
const events = await fazt.events.query({
  type: 'order.*',
  since: '2024-01-01',
  limit: 100
});
```

## Delivery Guarantees

| Mode | Guarantee |
|------|-----------|
| `fire-and-forget` | No guarantee (default) |
| `at-least-once` | Retry until ack |

```javascript
await fazt.events.emit('critical.action', data, {
  delivery: 'at-least-once',
  retries: 3,
  retryDelay: '1s'
});
```

## JS API

```javascript
fazt.events.emit(type, data, options?)
// options: { persist, ttl, delivery, retries, correlationId }

fazt.events.on(pattern, handler)
// pattern: exact type or wildcard (order.*)

fazt.events.off(pattern, handler)

fazt.events.once(pattern, handler)
// Auto-unsubscribe after first event

fazt.events.query(options)
// options: { type, since, until, source, limit }
// Only for persisted events
```

## Storage (for persisted events)

```sql
CREATE TABLE kernel_events (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    source TEXT NOT NULL,
    data TEXT,                  -- JSON
    correlation_id TEXT,
    created_at INTEGER,
    expires_at INTEGER
);

CREATE INDEX idx_events_type ON kernel_events(type, created_at);
CREATE INDEX idx_events_source ON kernel_events(source, created_at);
```

## CLI

```bash
# List recent events
fazt events list --type 'order.*' --limit 20

# Subscribe (for debugging)
fazt events watch --type 'system.*'

# Emit test event
fazt events emit app:test.ping '{"message": "hello"}'
```

## Example: Order Processing Pipeline

```javascript
// App: Shop (emits)
await fazt.events.emit('order.completed', {
  orderId: '123',
  items: [...],
  total: 99.99
});

// App: Inventory (subscribes)
fazt.events.on('app:shop.order.completed', async (event) => {
  for (const item of event.data.items) {
    await decrementStock(item.sku, item.quantity);
  }
});

// App: Notifications (subscribes)
fazt.events.on('app:shop.order.completed', async (event) => {
  await sendOrderConfirmation(event.data);
});

// Agent: Analytics (subscribes)
fazt.events.on('app:shop.order.*', async (event) => {
  await trackEvent(event.type, event.data);
});
```

## Example: Agent Monitoring

```javascript
// Agent subscribes to system events
fazt.events.on('system.limit.warning', async (event) => {
  const { resource, percent } = event.data;

  if (percent > 90) {
    await fazt.services.notify.send({
      title: 'Critical: Resource Alert',
      body: `${resource} at ${percent}%`,
      priority: 'high'
    });
  }
});

fazt.events.on('system.app.error', async (event) => {
  // Log error, maybe restart app, notify owner
  await fazt.log.error('App error', event.data);
});
```

## Limits

| Limit | Default |
|-------|---------|
| `maxEventSize` | 64 KB |
| `maxEventsPerSecond` | 100 (per app) |
| `maxSubscriptionsPerApp` | 50 |
| `persistedEventRetention` | 7 days |
