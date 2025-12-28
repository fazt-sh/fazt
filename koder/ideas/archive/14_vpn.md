This is a great next step. Based on our conversation, I have drafted the official **Idea Document** for the VPN capability. This document is designed to be checked into your repository (e.g., as `koder/ideas/14_vpn-gateway.md`) to serve as the "Sovereign Specification" for this module.

---

# Vision Document: Fazt v0.13 - The VPN Gateway Primitive

**Status**: Draft / Request for Comment (RFC)

**Target**: v0.13 Architecture

**Scope**: Sovereign Networking & Privacy Gateway

---

## 1. Executive Summary: The Invisible Shield

**Fazt v0.13** introduces the **VPN Gateway**, a built-in, userspace WireGuard implementation that transforms the binary from a web host into a **Sovereign Privacy Shield**. It allows users to route their device traffic through their personal server to mask their IP, encrypt their connection on public networks, and access "internal-only" apps without exposing them to the public internet.

This capability is a **Sovereign Primitive**: it resides quietly in the kernel, requiring zero external dependencies (like `wg-quick`) and remains dormant until the user chooses to provision a peer.

---

## 2. Architectural Pillars

### A. The "Quiet" Net Driver

The VPN is implemented as a core driver within the **`net/` (Network)** module.

* **Userspace Implementation**: Built using `wireguard-go` to maintain the **Zero Dependencies** promise.
* **Latent Readiness**: The UDP port only binds, and the interface only initializes, if at least one peer is configured in the database.
* **Attack Surface Parity**: If unused, the module consumes zero system resources and adds zero network exposure.

### B. Stable API Envelope (The Syscall Pattern)

The VPN is exposed via the **`syscall/`** layer using a protocol-agnostic API.

* **Internal Interface**: `fazt.net.vpn.*` in the JS runtime.
* **Abstraction**: The API focuses on "Trust Context" (e.g., `is_secure()`, `get_peer_identity()`) rather than low-level WireGuard keys.
* **Future-Proofing**: The protocol (WireGuard) is an implementation detail. The API envelope remains identical even if the underlying crypto-tunnel is upgraded in future binary versions.

---

## 3. Storage Pattern: Fluid Configuration

Following the **Micro-Document Storage Pattern** and **Evolutionary Database Design (EDD)**:

* **The Schema**: A `vpn_peers` table stores relational identifiers (Peer ID, Public Key) as primary columns.
* **The Blob**: A single `config` JSON column stores the "Fluid Data" (Device metadata, allowed IPs, custom tags).
* **Advantage**: This allows for "Sideloading" new metadata or device-specific parameters without performing SQL migrations that could break version compatibility.

---

## 4. Capability Surface (The Syscall Map)

Apps can interact with the VPN layer through a safe, read-only interface:

| Syscall | Capability |
| --- | --- |
| `fazt.net.vpn.status()` | Returns whether the current request is routed through the secure tunnel. |
| `fazt.net.vpn.peer_info()` | Provides metadata about the connected device (e.g., "Owner's iPhone"). |
| `fazt.net.vpn.authorize()` | Triggers a **Temporal Identity (TOTP)** prompt to elevate trust for sensitive actions. |

---

## 5. Strategic Value: The "Shadow" Cloud

The VPN Gateway unlocks a new class of "Shadow Apps" that are **Private by Physics**:

1. **VPN-Only Visibility**: Apps can be flagged in `app.json` to only be accessible via the `fazt0` interface. They effectively do not exist on the public internet.
2. **Portable Trust**: Since keys and config live in `data.db`, your entire secure network is as portable as your database.
3. **Zero-Config Gateway**: A user can generate a QR code in the **Dashboard (OS Shell)** and instantly have a private, encrypted gateway for all their mobile traffic.

---

### Architect's Advice

Keep the protocol implementation strictly isolated in `pkg/kernel/net/vpn/`. The rest of the OS should only see a "Secure Tunnel" boolean. This keeps the binary easy to reason about while providing the most powerful privacy primitive possible on a $6 VPS.

---

**Does this capture the "Quiet Existence" and "Great Unlock" aspects as you envisioned them?**