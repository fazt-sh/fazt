# Custom Domain Mapping

## Summary

Map any external domain to any Fazt app. Enables white-labeling, custom
domains for clients, and decoupling apps from the primary Fazt domain.

## How It Works

```
blog.example.com  ─┐
writings.me       ─┼──► app_x9z2k (My Blog App)
posts.alice.dev   ─┘
```

## CLI

```bash
# Map domain to app
fazt net domain map blog.example.com app_x9z2k

# Set as primary (used for canonical URLs)
fazt net domain map blog.example.com app_x9z2k --primary

# List mappings
fazt net domain list

# Remove mapping
fazt net domain unmap blog.example.com
```

## DNS Setup

User must configure DNS:

```
blog.example.com  A     <fazt-server-ip>
                  AAAA  <fazt-server-ipv6>
```

Or for Cloudflare proxy:

```
blog.example.com  CNAME  fazt.example.com
```

## HTTPS

CertMagic automatically provisions certificates:

1. Request arrives for `blog.example.com`
2. Kernel checks domain_mappings table
3. If mapped and no cert: issue via Let's Encrypt
4. Serve with HTTPS

## Storage

```sql
CREATE TABLE domain_mappings (
    domain TEXT PRIMARY KEY,
    app_uuid TEXT,
    is_primary BOOLEAN DEFAULT FALSE,
    ssl_provisioned BOOLEAN DEFAULT FALSE,
    created_at INTEGER
);
```

## JS Runtime

```javascript
// Get current request domain
const domain = fazt.net.domain.current();
// "blog.example.com"

// Check if primary domain
if (fazt.net.domain.isPrimary()) {
    // Handle canonical URL
}
```

## White-Label Use Case

Agency hosts client sites:
- Client gets `client-brand.com`
- Client never sees `fazt.example.com`
- Each client is a separate app
