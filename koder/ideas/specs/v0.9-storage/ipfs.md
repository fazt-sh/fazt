# IPFS-Lite: Content-Addressable Storage

## Summary

v0.9 introduces Content-Addressable Storage (CAS). Files are indexed by their
IPFS-compatible hashes (CIDs), enabling verifiability, de-duplication, and
permanent links.

## Rationale

### The Problem with Path-Based Storage

Traditional URLs are fragile:
- `https://example.com/photo.jpg` can change or disappear
- No way to verify content hasn't been modified
- Duplicate files stored multiple times

### The CID Solution

Content IDs (CIDs) are hashes of the file content:
- Same content = same CID, always
- CID is proof of integrity
- Global de-duplication

```
Path-based:  /apps/blog/photo.jpg (mutable, local)
CID-based:   /ipfs/QmXoypiz... (immutable, global)
```

## Architecture

### The Ghost Index

Decouple file metadata from content:

```
┌─────────────────────────────────────┐
│            files table              │
│  (path, cid, app_uuid, updated_at)  │
└─────────────────┬───────────────────┘
                  │ references
                  ▼
┌─────────────────────────────────────┐
│            blobs table              │
│         (cid, content)              │
│         Append-only                 │
└─────────────────────────────────────┘
```

### Schema

```sql
-- Content layer (immutable)
CREATE TABLE blobs (
    cid TEXT PRIMARY KEY,    -- IPFS-compatible multihash
    content BLOB,
    size INTEGER,
    created_at INTEGER
);

-- Metadata layer (mutable)
CREATE TABLE files (
    app_uuid TEXT,
    path TEXT,
    cid TEXT,                -- References blobs.cid
    mime_type TEXT,
    updated_at INTEGER,
    PRIMARY KEY (app_uuid, path)
);
```

### Write Flow

```
1. App writes file to /photo.jpg
2. Kernel computes CID from content
3. IF cid exists in blobs: skip content write (de-dup)
4. ELSE: insert into blobs
5. Update files table: path → cid mapping
```

### Read Flow

```
1. Request for /photo.jpg
2. Lookup files: path → cid
3. Lookup blobs: cid → content
4. Serve with CID-based ETag
```

## IPFS Gateway

### Standard URL Pattern

```
https://example.com/ipfs/{CID}
https://example.com/ipfs/{CID}/{PATH}
```

This matches the global IPFS gateway standard.

### Behavior

1. **Local Hit**: CID exists in `blobs` → serve directly
2. **Remote Proxy** (optional): Missing → fetch from `ipfs.io`, cache locally
3. **Folder Listings**: If CID is a directory (Merkle DAG) → render listing

### Reserved Path

The `/ipfs/` and `/ipns/` paths are intercepted before user apps:

```go
func (r *Router) ServeHTTP(w, req *http.Request) {
    if strings.HasPrefix(req.URL.Path, "/ipfs/") {
        r.kernel.FS.ServeIPFS(w, req)
        return
    }
    // ... normal routing
}
```

## CLI Commands

```bash
# Upload file, get CID
fazt storage upload ./photo.jpg
# Output: QmXoypizjW3WknFiJnKLwHCnL72vedxjQkDDP1mXWo6uco

# Get CID for existing path
fazt fs cid /apps/blog/photo.jpg
# Output: QmXoypizjW3WknFiJnKLwHCnL72vedxjQkDDP1mXWo6uco

# Deploy folder as Merkle DAG
fazt net deploy ./dist --slug blog
# Output: Root CID: QmRootHash...
```

## JS Runtime API

```javascript
// Get permanent IPFS URL for a path
const url = fazt.fs.ipfsUrl('/photo.jpg');
// Returns: "https://example.com/ipfs/QmXoypiz..."

// Get raw CID
const cid = fazt.fs.cid('/photo.jpg');
// Returns: "QmXoypiz..."

// Check if content exists by CID
const exists = await fazt.fs.hasCid('QmXoypiz...');
```

## Advanced Features

### Merkle DAG Folders

Folders become content-addressable:

```
root/
├── index.html  → QmIndex...
├── style.css   → QmStyle...
└── images/
    └── logo.png → QmLogo...

Root CID: QmRootFolder...
```

Share one CID for an entire versioned folder.

### Git Compatibility

Git objects are also content-addressed. The `blobs` table can store Git
objects, enabling de-duplication between repo data and deployed assets.

### External Pinning

Fazt is "lite"—no DHT or Bitswap. For global persistence:

```json
{
  "ipfs": {
    "pinning_service": "pinata",
    "api_key": "..."
  }
}
```

New CIDs are automatically sent to the pinning service.

## Benefits

| Feature | Description |
|---------|-------------|
| **Verifiability** | CID proves content integrity |
| **De-duplication** | Same file stored once |
| **Permanent Links** | CIDs never change |
| **Cache-Friendly** | Immutable = infinite cache |
| **Interoperability** | Compatible with global IPFS network |

## Open Questions

1. **Garbage Collection**: When to remove orphaned blobs?
2. **Size Limits**: Max blob size before streaming?
3. **Remote Fetch**: Auto-fetch missing CIDs from gateway?
