# Hooks - Bidirectional Webhook Service

## Summary

Hooks provides a unified system for sending and receiving webhooks. Inbound
hooks receive events from external services (Stripe, GitHub, etc.) with
automatic signature verification. Outbound hooks send events to external
URLs when things happen in your apps.

## Why a Service

Hooks is a convenience layer, not a core primitive:
- Signature verification is provider-specific (not universal)
- Retry logic and replay are app-level concerns
- Builds on top of kernel networking (`net/`)
- Fits alongside other services (forms, media, pdf)

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      EXTERNAL SERVICES                       │
│              Stripe    GitHub    Shopify    Custom           │
└──────┬─────────┬─────────┬─────────┬───────────────────────┘
       │         │         │         │
       │         │  POST   │         │
       ▼         ▼         ▼         ▼
┌─────────────────────────────────────────────────────────────┐
│                    INBOUND HOOKS                             │
│  /_hooks/{app}/{provider}                                   │
│                                                             │
│  1. Receive POST                                            │
│  2. Verify signature (provider-specific)                    │
│  3. Parse payload                                           │
│  4. Store event                                             │
│  5. Route to app handler                                    │
│  6. Retry on failure                                        │
└──────┬──────────────────────────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────────────────────────┐
│                      APP HANDLER                             │
│  api/hooks/{provider}.js                                    │
└─────────────────────────────────────────────────────────────┘


┌─────────────────────────────────────────────────────────────┐
│                      APP EVENT                               │
│  User signs up, order created, etc.                         │
└──────┬──────────────────────────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────────────────────────┐
│                    OUTBOUND HOOKS                            │
│                                                             │
│  1. App triggers event                                      │
│  2. Match against registered hooks                          │
│  3. POST to configured URLs                                 │
│  4. Retry on failure                                        │
│  5. Log delivery status                                     │
└──────┬─────────────────────────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────────────────────────┐
│                    EXTERNAL ENDPOINTS                        │
│         Slack    Zapier    Custom API    n8n                │
└─────────────────────────────────────────────────────────────┘
```

## Inbound Hooks

### Supported Providers

| Provider | Signature Method | Status |
|----------|-----------------|--------|
| Stripe | `Stripe-Signature` header (HMAC-SHA256) | Primary |
| GitHub | `X-Hub-Signature-256` header (HMAC-SHA256) | Primary |
| Shopify | `X-Shopify-Hmac-Sha256` header | Primary |
| Paddle | `Paddle-Signature` header | Planned |
| Twilio | Request validation | Planned |
| Slack | `X-Slack-Signature` header | Planned |
| Custom | Configurable or none | Primary |

### Configuration

In `app.json`:

```json
{
  "hooks": {
    "inbound": {
      "stripe": {
        "secret": "$STRIPE_WEBHOOK_SECRET",
        "handler": "api/hooks/stripe.js"
      },
      "github": {
        "secret": "$GITHUB_WEBHOOK_SECRET",
        "handler": "api/hooks/github.js"
      },
      "custom": {
        "handler": "api/hooks/custom.js",
        "verifySignature": false
      }
    }
  }
}
```

### Endpoint URLs

Each app gets dedicated inbound hook URLs:

```
https://{app}.{domain}/_hooks/stripe
https://{app}.{domain}/_hooks/github
https://{app}.{domain}/_hooks/custom
```

Configure these URLs in your external service's webhook settings.

### Handler Implementation

```javascript
// api/hooks/stripe.js
module.exports = async (event) => {
    // event is already verified and parsed
    console.log('Event type:', event.type);
    console.log('Event data:', event.data);

    switch (event.type) {
        case 'checkout.session.completed':
            await handleCheckoutComplete(event.data);
            break;

        case 'customer.subscription.deleted':
            await handleSubscriptionCanceled(event.data);
            break;

        case 'invoice.payment_failed':
            await handlePaymentFailed(event.data);
            break;
    }

    // Return 200 to acknowledge (or throw to trigger retry)
    return { received: true };
};
```

### Event Storage

All inbound events are stored for debugging and replay:

```sql
CREATE TABLE hook_events (
    id TEXT PRIMARY KEY,
    app_uuid TEXT NOT NULL,
    provider TEXT NOT NULL,
    event_type TEXT,
    payload_json TEXT NOT NULL,
    headers_json TEXT,
    received_at INTEGER NOT NULL,
    processed_at INTEGER,
    status TEXT DEFAULT 'pending',  -- pending, processed, failed
    error TEXT,
    attempts INTEGER DEFAULT 0,
    FOREIGN KEY (app_uuid) REFERENCES apps(uuid)
);

CREATE INDEX idx_hook_events_app ON hook_events(app_uuid, received_at);
CREATE INDEX idx_hook_events_status ON hook_events(status);
```

### Retry Logic

Failed handlers are retried with exponential backoff:

| Attempt | Delay |
|---------|-------|
| 1 | Immediate |
| 2 | 1 minute |
| 3 | 5 minutes |
| 4 | 30 minutes |
| 5 | 2 hours |

After 5 failures, event is marked `failed` and owner is notified.

### Replay

```javascript
// Replay a specific event
await fazt.services.hooks.replay(eventId);

// Replay all failed events for a provider
await fazt.services.hooks.replayFailed('stripe');

// Replay events in a time range
await fazt.services.hooks.replayRange({
    provider: 'stripe',
    from: '2024-01-01T00:00:00Z',
    to: '2024-01-02T00:00:00Z'
});
```

## Outbound Hooks

### Configuration

Register outbound hooks via API or CLI:

```javascript
// Register a hook
await fazt.services.hooks.register({
    event: 'user.created',
    url: 'https://hooks.slack.com/services/xxx',
    secret: 'my-signing-secret',  // Optional, for HMAC signature
    headers: {                     // Optional custom headers
        'X-Custom-Header': 'value'
    }
});

// Register with filtering
await fazt.services.hooks.register({
    event: 'order.*',              // Wildcard matching
    url: 'https://api.example.com/webhooks',
    filter: {                      // Only trigger if conditions match
        'data.total': { $gt: 100 }
    }
});
```

### Triggering Hooks

```javascript
// In your app code
await fazt.services.hooks.emit('user.created', {
    userId: '123',
    email: 'user@example.com',
    createdAt: new Date().toISOString()
});

// All registered hooks for 'user.created' will be called
```

### Payload Format

Outbound hooks send a standardized payload:

```json
{
  "id": "evt_abc123",
  "type": "user.created",
  "timestamp": "2024-01-15T10:30:00Z",
  "app": "myapp",
  "data": {
    "userId": "123",
    "email": "user@example.com",
    "createdAt": "2024-01-15T10:30:00Z"
  }
}
```

### Signature

If a secret is configured, outbound hooks include a signature header:

```
X-Hook-Signature: sha256=abc123...
```

Computed as: `HMAC-SHA256(secret, timestamp + "." + payload)`

Recipients can verify:

```javascript
const crypto = require('crypto');

function verifySignature(payload, signature, secret, timestamp) {
    const expected = crypto
        .createHmac('sha256', secret)
        .update(timestamp + '.' + JSON.stringify(payload))
        .digest('hex');

    return crypto.timingSafeEqual(
        Buffer.from(signature),
        Buffer.from('sha256=' + expected)
    );
}
```

### Delivery Logging

```sql
CREATE TABLE hook_deliveries (
    id TEXT PRIMARY KEY,
    app_uuid TEXT NOT NULL,
    hook_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    url TEXT NOT NULL,
    payload_json TEXT NOT NULL,
    status_code INTEGER,
    response_body TEXT,
    sent_at INTEGER NOT NULL,
    duration_ms INTEGER,
    success INTEGER DEFAULT 0,
    error TEXT,
    attempts INTEGER DEFAULT 1,
    FOREIGN KEY (app_uuid) REFERENCES apps(uuid)
);
```

### Retry Logic (Outbound)

Same as inbound: 5 attempts with exponential backoff.

## JS API

### Inbound

```javascript
// Query received events
const events = await fazt.services.hooks.events({
    provider: 'stripe',
    status: 'processed',
    limit: 50,
    since: '2024-01-01T00:00:00Z'
});

// Get specific event
const event = await fazt.services.hooks.event(eventId);

// Replay event
await fazt.services.hooks.replay(eventId);

// Replay failed events
await fazt.services.hooks.replayFailed(provider?);

// Stats
const stats = await fazt.services.hooks.stats('stripe');
// { received: 1234, processed: 1200, failed: 34, pending: 0 }
```

### Outbound

```javascript
// Register hook
const hook = await fazt.services.hooks.register({
    event: 'user.created',
    url: 'https://example.com/webhook',
    secret: 'optional-secret'
});

// List registered hooks
const hooks = await fazt.services.hooks.list();

// Update hook
await fazt.services.hooks.update(hookId, { url: 'https://new-url.com' });

// Delete hook
await fazt.services.hooks.delete(hookId);

// Emit event (triggers matching hooks)
await fazt.services.hooks.emit(eventType, data);

// Query deliveries
const deliveries = await fazt.services.hooks.deliveries({
    hookId: 'hook_123',
    success: false,
    limit: 20
});

// Retry failed delivery
await fazt.services.hooks.retryDelivery(deliveryId);
```

## HTTP API

### Inbound Endpoints

```
POST /_hooks/{provider}           # Receive webhook from provider
```

### Management Endpoints

```
GET  /api/hooks/events            # List inbound events
GET  /api/hooks/events/{id}       # Get event details
POST /api/hooks/events/{id}/replay # Replay event

GET  /api/hooks                   # List outbound hooks
POST /api/hooks                   # Register outbound hook
PUT  /api/hooks/{id}              # Update hook
DELETE /api/hooks/{id}            # Delete hook

GET  /api/hooks/deliveries        # List deliveries
POST /api/hooks/deliveries/{id}/retry # Retry delivery
```

## CLI

```bash
# Inbound
fazt hooks events --provider stripe --limit 20
fazt hooks event evt_123
fazt hooks replay evt_123
fazt hooks replay-failed --provider stripe
fazt hooks stats stripe

# Outbound
fazt hooks list
fazt hooks register user.created https://example.com/webhook
fazt hooks delete hook_123
fazt hooks deliveries --hook hook_123
fazt hooks retry-delivery del_123

# Testing
fazt hooks test stripe '{"type":"checkout.session.completed"}'
fazt hooks trigger user.created '{"userId":"123"}'
```

## Provider-Specific Verification

### Stripe

```go
func verifyStripe(payload []byte, header string, secret string) error {
    parts := parseStripeHeader(header)  // t=timestamp,v1=signature

    expectedSig := hmacSHA256(secret, parts.timestamp + "." + string(payload))

    if !hmac.Equal([]byte(parts.signature), []byte(expectedSig)) {
        return ErrInvalidSignature
    }

    // Check timestamp to prevent replay attacks
    if time.Now().Unix() - parts.timestamp > 300 {
        return ErrTimestampTooOld
    }

    return nil
}
```

### GitHub

```go
func verifyGitHub(payload []byte, header string, secret string) error {
    expectedSig := "sha256=" + hmacSHA256(secret, payload)

    if !hmac.Equal([]byte(header), []byte(expectedSig)) {
        return ErrInvalidSignature
    }

    return nil
}
```

### Shopify

```go
func verifyShopify(payload []byte, header string, secret string) error {
    expectedSig := base64(hmacSHA256(secret, payload))

    if header != expectedSig {
        return ErrInvalidSignature
    }

    return nil
}
```

## Example: Complete Stripe Integration

```javascript
// app.json
{
  "hooks": {
    "inbound": {
      "stripe": {
        "secret": "$STRIPE_WEBHOOK_SECRET",
        "handler": "api/hooks/stripe.js"
      }
    }
  }
}

// api/hooks/stripe.js
module.exports = async (event) => {
    const { type, data } = event;

    switch (type) {
        case 'checkout.session.completed': {
            const session = data.object;
            const userId = session.client_reference_id;
            const customerId = session.customer;

            // Update user with Stripe customer ID
            await fazt.storage.ds.update('users',
                { id: userId },
                { $set: { stripeCustomerId: customerId, plan: 'pro' } }
            );

            // Send welcome email
            await fazt.dev.email.send({
                to: session.customer_email,
                subject: 'Welcome to Pro!',
                html: welcomeEmailHtml
            });
            break;
        }

        case 'customer.subscription.deleted': {
            const subscription = data.object;
            const customerId = subscription.customer;

            // Downgrade user
            await fazt.storage.ds.update('users',
                { stripeCustomerId: customerId },
                { $set: { plan: 'free' } }
            );
            break;
        }

        case 'invoice.payment_failed': {
            const invoice = data.object;

            // Notify user
            await fazt.dev.email.send({
                to: invoice.customer_email,
                subject: 'Payment Failed',
                html: paymentFailedEmailHtml
            });
            break;
        }
    }

    return { received: true };
};
```

## Example: Outbound Hooks to Slack

```javascript
// Register hook to notify Slack on new orders
await fazt.services.hooks.register({
    event: 'order.created',
    url: process.env.SLACK_WEBHOOK_URL,
    transform: (event) => ({
        // Transform to Slack message format
        text: `New order #${event.data.orderId}`,
        blocks: [
            {
                type: 'section',
                text: {
                    type: 'mrkdwn',
                    text: `*New Order*\nOrder #${event.data.orderId}\nTotal: $${event.data.total}`
                }
            }
        ]
    })
});

// In your app, when order is created:
await fazt.services.hooks.emit('order.created', {
    orderId: '12345',
    total: 99.99,
    customer: 'Alice'
});

// Slack receives the transformed message automatically
```

## Limits

| Limit | Default |
|-------|---------|
| Max inbound payload size | 1 MB |
| Max outbound payload size | 256 KB |
| Event retention | 30 days |
| Delivery retention | 7 days |
| Max registered hooks per app | 50 |
| Max retries | 5 |
| Timeout per delivery | 30 seconds |

## Implementation Notes

- Inbound hooks run signature verification before touching app code
- Events stored immediately after verification, before handler runs
- Handler failures don't lose events (can replay)
- Outbound hooks use a queue with worker goroutines
- All times stored as Unix timestamps for consistency
