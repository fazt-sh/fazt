# Text Splitter

## Summary

Pure functions for splitting text into chunks with configurable size
and overlap. Essential for RAG pipelines, document indexing, and any
application that needs to process large documents in smaller pieces.

Inspired by LangChain's text splitters but implemented as pure Go with zero
external dependencies.

## Why lib (not services)

Text splitting is a pure function:
- Input: text string + options
- Output: array of strings
- No state, no I/O, no side effects
- Belongs in `fazt.lib` namespace

## Core Concept

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Original Document                             │
│  "Machine learning is a subset of artificial intelligence. It       │
│   enables computers to learn from data. Deep learning is a          │
│   subset of machine learning that uses neural networks..."          │
└─────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
                    ┌───────────────────────┐
                    │   fazt.lib.text.split │
                    │   chunkSize: 100      │
                    │   chunkOverlap: 20    │
                    └───────────────────────┘
                                │
        ┌───────────────────────┼───────────────────────┐
        ▼                       ▼                       ▼
┌───────────────┐       ┌───────────────┐       ┌───────────────┐
│   Chunk 1     │       │   Chunk 2     │       │   Chunk 3     │
│ "Machine      │       │ "from data.   │       │ "learning     │
│  learning is  │       │  Deep         │       │  that uses    │
│  a subset..." │       │  learning..." │       │  neural..."   │
└───────────────┘       └───────────────┘       └───────────────┘
        │                       │
        └───────┬───────────────┘
                │
         20 char overlap
         (maintains context)
```

## Usage

### Basic Split

```javascript
const chunks = fazt.lib.text.split(longDocument, {
    chunkSize: 1000,      // Target size in characters
    chunkOverlap: 200     // Overlap between chunks
});
// Returns: ["chunk1...", "chunk2...", "chunk3..."]
```

### With Custom Separators

```javascript
// Default separators: ["\n\n", "\n", " ", ""]
// Splits on paragraph first, then line, then word, then character

const chunks = fazt.lib.text.split(content, {
    chunkSize: 500,
    chunkOverlap: 50,
    separators: ["\n\n", "\n", ". ", " "]  // Custom hierarchy
});
```

### Keep Separators

```javascript
const chunks = fazt.lib.text.split(content, {
    chunkSize: 500,
    chunkOverlap: 50,
    keepSeparator: true  // Include separator in chunk
});
```

### Split Documents (with metadata)

```javascript
// Split array of LangChain-format documents
const docs = [
    { pageContent: "Long article about AI...", metadata: { source: "wiki" } },
    { pageContent: "Another long article...", metadata: { source: "blog" } }
];

const chunkedDocs = fazt.lib.text.splitDocuments(docs, {
    chunkSize: 500,
    chunkOverlap: 50
});
// Returns: [
//   { pageContent: "chunk1...", metadata: { source: "wiki" } },
//   { pageContent: "chunk2...", metadata: { source: "wiki" } },
//   { pageContent: "chunk1...", metadata: { source: "blog" } },
//   ...
// ]
```

## Algorithm

The recursive character splitter works as follows:

```
1. Try to split on first separator (e.g., "\n\n")
2. For each piece:
   - If piece < chunkSize: add to current chunk
   - If piece >= chunkSize: recursively split with next separator
3. Merge small pieces until they reach chunkSize
4. Maintain overlap by keeping last N characters from previous chunk
```

**Why recursive?** It preserves semantic boundaries:
- Prefers splitting on paragraphs over sentences
- Prefers sentences over words
- Only splits mid-word as last resort

## JS API

```javascript
// Basic split
fazt.lib.text.split(text, options?)
// options: { chunkSize, chunkOverlap, separators, keepSeparator, lenFunc }
// Returns: string[]

// Split with documents (preserves metadata)
fazt.lib.text.splitDocuments(documents, options?)
// documents: { pageContent, metadata }[]
// Returns: { pageContent, metadata }[]

// Create documents from text array
fazt.lib.text.createDocuments(texts, metadatas?, options?)
// texts: string[]
// metadatas: object[] (optional, same length as texts)
// Returns: { pageContent, metadata }[]

// Get character count (default length function)
fazt.lib.text.countChars(text)
// Returns: number (UTF-8 aware)
```

## Options

| Option          | Type     | Default                   | Description                     |
| --------------- | -------- | ------------------------- | ------------------------------- |
| `chunkSize`     | number   | 4000                      | Target chunk size in characters |
| `chunkOverlap`  | number   | 200                       | Overlap between chunks          |
| `separators`    | string[] | `["\n\n", "\n", " ", ""]` | Split hierarchy                 |
| `keepSeparator` | boolean  | false                     | Include separator in output     |
| `lenFunc`       | function | `countChars`              | Custom length function          |

## Examples

### RAG Pipeline

```javascript
// 1. Load document
const content = await fazt.storage.s3.get('documents/manual.txt');

// 2. Split into chunks
const chunks = fazt.lib.text.split(content, {
    chunkSize: 500,
    chunkOverlap: 50
});

// 3. Create documents with metadata
const docs = chunks.map((chunk, i) => ({
    pageContent: chunk,
    metadata: { source: 'manual.txt', chunk: i }
}));

// 4. Index in vector store
await fazt.storage.vector.addDocuments('manual', docs);
```

### Code Splitting

```javascript
// For code, use language-aware separators
const codeChunks = fazt.lib.text.split(sourceCode, {
    chunkSize: 1000,
    chunkOverlap: 100,
    separators: [
        "\nclass ",      // Class definitions
        "\nfunction ",   // Function definitions
        "\nconst ",      // Constants
        "\n\n",          // Blank lines
        "\n",            // Any line
        " "              // Words
    ]
});
```

### Batch Processing

```javascript
// Process multiple files
const files = await fazt.storage.s3.list('docs/*.md');
const allDocs = [];

for (const file of files) {
    const content = await fazt.storage.s3.get(file.path);
    const docs = fazt.lib.text.splitDocuments([
        { pageContent: content, metadata: { source: file.path } }
    ], { chunkSize: 500 });
    allDocs.push(...docs);
}

// Index all at once
await fazt.storage.vector.addDocuments('docs', allDocs);
```

## Implementation Notes

### Character Counting

Uses UTF-8 aware character counting (not bytes):

```go
func countChars(s string) int {
    return utf8.RuneCountInString(s)
}
```

### Merge Algorithm

```go
func mergeSplits(splits []string, separator string, chunkSize, overlap int) []string {
    var chunks []string
    var current []string
    total := 0

    for _, split := range splits {
        splitLen := countChars(split)

        // Would adding this split exceed chunk size?
        if total + splitLen > chunkSize && len(current) > 0 {
            // Emit current chunk
            chunks = append(chunks, joinWithSeparator(current, separator))

            // Keep overlap from end of current
            current, total = keepOverlap(current, separator, overlap)
        }

        current = append(current, split)
        total += splitLen
    }

    // Emit final chunk
    if len(current) > 0 {
        chunks = append(chunks, joinWithSeparator(current, separator))
    }

    return chunks
}
```

## CLI

```bash
# Split a file
fazt text split document.txt --chunk-size 500 --overlap 50

# Split and output as JSON
fazt text split document.txt --json

# Split multiple files
fazt text split docs/*.md --chunk-size 1000 > chunks.json
```

## Limits

| Limit          | Default       |
| -------------- | ------------- |
| Max input size | 10 MB         |
| Max chunk size | 100,000 chars |
| Max overlap    | chunkSize / 2 |

## Reference Implementation

LinGoose (MIT licensed) has a clean, tested Go implementation:
- **File**: `textsplitter/recursiveTextSplitter.go` (~110 lines)
- **Repo**: https://github.com/henomis/lingoose
- **Evaluation**: See `koder/ideas/lite-extractions.md` (EXTRACT verdict)

The algorithm is pure Go with stdlib-only dependencies. When implementing
Fazt's text splitter, reference LinGoose's battle-tested logic.

## Comparison with LangChain

This implementation is based on LangChain's `RecursiveCharacterTextSplitter`:

| Feature             | LangChain | Fazt               |
| ------------------- | --------- | ------------------ |
| Recursive splitting | ✓         | ✓                  |
| Custom separators   | ✓         | ✓                  |
| Chunk overlap       | ✓         | ✓                  |
| Keep separator      | ✓         | ✓                  |
| Document metadata   | ✓         | ✓                  |
| Token counting      | ✓         | Future             |
| Markdown-aware      | ✓         | Future (see STASH) |

The API is intentionally compatible so code can be ported between LangChain
and Fazt with minimal changes.
