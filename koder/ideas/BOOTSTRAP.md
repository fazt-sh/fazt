# Fazt Ideas: AI Onboarding

## Purpose

This folder contains the **architectural vision** for Fazt's evolution from a
static site host (v0.7) to a Personal Cloud OS (v0.16+).

Use this documentation to:
- Understand the target architecture
- Simulate future capability surfaces
- Brainstorm new ideas building on existing foundations
- Identify optimal build sequences

## Reading Order

1. **ROADMAP.md** - Version progression with brief descriptions
2. **SURFACE.md** - How the API/syscall surface evolves per version
3. **specs/** - Detailed specifications organized by version

## Current State: v0.7.x (Cartridge PaaS)

Fazt today is a single Go binary + SQLite database that provides:
- Static site hosting via VFS (files stored as BLOBs)
- Auto HTTPS via CertMagic
- Reserved domains: `admin.*`, `root.*`, `404.*`
- Basic serverless JS runtime (Goja)
- Analytics, redirects, webhooks

**Philosophy**: Binary is disposable. Database is precious.

## Target State: v0.16 (Personal Cloud OS)

The vision is a **Sovereign Cloud Platform** in a single binary:
- Full PaaS with intelligent runtime
- OS-like architecture (kernel, syscalls, drivers)
- AI-native capabilities
- Hardware-bound identity
- P2P mesh synchronization
- Protocol support (ActivityPub, Nostr)

## How to Use This Documentation

### For Brainstorming

1. Read ROADMAP.md to understand the version sequence
2. Pick a version to "simulate" as implemented
3. Reason about what capabilities are now available
4. Propose new ideas that build on that foundation

### For Implementation Planning

1. Read the spec for the target version
2. Identify dependencies on prior versions
3. Check SURFACE.md for the API contract
4. Build against that specification

## Conventions

### Document Structure

Each spec follows this format:

```markdown
# [Feature Name]

## Summary
One paragraph. What is it?

## Rationale
Why build this? What problem does it solve?

## Capability Surface
New syscalls, APIs, or CLI commands introduced.

## Implementation Notes
Technical constraints, dependencies, caveats.

## Open Questions
Unresolved design decisions.
```

### Naming

- Specs are organized by version: `specs/v0.X-codename/`
- Each version has a `README.md` summarizing its scope
- Individual features get atomic `.md` files

### Versioning Logic

| Range | Theme |
|-------|-------|
| v0.8 | Kernel pivot, stability |
| v0.9 | Storage evolution |
| v0.10 | Runtime capabilities |
| v0.11 | Distribution (marketplace) |
| v0.12 | Agentic capabilities |
| v0.13 | Network primitives |
| v0.14 | Security primitives |
| v0.15 | Identity layer |
| v0.16 | P2P mesh |

## Key Principles

1. **Single Binary** - No external dependencies ever
2. **Single Database** - All state in `data.db`
3. **Evolutionary Schema** - Append-only, no destructive migrations
4. **Resource Constrained** - Must run on 1GB RAM / $6 VPS
5. **AI-Native** - Designed for agent consumption and manipulation

## Session Protocol

When starting a brainstorming session:

1. State which version you're "simulating" as complete
2. List the capabilities now available
3. Propose ideas that build on that surface
4. Document new ideas following the spec template
5. Suggest where they fit in the version roadmap

---

## Adding New Ideas

### To Add a Feature to an Existing Version

1. Create `specs/v0.X-codename/feature-name.md` using the template below
2. Update `specs/v0.X-codename/README.md` to list the new document
3. Add new syscalls/CLI commands to `SURFACE.md` under that version
4. If it adds dependencies, update the dependency graph in `ROADMAP.md`

### To Create a New Version (v0.17+)

1. Create folder `specs/v0.17-codename/`
2. Create `specs/v0.17-codename/README.md` with theme and goals
3. Add individual feature specs
4. Add version section to `ROADMAP.md`
5. Add version section to `SURFACE.md`
6. Update dependency graph in `ROADMAP.md`

### Spec Template

```markdown
# Feature Name

## Summary

One paragraph. What is this feature?

## Rationale

Why build this? What problem does it solve?

## Capability Surface

New additions to `fazt.*` namespace, CLI, or HTTP API:

```javascript
fazt.namespace.method(args)
```

```bash
fazt command subcommand [options]
```

## Implementation Notes

- Technical constraints
- Dependencies on other features
- Resource considerations

## Open Questions

- Unresolved design decisions
- Trade-offs to discuss
```

### Naming Conventions

| Type | Pattern | Example |
|------|---------|---------|
| Version folder | `v0.X-codename/` | `v0.17-email/` |
| Feature spec | `feature-name.md` | `smtp-sink.md` |
| CLI commands | `fazt <module> <action>` | `fazt email receive` |
| JS namespace | `fazt.<module>.<method>()` | `fazt.email.inbox()` |

### Checklist for New Ideas

- [ ] Spec file created with all sections
- [ ] README.md updated to list new spec
- [ ] SURFACE.md updated with new APIs
- [ ] ROADMAP.md updated if new version or dependencies
- [ ] Dependencies clearly stated
- [ ] Resource impact considered (1GB RAM constraint)

---

## Extending Beyond v0.16

Ideas not yet assigned to versions are tracked in `ROADMAP.md` under
"Beyond v0.16". When ready to formalize:

1. Group related ideas into a themed version
2. Assign a codename that captures the theme
3. Create the version folder and specs
4. Move items from "Beyond v0.16" to the new version section
