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

## Log

| Date | Project | Verdict | Reason |
|------|---------|---------|--------|
| 2025-01-03 | [go-eino](https://github.com/cloudwego/eino) | NO-GO | Framework vs library mismatch. LangChain-for-Go wants to own execution model; Fazt uses simple imperative functions. |
