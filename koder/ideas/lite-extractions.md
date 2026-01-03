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
| 2026-01-03 | go-lingoose | EXTRACT | Text splitter + cosine algos    |
| 2026-01-03 | go-adk      | PATTERN | State prefix pattern worth study |
| 2026-01-03 | go-eino     | NO-GO   | Framework vs library mismatch    |

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
