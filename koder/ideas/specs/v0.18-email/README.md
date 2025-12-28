# v0.18 - Email

**Theme**: Inbound email processing.

## Summary

v0.18 adds an SMTP sink that receives emails at your domain and routes them
to apps. Each app can have its own email address prefix, and incoming emails
trigger serverless handlers.

## Goals

1. **Receive Emails**: Accept mail at `*@yourdomain.com`
2. **Route to Apps**: `support@` → support app, `orders@` → orders app
3. **Trigger Serverless**: Process emails with JS handlers
4. **Store & Query**: Access email history via `fazt.email`

## Key Capabilities

| Capability | Description |
|------------|-------------|
| SMTP sink | Receive on port 25 |
| Address routing | Local part → app mapping |
| Serverless trigger | `{ event: 'email', ... }` |
| Email storage | Query inbox via API |
| Attachments | Stored in `fazt.storage.s3` |

## Documents

- `sink.md` - SMTP server, routing, and API

## Dependencies

- v0.10 (Runtime): Serverless handlers for email processing
- v0.9 (Storage): Store emails and attachments

## DNS Setup

```
@ MX 10 mail.yourdomain.com.
mail A <fazt-server-ip>
```

Two DNS records. That's it for receiving.

## VPS Considerations

Many providers block port 25 by default (spam prevention).
Check provider documentation:
- DigitalOcean: Request unblock via support ticket
- Vultr: Unblock in control panel after 30 days
- Hetzner: Usually open
- AWS Lightsail: Request via support
