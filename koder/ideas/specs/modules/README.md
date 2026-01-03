# Modules

Extractable modules designed to work within Fazt but also standalone.
These are candidates for separate repositories in the org.

## Philosophy

Modules are:
- **Pure Go**: No CGO, embeddable anywhere
- **Zero-config**: Sensible defaults, convention over configuration
- **Extractable**: Minimal Fazt coupling, usable independently
- **Lite**: 80% of features for 20% of complexity

## Modules

| Module       | Purpose                        | Status  |
|--------------|--------------------------------|---------|
| jekyll-lite  | Jekyll-compatible static sites | Planned |

## Integration Pattern

Modules integrate with Fazt but don't depend on it:

```go
// Standalone usage
builder := jekylllite.New()
builder.SetSource("./my-blog")
builder.SetOutput("./_site")
builder.Build()

// Fazt integration
builder := jekylllite.New()
builder.SetSource("./my-blog")
builder.SetOutput(fazt.VFS(appID))  // Write to VFS
builder.Build()
```

## Extraction Criteria

A module is ready for extraction when:
1. Zero imports from `internal/` packages
2. Standalone CLI works without Fazt
3. Test coverage > 80%
4. Documentation complete
5. At least one production deployment
