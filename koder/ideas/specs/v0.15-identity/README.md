# v0.15 - Identity

**Theme**: Sovereign identity and authentication.

## Summary

v0.15 makes identity a kernel-level primitive. The owner has a hardware-bound
"Persona" that provides seamless SSO across all apps—no OAuth dances, no
password managers, no configuration.

## Documents

- `persona.md` - Hardware-bound identity
- `sso.md` - Zero-handshake single sign-on

## Key Capabilities

### Persona

- Cryptographic identity tied to kernel
- Not a password—a keypair
- Signs assertions for verification

### Sovereign SSO

- Auth at `os.<domain>` propagates to all subdomains
- Apps call `fazt.security.getPersona()` to get owner info
- No redirects, no OAuth, no secrets

## Future: OIDC Bridge

v0.18+ could expose the Persona as an OIDC provider:
- "Sign in with Fazt" for external sites
- Your domain becomes your identity
