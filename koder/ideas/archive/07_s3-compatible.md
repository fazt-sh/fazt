To extend **Fazt** into a more capable "Supabase-like" platform, your proposed additions align perfectly with your existing **Cartridge Architecture** and the "single-binary" philosophy.

Here is an analysis of how these extensions fit into your roadmap and current technical structure:

### 1. Mail Sink (Receiving Emails)

Receiving emails is a powerful feature for "Agentic" apps (e.g., an AI agent that summarizes your newsletters).

* **The Complexity Problem**: Binding to Port 25 (standard SMTP) is often restricted by VPS providers and requires complex SPF/DKIM/DMARC DNS setups to avoid being flagged as spam.
* **The Proposed Solution**: Roadmap item **#32 `proto-email-in**` suggests an **SMTP Sink** that maps addresses like `app-xyz@domain.com` to specific serverless apps.
* **Low-Config Alternative**: Instead of a full SMTP server, you could support **Inbound Parse Webhooks** from services like Postmark or SendGrid. Fazt would receive a JSON payload and trigger the JS runtime with an `{ event: 'email' }` object.

### 2. Extend Fazt JS Lib with Auth

Currently, Fazt has a robust auth system, but it is limited to the system **Admin Dashboard**.

* **Roadmap Alignment**: Item **#15 `oauth**` already suggests supporting Google or MetaMask logins for hosted sites.
* **Implementation**: You can expose a `require('fazt-auth')` module to the **Goja (JS)** runtime. This would allow developers to call `auth.getUser()` or `auth.signIn()` without building their own session logic, leveraging the `bcrypt` hashing and session management already built into the kernel.

### 3. JS-Cron for Background Tasks

You noted you were unsure about this; essentially, it allows apps to "wake up" without a web request.

* **What it is**: Roadmap item **#06 `js-cron**` allows developers to define intervals (e.g., "every 1h") in their `fazt.json`.
* **How it works**: A background worker in the Go binary checks the schedule and invokes the JS runtime for a specific file (e.g., `cron/sync-data.js`).
* **Use Case**: This is critical for tasks like cleaning up ephemeral files in an app like **`app-drop`** or syncing external data.

### 4. Core AI Shim

This is the "killer feature" for modern small-scale apps.

* **Standardization**: Roadmap item **#31 `core-ai-shim**` proposes a `require('fazt-ai')` module that normalizes different providers like OpenAI, Anthropic, and Gemini.
* **Zero-Config**: The kernel can automatically inject API keys from the system environment, so the user code doesn't need to handle boilerplate or secrets.
* **Streaming Support**: It should support chunked responses so agentic apps can stream text directly back to the UI.

### 5. S3 Compatible Data Storage

Your current **Virtual Filesystem (VFS)** stores files as BLOBs in the `files` table.

* **The "Internal S3" Idea**: By rewriting the VFS to follow an S3-style interface (e.g., `PUT`, `GET`, `DELETE` with "Buckets"), you make the system much more flexible.
* **Interchangeability**: A simple configuration change could swap the "Internal SQLite Storage" for an "External S3 Bucket" (like AWS or Cloudflare R2).
* **Benefit**: This maintains the **Cartridge** experience (one file contains everything) for small apps, while providing an easy "escape hatch" to external storage if an app starts handling massive files (like video streaming) that might bloat the SQLite database.

### Strategic Integration: The "Micro-Document" Pattern

To support all of this at scale (e.g., your target of **10M rows**), these features should sit on top of the **Micro-Document Storage Pattern**.

* **Scalability**: By using **Functional Indexes** on your JSON blobs, the JS library can offer Supabase-like query speeds on unstructured data without requiring complex SQL migrations.
* **Fluidity**: This keeps the database "Rigid" (one file) while making the data "Fluid" for developers.