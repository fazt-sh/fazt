# Notary Kernel

## Summary

The Notary Kernel provides cryptographic primitives at the kernel level. Apps
get security superpowers without importing crypto libraries.

## Primitives

### 1. Persona (Identity)

Every Fazt instance has a hardware-bound keypair:

```javascript
// Sign data
const signature = await fazt.security.sign(data);

// Verify signature
const valid = await fazt.security.verify(data, signature);

// Encrypt for self
const encrypted = await fazt.security.encrypt(data);

// Decrypt
const decrypted = await fazt.security.decrypt(encrypted);
```

### 2. Vaulting (Sealed Secrets)

Store secrets that survive only in memory:

```javascript
// Store secret (encrypted at rest)
await fazt.security.vault.store('api-key', 'sk-secret...');

// Retrieve (decrypted in vault)
const key = await fazt.security.vault.retrieve('api-key');

// App never sees raw key in logs/memory dumps
```

### 3. Attestation (Integrity)

Kernel can verify app code hasn't changed:

```go
func (k *Kernel) AttestApp(uuid string) bool {
    manifest := k.Apps.GetManifest(uuid)
    currentHash := k.FS.HashApp(uuid)
    return manifest.CodeHash == currentHash
}
```

### 4. Temporal (2FA)

Sensitive operations require time-based proof:

```javascript
// Force 2FA check
await fazt.security.totp.require();

// Only runs if owner enters TOTP code
await dangerousOperation();
```

## CLI

```bash
# Initialize kernel keypair
fazt security init

# Export public key (for sharing)
fazt security export-pubkey

# Sign a file
fazt security sign ./document.pdf

# Setup TOTP
fazt security totp init
# Shows QR code for authenticator app
```

## Use Cases

| App Type   | Primitives Used                        |
| ---------- | -------------------------------------- |
| Notes App  | Vaulting (encrypted notes)             |
| Wallet     | Persona (signing), TOTP (transactions) |
| Blog       | Attestation (verify content integrity) |
| AI Harness | TOTP (approve code changes)            |
