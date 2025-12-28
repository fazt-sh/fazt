# Persona

## Summary

The owner's identity is a kernel-managed keypair, not a password in a table.
This "Persona" can sign assertions, prove ownership, and authenticate across
all apps without configuration.

## The Keypair

```bash
# Generated during fazt init
fazt identity init

# Stored encrypted in data.db
# Decrypted only in kernel memory
```

## JS Runtime

```javascript
// Get owner profile
const persona = await fazt.security.getPersona();
// { username, email, publicKey }

// Check if current request is from owner
if (await fazt.security.isOwner()) {
    // Show admin features
}

// Generate signed assertion (for third-party verification)
const assertion = await fazt.security.signAssertion();
// { persona, timestamp, signature }
```

## Use Cases

### App Authentication

```javascript
// api/main.js
module.exports = async function(request) {
    if (!await fazt.security.isOwner()) {
        return { status: 401, body: 'Owner only' };
    }
    // Proceed with admin action
};
```

### External Verification

```javascript
// Third party can verify you own the server
const assertion = await fazt.security.signAssertion();
// Send to external service
// They verify signature against your public key
```

## Storage

```sql
CREATE TABLE identity (
    key TEXT PRIMARY KEY,
    value BLOB  -- Encrypted keypair, profile data
);
```

## Future: Identity Export

```bash
# Export identity for backup/migration
fazt identity export --output identity.enc

# Import on new server
fazt identity import identity.enc
```
