# Password Library

## Summary

Secure password hashing with Argon2id. The modern, memory-hard algorithm
recommended by OWASP. Handles salting, work factors, and timing-safe comparison.

## The Problem

```javascript
// Common mistakes that lead to breaches:
const hash = md5(password);           // Weak algorithm
const hash = sha256(password);        // No salt, too fast
const hash = sha256(password + salt); // Still too fast

// Timing attacks:
if (storedHash === inputHash) { ... } // Vulnerable to timing analysis
```

## The Solution

```javascript
// Secure by default
const hash = await fazt.lib.password.hash('user-password');
// "$argon2id$v=19$m=65536,t=3,p=4$..."

const valid = await fazt.lib.password.verify('user-password', hash);
// true/false (timing-safe)
```

## Usage

### Hash Password

```javascript
// Default settings (OWASP recommended)
const hash = await fazt.lib.password.hash('mypassword');
// "$argon2id$v=19$m=65536,t=3,p=4$randomsalt$hashvalue"

// The hash includes:
// - Algorithm identifier (argon2id)
// - Version (19)
// - Memory cost (64MB)
// - Time cost (3 iterations)
// - Parallelism (4 threads)
// - Random salt (16 bytes)
// - Hash output (32 bytes)
```

### Verify Password

```javascript
const hash = user.passwordHash; // From database

const valid = await fazt.lib.password.verify('entered-password', hash);

if (valid) {
  // Password correct
} else {
  // Password incorrect
}
```

### Check If Rehash Needed

```javascript
// When you upgrade security parameters, existing hashes may be outdated
const needsRehash = fazt.lib.password.needsRehash(hash);

if (needsRehash && passwordIsValid) {
  // Re-hash with current parameters
  const newHash = await fazt.lib.password.hash(password);
  await updateUserHash(userId, newHash);
}
```

### Custom Parameters

```javascript
// For high-security applications (slower but stronger)
const hash = await fazt.lib.password.hash('password', {
  memory: 128 * 1024,  // 128MB (default: 64MB)
  time: 4,             // 4 iterations (default: 3)
  parallelism: 4       // 4 threads (default: 4)
});

// For resource-constrained environments
const hash = await fazt.lib.password.hash('password', {
  memory: 32 * 1024,   // 32MB
  time: 3,
  parallelism: 2
});
```

## Security Properties

| Property               | Argon2id                       |
| ---------------------- | ------------------------------ |
| Memory-hard            | Yes (resists GPU/ASIC attacks) |
| Time-hard              | Yes (configurable iterations)  |
| Salt                   | Automatic (16 bytes random)    |
| Side-channel resistant | Yes (id variant)               |
| Output                 | 32 bytes (256 bits)            |

## Why Argon2id?

- **Winner of Password Hashing Competition (2015)**
- **OWASP recommended** for new applications
- **Memory-hard**: Attackers can't parallelize on GPUs
- **Argon2id variant**: Resistant to both side-channel and GPU attacks
- **Better than bcrypt/scrypt**: More configurable, better security margins

## JS API

```javascript
fazt.lib.password.hash(plaintext, options?)
// options: { memory: number, time: number, parallelism: number }
// Returns: string (PHC format hash)

fazt.lib.password.verify(plaintext, hash)
// Returns: boolean (timing-safe comparison)

fazt.lib.password.needsRehash(hash, options?)
// Returns: boolean (true if parameters below current defaults)

fazt.lib.password.config()
// Returns: { memory, time, parallelism } (current defaults)
```

## HTTP Endpoint

Not exposed via HTTP. Password operations must be server-side only.

## Go Library

Uses `golang.org/x/crypto/argon2`:

```go
import (
    "golang.org/x/crypto/argon2"
    "crypto/rand"
    "crypto/subtle"
)

func Hash(password string, opts Options) (string, error) {
    salt := make([]byte, 16)
    rand.Read(salt)

    hash := argon2.IDKey(
        []byte(password),
        salt,
        opts.Time,
        opts.Memory,
        opts.Parallelism,
        32,
    )

    // Encode to PHC format
    return encodePHC(hash, salt, opts), nil
}

func Verify(password, encoded string) bool {
    // Decode PHC format
    hash, salt, opts := decodePHC(encoded)

    // Compute hash with same parameters
    computed := argon2.IDKey(
        []byte(password),
        salt,
        opts.Time,
        opts.Memory,
        opts.Parallelism,
        32,
    )

    // Timing-safe comparison
    return subtle.ConstantTimeCompare(hash, computed) == 1
}
```

## Common Patterns

### User Registration

```javascript
async function register(email, password) {
  // Validate password strength first (app-level policy)
  if (password.length < 8) {
    throw new Error('Password too short');
  }

  const hash = await fazt.lib.password.hash(password);

  await fazt.storage.ds.insert('users', {
    email,
    passwordHash: hash,
    createdAt: Date.now()
  });
}
```

### User Login

```javascript
async function login(email, password) {
  const user = await fazt.storage.ds.findOne('users', { email });

  if (!user) {
    // Timing-safe: still do a hash comparison to prevent enumeration
    await fazt.lib.password.verify(password, '$argon2id$v=19$m=65536,t=3,p=4$dummy$dummy');
    throw new Error('Invalid credentials');
  }

  const valid = await fazt.lib.password.verify(password, user.passwordHash);

  if (!valid) {
    throw new Error('Invalid credentials');
  }

  // Check if we should upgrade the hash
  if (fazt.lib.password.needsRehash(user.passwordHash)) {
    const newHash = await fazt.lib.password.hash(password);
    await fazt.storage.ds.update('users', { id: user.id }, {
      passwordHash: newHash
    });
  }

  return createSession(user);
}
```

### Password Change

```javascript
async function changePassword(userId, currentPassword, newPassword) {
  const user = await fazt.storage.ds.findOne('users', { id: userId });

  const valid = await fazt.lib.password.verify(currentPassword, user.passwordHash);
  if (!valid) {
    throw new Error('Current password incorrect');
  }

  const newHash = await fazt.lib.password.hash(newPassword);

  await fazt.storage.ds.update('users', { id: userId }, {
    passwordHash: newHash,
    passwordChangedAt: Date.now()
  });
}
```

## Default Parameters

OWASP 2024 recommendations:

| Parameter     | Default      | Description            |
| ------------- | ------------ | ---------------------- |
| `memory`      | 65536 (64MB) | Memory cost in KB      |
| `time`        | 3            | Number of iterations   |
| `parallelism` | 4            | Number of threads      |
| `hashLength`  | 32           | Output length in bytes |
| `saltLength`  | 16           | Salt length in bytes   |

These defaults result in ~300ms hash time on typical server hardware.

## Limits

| Limit               | Default                     |
| ------------------- | --------------------------- |
| `maxPasswordLength` | 72 bytes (longer truncated) |
| `maxMemoryMB`       | 256                         |
| `maxTime`           | 10                          |
| `maxParallelism`    | 16                          |

## Implementation Notes

- ~30KB binary addition
- Pure Go (x/crypto/argon2, no CGO)
- Constant-time comparison prevents timing attacks
- PHC string format for hash portability
- Memory allocation capped to prevent DoS

