# Short URL Service

## Summary

Create short, shareable links. Useful for sharing, cleaner QR codes,
and click tracking. Works with any URL - internal paths or external links.

## Usage

### Create Short URL

```javascript
const short = await fazt.services.shorturl.create('/products/widget-pro-2024');
// Returns: { code: 'x7k2', url: '/_s/x7k2' }

// With custom code
const custom = await fazt.services.shorturl.create('/signup', {
  code: 'join'
});
// Returns: { code: 'join', url: '/_s/join' }
```

### Access

```
GET /_s/x7k2
→ 302 Redirect to /products/widget-pro-2024
```

### External URLs

```javascript
const ext = await fazt.services.shorturl.create('https://example.com/long/path', {
  code: 'ext'
});
// GET /_s/ext → redirects to https://example.com/long/path
```

## Click Tracking

Every access is logged:

```javascript
const stats = await fazt.services.shorturl.stats('x7k2');
// {
//   code: 'x7k2',
//   target: '/products/widget-pro-2024',
//   clicks: 142,
//   createdAt: '2024-01-15T10:30:00Z',
//   lastClickAt: '2024-01-20T14:22:00Z'
// }

// Detailed analytics
const details = await fazt.services.shorturl.clicks('x7k2', {
  limit: 100
});
// [
//   { timestamp: '...', ip: '...', userAgent: '...', referer: '...' },
//   ...
// ]
```

## Expiration

```javascript
// Expires after 7 days
const temp = await fazt.services.shorturl.create('/promo', {
  expiresIn: '7d'
});

// Expires at specific time
const scheduled = await fazt.services.shorturl.create('/flash-sale', {
  expiresAt: '2024-12-31T23:59:59Z'
});

// Max clicks then expire
const limited = await fazt.services.shorturl.create('/exclusive', {
  maxClicks: 100
});
```

After expiration, returns 410 Gone.

## QR Code Integration

Combine with QR service for cleaner codes:

```javascript
const short = await fazt.services.shorturl.create('/very/long/path/to/something');
// /_s/x7k2 is much shorter than /very/long/path/to/something
// → smaller, denser QR code

const qr = await fazt.services.qr.generate(`https://myapp.com${short.url}`);
```

## HTTP Endpoint

### Create (requires auth)

```
POST /_services/shorturl
Content-Type: application/json

{
  "target": "/products/123",
  "code": "prod123",      // optional
  "expiresIn": "30d"      // optional
}
```

### Redirect (public)

```
GET /_s/{code}
→ 302 Location: {target}
```

### Stats (requires auth)

```
GET /_services/shorturl/{code}/stats
```

## JS API

```javascript
fazt.services.shorturl.create(target, options?)
// options: { code, expiresIn, expiresAt, maxClicks }
// Returns: { code, url }

fazt.services.shorturl.get(code)
// Returns: { code, target, clicks, createdAt, expiresAt, maxClicks }

fazt.services.shorturl.update(code, options)
// options: { target, expiresAt, maxClicks }

fazt.services.shorturl.delete(code)

fazt.services.shorturl.list(options?)
// options: { limit, offset, order }
// Returns: [short urls]

fazt.services.shorturl.stats(code)
// Returns: { code, target, clicks, createdAt, lastClickAt }

fazt.services.shorturl.clicks(code, options?)
// options: { limit, offset }
// Returns: [click records]
```

## Storage

```sql
CREATE TABLE svc_shorturls (
    id TEXT PRIMARY KEY,
    app_uuid TEXT NOT NULL,
    code TEXT NOT NULL,
    target TEXT NOT NULL,
    clicks INTEGER DEFAULT 0,
    max_clicks INTEGER,
    created_at INTEGER,
    expires_at INTEGER,
    last_click_at INTEGER,
    UNIQUE(app_uuid, code)
);

CREATE TABLE svc_shorturl_clicks (
    id TEXT PRIMARY KEY,
    shorturl_id TEXT NOT NULL,
    timestamp INTEGER,
    ip TEXT,
    user_agent TEXT,
    referer TEXT,
    FOREIGN KEY (shorturl_id) REFERENCES svc_shorturls(id)
);

CREATE INDEX idx_shorturls_code ON svc_shorturls(app_uuid, code);
```

## Limits

| Limit                | Default |
| -------------------- | ------- |
| `maxPerApp`          | 10,000  |
| `maxCodeLength`      | 32      |
| `minCodeLength`      | 2       |
| `maxTargetLength`    | 2048    |
| `clickRetentionDays` | 90      |

## CLI

```bash
# Create short URL
fazt services shorturl create /products/123

# Create with custom code
fazt services shorturl create /signup --code join

# List all
fazt services shorturl list

# Show stats
fazt services shorturl stats x7k2

# Delete
fazt services shorturl delete x7k2
```

## Example: Share Links

```javascript
// api/share.js - Create shareable link
module.exports = async (req) => {
  const { path } = req.json;

  const short = await fazt.services.shorturl.create(path, {
    expiresIn: '30d'
  });

  const fullUrl = `https://${req.host}${short.url}`;

  return {
    json: {
      url: fullUrl,
      code: short.code
    }
  };
};
```

## Example: Campaign Tracking

```javascript
// Create campaign links
const campaigns = ['facebook', 'twitter', 'email'];

for (const campaign of campaigns) {
  await fazt.services.shorturl.create('/landing?utm_source=' + campaign, {
    code: `go-${campaign}`
  });
}

// /_s/go-facebook → /landing?utm_source=facebook
// /_s/go-twitter  → /landing?utm_source=twitter
// /_s/go-email    → /landing?utm_source=email

// Later, compare stats
for (const campaign of campaigns) {
  const stats = await fazt.services.shorturl.stats(`go-${campaign}`);
  console.log(`${campaign}: ${stats.clicks} clicks`);
}
```

## Example: One-Time Links

```javascript
// api/invite.js - Generate invite link
module.exports = async (req) => {
  const inviteId = generateId();

  // Store invite in database
  await fazt.storage.ds.insert('invites', {
    id: inviteId,
    email: req.json.email,
    status: 'pending'
  });

  // Create one-time link
  const short = await fazt.services.shorturl.create(`/accept-invite/${inviteId}`, {
    maxClicks: 1,
    expiresIn: '7d'
  });

  return {
    json: {
      inviteUrl: `https://${req.host}${short.url}`
    }
  };
};
```
