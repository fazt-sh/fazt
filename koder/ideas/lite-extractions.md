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

| Date       | Project  | Verdict | Reason                           |
|------------|----------|---------|----------------------------------|
| 2026-01-03 | go-adk   | PATTERN | State prefix pattern worth study |
| 2026-01-03 | go-eino  | NO-GO   | Framework vs library mismatch    |

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
