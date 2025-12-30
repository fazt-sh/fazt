# MCP Integration Discussion Handoff

This document captures the deep discussion about integrating MCP (Model Context Protocol) into Fazt, including the vision, architecture, and identified primitives.

---

## The Vision

Fazt should evolve from a **passive toolbox** to a **living digital entity**:

```
Today's Fazt:
  User → CLI → Fazt → serves apps
  (passive toolbox)

The Vision:
  User ↔ Claude ↔ Fazt ↔ World
         ↑            ↓
         └── pulse ←──┘
  (living entity with agency)
```

### Core Characteristics of the Living Entity

1. **Controllable via natural language** - MCP server exposes all operations
2. **Self-aware** - Pulse monitors health, resources, trends
3. **Autonomous** - Agent acts on observations without human intervention
4. **Prioritizes longevity** - Renews certs, cleans up, alerts on issues
5. **Builds and manages apps** - Full lifecycle from idea to production
6. **Has personality** - Configurable communication style

---

## What is MCP?

MCP (Model Context Protocol) is a protocol for connecting LLMs to external systems. It provides:

- **Tools**: Functions the LLM can call (e.g., `deploy`, `apps.list`)
- **Resources**: Data the LLM can read (e.g., `fazt://logs/app-name`)
- **Prompts**: Templates for common operations

The reference implementation is at `~/Projects/go-mcp/`.

### Key Types from go-mcp

```go
// Server exposes tools, prompts, resources
server := mcp.NewServer(&mcp.Implementation{
    Name: "fazt",
    Version: "0.12.0",
}, nil)

// Add tools with typed handlers
mcp.AddTool(server, &mcp.Tool{
    Name: "deploy",
    Description: "Deploy an app",
    InputSchema: deployArgsSchema,
}, deployHandler)

// Transport options
server.Run(ctx, &mcp.StdioTransport{})  // For Claude Desktop
// Or HTTP+SSE for remote access
```

---

## What MCP Provides for Fazt

### Example Conversations

**Deploying an app:**
```
User: "Launch a photography portfolio at photos.example.com"

Claude (via Fazt MCP):
1. infra.vps.create() → Spins up Hetzner server
2. infra.dns.set() → Points domain to new IP
3. apps.create() → Creates the app
4. deploy() → Deploys files
→ "Your site is live at https://photos.example.com"
```

**Checking health:**
```
User: "How is my server doing?"

Claude (via Fazt MCP):
1. pulse.status() → Gets health data
→ Natural language summary of CPU, memory, disk, recent errors
```

**Querying logs:**
```
User: "Show me errors from the shop app in the last hour"

Claude (via Fazt MCP):
1. logs.query({ app: 'shop', level: 'error', since: '1h' })
→ Formatted error logs with context
```

**Full lifecycle:**
```
User: "I want to launch a blog for my photography"

Claude:
1. Checks if VPS exists → infra.vps.create() if needed
2. Checks DNS → infra.dns.set() to point domain
3. Creates app → apps.create({ template: 'blog' })
4. Deploys → deploy(files)
5. Provisions SSL → net.ssl.provision()
→ "Your blog is live at https://photos.example.com"
```

---

## Architecture

### MCP Layers

```
┌─────────────────────────────────────────────────────────────────────┐
│  LAYER: MCP INTERFACE                                               │
│                                                                     │
│  ┌───────────────────────────────────────────────────────────────┐ │
│  │  mcp.kernel                                                    │ │
│  │  - System-wide MCP server                                     │ │
│  │  - Exposes ALL Fazt operations as tools                       │ │
│  │  - Owner authentication required                               │ │
│  │  - Used by Claude Desktop, Claude Code                         │ │
│  └───────────────────────────────────────────────────────────────┘ │
│                                                                     │
│  ┌───────────────────────────────────────────────────────────────┐ │
│  │  mcp.app                                                       │ │
│  │  - Per-app MCP endpoints                                       │ │
│  │  - Apps register their own tools/resources                     │ │
│  │  - Scoped to app permissions                                   │ │
│  │  - Can be exposed to app users                                 │ │
│  └───────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
```

### Full System Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         FAZT INSTANCE                               │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  MCP.KERNEL (nervous system)                                 │   │
│  │  Exposes all operations as tools for external AI             │   │
│  └──────────────────────────┬──────────────────────────────────┘   │
│                              │                                      │
│  ┌───────────┐  ┌───────────┴───────────┐  ┌───────────┐          │
│  │   PULSE   │  │        AGENT          │  │   INFRA   │          │
│  │           │  │                       │  │           │          │
│  │  observe  │─►│  decide + act         │─►│  create   │          │
│  │  health   │  │  autonomously         │  │  servers  │          │
│  └───────────┘  └───────────────────────┘  └───────────┘          │
│                              │                                      │
│                              ▼                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  KERNEL (proc, fs, net, storage, security, events, dev)     │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### The Autonomy Loop

```
OBSERVE (pulse) → DECIDE (ai) → ACT (kernel) → LEARN (events)
     ↑                                              │
     └──────────────────────────────────────────────┘
```

---

## Primitives Inventory

### Already Spec'd

| Primitive | Location | Role |
|-----------|----------|------|
| **infra** | v0.8 | VPS, DNS, Domain provisioning |
| **pulse** | v0.8 | Self-awareness, health monitoring |
| **devices** | v0.8 | External service abstraction |
| **events** | v0.8 | Internal communication |
| **ai shim** | v0.12 | LLM integration |
| **mcp** | v0.12 | Basic MCP server |

### Needs Refinement

| Primitive | Current State | Needed |
|-----------|---------------|--------|
| **mcp.kernel** | Basic in v0.12 | Full tool surface, auth, resources |
| **mcp.app** | Mentioned in v0.12 | Per-app tool registration API |

### Missing (NEW)

| Primitive | Purpose |
|-----------|---------|
| **agent** | Autonomy loop - the "soul" of the living entity |

---

## The Agent Primitive (Proposed)

The agent is what transforms Fazt from "controllable" to "autonomous".

### Without Agent
- Fazt is controllable via MCP (passive)
- You tell it what to do, it does it
- It reports health via pulse, but doesn't act on it

### With Agent
- Fazt is autonomous (active)
- It observes its own state (pulse)
- It decides what to do (ai shim)
- It acts on its own behalf (kernel)
- It learns from outcomes (events)

### Agent Structure

```
agent
├── config
│   ├── autonomy: low | medium | high
│   ├── personality: helpful | terse | cautious
│   └── notifications: email | slack | sms | none
│
├── loop
│   ├── subscribe to pulse beats
│   ├── evaluate alerts against policy
│   ├── decide action (via ai shim)
│   ├── execute or escalate (halt)
│   └── log decision for learning
│
└── memory
    ├── decision history
    ├── owner preferences (learned)
    └── escalation patterns
```

### Autonomy Levels

| Level | Behavior |
|-------|----------|
| **low** | Only notify, never act autonomously |
| **medium** | Act on safe operations (renew certs, cleanup), escalate destructive |
| **high** | Act on everything except payments and deletion |

### Example: Autonomous SSL Renewal

```
pulse detects: "SSL cert expires in 3 days"
     │
     ▼
agent evaluates: "SSL renewal is safe, autonomy=medium allows this"
     │
     ▼
agent acts: fazt.net.ssl.renew("blog.example.com")
     │
     ▼
agent notifies owner: "Renewed SSL for blog.example.com"
     │
     ▼
agent logs: { action: "ssl.renew", trigger: "pulse.ssl_expiring", outcome: "success" }
```

### Example: Escalation for Destructive Operation

```
pulse detects: "Disk 95% full, old logs consuming 20GB"
     │
     ▼
agent evaluates: "Log deletion is destructive, autonomy=medium requires escalation"
     │
     ▼
agent calls: fazt.halt({
    reason: "Disk 95% full. Delete 20GB of logs older than 30 days?",
    options: ["Delete logs", "Expand disk", "Ignore"]
})
     │
     ▼
owner responds via notification channel
     │
     ▼
agent acts based on response
```

---

## MCP Tool Surface

### Kernel Tools (mcp.kernel)

These are exposed to Claude Desktop and other MCP clients:

```
# App Management
apps.list                    List all apps
apps.get { id }              Get app details
apps.create { name, ... }    Create new app
apps.delete { id }           Delete app (destructive, triggers halt)

# Deployment
deploy { app, files }        Deploy files to app
deploy.status { app }        Get deployment status
deploy.rollback { app }      Rollback to previous version

# Infrastructure
infra.vps.list               List managed VPSes
infra.vps.create { ... }     Create VPS (triggers halt for cost)
infra.vps.status { id }      Get VPS health
infra.vps.destroy { id }     Destroy VPS (destructive, triggers halt)
infra.dns.set { ... }        Create/update DNS record
infra.dns.records { zone }   List DNS records

# Storage
storage.kv.get { key }       Get KV value
storage.kv.set { key, val }  Set KV value
storage.ds.query { ... }     Query document store

# Observability
pulse.status                 Current health assessment
pulse.ask { question }       Natural language query about system
pulse.history { hours }      Past health beats
logs.query { app, ... }      Query app logs

# Events
events.list { ... }          Query events
events.emit { type, data }   Emit event
```

### Kernel Resources (mcp.kernel)

Read-only data access for LLM context:

```
fazt://apps                      List of all apps
fazt://apps/{id}                 App details
fazt://apps/{id}/files/{path}    File contents
fazt://apps/{id}/logs            Recent logs
fazt://config                    System configuration
fazt://pulse/status              Current health
fazt://pulse/history             Health history
fazt://infra/vps                 Managed VPSes
fazt://infra/dns/{zone}          DNS records
```

### App Tools (mcp.app)

Apps can register their own tools:

```javascript
// In app's main.js
fazt.mcp.tool("posts.create", {
    description: "Create a new blog post",
    input: { title: "string", body: "string" },
    handler: async (args) => {
        const id = await fazt.storage.ds.insert("posts", args);
        return { id };
    }
});

fazt.mcp.tool("posts.list", {
    description: "List all blog posts",
    input: { limit: "number?" },
    handler: async (args) => {
        return fazt.storage.ds.query("posts", { limit: args.limit || 10 });
    }
});

fazt.mcp.resource("posts://{id}", {
    description: "Get a specific blog post",
    handler: async (uri, params) => {
        return fazt.storage.ds.get("posts", params.id);
    }
});
```

---

## Capability Gating

MCP sessions carry capability tokens:

```go
type MCPSession struct {
    Capabilities []string  // ["kernel:*", "app:blog:read"]
    Owner        bool
    AppID        string    // If scoped to an app
}

type Tool struct {
    Name     string
    Requires []string  // ["kernel:apps:write"]
    Handler  func(...)
}
```

### Capability Hierarchy

```
kernel:*                    Full kernel access (owner only)
kernel:apps:read            List/get apps
kernel:apps:write           Create/delete apps
kernel:storage:*            All storage ops
kernel:infra:*              VPS/DNS management
kernel:pulse:read           Query pulse
app:{uuid}:*                Full app access
app:{uuid}:read             Read app data only
```

---

## Transport Options

### Stdio (for Claude Desktop)

```bash
# Claude Desktop config
{
    "mcpServers": {
        "fazt": {
            "command": "fazt",
            "args": ["mcp", "start", "--stdio"]
        }
    }
}
```

### HTTP + SSE (for remote access)

```bash
# Start MCP server on port
fazt mcp start --port 3001 --token xxx

# Connect from remote
# Uses HTTP POST for requests, SSE for notifications
```

---

## Security Considerations

### Authentication

- **mcp.kernel**: Requires owner API key + optional TOTP
- **mcp.app**: Follows app's auth configuration

### Destructive Operations

Operations that cost money or destroy data trigger `halt()`:

```go
func (m *MCP) InfraVPSDestroy(id string) error {
    confirmed, err := m.kernel.Halt(HaltRequest{
        Reason: "Destroy server?",
        Data:   server,
    })
    if !confirmed {
        return ErrUserCanceled
    }
    return m.infra.VPSDestroy(id)
}
```

### Audit Trail

All MCP operations logged:

```go
m.events.Emit("mcp.call", MCPEvent{
    Tool:      "apps.delete",
    Input:     sanitize(input),
    Output:    sanitize(output),
    Session:   session.ID,
    Timestamp: time.Now(),
})
```

---

## Why MCP is Important

### 1. Accessibility
Non-technical users can manage cloud infrastructure through natural language.

### 2. AI-Native Platform
Fazt becomes a first-class citizen in the AI ecosystem. Any MCP client can control it.

### 3. Compound Automation
Single commands trigger complex multi-step workflows.

### 4. The Living Entity
MCP is what transforms Fazt from a tool into an entity you can converse with.

---

## What Makes It Valuable

**Comparison with other platforms:**

| Platform | Interaction | Autonomy |
|----------|-------------|----------|
| Heroku | Push to deploy | None |
| Vercel | Git-triggered | None |
| Cloudflare | Dashboard/CLI | None |
| **Fazt + MCP + Agent** | Conversational | Configurable |

The combination creates something unique:
- A personal cloud you can **talk to**
- That **manages itself**
- That **grows with you**
- That **learns your preferences**

It's the difference between a **tool** and a **partner**.

---

## Next Steps

1. **Refine mcp.kernel spec** - Full tool surface, auth, resources
2. **Spec mcp.app** - Per-app tool registration API
3. **Spec agent primitive** - Autonomy loop configuration and behavior
4. **Implementation** - Using go-mcp reference

---

## Questions to Resolve

1. Where does agent fit in the version roadmap? (v0.12 or new version?)
2. Should agent configuration be per-instance or per-app?
3. How does multi-node Fazt work with MCP? (mesh + MCP?)
4. What's the notification channel for agent escalations?

---

## Reference

- go-mcp implementation: `~/Projects/go-mcp/`
- Existing MCP spec: `koder/ideas/specs/v0.12-agentic/mcp.md`
- Pulse spec: `koder/ideas/specs/v0.8-kernel/pulse.md`
- Infra spec: `koder/ideas/specs/v0.8-kernel/infra.md`
