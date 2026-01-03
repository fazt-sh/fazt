# Devices - External Service Abstraction

## Summary

Devices (`/dev/*`) provide a unified interface to external services. Like Unix
device files abstract hardware, Fazt devices abstract third-party APIs. Apps
use clean, consistent interfaces; the kernel handles provider-specific details.

## Why Kernel-Level

Devices are fundamental infrastructure:
- Credential injection must be secure (kernel-managed)
- Rate limiting and retries are cross-cutting concerns
- Provider switching should be transparent to apps
- Follows OS architecture (`/dev/` for hardware abstraction)

## Philosophy

```
┌─────────────────────────────────────────────────────────────┐
│  App                                                         │
│  const customer = await fazt.dev.billing.createCustomer()   │
│                                                             │
│  App doesn't know or care if this is Stripe or Paddle       │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│  Kernel Device Layer                                         │
│                                                             │
│  1. Route to configured provider (Stripe)                   │
│  2. Inject credentials from config                          │
│  3. Handle rate limits, retries                             │
│  4. Normalize response to device interface                  │
│  5. Log for audit                                           │
└─────────────────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│  External Provider (Stripe API)                              │
└─────────────────────────────────────────────────────────────┘
```

## Available Devices

| Device    | Purpose                 | Providers                             |
| --------- | ----------------------- | ------------------------------------- |
| `billing` | Payments, subscriptions | Stripe, Paddle, LemonSqueezy          |
| `sms`     | Text messaging          | Twilio, MessageBird, Vonage           |
| `email`   | Transactional email     | SendGrid, Postmark, AWS SES           |
| `oauth`   | Social login            | Google, GitHub, Apple, Discord        |
| `storage` | Object storage          | (Already exists as `fazt.storage.s3`) |
| `ai`      | LLM inference           | (Already exists as `fazt.ai`)         |

Note: `storage` and `ai` are already implemented as top-level namespaces.
They're conceptually devices but predate this spec.

## Configuration

Devices are configured at the system level:

```bash
# Billing
fazt config set dev.billing.provider stripe
fazt config set dev.billing.secret_key sk_live_xxx
fazt config set dev.billing.publishable_key pk_live_xxx
fazt config set dev.billing.webhook_secret whsec_xxx

# SMS
fazt config set dev.sms.provider twilio
fazt config set dev.sms.account_sid ACxxx
fazt config set dev.sms.auth_token xxx
fazt config set dev.sms.from_number +1234567890

# Email
fazt config set dev.email.provider sendgrid
fazt config set dev.email.api_key SG.xxx
fazt config set dev.email.from_address hello@example.com
fazt config set dev.email.from_name "My App"

# OAuth (multiple providers allowed)
fazt config set dev.oauth.google.client_id xxx
fazt config set dev.oauth.google.client_secret xxx
fazt config set dev.oauth.github.client_id xxx
fazt config set dev.oauth.github.client_secret xxx
```

## Device: Billing

Unified payment and subscription interface.

### Providers

| Provider     | Status  | Notes                    |
| ------------ | ------- | ------------------------ |
| Stripe       | Primary | Full support             |
| Paddle       | Planned | MoR (Merchant of Record) |
| LemonSqueezy | Planned | MoR, simpler             |

### Interface

```javascript
// Customers
const customer = await fazt.dev.billing.customers.create({
    email: 'user@example.com',
    name: 'Alice Smith',
    metadata: { userId: '123' }
});

const customer = await fazt.dev.billing.customers.get(customerId);
const customer = await fazt.dev.billing.customers.update(customerId, { name: 'Alice Jones' });
const customers = await fazt.dev.billing.customers.list({ limit: 10 });

// Products & Prices
const products = await fazt.dev.billing.products.list();
const prices = await fazt.dev.billing.prices.list({ product: productId });

// Subscriptions
const subscription = await fazt.dev.billing.subscriptions.create({
    customer: customerId,
    price: priceId,
    metadata: { plan: 'pro' }
});

const subscription = await fazt.dev.billing.subscriptions.get(subscriptionId);
const subscription = await fazt.dev.billing.subscriptions.cancel(subscriptionId);
const subscriptions = await fazt.dev.billing.subscriptions.list({ customer: customerId });

// One-time Payments
const session = await fazt.dev.billing.checkout.create({
    customer: customerId,
    lineItems: [{ price: priceId, quantity: 1 }],
    successUrl: '/success',
    cancelUrl: '/cancel'
});
// Returns: { url: 'https://checkout.stripe.com/...' }

// Invoices
const invoices = await fazt.dev.billing.invoices.list({ customer: customerId });
const invoice = await fazt.dev.billing.invoices.get(invoiceId);

// Portal (customer self-service)
const portal = await fazt.dev.billing.portal.create({
    customer: customerId,
    returnUrl: '/account'
});
// Returns: { url: 'https://billing.stripe.com/...' }
```

### Webhook Events

Billing events are delivered via the hooks system:

```javascript
// api/webhooks/billing.js
module.exports = async (event) => {
    switch (event.type) {
        case 'subscription.created':
            await onSubscriptionCreated(event.data);
            break;
        case 'subscription.canceled':
            await onSubscriptionCanceled(event.data);
            break;
        case 'payment.succeeded':
            await onPaymentSucceeded(event.data);
            break;
        case 'payment.failed':
            await onPaymentFailed(event.data);
            break;
    }
};
```

### Normalized Event Types

| Event                   | Description          |
| ----------------------- | -------------------- |
| `customer.created`      | New customer         |
| `subscription.created`  | New subscription     |
| `subscription.updated`  | Subscription changed |
| `subscription.canceled` | Subscription ended   |
| `payment.succeeded`     | Payment completed    |
| `payment.failed`        | Payment failed       |
| `invoice.created`       | Invoice generated    |
| `invoice.paid`          | Invoice paid         |

## Device: SMS

Text messaging interface.

### Providers

| Provider    | Status  |
| ----------- | ------- |
| Twilio      | Primary |
| MessageBird | Planned |
| Vonage      | Planned |

### Interface

```javascript
// Send SMS
const message = await fazt.dev.sms.send({
    to: '+1234567890',
    body: 'Your verification code is 123456'
});
// Returns: { id: 'SM...', status: 'sent' }

// Send with options
const message = await fazt.dev.sms.send({
    to: '+1234567890',
    body: 'Your order has shipped!',
    from: '+0987654321',  // Override default from number
    statusCallback: '/api/sms-status'  // Webhook for delivery status
});

// Check status
const status = await fazt.dev.sms.status(messageId);
// Returns: { id, status: 'delivered' | 'failed' | 'sent', errorCode? }

// List messages (recent)
const messages = await fazt.dev.sms.list({ limit: 20 });
```

### Receiving SMS (Inbound)

Configure inbound SMS webhook in Twilio to point to `/_dev/sms/inbound`:

```javascript
// api/sms-inbound.js (registered via hooks)
module.exports = async (message) => {
    console.log(`From: ${message.from}`);
    console.log(`Body: ${message.body}`);

    // Reply
    return {
        reply: 'Thanks for your message!'
    };
};
```

## Device: Email

Transactional email interface.

### Providers

| Provider | Status  |
| -------- | ------- |
| SendGrid | Primary |
| Postmark | Planned |
| AWS SES  | Planned |

### Interface

```javascript
// Send simple email
await fazt.dev.email.send({
    to: 'user@example.com',
    subject: 'Welcome!',
    text: 'Thanks for signing up.',
    html: '<h1>Thanks for signing up!</h1>'
});

// Send with options
await fazt.dev.email.send({
    to: ['user1@example.com', 'user2@example.com'],
    cc: 'manager@example.com',
    bcc: 'archive@example.com',
    from: 'support@example.com',  // Override default
    replyTo: 'help@example.com',
    subject: 'Your order #12345',
    html: orderEmailHtml,
    attachments: [
        {
            filename: 'invoice.pdf',
            content: pdfBuffer,
            contentType: 'application/pdf'
        }
    ]
});

// Send using template (provider-specific templates)
await fazt.dev.email.sendTemplate({
    to: 'user@example.com',
    templateId: 'd-abc123',  // SendGrid template ID
    dynamicData: {
        name: 'Alice',
        orderNumber: '12345'
    }
});

// Check email status (if supported by provider)
const status = await fazt.dev.email.status(messageId);
```

## Device: OAuth

Social login and OAuth integration.

### Providers

| Provider  | Status  |
| --------- | ------- |
| Google    | Primary |
| GitHub    | Primary |
| Apple     | Planned |
| Discord   | Planned |
| Twitter/X | Planned |

### Interface

```javascript
// Generate authorization URL
const authUrl = fazt.dev.oauth.authorize({
    provider: 'google',
    redirectUri: '/auth/callback',
    scopes: ['email', 'profile'],
    state: 'random-state-string'
});
// Returns: 'https://accounts.google.com/oauth/authorize?...'

// Exchange code for tokens (in callback handler)
const tokens = await fazt.dev.oauth.callback({
    provider: 'google',
    code: req.query.code,
    redirectUri: '/auth/callback'
});
// Returns: { accessToken, refreshToken?, expiresIn, tokenType }

// Get user info
const user = await fazt.dev.oauth.userinfo({
    provider: 'google',
    accessToken: tokens.accessToken
});
// Returns: { id, email, name, picture, ... }

// Refresh token
const newTokens = await fazt.dev.oauth.refresh({
    provider: 'google',
    refreshToken: tokens.refreshToken
});
```

### Example: Complete OAuth Flow

```javascript
// api/auth/[provider].js
module.exports = async (req) => {
    const provider = req.params.provider;

    const authUrl = fazt.dev.oauth.authorize({
        provider,
        redirectUri: `/auth/${provider}/callback`,
        scopes: provider === 'google'
            ? ['email', 'profile']
            : ['user:email'],
        state: generateState()
    });

    return { redirect: authUrl };
};

// api/auth/[provider]/callback.js
module.exports = async (req) => {
    const provider = req.params.provider;
    const { code, state } = req.query;

    // Verify state...

    const tokens = await fazt.dev.oauth.callback({
        provider,
        code,
        redirectUri: `/auth/${provider}/callback`
    });

    const userInfo = await fazt.dev.oauth.userinfo({
        provider,
        accessToken: tokens.accessToken
    });

    // Create or update user in your database
    const user = await upsertUser({
        provider,
        providerId: userInfo.id,
        email: userInfo.email,
        name: userInfo.name
    });

    // Create session
    const session = await createSession(user.id);

    return {
        redirect: '/dashboard',
        cookies: { session: session.token }
    };
};
```

## Error Handling

All devices use consistent error types:

```javascript
try {
    await fazt.dev.billing.customers.create({ email: 'invalid' });
} catch (e) {
    if (e.code === 'DEVICE_VALIDATION_ERROR') {
        // Invalid input
        console.log(e.field, e.message);
    }
    if (e.code === 'DEVICE_PROVIDER_ERROR') {
        // Provider returned error
        console.log(e.providerCode, e.message);
    }
    if (e.code === 'DEVICE_NOT_CONFIGURED') {
        // Device not set up
        console.log('Billing not configured');
    }
    if (e.code === 'DEVICE_RATE_LIMITED') {
        // Hit provider rate limit
        console.log('Retry after:', e.retryAfter);
    }
}
```

## Built-in Features

All devices automatically provide:

### Credential Injection
```go
// Kernel injects credentials before calling provider
func (d *BillingDevice) CreateCustomer(input CreateCustomerInput) (*Customer, error) {
    client := stripe.New(d.config.SecretKey)  // Injected from config
    // ...
}
```

### Rate Limiting
```go
// Kernel respects provider rate limits
if response.StatusCode == 429 {
    retryAfter := parseRetryAfter(response)
    return nil, &DeviceError{
        Code: "DEVICE_RATE_LIMITED",
        RetryAfter: retryAfter,
    }
}
```

### Automatic Retries
```go
// Kernel retries transient failures
func (d *Device) callWithRetry(fn func() error) error {
    for attempt := 0; attempt < 3; attempt++ {
        err := fn()
        if err == nil || !isTransient(err) {
            return err
        }
        time.Sleep(backoff(attempt))
    }
    return err
}
```

### Audit Logging
```go
// All device calls logged
func (d *Device) log(operation string, input, output interface{}, err error) {
    d.events.Emit("device.call", DeviceCallEvent{
        Device: d.name,
        Operation: operation,
        Input: sanitize(input),  // Remove sensitive fields
        Success: err == nil,
        Timestamp: time.Now(),
    })
}
```

## CLI

```bash
# List configured devices
fazt dev list
# billing: stripe (configured)
# sms: twilio (configured)
# email: sendgrid (configured)
# oauth.google: (configured)
# oauth.github: (configured)

# Test device connectivity
fazt dev test billing
# Billing (Stripe): Connected
# - Account: acct_xxx
# - Mode: live

fazt dev test sms
# SMS (Twilio): Connected
# - Account: ACxxx
# - From: +1234567890

# View device logs
fazt dev logs billing --last 1h

# Check rate limit status
fazt dev limits billing
# Stripe rate limit: 95/100 requests remaining
# Resets in: 45 seconds
```

## Future Devices

Potential additions:

| Device      | Purpose                                     |
| ----------- | ------------------------------------------- |
| `push`      | Push notifications (FCM, APNs)              |
| `analytics` | Event tracking (Mixpanel, Amplitude)        |
| `maps`      | Geocoding, directions (Google Maps, Mapbox) |
| `cdn`       | Asset delivery (Cloudflare, BunnyCDN)       |
| `search`    | External search (Algolia, Typesense)        |

## Implementation Notes

- Each device is a Go package under `pkg/kernel/dev/`
- Providers implement a common interface per device
- Config stored in kernel config table
- Credentials never exposed to JS runtime (only kernel)
- Device calls go through kernel, not direct HTTP from apps
