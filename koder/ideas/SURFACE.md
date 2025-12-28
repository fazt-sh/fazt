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
+ fazt storage migrate|backup|restore
+ fazt security root-pass
~ fazt server *                 # Deprecated, use fazt proc
```

### HTTP API

```
+ GET    /api/kernel/status
+ GET    /api/kernel/metrics
+ POST   /api/apps                      # Create app
+ GET    /api/apps/{uuid}               # Get by UUID
+ PUT    /api/apps/{uuid}               # Update app
+ DELETE /api/apps/{uuid}               # Delete app
~ /api/sites/* → /api/apps/*            # Renamed
```

### JS Runtime (fazt namespace)

```javascript
+ fazt.app.id                   // Current app UUID
+ fazt.app.name                 // Current app name
+ fazt.app.env                  // Environment variables
+ fazt.log.info|warn|error()    // Structured logging
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
```

### HTTP API

```
+ POST   /api/apps/{uuid}/invoke        # Trigger function
+ GET    /api/apps/{uuid}/logs
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

// app.json additions
{
  "cron": [
    { "schedule": "0 * * * *", "handler": "api/hourly.js" }
  ]
}
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
+ fazt services media resize|cache
+ fazt services search list|show|query|reindex
```

### HTTP API

```
+ POST   /_services/forms/{name}                  # Form submission endpoint
+ GET    /_services/media/{path}?w=&h=&thumb=     # On-the-fly processing
+ GET    /_services/qr?data=&size=                # QR generation
+ GET    /_services/search?q=                     # Search endpoint
+ POST   /_services/markdown/render               # Compile markdown
+ GET    /_services/comments/{target}             # Get comments
+ POST   /_services/comments/{target}             # Add comment
+ GET    /_s/{code}                          # Short URL redirect
+ POST   /_services/captcha                       # Create captcha
+ POST   /_services/captcha/{id}/verify           # Verify answer
```

### JS Runtime

```javascript
// Forms (dumb bucket)
+ fazt.services.forms.list(name, options?)
+ fazt.services.forms.get(name, id)
+ fazt.services.forms.delete(name, id)
+ fazt.services.forms.count(name)
+ fazt.services.forms.clear(name)

// Media (image processing)
+ fazt.services.media.resize(path, options)
+ fazt.services.media.thumbnail(path, size)
+ fazt.services.media.crop(path, options)
+ fazt.services.media.optimize(path, options?)
+ fazt.services.media.convert(path, format)
+ fazt.services.media.info(path)

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

// QR
+ fazt.services.qr.generate(data, options?)
+ fazt.services.qr.dataUrl(data, options?)
// options: { size }

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
├── ai
│   ├── complete(), stream(), embed()
├── git
│   ├── commit(), diff(), log(), rollback()
├── kernel
│   ├── deploy(), status(), apps, limits()
├── net
│   ├── vpn
│   │   ├── status(), peer(), authorize()
│   ├── domain
│   │   ├── current(), isPrimary()
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
├── services
    ├── forms
    │   ├── list(), get(), delete(), count(), clear()
    ├── media
    │   ├── resize(), thumbnail(), crop(), optimize(), convert(), info()
    ├── markdown
    │   ├── render(), renderFile(), extract()
    ├── search
    │   ├── index(), indexFiles(), query(), reindex(), dropIndex(), indexes()
    ├── qr
    │   ├── generate(), dataUrl()
    ├── comments
    │   ├── add(), list(), get(), update(), delete()
    │   ├── hide(), show(), approve(), count()
    ├── shorturl
    │   ├── create(), get(), update(), delete(), list()
    │   ├── stats(), clicks()
    ├── captcha
        ├── create(), verify()
```

### CLI Command Groups

```
fazt proc       # Process lifecycle
fazt fs         # Filesystem operations
fazt net        # Networking (routes, vpn, domains)
fazt storage    # Storage operations
fazt security   # Cryptographic operations
fazt identity   # Owner identity
fazt app        # App management
fazt marketplace # App sources
fazt mcp        # AI agent protocol
fazt mesh       # P2P synchronization
fazt limits     # Resource limits (presets, show, reset)
fazt realtime   # WebSocket pub/sub
fazt email      # SMTP sink
fazt worker     # Background jobs
fazt services   # Services (forms, media, markdown, search, qr)
```
