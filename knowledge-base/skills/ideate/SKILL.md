---
name: fazt-ideate
description: Brainstorm ideas for Fazt's evolution - uses reasoning model for strategic planning and creative exploration of architectural possibilities.
model: opus
allowed-tools: Read, Write, Edit, Glob, Grep
---

# Fazt Ideation Session

## Context Loading

Load context in layers - philosophy first, then build capability surface:

### Layer 1: Philosophy (required)

1. Read `koder/philosophy/VISION.md` - Strategic vision, swarm model
2. Read `koder/philosophy/CORE.md` - Foundational principles
3. Read `koder/philosophy/SENSORS.md` - Sensor philosophy + Go libraries
4. Read `koder/philosophy/EVOLUTION.md` - How we got here

### Layer 2: Roadmap Overview (required)

5. Read `koder/ideas/ROADMAP.md` - Version progression
6. Read `koder/ideas/SURFACE.md` - API surface evolution

### Layer 3: Capability Surface (required)

Read specs progressively to build the full capability surface:

```
koder/ideas/specs/v0.8-kernel/     → Kernel primitives, events, devices
koder/ideas/specs/v0.9-storage/    → Storage layer
koder/ideas/specs/v0.10-runtime/   → Serverless, stdlib, sandbox
koder/ideas/specs/v0.11-distribution/ → Marketplace, manifest
koder/ideas/specs/v0.12-agentic/   → AI harness, MCP, ai-shim
koder/ideas/specs/v0.13-network/   → Domains, VPN
koder/ideas/specs/v0.14-security/  → RLS, notary, halt
koder/ideas/specs/v0.15-identity/  → Persona, identity
koder/ideas/specs/v0.16-mesh/      → P2P, protocols, discovery
koder/ideas/specs/v0.17-realtime/  → WebSocket
koder/ideas/specs/v0.18-email/     → Email sink
koder/ideas/specs/v0.19-workers/   → Background jobs
koder/ideas/specs/v0.20-services/  → Libraries (sanitize, markdown, etc.)
```

For each version directory:
1. Read README.md first (if exists) for overview
2. Scan *.md files to understand capabilities
3. Note Go libraries specified (implementation decisions)

**Assume**: Everything up to v0.20 is the planned capability surface.
Ideation proposes extensions, simplifications, or new directions beyond this.

## Brainstorming Directions

Each session, pick **2-3 directions** from below.
Vary the selection to keep sessions fresh.
Consider user's input/arguments if provided.

### Direction A: Extend Primitives
- What new kernel primitives would unlock significant capability?
- Apply the primitive-vs-pattern test (only propose true primitives)
- Consider: Does this compose with existing primitives?

### Direction B: Extend Services
- What higher-level services would provide value?
- Services should compose from kernel primitives
- Consider: Is this general enough to be a platform service?

### Direction D: Use Case Expansion
- What valuable use cases are *almost* supported but need 1-2 additions?
- What's the minimum change for maximum new capability?
- Think: "If we just added X, we could do Y, Z, and W"

### Direction J: Gap Analysis
- Compare Fazt to alternatives (Coolify, Dokku, etc.)
- What do they have that we don't (and should we care)?
- What do we have that they don't (and how do we emphasize it)?

## Quality Gates

For any concrete proposal, evaluate:

### Primitive Test
```
Is it a primitive?     Can it be decomposed further?
Is it composable?      Can existing primitives achieve this?
Is it necessary?       Does it enable things otherwise impossible?
```

### Philosophy Alignment
```
Single binary:         Does this fit in one executable?
Pure Go:               Can this be done without CGO?
JSON everywhere:       Does data flow as JSON?
Events as spine:       Does this integrate via events?
```

### Value Assessment
```
Intrinsic value:       What's the value if friction were zero?
Current friction:      Why isn't this used more today?
Fazt collapse:         Can Fazt dramatically reduce friction?
```

### Cost Assessment
```
Complexity:            How much does this complicate the system?
Maintenance:           What's the ongoing burden?
Binary size:           Significant impact?
```

## Output Format

For each brainstorming direction explored:

### [Direction Name]

**Observation:** What did you notice?

**Analysis:** Why does this matter?

**Proposal (if any):**
```
Name: [Feature/Change]
Type: primitive | service | simplification | removal
Layer: kernel | runtime | services | philosophy

What: One sentence
Why: Value proposition
Cost: low | medium | high
Alignment: How it fits the philosophy

API Surface (if applicable):
- fazt.namespace.method()
- fazt cli command
```

**Alternative considered:** What else could address this?

## Conversation Flow

- Present 2-3 directions with analysis
- User may ask to explore other directions
- User may challenge proposals
- User may say "build it" for a specific idea
- Be willing to say "on reflection, this doesn't fit"

## When User Approves an Idea

1. Determine which version it fits (or propose new)
2. Create spec: `koder/ideas/specs/v0.X-*/feature-name.md`
3. Update version README.md if exists
4. Update SURFACE.md with new APIs
5. Keep specs concise, match existing style

Then: "Added. Run `/koder-ideate` again when ready for more."

## Session Goals

Every ideation session should aim to:
- **Improve**: Make Fazt better at what it does
- **Extend**: Add capability where valuable
- **Align**: Ensure coherence with philosophy
- **Prioritize**: Focus on adoption and robustness

The best ideas are often subtractions, not additions.
