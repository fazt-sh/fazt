# Sandbox (Safe Code Execution)

## Summary

Isolated execution environment for running untrusted code. Agents can
generate and execute code safely. Apps can evaluate dynamic expressions.
Strict resource limits and no access to system primitives by default.

## Why Runtime-Level

Sandbox extends the JS runtime with isolated execution:
- Agents generate code that needs to run safely
- User-submitted code (formulas, templates) needs isolation
- Dynamic behavior without compromising system security
- Same Goja engine, stricter constraints

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Normal Runtime                          │
│   (full fazt.* access, network, storage, etc.)               │
├─────────────────────────────────────────────────────────────┤
│                        Sandbox                                │
│   ┌─────────────────────────────────────────────────────┐   │
│   │  • No fazt.* access (by default)                     │   │
│   │  • No network                                        │   │
│   │  • No filesystem                                     │   │
│   │  • No require()                                      │   │
│   │  • CPU/memory limits                                 │   │
│   │  • Timeout enforced                                  │   │
│   └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## Usage

### Basic Execution

```javascript
const result = await fazt.sandbox.exec({
  code: `
    const items = JSON.parse(input);
    return items.filter(x => x > 10).map(x => x * 2);
  `,
  input: '[5, 15, 20, 8, 30]'
});

// result: [30, 40, 60]
```

### With Context

```javascript
const result = await fazt.sandbox.exec({
  code: `
    return data.items.reduce((sum, item) => sum + item.price, 0);
  `,
  context: {
    data: {
      items: [
        { name: 'A', price: 10 },
        { name: 'B', price: 20 }
      ]
    }
  }
});

// result: 30
```

### Timeout and Limits

```javascript
const result = await fazt.sandbox.exec({
  code: heavyComputation,
  timeout: 5000,           // 5 seconds max
  memoryLimit: '32MB'      // Memory limit
});
```

## Input/Output

### Input Methods

```javascript
// Via 'input' variable (string, parsed in code)
await fazt.sandbox.exec({
  code: `return JSON.parse(input).length`,
  input: '[1,2,3]'
});

// Via 'context' object (direct access)
await fazt.sandbox.exec({
  code: `return context.items.length`,
  context: { items: [1, 2, 3] }
});

// Both
await fazt.sandbox.exec({
  code: `return context.multiplier * JSON.parse(input)`,
  input: '5',
  context: { multiplier: 10 }
});
// result: 50
```

### Output

Return value is serialized to JSON:

```javascript
// Primitives
return 42;                  // 42
return "hello";             // "hello"
return true;                // true

// Objects/Arrays
return { a: 1, b: 2 };      // {"a":1,"b":2}
return [1, 2, 3];           // [1,2,3]

// Undefined/null
return undefined;           // null
return null;                // null

// Functions (not serializable)
return () => {};            // Error
```

## Available APIs

### Default (Minimal)

```javascript
// These are available by default:
JSON.parse(), JSON.stringify()
parseInt(), parseFloat()
String(), Number(), Boolean(), Array(), Object()
Math.*
Date (read-only, no system time modification)
console.log() (captured, not printed)
```

### No Access (by default)

```javascript
// These are NOT available:
fazt.*                      // No system access
require()                   // No imports
fetch()                     // No network
process                     // No process info
global                      // No global object
eval()                      // No nested eval
Function()                  // No function constructor
```

## Capability Grants

Explicitly grant specific capabilities:

```javascript
const result = await fazt.sandbox.exec({
  code: `
    const data = await fetch('https://api.example.com/data');
    return data.json();
  `,
  capabilities: ['net:fetch'],   // Grant network access
  allowedDomains: ['api.example.com']
});
```

### Available Capabilities

| Capability | Grants |
|------------|--------|
| `net:fetch` | `fetch()` function (via proxy) |
| `storage:kv` | `fazt.storage.kv.*` |
| `storage:ds` | `fazt.storage.ds.*` |
| `time:now` | Accurate `Date.now()` |
| `crypto:random` | `crypto.getRandomValues()` |

```javascript
await fazt.sandbox.exec({
  code: `
    const value = await fazt.storage.kv.get('key');
    return value * 2;
  `,
  capabilities: ['storage:kv']
});
```

## Error Handling

```javascript
try {
  const result = await fazt.sandbox.exec({
    code: `throw new Error('oops')`,
    timeout: 5000
  });
} catch (e) {
  if (e.code === 'SANDBOX_ERROR') {
    console.log(e.message);      // "oops"
    console.log(e.line);         // 1
    console.log(e.column);       // 7
  }
  if (e.code === 'SANDBOX_TIMEOUT') {
    console.log('Code took too long');
  }
  if (e.code === 'SANDBOX_MEMORY') {
    console.log('Memory limit exceeded');
  }
}
```

## Console Capture

```javascript
const result = await fazt.sandbox.exec({
  code: `
    console.log('step 1');
    console.log('step 2');
    return 'done';
  `
});

// result.value: 'done'
// result.logs: ['step 1', 'step 2']
```

## Use Cases

### 1. Agent-Generated Code

```javascript
// Agent generates code based on user request
const userRequest = "calculate total with 10% tax";
const generatedCode = await fazt.ai.complete(
  `Generate JS code to: ${userRequest}.
   Input: array of {price, qty}. Return total.`
);

// Execute safely
const result = await fazt.sandbox.exec({
  code: generatedCode,
  context: {
    items: [
      { price: 10, qty: 2 },
      { price: 5, qty: 3 }
    ]
  },
  timeout: 1000
});
```

### 2. User Formulas

```javascript
// User defines custom formula for their dashboard
const userFormula = user.settings.customFormula;
// e.g., "data.sales - data.costs"

const result = await fazt.sandbox.exec({
  code: `return ${userFormula}`,
  context: {
    data: {
      sales: 10000,
      costs: 7500
    }
  }
});
// result: 2500
```

### 3. Template Expressions

```javascript
// Dynamic template with expressions
const template = "Hello {{name}}, your total is ${{total * 1.1}}";

const rendered = await fazt.sandbox.exec({
  code: `
    return template.replace(/\\{\\{(.+?)\\}\\}/g, (_, expr) => {
      return eval(expr);  // Safe because we're in sandbox
    });
  `,
  context: {
    template,
    name: 'Alice',
    total: 100
  }
});
// result: "Hello Alice, your total is $110"
```

### 4. Data Transformation

```javascript
// User-defined transformation
const transform = userConfig.transform;
// e.g., "items.map(x => ({ ...x, total: x.price * x.qty }))"

const result = await fazt.sandbox.exec({
  code: `return ${transform}`,
  context: {
    items: rawData
  }
});
```

## JS API

```javascript
fazt.sandbox.exec(options)
// options:
//   code: string (required)
//   input: string (available as 'input' variable)
//   context: object (available as 'context' variable)
//   timeout: number (ms, default 5000)
//   memoryLimit: string (default '64MB')
//   capabilities: string[]
//   allowedDomains: string[] (if net:fetch granted)
//
// Returns: { value, logs }
// Throws: SandboxError with code, message, line, column

fazt.sandbox.validate(code)
// Static analysis - checks for disallowed patterns
// Returns: { valid: boolean, errors: string[] }
```

## Static Validation

Pre-check code before execution:

```javascript
const check = await fazt.sandbox.validate(`
  const x = eval('1+1');  // Not allowed
`);

// { valid: false, errors: ['eval() is not allowed'] }
```

Patterns detected:
- `eval()`, `Function()`
- `require()`, `import`
- Global assignments
- Infinite loop patterns (heuristic)

## Limits

| Limit | Default | Max |
|-------|---------|-----|
| `timeout` | 5s | 60s |
| `memoryLimit` | 64 MB | 256 MB |
| `codeSize` | 100 KB | 1 MB |
| `outputSize` | 1 MB | 10 MB |
| `stackDepth` | 100 | 1000 |

## CLI

```bash
# Execute code
fazt sandbox exec 'return 1 + 1'

# With input
fazt sandbox exec 'return JSON.parse(input).length' --input '[1,2,3]'

# From file
fazt sandbox exec --file transform.js --input data.json

# Validate
fazt sandbox validate 'const x = eval("1")'
```

## Implementation Notes

- Uses Goja (Go JS engine) with restricted global object
- Memory limit via Go runtime memory tracking
- Timeout via context cancellation
- No goroutine/concurrency in sandbox code
- Capability grants inject specific functions into global scope
