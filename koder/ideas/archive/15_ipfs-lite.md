# Vision Document: Fazt v0.9 - IPFS-Lite & Content-Addressable Storage

**Status**: Draft / Request for Comment (RFC)

**Target**: v0.9 Architecture

**Scope**: Storage Evolution & P2P Primitives

---

## 1. Executive Summary: From Locations to Truth

Fazt v0.8 established the **Kernel** metaphor, treating the binary as an OS. Fazt v0.9 evolves the **`storage/`** and **`fs/`** modules from path-based storage to **Content-Addressable Storage (CAS)**.

By indexing files by their IPFS-compatible hashes (CIDs), Fazt becomes a "notary" for digital content. It provides the benefits of the decentralized web—verifiability, de-duplication, and permanence—without the resource overhead of a full P2P swarm.

---

## 2. Architectural Pivot: The "Ghost Index"

The Kernel transparently decouples file metadata (names/paths) from the underlying data (blobs).

### A. The Content Layer (Immutable)

* **Table**: `blobs` (or `hashes`).
* **Primary Key**: `cid` (IPFS-compatible Multihash).
* **Value**: `content` (BLOB).
* **Behavior**: Append-only. If a hash exists, the write is ignored (Global De-duplication).

### B. The Metadata Layer (Mutable)

* **Table**: `files`.
* **Columns**: `path`, `cid` (Foreign Key), `site_id`, `updated_at`.
* **Behavior**: Maps a human-readable path to a verifiable CID.

---

## 3. The "Standard Gateway" Pattern

To ensure drop-in compatibility with the global IPFS ecosystem, Fazt adopts the standard gateway URL structure.

* **Standard URL**: `https://<domain>/ipfs/<CID>/<PATH>`
* **Reserved Path**: The **`net/`** module intercepts all `/ipfs/` and `/ipns/` requests before they hit user apps.
* **Behavior**:
1. **Local Hit**: If the CID exists in the local `blobs` table, serve it directly.
2. **Remote Proxy**: If missing, the Kernel optionally proxies the request from a public gateway (e.g., `ipfs.io`) and caches the result locally.
3. **Folder Listings**: If a CID represents a directory (Merkle DAG), the Kernel renders a standard, clean file listing.



---

## 4. CLI & API Integration (OS Nomenclature)

The CLI and JS runtime are updated to treat storage as a system-level "Syscall".

### A. CLI Commands

* **`fazt storage upload <FILE>`**: Generates a CID, stores the blob, and returns the IPFS link.
* **`fazt net deploy <FOLDER>`**: Hashes the folder into a Merkle DAG, uploads blobs, and maps a domain to the Root CID.
* **`fazt fs cid <PATH>`**: Returns the CID for any path on a hosted site.

### B. JS Runtime (`fazt.fs`)

```javascript
// Get a permanent link for a resource
const link = fazt.fs.getIPFS("/photo.jpg"); 
// Result: https://zyt.app/ipfs/QmXoyp...

```

---

## 5. Advanced Capabilities

### A. Merkle DAG Folders

Folders are no longer just "lists of files." They are directory blocks containing filenames and CIDs. This allows a user to share a single **Root CID** that represents an entire versioned folder.

### B. Git Compatibility

Since Git is also content-addressable, the `blobs` table can natively store Git objects. This allows Fazt to act as a Git remote that automatically de-duplicates repository data against deployed site assets.

### C. External Pinning

The Kernel remains "lite" by not running a DHT or Bitswap. For global persistence, a "Pinning Driver" in the **`driver/`** module can automatically send CIDs to services like Pinata via API.

---

## 6. Architect's Verdict

This implementation achieves the "Unbelievable Achievement" of v0.8: A robust Cloud Platform in a single binary. It adds **Identity-Location Decoupling** with negligible CPU/RAM overhead, making Fazt v0.9 the most efficient gateway to the permanent web for $6/mo VPS users.