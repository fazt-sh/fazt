---
description: Brainstorm ideas for Fazt's evolution (project)
model: opus
allowed-tools: Read, Write, Edit, Glob, Grep
---

# Fazt Ideation Session

## Context Loading

First, load the philosophical and technical foundation:

1. **Philosophy** (required):
   - Read `koder/philosophy/CORE.md` - Foundational principles
   - Read `koder/philosophy/SENSORS.md` - Sensor-specific philosophy
   - Read `koder/philosophy/EVOLUTION.md` - How we got here

2. **Technical Surface** (required):
   - Read `koder/ideas/ROADMAP.md` - Version progression
   - Read `koder/ideas/SURFACE.md` - Current API surface

3. **Assume**: Everything up to v0.20 is implemented. Sensor/percept/effect architecture is planned.

## Brainstorming Directions

Each session, pick **2-3 directions** from below. Vary the selection to keep sessions fresh. Consider user's input/arguments if provided.

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
