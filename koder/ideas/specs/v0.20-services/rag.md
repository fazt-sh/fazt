# RAG Service

## Summary

Retrieval-Augmented Generation (RAG) made simple. Define a pipeline once,
ingest documents, ask questions. Fazt handles chunking, embedding, storage,
retrieval, and context injection.

## What is RAG?

RAG solves a fundamental LLM limitation: **LLMs don't know your data.**

Ask GPT-4 "What's our refund policy?" and it will hallucinate an answer.
RAG fixes this by:

1. Finding relevant chunks from YOUR documents
2. Injecting them into the prompt as context
3. LLM answers based on your actual data

```
Without RAG:
  User: "What's our refund policy?"
  LLM:  "Generally, refund policies vary..." (hallucination)

With RAG:
  User: "What's our refund policy?"
  [System finds: "Refunds available within 30 days..." from docs/policy.md]
  LLM:  "You can get a refund within 30 days of purchase." (grounded)
```

## The Manual Workflow (What RAG Collapses)

Without this service, implementing RAG requires 6 steps:

```javascript
// === STEP 1: Chunk documents ===
const text = await fazt.fs.read('docs/policy.md');
const chunks = fazt.lib.text.split(text, {
    chunkSize: 500,
    chunkOverlap: 50
});

// === STEP 2: Create documents with metadata ===
const docs = chunks.map((chunk, i) =>
    fazt.lib.document.create(chunk, { source: 'policy.md', index: i })
);

// === STEP 3: Create vector collection ===
await fazt.storage.vector.createCollection('knowledge', {
    embeddingModel: 'openai/text-embedding-3-small'
});

// === STEP 4: Embed and store ===
await fazt.storage.vector.addDocuments('knowledge', docs);

// === STEP 5: Query (on each user question) ===
const relevant = await fazt.storage.vector.query('knowledge', userQuestion, {
    limit: 5
});

// === STEP 6: Build prompt and call LLM ===
const context = relevant.map(d => d.pageContent).join('\n\n');
const answer = await fazt.ai.complete(`
Context:
${context}

Question: ${userQuestion}
Answer based only on the context above.
`);
```

**Problems with manual approach:**
- Boilerplate for every RAG app
- Easy to get chunking wrong
- Must manage collection lifecycle
- Prompt engineering repeated everywhere
- No incremental updates

## The RAG Service

One pipeline definition. Automatic everything.

### Create Pipeline

```javascript
// Define once
await fazt.services.rag.create('support-docs', {
    // What to index
    sources: ['docs/**/*.md', 'faq/**/*.md'],

    // How to chunk (optional, good defaults)
    chunkSize: 500,
    chunkOverlap: 50,

    // Which embedding model
    embedModel: 'openai/text-embedding-3-small'
});
```

### Ingest Documents

```javascript
// Index all sources
await fazt.services.rag.ingest('support-docs');

// Or ingest specific content
await fazt.services.rag.ingestText('support-docs', content, {
    metadata: { source: 'user-upload', title: 'Custom Doc' }
});

// Incremental: only new/changed files
await fazt.services.rag.sync('support-docs');
```

### Query

```javascript
// Just retrieval (get relevant chunks)
const chunks = await fazt.services.rag.retrieve('support-docs',
    'How do I reset my password?',
    { limit: 5 }
);
// Returns: [{ content, metadata, score }, ...]

// Full RAG answer
const result = await fazt.services.rag.ask('support-docs',
    'How do I reset my password?'
);
// Returns: { answer, sources: [{ content, metadata }] }
```

### That's It

Three methods for 90% of RAG use cases:
1. `create()` - Define pipeline
2. `ingest()` - Index documents
3. `ask()` - Get grounded answers

## Architecture

RAG composes existing primitives:

```
┌─────────────────────────────────────────────────────────────────┐
│                    fazt.services.rag                             │
│                                                                  │
│  ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐     │
│  │ Pipeline │   │  Ingest  │   │ Retrieve │   │   Ask    │     │
│  │  Config  │   │  Engine  │   │  Engine  │   │  Engine  │     │
│  └────┬─────┘   └────┬─────┘   └────┬─────┘   └────┬─────┘     │
│       │              │              │              │            │
└───────┼──────────────┼──────────────┼──────────────┼────────────┘
        │              │              │              │
        ▼              ▼              ▼              ▼
   ┌─────────┐   ┌──────────┐   ┌──────────┐   ┌─────────┐
   │ SQLite  │   │ lib.text │   │ storage  │   │ ai      │
   │ config  │   │ .split() │   │ .vector  │   │.complete│
   └─────────┘   └──────────┘   └──────────┘   └─────────┘
```

**No new capabilities.** RAG is pure composition:
- `fazt.lib.text.split()` for chunking
- `fazt.lib.document.create()` for document format
- `fazt.storage.vector.*` for embedding storage
- `fazt.ai.complete()` for LLM calls

## JS API

```javascript
// === Pipeline Management ===

fazt.services.rag.create(name, options)
// options: {
//   sources: string[],           // Glob patterns for files
//   chunkSize: number,           // Default: 500
//   chunkOverlap: number,        // Default: 50
//   embedModel: string,          // Default: 'openai/text-embedding-3-small'
//   systemPrompt: string,        // Custom RAG prompt (advanced)
//   metadata: object             // Pipeline metadata
// }

fazt.services.rag.get(name)
// Returns pipeline config and stats

fazt.services.rag.list()
// Returns all pipelines

fazt.services.rag.update(name, options)
// Update pipeline config (re-ingest may be needed)

fazt.services.rag.delete(name)
// Delete pipeline and all indexed data


// === Ingestion ===

fazt.services.rag.ingest(name, options?)
// Index all sources defined in pipeline
// options: { force: boolean }  // Re-index even if unchanged

fazt.services.rag.ingestText(name, text, options?)
// Index arbitrary text
// options: { metadata: object }

fazt.services.rag.ingestUrl(name, url, options?)
// Fetch and index URL content
// options: { metadata: object, selector: string }

fazt.services.rag.sync(name)
// Incremental sync: index new/changed, remove deleted

fazt.services.rag.clear(name)
// Remove all indexed documents (keep pipeline)


// === Retrieval ===

fazt.services.rag.retrieve(name, query, options?)
// Get relevant chunks without LLM
// options: { limit: number, threshold: number, filter: object }
// Returns: [{ content, metadata, score }]


// === Question Answering ===

fazt.services.rag.ask(name, question, options?)
// Full RAG: retrieve + LLM answer
// options: {
//   limit: number,              // Chunks to retrieve (default: 5)
//   model: string,              // LLM model (default: from fazt.ai config)
//   temperature: number,        // Default: 0.3 (factual)
//   includeSources: boolean,    // Return source chunks (default: true)
//   systemPrompt: string        // Override pipeline prompt
// }
// Returns: { answer: string, sources: [{ content, metadata }] }

fazt.services.rag.chat(name, messages, options?)
// Multi-turn RAG conversation
// messages: [{ role: 'user'|'assistant', content: string }]
// Returns: { answer, sources }
```

## CLI

```bash
# Pipeline management
fazt rag list
fazt rag create support-docs --sources "docs/**/*.md" --embed-model openai/text-embedding-3-small
fazt rag show support-docs
fazt rag delete support-docs

# Ingestion
fazt rag ingest support-docs
fazt rag ingest support-docs --force        # Re-index all
fazt rag sync support-docs                  # Incremental
fazt rag clear support-docs                 # Remove indexed data

# Querying
fazt rag retrieve support-docs "password reset" --limit 5
fazt rag ask support-docs "How do I reset my password?"

# Stats
fazt rag stats support-docs
# Output:
#   Pipeline: support-docs
#   Sources: docs/**/*.md, faq/**/*.md
#   Documents: 47 files
#   Chunks: 892
#   Last sync: 2024-01-15 10:30:00
```

## HTTP API

```
# Pipeline management
GET    /api/rag/pipelines
POST   /api/rag/pipelines
GET    /api/rag/pipelines/{name}
PUT    /api/rag/pipelines/{name}
DELETE /api/rag/pipelines/{name}

# Ingestion
POST   /api/rag/pipelines/{name}/ingest
POST   /api/rag/pipelines/{name}/ingest/text
POST   /api/rag/pipelines/{name}/ingest/url
POST   /api/rag/pipelines/{name}/sync
DELETE /api/rag/pipelines/{name}/documents

# Querying
POST   /api/rag/pipelines/{name}/retrieve
Body: { "query": "...", "limit": 5 }

POST   /api/rag/pipelines/{name}/ask
Body: { "question": "...", "options": {} }

POST   /api/rag/pipelines/{name}/chat
Body: { "messages": [...], "options": {} }

# Stats
GET    /api/rag/pipelines/{name}/stats
```

## Storage Schema

```sql
-- Pipeline definitions
CREATE TABLE rag_pipelines (
    id TEXT PRIMARY KEY,
    app_uuid TEXT NOT NULL,
    name TEXT NOT NULL,
    config_json TEXT NOT NULL,      -- sources, chunk settings, model
    stats_json TEXT,                -- document count, last sync
    created_at INTEGER,
    updated_at INTEGER,
    UNIQUE(app_uuid, name)
);

-- Indexed file tracking (for incremental sync)
CREATE TABLE rag_sources (
    id TEXT PRIMARY KEY,
    pipeline_id TEXT NOT NULL,
    path TEXT NOT NULL,             -- File path or URL
    hash TEXT NOT NULL,             -- Content hash for change detection
    chunk_count INTEGER,
    indexed_at INTEGER,
    FOREIGN KEY (pipeline_id) REFERENCES rag_pipelines(id) ON DELETE CASCADE
);

-- Documents stored in fazt.storage.vector (not duplicated here)
-- Vector collection name: rag_{pipeline_id}
```

## Default System Prompt

```
You are a helpful assistant. Answer the user's question based ONLY on the
provided context. If the context doesn't contain enough information to
answer the question, say "I don't have enough information to answer that."

Do not make up information. Do not use knowledge outside the provided context.
If you quote from the context, be accurate.

Context:
{context}

Question: {question}
```

Customizable via `systemPrompt` option.

## Example: Documentation Chatbot

```javascript
// === Setup (run once) ===

// Create pipeline
await fazt.services.rag.create('docs', {
    sources: ['docs/**/*.md', 'blog/**/*.md'],
    chunkSize: 500,
    chunkOverlap: 50,
    embedModel: 'openai/text-embedding-3-small'
});

// Index everything
await fazt.services.rag.ingest('docs');
console.log('Indexed!');


// === API endpoint ===

// api/chat.js
module.exports = async (req) => {
    const { question } = req.body;

    if (!question) {
        return { status: 400, json: { error: 'Missing question' } };
    }

    const result = await fazt.services.rag.ask('docs', question);

    return {
        json: {
            answer: result.answer,
            sources: result.sources.map(s => ({
                file: s.metadata.source,
                preview: s.content.slice(0, 100) + '...'
            }))
        }
    };
};


// === Cron: Keep index fresh ===

// app.json
{
    "cron": [
        { "schedule": "0 * * * *", "handler": "api/sync-docs.js" }
    ]
}

// api/sync-docs.js
module.exports = async () => {
    await fazt.services.rag.sync('docs');
    return { json: { synced: true } };
};
```

## Example: Support Ticket Classifier

```javascript
// Index past tickets with their resolutions
await fazt.services.rag.create('tickets', {
    sources: [],  // We'll ingest programmatically
    embedModel: 'openai/text-embedding-3-small'
});

// Index historical tickets
for (const ticket of historicalTickets) {
    await fazt.services.rag.ingestText('tickets',
        `Issue: ${ticket.subject}\n\nResolution: ${ticket.resolution}`,
        { metadata: { ticketId: ticket.id, category: ticket.category } }
    );
}

// When new ticket arrives, find similar past tickets
async function suggestResolution(newTicket) {
    const similar = await fazt.services.rag.retrieve('tickets',
        newTicket.subject + '\n' + newTicket.body,
        { limit: 3 }
    );

    return similar.map(s => ({
        pastTicketId: s.metadata.ticketId,
        category: s.metadata.category,
        resolution: s.content.split('Resolution: ')[1],
        similarity: s.score
    }));
}
```

## Limits

| Limit | Default |
|-------|---------|
| Max pipelines per app | 10 |
| Max sources per pipeline | 100 |
| Max chunks per pipeline | 100,000 |
| Max chunk size | 2,000 chars |
| Max question length | 1,000 chars |
| Retrieve limit | 20 |

## Cost Estimation

RAG costs are primarily embedding API calls:

| Operation | Cost (text-embedding-3-small) |
|-----------|-------------------------------|
| Ingest 1,000 docs (~500 chunks each) | ~$0.05 |
| 1,000 questions | ~$0.02 |
| LLM answers (GPT-4) | ~$0.03 per answer |

For personal use, expect <$1/month.

## Philosophy Alignment

- **Composition**: RAG composes existing primitives (no new capabilities)
- **Zero config**: Good defaults, works immediately
- **Personal scale**: Optimized for <100k documents
- **JSON everywhere**: All data flows as JSON
- **Single DB**: Everything in SQLite (via vector store)

## Future Enhancements

Potential improvements (not in initial release):

1. **Hybrid Search**: Combine vector + full-text (Bleve) for better recall
2. **Reranking**: Use cross-encoder to rerank retrieved chunks
3. **Streaming Answers**: Stream LLM response as it generates
4. **Conversation Memory**: Automatic context window management
5. **Citation Extraction**: Parse which sources supported which claims
