# Testing Patterns for Fazt Apps

Comprehensive testing framework for validating data flow, UI/UX, and state management.

## Architecture

```
┌─────────────────────────┐
│  Test Harness           │  Control panel with test controls
│  (standalone.html)      │
│                         │
│  ┌─────────────────┐   │
│  │  App (iframe)   │   │  Your fazt app
│  │                 │   │
│  │  + Test Runner  │◄──┼──► WebSocket Server (port 7777)
│  │    Injected     │   │    - Collects results
│  └─────────────────┘   │    - Real-time dashboard
└─────────────────────────┘    - Saves logs
```

## Files Structure

```
app-name/
├── test/
│   ├── server.js         # WebSocket server + dashboard
│   ├── runner.js         # Browser test runner
│   ├── standalone.html   # Test harness (self-contained)
│   ├── package.json      # Dependencies
│   └── README.md         # Documentation
```

## Test Runner (runner.js)

Browser-side test runner that simulates user workflows:

```javascript
class TestRunner {
  constructor() {
    this.results = []
    this.ws = null
  }

  // Connect to WebSocket server
  async connect(wsUrl = 'ws://192.168.64.3:7777') {
    return new Promise((resolve) => {
      this.ws = new WebSocket(wsUrl)
      this.ws.onopen = () => {
        this.send({ type: 'connected', userAgent: navigator.userAgent })
        resolve()
      }
      this.ws.onerror = () => resolve() // Continue standalone
    })
  }

  // Simulate clicks
  async click(selector) {
    const el = document.querySelector(selector)
    if (!el) throw new Error(`Element not found: ${selector}`)
    el.click()
    return this.wait(100)
  }

  // Set input values
  async setValue(selector, value) {
    const el = document.querySelector(selector)
    if (!el) throw new Error(`Element not found: ${selector}`)
    el.value = value
    el.dispatchEvent(new Event('input', { bubbles: true }))
    return this.wait(50)
  }

  // Select dropdown
  async select(selector, value) {
    const el = document.querySelector(selector)
    if (!el) throw new Error(`Element not found: ${selector}`)
    el.value = value
    el.dispatchEvent(new Event('change', { bubbles: true }))
    return this.wait(50)
  }

  // Wait for condition
  async waitFor(conditionFn, timeout = 5000, message = 'Condition not met') {
    const start = Date.now()
    while (Date.now() - start < timeout) {
      if (conditionFn()) return
      await this.wait(100)
    }
    throw new Error(`Timeout: ${message}`)
  }

  // Assertions
  assert(condition, message) {
    if (!condition) throw new Error(`Assertion failed: ${message}`)
  }

  assertExists(selector, message) {
    const el = document.querySelector(selector)
    this.assert(el !== null, message || `Element should exist: ${selector}`)
    return el
  }

  assertText(selector, expected, message) {
    const el = this.assertExists(selector)
    const actual = el.textContent.trim()
    this.assert(
      actual.includes(expected),
      message || `Expected "${expected}" in "${actual}"`
    )
  }

  // Access Vue state
  getState() {
    const app = document.querySelector('#app').__vue_app__
    const rootComponent = app._instance
    if (!rootComponent) throw new Error('Root component not found')
    return rootComponent.ctx // Returns reactive state
  }

  // Run individual test
  async test(name, fn) {
    const test = { name, status: 'running', startedAt: Date.now() }
    console.log(`🧪 ${name}`)
    this.send({ type: 'test-start', test: name })

    try {
      await fn()
      test.status = 'passed'
      test.duration = Date.now() - test.startedAt
      console.log(`✅ ${name} (${test.duration}ms)`)
      this.send({ type: 'test-pass', test: name, duration: test.duration })
    } catch (error) {
      test.status = 'failed'
      test.error = error.message
      test.duration = Date.now() - test.startedAt
      console.error(`❌ ${name}: ${error.message}`)
      this.send({ type: 'test-fail', test: name, error: error.message })
    }

    this.results.push(test)
  }

  wait(ms) {
    return new Promise(resolve => setTimeout(resolve, ms))
  }

  send(data) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data))
    }
  }
}
```

## Example Test Workflows

### Test 1: App Initialization
```javascript
await runner.test('App loads successfully', async () => {
  runner.assertExists('#app', 'App root should exist')
  runner.assertExists('h1', 'Header should exist')
  await runner.wait(500) // Wait for initial data load
})
```

### Test 2: Data Operations
```javascript
await runner.test('Can create item', async () => {
  const state = runner.getState()
  const initialCount = state.items.value.length

  // Open form
  await runner.click('button[aria-label="Add item"]')
  await runner.waitFor(() => document.querySelector('form'))

  // Fill and submit
  await runner.setValue('input[name="title"]', 'Test Item')
  await runner.click('button[type="submit"]')

  // Verify
  await runner.waitFor(
    () => state.items.value.length > initialCount,
    2000,
    'Item should be added'
  )

  runner.assert(
    state.items.value.length === initialCount + 1,
    'Item count should increase'
  )
})
```

### Test 3: State Updates
```javascript
await runner.test('State updates correctly', async () => {
  const state = runner.getState()

  // Perform action
  await runner.click('.increment-button')

  // Verify state changed
  await runner.waitFor(
    () => state.counter.value === 1,
    1000,
    'Counter should increment'
  )
})
```

### Test 4: API Integration
```javascript
await runner.test('API persists data', async () => {
  const state = runner.getState()

  // Create item
  await runner.click('button.add')
  await runner.setValue('input', 'API Test')
  await runner.click('button[type="submit"]')

  // Wait for API call
  await runner.wait(500)

  // Refresh and verify persistence
  location.reload()
  await runner.wait(1000)

  const newState = runner.getState()
  runner.assert(
    newState.items.value.some(i => i.name === 'API Test'),
    'Item should persist after reload'
  )
})
```

### Test 5: UI Interactions
```javascript
await runner.test('Modal opens and closes', async () => {
  // Open modal
  await runner.click('button[aria-label="Settings"]')
  await runner.waitFor(
    () => document.querySelector('[role="dialog"]'),
    1000
  )

  // Close with backdrop
  await runner.click('.bg-black\\/50')
  await runner.wait(100)

  runner.assertNotExists('[role="dialog"]', 'Modal should close')
})
```

## WebSocket Server (server.js)

Node.js server for test orchestration:

```javascript
const WebSocket = require('ws')
const http = require('http')

class TestServer {
  constructor() {
    this.clients = new Map()
    this.sessions = new Map()
  }

  start() {
    this.wss = new WebSocket.Server({ port: 7777 })

    this.wss.on('connection', (ws) => {
      const clientId = this.generateId()
      const session = {
        id: this.generateId(),
        clientId,
        startedAt: Date.now(),
        tests: [],
        status: 'connected'
      }

      this.clients.set(clientId, { ws, session })

      ws.on('message', (data) => {
        const msg = JSON.parse(data.toString())
        this.handleMessage(clientId, msg, session)
      })
    })

    console.log('🧪 Test Server: ws://192.168.64.3:7777')
  }

  handleMessage(clientId, msg, session) {
    switch (msg.type) {
      case 'test-start':
        session.tests.push({ name: msg.test, status: 'running' })
        break
      case 'test-pass':
        const passTest = session.tests.find(t => t.name === msg.test)
        if (passTest) {
          passTest.status = 'passed'
          passTest.duration = msg.duration
        }
        break
      case 'test-fail':
        const failTest = session.tests.find(t => t.name === msg.test)
        if (failTest) {
          failTest.status = 'failed'
          failTest.error = msg.error
        }
        break
      case 'summary':
        session.status = 'completed'
        session.summary = msg
        this.saveResults(session)
        break
    }
  }

  saveResults(session) {
    const filename = `test-${session.id}-${Date.now()}.json`
    fs.writeFileSync(`results/${filename}`, JSON.stringify(session, null, 2))
  }
}

new TestServer().start()
```

## Test Harness (standalone.html)

Self-contained HTML file that loads app in iframe and injects tests:

```html
<!DOCTYPE html>
<html>
<head>
  <title>Test Harness</title>
  <style>
    /* Control panel styling */
  </style>
</head>
<body>
  <div class="controls">
    <button onclick="runTests()">▶️ Run All Tests</button>
    <button onclick="refreshApp()">🔄 Refresh App</button>
  </div>

  <div class="log" id="log"></div>

  <iframe id="app" src="http://app.192.168.64.3.nip.io:8080"></iframe>

  <script>
    async function runTests() {
      const iframe = document.getElementById('app')

      // Inject test runner
      const script = iframe.contentDocument.createElement('script')
      script.textContent = /* TestRunner code */
      iframe.contentDocument.head.appendChild(script)

      // Monitor console output
      iframe.contentWindow.console.log = function(...args) {
        addLog(args.join(' '))
      }
    }
  </script>
</body>
</html>
```

## Setup Instructions

### 1. Create test directory

```bash
mkdir -p servers/<peer>/<app-name>/test
cd servers/<peer>/<app-name>/test
```

### 2. Install dependencies

```bash
npm init -y
npm install ws
```

### 3. Start test server

```bash
node server.js
```

### 4. Serve test harness

```bash
python3 -m http.server 7778
```

### 5. Run tests

Open: `http://192.168.64.3:7778/standalone.html`
Click: "▶️ Run All Tests"

## Dashboard

WebSocket server provides real-time dashboard at `http://192.168.64.3:7777`:

- Test status (PASSED/FAILED)
- Total, passed, failed counts
- Duration for each test
- Error messages
- Auto-refreshes every 2 seconds

## Common Test Patterns

### Testing Data Flow

```javascript
await runner.test('Data persists to server', async () => {
  const state = runner.getState()

  // Create data
  await createItem('Test')

  // Verify in state
  runner.assert(state.items.value.length > 0)

  // Verify in database (via debug endpoint)
  const response = await fetch('/_fazt/storage')
  const data = await response.json()
  runner.assert(data.documents.some(d => d.name === 'Test'))
})
```

### Testing State Management

```javascript
await runner.test('State updates reactively', async () => {
  const state = runner.getState()
  const initial = state.counter.value

  await runner.click('.increment')

  runner.assert(state.counter.value === initial + 1)
})
```

### Testing Error Handling

```javascript
await runner.test('Shows error on invalid input', async () => {
  await runner.setValue('input[type="email"]', 'invalid')
  await runner.click('button[type="submit"]')

  await runner.waitFor(
    () => document.querySelector('.error-message'),
    1000
  )
})
```

### Testing Cascading Updates

```javascript
await runner.test('Cascading updates work', async () => {
  const state = runner.getState()

  // Create transaction (should update category total)
  await createTransaction({ category: 'Food', amount: 50 })

  const category = state.categories.value.find(c => c.name === 'Food')
  runner.assert(category.totalSpent === 50, 'Category total should update')
})
```

## Debugging Tests

### View logs in real-time

```bash
tail -f /tmp/test-server.log
```

### Check test results

```bash
ls -lt test/results/ | head -5
cat test/results/test-*.json | jq
```

### Access debug endpoints

```
http://app.192.168.64.3.nip.io:8080/_fazt/info
http://app.192.168.64.3.nip.io:8080/_fazt/storage
http://app.192.168.64.3.nip.io:8080/_fazt/errors
```

## Best Practices

1. **Test data flow first** - Verify data persists to server
2. **Test state management** - Confirm reactive updates work
3. **Test UI interactions** - Verify clicks, forms, modals
4. **Test error cases** - Check error handling works
5. **Test edge cases** - Empty states, boundaries, etc.

6. **Use descriptive test names** - Clearly state what's being tested
7. **Wait for async operations** - Use `waitFor()` for conditions
8. **Access Vue state directly** - Use `getState()` for assertions
9. **Monitor WebSocket** - Use dashboard for real-time feedback
10. **Save test results** - Review JSON logs after runs

## Integration with /fazt-app

When building apps with `/fazt-app`:

1. Build the app normally
2. Add testing framework (copy from template)
3. Write test workflows for key features
4. Run tests before deploying to production
5. Use test insights to improve app

Testing validates:
- ✅ Data saved to server (not localStorage)
- ✅ State management working correctly
- ✅ UI interactions functioning
- ✅ API endpoints responding
- ✅ Cascading updates happening
- ✅ Error handling in place

## Example: Full Test Suite

```javascript
async function runAll() {
  await this.test('App loads', async () => {
    this.assertExists('#app')
  })

  await this.test('Has initial data', async () => {
    const state = this.getState()
    this.assert(state.items.value.length >= 0)
  })

  await this.test('Can create item', async () => {
    // ... creation test
  })

  await this.test('Can edit item', async () => {
    // ... edit test
  })

  await this.test('Can delete item', async () => {
    // ... delete test
  })

  await this.test('Data persists', async () => {
    // ... persistence test
  })

  this.printSummary()
}
```

## Reference Implementation

See CashFlow app (`servers/<peer>/cashflow/test/`) for complete working example:
- 9 comprehensive tests
- Full test runner implementation
- WebSocket server with dashboard
- Standalone test harness
- Real-time monitoring
- Result persistence
