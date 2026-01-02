# API Profiles

## Summary

Named connections to external APIs with managed credentials, automatic auth
refresh, rate limiting, and declarative endpoint definitions. Collapses the
friction of personal API access to near-zero for users and AI agents.

## Why a Service

API Profiles extends the `fazt.dev` pattern to user-owned external services:
- `fazt.dev.*` connects to platform-provided services (Stripe, Twilio)
- `fazt.services.api.*` connects to user-defined external APIs
- Reuses kernel primitives: vault (secrets), oauth (auth), net (fetch)
- Enables "personal API gateway" use case

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        APP CODE                              │
│  const bank = await fazt.api.connect('my-bank');            │
│  const balance = await bank.call('balance');                │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                    API PROFILES SERVICE                      │
│                                                              │
│  1. Load profile definition                                 │
│  2. Retrieve credentials from vault                         │
│  3. Check/refresh auth tokens                               │
│  4. Apply rate limiting                                     │
│  5. Make request via fazt.net.fetch                         │
│  6. Handle response/errors                                  │
│  7. Log for debugging                                       │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                    EXTERNAL API                              │
│            Bank API    Fitness API    Calendar API          │
└─────────────────────────────────────────────────────────────┘
```

## Profile Definition

```javascript
{
    "name": "my-bank",
    "description": "Personal bank account API",
    "baseUrl": "https://api.mybank.com/v2",

    // Authentication
    "auth": {
        "type": "oauth",              // oauth, apikey, basic, bearer, custom
        "clientId": "$BANK_CLIENT_ID",
        "clientSecret": "$BANK_CLIENT_SECRET",
        "authUrl": "https://auth.mybank.com/oauth/authorize",
        "tokenUrl": "https://auth.mybank.com/oauth/token",
        "scopes": ["accounts:read", "transactions:read"],
        "refreshEnabled": true
    },

    // Rate limiting
    "rateLimit": {
        "requests": 100,
        "per": "minute"
    },

    // Default headers
    "headers": {
        "Accept": "application/json",
        "X-API-Version": "2024-01"
    },

    // Endpoint definitions
    "endpoints": {
        "balance": {
            "method": "GET",
            "path": "/accounts/balance"
        },
        "transactions": {
            "method": "GET",
            "path": "/accounts/{accountId}/transactions",
            "params": {
                "accountId": { "required": true },
                "from": { "type": "date" },
                "to": { "type": "date" },
                "limit": { "type": "number", "default": 50 }
            }
        },
        "transfer": {
            "method": "POST",
            "path": "/transfers",
            "body": {
                "from": { "required": true },
                "to": { "required": true },
                "amount": { "required": true, "type": "number" },
                "currency": { "default": "USD" }
            },
            "confirm": true  // Requires fazt.halt() confirmation
        }
    }
}
```

## Auth Types

### OAuth 2.0

```javascript
{
    "auth": {
        "type": "oauth",
        "clientId": "$CLIENT_ID",
        "clientSecret": "$CLIENT_SECRET",
        "authUrl": "https://provider.com/oauth/authorize",
        "tokenUrl": "https://provider.com/oauth/token",
        "scopes": ["read", "write"],
        "refreshEnabled": true,
        "pkce": true  // Optional: use PKCE flow
    }
}
```

### API Key

```javascript
{
    "auth": {
        "type": "apikey",
        "key": "$API_KEY",
        "header": "X-API-Key"       // or "query": "api_key"
    }
}
```

### Bearer Token

```javascript
{
    "auth": {
        "type": "bearer",
        "token": "$BEARER_TOKEN"
    }
}
```

### Basic Auth

```javascript
{
    "auth": {
        "type": "basic",
        "username": "$USERNAME",
        "password": "$PASSWORD"
    }
}
```

### Custom Auth

```javascript
{
    "auth": {
        "type": "custom",
        "handler": "api/auth/custom-api.js"  // Your auth logic
    }
}

// api/auth/custom-api.js
module.exports = async (request, credentials) => {
    // Add custom headers, sign request, etc.
    request.headers['X-Signature'] = sign(request.body, credentials.secret);
    return request;
};
```

## Usage

### Setup Profile

```javascript
// Register profile from JSON
await fazt.services.api.register({
    name: 'my-bank',
    baseUrl: 'https://api.mybank.com/v2',
    auth: {
        type: 'oauth',
        clientId: process.env.BANK_CLIENT_ID,
        // ... auth config
    },
    endpoints: {
        balance: { method: 'GET', path: '/accounts/balance' },
        // ...
    }
});

// Or load from file
await fazt.services.api.registerFromFile('profiles/my-bank.json');
```

### Connect and Call

```javascript
// Get connected client
const bank = await fazt.api.connect('my-bank');

// Call defined endpoint
const balance = await bank.call('balance');
console.log(balance);  // { available: 1234.56, currency: 'USD' }

// Call with parameters
const txns = await bank.call('transactions', {
    accountId: 'acct_123',
    from: '2024-01-01',
    limit: 100
});

// Raw request (when endpoint not predefined)
const custom = await bank.request({
    method: 'GET',
    path: '/accounts/summary',
    query: { year: 2024 }
});
```

### OAuth Flow

```javascript
// For OAuth profiles, first authorize:
const authUrl = await fazt.services.api.authorize('my-bank', {
    redirectUri: 'https://myapp.fazt.sh/callback'
});
// Redirect user to authUrl

// Handle callback
await fazt.services.api.handleCallback('my-bank', {
    code: req.query.code,
    state: req.query.state
});

// Now ready to use
const bank = await fazt.api.connect('my-bank');
```

### Token Refresh

Automatic. When access token expires:
1. Service detects 401 response
2. Uses refresh token to get new access token
3. Retries original request
4. Stores new tokens in vault

## Rate Limiting

```javascript
// Profile defines limits
{
    "rateLimit": {
        "requests": 100,
        "per": "minute"
    }
}

// Service tracks usage
const status = await fazt.services.api.rateLimit('my-bank');
// { remaining: 87, resetAt: 1705312260000 }

// Calls automatically wait if limit reached
// Or throws if wait would exceed timeout
```

## Caching

```javascript
// Per-endpoint cache config
{
    "endpoints": {
        "balance": {
            "method": "GET",
            "path": "/accounts/balance",
            "cache": {
                "ttl": "5m",           // Cache for 5 minutes
                "key": "balance"       // Cache key (default: path + params)
            }
        }
    }
}

// Bypass cache
const balance = await bank.call('balance', {}, { cache: false });

// Clear cache
await fazt.services.api.clearCache('my-bank');
await fazt.services.api.clearCache('my-bank', 'balance');
```

## Error Handling

```javascript
try {
    const result = await bank.call('transfer', {
        from: 'acct_123',
        to: 'acct_456',
        amount: 100
    });
} catch (error) {
    if (error.code === 'RATE_LIMITED') {
        console.log(`Retry after ${error.retryAfter}ms`);
    } else if (error.code === 'AUTH_EXPIRED') {
        // Token refresh failed, need reauthorization
        const authUrl = await fazt.services.api.authorize('my-bank');
    } else if (error.code === 'API_ERROR') {
        console.log(`API returned ${error.status}: ${error.message}`);
    }
}
```

## JS API

```javascript
// Profile management
fazt.services.api.register(profile)
fazt.services.api.registerFromFile(path)
fazt.services.api.list()
fazt.services.api.get(name)
fazt.services.api.update(name, updates)
fazt.services.api.delete(name)
fazt.services.api.test(name)          // Test connectivity

// OAuth flow
fazt.services.api.authorize(name, options?)
// options: { redirectUri, state, scopes }
fazt.services.api.handleCallback(name, params)
// params: { code, state }

// Connection
fazt.api.connect(name)                // Returns bound client
// client.call(endpoint, params?, options?)
// client.request(options)
// client.status()

// Rate limiting
fazt.services.api.rateLimit(name)     // Get current status

// Caching
fazt.services.api.clearCache(name, endpoint?)

// Logs
fazt.services.api.logs(name, options?)
// options: { limit, since, status }
```

## CLI

```bash
# Profile management
fazt api list
fazt api add my-bank --from profiles/bank.json
fazt api show my-bank
fazt api test my-bank
fazt api remove my-bank

# OAuth authorization
fazt api authorize my-bank
# Opens browser, handles callback

# Test endpoint
fazt api call my-bank balance
fazt api call my-bank transactions --accountId acct_123 --limit 10

# Rate limit status
fazt api rate-limit my-bank

# Logs
fazt api logs my-bank --limit 20
fazt api logs my-bank --errors-only

# Cache
fazt api cache clear my-bank
```

## HTTP API

```
GET    /api/profiles                  # List profiles
POST   /api/profiles                  # Register profile
GET    /api/profiles/{name}           # Get profile
PUT    /api/profiles/{name}           # Update profile
DELETE /api/profiles/{name}           # Delete profile
POST   /api/profiles/{name}/test      # Test connectivity

GET    /api/profiles/{name}/authorize # Start OAuth flow
POST   /api/profiles/{name}/callback  # Handle OAuth callback

POST   /api/profiles/{name}/call      # Call endpoint
# Body: { endpoint, params }

GET    /api/profiles/{name}/rate-limit
GET    /api/profiles/{name}/logs
DELETE /api/profiles/{name}/cache
```

## Storage

```sql
CREATE TABLE api_profiles (
    id TEXT PRIMARY KEY,
    app_uuid TEXT,
    name TEXT NOT NULL,
    config_json TEXT NOT NULL,        -- Profile definition (sans secrets)
    created_at INTEGER,
    updated_at INTEGER,
    UNIQUE(app_uuid, name)
);

CREATE TABLE api_tokens (
    profile_id TEXT PRIMARY KEY,
    access_token_encrypted TEXT,
    refresh_token_encrypted TEXT,
    expires_at INTEGER,
    scopes TEXT,
    updated_at INTEGER,
    FOREIGN KEY (profile_id) REFERENCES api_profiles(id)
);

CREATE TABLE api_logs (
    id TEXT PRIMARY KEY,
    profile_id TEXT NOT NULL,
    endpoint TEXT,
    method TEXT,
    path TEXT,
    status_code INTEGER,
    duration_ms INTEGER,
    error TEXT,
    created_at INTEGER,
    FOREIGN KEY (profile_id) REFERENCES api_profiles(id)
);

CREATE INDEX idx_api_logs_profile ON api_logs(profile_id, created_at);

CREATE TABLE api_cache (
    profile_id TEXT NOT NULL,
    cache_key TEXT NOT NULL,
    response_json TEXT,
    expires_at INTEGER,
    PRIMARY KEY (profile_id, cache_key)
);
```

## Predefined Profiles

Common API profiles bundled with Fazt:

```javascript
// Built-in profiles (just add credentials)
await fazt.services.api.registerBuiltin('github', {
    token: process.env.GITHUB_TOKEN
});

await fazt.services.api.registerBuiltin('openai', {
    apiKey: process.env.OPENAI_API_KEY
});

await fazt.services.api.registerBuiltin('spotify', {
    clientId: process.env.SPOTIFY_CLIENT_ID,
    clientSecret: process.env.SPOTIFY_CLIENT_SECRET
});
```

Built-in profiles:
- `github` - GitHub API
- `openai` - OpenAI API
- `anthropic` - Anthropic API
- `spotify` - Spotify API
- `notion` - Notion API
- `linear` - Linear API
- `todoist` - Todoist API
- `google-calendar` - Google Calendar
- `google-fitness` - Google Fit
- `strava` - Strava API
- `withings` - Withings Health API

## Example: Personal Dashboard

```javascript
// Register APIs
await fazt.services.api.register({
    name: 'bank',
    baseUrl: 'https://api.mybank.com',
    auth: { type: 'oauth', ... },
    endpoints: {
        balance: { method: 'GET', path: '/balance' }
    }
});

await fazt.services.api.registerBuiltin('google-calendar', {
    clientId: process.env.GOOGLE_CLIENT_ID,
    clientSecret: process.env.GOOGLE_CLIENT_SECRET
});

await fazt.services.api.registerBuiltin('strava', {
    clientId: process.env.STRAVA_CLIENT_ID,
    clientSecret: process.env.STRAVA_CLIENT_SECRET
});

// Dashboard handler
module.exports = async (request) => {
    const [bank, calendar, strava] = await Promise.all([
        fazt.api.connect('bank'),
        fazt.api.connect('google-calendar'),
        fazt.api.connect('strava')
    ]);

    const [balance, events, activities] = await Promise.all([
        bank.call('balance'),
        calendar.call('events', { maxResults: 5 }),
        strava.call('activities', { per_page: 5 })
    ]);

    return {
        finances: {
            balance: balance.available,
            currency: balance.currency
        },
        calendar: events.items.map(e => ({
            title: e.summary,
            when: e.start.dateTime
        })),
        fitness: activities.map(a => ({
            type: a.type,
            distance: a.distance,
            date: a.start_date
        }))
    };
};
```

## Example: AI Agent API Access

```javascript
// Agent can query any registered API
const agent = fazt.ai.complete(`
    Using the available APIs, answer: What's my bank balance
    and do I have any meetings today?
`, {
    tools: [
        fazt.api.asTool('bank'),      // Expose bank API as tool
        fazt.api.asTool('google-calendar')
    ]
});

// Agent receives tool definitions:
// - bank.balance: Get account balance
// - bank.transactions: Get recent transactions
// - google-calendar.events: Get calendar events
// - google-calendar.create: Create calendar event
```

## Example: Webhook-Triggered Sync

```javascript
// Sync data when webhook received
fazt.events.on('hooks.strava.activity.created', async (event) => {
    const strava = await fazt.api.connect('strava');
    const activity = await strava.call('activity', {
        id: event.data.object_id
    });

    // Store in local database
    await fazt.storage.ds.insert('activities', {
        externalId: activity.id,
        type: activity.type,
        distance: activity.distance,
        duration: activity.moving_time,
        date: activity.start_date
    });
});
```

## Security Considerations

- All tokens stored encrypted in vault
- Profile definitions don't contain secrets (reference $ENV vars)
- Sensitive endpoints can require `confirm: true` (triggers `fazt.halt()`)
- OAuth state parameter prevents CSRF
- Token refresh happens server-side only
- Rate limiting prevents abuse

## Limits

| Limit | Default |
|-------|---------|
| Max profiles per app | 50 |
| Max endpoints per profile | 100 |
| Request timeout | 30 seconds |
| Max response size | 10 MB |
| Log retention | 7 days |
| Cache TTL max | 24 hours |
