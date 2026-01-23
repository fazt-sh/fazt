# Plan: Fazt DevTools

**Status**: Draft
**Priority**: High
**Version Target**: v0.11.x

## Problem Statement

Building fazt apps has been error-prone due to:
1. Fragmented observability (server logs separate from client)
2. No structured way for LLM agents to understand app state
3. Testing requires manual browser interaction
4. State management bugs are subtle and hard to diagnose
5. No unified feedback loop for rapid iteration
6. Apps need to be "build-free, but buildable" (work raw AND built)
7. Need robust patterns that scale to large multi-page apps

## Design Philosophy: Build-Free, But Buildable

Fazt apps MUST work in two modes with the SAME source code:

```
Source Code (identical files)
         │
         ├──► Serve directly (no build)
         │    - python -m http.server
         │    - Any static server
         │    - Constrained environments (Pi, phones)
         │    - Future embedded LLM harness
         │
         └──► Build with Vite/Bun
              - Tree shaking, minification
              - Build-time error catching
              - Optimized production bundles
```

This enables:
- Rapid iteration (serve raw, no build step)
- Production optimization (build when needed)
- Deployment anywhere (no Node.js requirement)
- Future embedded LLM can build apps on constrained devices

## Goals

1. **Unified Observability**: Single stream of all events (server + client + tests)
2. **LLM-Optimized**: Structured JSON output for agent consumption
3. **Zero Setup**: Works automatically for any fazt app
4. **Tight Feedback Loop**: Change → rebuild → test → see results instantly
5. **Injectable Analysis**: Library of scripts for different analysis needs

## Non-Goals

- Browser extension (humans can use web UI)
- Complex UI framework (keep it simple)
- Production monitoring (this is for development)

---

## App Architecture Standards

### Import Maps (Required)

All fazt apps use import maps for clean, CDN-backed imports:

```html
<!-- index.html -->
<script type="importmap">
{
  "imports": {
    "vue": "https://unpkg.com/vue@3/dist/vue.esm-browser.prod.js",
    "pinia": "https://unpkg.com/pinia@2/dist/pinia.esm-browser.js",
    "vue-router": "https://unpkg.com/vue-router@4/dist/vue-router.esm-browser.js"
  }
}
</script>
```

Benefits:
- Clean imports: `import { ref } from 'vue'` (not long CDN URLs)
- Works raw (browser resolves via importmap)
- Works built (Vite/Bun resolves to node_modules)
- Easy to swap CDN providers

### Directory Structure (Multi-Page Apps)

```
app-name/
├── manifest.json           # App metadata
├── index.html              # Entry point with importmap
├── src/
│   ├── main.js             # App initialization
│   ├── router.js           # Client-side routing
│   ├── stores/             # Pinia stores (state management)
│   │   ├── index.js        # Store exports
│   │   ├── auth.js         # Auth state
│   │   └── [feature].js    # Feature-specific stores
│   ├── pages/              # Route components (full pages)
│   │   ├── Home.js
│   │   ├── Settings.js
│   │   └── [Feature]/
│   │       ├── Index.js
│   │       └── Detail.js
│   ├── components/         # Reusable components
│   │   ├── ui/             # Generic UI (Button, Modal, Card)
│   │   │   ├── Button.js
│   │   │   ├── Modal.js
│   │   │   └── index.js
│   │   └── [feature]/      # Feature-specific components
│   │       └── TransactionCard.js
│   └── lib/                # Utilities
│       ├── api.js          # API client
│       ├── session.js      # Session management
│       └── utils.js        # Helpers
├── api/                    # Serverless functions
│   └── main.js
└── static/                 # Static assets
    └── icons/
```

### State Management: Pinia

Use Pinia (Vue's official state management). Works without build:

```javascript
// src/stores/transactions.js
import { defineStore } from 'pinia'
import { api } from '../lib/api.js'

export const useTransactionStore = defineStore('transactions', {
  state: () => ({
    items: [],
    loading: false,
    error: null
  }),

  getters: {
    expenses: (state) => state.items.filter(t => t.type === 'expense'),
    income: (state) => state.items.filter(t => t.type === 'income'),
    total: (state) => state.items.reduce((sum, t) =>
      sum + (t.type === 'income' ? t.amount : -t.amount), 0
    )
  },

  actions: {
    async load() {
      this.loading = true
      this.error = null
      try {
        const data = await api.get('/api/transactions')
        this.items = data.transactions
      } catch (e) {
        this.error = e.message
      } finally {
        this.loading = false
      }
    },

    async create(transaction) {
      const created = await api.post('/api/transactions', transaction)
      this.items.push(created)
      return created
    }
  }
})
```

### Component Pattern

Components as plain JS with template strings:

```javascript
// src/components/ui/Button.js
export const Button = {
  props: {
    variant: { type: String, default: 'primary' },
    size: { type: String, default: 'md' },
    loading: { type: Boolean, default: false },
    disabled: { type: Boolean, default: false }
  },

  emits: ['click'],

  template: `
    <button
      :class="[
        'btn',
        'btn-' + variant,
        'btn-' + size,
        { 'btn-loading': loading }
      ]"
      :disabled="disabled || loading"
      @click="$emit('click', $event)"
    >
      <span v-if="loading" class="spinner"></span>
      <slot></slot>
    </button>
  `
}
```

### Page Pattern

Pages use stores and components:

```javascript
// src/pages/Transactions.js
import { useTransactionStore } from '../stores/transactions.js'
import { TransactionCard } from '../components/transactions/TransactionCard.js'
import { Button } from '../components/ui/Button.js'

export const TransactionsPage = {
  components: { TransactionCard, Button },

  setup() {
    const store = useTransactionStore()

    onMounted(() => {
      store.load()
    })

    return { store }
  },

  template: `
    <div class="page transactions-page">
      <header class="page-header">
        <h1>Transactions</h1>
        <Button @click="showAddModal = true">Add</Button>
      </header>

      <div v-if="store.loading" class="loading">Loading...</div>
      <div v-else-if="store.error" class="error">{{ store.error }}</div>
      <div v-else class="transaction-list">
        <TransactionCard
          v-for="txn in store.items"
          :key="txn.id"
          :transaction="txn"
        />
      </div>
    </div>
  `
}
```

### Shared Component Library (Future)

Goal: Build components once, use across all apps.

```javascript
// Future: import from shared library
import { Button, Modal, Card } from '@fazt/ui'
import { useSession, useApi } from '@fazt/core'
```

For now, copy patterns. Later, extract to shared package.

---

## What Already Exists in Fazt

**Debug Mode** (`internal/debug/debug.go`):
- `FAZT_DEBUG=1` enables structured logging
- Categories: `storage`, `runtime`, `sql`
- Functions: `debug.StorageOp()`, `debug.RuntimeReq()`, `debug.SQL()`

**Existing /_fazt Endpoints** (`internal/handlers/agent_handler.go`):
| Endpoint | Purpose |
|----------|---------|
| `/_fazt/info` | App metadata, file count, storage stats |
| `/_fazt/storage` | List all KV keys with sizes |
| `/_fazt/storage/{key}` | Get specific KV value |
| `/_fazt/snapshot` | Create named snapshot |
| `/_fazt/restore/{name}` | Restore from snapshot |
| `/_fazt/snapshots` | List available snapshots |
| `/_fazt/logs` | Recent execution logs |
| `/_fazt/errors` | Error logs only |

**Storage API** (`internal/storage/`):
- `fazt.storage.kv` - Key-value with TTL
- `fazt.storage.ds` - Document store with MongoDB-style queries
- `fazt.storage.s3` - Blob storage

**Runtime Globals** (`internal/runtime/`):
- `request` - HTTP request object
- `respond()` - Response helper
- `console` - Logging (persisted to `site_logs`)
- `fazt.app`, `fazt.env`, `fazt.log`, `fazt.storage`

**Execution Logging** (`site_logs` table):
- All console.* and fazt.log.* calls persisted
- Retrieved via `/_fazt/logs` and `/_fazt/errors`

## What's Missing (To Build)

```
┌─────────────────────────────────────────────────────────────┐
│                         FAZT SERVER                          │
│                                                              │
│  EXISTING:                      NEW (DevTools):              │
│  ┌──────────────────────┐      ┌──────────────────────────┐ │
│  │ /_fazt/info          │      │ /_fazt/stream            │ │
│  │ /_fazt/storage       │      │   - SSE real-time events │ │
│  │ /_fazt/logs          │      │   - Unified server+client│ │
│  │ /_fazt/errors        │      │                          │ │
│  │ /_fazt/snapshot      │──────│ /_fazt/events            │ │
│  │ /_fazt/restore       │      │   - POST from client JS  │ │
│  │ /_fazt/snapshots     │      │   - Aggregates to stream │ │
│  │                      │      │                          │ │
│  │ debug.StorageOp()    │      │ /_fazt/scripts           │ │
│  │ debug.RuntimeReq()   │      │   - Injectable JS library│ │
│  │ site_logs table      │      │   - test-runner.js       │ │
│  └──────────────────────┘      │   - state-inspector.js   │ │
│                                │   - error-catcher.js     │ │
│                                └──────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## API Design

### GET /_fazt/stream

Server-Sent Events (SSE) endpoint for unified logging.

```bash
curl -N "http://app.domain/_fazt/stream?session=xyz"
```

Event format:
```json
{"ts":1706123456789,"type":"storage","op":"insert","collection":"txn","rows":1,"ms":4.2}
{"ts":1706123456790,"type":"api","method":"POST","path":"/api/transactions","status":201,"ms":12}
{"ts":1706123456800,"type":"test","name":"Create transaction","status":"pass","ms":234}
{"ts":1706123456850,"type":"state","key":"transactions","length":5}
{"ts":1706123456900,"type":"error","message":"TypeError: x is undefined","stack":"...","source":"client"}
```

Event types:
- `storage` - ds/kv/s3 operations
- `api` - HTTP requests to /api/*
- `test` - Test runner events
- `state` - Vue state changes
- `error` - Errors (server or client)
- `perf` - Performance metrics
- `event` - User interactions (clicks, inputs)
- `log` - Console.log from client

### POST /_fazt/events

Receive events from injected client-side scripts.

```javascript
// Client-side script sends:
fetch('/_fazt/events', {
  method: 'POST',
  body: JSON.stringify({
    type: 'state',
    key: 'transactions',
    length: 5,
    session: 'sun-teal-kiwi'
  })
})
```

### GET /_fazt/scripts/:name

Serve injectable scripts.

```bash
# Get test runner
curl http://app.domain/_fazt/scripts/test-runner.js

# Get all scripts combined
curl http://app.domain/_fazt/scripts/all.js

# Get specific bundle
curl http://app.domain/_fazt/scripts/bundle.js?include=test-runner,state-inspector
```

### GET /_fazt/devtools

Web UI for human developers (optional, low priority).

## Injectable Scripts

### 1. test-runner.js

Workflow testing with assertions.

```javascript
// Injected into app, reports to /_fazt/events
class TestRunner {
  async test(name, fn) { ... }
  assert(condition, message) { ... }
  assertExists(selector) { ... }
  click(selector) { ... }
  // Reports: {"type":"test","name":"...","status":"pass/fail","ms":...}
}
```

### 2. state-inspector.js

Monitor Vue reactive state.

```javascript
// Hooks into Vue reactivity system
// Reports state changes to stream
function watchState(component) {
  watch(() => component.categories, (newVal) => {
    report({ type: 'state', key: 'categories', length: newVal.length })
  })
}
```

### 3. error-catcher.js

Global error handling.

```javascript
window.onerror = (msg, url, line, col, error) => {
  report({ type: 'error', message: msg, stack: error?.stack, source: 'client' })
}
window.onunhandledrejection = (event) => {
  report({ type: 'error', message: event.reason?.message, source: 'promise' })
}
```

### 4. storage-monitor.js

Hook into fazt.storage calls (client-side).

```javascript
// Wrap fazt.storage methods to log calls
const originalInsert = fazt.storage.ds.insert
fazt.storage.ds.insert = function(collection, doc) {
  report({ type: 'storage', op: 'insert', collection, source: 'client' })
  return originalInsert.apply(this, arguments)
}
```

### 5. perf-monitor.js

Performance metrics.

```javascript
// Report Web Vitals
new PerformanceObserver((list) => {
  for (const entry of list.getEntries()) {
    report({ type: 'perf', metric: entry.name, value: entry.value })
  }
}).observe({ type: 'largest-contentful-paint', buffered: true })
```

### 6. dom-analyzer.js

DOM structure and accessibility checks.

```javascript
function analyzeDOM() {
  const issues = []
  // Check for missing alt text, aria labels, etc.
  document.querySelectorAll('img:not([alt])').forEach(img => {
    issues.push({ type: 'a11y', issue: 'missing-alt', element: img.outerHTML })
  })
  return issues
}
```

## Implementation Phases

### Phase 0: App Architecture Templates (NOW)
Already done in `.claude/skills/fazt-app/templates/`:
- [x] Import maps pattern
- [x] Multi-page structure (router, stores, pages, components)
- [x] Pinia state management
- [x] API client with session
- [ ] Example page templates
- [ ] Example component library (ui/)

### Phase 1: Real-Time Streaming (v0.11.0)
Build on existing `/_fazt/logs` and `debug.StorageOp()`:
- [ ] `/_fazt/stream` SSE endpoint (real-time version of logs)
- [ ] `/_fazt/events` POST endpoint (receive client-side events)
- [ ] Hook `debug.StorageOp()` into stream (already logs, just broadcast)
- [ ] Hook `debug.RuntimeReq()` into stream
- [ ] Event buffer with configurable retention

### Phase 2: Injectable Scripts (v0.11.1)
- [ ] `/_fazt/scripts/:name` endpoint (serve from embedded assets)
- [ ] test-runner.js - workflow testing, reports to `/_fazt/events`
- [ ] error-catcher.js - global error handler
- [ ] state-inspector.js - Vue state → stream

### Phase 3: App Testing Framework (v0.11.2)
- [ ] storage-monitor.js - hook client-side fazt.storage calls
- [ ] perf-monitor.js - Web Vitals reporting
- [ ] Standardized test workflow for all fazt apps
- [ ] agent-browser integration guide

### Phase 4: Shared Component Library (v0.12.x)
- [ ] Extract common components to `@fazt/ui`
- [ ] Extract common utilities to `@fazt/core`
- [ ] CDN hosting for shared packages
- [ ] Apps can import shared components

### Phase 5: Web UI (optional, v0.12.x)
- [ ] `/_fazt/devtools` web interface (for humans)
- [ ] Real-time event viewer
- [ ] State tree inspector
- [ ] Visual test runner

## Leveraging Existing Systems

### Already Have: Debug Logging
```go
// internal/debug/debug.go - already logs storage ops
debug.StorageOp("insert", appID, collection, query, rows, duration)
debug.RuntimeReq(reqID, app, path, status, duration)
```
**Action**: Broadcast these to SSE stream (not just stderr)

### Already Have: Execution Logs
```go
// internal/runtime/handler.go - persists to site_logs
persistLogs(appID, result.Logs, result.Error)
```
**Action**: Also send to SSE stream in real-time

### Already Have: Agent Endpoints
```
/_fazt/info     → App metadata
/_fazt/storage  → KV keys
/_fazt/logs     → Historical logs
/_fazt/errors   → Error logs
```
**Action**: Add `/_fazt/stream` for real-time, `/_fazt/events` for client → server

### Storage Operations
Current: Logged to stderr with `[DEBUG storage]`
New: Also emitted as events to stream

### Runtime Errors
Current: Logged to stderr
New: Also emitted as events, include stack traces

## LLM Agent Workflow

```bash
# 1. Start watching stream (in background)
curl -N "http://cashflow.192.168.64.3.nip.io:8080/_fazt/stream" &

# 2. Open app with agent-browser
agent-browser open http://cashflow.192.168.64.3.nip.io:8080

# 3. Inject test runner
agent-browser eval "$(curl -s http://cashflow.192.168.64.3.nip.io:8080/_fazt/scripts/test-runner.js)"

# 4. Run tests
agent-browser eval "window.testRunner.runAll()"

# 5. Stream shows:
{"type":"test","name":"App loads","status":"pass","ms":234}
{"type":"storage","op":"find","collection":"categories","rows":12,"ms":2.1}
{"type":"test","name":"Create transaction","status":"pass","ms":567}
{"type":"storage","op":"insert","collection":"transactions","rows":1,"ms":4.5}
{"type":"storage","op":"update","collection":"categories","$inc":true,"ms":3.2}

# 6. If something fails, agent sees exactly what happened:
{"type":"test","name":"Category updates","status":"fail","error":"Expected 50, got 0"}
{"type":"storage","op":"update","collection":"categories","rows":0,"ms":1.1}
# ^ Agent can see: update returned 0 rows - query didn't match!
```

## Success Criteria

1. **Single curl command** shows all app activity
2. **LLM agent** can diagnose issues from stream alone
3. **Test failures** include enough context to fix
4. **No manual browser DevTools** needed for debugging
5. **Works on any fazt app** without app-specific setup

## Open Questions

1. Should stream require authentication?
2. How long to buffer events? (memory concern)
3. Should we support WebSocket in addition to SSE?
4. How to handle high-frequency events (throttling)?

## Related Documents

- `koder/ideas/specs/v0.10-runtime/` - Runtime enhancements
- `CLAUDE.md` - Debug mode documentation
- `.claude/skills/fazt-app/` - App building patterns

## Notes

This plan focuses on **LLM agent experience** over human UI. The web interface
is optional and low priority. The core value is the unified stream that gives
agents complete visibility into app behavior.
