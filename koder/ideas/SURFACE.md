# Capability Surface Evolution

This document tracks the **API surface** as it evolves across versions.

Use this to understand what syscalls, CLI commands, and JS APIs are available
at each version checkpoint.

---

## Notation

- `CLI:` - Shell command
- `JS:` - Available in serverless runtime
- `API:` - HTTP endpoint
- `+` - Added in this version
- `~` - Modified in this version

---

## v0.7 - Cartridge PaaS (Current)

### CLI Commands

```
fazt server start [--port] [--env]
fazt server init
fazt server status
fazt server set-config
fazt server set-credentials
fazt server reset-admin
fazt service install [--domain] [--email] [--https]
fazt service logs
fazt client set-auth-token
fazt client deploy <dir> [--name]
fazt deploy <dir> [--name]              # Alias
fazt backup create
fazt backup restore
```

### HTTP API

```
POST   /api/login
GET    /api/user/me
GET    /api/sites
GET    /api/sites/{id}
GET    /api/sites/{id}/files
GET    /api/sites/{id}/files/{path}
POST   /api/deploy
GET    /api/analytics
GET    /api/redirects
POST   /api/redirects
DELETE /api/redirects/{id}
GET    /api/webhooks
POST   /api/webhooks
PUT    /api/webhooks/{id}
DELETE /api/webhooks/{id}
GET    /api/system/health
GET    /api/system/limits
GET    /api/system/cache
GET    /api/system/db
GET    /api/system/config
```

### JS Runtime

```javascript
// Available in main.js
request                         // Incoming request object
response                        // Response builder
console.log()                   // Logging
```

---

## v0.8 - Kernel

### CLI Commands

```
+ fazt proc start|stop|restart|upgrade
+ fazt fs mount|unmount|ls
+ fazt net route add|remove|list
+ fazt net logs|allow|limits            # Egress proxy
+ fazt storage migrate|backup|restore
+ fazt storage cleanup [--older-than]     # Hard delete soft-deleted data
+ fazt storage vacuum                      # Forensic cleanup (SQLite VACUUM)
+ fazt app export <app> [--output]         # Export app as cartridge
+ fazt app import <file> [--mode]          # Import cartridge
+ fazt app delete <app> [--purge]          # Soft/hard delete app
+ fazt app restore <app>                   # Restore soft-deleted app
+ fazt user export --user <id> [--app]     # Export user data (GDPR)
+ fazt user delete <id> [--purge]          # Delete user data
+ fazt user anonymize <id>                 # Anonymize user (keep data, scrub identity)
+ fazt security root-pass
+ fazt events list|watch|emit           # Event bus
+ fazt pulse status|ask|history|beat    # Cognitive observability
+ fazt dev list|test|logs|limits        # External service devices
+ fazt dev config <device> --token      # Configure devices
+ fazt dev infra vps list|create|destroy|status|ssh|exec  # VPS management
+ fazt dev infra dns zones|records|set|delete|check       # DNS management
+ fazt dev infra domain check|search|register|list        # Domain management
+ fazt dev infra ssh-keys list|add|remove                 # SSH key management
+ fazt beacon status|scan|set-name      # Local network discovery
+ fazt time status|sync|peers           # Local time consensus
+ fazt chirp send|listen|encode|decode  # Audio data transfer
+ fazt mnemonic encode|decode           # Human-channel exchange
~ fazt server *                         # Deprecated, use fazt proc
```

### HTTP API

```
+ GET    /api/kernel/status
+ GET    /api/kernel/metrics
+ POST   /api/apps                      # Create app
+ GET    /api/apps/{uuid}               # Get by UUID
+ PUT    /api/apps/{uuid}               # Update app
+ DELETE /api/apps/{uuid}               # Delete app
+ GET    /api/events                    # Query events
+ GET    /api/net/logs                  # Proxy logs
+ GET    /api/pulse/status              # Current health
+ GET    /api/pulse/history             # Past beats
+ POST   /api/pulse/ask                 # Natural language query
+ GET    /api/dev/{device}/status       # Device status
+ POST   /api/apps/{uuid}/export        # Export app as cartridge
+ POST   /api/apps/import               # Import cartridge
+ POST   /api/apps/{uuid}/restore       # Restore soft-deleted app
+ GET    /api/users/{id}/export         # Export user data (GDPR)
+ DELETE /api/users/{id}                # Delete user
+ POST   /api/users/{id}/anonymize      # Anonymize user
~ /api/sites/* → /api/apps/*            # Renamed
```

### JS Runtime (fazt namespace)

**Note**: Provenance (app_id, user_id tracking) is automatic and invisible.
All storage operations automatically include provenance context. Apps don't
need to do anything special - data ownership is handled by the kernel.

```javascript
+ fazt.app.id                   // Current app UUID
+ fazt.app.name                 // Current app name
+ fazt.app.env                  // Environment variables
+ fazt.log.info|warn|error()    // Structured logging

// Events (IPC)
+ fazt.events.emit(type, data, options?)
+ fazt.events.on(pattern, handler)
+ fazt.events.off(pattern, handler)
+ fazt.events.once(pattern, handler)
+ fazt.events.query(options)

// Network proxy
+ fazt.net.fetch(url, options?)
// options: auth, cache, retry, timeout, etc.
+ fazt.net.logs(options?)

// Pulse (cognitive observability)
+ fazt.pulse.status()              // Current health assessment
+ fazt.pulse.history(hours)        // Past beats with metrics
+ fazt.pulse.insights(hours?)      // LLM-generated insights
+ fazt.pulse.ask(question)         // Natural language query
+ fazt.pulse.trend(metric, hours)  // Time-series for charting

// Devices (external service abstraction)
+ fazt.dev.billing.customers.create|get|update|list
+ fazt.dev.billing.subscriptions.create|get|cancel|list
+ fazt.dev.billing.checkout.create
+ fazt.dev.billing.portal.create
+ fazt.dev.billing.invoices.list|get
+ fazt.dev.sms.send(options)       // Send SMS
+ fazt.dev.sms.status(id)          // Check delivery status
+ fazt.dev.email.send(options)     // Send email
+ fazt.dev.email.sendTemplate(options)
+ fazt.dev.oauth.authorize(options)   // Generate auth URL
+ fazt.dev.oauth.callback(options)    // Exchange code for tokens
+ fazt.dev.oauth.userinfo(options)    // Get user info
+ fazt.dev.oauth.refresh(options)     // Refresh token

// Infrastructure (cloud substrate)
+ fazt.dev.infra.vps.types(provider)       // List server types
+ fazt.dev.infra.vps.regions(provider)     // List regions
+ fazt.dev.infra.vps.create(options)       // Create VPS
+ fazt.dev.infra.vps.list()                // List managed VPSes
+ fazt.dev.infra.vps.get(id)               // Get VPS details
+ fazt.dev.infra.vps.status(id)            // VPS health status
+ fazt.dev.infra.vps.exec(id, command)     // Execute command
+ fazt.dev.infra.vps.reboot(id)            // Reboot VPS
+ fazt.dev.infra.vps.destroy(id)           // Destroy VPS (triggers halt)
+ fazt.dev.infra.vps.sshKeys(provider)     // List SSH keys
+ fazt.dev.infra.vps.addSshKey(provider, options)
+ fazt.dev.infra.vps.removeSshKey(provider, name)

+ fazt.dev.infra.dns.zones()               // List managed zones
+ fazt.dev.infra.dns.records(zone)         // List records
+ fazt.dev.infra.dns.set(zone, record)     // Create/update record
+ fazt.dev.infra.dns.delete(zone, record)  // Delete record
+ fazt.dev.infra.dns.setMany(zone, records)// Bulk operations
+ fazt.dev.infra.dns.checkPropagation(fqdn)// Check DNS propagation

+ fazt.dev.infra.domain.check(domain)      // Check availability
+ fazt.dev.infra.domain.search(query)      // Search available domains
+ fazt.dev.infra.domain.register(options)  // Register (triggers halt)
+ fazt.dev.infra.domain.list()             // List owned domains
+ fazt.dev.infra.domain.get(domain)        // Get domain details
+ fazt.dev.infra.domain.update(domain, options)

// Beacon (local discovery - usually automatic, explicit use optional)
+ fazt.beacon.discover(options?)      // Find nearby Fazt nodes
+ fazt.beacon.on('found', handler)    // Live discovery
+ fazt.beacon.on('lost', handler)     // Node disappeared

// Timekeeper (time consensus - usually automatic, explicit use optional)
+ fazt.time.now()                     // Consensus time (or system time)
+ fazt.time.status()                  // { local, consensus, drift, sources }
+ fazt.time.drift()                   // Milliseconds of drift
+ fazt.time.peers()                   // Contributing peers
+ fazt.time.sync()                    // Force sync

// Chirp (audio transfer - explicit use only)
+ fazt.chirp.encode(data, options?)   // Data to audio buffer
+ fazt.chirp.decode(audioBuffer)      // Audio buffer to data
+ fazt.chirp.send(data, options?)     // Play via speaker
+ fazt.chirp.listen(options?)         // Capture via microphone

// Mnemonic (human-channel exchange - explicit use only)
+ fazt.mnemonic.encode(data)          // Data to word sequence
+ fazt.mnemonic.decode(words)         // Words to data
+ fazt.mnemonic.validate(words)       // Check validity
```

---

## v0.9 - Storage

### CLI Commands

```
+ fazt storage upload <file>
+ fazt storage cid <path>
+ fazt storage gc                       # Garbage collect orphaned blobs
```

### HTTP API

```
+ GET    /ipfs/{cid}                    # IPFS gateway
+ GET    /ipfs/{cid}/{path}             # IPFS gateway with path
+ GET    /api/apps/{uuid}/storage/kv
+ POST   /api/apps/{uuid}/storage/kv
+ GET    /api/apps/{uuid}/storage/ds/{collection}
+ POST   /api/apps/{uuid}/storage/ds/{collection}
```

### JS Runtime

```javascript
+ fazt.storage.kv.get(key)
+ fazt.storage.kv.set(key, value, ttl?)
+ fazt.storage.kv.delete(key)
+ fazt.storage.kv.list(prefix?)

+ fazt.storage.ds.find(collection, query)
+ fazt.storage.ds.insert(collection, doc)
+ fazt.storage.ds.update(collection, query, update)
+ fazt.storage.ds.delete(collection, query)

+ fazt.storage.s3.put(key, data, mime?)
+ fazt.storage.s3.get(key)
+ fazt.storage.s3.delete(key)
+ fazt.storage.s3.list(prefix?)

+ fazt.fs.cid(path)                     // Get CID for file
+ fazt.fs.ipfsUrl(path)                 // Get /ipfs/ URL
```

---

## v0.10 - Runtime

### CLI Commands

```
+ fazt app run <uuid> [--cron]          # Manual trigger
+ fazt app logs <uuid>                  # View app logs
+ fazt sandbox exec '<code>'            # Execute in sandbox
+ fazt sandbox validate '<code>'        # Validate code
+ fazt wasm list                        # List loaded modules (admin)
+ fazt wasm stats <module>              # Module stats (admin)
+ fazt wasm cache clear                 # Clear module cache (admin)
```

### HTTP API

```
+ POST   /api/apps/{uuid}/invoke        # Trigger function
+ GET    /api/apps/{uuid}/logs
+ POST   /api/sandbox/exec              # Execute in sandbox
```

### JS Runtime

```javascript
+ require('./local-file.js')            // Local imports
+ require('lodash')                     // Stdlib: lodash
+ require('cheerio')                    // Stdlib: cheerio
+ require('uuid')                       // Stdlib: uuid
+ require('zod')                        // Stdlib: zod
+ require('marked')                     // Stdlib: marked

+ fazt.schedule(delayMs, state)         // Schedule future execution
+ fazt.cron.next()                      // Next scheduled run time

// Sandbox (safe code execution)
+ fazt.sandbox.exec(options)
// options: { code, input, context, timeout, memoryLimit, capabilities }
// Returns: { value, logs }
+ fazt.sandbox.validate(code)
// Returns: { valid, errors }

// app.json additions
{
  "cron": [
    { "schedule": "0 * * * *", "handler": "api/hourly.js" }
  ]
}
```

### Kernel Primitives (Internal, not exposed to JS)

```go
// WASM Runtime - for kernel services only
+ wasm.Load(ctx, bytes, config)         // Load WASM module
+ wasm.NewCache(config)                 // Module cache
+ module.Call(ctx, fn, args)            // Invoke function
+ module.WriteBytes(data)               // Write to WASM memory
+ module.ReadBytes(ptr, len)            // Read from WASM memory
+ module.SetFuelLimit(n)                // CPU limit
+ module.Export(name, fn)               // Host function

// Config options
// - MemoryLimit: 64MB default, 256MB max
// - FuelLimit: instruction budget (~1B = 1 second)
// - Embedded modules: libimage, libpdf, libxlsx
```

---

## v0.11 - Distribution

### CLI Commands

```
+ fazt app install <name> [--source <url>]
+ fazt app update <name>
+ fazt app remove <name>
+ fazt app list [--source personal|marketplace]
+ fazt marketplace add <git-url>
+ fazt marketplace remove <name>
+ fazt marketplace sync
```

### HTTP API

```
+ GET    /api/marketplace
+ POST   /api/marketplace/install
+ GET    /api/apps/{uuid}/manifest
```

### JS Runtime

```javascript
+ fazt.app.source                       // 'personal' | 'marketplace'
+ fazt.app.version                      // Installed version
+ fazt.app.manifest                     // Parsed app.json

// app.json additions
{
  "name": "my-app",
  "version": "1.0.0",
  "permissions": ["storage:kv", "net:fetch"],
  "env": ["API_KEY"]
}
```

---

## v0.12 - Agentic

### CLI Commands

```
+ fazt mcp start [--port]               # Start MCP server
+ fazt mcp status
```

### HTTP API

```
+ POST   /mcp/tools/list
+ POST   /mcp/tools/call
+ POST   /api/apps/{uuid}/git/commit
+ GET    /api/apps/{uuid}/git/log
+ POST   /api/apps/{uuid}/git/rollback
```

### JS Runtime

```javascript
+ fazt.ai.complete(prompt, options)
+ fazt.ai.stream(prompt, onChunk)
+ fazt.ai.embed(text)

// Options: { model, temperature, maxTokens, provider }
// Providers: 'openai', 'anthropic', 'gemini', 'ollama'

+ fazt.git.commit(message)
+ fazt.git.diff()
+ fazt.git.log(limit?)
+ fazt.git.rollback(commitId)

+ fazt.kernel.deploy(appId, files)
+ fazt.kernel.status()
+ fazt.kernel.apps.list()
```

---

## v0.13 - Network

### CLI Commands

```
+ fazt net vpn init
+ fazt net vpn add-peer [--name]
+ fazt net vpn remove-peer <id>
+ fazt net vpn qr <peerId>
+ fazt net domain map <domain> <appUuid>
+ fazt net domain unmap <domain>
+ fazt net domain list
```

### HTTP API

```
+ GET    /api/vpn/peers
+ POST   /api/vpn/peers
+ DELETE /api/vpn/peers/{id}
+ GET    /api/vpn/peers/{id}/config      # WireGuard config
+ GET    /api/domains
+ POST   /api/domains
+ DELETE /api/domains/{id}
```

### JS Runtime

```javascript
+ fazt.net.vpn.status()                 // Is request via VPN?
+ fazt.net.vpn.peer()                   // Connected peer info
+ fazt.net.vpn.authorize()              // Elevate trust (TOTP)

+ fazt.net.domain.current()             // Current request domain
+ fazt.net.domain.isPrimary()           // Is primary domain?

// app.json additions
{
  "visibility": "vpn-only"              // Only accessible via VPN
}
```

---

## v0.14 - Security

### CLI Commands

```
+ fazt security init                    # Generate kernel keypair
+ fazt security export-pubkey
+ fazt security sign <file>
+ fazt security verify <file> <sig>
+ fazt security totp init
+ fazt security totp verify <code>
```

### HTTP API

```
+ GET    /api/security/pubkey
+ POST   /api/security/sign
+ POST   /api/security/verify
+ POST   /api/security/totp/init
+ POST   /api/security/totp/verify
```

### JS Runtime

```javascript
+ fazt.security.sign(data)
+ fazt.security.verify(data, signature, pubkey?)
+ fazt.security.encrypt(data, pubkey?)
+ fazt.security.decrypt(data)

+ fazt.security.vault.store(key, secret)
+ fazt.security.vault.retrieve(key)
+ fazt.security.vault.delete(key)

+ fazt.halt(reason, data)               // Pause for human approval
+ fazt.security.totp.require()          // Force 2FA check

// Automatic RLS: All storage queries filtered by app_id
```

---

## v0.15 - Identity

### CLI Commands

```
+ fazt identity init                    # Setup owner persona
+ fazt identity export                  # Export identity
+ fazt identity import <file>           # Import identity
```

### HTTP API

```
+ GET    /api/identity/persona
+ POST   /api/identity/assertion        # Generate signed assertion
+ GET    /.well-known/openid-configuration  # OIDC discovery
+ GET    /oauth/authorize               # OIDC authorization
+ POST   /oauth/token                   # OIDC token
```

### JS Runtime

```javascript
+ fazt.security.getPersona()            // Owner profile
+ fazt.security.isOwner()               // Is owner making request?
+ fazt.security.signAssertion()         // Generate identity proof
+ fazt.security.requireAuth()           // Force login

// Automatic SSO: All subdomains inherit auth state
```

---

## v0.16 - Mesh

### CLI Commands

```
+ fazt mesh init                        # Initialize mesh identity
+ fazt mesh join <peer-url>             # Join mesh network
+ fazt mesh leave
+ fazt mesh status
+ fazt mesh sync                        # Force sync
```

### HTTP API

```
+ GET    /api/mesh/peers
+ POST   /api/mesh/peers
+ DELETE /api/mesh/peers/{id}
+ POST   /api/mesh/sync
+ GET    /api/mesh/status
```

### JS Runtime

```javascript
+ fazt.mesh.peers()                     // List mesh peers
+ fazt.mesh.broadcast(data)             // Broadcast to all peers
+ fazt.mesh.sync()                      // Force sync

+ fazt.protocols.activitypub.actor()    // ActivityPub actor
+ fazt.protocols.activitypub.inbox()    // Inbox messages
+ fazt.protocols.activitypub.post(content)

+ fazt.protocols.nostr.pubkey()         // Nostr public key
+ fazt.protocols.nostr.sign(event)      // Sign Nostr event
+ fazt.protocols.nostr.publish(event)   // Publish to relays
```

---

## v0.17 - Realtime

### CLI Commands

```
+ fazt realtime status                 # Show connection stats
+ fazt realtime channels [--app]       # List active channels
+ fazt realtime kick <clientId>        # Disconnect client
```

### HTTP API

```
+ GET    wss://app.domain/_ws          # WebSocket endpoint
+ GET    /api/apps/{uuid}/realtime/channels
+ GET    /api/apps/{uuid}/realtime/stats
+ POST   /api/apps/{uuid}/realtime/broadcast
```

### JS Runtime

```javascript
+ fazt.realtime.broadcast(channel, data)    // Send to all subscribers
+ fazt.realtime.subscribers(channel)        // List client IDs
+ fazt.realtime.count(channel)              // Subscriber count
+ fazt.realtime.kick(clientId)              // Disconnect client

// Channel types:
// - public:{name}    - Anyone can subscribe
// - private:{name}   - Requires auth callback
// - presence:{name}  - Tracks member joins/leaves

// Client protocol (JSON over WebSocket):
// { "type": "subscribe", "channel": "public:chat" }
// { "type": "unsubscribe", "channel": "public:chat" }
// { "type": "message", "channel": "public:chat", "data": {...} }
```

---

## v0.18 - Email

### CLI Commands

```
+ fazt email status                    # SMTP sink status
+ fazt email list [--app]              # List received emails
+ fazt email show <id>                 # View email details
+ fazt email purge --older-than 30d    # Clean old emails
```

### HTTP API

```
+ GET    /api/apps/{uuid}/email
+ GET    /api/apps/{uuid}/email/{id}
+ GET    /api/apps/{uuid}/email/{id}/attachment/{idx}
+ POST   /api/apps/{uuid}/email/{id}/processed
+ DELETE /api/apps/{uuid}/email/{id}
```

### JS Runtime

```javascript
+ fazt.email.list(options)             // { limit, offset, unprocessed }
+ fazt.email.get(id)                   // Full email object
+ fazt.email.attachment(id, index)     // Get attachment buffer
+ fazt.email.markProcessed(id)         // Mark as handled
+ fazt.email.delete(id)                // Remove email

// Email object:
// {
//   id, from, to, subject, textBody, htmlBody,
//   attachments: [{ filename, contentType, size }],
//   receivedAt, processed
// }

// Routing: local part maps to app slug
// support@domain.com → app with slug "support"
```

---

## v0.19 - Workers

### CLI Commands

```
+ fazt worker list [--app] [--status]  # List jobs
+ fazt worker show <jobId>             # Job details
+ fazt worker cancel <jobId>           # Cancel running job
+ fazt worker dead-letter list         # Failed jobs
+ fazt worker dead-letter retry <id>   # Retry failed job
+ fazt worker purge --older-than 7d    # Clean old jobs
```

### HTTP API

```
+ GET    /api/apps/{uuid}/workers
+ GET    /api/apps/{uuid}/workers/{id}
+ POST   /api/apps/{uuid}/workers/{id}/cancel
+ GET    /api/apps/{uuid}/workers/dead-letter
+ POST   /api/apps/{uuid}/workers/dead-letter/{id}/retry
```

### JS Runtime

```javascript
+ fazt.worker.spawn(handler, options)  // Create background job
+ fazt.worker.get(id)                  // Job status
+ fazt.worker.list(options)            // { status, limit, order }
+ fazt.worker.cancel(id)               // Cancel job
+ fazt.worker.wait(id, options)        // Poll until done

+ fazt.worker.deadLetter.list()        // Failed jobs
+ fazt.worker.deadLetter.get(id)       // Failed job details
+ fazt.worker.deadLetter.retry(id)     // Retry failed job
+ fazt.worker.deadLetter.delete(id)    // Acknowledge failure

// Spawn options:
// {
//   data: {...},              // Passed to handler
//   timeout: '5m',            // Max runtime (max: 30m)
//   retry: 3,                 // Retry attempts
//   retryDelay: '1m',         // Delay between retries
//   retryBackoff: 'exponential',
//   priority: 'normal',       // 'low' | 'normal' | 'high'
//   delay: '10s',             // Delay before first run
//   uniqueKey: 'job-123'      // Prevent duplicates
// }

// In worker handler:
// job.id, job.data, job.attempt
// job.progress(percent)       // Report 0-100
// job.log(message)            // Add log entry
// job.checkpoint(state)       // Save state for resume
```

---

## v0.20 - Services

### CLI Commands

```
+ fazt services forms list|show|export|purge
+ fazt services image resize|cache           # Renamed from media
+ fazt services pdf render|info|list|purge
+ fazt services search list|show|query|reindex
+ fazt hooks events|replay|stats          # Inbound
+ fazt hooks list|register|delete         # Outbound
+ fazt hooks deliveries|retry-delivery    # Delivery management
```

### HTTP API

```
+ POST   /_services/forms/{name}                  # Form submission endpoint
+ GET    /_services/image/{path}?w=&h=&thumb=     # On-the-fly processing
+ POST   /_services/pdf/render                    # HTML to PDF
+ GET    /_services/pdf/render?file={path}        # File to PDF
+ GET    /_services/qr?data=&size=                # QR generation
+ GET    /_services/barcode?data=&format=        # Barcode generation
+ GET    /_services/search?q=                     # Search endpoint
+ POST   /_services/markdown/render               # Compile markdown
+ GET    /_services/comments/{target}             # Get comments
+ POST   /_services/comments/{target}             # Add comment
+ GET    /_s/{code}                               # Short URL redirect
+ POST   /_services/captcha                       # Create captcha
+ POST   /_services/captcha/{id}/verify           # Verify answer
+ POST   /_hooks/{provider}                       # Inbound webhook receiver
+ GET    /api/hooks/events                        # List inbound events
+ POST   /api/hooks/events/{id}/replay            # Replay event
+ GET    /api/hooks                               # List outbound hooks
+ POST   /api/hooks                               # Register outbound hook
+ DELETE /api/hooks/{id}                          # Delete hook
+ GET    /api/hooks/deliveries                    # List deliveries
```

### JS Runtime

```javascript
// === SERVICES (stateful, I/O, lifecycle) ===

// Forms (dumb bucket)
+ fazt.services.forms.list(name, options?)
+ fazt.services.forms.get(name, id)
+ fazt.services.forms.delete(name, id)
+ fazt.services.forms.count(name)
+ fazt.services.forms.clear(name)

// Image (renamed from media)
+ fazt.services.image.resize(path, options)
+ fazt.services.image.thumbnail(path, size)
+ fazt.services.image.crop(path, options)
+ fazt.services.image.optimize(path, options?)
+ fazt.services.image.convert(path, format)
+ fazt.services.image.info(path)
+ fazt.services.image.blurhash(path, options?)
+ fazt.services.image.blurhashDataUrl(hash, options?)

// PDF (HTML/CSS to PDF via WASM)
+ fazt.services.pdf.fromHtml(html, options?)
+ fazt.services.pdf.fromFile(path, options?)
+ fazt.services.pdf.fromUrl(url, options?)
+ fazt.services.pdf.merge(paths, options?)
+ fazt.services.pdf.info(path)
+ fazt.services.pdf.delete(path)
// options: { pageSize, margin, orientation, output }
// output: 'path' (default) | 'bytes'

// Markdown
+ fazt.services.markdown.render(content, options?)
+ fazt.services.markdown.renderFile(path, options?)
+ fazt.services.markdown.extract(content)
// options: { css, highlight, toc, shortcodes, template }
// Shortcodes: {{form}}, {{youtube}}, {{qr}}, {{include}}, {{toc}}

// Search
+ fazt.services.search.index(collection, options)
+ fazt.services.search.indexFiles(glob, options?)
+ fazt.services.search.query(term, options?)
+ fazt.services.search.reindex(collection)
+ fazt.services.search.dropIndex(collection)
+ fazt.services.search.indexes()

// QR & Barcode (consolidated)
+ fazt.services.qr.generate(data, options?)
+ fazt.services.qr.dataUrl(data, options?)
+ fazt.services.qr.svg(data, options?)
+ fazt.services.qr.barcode(data, options)
+ fazt.services.qr.barcodeDataUrl(data, options)
// options: { size, format }

// Comments
+ fazt.services.comments.add(target, options)
+ fazt.services.comments.list(target, options?)
+ fazt.services.comments.get(id)
+ fazt.services.comments.update(id, options)
+ fazt.services.comments.delete(id)
+ fazt.services.comments.hide(id)
+ fazt.services.comments.show(id)
+ fazt.services.comments.approve(id)
+ fazt.services.comments.count(target)
// options: { body, author, authorName, meta, parentId }

// Short URL
+ fazt.services.shorturl.create(target, options?)
+ fazt.services.shorturl.get(code)
+ fazt.services.shorturl.update(code, options)
+ fazt.services.shorturl.delete(code)
+ fazt.services.shorturl.list(options?)
+ fazt.services.shorturl.stats(code)
+ fazt.services.shorturl.clicks(code, options?)
// options: { code, expiresIn, expiresAt, maxClicks }

// Captcha
+ fazt.services.captcha.create(options?)
+ fazt.services.captcha.verify(id, answer)
// options: { type } - 'math' | 'text'

// Hooks (bidirectional webhooks)
+ fazt.services.hooks.events(options?)       // Query inbound events
+ fazt.services.hooks.event(id)              // Get specific event
+ fazt.services.hooks.replay(id)             // Replay event
+ fazt.services.hooks.replayFailed(provider?)
+ fazt.services.hooks.stats(provider)        // Inbound stats
+ fazt.services.hooks.register(options)      // Register outbound hook
+ fazt.services.hooks.list()                 // List outbound hooks
+ fazt.services.hooks.update(id, options)
+ fazt.services.hooks.delete(id)
+ fazt.services.hooks.emit(type, data)       // Trigger outbound
+ fazt.services.hooks.deliveries(options?)   // Query deliveries
+ fazt.services.hooks.retryDelivery(id)

// Rate Limiting (in fazt.limits namespace)
+ fazt.limits.rate.status(key)
+ fazt.limits.rate.check(key, options)
+ fazt.limits.rate.consume(key, options)

// === LIB (pure functions, no state, no side effects) ===

// Money (decimal arithmetic)
+ fazt.lib.money.add(...amounts)
+ fazt.lib.money.subtract(a, b)
+ fazt.lib.money.multiply(amount, factor)
+ fazt.lib.money.divide(amount, divisor, options?)
+ fazt.lib.money.percent(amount, percent)
+ fazt.lib.money.addPercent(amount, percent)
+ fazt.lib.money.subtractPercent(amount, percent)
+ fazt.lib.money.format(cents, currency, options?)
+ fazt.lib.money.parse(string, currency)
+ fazt.lib.money.compare(a, b)
+ fazt.lib.money.min(...amounts)
+ fazt.lib.money.max(...amounts)
+ fazt.lib.money.split(amount, parts)
+ fazt.lib.money.allocate(amount, ratios)
+ fazt.lib.money.currency(code)
+ fazt.lib.money.currencies()

// Humanize (human-readable formatting)
+ fazt.lib.humanize.bytes(bytes, options?)
+ fazt.lib.humanize.time(timestamp, options?)
+ fazt.lib.humanize.duration(ms, options?)
+ fazt.lib.humanize.number(n, options?)
+ fazt.lib.humanize.compact(n, options?)
+ fazt.lib.humanize.ordinal(n)
+ fazt.lib.humanize.plural(count, singular, plural?, options?)
+ fazt.lib.humanize.truncate(text, length, options?)
+ fazt.lib.humanize.list(items, options?)

// Timezone (IANA timezone handling)
+ fazt.lib.timezone.now(tz)
+ fazt.lib.timezone.convert(time, fromTz, toTz, options?)
+ fazt.lib.timezone.parse(time, tz, options?)
+ fazt.lib.timezone.format(timestamp, tz, options?)
+ fazt.lib.timezone.isDST(tz, time?)
+ fazt.lib.timezone.transitions(tz, year)
+ fazt.lib.timezone.info(tz)
+ fazt.lib.timezone.list(options?)
+ fazt.lib.timezone.search(query)
+ fazt.lib.timezone.offset(fromTz, toTz, time?)
+ fazt.lib.timezone.offsetFromUTC(tz, time?)
+ fazt.lib.timezone.next(time, tz)
+ fazt.lib.timezone.scheduleDaily(time, tz)
+ fazt.lib.timezone.isWithin(timestamp, tz, range)

// Sanitize (HTML/text sanitization)
+ fazt.lib.sanitize.html(input, options?)
// options: { policy: 'strict'|'basic'|'rich', allow: [], allowAttrs: {} }
+ fazt.lib.sanitize.text(input)
+ fazt.lib.sanitize.markdown(input, options?)
+ fazt.lib.sanitize.url(input, options?)

// Password (secure hashing)
+ fazt.lib.password.hash(plaintext, options?)
+ fazt.lib.password.verify(plaintext, hash)
+ fazt.lib.password.needsRehash(hash)
+ fazt.lib.password.config()

// Geo (geographic primitives)
+ fazt.lib.geo.distance(lat1, lon1, lat2, lon2, options?)
+ fazt.lib.geo.fromIP(ip)
+ fazt.lib.geo.countryFromIP(ip)
+ fazt.lib.geo.contains(polygon, point)
+ fazt.lib.geo.inBounds(bbox, lat, lon)
+ fazt.lib.geo.bounds(points)
+ fazt.lib.geo.timezone(lat, lon)
+ fazt.lib.geo.countryAt(lat, lon)
+ fazt.lib.geo.nearby(items, center, options)

// Mime (mimetype detection - extracted from media)
+ fazt.lib.mime.detect(path)
+ fazt.lib.mime.fromBytes(buffer)
+ fazt.lib.mime.toExt(mime)
+ fazt.lib.mime.fromExt(ext)
+ fazt.lib.mime.isImage(path)
+ fazt.lib.mime.is(path, mime)
```

---

## Summary: Full Surface at v0.20

### fazt.* Namespace Tree

```
fazt
├── app
│   ├── id, name, env, source, version, manifest
├── log
│   ├── info(), warn(), error()
├── storage
│   ├── kv
│   │   ├── get(), set(), delete(), list()
│   ├── ds
│   │   ├── find(), insert(), update(), delete()
│   ├── s3
│   │   ├── put(), get(), delete(), list()
├── fs
│   ├── cid(), ipfsUrl()
├── schedule()
├── cron
│   ├── next()
├── sandbox
│   ├── exec(), validate()
├── ai
│   ├── complete(), stream(), embed()
├── git
│   ├── commit(), diff(), log(), rollback()
├── kernel
│   ├── deploy(), status(), apps, limits()
├── net
│   ├── fetch(), logs()
│   ├── vpn
│   │   ├── status(), peer(), authorize()
│   ├── domain
│   │   ├── current(), isPrimary()
├── events
│   ├── emit(), on(), off(), once(), query()
├── pulse
│   ├── status(), history(), insights(), ask(), trend()
├── beacon
│   ├── discover(), on(), announce()
├── time
│   ├── now(), status(), drift(), peers(), sync()
├── chirp
│   ├── encode(), decode(), send(), listen()
├── mnemonic
│   ├── encode(), decode(), validate()
├── dev
│   ├── billing
│   │   ├── customers, subscriptions, checkout, portal, invoices
│   ├── sms
│   │   ├── send(), status()
│   ├── email
│   │   ├── send(), sendTemplate()
│   ├── oauth
│   │   ├── authorize(), callback(), userinfo(), refresh()
│   ├── infra
│       ├── vps
│       │   ├── types(), regions(), create(), list(), get(), status()
│       │   ├── exec(), reboot(), destroy(), sshKeys(), addSshKey()
│       ├── dns
│       │   ├── zones(), records(), set(), delete(), setMany()
│       │   ├── checkPropagation()
│       ├── domain
│           ├── check(), search(), register(), list(), get(), update()
├── security
│   ├── sign(), verify(), encrypt(), decrypt()
│   ├── vault
│   │   ├── store(), retrieve(), delete()
│   ├── totp
│   │   ├── require()
│   ├── getPersona(), isOwner(), signAssertion(), requireAuth()
├── halt()
├── mesh
│   ├── peers(), broadcast(), sync()
├── protocols
│   ├── activitypub
│   │   ├── actor(), inbox(), post()
│   ├── nostr
│       ├── pubkey(), sign(), publish()
├── realtime
│   ├── broadcast(), subscribers(), count(), kick()
├── email
│   ├── list(), get(), attachment(), markProcessed(), delete()
├── worker
│   ├── spawn(), get(), list(), cancel(), wait()
│   ├── deadLetter
│       ├── list(), get(), retry(), delete()
├── limits
│   ├── rate
│       ├── status(), check(), consume()
├── services                              # Stateful services
│   ├── forms
│   │   ├── list(), get(), delete(), count(), clear()
│   ├── image                             # Renamed from media
│   │   ├── resize(), thumbnail(), crop(), optimize(), convert(), info()
│   │   ├── blurhash(), blurhashDataUrl()
│   ├── pdf
│   │   ├── fromHtml(), fromFile(), fromUrl(), merge(), info(), delete()
│   ├── markdown
│   │   ├── render(), renderFile(), extract()
│   ├── search
│   │   ├── index(), indexFiles(), query(), reindex(), dropIndex(), indexes()
│   ├── qr                                # Consolidated (includes barcode)
│   │   ├── generate(), dataUrl(), svg()
│   │   ├── barcode(), barcodeDataUrl()
│   ├── comments
│   │   ├── add(), list(), get(), update(), delete()
│   │   ├── hide(), show(), approve(), count()
│   ├── shorturl
│   │   ├── create(), get(), update(), delete(), list()
│   │   ├── stats(), clicks()
│   ├── captcha
│   │   ├── create(), verify()
│   └── hooks
│       ├── events(), event(), replay(), replayFailed(), stats()
│       ├── register(), list(), update(), delete(), emit()
│       ├── deliveries(), retryDelivery()
└── lib                                   # Pure functions (no state)
    ├── money
    │   ├── add(), subtract(), multiply(), divide()
    │   ├── percent(), addPercent(), subtractPercent()
    │   ├── format(), parse(), compare(), min(), max()
    │   ├── split(), allocate(), currency(), currencies()
    ├── humanize
    │   ├── bytes(), time(), duration(), number(), compact()
    │   ├── ordinal(), plural(), truncate(), list()
    ├── timezone
    │   ├── now(), convert(), parse(), format()
    │   ├── isDST(), transitions(), info(), list(), search()
    │   ├── offset(), offsetFromUTC()
    │   ├── next(), scheduleDaily(), isWithin()
    ├── sanitize
    │   ├── html(), text(), markdown(), url()
    ├── password
    │   ├── hash(), verify(), needsRehash(), config()
    ├── geo
    │   ├── distance(), fromIP(), countryFromIP()
    │   ├── contains(), inBounds(), bounds()
    │   ├── timezone(), countryAt(), nearby()
    └── mime
        ├── detect(), fromBytes()
        ├── toExt(), fromExt()
        ├── isImage(), is()
```

### CLI Command Groups

```
fazt proc       # Process lifecycle
fazt fs         # Filesystem operations
fazt net        # Networking (routes, vpn, domains, egress proxy)
fazt events     # Internal event bus
fazt storage    # Storage operations
fazt security   # Cryptographic operations
fazt identity   # Owner identity
fazt beacon     # Local network discovery (mDNS)
fazt time       # Local time consensus
fazt chirp      # Audio data transfer
fazt mnemonic   # Human-channel data exchange
fazt app        # App management
fazt sandbox    # Safe code execution
fazt marketplace # App sources
fazt mcp        # AI agent protocol
fazt mesh       # P2P synchronization
fazt limits     # Resource limits (presets, show, reset)
fazt realtime   # WebSocket pub/sub
fazt email      # SMTP sink
fazt worker     # Background jobs
fazt services   # Services (forms, image, pdf, markdown, search, qr)
fazt pulse      # Cognitive observability (health, ask, insights)
fazt dev        # External service devices (billing, sms, email, oauth, infra)
fazt hooks      # Bidirectional webhooks (inbound, outbound)
```
