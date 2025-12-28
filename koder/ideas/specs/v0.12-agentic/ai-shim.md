# AI Shim

## Summary

`fazt.ai` provides a unified interface for calling LLM providers. Apps can
use OpenAI, Anthropic, Gemini, or Ollama without managing credentials or
handling provider-specific APIs.

## API

### Completion

```javascript
const result = await fazt.ai.complete('Summarize this: ...', {
    model: 'gpt-4',           // Optional, defaults to system default
    provider: 'openai',       // Optional, auto-detected from model
    temperature: 0.7,
    maxTokens: 1000
});

console.log(result.text);
console.log(result.usage);    // { prompt: 100, completion: 50 }
```

### Streaming

```javascript
await fazt.ai.stream('Write a story about...', {
    onChunk: (text) => {
        socket.send(text);    // Stream to client
    },
    onDone: (full) => {
        console.log('Complete:', full.text);
    }
});
```

### Embeddings

```javascript
const embedding = await fazt.ai.embed('semantic search query');
// Returns: { vector: [0.1, 0.2, ...], dimensions: 1536 }
```

## Provider Configuration

System-level config (set by admin):

```bash
fazt config set ai.openai.key sk-...
fazt config set ai.anthropic.key sk-ant-...
fazt config set ai.default openai/gpt-4
```

App-level config (in env):

```json
{
  "env": [
    { "name": "OPENAI_API_KEY", "required": true }
  ]
}
```

App keys override system keys.

## Supported Providers

| Provider | Models | Notes |
|----------|--------|-------|
| OpenAI | gpt-4, gpt-3.5-turbo | Default provider |
| Anthropic | claude-3-opus, claude-3-sonnet | |
| Google | gemini-pro, gemini-flash | |
| Ollama | llama3, mistral, etc. | Local inference |

## Permissions

Requires `ai:complete` permission in `app.json`:

```json
{
  "permissions": ["ai:complete"]
}
```

## Cost Tracking

```javascript
const result = await fazt.ai.complete('...');
console.log(result.cost);  // { usd: 0.003 }
```

Dashboard shows aggregated AI costs per app.
