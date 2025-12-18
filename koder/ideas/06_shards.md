# Micro-Document Storage Pattern

## High-Performance NoSQL inside the Cartridge Kernel

### 1. What is it?

A storage pattern using **SQLite** as a document database.
It replaces giant JSON blobs with **Sharded Rows**.
It uses **Functional Indexes** to query JSON keys at B-Tree speeds.

### 2. Significance

Allows Fazt to scale to **millions of records** without custom tables.
Enables relational-speed lookups on unstructured data.
Keeps the database "Rigid" while making the data "Fluid."

### 3. Core Components

* **Prefix Sharding**: Keys are split (e.g., `checkin:2025:shard_1`).
* **Functional Indexing**: Indexing `json_extract(value, '$.id')`.
* **Batch Buffering**: RAM-to-Disk flushing via the `proc` layer.
* **JSONB Storage**: Binary JSON format for 2-3x faster parsing.

### 4. Advantages vs. Disadvantages

**Advantages:**

* **Scale**: Handles 500k+ events/day on a $6 VPS.
* **Speed**: Instant searches on specific JSON fields.
* **Simplicity**: No SQL migrations needed for new features.

**Disadvantages:**

* **Redundancy**: Data is stored in the table and the index.
* **Cold Storage**: Non-indexed fields still require slow table scans.

### 5. Implementation Rules

* **Index the Intent**: Only index keys used for `WHERE` or `JOIN`.
* **Shard by Pattern**: Auto-split rows when a key prefix hits size limits.
* **Transparent API**: The `syscall` layer handles indexing automatically.
* **EDD Compliance**: Never drop an index; only add better ones.

### 6. Architect's Advice

Treat the DB as a **Searchable Log**.
Use the **Kernel** to hide sharding logic from the app.
Store everything as JSON; index only what matters for speed.
Sanity is found in **Batching**; Uptime is found in **Indexes**.

---

If you like this, I can define the **`syscall` interface** that allows a serverless app to "Register" an index for a specific JSON key pattern. Do you want to see that?