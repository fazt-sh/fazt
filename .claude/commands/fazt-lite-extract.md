---
description: Evaluate if a library/project can be "lite-extracted" into Fazt
model: opus
allowed-tools: Read, Write, Edit, Glob, Grep, Bash, Task, WebFetch, WebSearch
---

# Lite Extraction Analysis

## The "Lite" Philosophy

Fazt has a pattern of **lite implementations**:

| Full Thing | Fazt Lite | What's Kept | What's Dropped |
|------------|-----------|-------------|----------------|
| IPFS | ipfs-lite | CID addressing, dedup, gateway | DHT, Bitswap, p2p |
| Pinecone/Weaviate | vector-lite | Similarity search, embeddings | HNSW, ANN indexes |
| Kernel WireGuard | wireguard-go | Full protocol | Kernel module |
| LangChain | text/document | Splitting, Document format | Chains, agents, memory |

**The pattern**: Extract 5-20% of features that provide 80%+ of the value for personal-scale use cases.

## Input

The user will provide one of:
- A path to a local project: `~/Projects/go-eino`
- A library name: `chromem-go`
- A concept: `CRDT synchronization`
- A URL: `github.com/example/project`

## Phase 1: Deep Research

### For Local Paths
1. Explore the codebase structure (`tree`, `ls`)
2. Read README, documentation
3. Identify core abstractions (main types, interfaces)
4. Understand the dependency graph
5. Note: Is it pure Go? CGO? External deps?

### For Libraries/Concepts
1. Search the web for documentation, GitHub repos
2. Fetch and read key documentation
3. Understand the problem it solves
4. Identify core vs peripheral features

### Key Questions During Research
- What problem does this solve?
- What are the core primitives vs convenience wrappers?
- What's the minimum viable subset?
- What would break Fazt's constraints (CGO, external deps)?

## Phase 2: Extraction Analysis

### Feature Decomposition

Create a feature matrix:

```
Feature                    | Complexity | Value | Fazt-Compatible
---------------------------|------------|-------|----------------
[core feature 1]           | low/med/hi | hi    | yes/no/maybe
[core feature 2]           | ...        | ...   | ...
[convenience feature 1]    | ...        | ...   | ...
```

### Lite Extraction Candidates

For each potential extraction:

```
Candidate: [name]
From: [source library/concept]

Extract:
- [feature 1]
- [feature 2]

Drop:
- [feature that adds complexity without proportional value]
- [feature that requires CGO or external deps]

Fazt API Shape:
- fazt.namespace.method()
- fazt cli command

Implementation Notes:
- [how it would work in pure Go]
- [what SQLite schema if needed]
- [binary size impact estimate]
```

### Philosophy Check

```
Single binary:    Can this be embedded? [yes/no + explanation]
Pure Go:          Any CGO required? [yes/no + what would need rewriting]
Single database:  Fits in SQLite? [yes/no + schema sketch]
JSON everywhere:  Data as JSON? [yes/no]
Events as spine:  Emits/consumes events? [yes/no]
Personal scale:   Works for <100k items? [yes/no]
```

## Phase 3: Verdict

### GO Verdict
If extraction makes sense:

```
VERDICT: EXTRACT

Name: [fazt-compatible name]
Type: primitive | service | lib
Layer: kernel | runtime | services | lib
Version: v0.X

What: [one sentence]
Why: [value proposition for Fazt users]

Extraction Ratio: [X% of original features]
Value Ratio: [Y% of original value retained]

Next Steps:
1. Create spec at koder/ideas/specs/v0.X-*/[name].md
2. Update SURFACE.md
3. [any prototype/POC suggestions]
```

### NO-GO Verdict
If extraction doesn't make sense:

```
VERDICT: NO-GO

Reason: [primary reason]

Details:
- [specific blocker 1]
- [specific blocker 2]

Alternatives Considered:
- [what else could solve the underlying need]
- [existing Fazt features that partially address this]

Revisit If:
- [conditions under which this might become viable]
```

### DEFER Verdict
If interesting but not ready:

```
VERDICT: DEFER

Reason: [why not now]

Blocking On:
- [prerequisite 1]
- [prerequisite 2]

Track In: [where to note this for future]
```

## Important Guidelines

1. **Be skeptical by default** - Most extractions don't make sense
2. **Validate the need first** - Does Fazt actually need this capability?
3. **Prefer composition** - Can existing primitives achieve this?
4. **Check for prior art** - Is there already a pure Go implementation?
5. **Size matters** - What's the binary size impact?
6. **Never for the sake of it** - Only extract if there's clear value

## Anti-Patterns to Avoid

- Extracting because the library is popular
- Extracting the whole thing (that's not "lite")
- Adding features Fazt users won't use
- Breaking the single-binary promise
- Introducing CGO "just for this one thing"

## Output Format

Present findings conversationally:

1. **What I Found**: Summary of the library/project
2. **Core Capabilities**: What it actually does
3. **Extraction Opportunity**: What subset makes sense (or doesn't)
4. **Verdict**: GO / NO-GO / DEFER with reasoning
5. **If GO**: Proposed spec outline

Be honest if something doesn't fit. "This is cool but doesn't belong in Fazt" is a valid conclusion.
