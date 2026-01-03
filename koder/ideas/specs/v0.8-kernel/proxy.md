# Network Proxy (Egress Control)

## Summary

Kernel-level HTTP proxy for all outbound network requests from apps.
Provides control, visibility, and safety for external API access.
All `fetch()` calls from apps go through this layer.

## Why Kernel-Level

Network egress is a security and resource boundary:
- Apps shouldn't make unconstrained external requests
- Owner needs visibility into what's being accessed
- Rate limiting prevents runaway costs
- Credential injection keeps secrets out of app code
- Caching reduces external dependencies

## Architecture

```
┌─────────┐                    ┌─────────────┐                ┌──────────┐
│   App   │ ── fazt.net.fetch ─│   Proxy     │ ── HTTP(S) ──► │ External │
└─────────┘                    │  (kernel)   │                │   API    │
                               ├─────────────┤                └──────────┘
                               │ • Auth      │
                               │ • Cache     │
                               │ • Rate Limit│
                               │ • Logging   │
                               │ • Allow/Deny│
                               └─────────────┘
```

## Usage

### Basic Fetch

```javascript
// In app runtime, fazt.net.fetch replaces global fetch
const response = await fazt.net.fetch('https://api.example.com/data');
const data = await response.json();
```

### With Options

```javascript
const response = await fazt.net.fetch('https://api.stripe.com/v1/charges', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({ amount: 1000 }),

  // Proxy options
  auth: 'STRIPE_KEY',        // Inject from secrets
  cache: '5m',               // Cache response
  retry: 3,                  // Retry on failure
  timeout: 30000             // Timeout in ms
});
```

## Credential Injection

Secrets are injected by the proxy, never exposed to app code:

```javascript
// App code - no secrets visible
await fazt.net.fetch('https://api.openai.com/v1/chat/completions', {
  method: 'POST',
  auth: 'OPENAI_KEY',        // Reference, not value
  body: JSON.stringify({ ... })
});

// Proxy injects: Authorization: Bearer sk-...
```

### Auth Modes

| Mode            | Header                                     |
| --------------- | ------------------------------------------ |
| `bearer`        | `Authorization: Bearer {secret}` (default) |
| `basic`         | `Authorization: Basic {base64}`            |
| `header:{name}` | Custom header: `{name}: {secret}`          |
| `query:{param}` | Query param: `?{param}={secret}`           |

```javascript
await fazt.net.fetch(url, {
  auth: 'API_KEY',
  authMode: 'header:X-API-Key'
});
// Adds: X-API-Key: {secret value}
```

## Response Caching

```javascript
const response = await fazt.net.fetch('https://api.weather.com/current', {
  cache: '10m'               // Cache for 10 minutes
});

// Subsequent calls within 10m return cached response
// Cache key: method + url + relevant headers
```

### Cache Control

```javascript
await fazt.net.fetch(url, {
  cache: '1h',
  cacheKey: 'weather-nyc',   // Custom cache key
  cacheVary: ['Accept'],     // Vary by these headers
});

// Force refresh
await fazt.net.fetch(url, {
  cache: '1h',
  cacheRefresh: true         // Bypass cache, update it
});
```

## Rate Limiting

Protect against runaway requests and API costs:

```javascript
// Global limit (kernel config)
fazt config set net.proxy.rateLimit 100/min

// Per-domain limits
fazt config set net.proxy.limits.api.openai.com 60/min
fazt config set net.proxy.limits.api.stripe.com 100/min
```

When limit exceeded:
- Request queued (if within buffer)
- Or rejected with 429

```javascript
try {
  await fazt.net.fetch(url);
} catch (e) {
  if (e.code === 'RATE_LIMITED') {
    // Wait and retry
  }
}
```

## Retry with Backoff

```javascript
await fazt.net.fetch(url, {
  retry: 3,                  // Max attempts
  retryDelay: 1000,          // Initial delay (ms)
  retryBackoff: 'exponential', // 1s, 2s, 4s
  retryOn: [500, 502, 503]   // Status codes to retry
});
```

## Allow/Deny Lists

Owner controls which domains apps can access:

```yaml
# kernel config
net:
  proxy:
    mode: allowlist          # or 'denylist' or 'open'
    allowlist:
      - api.stripe.com
      - api.openai.com
      - api.anthropic.com
      - '*.amazonaws.com'
    denylist:
      - localhost
      - '*.internal'
      - '10.*'
      - '192.168.*'
```

```javascript
// If domain not allowed:
await fazt.net.fetch('https://blocked.com/api');
// Error: Domain not in allowlist
```

## Request Logging

All requests logged for visibility:

```javascript
// Query logs
const logs = await fazt.net.logs({
  since: '1h',
  domain: 'api.openai.com',
  status: 'error'
});

// [
//   {
//     timestamp: ...,
//     method: 'POST',
//     url: 'https://api.openai.com/v1/...',
//     status: 200,
//     latency: 1234,
//     cached: false,
//     app: 'my-app',
//     bytes: 4567
//   }
// ]
```

### Log Retention

```yaml
net:
  proxy:
    logRetention: 7d
    logLevel: info           # debug, info, error
```

## Timeout

```javascript
await fazt.net.fetch(url, {
  timeout: 30000             // 30 seconds
});

// Default: 60 seconds
// Max: 5 minutes
```

## JS API

```javascript
fazt.net.fetch(url, options?)
// Standard fetch + proxy options:
// - auth: secret name
// - authMode: 'bearer' | 'basic' | 'header:{name}' | 'query:{param}'
// - cache: duration string
// - cacheKey: custom key
// - cacheRefresh: boolean
// - retry: number
// - retryDelay: number (ms)
// - retryBackoff: 'fixed' | 'exponential'
// - retryOn: number[]
// - timeout: number (ms)

fazt.net.logs(options?)
// Query request logs
// options: { since, until, domain, app, status, limit }
```

## CLI

```bash
# View recent requests
fazt net logs --since 1h --domain api.openai.com

# Test fetch
fazt net fetch https://api.example.com/health

# Manage allowlist
fazt net allow add api.newservice.com
fazt net allow remove api.oldservice.com
fazt net allow list

# View rate limit status
fazt net limits

# Clear cache
fazt net cache clear --domain api.weather.com
```

## Storage

```sql
CREATE TABLE kernel_net_logs (
    id TEXT PRIMARY KEY,
    app_uuid TEXT,
    method TEXT,
    url TEXT,
    domain TEXT,
    status INTEGER,
    latency_ms INTEGER,
    request_bytes INTEGER,
    response_bytes INTEGER,
    cached INTEGER,
    error TEXT,
    created_at INTEGER
);

CREATE INDEX idx_net_logs_domain ON kernel_net_logs(domain, created_at);
CREATE INDEX idx_net_logs_app ON kernel_net_logs(app_uuid, created_at);

CREATE TABLE kernel_net_cache (
    cache_key TEXT PRIMARY KEY,
    url TEXT,
    response_headers TEXT,
    response_body BLOB,
    created_at INTEGER,
    expires_at INTEGER
);
```

## Limits

| Limit                        | Default       |
| ---------------------------- | ------------- |
| `maxRequestsPerMinute`       | 100 (per app) |
| `maxRequestsPerMinuteGlobal` | 500           |
| `maxRequestBodySize`         | 10 MB         |
| `maxResponseBodySize`        | 50 MB         |
| `maxCacheSizeMB`             | 100           |
| `defaultTimeout`             | 60s           |
| `maxTimeout`                 | 5m            |
| `logRetention`               | 7d            |

## Example: External API Integration

```javascript
// api/weather.js
module.exports = async (req) => {
  const city = req.query.city;

  const response = await fazt.net.fetch(
    `https://api.weather.com/v1/current?city=${city}`,
    {
      auth: 'WEATHER_API_KEY',
      cache: '15m',
      retry: 2
    }
  );

  if (!response.ok) {
    return { status: 502, json: { error: 'Weather API failed' } };
  }

  const data = await response.json();

  return {
    json: {
      city: data.city,
      temp: data.temperature,
      conditions: data.conditions
    }
  };
};
```

## Example: Agent with Rate-Limited API

```javascript
// Agent making many API calls
async function processItems(items) {
  const results = [];

  for (const item of items) {
    try {
      const response = await fazt.net.fetch(
        'https://api.processor.com/analyze',
        {
          method: 'POST',
          auth: 'PROCESSOR_KEY',
          body: JSON.stringify(item),
          retry: 3
        }
      );

      results.push(await response.json());
    } catch (e) {
      if (e.code === 'RATE_LIMITED') {
        // Proxy handles rate limiting, wait for queue
        await sleep(1000);
      }
    }
  }

  return results;
}
```

## Security Considerations

1. **No localhost access** - Apps cannot access localhost or internal IPs
2. **HTTPS only** - HTTP requests upgraded or blocked
3. **No credential leakage** - Secrets injected by proxy, not visible to app
4. **Request signing** - Optional HMAC signing for requests
5. **Response validation** - Optional schema validation
