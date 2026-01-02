# Temporal Patterns

## Summary

First-class temporal operators on the event bus. Enables windowing, debouncing,
sequence detection, and absence monitoring without per-app temporal logic.
Foundation for reactive behaviors in sensor/percept/effect pipelines.

## Why Kernel-Level

Temporal patterns are a primitive, not a service:
- Sensors emit high-frequency data needing windowing
- Percepts need to detect patterns over time
- Effects need throttling and debouncing
- Without kernel support, every app reinvents temporal logic
- Composes with existing events primitive

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     EVENT BUS                                │
│                                                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐    │
│  │  window  │  │ sliding  │  │ debounce │  │ throttle │    │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘    │
│       │             │             │             │            │
│  ┌────┴─────┐  ┌────┴─────┐                               │
│  │ sequence │  │ absence  │                               │
│  └──────────┘  └──────────┘                               │
│                                                              │
│  All operators emit new events or invoke handlers           │
└─────────────────────────────────────────────────────────────┘
```

## Operators

### Window

Collect events over a fixed duration, emit batch at window end.

```javascript
// Collect temperature readings for 5 minutes, process batch
fazt.events.window('sensor.temperature', '5m', async (events) => {
    const avg = events.reduce((sum, e) => sum + e.data.value, 0) / events.length;
    console.log(`5-minute average: ${avg}°C`);

    // Emit derived event
    await fazt.events.emit('metric.temperature.avg', {
        value: avg,
        count: events.length,
        window: '5m'
    });
});
```

**Parameters:**
- `pattern`: Event type pattern (supports wildcards)
- `duration`: Window size ('30s', '5m', '1h', etc.)
- `handler`: Receives array of events at window end

**Behavior:**
- Non-overlapping (tumbling) windows
- Empty windows don't invoke handler
- Handler receives events sorted by timestamp

### Sliding Window

Overlapping windows with configurable step.

```javascript
// 10-minute sliding window, evaluated every minute
fazt.events.sliding('sensor.temperature', '10m', '1m', async (events) => {
    const max = Math.max(...events.map(e => e.data.value));

    if (max > 35) {
        await fazt.events.emit('alert.temperature.high', { value: max });
    }
});
```

**Parameters:**
- `pattern`: Event type pattern
- `window`: Window size
- `step`: How often to evaluate
- `handler`: Receives events in window at each step

**Use cases:**
- Moving averages
- Rolling max/min
- Trend detection

### Debounce

Delay handler until events stop arriving for specified duration.

```javascript
// Only process after 500ms of no new events
fazt.events.debounce('ui.search.input', '500ms', async (event) => {
    // event is the last one received
    const results = await search(event.data.query);
    await fazt.events.emit('ui.search.results', { results });
});
```

**Parameters:**
- `pattern`: Event type pattern
- `delay`: Quiet period before firing
- `handler`: Receives last event only

**Use cases:**
- Search-as-you-type
- Input stabilization
- Sensor noise filtering

### Throttle

Limit handler invocations to at most once per duration.

```javascript
// At most one notification per minute for the same alert
fazt.events.throttle('alert.*', '1m', async (event) => {
    await fazt.services.notify.send({
        title: event.type,
        body: event.data.message
    });
}, { key: (e) => e.type });  // Group by event type
```

**Parameters:**
- `pattern`: Event type pattern
- `interval`: Minimum time between invocations
- `handler`: Receives first event in interval
- `options.key`: Optional function to group throttling

**Behavior:**
- First event fires immediately
- Subsequent events in interval are dropped
- Optional: emit count of dropped events

### Sequence

Detect ordered sequence of events within timeout.

```javascript
// Detect "door opened, then motion, then door closed" within 2 minutes
fazt.events.sequence([
    'sensor.door.opened',
    'sensor.motion.detected',
    'sensor.door.closed'
], '2m', async (events) => {
    // events[0] = door opened
    // events[1] = motion detected
    // events[2] = door closed

    await fazt.events.emit('pattern.entry.detected', {
        duration: events[2].timestamp - events[0].timestamp,
        events: events.map(e => e.id)
    });
});
```

**Parameters:**
- `patterns[]`: Ordered array of event patterns
- `timeout`: Max time for complete sequence
- `handler`: Receives array of matching events

**Behavior:**
- Events must arrive in order
- Partial sequences timeout silently
- First matching event starts the clock
- Handler receives one event per pattern

### Absence

Trigger when expected event doesn't arrive within duration.

```javascript
// Alert if no heartbeat for 5 minutes
fazt.events.absence('system.heartbeat', '5m', async (lastSeen) => {
    await fazt.services.notify.send({
        title: 'Heartbeat Missing',
        body: `No heartbeat since ${new Date(lastSeen).toISOString()}`,
        priority: 'high'
    });
});
```

**Parameters:**
- `pattern`: Event type pattern
- `duration`: Max allowed gap
- `handler`: Receives timestamp of last seen event (or null)

**Behavior:**
- Handler invoked once when absence detected
- Resets when matching event arrives
- Can fire repeatedly if events keep not arriving

## Advanced Usage

### Chaining Operators

```javascript
// Sliding average, throttled output
fazt.events.sliding('sensor.temperature', '5m', '30s', async (events) => {
    const avg = events.reduce((sum, e) => sum + e.data.value, 0) / events.length;
    await fazt.events.emit('metric.temperature.avg', { value: avg });
});

fazt.events.throttle('metric.temperature.avg', '5m', async (event) => {
    if (event.data.value > 30) {
        await fazt.services.notify.send({
            title: 'Temperature Warning',
            body: `5-minute average: ${event.data.value}°C`
        });
    }
});
```

### Conditional Sequences

```javascript
// Sequence with conditions on intermediate events
fazt.events.sequence([
    { pattern: 'sensor.door.opened', where: { data: { door: 'front' } } },
    { pattern: 'sensor.motion.*' },
    { pattern: 'sensor.door.closed', where: { data: { door: 'front' } } }
], '2m', handler);
```

### Named Temporal Rules

```javascript
// Register a named rule (persisted, survives restart)
await fazt.events.temporal.register({
    name: 'high-temp-alert',
    type: 'absence',
    pattern: 'sensor.temperature',
    duration: '10m',
    emit: 'alert.sensor.missing',  // Auto-emit on trigger
    notify: true                    // Also send notification
});

// List registered rules
const rules = await fazt.events.temporal.list();

// Remove rule
await fazt.events.temporal.remove('high-temp-alert');
```

## JS API

```javascript
// Window operators
fazt.events.window(pattern, duration, handler)
fazt.events.sliding(pattern, window, step, handler)

// Rate limiting
fazt.events.debounce(pattern, delay, handler)
fazt.events.throttle(pattern, interval, handler, options?)
// options: { key, emitDropped }

// Pattern detection
fazt.events.sequence(patterns, timeout, handler)
// patterns: string[] or { pattern, where }[]

// Absence detection
fazt.events.absence(pattern, duration, handler)

// Named rules (persistent)
fazt.events.temporal.register(config)
// config: { name, type, pattern, duration, emit?, notify?, handler? }

fazt.events.temporal.list()
fazt.events.temporal.get(name)
fazt.events.temporal.remove(name)
fazt.events.temporal.pause(name)
fazt.events.temporal.resume(name)

// Cleanup
fazt.events.temporal.clear()  // Remove all in-flight state
```

## CLI

```bash
# List registered temporal rules
fazt events temporal list

# Register a rule
fazt events temporal register \
    --name high-temp \
    --type sliding \
    --pattern 'sensor.temperature' \
    --window 5m \
    --step 1m \
    --emit 'metric.temperature.avg'

# Remove rule
fazt events temporal remove high-temp

# Pause/resume
fazt events temporal pause high-temp
fazt events temporal resume high-temp

# Show rule details
fazt events temporal show high-temp

# Debug: show in-flight windows
fazt events temporal debug --pattern 'sensor.*'
```

## Storage

```sql
-- Persistent temporal rules
CREATE TABLE temporal_rules (
    id TEXT PRIMARY KEY,
    app_uuid TEXT,
    name TEXT NOT NULL,
    type TEXT NOT NULL,        -- window, sliding, debounce, throttle, sequence, absence
    config_json TEXT NOT NULL, -- Pattern, duration, etc.
    enabled INTEGER DEFAULT 1,
    created_at INTEGER,
    updated_at INTEGER,
    UNIQUE(app_uuid, name)
);

-- In-flight window state (for crash recovery)
CREATE TABLE temporal_state (
    rule_id TEXT NOT NULL,
    window_key TEXT NOT NULL,   -- Pattern + window start
    events_json TEXT,           -- Collected events
    started_at INTEGER,
    expires_at INTEGER,
    PRIMARY KEY (rule_id, window_key)
);

CREATE INDEX idx_temporal_state_expires ON temporal_state(expires_at);
```

## Implementation Notes

### Memory Management

- In-memory buffers for active windows
- Periodic checkpoint to SQLite for crash recovery
- LRU eviction for inactive patterns
- Configurable max events per window

### Timing

- Uses monotonic clock for intervals
- Wall clock for timestamps in events
- Drift-tolerant (graceful with `fazt.time` consensus)

### Crash Recovery

- Active windows persisted every 30 seconds
- On restart, pending windows resumed
- Handlers re-invoked for windows that should have fired

## Example: Sensor Monitoring Pipeline

```javascript
// 1. Debounce rapid sensor readings
fazt.events.debounce('sensor.temperature.raw', '1s', async (event) => {
    await fazt.events.emit('sensor.temperature', event.data);
});

// 2. Compute 5-minute sliding average
fazt.events.sliding('sensor.temperature', '5m', '1m', async (events) => {
    const avg = events.reduce((sum, e) => sum + e.data.value, 0) / events.length;
    const max = Math.max(...events.map(e => e.data.value));
    const min = Math.min(...events.map(e => e.data.value));

    await fazt.events.emit('metric.temperature', { avg, max, min });
});

// 3. Throttle alerts to prevent spam
fazt.events.throttle('metric.temperature', '15m', async (event) => {
    if (event.data.avg > 35) {
        await fazt.services.notify.send({
            title: 'High Temperature Alert',
            body: `Average: ${event.data.avg}°C (max: ${event.data.max}°C)`
        });
    }
});

// 4. Detect sensor absence
fazt.events.absence('sensor.temperature', '10m', async () => {
    await fazt.services.notify.send({
        title: 'Sensor Offline',
        body: 'Temperature sensor not reporting',
        priority: 'high'
    });
});
```

## Example: User Behavior Detection

```javascript
// Detect rage clicks (5+ clicks in 2 seconds)
fazt.events.window('ui.click', '2s', async (events) => {
    if (events.length >= 5) {
        const element = events[0].data.element;
        await fazt.events.emit('ux.rage_click', {
            element,
            count: events.length
        });
    }
});

// Detect checkout abandonment (cart added, no purchase in 30m)
fazt.events.sequence([
    'shop.cart.add',
    { pattern: 'shop.checkout.complete', invert: true }  // NOT this event
], '30m', async (events) => {
    await fazt.events.emit('shop.cart.abandoned', {
        cartId: events[0].data.cartId
    });
});
```

## Limits

| Limit | Default |
|-------|---------|
| Max concurrent windows | 1000 (per app) |
| Max events per window | 10000 |
| Max window duration | 24 hours |
| Max sequence length | 10 |
| Checkpoint interval | 30 seconds |
