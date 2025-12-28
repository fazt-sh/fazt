# Email Sink

## Summary

SMTP server that receives emails at your domain, routes them to apps based
on the local part (prefix), and triggers serverless handlers. Emails are
stored and queryable via `fazt.email`.

## Architecture

```
Internet                    Fazt Kernel                    App
   │                            │                           │
   │── SMTP (port 25) ─────────►│                           │
   │                            │ Parse email               │
   │                            │ Route by local part       │
   │                            │ Store in DB               │
   │                            │────── trigger ───────────►│
   │                            │                           │ api/email.js
   │                            │◄───── response ───────────│
   │◄── 250 OK ─────────────────│                           │
```

## Routing

### Default: Local Part → App Slug

```
support@domain.com  →  app with slug "support"
orders@domain.com   →  app with slug "orders"
blog@domain.com     →  app with slug "blog"
```

If no app matches the local part, email goes to a fallback (configurable).

### Explicit Routing Table

For complex mappings, configure in kernel:

```bash
fazt email route add "helpdesk@" app_support_uuid
fazt email route add "sales-*@" app_crm_uuid      # Wildcard
fazt email route add "*@" app_catchall_uuid       # Catch-all
```

### Per-App Addresses

Apps can claim multiple addresses in `app.json`:

```json
{
  "email": {
    "addresses": ["support", "help", "contact"],
    "handler": "api/email.js"
  }
}
```

## Email Object

When an email arrives, it's parsed into:

```javascript
{
  id: 'email_abc123',
  from: {
    address: 'sender@example.com',
    name: 'John Doe'
  },
  to: [{
    address: 'support@yourdomain.com',
    name: ''
  }],
  cc: [],
  bcc: [],
  subject: 'Help with my order',
  text: 'Plain text body...',
  html: '<p>HTML body...</p>',
  attachments: [{
    id: 'att_xyz',
    filename: 'receipt.pdf',
    contentType: 'application/pdf',
    size: 102400,
    cid: 'QmXyz...'  // Stored in fazt.storage.s3
  }],
  headers: {
    'message-id': '<abc@example.com>',
    'date': 'Sat, 28 Dec 2024 10:00:00 +0000',
    ...
  },
  receivedAt: 1735383600000,
  processed: false,
  appUuid: 'app_support'
}
```

## Serverless Trigger

### Handler: api/email.js

```javascript
module.exports = async function(request) {
    // request.event === 'email'
    const email = request.email;

    console.log(`New email from ${email.from.address}`);
    console.log(`Subject: ${email.subject}`);

    // Process the email
    if (email.subject.includes('urgent')) {
        await notifySlack(email);
    }

    // Store in your own collection if needed
    await fazt.storage.ds.insert('tickets', {
        emailId: email.id,
        from: email.from.address,
        subject: email.subject,
        status: 'open'
    });

    // Mark as processed
    await fazt.email.markProcessed(email.id);

    return { success: true };
};
```

### Alternative: api/main.js with event routing

```javascript
module.exports = async function(request) {
    if (request.event === 'email') {
        return handleEmail(request.email);
    }
    if (request.event === 'http') {
        return handleHttp(request);
    }
};
```

## Server-Side API (fazt.email)

### Query Inbox

```javascript
// Get recent emails
const emails = await fazt.email.list({
    limit: 50,
    offset: 0,
    unprocessed: true
});

// Search emails
const results = await fazt.email.search({
    from: '*@example.com',
    subject: 'order',
    after: '2024-12-01'
});
```

### Get Single Email

```javascript
const email = await fazt.email.get('email_abc123');
```

### Get Attachment

```javascript
const attachment = await fazt.email.attachment('att_xyz');
// Returns: { data: Buffer, contentType: '...', filename: '...' }

// Or get URL (served via IPFS gateway)
const url = await fazt.email.attachmentUrl('att_xyz');
// "https://domain.com/ipfs/QmXyz..."
```

### Mark Processed

```javascript
await fazt.email.markProcessed('email_abc123');
```

### Delete Email

```javascript
await fazt.email.delete('email_abc123');
// Also deletes attachments from storage
```

### Reply (Future: v0.18.1)

```javascript
await fazt.email.reply('email_abc123', {
    text: 'Thanks for contacting us...',
    html: '<p>Thanks for contacting us...</p>'
});
// Requires SMTP relay configuration
```

## Storage

### Email Table

```sql
CREATE TABLE emails (
    id TEXT PRIMARY KEY,
    app_uuid TEXT NOT NULL,
    from_address TEXT,
    from_name TEXT,
    to_addresses TEXT,      -- JSON array
    cc_addresses TEXT,      -- JSON array
    subject TEXT,
    text_body TEXT,
    html_body TEXT,
    headers TEXT,           -- JSON object
    attachments TEXT,       -- JSON array of attachment refs
    received_at INTEGER,
    processed_at INTEGER,
    deleted_at INTEGER
);

CREATE INDEX idx_emails_app ON emails(app_uuid, received_at DESC);
CREATE INDEX idx_emails_from ON emails(app_uuid, from_address);
```

### Attachments

Stored via `fazt.storage.s3` (internal or external):
- Key: `emails/{email_id}/{attachment_id}`
- Content-addressable via CID for deduplication

## Limits

| Limit | Default | Description |
|-------|---------|-------------|
| `maxInboundPerHour` | 1000 | Emails per hour (all apps) |
| `maxInboundPerAppPerHour` | 100 | Per app |
| `maxMessageSizeMB` | 10 | Including attachments |
| `maxAttachmentSizeMB` | 25 | Single attachment |
| `maxAttachmentsPerEmail` | 20 | |
| `retentionDays` | 30 | Auto-delete after |

### Limit Behavior

- **At 80%**: Warning to owner
- **At 100%**: Reject with 452 (insufficient storage)
- **Oversized**: Reject with 552 (message too large)

## Spam Prevention

### SPF Check

Verify sender's SPF record. Configurable strictness:

```bash
fazt config set email.spf strict|soft|none
```

- `strict`: Reject SPF fail
- `soft`: Accept but flag
- `none`: No SPF checking

### Rate Limiting

Per-sender rate limits:

```bash
fazt config set email.rateLimit.perSender 100/hour
```

### Blocklist

```bash
fazt email block add "spammer@example.com"
fazt email block add "*@spamdomain.com"
```

## SMTP Implementation

### Port

- **Port 25**: Standard SMTP (requires VPS unblocking)
- Kernel only binds if email feature is enabled

### TLS

- STARTTLS supported and encouraged
- Uses same CertMagic certificates as HTTPS

### Commands Supported

```
EHLO, HELO
MAIL FROM
RCPT TO
DATA
QUIT
STARTTLS
RSET
NOOP
```

### Response Codes

| Code | Meaning |
|------|---------|
| 250 | OK |
| 354 | Start mail input |
| 421 | Service not available |
| 450 | Mailbox unavailable (temp) |
| 452 | Insufficient storage |
| 500 | Syntax error |
| 550 | Mailbox not found |
| 552 | Message too large |

## CLI

```bash
# Enable email sink
fazt email enable

# Check status
fazt email status

# List recent emails
fazt email list --app app_support --limit 10

# View email
fazt email show email_abc123

# Configure routing
fazt email route list
fazt email route add "prefix@" app_uuid

# Manage blocklist
fazt email block list
fazt email block add "spam@example.com"
```

## DNS Verification

Kernel provides DNS check:

```bash
fazt email dns-check

# Output:
# MX Record: ✓ mail.yourdomain.com
# A Record: ✓ 123.45.67.89
# Port 25: ✓ Open
# Ready to receive email!
```

## Example: Support Ticket System

```javascript
// api/email.js
const { z } = require('zod');

module.exports = async function(request) {
    const email = request.email;

    // Create ticket
    const ticket = await fazt.storage.ds.insert('tickets', {
        id: `ticket_${Date.now()}`,
        emailId: email.id,
        from: email.from.address,
        subject: email.subject,
        body: email.text,
        status: 'open',
        priority: email.subject.toLowerCase().includes('urgent')
            ? 'high' : 'normal',
        createdAt: Date.now()
    });

    // Notify via realtime
    await fazt.realtime.broadcast('tickets', {
        event: 'new_ticket',
        ticket
    });

    // Mark email processed
    await fazt.email.markProcessed(email.id);

    return { ticketId: ticket.id };
};
```
