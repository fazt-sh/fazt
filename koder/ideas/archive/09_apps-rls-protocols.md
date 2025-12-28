The architecture for **Fazt v0.8** represents a pivot from a simple web host to a **Personal Cloud OS** designed for privacy, "Vibe Coding," and AI agents. By combining lightweight execution with powerful kernel-level abstractions, the system provides a "Digital Vault" for the individual.

Here is the combined summary of the core pillars discussed:

### 1. Kernel-Level Row Level Security (RLS)

Instead of complex database-level policies, Fazt implements security at the **Go Kernel** layer.

* **Automatic Multi-Tenancy**: The Kernel identifies the `app_id` and `user_id` from the active session or JWT.
* **Query Rewriting**: When an app calls `fazt.storage`, the Go-based storage wrapper automatically appends filters (e.g., `WHERE app_id = ? AND user_id = ?`) to every SQLite query.
* **Developer Simplicity**: Solo developers can build multi-user apps without writing manual security checks, as the Kernel ensures users can only access their own data.

### 2. The "Everything is an App" Paradigm

Fazt treats applications as modular "Cartridges" that extend the OS's capabilities effortlessly.

* **GitHub Installation**: Apps are installed by simply providing a GitHub URL. The Kernel fetches the code and hydrates it into the **Virtual Filesystem (VFS)** in SQLite.
* **Security & Isolation**: Every app is strictly namespaced by its **App ID**. The Kernel ensures an app cannot access the files or database rows of another app unless explicitly permitted.
* **Real-time Collaboration (Yjs)**: Apps can implement "Google Docs" style collaboration using **CRDTs (Yjs)**. Update deltas are broadcast via the Kernel's WebSocket hub and stored as binary BLOBs in the KV store, allowing for 10-person editing with minimal overhead.
* **Examples of "Batteries-Included" Apps**:
* **Fazt.Drive**: Personal storage with S3-ready abstraction.
* **Fazt.Chat**: A micro-Discord for small, private communities.
* **Fazt.Notes**: Collaborative markdown editor.
* **Fazt.Bot**: A command center for agentic micro-tools.



### 3. "None-by-Default" Permission Layer

To maintain the **"Unsinkable Kernel"** philosophy, Fazt uses a strictly controlled permission system.

* **Manifest-Based Requests**: Apps request specific privileges (e.g., access to the global scheduler or raw networking) via the `app.json` manifest.
* **User Consent**: During installation, users must explicitly grant these high-risk permissions.
* **Sandbox Safety**: By default, apps have no access to system services. This prevents malicious "Git-installed" code from crashing the 1GB RAM VPS or exhausting resources.

### 4. Decentralized Protocol Support

Fazt prioritizes **protocols over platforms**, allowing users to integrate with global decentralized networks.

* **ActivityPub (Mastodon Lite)**: By implementing ActivityPub, Fazt apps can federate with the main Mastodon network, giving you a sovereign account at `@you@yourdomain.com`.
* **Nostr**: Native support for Nostr allows Fazt to act as a personal relay or client, ensuring your social identity is cryptographically signed and platform-independent.
* **BitChat/Noise**: Enables encrypted, serverless P2P messaging that can work over global Nostr relays or local mesh networks.
* **System Glue (ntfy & Dkron)**: Bundling **ntfy** provides centralized push notifications for all apps, while **Dkron** serves as the engine for the **Hibernate Architecture**, waking apps up for scheduled background tasks.
* **Personal GH (Gitea)**: Integrating a lightweight, pure-Go Git service like Gitea allows users to host and manage their own code repos entirely within their personal "Digital Vault".

This combined architecture empowers the "long tail" of developers—the bottom 95%—to host unlimited, private, and capable projects on a single $6 VPS while maintaining absolute control over their digital life.