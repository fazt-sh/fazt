---
description: Brainstorm ideas for Fazt's evolution
allowed-tools: Read, Write, Edit, Glob
---

# Fazt Ideation Session

## Setup

1. Read `koder/ideas/ROADMAP.md` and `koder/ideas/SURFACE.md`
2. Simulate that everything up to v0.16 is implemented
3. Understand the full capability surface

## Present Ideas

Generate 3-4 ideas that extend Fazt. For each:

```
1. **Idea Name**
   - What it does (1 sentence)
   - Key capability it adds
   - What it enables
```

Keep it brief. Numbers for easy reference.

## Conversation

- User may ask questions about specific ideas
- User may combine ideas or add their own twists
- User may say "build idea 2" or "combine 1 and 3 into..."

## When User Approves an Idea

Create the spec file:
1. Determine which version it fits (or propose new one)
2. Create `koder/ideas/specs/v0.X-*/feature-name.md`
3. Update the version's README.md
4. Update SURFACE.md with new APIs
5. Keep it concise—same style as existing specs

Then say: "Added. Run `/koder-ideate` again when ready for more."
