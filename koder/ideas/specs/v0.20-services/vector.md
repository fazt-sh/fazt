# Vector Store

## Summary

SQLite-backed vector storage for semantic search and RAG applications.
Stores document embeddings in the main Fazt database, maintaining the
"One Binary + One DB" philosophy while enabling similarity search.

Inspired by chromem-go's API but with SQLite persistence instead of
file-based storage.

## Why storage.vector (not services)

Vector storage is a persistence layer:
- Stores data in SQLite (like `storage.kv`, `storage.ds`)
- Provides query capabilities over stored data
- Part of the core storage subsystem
- `fazt.storage.vector` parallels `fazt.storage.kv` and `fazt.storage.ds`

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              fazt.storage.vector                         │
│                                                                         │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐               │
│  │ Collections │     │  Documents  │     │   Query     │               │
│  │ Management  │     │  + Vectors  │     │   Engine    │               │
│  └─────────────┘     └─────────────┘     └─────────────┘               │
│         │                   │                   │                       │
│         └───────────────────┴───────────────────┘                       │
│                             │                                           │
│                             ▼                                           │
│                      ┌─────────────┐                                    │
│                      │   SQLite    │                                    │
│                      │   (data.db) │                                    │
│                      └─────────────┘                                    │
└─────────────────────────────────────────────────────────────────────────┘
                              │
                              │ Embedding creation
                              ▼
                      ┌─────────────┐
                      │ fazt.ai     │
                      │ .embed()    │
                      └─────────────┘
```

## Storage Schema

```sql
-- Collections (namespaces for documents)
CREATE TABLE vector_collections (
    id TEXT PRIMARY KEY,
    app_uuid TEXT,
    name TEXT NOT NULL,
    metadata_json TEXT,
    embedding_model TEXT,      -- e.g., 'openai/text-embedding-3-small'
    dimensions INTEGER,        -- e.g., 1536
    document_count INTEGER DEFAULT 0,
    created_at INTEGER,
    updated_at INTEGER,
    UNIQUE(app_uuid, name)
);

-- Documents with embeddings
CREATE TABLE vector_documents (
    id TEXT PRIMARY KEY,
    collection_id TEXT NOT NULL,
    content TEXT,              -- pageContent
    metadata_json TEXT,        -- Document metadata
    embedding BLOB NOT NULL,   -- []float32 as little-endian binary
    created_at INTEGER,
    FOREIGN KEY (collection_id) REFERENCES vector_collections(id) ON DELETE CASCADE
);

CREATE INDEX idx_vector_docs_collection ON vector_documents(collection_id);
```

### Embedding Storage Format

Vectors stored as binary blobs for efficiency:

```go
// Encode: []float32 → []byte
func encodeEmbedding(v []float32) []byte {
    buf := make([]byte, len(v)*4)
    for i, f := range v {
        binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(f))
    }
    return buf
}

// Decode: []byte → []float32
func decodeEmbedding(b []byte) []float32 {
    v := make([]float32, len(b)/4)
    for i := range v {
        v[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*4:]))
    }
    return v
}
```

## Usage

### Create Collection

```javascript
// Create a collection with auto-embedding
const collection = await fazt.storage.vector.createCollection('articles', {
    embeddingModel: 'openai/text-embedding-3-small',  // Uses fazt.ai.embed()
    metadata: { description: 'Knowledge base articles' }
});

// Or with custom embedding function
const collection = await fazt.storage.vector.createCollection('custom', {
    dimensions: 768,  // Must specify if not using built-in model
    metadata: { description: 'Custom embeddings' }
});
```

### Add Documents

```javascript
// Add documents (embeddings created automatically)
await fazt.storage.vector.addDocuments('articles', [
    { pageContent: 'Machine learning is...', metadata: { source: 'wiki' } },
    { pageContent: 'Neural networks are...', metadata: { source: 'blog' } }
]);

// Add with pre-computed embeddings
await fazt.storage.vector.addDocuments('custom', [
    {
        pageContent: 'Some text...',
        metadata: { source: 'file.txt' },
        embedding: [0.1, 0.2, 0.3, ...]  // Pre-computed
    }
]);

// Add with concurrency control
await fazt.storage.vector.addDocuments('articles', docs, {
    concurrency: 5,  // Max concurrent embedding API calls
    batchSize: 100   // Commit every N documents
});
```

### Query (Similarity Search)

```javascript
// Query by text (auto-embedded)
const results = await fazt.storage.vector.query('articles', 'What is deep learning?', {
    limit: 10
});
// Returns LangChain-format documents with scores:
// [
//   { pageContent: '...', metadata: { source: 'wiki' }, score: 0.95 },
//   { pageContent: '...', metadata: { source: 'blog' }, score: 0.87 },
// ]

// Query by embedding
const results = await fazt.storage.vector.queryEmbedding('articles', queryVector, {
    limit: 10
});

// Query with metadata filter
const results = await fazt.storage.vector.query('articles', 'deep learning', {
    limit: 10,
    where: { source: 'wiki' }  // Only wiki articles
});

// Query with content filter
const results = await fazt.storage.vector.query('articles', 'deep learning', {
    limit: 10,
    whereContent: { contains: 'neural' }  // Must contain 'neural'
});
```

### Get Documents

```javascript
// Get by ID
const doc = await fazt.storage.vector.get('articles', 'doc-123');

// Get by metadata
const docs = await fazt.storage.vector.getByMetadata('articles', {
    source: 'wiki'
});

// List all document IDs
const ids = await fazt.storage.vector.listIds('articles');

// List all documents (without embeddings for efficiency)
const docs = await fazt.storage.vector.list('articles', {
    limit: 100,
    offset: 0
});

// Count documents
const count = await fazt.storage.vector.count('articles');
```

### Delete Documents

```javascript
// Delete by ID
await fazt.storage.vector.delete('articles', 'doc-123');

// Delete by IDs
await fazt.storage.vector.deleteMany('articles', ['doc-1', 'doc-2', 'doc-3']);

// Delete by metadata
await fazt.storage.vector.deleteByMetadata('articles', { source: 'outdated' });

// Clear collection (delete all documents)
await fazt.storage.vector.clear('articles');
```

### Collection Management

```javascript
// List collections
const collections = await fazt.storage.vector.listCollections();
// [{ name: 'articles', documentCount: 1234, dimensions: 1536 }, ...]

// Get collection info
const info = await fazt.storage.vector.getCollection('articles');

// Delete collection (and all documents)
await fazt.storage.vector.deleteCollection('articles');

// Update collection metadata
await fazt.storage.vector.updateCollection('articles', {
    metadata: { description: 'Updated description' }
});
```

## JS API

```javascript
// Collection management
fazt.storage.vector.createCollection(name, options?)
// options: { embeddingModel, dimensions, metadata }
fazt.storage.vector.getCollection(name)
fazt.storage.vector.listCollections()
fazt.storage.vector.deleteCollection(name)
fazt.storage.vector.updateCollection(name, options)

// Document operations
fazt.storage.vector.addDocuments(collection, documents, options?)
// documents: { pageContent, metadata, embedding? }[]
// options: { concurrency, batchSize }

fazt.storage.vector.get(collection, id)
fazt.storage.vector.getByMetadata(collection, where)
fazt.storage.vector.list(collection, options?)
fazt.storage.vector.listIds(collection)
fazt.storage.vector.count(collection)

fazt.storage.vector.delete(collection, id)
fazt.storage.vector.deleteMany(collection, ids)
fazt.storage.vector.deleteByMetadata(collection, where)
fazt.storage.vector.clear(collection)

// Query (similarity search)
fazt.storage.vector.query(collection, text, options?)
// options: { limit, where, whereContent }
// Returns: { pageContent, metadata, score }[]

fazt.storage.vector.queryEmbedding(collection, embedding, options?)
// Same options, but with pre-computed query embedding
```

## Similarity Search Algorithm

Uses cosine similarity (same as chromem-go):

```go
// Cosine similarity between two normalized vectors = dot product
func cosineSimilarity(a, b []float32) float32 {
    var dot float32
    for i := range a {
        dot += a[i] * b[i]
    }
    return dot
}

// Vectors are normalized on insertion
func normalize(v []float32) []float32 {
    var sum float32
    for _, f := range v {
        sum += f * f
    }
    norm := float32(math.Sqrt(float64(sum)))
    result := make([]float32, len(v))
    for i, f := range v {
        result[i] = f / norm
    }
    return result
}
```

### Query Execution

```go
func Query(collectionID, queryEmbedding []float32, limit int) []ScoredDocument {
    // 1. Load all document embeddings from SQLite
    rows, _ := db.Query(`
        SELECT id, content, metadata_json, embedding
        FROM vector_documents
        WHERE collection_id = ?
    `, collectionID)

    // 2. Compute similarity for each document
    type scored struct {
        doc   Document
        score float32
    }
    var results []scored

    for rows.Next() {
        var id, content, metadataJSON string
        var embeddingBlob []byte
        rows.Scan(&id, &content, &metadataJSON, &embeddingBlob)

        embedding := decodeEmbedding(embeddingBlob)
        score := cosineSimilarity(queryEmbedding, embedding)

        results = append(results, scored{
            doc:   Document{PageContent: content, Metadata: parseJSON(metadataJSON)},
            score: score,
        })
    }

    // 3. Sort by score descending
    sort.Slice(results, func(i, j int) bool {
        return results[i].score > results[j].score
    })

    // 4. Return top N
    if len(results) > limit {
        results = results[:limit]
    }

    return results
}
```

### Performance Characteristics

| Documents | Query Time | Memory |
|-----------|-----------|--------|
| 1,000 | ~1ms | ~6 MB |
| 10,000 | ~10ms | ~60 MB |
| 100,000 | ~100ms | ~600 MB |

For most personal/small team use cases (<100k documents), this is sufficient.

## Integration with fazt.ai.embed()

```javascript
// Collection with embedding model auto-calls fazt.ai.embed()
await fazt.storage.vector.createCollection('articles', {
    embeddingModel: 'openai/text-embedding-3-small'
});

// When adding documents, embeddings are created automatically:
await fazt.storage.vector.addDocuments('articles', [
    { pageContent: 'Some text...' }
]);
// Internally calls: fazt.ai.embed('Some text...', { model: 'text-embedding-3-small' })

// Queries also auto-embed:
await fazt.storage.vector.query('articles', 'search query');
// Internally calls: fazt.ai.embed('search query', { model: 'text-embedding-3-small' })
```

### Supported Embedding Models

| Model | Provider | Dimensions | Cost |
|-------|----------|------------|------|
| `text-embedding-3-small` | OpenAI | 1536 | $0.02/1M tokens |
| `text-embedding-3-large` | OpenAI | 3072 | $0.13/1M tokens |
| `text-embedding-ada-002` | OpenAI | 1536 | $0.10/1M tokens |
| `nomic-embed-text` | Ollama | 768 | Free (local) |
| `mxbai-embed-large` | Ollama | 1024 | Free (local) |

## Complete RAG Example

```javascript
// === INDEXING (run once) ===

// 1. Create collection
await fazt.storage.vector.createCollection('site-docs', {
    embeddingModel: 'openai/text-embedding-3-small'
});

// 2. Crawl and process pages
const urls = await crawlSite('https://example.com');  // Returns 2000 URLs
const allDocs = [];

for (const url of urls) {
    const html = await fetch(url).then(r => r.text());
    const text = extractText(html);  // Strip HTML

    // Split into chunks
    const chunks = fazt.lib.text.split(text, {
        chunkSize: 500,
        chunkOverlap: 50
    });

    // Create documents with metadata
    for (let i = 0; i < chunks.length; i++) {
        allDocs.push(
            fazt.lib.document.create(chunks[i], {
                url,
                chunkIndex: i,
                title: extractTitle(html)
            })
        );
    }
}

// 3. Index all documents
console.log(`Indexing ${allDocs.length} chunks...`);
await fazt.storage.vector.addDocuments('site-docs', allDocs, {
    concurrency: 10,  // Parallel embedding calls
    batchSize: 100    // Progress updates
});
// ~20,000 chunks at $0.02/1M tokens ≈ $0.01 total cost


// === QUERYING (every user question) ===

async function answerQuestion(userQuestion) {
    // 1. Find relevant chunks
    const relevantDocs = await fazt.storage.vector.query(
        'site-docs',
        userQuestion,
        { limit: 5 }
    );

    // 2. Build context for LLM
    const context = relevantDocs
        .map(d => `[Source: ${d.metadata.url}]\n${d.pageContent}`)
        .join('\n\n---\n\n');

    // 3. Generate answer with RAG
    const answer = await fazt.ai.complete(`
You are a helpful assistant. Answer the user's question based ONLY on the
provided context. If the context doesn't contain enough information, say so.

Context:
${context}

Question: ${userQuestion}

Answer:`, {
        model: 'gpt-4',
        temperature: 0.3
    });

    return {
        answer: answer.text,
        sources: relevantDocs.map(d => d.metadata.url)
    };
}

// Usage
const result = await answerQuestion("How do I reset my password?");
console.log(result.answer);
console.log("Sources:", result.sources);
```

## CLI

```bash
# Collection management
fazt vector collections list
fazt vector collections create articles --model openai/text-embedding-3-small
fazt vector collections delete articles
fazt vector collections info articles

# Document operations
fazt vector add articles --file docs.json
fazt vector add articles --content "Some text..." --metadata '{"source":"cli"}'
fazt vector list articles --limit 10
fazt vector count articles
fazt vector delete articles doc-123

# Query
fazt vector query articles "What is machine learning?" --limit 5
fazt vector query articles "password reset" --where '{"source":"help"}'

# Bulk operations
fazt vector import articles backup.json
fazt vector export articles > backup.json
fazt vector clear articles
```

## HTTP API

```
# Collections
GET    /api/vector/collections
POST   /api/vector/collections
GET    /api/vector/collections/{name}
DELETE /api/vector/collections/{name}

# Documents
POST   /api/vector/collections/{name}/documents
GET    /api/vector/collections/{name}/documents
GET    /api/vector/collections/{name}/documents/{id}
DELETE /api/vector/collections/{name}/documents/{id}

# Query
POST   /api/vector/collections/{name}/query
Body: { "text": "query string", "limit": 10, "where": {} }

# Bulk
POST   /api/vector/collections/{name}/import
GET    /api/vector/collections/{name}/export
DELETE /api/vector/collections/{name}/clear
```

## Limits

| Limit | Default |
|-------|---------|
| Max collections per app | 100 |
| Max documents per collection | 1,000,000 |
| Max document content size | 100 KB |
| Max metadata size | 10 KB |
| Max embedding dimensions | 4096 |
| Query timeout | 30 seconds |

## Future Enhancements

Potential improvements (not in initial release):

1. **HNSW Index**: Approximate nearest neighbor for faster queries at scale
2. **Hybrid Search**: Combine vector search with full-text (Bleve)
3. **Incremental Updates**: Update embeddings without re-indexing
4. **Compression**: Quantized vectors for smaller storage
5. **Sharding**: Distribute across multiple SQLite files

For now, brute-force search is sufficient for personal-scale use cases.
