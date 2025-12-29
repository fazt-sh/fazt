---
description: Brainstorm ideas for Fazt's evolution (project)
model: opus
allowed-tools: Read, Write, Edit, Glob
---

# Fazt Ideation Session

## Setup

1. Read `koder/ideas/ROADMAP.md` and `koder/ideas/SURFACE.md`
2. Simulate that everything up to v0.16 is implemented
3. Understand the full capability surface

## Quality Gates (Apply Before Proposing)

For each potential idea, evaluate:

### 1. Primitive or Pattern?
- **Primitive**: Fundamental building block that can't be decomposed further
- **Pattern**: Composition of existing primitives
- Only propose primitives. Patterns can be documentation/examples, not features.

### 2. Composable from Existing?
- Can this be achieved by combining existing primitives?
- If yes → don't propose (it's a pattern)
- If no → continue evaluation

### 3. Value Assessment
```
Intrinsic Value:    What's the value if friction were zero?
Current Friction:   Why isn't this used more today?
Fazt Collapse:      Can Fazt dramatically reduce the friction?
```
Note: Low current usage doesn't mean low value. High-value + high-friction + low-usage = OPPORTUNITY (e.g., encryption, HTTPS).

### 4. Layer Assignment
- **Kernel**: Core OS primitives (proc, fs, net, storage, security, events)
- **Runtime**: Execution environment (JS engine, sandbox, cron, WASM)
- **Services**: Higher-level utilities (forms, media, pdf, search)

### 5. Cost Assessment
- Binary size impact
- Complexity added to codebase
- Maintenance burden
- Does it require new dependencies?

### Decision Formula
```
Propose only if:
  - True primitive (not pattern)
  - Cannot compose from existing
  - (Intrinsic Value × Friction Collapse) > Complexity Cost
```

## Present Ideas

Generate 2-3 ideas that pass the quality gates. For each:

```
**Idea Name** (Layer: kernel|runtime|services)

What: One sentence description
Primitive: The irreducible capability it adds
Value/Friction: Why this matters (intrinsic value + friction it collapses)
Cost: Complexity and binary impact (low|medium|high)

API Surface (minimal):
- `fazt.namespace.method()`
- `fazt cli command`
```

Keep it brief. Numbers for easy reference.

## Conversation

- User may ask questions about specific ideas
- User may challenge whether something is truly a primitive
- User may say "build idea 2" or combine ideas
- Be willing to say "on reflection, this is a pattern not a primitive"

## When User Approves an Idea

Create the spec file:
1. Determine which version it fits (or propose new one)
2. Create `koder/ideas/specs/v0.X-*/feature-name.md`
3. Update the version's README.md if it exists
4. Update SURFACE.md with new APIs
5. Keep it concise—same style as existing specs

Then say: "Added. Run `/koder-ideate` again when ready for more."
