# Lite Extraction Log

Projects evaluated for potential "lite" extraction into Fazt.

The "lite" pattern: Extract 5-20% of features that provide 80%+ of value
for personal-scale use cases. Examples: ipfs-lite, vector-lite, wireguard-go.

## Evaluation Criteria

- **Single binary**: Can it be embedded without external deps?
- **Pure Go**: No CGO required?
- **SQLite fit**: Data model works with single DB?
- **Personal scale**: Optimized for <100k items?
- **Composable**: Works with existing Fazt primitives?

## Verdicts

- **EXTRACT**: Code worth pulling into Fazt
- **USE-AS-IS**: Import whole library - already optimal for embedding
- **NO-GO**: Doesn't fit, won't revisit soon
- **DEFER**: Interesting but blocked on prerequisites
- **PATTERN**: No code to extract, but valuable design pattern

## Log

| Date       | Project       | Verdict   | Reason                           |
|------------|---------------|-----------|----------------------------------|
| 2026-01-03 | backlite      | USE-AS-IS | SQLite task queues, perfect fit  |
| 2026-01-03 | gortsplib     | USE-AS-IS | RTSP client/server for IP cameras|
| 2026-01-03 | mochi-mqtt    | USE-AS-IS | Embeddable MQTT broker for IoT   |
| 2026-01-03 | peerdiscovery | USE-AS-IS | UDP multicast LAN peer discovery |
| 2026-01-03 | pion/webrtc   | USE-AS-IS | P2P, NAT traversal, data channels|
| 2026-01-03 | bluemonday    | USE-AS-IS | HTML sanitization, security-safe |
| 2026-01-03 | gjson         | USE-AS-IS | Zero-dep JSON path queries       |
| 2026-01-03 | goldmark      | USE-AS-IS | Zero-dep markdown parser         |
| 2026-01-03 | go-expr       | USE-AS-IS | Zero-dep expression engine       |
| 2026-01-03 | go-redka      | PATTERN   | Redis-on-SQLite schema patterns  |
| 2026-01-03 | go-crush      | PATTERN   | Skills, permissions, LSP routing |
| 2026-01-03 | go-lingoose   | EXTRACT   | Text splitter + cosine algos     |
| 2026-01-03 | go-adk        | PATTERN   | State prefix pattern worth study |
| 2026-01-03 | go-eino       | NO-GO     | Framework vs library mismatch    |
| 2026-01-03 | alice         | NO-GO     | ~40 lines, not needed            |
| 2026-01-03 | atreugo       | NO-GO     | Fasthttp framework, use stdlib   |
| 2026-01-03 | easyprobe     | NO-GO     | Standalone app, not library      |
| 2026-01-03 | gorm          | NO-GO     | ORM overhead, Fazt uses raw SQL  |
| 2026-01-03 | htmgo         | NO-GO     | Framework, Fazt has React admin  |
| 2026-01-03 | langchaingo   | NO-GO     | Framework vs library mismatch    |
| 2026-01-03 | localai       | NO-GO     | CGO deps (llama.cpp), not embed  |
| 2026-01-03 | upper/db      | NO-GO     | Multi-DB abstraction, SQLite-only|
| 2026-01-03 | periph.io     | SKIP      | Already in SENSORS.md spec       |
| 2026-01-03 | mcp-go-sdk    | SKIP      | Already referenced, planned      |

## Note: Patterns vs Code

Patterns (architectural approaches, designs) are **always extractable** -
ideas aren't copyrightable. Even from restrictively-licensed projects like
go-crush (FSL), we document patterns for independent implementation.

If you find a **permissive Go project** with similar agentic patterns,
evaluate it - reference code accelerates implementation even when patterns
are already documented.

---

### backlite

- **URL**: https://github.com/mikestefanello/backlite
- **What**: Type-safe, persistent task queues with SQLite backend
- **Verdict**: USE-AS-IS
- **License**: MIT

**Why USE-AS-IS**: Already perfectly designed for Fazt's cartridge model:
- Pure Go core (CGO only in tests, driver-agnostic)
- Designed specifically for SQLite embedding
- ~1200 lines, minimal binary impact (~50KB)
- Type-safe generics, worker pool, retry/backoff
- Transaction support (add tasks in your app's transaction!)

```go
// Define task type
type SendEmail struct {
    To      string
    Subject string
}

func (t SendEmail) Config() backlite.QueueConfig {
    return backlite.QueueConfig{
        Name:        "email",
        MaxAttempts: 3,
        Backoff:     5 * time.Second,
    }
}

// Register and use
queue := backlite.NewQueue[SendEmail](func(ctx context.Context, t SendEmail) error {
    return email.Send(t.To, t.Subject)
})
client.Register(queue)

// Add tasks (even in transactions!)
client.Add(SendEmail{To: "user@example.com"}).Tx(tx).Wait(5 * time.Minute).Save()
```

**Schema** (2 tables, perfect for cartridge):
```sql
backlite_tasks (id, queue, task BLOB, wait_until, attempts, ...)
backlite_tasks_completed (id, queue, succeeded, error, expires_at, ...)
```

**Key features**:
- No database polling (notification pattern)
- Graceful shutdown (wait for workers)
- Panic recovery (auto-marks failed)
- Retention policies (keep failed tasks for debugging)
- Optional web UI (skip - Fazt has admin)

**Fazt applicability**: HIGH - Runtime layer
- Queue notification emails during deployments
- Schedule background cleanup tasks
- Process webhooks asynchronously
- Any "do this later" pattern

---

### gortsplib

- **URL**: https://github.com/bluenviron/gortsplib
- **What**: Pure Go RTSP client/server for IP cameras
- **Verdict**: USE-AS-IS
- **License**: MIT

**Why USE-AS-IS**: THE standard for RTSP in Go:
- Pure Go (~23k lines, no CGO)
- Written for MediaMTX (popular media server)
- IP cameras speak RTSP natively
- pion/* deps overlap with pion/webrtc (no extra cost)

**Features**:
- Client: read streams from IP cameras (UDP, TCP, multicast)
- Server: re-stream to multiple clients
- RTSPS/TLS, RTSP-over-HTTP, RTSP-over-WebSocket
- Auto transport protocol switching

**Binary impact**: ~800KB (with mediacommon codec lib)

```go
// Connect to IP camera
client := gortsplib.Client{}
client.Start("rtsp://camera.local:554/stream")

// Read frames
client.OnPacketRTP = func(medi *media.Media, pkt *rtp.Packet) {
    // Decode frame, send to VLM for perception
    fazt.events.Emit("sensor.camera.frame", frame)
}
```

**Fazt applicability**: HIGH - Sensors layer
- Read IP camera streams (RTSP is standard protocol)
- Pipe frames to VLM for perception (Percept layer)
- Re-stream to browsers via WebRTC
- Events: `sensor.camera.frame`

---

### mochi-mqtt

- **URL**: https://github.com/mochi-mqtt/server
- **What**: Embeddable MQTT v5 broker for IoT/telemetry
- **Verdict**: USE-AS-IS
- **License**: MIT

**Why USE-AS-IS**: Standard IoT protocol, designed for embedding:
- Pure Go core (~11k lines, no CGO)
- Full MQTT v5 + v3.1.1 compatibility
- Heavy deps (badger, redis, etc.) are OPTIONAL hooks
- Hook system allows custom SQLite persistence

**Dependency architecture**:
```
Core (always):     stdlib + gorilla/websocket + rs/xid
Optional hooks:    badger, bolt, pebble, redis (don't import = no deps)
```

**Binary impact**: ~500KB (core only)

```go
// Embed MQTT broker in Fazt
server := mqtt.New(nil)
server.AddHook(new(FaztSqliteHook), nil)  // Custom storage
server.AddListener(listeners.NewTCP("tcp", ":1883", nil))
server.Serve()

// Sensors publish to: sensors/{node_id}/{type}
// Payload: {"value": 22.5, "unit": "celsius"}
```

**Fazt applicability**: HIGH - Sensors layer
- Standard MQTT protocol (sensors speak this)
- Sensors publish readings, Fazt subscribes
- Write SQLite storage hook for cartridge persistence
- Events: `mqtt.message`, `mqtt.client.connected`

---

### peerdiscovery

- **URL**: https://github.com/schollz/peerdiscovery
- **What**: Pure-Go UDP multicast local peer discovery
- **Verdict**: USE-AS-IS
- **License**: MIT

**Why USE-AS-IS**: Exactly what Fazt's Beacon needs for mesh discovery:
- Pure Go (~450 lines, uses golang.org/x/net for multicast)
- Custom payloads (for node identity, capabilities, load)
- Callbacks: `Notify` (peer found), `NotifyLost` (peer gone)
- IPv4 + IPv6 dual-stack support
- Peer garbage collection built-in

```go
// Long-running discovery with callbacks
pd, _ := peerdiscovery.NewPeerDiscovery(peerdiscovery.Settings{
    Payload: []byte(`{"node":"fazt-abc123","caps":["hosting","ai"]}`),
    Notify: func(d peerdiscovery.Discovered) {
        fmt.Printf("Found: %s\n", d.Address)
    },
    NotifyLost: func(lp peerdiscovery.LostPeer) {
        fmt.Printf("Lost: %s\n", lp.Address)
    },
})
```

**How it works**:
1. Broadcast on UDP multicast (default 239.255.255.250:9999)
2. Listen for other peers broadcasting
3. Exchange custom payloads (JSON node info)
4. GC removes peers not seen recently

**Fazt applicability**: HIGH - Mesh layer (Beacon)
- Discover other Fazt nodes on local network
- Exchange node identity + capabilities
- Feed discovered peers into mesh topology
- Events: `beacon.peer.found`, `beacon.peer.lost`

---

### pion/webrtc

- **URL**: https://github.com/pion/webrtc
- **What**: Pure Go WebRTC implementation (P2P, NAT traversal)
- **Verdict**: USE-AS-IS
- **License**: MIT

**Why USE-AS-IS**: THE standard for P2P in Go, essential for mesh vision:
- Pure Go (~80k lines across pion/* ecosystem)
- NAT traversal (ICE) - critical for WAN mesh
- Browser ↔ Fazt node real-time communication
- Encrypted P2P data channels
- Audio/video streaming capability

**Binary impact**: ~3-4MB (acceptable for platform scope)

**Pion is modular** - can import pieces:
- `pion/ice` (~500KB) - just NAT traversal
- `pion/datachannel` - just data channels
- `pion/stun` - just STUN protocol

**Fazt applicability**: HIGH - Mesh + Sensors + Admin
- Mesh: P2P connectivity across NAT boundaries
- Sensors: Video streaming to browsers
- Admin: Real-time browser ↔ node communication
- Chirp: Audio data channel

---

### bluemonday

- **URL**: https://github.com/microcosm-cc/bluemonday
- **What**: HTML sanitization library
- **Verdict**: USE-AS-IS
- **License**: BSD 3-Clause

**Why USE-AS-IS**: HTML sanitization is security-critical and easy to get wrong.
Bluemonday is the Go standard, battle-tested, whitelist-based approach.

```go
p := bluemonday.UGCPolicy()  // User-generated content preset
safe := p.Sanitize(untrustedHTML)
```

**Fazt applicability**: HIGH if hosting user content - comments, form submissions,
markdown with embedded HTML. Security-critical, don't reinvent.

---

### gjson

- **URL**: https://github.com/tidwall/gjson
- **What**: Fast JSON path queries without unmarshaling
- **Verdict**: USE-AS-IS
- **License**: MIT

**Why USE-AS-IS**: Zero deps, ~2k lines. Query JSON blobs without full parse.

```go
value := gjson.Get(json, "name.last")        // Simple path
value := gjson.Get(json, "friends.#.name")   // Array iteration
value := gjson.Get(json, `friends.#(age>40).name`)  // Filtering
```

**Fazt applicability**: HIGH - serverless functions querying large JSON in SQLite
without memory overhead of full unmarshal.

---

### goldmark

- **URL**: https://github.com/yuin/goldmark
- **What**: CommonMark compliant markdown parser
- **Verdict**: USE-AS-IS
- **License**: MIT

**Why USE-AS-IS**: Zero deps, highly extensible, standard Go choice for markdown.

**Fazt applicability**: MEDIUM - auto-render .md files in VFS as HTML pages.
Could power documentation sites or blog-style deployments.

---

### go-expr

- **URL**: https://github.com/expr-lang/expr
- **What**: Safe, fast expression language for Go
- **Verdict**: USE-AS-IS
- **License**: MIT

**Why USE-AS-IS**: Already perfectly designed for embedding:
- Zero dependencies (stdlib only)
- ~31k lines but only ~200KB binary impact
- `DisableAllBuiltins()` + `EnableBuiltin()` for lite mode
- Compile-once/run-many pattern for performance
- Static type checking catches errors before runtime

**Used by**: Google, Uber, Argo, OpenTelemetry, CrowdSec, CoreDNS

---

#### Built-in Functions (All Disableable)

```
Collection: all, none, any, one, filter, map, find, count, sum, groupBy, sortBy, reduce
Math:       abs, ceil, floor, round, int, float, max, min, mean, median
String:     len, trim, upper, lower, split, replace, join, hasPrefix, hasSuffix
Encoding:   toJSON, fromJSON, toBase64, fromBase64
Time:       now, duration, date, timezone
```

---

#### API Pattern

```go
// Define environment struct (type-safe)
type Env struct {
    User    User
    Request Request
}

// Compile with type checking (once at startup)
program, err := expr.Compile(
    `User.Role in ["admin", "mod"] || Request.IP in Whitelist`,
    expr.Env(Env{}),
    expr.AsBool(),
)

// Run per-request (fast bytecode VM)
allowed, err := expr.Run(program, env)
```

---

#### Fazt Use Cases

1. **Dynamic routing rules**:
   ```json
   {"pattern": "/admin/*", "expr": "user.Role == \"admin\""}
   ```

2. **Webhook filters**:
   ```json
   {"event": "push", "expr": "len(commits) > 0 && branch == \"main\""}
   ```

3. **Access control**:
   ```json
   {"resource": "billing", "expr": "user.Plan in [\"pro\", \"team\"]"}
   ```

**Fazt applicability**: HIGH - enables user-defined logic without JS runtime
overhead. Compile rules at config load, evaluate per-request.

---

### go-redka

- **URL**: https://github.com/nalgeon/redka
- **What**: Redis reimplementation using SQLite/PostgreSQL backend
- **Verdict**: PATTERN
- **License**: BSD 3-Clause (permissive)

**Why PATTERN not EXTRACT**: Despite permissive license, Fazt already has KV
via chromem-go. The value is in schema design for SQL-backed data structures,
not code extraction.

---

#### Pattern 1: Unified Key Metadata Table

**Problem**: Store multiple data types (strings, lists, sets, hashes) in SQL.

**Solution**:
```sql
-- Single table for key metadata
rkey (id, key, type, version, etime, mtime, len)
  ↓ foreign key (kid)
rstring (kid, value)      -- type=1
rlist (kid, pos, elem)    -- type=2
rset (kid, elem)          -- type=3
rhash (kid, field, value) -- type=4
rzset (kid, elem, score)  -- type=5
```

Single source of truth for key existence, type, expiration, and length.
Type-specific tables store only values. `ON DELETE CASCADE` handles cleanup.

**Fazt applicability**: HIGH if extending KV to richer data structures

---

#### Pattern 2: Real-Valued List Positions

**Problem**: Efficient list insertions without renumbering.

**Solution**:
```sql
rlist (kid INTEGER, pos REAL, elem BLOB)
-- pos is float, not integer!
```

Insert between pos=1.0 and pos=2.0 → use pos=1.5. No renumbering needed.
Allows O(1) insertions anywhere in list. View uses `ORDER BY pos` +
`row_number()` for display indices.

**Fazt applicability**: HIGH for ordered collections

---

#### Pattern 3: View-Based Expiration Filtering

**Problem**: Filter expired keys without scanning every query.

**Solution**:
```sql
create view vkey as
select id as kid, key, type, len, ...
from rkey
where etime is null or etime > unixepoch('subsec');
```

All queries go through views that filter expired keys automatically.
Background goroutine cleans up expired keys every 60 seconds.

**Fazt applicability**: MEDIUM - Fazt KV doesn't currently support TTL

---

#### Pattern 4: Trigger-Maintained Consistency

**Problem**: Keep `len` count accurate without application logic.

**Solution**:
```sql
create trigger rset_on_insert after insert on rset
begin
    update rkey set len = len + 1 where id = new.kid;
end;
```

Database maintains referential integrity and derived values. Application
code stays simple.

**Fazt applicability**: MEDIUM - useful for complex data structures

---

**Other notable patterns**:
- Separate RW/RO connection pools for concurrency
- Dialect abstraction (SQLite vs PostgreSQL with same API)
- Multiple SQLite driver support (mattn/CGO, ncruces/WASM, modernc/pure Go)

---

### go-crush

- **URL**: https://github.com/charmbracelet/crush
- **What**: Charm's AI coding assistant CLI (like Claude Code)
- **Verdict**: PATTERN (code NO-GO due to license)
- **License**: FSL-1.1-MIT (restricts competing use)

**Why PATTERN not EXTRACT**: FSL license prohibits copying code for
competing products. But patterns are ideas - not copyrightable. These
architectural approaches can be implemented independently.

---

#### Pattern 1: Agent Skills Discovery

**Problem**: Allow extensible agent capabilities via filesystem.

**Solution**:
```
1. Define skill locations: ~/.config/app/skills/, ./project-skills/
2. Walk directories, find SKILL.md files
3. Parse: YAML frontmatter (metadata) + markdown body (instructions)
4. Validate: name format, required fields, length limits
5. Inject into system prompt as structured XML
```

**Skill file format** (agentskills.io spec):
```markdown
---
name: skill-name           # alphanumeric + hyphens, matches dirname
description: What it does  # Required, <1024 chars
license: MIT               # Optional
compatibility: "..."       # Optional, <500 chars
metadata:                  # Optional key-value
  author: someone
---
Instructions for the agent to follow when this skill is activated.
Can reference files in the skill directory.
```

**Prompt injection**:
```xml
<available_skills>
  <skill>
    <name>skill-name</name>
    <description>What it does</description>
    <location>/path/to/skill</location>
  </skill>
</available_skills>
```

**Fazt applicability**: HIGH - matches managed skills concept in harness.md

---

#### Pattern 2: Permission Request/Grant Flow

**Problem**: Block tool execution until user approves, with memory.

**Solution**:
```
┌─────────┐     ┌─────────┐     ┌─────────┐
│  Agent  │────▶│ PermSvc │────▶│   UI    │
│  calls  │     │ Request │     │ prompts │
│  tool   │     │ blocks  │     │  user   │
└─────────┘     └────┬────┘     └────┬────┘
                     │               │
                     │◀──────────────┘
                     │  Grant/Deny
                     ▼
              ┌─────────────┐
              │  Unblock    │
              │  caller     │
              └─────────────┘
```

**Key components**:
```go
type PermissionRequest struct {
    ID          string
    ToolName    string    // e.g., "bash", "edit"
    Action      string    // e.g., "execute", "write"
    Path        string    // affected path
    Description string    // human-readable
}

type Service interface {
    Request(req) bool           // Blocks until decision
    Grant(req)                  // One-time approval
    GrantPersistent(req)        // Remember for session
    Deny(req)
}
```

**Optimization layers**:
1. Allowlist: `["view", "ls", "grep"]` - auto-approve
2. Session memory: Same tool+action+path → auto-approve
3. YOLO mode: Skip all prompts (dangerous)

**Fazt applicability**: MEDIUM - relevant for interactive CLI mode

---

#### Pattern 3: Multi-LSP Routing

**Problem**: Route files to correct LSP server by type.

**Solution**:
```go
type LSPClient struct {
    name      string      // "gopls", "typescript-language-server"
    fileTypes []string    // [".go"], [".ts", ".tsx"]
    // ...
}

func (c *LSPClient) HandlesFile(path string) bool {
    ext := filepath.Ext(path)
    return slices.Contains(c.fileTypes, ext)
}

// Router finds appropriate LSP for a file
func (r *Router) GetClient(path string) *LSPClient {
    for _, client := range r.clients {
        if client.HandlesFile(path) {
            return client
        }
    }
    return nil
}
```

**State management per client**:
- `openFiles` map: track which files are open
- `diagnostics` versioned map: cache diagnostics, detect changes
- `serverState`: Starting → Ready → Error

**Fazt applicability**: HIGH - needed for LSP integration in v0.12

---

**Implementation approach for Fazt**:
- Skills: Implement from agentskills.io spec + this pattern
- Permissions: Implement pattern with Fazt's event system
- LSP: Use `charmbracelet/x/powernap` (MIT) + this routing pattern
- MCP: Use `modelcontextprotocol/go-sdk` (Apache 2.0)

---

### go-lingoose

- **URL**: https://github.com/henomis/lingoose
- **What**: Modular Go framework for LLM applications
- **Verdict**: EXTRACT
- **License**: MIT

**Why EXTRACT**: Unlike go-eino/go-adk, LinGoose is truly modular. Each
package is standalone with minimal deps. Two algorithms directly match
Fazt specs and are worth lifting.

**Extracted algorithms:**

1. **RecursiveCharacterTextSplitter** (~110 lines, stdlib only)
   ```go
   // Core algorithm:
   // 1. Find first separator that exists in text
   // 2. Split by that separator
   // 3. For chunks > chunkSize: recursively split with next separator
   // 4. For chunks <= chunkSize: merge with overlap
   separators := []string{"\n\n", "\n", " ", ""}
   ```
   Location: `textsplitter/recursiveTextSplitter.go`
   Use in: `text-splitter.md` implementation

2. **Cosine Similarity** (~30 lines, stdlib only)
   ```go
   // cosine = (a·b) / (|a| * |b|)
   // Handles mismatched vector lengths gracefully
   ```
   Location: `index/vectordb/jsondb/jsondb.go:198-227`
   Use in: `vector.md` implementation

**Pattern also extracted: VectorDB Interface**
```go
type VectorDB interface {
    Insert(context.Context, []Data) error
    Search(context.Context, []float64, *Options) (SearchResults, error)
    Delete(ctx context.Context, ids []string) error
}
```
Separates index orchestration from storage. Fazt's vector.md should
adopt this pattern for pluggable backends.

**What's dropped:**
- Embedder implementations (Fazt has ai-shim)
- JSON file storage (Fazt uses SQLite)
- External loaders (PDF requires pdftotext binary)
- Observer/tracing (Fazt has events)

**Extraction ratio**: ~5% of LinGoose code, ~60% of RAG value

---

### go-adk

- **URL**: https://github.com/google/adk-go
- **What**: Google's Agent Development Kit for Go
- **Verdict**: PATTERN

**Why not EXTRACT**: Framework vs library mismatch (same as go-eino).
Gemini-centric types throughout. Fazt already has ai-shim, mcp, harness.

**Pattern extracted**: State Prefix Convention

```
app:key    → shared across all users of an app
user:key   → shared across a user's sessions
temp:key   → in-memory only, not persisted
key        → session-scoped (default)
```

Single API (`state.get/set`), automatic scoping via prefix, transparent
merging on read. Elegant solution for hierarchical state management.

**Fazt applicability**: MEDIUM - solves problem Fazt doesn't strongly
have yet (personal PaaS, often single-user). Revisit at v0.17 (Realtime)
when session-awareness becomes more important.

**Lighter adoption**: Add `temp:` prefix support to KV for auto-expiring
scratch data. 20% cost, 80% value.

---

### go-eino

- **URL**: https://github.com/cloudwego/eino
- **What**: LangChain-for-Go, LLM application framework
- **Verdict**: NO-GO

**Why NO-GO**: Wants to own execution model; Fazt uses simple
imperative functions. Framework, not extractable library.
