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
- **NO-GO**: Doesn't fit, won't revisit soon
- **DEFER**: Interesting but blocked on prerequisites
- **PATTERN**: No code to extract, but valuable design pattern

## Log

| Date       | Project     | Verdict | Reason                           |
|------------|-------------|---------|----------------------------------|
| 2026-01-03 | go-crush    | PATTERN | Skills, permissions, LSP routing |
| 2026-01-03 | go-lingoose | EXTRACT | Text splitter + cosine algos    |
| 2026-01-03 | go-adk      | PATTERN | State prefix pattern worth study |
| 2026-01-03 | go-eino     | NO-GO   | Framework vs library mismatch    |

## Note: Patterns vs Code

Patterns (architectural approaches, designs) are **always extractable** -
ideas aren't copyrightable. Even from restrictively-licensed projects like
go-crush (FSL), we document patterns for independent implementation.

If you find a **permissive Go project** with similar agentic patterns,
evaluate it - reference code accelerates implementation even when patterns
are already documented.

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
