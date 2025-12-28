# Search Service

## Summary

Full-text search over app content. Index documents from storage or VFS,
query with relevance ranking. Uses Bleve (pure Go search library).

## Capabilities

| Feature | Description |
|---------|-------------|
| Indexing | Index documents from ds collections or files |
| Full-text | Tokenized, stemmed text search |
| Relevance | Results ranked by match quality |
| Facets | Filter by field values |
| Highlight | Show matching snippets |

## Usage

### Index a Collection

```javascript
// Index all documents in a collection
await fazt.services.search.index('posts', {
  fields: ['title', 'body', 'tags']   // Fields to index
});
```

Indexing happens in background. New documents auto-indexed on insert.

### Query

```javascript
const results = await fazt.services.search.query('hello world', {
  collections: ['posts'],    // Which indexes to search
  limit: 20,
  offset: 0
});

// Returns:
// {
//   total: 42,
//   hits: [
//     { id: 'doc_123', score: 2.5, collection: 'posts', highlight: {...} },
//     ...
//   ]
// }
```

### Get Full Document

```javascript
// Search returns IDs, fetch full documents separately
const doc = await fazt.storage.ds.get('posts', hit.id);
```

## Index Configuration

### Basic

```javascript
await fazt.services.search.index('posts', {
  fields: ['title', 'body']
});
```

### With Field Weights

```javascript
await fazt.services.search.index('posts', {
  fields: {
    title: { boost: 2.0 },     // Title matches rank higher
    body: { boost: 1.0 },
    tags: { boost: 1.5 }
  }
});
```

### Index VFS Files

```javascript
// Index markdown files
await fazt.services.search.indexFiles('docs/**/*.md', {
  extractTitle: true,        // Use first H1 as title
  extractBody: true          // Full text content
});
```

### Index HTML Pages

For static sites with HTML pages:

```javascript
// Index all HTML files
await fazt.services.search.indexFiles('**/*.html', {
  stripTags: true,           // Extract text from HTML
  extractTitle: true         // Use <title> or first <h1>
});
```

Query results return file paths directly usable as URLs:

```javascript
const results = await fazt.services.search.query('pricing');
// {
//   hits: [
//     {
//       id: '/pricing.html',      // ← File path = URL
//       title: 'Pricing Plans',
//       highlight: '...our <mark>pricing</mark> is simple...'
//     },
//     {
//       id: '/docs/faq.html',
//       title: 'FAQ',
//       highlight: '...questions about <mark>pricing</mark>...'
//     }
//   ]
// }
```

Frontend usage:

```html
<ul id="results"></ul>
<script>
fetch(`/_services/search?q=${query}`)
  .then(r => r.json())
  .then(data => {
    data.hits.forEach(hit => {
      const li = document.createElement('li');
      li.innerHTML = `
        <a href="${hit.id}">${hit.title}</a>
        <p>${hit.highlight}</p>
      `;
      document.getElementById('results').appendChild(li);
    });
  });
</script>
```

## Query Options

```javascript
await fazt.services.search.query(term, {
  collections: ['posts', 'pages'],  // Default: all indexed
  limit: 20,                        // Default: 10
  offset: 0,

  // Filters
  filter: {
    category: 'tech',               // Exact match
    date: { $gte: '2024-01-01' }    // Range
  },

  // Facets
  facets: ['category', 'author'],   // Return counts per value

  // Highlighting
  highlight: true,                  // Include snippets
  highlightFields: ['body'],        // Which fields
  highlightSize: 150                // Snippet length
});
```

## Response Format

```javascript
{
  total: 42,                        // Total matches
  took: 12,                         // Milliseconds
  hits: [
    {
      id: 'doc_123',
      collection: 'posts',
      score: 2.5,
      highlight: {
        body: '...matching <mark>hello world</mark> text...'
      }
    }
  ],
  facets: {
    category: [
      { value: 'tech', count: 15 },
      { value: 'news', count: 12 }
    ]
  }
}
```

## Index Management

```javascript
// Rebuild index from scratch
await fazt.services.search.reindex('posts');

// Remove index
await fazt.services.search.dropIndex('posts');

// List indexes
const indexes = await fazt.services.search.indexes();
// [{ name: 'posts', docCount: 1234, size: '2.4 MB' }, ...]
```

## Auto-Indexing

Once an index is created, new documents are automatically indexed:

```javascript
// Setup once
await fazt.services.search.index('posts', { fields: ['title', 'body'] });

// Later, new docs auto-indexed
await fazt.storage.ds.insert('posts', {
  title: 'New Post',
  body: 'Content here...'
});
// Automatically added to search index
```

## HTTP Endpoint

### Search

```
GET /_services/search?q=hello+world&collections=posts,pages&limit=20
```

Response:

```json
{
  "total": 42,
  "hits": [
    { "id": "doc_123", "collection": "posts", "score": 2.5 }
  ]
}
```

### With filters

```
GET /_services/search?q=hello&filter.category=tech&facets=category,author
```

## JS API

```javascript
fazt.services.search.index(collection, options)
// options: { fields }
// Creates or updates index config

fazt.services.search.indexFiles(glob, options)
// options: { extractTitle, extractBody }
// Index VFS files

fazt.services.search.query(term, options)
// options: { collections, limit, offset, filter, facets, highlight }
// Returns: { total, hits, facets }

fazt.services.search.reindex(collection)
// Rebuild index

fazt.services.search.dropIndex(collection)
// Remove index

fazt.services.search.indexes()
// List all indexes
```

## Go Libraries

```go
import (
    "github.com/blevesearch/bleve/v2"
)
```

Bleve is pure Go, no CGO required.

## Storage

Indexes stored in SQLite:

```sql
CREATE TABLE svc_search_indexes (
    id TEXT PRIMARY KEY,
    app_uuid TEXT NOT NULL,
    collection TEXT NOT NULL,
    config TEXT NOT NULL,        -- JSON
    created_at INTEGER
);

CREATE TABLE svc_search_docs (
    id TEXT PRIMARY KEY,
    index_id TEXT NOT NULL,
    doc_id TEXT NOT NULL,
    content TEXT NOT NULL,       -- Indexed text
    FOREIGN KEY (index_id) REFERENCES svc_search_indexes(id)
);
```

Bleve index data stored as blobs or in VFS under `_search/`.

## Limits

| Limit | Default |
|-------|---------|
| `maxIndexesPerApp` | 10 |
| `maxDocsPerIndex` | 100,000 |
| `maxIndexSizeMB` | 100 |
| `maxQueryTime` | 5s |

## CLI

```bash
# List indexes
fazt services search list

# Show index info
fazt services search show posts

# Rebuild index
fazt services search reindex posts

# Search from CLI
fazt services search query "hello world" --collection posts
```

## Example: Blog Search

**Setup (run once):**

```javascript
// api/setup-search.js
module.exports = async () => {
  await fazt.services.search.index('posts', {
    fields: {
      title: { boost: 2.0 },
      body: { boost: 1.0 },
      tags: { boost: 1.5 }
    }
  });

  return { json: { ok: true } };
};
```

**Search endpoint:**

```javascript
// api/search.js
module.exports = async (req) => {
  const q = req.query.q;
  if (!q) {
    return { status: 400, json: { error: 'Missing query' } };
  }

  const results = await fazt.services.search.query(q, {
    collections: ['posts'],
    limit: 20,
    highlight: true
  });

  // Fetch full documents for hits
  const posts = await Promise.all(
    results.hits.map(async (hit) => {
      const post = await fazt.storage.ds.get('posts', hit.id);
      return {
        ...post,
        _score: hit.score,
        _highlight: hit.highlight
      };
    })
  );

  return {
    json: {
      total: results.total,
      posts
    }
  };
};
```

**Frontend:**

```html
<form action="/api/search" method="GET">
  <input name="q" placeholder="Search posts...">
  <button type="submit">Search</button>
</form>
```
