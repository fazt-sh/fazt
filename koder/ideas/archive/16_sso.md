# Vision Document: Fazt v0.16 - The Sovereign Identity Provider (Persona Proximity)

**Status**: Draft / Request for Comment (RFC)

**Target**: v0.16 Architecture

**Scope**: Sovereign Identity & Zero-Effort Single Sign-On (SSO)

---

## 1. Executive Summary: The Trust Circle

**Fazt v0.16** introduces the **Sovereign Identity Provider (IdP)**. Building upon the **Persona** primitive established in the Notary Kernel, this module eliminates the complexity of OAuth 2.0 and OpenID Connect (OIDC) for apps living within your personal domain.

Instead of treating subdomains (e.g., `notes.abc.com`) as separate entities requiring complex handshakes, the Kernel treats them as members of a **Sovereign Trust Circle**. When you are authenticated at the "OS Shell" (`abc.com`), every child application instantly recognizes your "Persona" through a zero-latency kernel lookup, providing a "Sign in with My Domain" experience with zero configuration.

---

## 2. Architectural Pillars

### A. The Root of Trust

The Kernel leverages its position as the sole manager of both the root domain and all subdomains.

* **Session Inheritance**: The Admin Dashboard and primary authentication state are pinned in the Kernel's memory.
* **Implicit Verification**: When a request hits a subdomain, the Kernel's security middleware inspects the session context before the request ever reaches the user app.

### B. Hardware-Bound Persona

Following the Notary Kernel specification, every Fazt instance is born with a hardware-bound keypair.

* **The Blueprint**: Identity is not a password stored in a table, but a cryptographic proof managed by the Kernel's executive layer.
* **Asymmetric Assertions**: For apps requiring formal tokens, the Kernel generates OIDC-compatible assertions signed by its internal vault.

### C. Zero-Handshake SSO

We reject the "Redirect Dance" typical of Google/Github SSOs.

* **Internal Access**: Apps use a native `syscall` to ask the Kernel for the current user's identity.
* **No Secrets**: Because the Kernel manages the App UUID and the VFS, it does not require apps to store "Client Secrets".

---

## 3. The "Persona" Syscall Map

Applications interact with the identity layer through the `fazt.security` namespace:

| Syscall | Capability |
| --- | --- |
| `fazt.security.getPersona()` | Returns the authenticated Owner's profile (Username, Email, Public Key). |
| `fazt.security.isOwner()` | Boolean check to see if the current request originates from the primary owner. |
| `fazt.security.signAssertion()` | Generates a cryptographically signed identity packet for third-party verification. |
| `fazt.security.requireAuth()` | A kernel-level "halt" that forces a login/TOTP prompt if the user is not authenticated. |

---

## 4. Strategic Value: Sovereign SSO

This architecture transforms Fazt into a personal **Identity Anchor** for your digital life:

1. **Effortless Multi-App Usage**: You can host 50 different micro-tools (Notes, CRM, Budget, etc.) and move between them as seamlessly as switching tabs in a browser.
2. **Privacy by Physics**: Your identity data never leaves your binary. Third-party apps only receive assertions signed by your server, keeping your raw credentials vaulted.
3. **The OIDC Bridge**: In future versions (v0.18+), your Fazt instance can act as a "Sign in with Fazt" provider for *external* websites, allowing you to use your personal domain as your global login across the web.

---

### Architect's Advice

The goal of "Persona Proximity" is to make the developer feel like they are writing for a single-user machine, even if they are building a complex multi-domain ecosystem. By moving identity verification into the **Kernel**, we earn the right to build "Vibe-coded" apps that don't need to worry about the "login problem".