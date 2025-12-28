# v0.14 - Security

**Theme**: Cryptographic primitives and trust infrastructure.

## Summary

v0.14 introduces the "Notary Kernel"—a set of cryptographic primitives that
make security a native capability rather than an afterthought.

## Documents

- `notary.md` - Cryptographic primitives
- `rls.md` - Kernel-level row security
- `halt.md` - Human-in-the-loop syscall

## Key Primitives

| Primitive | Description |
|-----------|-------------|
| **Persona** | Hardware-bound keypair per process |
| **Attestation** | Code integrity verification |
| **Vaulting** | Sealed memory for secrets |
| **Temporal** | TOTP integration for sensitive ops |

## Dependencies

- v0.8 (Kernel): Security module foundation
