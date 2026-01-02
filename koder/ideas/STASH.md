# Ideas Stash

Ideas evaluated but deferred for future consideration.

---

## From: go-langchain Analysis (2026-01-02)

### Markdown-Aware Text Splitter

**Source:** `go-langchain/textsplitter/markdown_splitter.go`

**What:** Semantic markdown chunking that preserves document structure.
- Parses markdown AST
- Preserves heading hierarchy in chunks
- Handles tables, code blocks, lists
- ~850 lines

**Dependency:** `gitlab.com/golang-commonmark/markdown` (pure Go)

**Why stashed:**
- RecursiveCharacter splitter covers 90% of use cases
- Markdown splitter adds complexity for specialized use
- Revisit when users request markdown-specific chunking

**Potential surface:**
```javascript
fazt.lib.text.splitMarkdown(content, {
    chunkSize: 1000,
    headingHierarchy: true,
    codeBlocks: true
})
```

---

### HTTP Record/Replay Testing (httprr)

**Source:** `go-langchain/internal/httprr/`

**What:** Deterministic HTTP testing utility.
- Record real HTTP interactions to files
- Replay in tests for determinism
- Request/response scrubbing for secrets
- ~1500 lines, zero dependencies

**Why stashed:**
- Internal development tool, not user-facing
- Fazt already has testing patterns
- Valuable for `fazt.dev.*` integration testing
- Revisit when adding more external service integrations

**Potential use:** Internal testing infrastructure for device integrations.

---

## Evaluation Criteria for Unstashing

1. User demand (requests, issues)
2. Prerequisite features implemented
3. Clear integration path exists
4. Complexity/value ratio acceptable
