# Protocol Support

## Summary

Fazt can participate in decentralized protocols, making your instance a
citizen of the wider federated web.

## ActivityPub

### What It Is

The protocol behind Mastodon, Pixelfed, and other fediverse apps. Implement
ActivityPub and your Fazt instance can:
- Follow/be followed by Mastodon users
- Post content visible across the fediverse
- Receive replies and interactions

### Implementation

```javascript
// Get your ActivityPub actor
const actor = await fazt.protocols.activitypub.actor();
// @you@yourdomain.com

// Post to fediverse
await fazt.protocols.activitypub.post({
    content: 'Hello from my personal cloud!',
    visibility: 'public'
});

// Read inbox
const messages = await fazt.protocols.activitypub.inbox();
```

### Endpoints

```
GET  /.well-known/webfinger
GET  /activitypub/actor
POST /activitypub/inbox
GET  /activitypub/outbox
```

## Nostr

### What It Is

A decentralized social protocol using cryptographic keys. Your Persona
keypair IS your Nostr identity.

### Implementation

```javascript
// Get Nostr public key (derived from Persona)
const pubkey = await fazt.protocols.nostr.pubkey();

// Sign and publish event
const event = {
    kind: 1,  // Text note
    content: 'Hello Nostr!'
};
await fazt.protocols.nostr.publish(event);

// Read from relays
const notes = await fazt.protocols.nostr.fetch({
    authors: [pubkey],
    kinds: [1],
    limit: 10
});
```

### Relay Configuration

```json
{
  "nostr": {
    "relays": [
      "wss://relay.damus.io",
      "wss://nos.lol"
    ]
  }
}
```

## Benefits

| Protocol | Benefit |
|----------|---------|
| ActivityPub | Federate with Mastodon ecosystem |
| Nostr | Censorship-resistant identity |

## Use Case: Personal Social Presence

Build a social app on Fazt that:
1. Posts to both ActivityPub and Nostr
2. Receives interactions from both networks
3. Stores everything in your `data.db`
4. You own your social graph
