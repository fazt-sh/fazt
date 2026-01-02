# Document Format

## Summary

LangChain-compatible document format for interoperability with the LangChain
ecosystem. Provides exact format compatibility for import/export of documents
between Fazt and any LangChain-based tool (Python, JavaScript, Go).

## Why Exact Compatibility

LangChain is the dominant framework for LLM applications:
- Python: 100k+ GitHub stars
- JavaScript: 15k+ stars
- Go (langchaingo): 5k+ stars

By matching their Document format exactly, Fazt users can:
- Import documents from any LangChain pipeline
- Export documents to any vector store
- Use existing LangChain tools and tutorials
- Migrate between Fazt and LangChain seamlessly

## The LangChain Document Format

```typescript
// This is THE canonical format used across all LangChain implementations
interface Document {
    pageContent: string;           // The text content
    metadata: Record<string, any>; // Arbitrary key-value pairs
}

// When returned from similarity search, includes score
interface DocumentWithScore extends Document {
    score: number;  // 0.0 to 1.0, higher = more similar
}
```

**Important:** The field names are `pageContent` (camelCase) and `metadata`,
not `page_content` or `content`. This is the exact format used by:
- Python LangChain
- LangChain.js
- LangChainGo

## Usage

### Create Document

```javascript
// Simple
const doc = fazt.lib.document.create("Machine learning is...");

// With metadata
const doc = fazt.lib.document.create(
    "Machine learning is a subset of AI...",
    {
        source: "wikipedia",
        title: "Machine Learning",
        url: "https://en.wikipedia.org/wiki/Machine_learning",
        timestamp: Date.now()
    }
);
// Returns: { pageContent: "...", metadata: { source: "...", ... } }
```

### From/To JSON

```javascript
// Parse from JSON (e.g., from LangChain export)
const doc = fazt.lib.document.fromJSON({
    pageContent: "Content here...",
    metadata: { source: "file.txt" }
});

// Convert to JSON (for LangChain import)
const json = fazt.lib.document.toJSON(doc);
// { "pageContent": "...", "metadata": { ... } }

// Batch conversion
const docs = fazt.lib.document.fromJSONArray(jsonArray);
const jsonArray = fazt.lib.document.toJSONArray(docs);
```

### Add Score (for search results)

```javascript
// After similarity search, add score to document
const scoredDoc = fazt.lib.document.withScore(doc, 0.95);
// { pageContent: "...", metadata: { ... }, score: 0.95 }
```

### Validate Format

```javascript
// Check if object is valid LangChain Document
const isValid = fazt.lib.document.isValid(obj);
// Returns: boolean

// Validate and throw if invalid
fazt.lib.document.validate(obj);
// Throws: Error if not valid Document format
```

### Convert from Other Formats

```javascript
// From chromem-go format
const doc = fazt.lib.document.fromChromem({
    ID: "doc-123",
    Content: "Text here...",
    Metadata: { source: "file.txt" },
    Embedding: [0.1, 0.2, ...]
});
// Returns: { pageContent: "Text here...", metadata: { id: "doc-123", source: "file.txt" } }

// To chromem-go format
const chromemDoc = fazt.lib.document.toChromem(doc, "doc-123");
// Returns: { ID: "doc-123", Content: "...", Metadata: { ... } }
```

## JS API

```javascript
// Create documents
fazt.lib.document.create(pageContent, metadata?)
// Returns: { pageContent, metadata }

// JSON conversion
fazt.lib.document.fromJSON(json)
fazt.lib.document.toJSON(doc)
fazt.lib.document.fromJSONArray(jsonArray)
fazt.lib.document.toJSONArray(docs)

// Score handling
fazt.lib.document.withScore(doc, score)
// Returns: { pageContent, metadata, score }

// Validation
fazt.lib.document.isValid(obj)
// Returns: boolean
fazt.lib.document.validate(obj)
// Throws if invalid

// Format conversion
fazt.lib.document.fromChromem(chromemDoc)
fazt.lib.document.toChromem(doc, id)

// Utilities
fazt.lib.document.getContent(doc)
// Returns: string (the pageContent)
fazt.lib.document.getMetadata(doc, key)
// Returns: any (metadata value)
fazt.lib.document.setMetadata(doc, key, value)
// Returns: new Document with updated metadata
fazt.lib.document.mergeMetadata(doc, additionalMetadata)
// Returns: new Document with merged metadata
```

## Type Definitions

```typescript
// Core Document type (matches LangChain exactly)
interface Document {
    pageContent: string;
    metadata: Record<string, any>;
}

// Document with similarity score (for search results)
interface ScoredDocument extends Document {
    score: number;
}

// chromem-go's Document type (for conversion)
interface ChromemDocument {
    ID: string;
    Content: string;
    Metadata: Record<string, string>;  // Note: string values only
    Embedding?: number[];
}
```

## Format Comparison

| Field | LangChain | chromem-go | Fazt |
|-------|-----------|------------|------|
| Content | `pageContent` | `Content` | `pageContent` |
| Metadata type | `map[string]any` | `map[string]string` | `map[string]any` |
| ID | in metadata | `ID` | in metadata |
| Score | `score` | `Similarity` | `score` |
| Embedding | separate | `Embedding` | separate |

**Fazt follows LangChain exactly.** Conversion to/from chromem-go is handled
by the `fromChromem`/`toChromem` functions.

## Examples

### Load from LangChain Export

```javascript
// LangChain Python exports documents as JSON
const exported = JSON.parse(await Deno.readTextFile("langchain_docs.json"));

// Convert to Fazt documents
const docs = fazt.lib.document.fromJSONArray(exported);

// Index in vector store
await fazt.storage.vector.addDocuments("knowledge-base", docs);
```

### Export for LangChain

```javascript
// Get documents from vector store
const docs = await fazt.storage.vector.getAll("knowledge-base");

// Export as LangChain-compatible JSON
const json = fazt.lib.document.toJSONArray(docs);
await Deno.writeTextFile("export.json", JSON.stringify(json, null, 2));

// Can now be loaded in Python:
// from langchain.schema import Document
// docs = [Document(**d) for d in json.load(open("export.json"))]
```

### Integration with Text Splitter

```javascript
// Split returns LangChain-compatible documents
const chunks = fazt.lib.text.splitDocuments([
    fazt.lib.document.create("Long article here...", { source: "article.md" })
], { chunkSize: 500 });

// chunks is already in correct format for vector store
await fazt.storage.vector.addDocuments("articles", chunks);
```

### Search Result with Score

```javascript
// Query returns scored documents
const results = await fazt.storage.vector.query("articles", "machine learning");
// [
//   { pageContent: "...", metadata: { source: "..." }, score: 0.95 },
//   { pageContent: "...", metadata: { source: "..." }, score: 0.87 },
// ]

// Check score
results.forEach(doc => {
    console.log(`${doc.metadata.source}: ${(doc.score * 100).toFixed(1)}% match`);
});
```

## Implementation Notes

### Metadata Handling

```go
// Metadata can contain any JSON-serializable value
type Document struct {
    PageContent string         `json:"pageContent"`
    Metadata    map[string]any `json:"metadata"`
}

// When score is present (search results)
type ScoredDocument struct {
    Document
    Score float32 `json:"score"`
}
```

### Validation Rules

```go
func Validate(doc any) error {
    m, ok := doc.(map[string]any)
    if !ok {
        return errors.New("document must be an object")
    }

    pageContent, ok := m["pageContent"]
    if !ok {
        return errors.New("document must have pageContent field")
    }
    if _, ok := pageContent.(string); !ok {
        return errors.New("pageContent must be a string")
    }

    if metadata, ok := m["metadata"]; ok {
        if _, ok := metadata.(map[string]any); !ok {
            return errors.New("metadata must be an object")
        }
    }

    return nil
}
```

### chromem-go Conversion

```go
// Note: chromem-go metadata is map[string]string, not map[string]any
// We convert by JSON encoding non-string values
func ToChromem(doc Document, id string) chromem.Document {
    metadata := make(map[string]string)
    for k, v := range doc.Metadata {
        switch val := v.(type) {
        case string:
            metadata[k] = val
        default:
            // JSON encode non-string values
            bytes, _ := json.Marshal(val)
            metadata[k] = string(bytes)
        }
    }
    return chromem.Document{
        ID:       id,
        Content:  doc.PageContent,
        Metadata: metadata,
    }
}
```

## CLI

```bash
# Validate document JSON
fazt document validate doc.json

# Convert between formats
fazt document convert --from langchain --to chromem input.json output.json

# Create document from text file
fazt document create --source article.md "$(cat article.md)"
```

## Why Not Just Use chromem-go's Format?

chromem-go's format differs in important ways:

1. **Field names**: `Content` vs `pageContent`
2. **Metadata type**: `map[string]string` vs `map[string]any`
3. **ID handling**: Explicit field vs in metadata

The LangChain format is the **industry standard**. Supporting it exactly means:
- Users don't need to transform data
- Tutorials and examples work as-is
- Integration with Python/JS ecosystems is seamless

Fazt handles the conversion internally so users work with the standard format.
