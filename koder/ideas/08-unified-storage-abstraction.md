# Unified Storage Abstraction: The "Everything is a Resource" Pattern

## 1. Vision
Provide a single, provider-agnostic JS API (`fazt.storage`) that handles all
application state. Whether the data lives in a local SQLite file or a global
AWS region, the application code remains identical.

## 2. Rationale
Startups need to move fast without infra-bloat. Fazt allows them to build using
enterprise patterns (S3, Document Stores) on a $5 VPS. When they hit scale,
they swap a config line—not their code.

## 3. The Four Pillars
* **`fazt.storage.kv`**: High-speed, persistent Key-Value store.
* **`fazt.storage.ds`**: Document Store using the Micro-Document Sharding pattern.
* **`fazt.storage.rd`**: Relational DB using namespaced virtual tables.
* **`fazt.storage.s3`**: Blob storage for user-generated content and media.

## 4. Implementation Logic
The Kernel intercepts storage calls and routes them based on the `app.json` 
provider configuration:
* **Internal**: Routes to local SQLite tables (e.g., `app_id_orders`).
* **External**: Translates calls to S3 APIs, Postgres drivers, or Redis.

## 5. The Migration Edge (The "Escape Hatch")
Since the Kernel manages the abstraction, it provides built-in migration 
tooling to stream internal SQLite data to external providers with zero 
application-level code changes.

## 6. Architect's Advice
Design the API to match industry standards (e.g., S3's PUT/GET). This ensures
that AI agents can write valid Fazt code using pre-existing knowledge and 
makes the platform immediately familiar to senior developers.