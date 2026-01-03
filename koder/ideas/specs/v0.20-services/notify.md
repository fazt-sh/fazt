# Notifications Service

## Summary

Unified notification delivery with user preferences, multi-channel fallback,
batching, and history. Sits between kernel devices (`fazt.dev.*`) and apps,
providing a single abstraction for "tell the human."

## Why a Service

Notifications are a common pattern, not a primitive:
- Builds on kernel devices (sms, email) and realtime
- Adds user preferences (quiet hours, digest mode)
- Adds delivery intelligence (fallback, batching)
- Every app sending notifications duplicates this logic

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        APP CODE                              │
│  await fazt.services.notify.send({ title, body, ... })      │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                   NOTIFICATIONS SERVICE                      │
│                                                              │
│  1. Check user preferences                                  │
│  2. Apply quiet hours / do-not-disturb                      │
│  3. Batch if digest mode                                    │
│  4. Select channel(s)                                       │
│  5. Deliver via kernel devices                              │
│  6. Track delivery status                                   │
│  7. Store in history                                        │
└───────┬─────────────┬─────────────┬─────────────┬───────────┘
        │             │             │             │
        ▼             ▼             ▼             ▼
   ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐
   │  Push   │  │  Email  │  │   SMS   │  │  In-App │
   │(realtime│  │(dev.email│ │(dev.sms)│  │(realtime│
   └─────────┘  └─────────┘  └─────────┘  └─────────┘
```

## Notification Object

```javascript
{
    id: 'ntf_abc123',
    to: 'user_123',           // User ID, or omit for owner
    title: 'Order Shipped',
    body: 'Your order #12345 has shipped.',
    priority: 'normal',       // low, normal, high, urgent
    channels: ['push', 'email'],
    category: 'orders',       // For filtering/preferences
    data: {                   // App-specific payload
        orderId: '12345',
        trackingUrl: 'https://...'
    },
    actions: [                // Optional action buttons
        { id: 'view', label: 'View Order' },
        { id: 'track', label: 'Track Package' }
    ],
    createdAt: 1705312200000,
    deliveredAt: 1705312201000,
    readAt: null
}
```

## Usage

### Basic Send

```javascript
// Send to owner (default)
await fazt.services.notify.send({
    title: 'Backup Complete',
    body: 'Your daily backup finished successfully.'
});

// Send to specific user
await fazt.services.notify.send({
    to: 'user_123',
    title: 'New Comment',
    body: 'Alice replied to your post.',
    category: 'social'
});
```

### Priority Levels

```javascript
// Low: batched, may be delayed
await fazt.services.notify.send({
    title: 'Weekly Summary',
    body: '...',
    priority: 'low'
});

// Normal: standard delivery (default)
await fazt.services.notify.send({
    title: 'Order Shipped',
    body: '...',
    priority: 'normal'
});

// High: bypasses batching, respects quiet hours
await fazt.services.notify.send({
    title: 'Payment Failed',
    body: '...',
    priority: 'high'
});

// Urgent: bypasses everything, always delivers
await fazt.services.notify.send({
    title: 'Security Alert',
    body: 'New login from unknown device',
    priority: 'urgent'
});
```

### Channel Selection

```javascript
// Specific channels
await fazt.services.notify.send({
    title: 'Important Update',
    body: '...',
    channels: ['email', 'sms']
});

// Auto (service picks based on preferences/priority)
await fazt.services.notify.send({
    title: 'Update',
    body: '...',
    channels: 'auto'  // Default
});
```

### With Actions

```javascript
await fazt.services.notify.send({
    title: 'New Friend Request',
    body: 'Bob wants to connect.',
    actions: [
        { id: 'accept', label: 'Accept', primary: true },
        { id: 'decline', label: 'Decline' }
    ],
    data: { requestId: 'req_123' }
});

// Handle action (in your app)
fazt.events.on('notification.action', async (event) => {
    const { notificationId, actionId, data } = event.data;

    if (actionId === 'accept') {
        await acceptFriendRequest(data.requestId);
    }
});
```

## User Preferences

### Get Preferences

```javascript
// Get current user's preferences
const prefs = await fazt.services.notify.preferences.get();

// Get specific user's preferences (owner only)
const prefs = await fazt.services.notify.preferences.get('user_123');
```

### Set Preferences

```javascript
await fazt.services.notify.preferences.set({
    // Channel preferences (ordered by preference)
    channels: ['push', 'email'],  // Try push first, fall back to email

    // Quiet hours
    quietHours: {
        enabled: true,
        start: '22:00',
        end: '08:00',
        timezone: 'America/New_York'
    },

    // Digest mode (batch low-priority notifications)
    digest: {
        enabled: true,
        frequency: 'daily',     // hourly, daily, weekly
        time: '09:00',
        timezone: 'America/New_York'
    },

    // Per-category settings
    categories: {
        'marketing': { enabled: false },
        'social': { channels: ['push'] },
        'security': { bypassQuietHours: true }
    },

    // Do not disturb (temporary)
    dnd: {
        enabled: false,
        until: null
    }
});
```

### Preference Schema

```javascript
{
    channels: ['push', 'email', 'sms'],  // Default order
    quietHours: {
        enabled: false,
        start: '22:00',
        end: '08:00',
        timezone: 'UTC'
    },
    digest: {
        enabled: false,
        frequency: 'daily',
        time: '09:00',
        timezone: 'UTC'
    },
    categories: {
        // Per-category overrides
    },
    dnd: {
        enabled: false,
        until: null  // Timestamp
    }
}
```

## Notification History

```javascript
// List notifications
const notifications = await fazt.services.notify.history({
    limit: 50,
    unreadOnly: false,
    category: 'orders'
});

// Get specific notification
const notif = await fazt.services.notify.get('ntf_abc123');

// Mark as read
await fazt.services.notify.markRead('ntf_abc123');

// Mark all as read
await fazt.services.notify.markAllRead();

// Delete notification
await fazt.services.notify.delete('ntf_abc123');

// Count unread
const count = await fazt.services.notify.unreadCount();
```

## Channel Implementations

### Push (In-App / Realtime)

Uses `fazt.realtime` to deliver to connected clients:

```javascript
// Delivered via WebSocket to channel: notification:{userId}
// Client subscribes to receive real-time notifications
```

### Email

Uses `fazt.dev.email`:

```javascript
// Auto-generates email from notification
// Subject: title
// Body: body + actions rendered as links
// Uses app's configured email template (optional)
```

### SMS

Uses `fazt.dev.sms`:

```javascript
// Truncated to SMS length
// Actions rendered as short URLs
// Only for high/urgent priority by default
```

## Delivery Logic

```
1. Receive notification request
2. Resolve recipient (to → user record)
3. Check DND status
   - If DND and priority < urgent: queue for later
4. Check quiet hours
   - If quiet hours and priority < high: queue for later
5. Check digest mode
   - If digest and priority == low: add to digest batch
6. Select channels
   - If 'auto': use preference order
   - If explicit: use specified channels
7. For each channel:
   a. Check category preferences
   b. Attempt delivery
   c. On failure: try next channel
   d. Log delivery status
8. Store in history
9. Emit event: notification.sent
```

## JS API

```javascript
// Send notification
fazt.services.notify.send(options)
// options: { to?, title, body, priority?, channels?, category?, data?, actions? }
// Returns: { id, deliveredVia }

// Preferences
fazt.services.notify.preferences.get(userId?)
fazt.services.notify.preferences.set(options)
// options: { channels, quietHours, digest, categories, dnd }

// History
fazt.services.notify.history(options?)
// options: { limit, offset, unreadOnly, category, since }
fazt.services.notify.get(id)
fazt.services.notify.markRead(id)
fazt.services.notify.markAllRead(options?)
fazt.services.notify.delete(id)
fazt.services.notify.unreadCount(options?)

// Bulk operations
fazt.services.notify.sendBulk(notifications[])
// Returns: { sent, failed, batched }

// Templates (optional)
fazt.services.notify.templates.create(name, template)
fazt.services.notify.templates.list()
fazt.services.notify.templates.get(name)
fazt.services.notify.templates.delete(name)
fazt.services.notify.sendTemplate(name, variables, options)
```

## HTTP API

```
POST   /api/notify                    # Send notification
GET    /api/notify                    # List history
GET    /api/notify/unread             # Unread count
GET    /api/notify/{id}               # Get notification
POST   /api/notify/{id}/read          # Mark read
POST   /api/notify/read-all           # Mark all read
DELETE /api/notify/{id}               # Delete

GET    /api/notify/preferences        # Get preferences
PUT    /api/notify/preferences        # Set preferences

GET    /api/notify/templates          # List templates
POST   /api/notify/templates          # Create template
GET    /api/notify/templates/{name}   # Get template
DELETE /api/notify/templates/{name}   # Delete template
```

## CLI

```bash
# Send notification
fazt notify send --title "Test" --body "Hello world"
fazt notify send --to user_123 --title "Alert" --priority high

# List notifications
fazt notify list --limit 20
fazt notify list --unread

# Preferences
fazt notify preferences
fazt notify preferences set --quiet-hours "22:00-08:00"
fazt notify preferences set --digest daily
fazt notify preferences set --dnd 2h

# History
fazt notify history --category orders
fazt notify show ntf_abc123

# Templates
fazt notify templates list
fazt notify templates show welcome
```

## Storage

```sql
CREATE TABLE notifications (
    id TEXT PRIMARY KEY,
    app_uuid TEXT NOT NULL,
    user_id TEXT,              -- NULL for owner
    title TEXT NOT NULL,
    body TEXT,
    priority TEXT DEFAULT 'normal',
    category TEXT,
    channels_json TEXT,        -- Requested channels
    delivered_via TEXT,        -- Actual delivery channel
    data_json TEXT,
    actions_json TEXT,
    created_at INTEGER NOT NULL,
    delivered_at INTEGER,
    read_at INTEGER,
    FOREIGN KEY (app_uuid) REFERENCES apps(uuid)
);

CREATE INDEX idx_notifications_user ON notifications(app_uuid, user_id, created_at);
CREATE INDEX idx_notifications_unread ON notifications(app_uuid, user_id, read_at) WHERE read_at IS NULL;

CREATE TABLE notification_preferences (
    app_uuid TEXT NOT NULL,
    user_id TEXT,              -- NULL for owner
    preferences_json TEXT NOT NULL,
    updated_at INTEGER,
    PRIMARY KEY (app_uuid, user_id)
);

CREATE TABLE notification_digest (
    id TEXT PRIMARY KEY,
    app_uuid TEXT NOT NULL,
    user_id TEXT,
    notifications_json TEXT,   -- Batched notification IDs
    scheduled_for INTEGER,
    sent_at INTEGER
);

CREATE TABLE notification_templates (
    app_uuid TEXT NOT NULL,
    name TEXT NOT NULL,
    template_json TEXT NOT NULL,
    created_at INTEGER,
    PRIMARY KEY (app_uuid, name)
);
```

## Events

```javascript
// Emitted when notification sent
'notification.sent': {
    id: 'ntf_abc123',
    to: 'user_123',
    channel: 'email',
    priority: 'normal'
}

// Emitted when action taken
'notification.action': {
    notificationId: 'ntf_abc123',
    actionId: 'accept',
    data: { ... }
}

// Emitted when read
'notification.read': {
    id: 'ntf_abc123',
    userId: 'user_123'
}
```

## Example: Multi-Channel Fallback

```javascript
// Send with fallback chain
const result = await fazt.services.notify.send({
    to: 'user_123',
    title: 'Important Update',
    body: 'Please review your account settings.',
    channels: ['push', 'email', 'sms'],
    priority: 'high'
});

// Service tries:
// 1. Push (realtime) - user not connected, fails
// 2. Email - succeeds
// result.deliveredVia = 'email'
```

## Example: Digest Mode

```javascript
// User preference: digest daily at 9am
// Low-priority notifications are batched

await fazt.services.notify.send({
    to: 'user_123',
    title: 'New follower',
    body: 'Alice followed you',
    priority: 'low',
    category: 'social'
});

await fazt.services.notify.send({
    to: 'user_123',
    title: 'New follower',
    body: 'Bob followed you',
    priority: 'low',
    category: 'social'
});

// At 9am, user receives single email:
// "You have 2 new followers: Alice, Bob"
```

## Example: Security Alerts

```javascript
// Category configured to bypass quiet hours
await fazt.services.notify.send({
    title: 'New Login Detected',
    body: 'Login from Chrome on Windows from IP 1.2.3.4',
    category: 'security',
    priority: 'urgent',
    actions: [
        { id: 'secure', label: 'Secure Account', primary: true },
        { id: 'ignore', label: 'This was me' }
    ],
    data: {
        ip: '1.2.3.4',
        device: 'Chrome on Windows',
        sessionId: 'sess_123'
    }
});

// Delivers immediately regardless of DND/quiet hours
// Via all enabled channels
```

## Limits

| Limit                        | Default        |
| ---------------------------- | -------------- |
| Max notifications per hour   | 100 (per user) |
| Max actions per notification | 4              |
| Max title length             | 100 chars      |
| Max body length              | 1000 chars     |
| History retention            | 30 days        |
| Digest batch size            | 50             |
