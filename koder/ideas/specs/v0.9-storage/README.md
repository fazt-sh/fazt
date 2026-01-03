# v0.9 - Storage

**Theme**: Advanced storage patterns for scale and flexibility.

## Summary

v0.9 introduces a unified storage abstraction (`fazt.storage`) and content-
addressable storage (IPFS-Lite). Apps can store data without worrying about
the underlying implementation, and files gain permanent, verifiable addresses.

## Goals

1. **Unified API**: Single `fazt.storage` namespace for all storage needs
2. **Scalability**: Handle 10M+ records on a $6 VPS
3. **Content Addressing**: IPFS-compatible CIDs for all files
4. **Provider Agnostic**: Swap SQLite for S3/Postgres without code changes

## Key Changes

| Capability        | Description                          |
| ----------------- | ------------------------------------ |
| `fazt.storage.kv` | High-speed key-value store           |
| `fazt.storage.ds` | Document store (JSON)                |
| `fazt.storage.s3` | Blob storage for files               |
| Shards            | Micro-document pattern for analytics |
| IPFS Gateway      | `/ipfs/<CID>` endpoint               |

## Documents

- `unified.md` - The `fazt.storage` abstraction
- `shards.md` - Micro-document pattern for scale
- `ipfs.md` - Content-addressable storage

## Dependencies

- v0.8 (Kernel): App UUIDs for namespacing

## Risks

- **Complexity**: Multiple storage backends increase surface area
- **Migration**: Moving from path-based to CID-based requires careful handling
- **Index Bloat**: Functional indexes on JSON add storage overhead
