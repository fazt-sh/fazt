# v0.8 - Kernel

**Theme**: Paradigm shift from "web server" to "operating system".

## Summary

v0.8 transforms Fazt from a web hosting tool into a **Personal Cloud OS**.
The core architectural change is adopting OS-like terminology, structure,
and mental models throughout the codebase.

## Goals

1. **Stability First**: Implement safeguards for 1GB RAM environments
2. **OS Nomenclature**: Rename internals to match OS concepts
3. **App Identity**: Introduce UUIDs, decouple from subdomains
4. **Schema Evolution**: Adopt EDD (append-only migrations)

## Key Changes

| Before (v0.7) | After (v0.8) |
|---------------|--------------|
| Sites | Apps |
| Subdomain-based identity | UUID-based identity |
| `internal/hosting/` | `pkg/kernel/fs/` |
| `internal/config/` | `pkg/kernel/storage/` |
| Destructive migrations | Append-only schema (EDD) |
| No resource limits | Circuit breakers |

## Documents

- `pivot.md` - The paradigm shift rationale
- `nomenclature.md` - OS-like naming conventions
- `edd.md` - Evolutionary Database Design
- `safeguards.md` - Circuit breakers and degradation
- `limits.md` - Per-subsystem soft limits
- `apps.md` - "Everything is an App" model
- `provenance.md` - Data lineage (app_id, user_id on everything)
- `events.md` - Internal event bus (IPC)
- `proxy.md` - Network egress control
- `pulse.md` - Cognitive observability (system self-awareness)
- `devices.md` - External service abstraction (/dev/*)
- `infra.md` - Cloud infrastructure abstraction (VPS, DNS, Domain)
- `beacon.md` - Local network discovery (mDNS)
- `timekeeper.md` - Local time consensus (without NTP)
- `chirp.md` - Audio-based data transfer
- `mnemonic.md` - Human-channel data exchange (word sequences)

## Dependencies

- None (builds on v0.7)

## Risks

- **Migration Complexity**: Renaming sites → apps requires careful migration
- **Breaking Changes**: CLI commands will change significantly
- **Scope Creep**: OS metaphor can invite over-engineering
