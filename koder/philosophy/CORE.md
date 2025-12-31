# Fazt Philosophy

This document defines what Fazt is, why it exists, and how decisions are made. It is the foundation from which all other design documents derive.

---

## What Fazt Is

Fazt is personal infrastructure for the AI era.

A single binary and a single database that together form a portable, sovereign compute substrate. Deploy it anywhere—laptop, phone, VPS, Raspberry Pi, embedded device. Own everything it contains. Take it with you.

Fazt is the brain for personal digital life: hosting applications, storing data, perceiving environment, reasoning about state, acting on intent. Not a service you rent, but infrastructure you own.

---

## Why Fazt Exists

The current model of personal computing is dependency:
- Your data lives on someone else's servers
- Your applications run on someone else's infrastructure
- Your digital identity is granted by someone else's systems
- Your capabilities are limited by someone else's decisions

Fazt exists to invert this relationship.

You own your infrastructure. You control your data. Your identity is cryptographic, not granted. Your capabilities grow with your needs, not with your subscription tier.

This is not anti-cloud ideology. It is pro-sovereignty pragmatism. Some things belong in the cloud. Many things—especially in an era of AI agents acting on your behalf—should belong to you.

---

## Heritage

Fazt synthesizes lessons from four traditions. Each contributed something essential. None is adopted wholesale.

### Unix (1970s)

**Lesson:** Simplicity through composition.

Unix proved that small tools, doing one thing well, connected by a universal interface (text streams), create systems more powerful than monolithic alternatives.

**What Fazt takes:**
- Small, focused components
- Composition over integration
- Universal interfaces between components

**What Fazt adapts:**
- JSON instead of plain text (structured, typed, LLM-native)
- Events instead of pipes (asynchronous, persistent, queryable)

### Apple (1984–)

**Lesson:** Decisions are a gift.

Apple proved that strong opinions, consistently applied, create experiences that "just work." Users don't want infinite configuration. They want things that work beautifully out of the box.

**What Fazt takes:**
- Opinionated defaults
- Integrated experience
- Craft as a value

**What Fazt adapts:**
- Open source, not proprietary
- Inspectable, not black-boxed
- Community-owned, not corporate-controlled

### Linux (1991–)

**Lesson:** Openness enables trust.

Linux proved that transparency—in code, process, and governance—creates systems people can rely on for decades. What you can inspect, you can trust. What you can modify, you can adapt.

**What Fazt takes:**
- Open source everything
- Transparent operation
- Community participation

**What Fazt adapts:**
- More opinionated than typical Linux projects
- Fewer choices, more coherence
- BDFL decision-making for design consistency

### Bitcoin (2009–)

**Lesson:** Verification beats trust.

Bitcoin proved that cryptographic proof can replace institutional trust. Identity can be mathematical. History can be immutable. Verification can be universal.

**What Fazt takes:**
- Cryptographic identity
- Immutable event history
- Verification over trust

**What Fazt adapts:**
- Not a blockchain (unnecessary for single-owner systems)
- Not decentralized consensus (owner is the authority)
- The trust model, not the implementation

---

## Core Principles

These are not guidelines. They are constraints. Violating them requires amending this document with full justification.

### 1. Single Binary

Fazt is one executable file. No runtime dependencies. No installers. No package managers. Copy the binary, run it, done.

**Why:** Deployment complexity is the enemy of adoption and reliability. A single binary can be verified, distributed, and trusted. Dependencies rot.

**Implication:** Features that require external binaries are either embedded (WASM), deferred (not yet implemented), or delegated (external daemon protocol).

### 2. Single Database

Fazt uses one SQLite database. All state lives there. No Redis. No Postgres. No files scattered across the filesystem.

**Why:** Backup is copying one file. Migration is copying one file. The database is the source of truth, completely and without exception.

**Implication:** Features that need different storage patterns (time-series, full-text search, etc.) are implemented on top of SQLite, not alongside it.

### 3. Pure Go

No CGO. Ever. For any reason.

**Why:** CGO breaks cross-compilation, reproducible builds, and deployment simplicity. It introduces external dependencies that rot and diverge across platforms. The cost is permanent. The benefit is temporary.

**Implication:** Features requiring CGO are either implemented in pure Go (syscalls are acceptable), deferred until pure implementations exist, or delegated to external daemons.

See: `koder/philosophy/SENSORS.md` for extended discussion.

### 4. JSON Everywhere

Every data structure is JSON. Every event payload is JSON. Every API speaks JSON. Every storage format is JSON (in SQLite columns).

**Why:** JSON is the lingua franca of the internet, the native format of LLMs, and the universal interchange format for APIs. A system that speaks JSON can integrate with anything. A system that speaks JSON can be understood by AI.

**Implication:** Binary data is base64-encoded or stored as BLOBs with JSON metadata. Performance-critical paths can use binary internally but expose JSON at boundaries.

### 5. Events as Spine

Components communicate through events. Events are the universal transport, like pipes in Unix.

**Why:** Events enable decoupling (producers don't know consumers), history (events are stored), replay (debugging and recovery), and composition (rules combine events into behaviors).

**Implication:** Direct function calls between major subsystems are discouraged. The event bus is the integration layer.

### 6. Schema Validation

Every JSON structure has a schema. Events are validated on emission. Data is validated at boundaries.

**Why:** Unvalidated data is a source of bugs, security holes, and integration failures. Schemas are documentation that the computer enforces. With schemas, any tool can understand any data.

**Implication:** New event types require schema registration. Schema changes require versioning. Invalid data is rejected, not silently accepted.

### 7. Timestamps are Integers

All timestamps are Unix milliseconds (integer). No ISO strings. No timezone handling in storage.

**Why:** Integer comparison is simple, fast, and unambiguous. Range queries are trivial. No parsing, no timezone bugs, no locale issues.

**Implication:** Display formatting happens at render time, not storage time. Timezones are a presentation concern, not a data concern.

### 8. Cartridge Model

The binary is disposable. The database is precious.

**Why:** Binaries can be downloaded, rebuilt, replaced. Data cannot. This asymmetry should inform all decisions about what goes where.

**Implication:** Configuration lives in the database, not in files. State lives in the database, not in memory. The binary is stateless.

---

## Design Philosophy

These are values, not rules. They guide judgment when principles don't dictate answers.

### Opinionated Over Configurable

When there's a choice, choose. Don't expose the choice to users. One right way is better than ten possible ways.

**Test:** If adding a configuration option, ask: "Is there a correct answer we should just implement?" Usually, yes.

### Craft Over Speed

Quality matters, even in places users won't see. Elegant internals lead to reliable behavior. Hacks compound.

**Test:** Would you be proud to show this code to a peer you respect?

### Patience Over Compromise

Missing features are better than broken features. Wait until you can do it right. Users can work around absence. They cannot work around dysfunction.

**Test:** If you're tempted to ship something "good enough," ask: "Will we ever fix this, or will it become permanent?"

### Simple Over Clever

Simple code is debuggable, maintainable, and trustworthy. Clever code is a liability. Future you is not as smart as present you thinks.

**Test:** Can someone unfamiliar with the codebase understand this in five minutes?

### Explicit Over Magic

No hidden behavior. No action at a distance. If something happens, it should be traceable to a clear cause.

**Test:** When debugging, can you follow cause to effect without guessing?

---

## What Fazt Is Not

Clarity about boundaries prevents scope creep and misaligned expectations.

**Fazt is not a general-purpose cloud platform.** It doesn't try to replace AWS. It serves individuals and small groups who want sovereignty over their infrastructure.

**Fazt is not infinitely configurable.** It has opinions. If those opinions don't fit your needs, Fazt may not be for you. That's okay.

**Fazt is not moving fast and breaking things.** Stability is a feature. APIs don't change capriciously. Upgrades don't destroy data.

**Fazt is not optimizing for every use case.** It optimizes for the AI era: LLM integration, multi-modal awareness, autonomous agents. Traditional web hosting is supported but not the focus.

**Fazt is not a VC-scale startup play.** It doesn't need to be everything to everyone. It needs to be excellent for its intended purpose.

---

## Decision Framework

When evaluating changes—features, refactors, dependencies—apply these questions in order:

### 1. Purpose Alignment
Does this serve Fazt's purpose of personal, sovereign infrastructure for the AI era?

If no: Don't do it.

### 2. Principle Compliance
Does this honor the core principles (single binary, single database, pure Go, JSON, events, schemas, timestamps, cartridge model)?

If no: Either don't do it, or amend the principles with full justification.

### 3. Heritage Consistency
Does this feel consistent with the heritage (Unix simplicity, Apple polish, Linux openness, Bitcoin verification)?

If no: Reconsider the approach.

### 4. Simplicity Check
Is this the simplest solution? Could it be simpler?

If not simplest: Simplify before proceeding.

### 5. Pride Test
Would you be proud to explain this decision to someone you respect?

If no: Reconsider.

---

## Evolution

This document can change. Principles can evolve. But evolution must be explicit.

### Amendment Process

1. **Identify the change:** What principle or value is being modified?
2. **Justify fully:** Why is the current state wrong? Why is the new state right?
3. **Consider consequences:** What else changes if this changes?
4. **Document the amendment:** Add to this document with date and reasoning.
5. **Preserve history:** Previous versions remain accessible.

### What Cannot Change

Some things are definitional. Changing them would make Fazt into something else:

- Single binary, single database (the Cartridge model)
- Personal sovereignty as the purpose
- Open source

Everything else is negotiable with sufficient justification.

---

## Closing

Fazt is an attempt to build infrastructure worthy of trust. Trust is earned through consistency, transparency, and craft. This document exists to maintain that consistency across time and contributors.

When in doubt, return here. When tempted to compromise, return here. When the path is unclear, return here.

The philosophy is the anchor.
