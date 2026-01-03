# The Fazt Vision

*Read CORE.md for principles. Read EVOLUTION.md for history.
Read this for where we're going.*

---

## The One-Sentence Version

**Fazt is sovereign compute infrastructure for individuals—a personal operating
system that runs everywhere you need computation, from your phone to your
washing machine.**

---

## What Fazt Is Becoming

Fazt began as a PaaS (deploy static sites to your VPS). It evolved into
something much larger:

**A universal compute substrate for personal life.**

| What It Was       | What It Is                        |
|-------------------|-----------------------------------|
| Web server        | Operating system                  |
| Host sites        | Run any compute/data task         |
| One VPS           | Swarm of devices                  |
| Cloud alternative | Sovereign infrastructure          |
| Developer tool    | AI-native platform anyone can use |

The core insight: In an AI era, individuals need compute infrastructure
as much as enterprises do. But they need it on *their* terms—private,
portable, resilient, owned.

---

## The Swarm Model

A mature Fazt deployment is not one node. It's many:

```
┌─────────────────────────────────────────────────────────────┐
│                    YOUR FAZT SWARM                          │
│                                                             │
│   ┌──────────┐   ┌──────────┐   ┌──────────┐               │
│   │  Laptop  │   │  Phone   │   │   VPS    │               │
│   │  (dev)   │   │ (mobile) │   │ (public) │               │
│   └────┬─────┘   └────┬─────┘   └────┬─────┘               │
│        │              │              │                      │
│        └──────────────┼──────────────┘                      │
│                       │                                     │
│              ┌────────┴────────┐                            │
│              │   Secure Mesh   │                            │
│              └────────┬────────┘                            │
│                       │                                     │
│        ┌──────────────┼──────────────┐                      │
│        │              │              │                      │
│   ┌────┴─────┐   ┌────┴─────┐   ┌────┴─────┐               │
│   │  Pi w/   │   │ Security │   │  Smart   │               │
│   │  Camera  │   │  System  │   │ Appliance│               │
│   └──────────┘   └──────────┘   └──────────┘               │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

Each node:
- Runs the same single binary
- Has its own SQLite database (cartridge)
- Meshes with other nodes securely
- Contributes sensors, compute, storage
- Works autonomously when isolated
- Syncs when connected

**Your swarm is your personal cloud.** Not rented. Owned.

---

## The Resilience Imperative

Fazt nodes must work when:
- Network is slow
- Network is unreliable
- Network is unavailable
- Network is *denied*

This isn't paranoia. It's reality:
- Travel through connectivity dead zones
- Infrastructure outages (storms, disasters)
- Adversarial network conditions
- Air-gapped environments
- Edge deployments without backhaul

**Design constraint:** Every Fazt node must be useful in isolation.
Mesh connectivity enhances capability. It never becomes a requirement.

This is why Fazt has:
- **Beacon**: mDNS discovery on local network
- **Chirp**: Data transfer via audio (when all else fails)
- **Mnemonic**: Human-channel key exchange
- **Timekeeper**: Time consensus without NTP

Extreme? Maybe. But "works without internet" is a feature that matters
when it matters.

---

## AI Native

"AI native" means two things:

### 1. AI Lowers the Floor

Traditional infrastructure requires expertise:
- Command lines
- Configuration files
- Networking knowledge
- System administration

Fazt + AI agents = anyone can use it:
- Natural language commands
- Intelligent defaults
- Self-healing systems
- Explained errors

**Goal:** Your non-technical family member can run their own Fazt node.

### 2. AI Raises the Ceiling

For those who push further:
- Harness apps (self-modifying code)
- MCP integration (agents use Fazt tools)
- RAG pipelines (knowledge retrieval)
- Perception (understand environment via VLM)
- Autonomous action (effect layer)

**Goal:** An AI researcher can build sophisticated agentic systems on Fazt.

The floor and ceiling move in opposite directions. That's the power of
AI-native design.

---

## The Device Spectrum

Fazt runs on anything that runs Go:

- **Phone**: Mobile presence (notifications, location, personal data)
- **Laptop**: Daily driver (apps, files, local AI)
- **Raspberry Pi**: Sensor node (cameras, environment)
- **VPS**: Public endpoint (web apps, API gateway)
- **NAS**: Storage node (bulk storage, backup)
- **Smart appliance**: Embedded (washing machine, thermostat)
- **Vehicle**: Mobile compute (GPS, dashcam, trips)
- **Security system**: Always-on (cameras, motion, alerting)

Same binary everywhere. Same mental model.
Different capabilities based on hardware.

---

## Universal Compute Substrate

What can Fazt do? Almost any compute/data task an individual needs:

**Hosting**
- Static sites
- Dynamic apps
- APIs
- Serverless functions

**Data**
- KV store
- Document store
- Vector store (embeddings)
- Blob storage
- Full-text search

**AI**
- LLM integration (multiple providers)
- RAG pipelines
- Vision (VLM)
- Agents (MCP server)

**Services**
- Email (send/receive)
- Notifications (push, SMS, webhooks)
- Forms
- Comments
- Short URLs
- PDF generation
- Image processing

**Perception** (sensor layer)
- Temperature, humidity, pressure
- Camera feeds
- GPS
- Audio
- Motion

**Action** (effect layer)
- Control devices
- Send alerts
- Call APIs
- Physical actuation

**Coordination**
- Cron scheduling
- Background jobs
- Event processing
- Multi-node distribution

**Security**
- Cryptographic identity
- Secrets vault
- OAuth provider
- VPN (WireGuard)

This is not feature creep. This is recognizing that "personal compute"
spans all these domains—and they should compose rather than requiring
separate tools.

---

## Why Single Binary + Single DB

The cartridge model isn't a technical curiosity. It's strategic:

**Portability**: Copy one file, run anywhere.

**Backup**: Copy one file, have everything.

**Migration**: Copy one file, move completely.

**Trust**: One thing to verify, one thing to secure.

**Resilience**: No external dependencies to fail.

**Simplicity**: No "which config file" or "which database".

The entire state of your personal compute infrastructure is one SQLite
file. The entire runtime is one binary. This isn't minimalism for
aesthetics. It's minimalism for reliability.

---

## The Competition

Fazt competes with:

- **PaaS** (Coolify, CapRover): Fazt is broader, not just hosting
- **Home automation** (Home Assistant): Fazt unifies compute/data/AI
- **Personal cloud** (Nextcloud): Fazt is single binary, not PHP stack
- **IoT platforms**: Fazt is pure Go, not proprietary
- **AI frameworks** (LangChain): Fazt is infrastructure, not library

Fazt's unique position: **The unifying layer.**

Instead of:
- Home Assistant for automation
- Nextcloud for files
- Self-hosted email
- Separate AI tools
- Disconnected IoT devices

One system. One mental model. One backup.

---

## The 10-Year Vision

**Year 1-2 (current):** Core functionality. Hosting, storage, AI, basic
sensors. Single-node focus with mesh foundations.

**Year 3-5:** Swarm maturity. Multi-node coordination becomes natural.
Sensor coverage expands. Effect layer enables real-world actions.
Community app ecosystem.

**Year 5-10:** Fazt becomes infrastructure you assume exists, like
electricity. Your devices run Fazt. Your home runs Fazt. Your data lives
in your swarm. AI agents act on your behalf through infrastructure you own.

This isn't about building a company. It's about building infrastructure
worthy of trust for decades.

---

## Principles (Abbreviated)

From CORE.md, the non-negotiables:

1. **Single Binary**: One executable, no dependencies
2. **Single Database**: One SQLite file, all state
3. **Pure Go**: No CGO, ever
4. **JSON Everywhere**: Universal data format
5. **Events as Spine**: Decoupled composition
6. **Cartridge Model**: Disposable binary, precious data

These aren't preferences. They're the physics of the system. Violating
them changes what Fazt is.

---

## Who This Is For

**Individuals who want sovereignty.**

Not everyone wants to run their own infrastructure. That's fine. Cloud
services exist.

But for those who:
- Care about privacy
- Want to own their data
- Distrust centralized platforms
- Need to work offline
- Want AI agents under their control
- Like understanding their systems
- Value long-term stability over features

Fazt is for you.

---

## Closing

Fazt is ambitious. A single binary that does all this? A swarm of nodes
managing your entire compute life? Resilient to network denial?

Yes. That's the goal.

Not because it's easy. Because it's necessary.

The future is personal AI infrastructure. Either you own it, or someone
else does. Fazt is the "own it" option.
