# Infrastructure - Cloud Substrate Abstraction

## Summary

Infrastructure (`dev.infra.*`) provides a unified interface to cloud
infrastructure providers. Like devices abstract third-party APIs, infra
abstracts the substrate that Fazt runs ON - VPS servers, DNS records,
and domain registration.

This enables full app lifecycle automation via natural language:
"Create a photography site at photos.example.com" becomes actionable.

## Why Under `dev.*`

Infrastructure providers are external services, just like Stripe or Twilio:
- Accessed via HTTP APIs
- Require API credentials
- Can fail, need health checks
- Unified under one namespace for dashboard visibility

The distinction from other devices:
- **infra**: Where Fazt RUNS (substrate)
- **billing, sms, etc**: What Fazt USES (operational)

Both are "devices" in the OS sense - external interfaces abstracted
by the kernel.

## Philosophy

```
┌─────────────────────────────────────────────────────────────┐
│  User (via Claude Desktop)                                   │
│  "Launch my app at blog.example.com"                        │
└─────────────────────┬───────────────────────────────────────┘
                      │ MCP
                      ▼
┌─────────────────────────────────────────────────────────────┐
│  Fazt Kernel                                                 │
│                                                             │
│  1. dev.infra.vps.create("hetzner", {...})                 │
│     → Spin up server, install Fazt via cloud-init          │
│                                                             │
│  2. dev.infra.dns.set("example.com", {...})                │
│     → Point blog.example.com to new IP                      │
│                                                             │
│  3. apps.create({name: "blog"})                            │
│     → Deploy to new Fazt instance                          │
│                                                             │
│  4. net.ssl.provision("blog.example.com")                  │
│     → Auto HTTPS via CertMagic                             │
└─────────────────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│  External Providers                                          │
│  Hetzner API │ Cloudflare API │ DigitalOcean API           │
└─────────────────────────────────────────────────────────────┘
```

## Available Infrastructure Types

| Type     | Purpose                 | Providers                       |
| -------- | ----------------------- | ------------------------------- |
| `vps`    | Virtual private servers | Hetzner, DigitalOcean, Vultr    |
| `dns`    | DNS record management   | Cloudflare, Hetzner DNS, DO DNS |
| `domain` | Domain registration     | Cloudflare Registrar            |

## Configuration

Infrastructure providers are configured at the system level:

```bash
# VPS Providers
fazt dev config infra.vps.hetzner --token xxx
fazt dev config infra.vps.digitalocean --token xxx
fazt dev config infra.vps.vultr --token xxx

# DNS Providers
fazt dev config infra.dns.cloudflare --token xxx
fazt dev config infra.dns.hetzner --token xxx

# Domain Provider (optional, for domain purchases)
fazt dev config infra.domain.cloudflare --token xxx

# Add managed DNS zones
fazt dev infra dns add-zone example.com --provider cloudflare
fazt dev infra dns add-zone another.com --provider hetzner

# Set default providers
fazt dev config infra.vps.default hetzner
fazt dev config infra.dns.default cloudflare
```

## Infrastructure: VPS

Virtual private server provisioning and management.

### Providers

| Provider     | Status  | API Quality | Pricing            |
| ------------ | ------- | ----------- | ------------------ |
| Hetzner      | Primary | Excellent   | Best EU value      |
| DigitalOcean | Primary | Excellent   | Popular, good docs |
| Vultr        | Planned | Good        | Global coverage    |

### Interface

```javascript
// List available server types
const types = await fazt.dev.infra.vps.types('hetzner');
// Returns: [
//   { id: 'cx22', vcpus: 2, ram: 4096, disk: 40, price: 3.79 },
//   { id: 'cx32', vcpus: 4, ram: 8192, disk: 80, price: 6.99 },
//   ...
// ]

// List available regions
const regions = await fazt.dev.infra.vps.regions('hetzner');
// Returns: [
//   { id: 'fsn1', name: 'Falkenstein', country: 'DE' },
//   { id: 'nbg1', name: 'Nuremberg', country: 'DE' },
//   { id: 'hel1', name: 'Helsinki', country: 'FI' },
//   ...
// ]

// Create a new VPS
const server = await fazt.dev.infra.vps.create({
    provider: 'hetzner',          // Optional if default set
    name: 'fazt-prod-1',
    type: 'cx22',                 // 2 vCPU, 4GB RAM
    region: 'fsn1',
    image: 'ubuntu-24.04',
    sshKeys: ['my-key'],          // SSH key names from provider
    installFazt: true,            // Auto-install Fazt via cloud-init
    faztConfig: {                 // Optional Fazt configuration
        domain: 'example.com',
        adminEmail: 'admin@example.com'
    }
});
// Returns: {
//   id: 'srv-abc123',
//   provider: 'hetzner',
//   name: 'fazt-prod-1',
//   ip: '123.45.67.89',
//   ipv6: '2a01:...',
//   status: 'running',
//   type: 'cx22',
//   region: 'fsn1',
//   createdAt: '2025-01-15T...',
//   monthlyCost: 3.79
// }

// List all managed VPSes
const servers = await fazt.dev.infra.vps.list();

// Get server details
const server = await fazt.dev.infra.vps.get('srv-abc123');

// Get server status with health check
const status = await fazt.dev.infra.vps.status('srv-abc123');
// Returns: {
//   id: 'srv-abc123',
//   status: 'running',
//   faztStatus: 'healthy',      // null if Fazt not installed
//   faztVersion: '0.8.0',
//   uptime: 86400,
//   load: [0.5, 0.3, 0.2],
//   diskUsed: 5.2,              // GB
//   diskTotal: 40,
//   memUsed: 1.2,               // GB
//   memTotal: 4
// }

// Execute command on remote server
const result = await fazt.dev.infra.vps.exec('srv-abc123', 'fazt pulse status');
// Returns: { stdout, stderr, exitCode }

// Reboot server
await fazt.dev.infra.vps.reboot('srv-abc123');

// Destroy server (triggers fazt.halt for confirmation)
await fazt.dev.infra.vps.destroy('srv-abc123');
// IMPORTANT: This is destructive. Kernel calls fazt.halt() for confirmation.
```

### Cloud-Init Template

When `installFazt: true`, the kernel generates a cloud-init script:

```yaml
#cloud-config
package_update: true
packages:
  - curl
  - sqlite3

runcmd:
  - curl -fsSL https://fazt.sh/install | sh
  - fazt server init --domain ${domain} --email ${adminEmail}
  - fazt service install --https
  - systemctl enable fazt
  - systemctl start fazt

write_files:
  - path: /etc/fazt/cloud-init-complete
    content: |
      installed_at: ${timestamp}
      installed_by: ${fazt_instance_id}
```

### SSH Key Management

```javascript
// List SSH keys registered with provider
const keys = await fazt.dev.infra.vps.sshKeys('hetzner');

// Add SSH key to provider
await fazt.dev.infra.vps.addSshKey('hetzner', {
    name: 'my-laptop',
    publicKey: 'ssh-ed25519 AAAA...'
});

// Remove SSH key
await fazt.dev.infra.vps.removeSshKey('hetzner', 'my-laptop');
```

## Infrastructure: DNS

DNS record management for managed zones.

### Providers

| Provider     | Status  | Free Tier | Notes                         |
| ------------ | ------- | --------- | ----------------------------- |
| Cloudflare   | Primary | Unlimited | Best choice, fast propagation |
| Hetzner DNS  | Primary | Yes       | Good if using Hetzner VPS     |
| DigitalOcean | Planned | Yes       | Good if using DO              |

### Prerequisite: DNS Delegation

Before Fazt can manage DNS, the domain must use the provider's nameservers.
This is a ONE-TIME manual step:

1. Add domain to Cloudflare (free account)
2. Cloudflare provides nameservers: `ns1.cloudflare.com`, `ns2.cloudflare.com`
3. Go to registrar (Namecheap, GoDaddy, etc.)
4. Change nameservers to Cloudflare's
5. Wait 1-24 hours for propagation

After this, all DNS changes are API-driven forever.

### Interface

```javascript
// List managed zones
const zones = await fazt.dev.infra.dns.zones();
// Returns: [
//   { zone: 'example.com', provider: 'cloudflare', records: 12 },
//   { zone: 'another.com', provider: 'hetzner', records: 5 }
// ]

// List records in a zone
const records = await fazt.dev.infra.dns.records('example.com');
// Returns: [
//   { type: 'A', name: '@', content: '123.45.67.89', ttl: 300, proxied: true },
//   { type: 'A', name: 'blog', content: '123.45.67.89', ttl: 300, proxied: true },
//   { type: 'MX', name: '@', content: 'mail.example.com', priority: 10, ttl: 3600 },
//   { type: 'TXT', name: '@', content: 'v=spf1 ...', ttl: 3600 },
//   ...
// ]

// Create or update a record (upsert behavior)
await fazt.dev.infra.dns.set('example.com', {
    type: 'A',
    name: 'blog',                 // 'blog' → blog.example.com, '@' → example.com
    content: '123.45.67.89',
    ttl: 300,                     // Optional, default 300
    proxied: true                 // Cloudflare only, default true
});

// Create MX record for email
await fazt.dev.infra.dns.set('example.com', {
    type: 'MX',
    name: '@',
    content: 'mail.example.com',
    priority: 10,
    ttl: 3600
});

// Create TXT record (SPF, DKIM, verification)
await fazt.dev.infra.dns.set('example.com', {
    type: 'TXT',
    name: '@',
    content: 'v=spf1 ip4:123.45.67.89 -all'
});

// Create CNAME record
await fazt.dev.infra.dns.set('example.com', {
    type: 'CNAME',
    name: 'www',
    content: 'example.com'
});

// Delete a record
await fazt.dev.infra.dns.delete('example.com', {
    type: 'A',
    name: 'old-subdomain'
});

// Bulk operations
await fazt.dev.infra.dns.setMany('example.com', [
    { type: 'A', name: 'api', content: '123.45.67.89' },
    { type: 'A', name: 'cdn', content: '123.45.67.89' },
    { type: 'TXT', name: '_dmarc', content: 'v=DMARC1; p=none' }
]);

// Check propagation status
const propagation = await fazt.dev.infra.dns.checkPropagation('blog.example.com');
// Returns: {
//   expected: '123.45.67.89',
//   results: [
//     { resolver: '8.8.8.8', value: '123.45.67.89', propagated: true },
//     { resolver: '1.1.1.1', value: '123.45.67.89', propagated: true },
//     { resolver: '9.9.9.9', value: null, propagated: false }
//   ],
//   allPropagated: false
// }
```

### Record Types

| Type    | Purpose               | Example                           |
| ------- | --------------------- | --------------------------------- |
| `A`     | IPv4 address          | `blog.example.com → 123.45.67.89` |
| `AAAA`  | IPv6 address          | `blog.example.com → 2001:db8::1`  |
| `CNAME` | Alias                 | `www.example.com → example.com`   |
| `MX`    | Mail server           | `example.com → mail.example.com`  |
| `TXT`   | Text data             | SPF, DKIM, DMARC, verification    |
| `NS`    | Nameserver            | Usually managed by provider       |
| `CAA`   | Certificate authority | Restrict who can issue SSL        |

## Infrastructure: Domain

Domain registration (optional, for end-to-end automation).

### Providers

| Provider   | Status  | Notes                          |
| ---------- | ------- | ------------------------------ |
| Cloudflare | Primary | At-cost pricing, excellent API |

### Interface

```javascript
// Check domain availability
const available = await fazt.dev.infra.domain.check('example.com');
// Returns: {
//   domain: 'example.com',
//   available: false
// }

const available = await fazt.dev.infra.domain.check('my-new-domain.com');
// Returns: {
//   domain: 'my-new-domain.com',
//   available: true,
//   price: 10.11,
//   currency: 'USD',
//   premium: false
// }

// Search for available domains
const suggestions = await fazt.dev.infra.domain.search('my-startup');
// Returns: [
//   { domain: 'my-startup.com', available: true, price: 10.11 },
//   { domain: 'my-startup.io', available: true, price: 32.00 },
//   { domain: 'my-startup.dev', available: true, price: 12.00 },
//   { domain: 'mystartup.com', available: false },
//   ...
// ]

// Register a domain (triggers fazt.halt for confirmation)
const registration = await fazt.dev.infra.domain.register({
    domain: 'my-new-domain.com',
    years: 1,
    autoRenew: true,
    privacy: true,                // WHOIS privacy
    registrant: {
        firstName: 'Alice',
        lastName: 'Smith',
        email: 'alice@example.com',
        phone: '+1.5551234567',
        address: '123 Main St',
        city: 'San Francisco',
        state: 'CA',
        postalCode: '94102',
        country: 'US'
    }
});
// IMPORTANT: This charges money. Kernel calls fazt.halt() for confirmation.
// Returns: {
//   domain: 'my-new-domain.com',
//   status: 'registered',
//   expiresAt: '2026-01-15',
//   autoRenew: true,
//   nameservers: ['ns1.cloudflare.com', 'ns2.cloudflare.com']
// }

// List owned domains
const domains = await fazt.dev.infra.domain.list();

// Get domain details
const domain = await fazt.dev.infra.domain.get('example.com');

// Update auto-renew
await fazt.dev.infra.domain.update('example.com', { autoRenew: false });

// Transfer domain (from another registrar)
await fazt.dev.infra.domain.transfer({
    domain: 'example.com',
    authCode: 'abc123',           // EPP auth code from current registrar
    registrant: { ... }
});
```

## CLI

```bash
# Configuration
fazt dev config infra.vps.hetzner --token xxx
fazt dev config infra.dns.cloudflare --token xxx

# VPS Management
fazt dev infra vps list
fazt dev infra vps types hetzner
fazt dev infra vps regions hetzner
fazt dev infra vps create hetzner --name prod-1 --type cx22 --region fsn1
fazt dev infra vps status prod-1
fazt dev infra vps ssh prod-1
fazt dev infra vps exec prod-1 "fazt pulse status"
fazt dev infra vps destroy prod-1

# DNS Management
fazt dev infra dns zones
fazt dev infra dns add-zone example.com --provider cloudflare
fazt dev infra dns records example.com
fazt dev infra dns set example.com --type A --name blog --content 1.2.3.4
fazt dev infra dns set example.com --type MX --name @ --content mail.example.com --priority 10
fazt dev infra dns delete example.com --type A --name old-subdomain
fazt dev infra dns check blog.example.com

# Domain Management (optional)
fazt dev infra domain check my-domain.com
fazt dev infra domain search my-startup
fazt dev infra domain register my-domain.com --years 1
fazt dev infra domain list

# SSH Keys
fazt dev infra ssh-keys list hetzner
fazt dev infra ssh-keys add hetzner --name laptop --key "ssh-ed25519 ..."
```

## MCP Tools

When MCP is enabled (v0.12), infra operations are exposed as tools:

```
Tools:
  infra.vps.list          List all managed VPSes
  infra.vps.create        Create new VPS (requires confirmation)
  infra.vps.status        Get VPS status and health
  infra.vps.destroy       Destroy VPS (requires confirmation)

  infra.dns.zones         List managed DNS zones
  infra.dns.records       List records in a zone
  infra.dns.set           Create or update DNS record
  infra.dns.delete        Delete DNS record

  infra.domain.check      Check domain availability
  infra.domain.register   Register domain (requires confirmation)

Resources:
  fazt://infra/vps                  List of all VPSes
  fazt://infra/vps/{id}             VPS details
  fazt://infra/dns/{zone}           DNS records for zone
  fazt://infra/domains              List of owned domains
```

## Security Considerations

### Credential Storage

Infra credentials are powerful (can delete servers, rack up bills). They are:
- Stored encrypted in kernel config
- Never exposed to JS runtime
- Only accessible to kernel-level operations
- Logged in audit trail

```go
// Credentials stored encrypted
type InfraConfig struct {
    VPS map[string]struct {
        Token     EncryptedString `json:"token"`
        LastUsed  time.Time       `json:"last_used"`
    } `json:"vps"`
    DNS map[string]struct {
        Token     EncryptedString `json:"token"`
        ZoneIDs   map[string]string `json:"zone_ids"` // domain → zone ID
    } `json:"dns"`
}
```

### Destructive Operations

Operations that cost money or destroy resources require confirmation:

```go
// In kernel
func (i *Infra) VPSDestroy(id string) error {
    server, _ := i.VPSGet(id)

    // Require human confirmation
    confirmed, err := i.kernel.Halt(HaltRequest{
        Reason: fmt.Sprintf("Destroy server %s (%s)?", server.Name, server.IP),
        Data: map[string]any{
            "server": server,
            "action": "destroy",
        },
    })
    if !confirmed {
        return ErrUserCanceled
    }

    return i.providers.vps[server.Provider].Destroy(server.ProviderID)
}
```

### Rate Limiting

Infra APIs have rate limits. The kernel:
- Tracks rate limit headers
- Queues requests if approaching limits
- Backs off on 429 responses

### Audit Trail

All infra operations are logged:

```go
// Every infra call logged
i.events.Emit("infra.operation", InfraEvent{
    Type:      "vps.create",
    Provider:  "hetzner",
    Input:     sanitize(input),
    Output:    sanitize(output),
    Success:   err == nil,
    Timestamp: time.Now(),
})
```

## Error Handling

```javascript
try {
    await fazt.dev.infra.vps.create({ ... });
} catch (e) {
    if (e.code === 'INFRA_NOT_CONFIGURED') {
        // Provider not configured
        console.log('Run: fazt dev config infra.vps.hetzner --token xxx');
    }
    if (e.code === 'INFRA_PROVIDER_ERROR') {
        // Provider returned error
        console.log(e.providerCode, e.message);
    }
    if (e.code === 'INFRA_RATE_LIMITED') {
        // Hit provider rate limit
        console.log('Retry after:', e.retryAfter);
    }
    if (e.code === 'INFRA_QUOTA_EXCEEDED') {
        // Provider account quota (e.g., max servers)
        console.log(e.message);
    }
    if (e.code === 'INFRA_INSUFFICIENT_FUNDS') {
        // Provider account needs funding
        console.log('Add funds to your', e.provider, 'account');
    }
    if (e.code === 'INFRA_CANCELED') {
        // User canceled destructive operation
        console.log('Operation canceled by user');
    }
}
```

## Dashboard UI

Infra appears in the unified External Services dashboard:

```
┌─────────────────────────────────────────────────────────────────────┐
│  EXTERNAL SERVICES → INFRASTRUCTURE                                 │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  VPS PROVIDERS                                         [+ Add]      │
│  ┌──────────────┐ ┌──────────────┐                                 │
│  │ ☁️  Hetzner   │ │ 🌊 DigitalOcean │                               │
│  │              │ │                 │                               │
│  │ ✓ Connected  │ │ ○ Not setup    │                               │
│  │ 2 servers    │ │                 │                               │
│  │ €15.58/mo    │ │                 │                               │
│  └──────────────┘ └──────────────┘                                 │
│                                                                     │
│  DNS PROVIDERS                                         [+ Add]      │
│  ┌──────────────┐                                                  │
│  │ 🌐 Cloudflare │                                                  │
│  │              │                                                  │
│  │ ✓ Connected  │                                                  │
│  │ 3 zones      │                                                  │
│  │ Free tier    │                                                  │
│  └──────────────┘                                                  │
│                                                                     │
│  MANAGED SERVERS                                                    │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ Name         │ Provider │ IP             │ Status  │ Cost   │   │
│  ├─────────────────────────────────────────────────────────────┤   │
│  │ fazt-prod-1  │ Hetzner  │ 123.45.67.89   │ ● healthy │ €3.79 │   │
│  │ fazt-staging │ Hetzner  │ 123.45.67.90   │ ● healthy │ €3.79 │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  MANAGED ZONES                                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ Zone           │ Provider   │ Records │ Status              │   │
│  ├─────────────────────────────────────────────────────────────┤   │
│  │ example.com    │ Cloudflare │ 12      │ ✓ Active            │   │
│  │ another.com    │ Cloudflare │ 5       │ ✓ Active            │   │
│  │ test.dev       │ Hetzner    │ 3       │ ✓ Active            │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

## Implementation Notes

- Each provider is a Go package under `pkg/kernel/dev/infra/`
- Providers implement common interfaces (`VPSProvider`, `DNSProvider`)
- Cloud-init templates in `pkg/kernel/dev/infra/templates/`
- SSH key management via provider APIs (no local key storage)
- All operations go through kernel (no direct calls from JS)

## Dependencies

- `github.com/hetznercloud/hcloud-go/v2` - Hetzner Cloud API
- `github.com/cloudflare/cloudflare-go` - Cloudflare API
- `github.com/digitalocean/godo` - DigitalOcean API (planned)
- `github.com/vultr/govultr/v3` - Vultr API (planned)

## Binary Size Impact

| Dependency    | Size   | Notes            |
| ------------- | ------ | ---------------- |
| hcloud-go     | ~200KB | Hetzner SDK      |
| cloudflare-go | ~400KB | Cloudflare SDK   |
| godo          | ~300KB | DigitalOcean SDK |
| govultr       | ~200KB | Vultr SDK        |

Total: ~1.1MB additional binary size (with all providers).

Consider: Build tags to include only needed providers.
