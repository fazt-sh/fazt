# Standard Library

## Summary

Fazt embeds ES5-compatible builds of essential utilities directly into the
binary. Apps can `require('lodash')` without npm install or bundlers.

## Rationale

### The Problem

AI agents and casual developers face friction:
- Need to run `npm install` before deploying
- Need a build step for bundling
- Need to understand package management

### The Solution

Pre-bundle the top 10 utilities. Zero-build deployment.

```javascript
// Just works. No npm. No webpack.
const _ = require('lodash');
const cheerio = require('cheerio');
const { v4: uuid } = require('uuid');
```

## Included Libraries

| Library     | Version | Use Case                      |
| ----------- | ------- | ----------------------------- |
| `lodash`    | 4.x     | Array/object manipulation     |
| `cheerio`   | 1.x     | HTML parsing and manipulation |
| `uuid`      | 9.x     | Generate unique identifiers   |
| `zod`       | 3.x     | Schema validation             |
| `marked`    | 9.x     | Markdown to HTML              |
| `dayjs`     | 1.x     | Date manipulation             |
| `validator` | 13.x    | String validation             |

## Usage

### lodash

```javascript
const _ = require('lodash');

const users = [
    { name: 'Alice', age: 30 },
    { name: 'Bob', age: 25 }
];

const sorted = _.sortBy(users, 'age');
const grouped = _.groupBy(users, u => u.age > 27 ? 'senior' : 'junior');
```

### cheerio

```javascript
const cheerio = require('cheerio');

const html = '<div class="post"><h1>Title</h1></div>';
const $ = cheerio.load(html);

const title = $('h1').text();  // "Title"
$('.post').addClass('featured');
```

### uuid

```javascript
const { v4: uuid } = require('uuid');

const id = uuid();  // "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"
```

### zod

```javascript
const { z } = require('zod');

const UserSchema = z.object({
    email: z.string().email(),
    age: z.number().min(0).max(120)
});

const result = UserSchema.safeParse(request.json);
if (!result.success) {
    return { status: 400, json: result.error };
}
```

### marked

```javascript
const { marked } = require('marked');

const markdown = '# Hello\n\nThis is **bold**.';
const html = marked(markdown);
// "<h1>Hello</h1><p>This is <strong>bold</strong>.</p>"
```

### dayjs

```javascript
const dayjs = require('dayjs');

const now = dayjs();
const formatted = now.format('YYYY-MM-DD');
const nextWeek = now.add(7, 'day');
```

### validator

```javascript
const validator = require('validator');

validator.isEmail('foo@bar.com');     // true
validator.isURL('https://foo.com');   // true
validator.isUUID('abc-123');          // false
```

## Implementation

### Embedding

Libraries are compiled to ES5 and embedded in the binary:

```go
//go:embed stdlib/lodash.min.js
var lodashSource string

//go:embed stdlib/cheerio.min.js
var cheerioSource string

var stdlib = map[string]string{
    "lodash":  lodashSource,
    "cheerio": cheerioSource,
    // ...
}
```

### Resolution

When `require('name')` is called:

1. Check if `name` is in stdlib map
2. If yes: return pre-compiled source
3. If no: attempt local file resolution
4. If no local file: throw error

### Virtual Modules

Stdlib modules are served from memory with sub-millisecond load time:

```go
func (r *Runtime) Require(name string) (goja.Value, error) {
    if source, ok := stdlib[name]; ok {
        return r.vm.RunString(source)
    }
    return r.requireLocal(name)
}
```

## Maintenance

### Update Process

1. Download new library version
2. Compile to ES5 (for Goja compatibility)
3. Minify
4. Embed in binary
5. Release new Fazt version

### Security Updates

If a stdlib library has a vulnerability:
1. Patch the embedded version
2. Release hotfix binary
3. Users run `fazt proc upgrade`

### Version Locking

Stdlib versions are tied to Fazt versions:

| Fazt | lodash  | cheerio | uuid  |
| ---- | ------- | ------- | ----- |
| 0.10 | 4.17.21 | 1.0.0   | 9.0.0 |
| 0.11 | 4.17.22 | 1.0.1   | 9.0.1 |

## Limitations

1. **No npm**: Can't install arbitrary packages
2. **ES5 Only**: Modern syntax may not work
3. **Curated List**: Only included libraries available
4. **Version Lock**: Can't pick specific versions

## Go-Backed Utilities

Some utilities are implemented in Go for performance and exposed to JS:

### fazt.json.get (gjson)

Fast JSON path queries without full unmarshal. Uses
[gjson](https://github.com/tidwall/gjson) - zero deps, ~2k lines.

```javascript
const data = '{"user":{"name":"Alice","tags":["admin","active"]}}';

// Path query
fazt.json.get(data, "user.name");           // "Alice"
fazt.json.get(data, "user.tags.0");         // "admin"
fazt.json.get(data, "user.tags.#");         // 2 (array length)

// Wildcards and filters
fazt.json.get(data, "user.tags.#(=admin)"); // "admin"

// Multi-path
fazt.json.getMany(data, ["user.name", "user.tags.#"]);
// ["Alice", 2]
```

**Use case**: Query large JSON blobs stored in SQLite without loading entire
document into JS memory. Path expressions evaluate in Go, only matched values
cross the boundary.

### fazt.expr (go-expr)

Safe expression evaluation for dynamic rules. Uses
[expr-lang/expr](https://github.com/expr-lang/expr) - zero deps, ~200KB.

```javascript
// Compile once at config load, evaluate per-request
const rule = fazt.expr.compile('user.Role in ["admin", "mod"]');

// Evaluate with context
const allowed = fazt.expr.run(rule, {
    user: { Role: "admin", Plan: "pro" }
});  // true

// One-shot evaluation (compiles each time)
fazt.expr.eval('len(items) > 0 && total >= 100', {
    items: [1, 2, 3],
    total: 150
});  // true
```

**Built-in functions**: `len`, `all`, `any`, `filter`, `map`, `max`, `min`,
`upper`, `lower`, `trim`, `now`, `duration`

**Use cases**:
- Dynamic routing: `{"path": "/admin/*", "expr": "user.Role == 'admin'"}`
- Webhook filters: `{"event": "push", "expr": "branch == 'main'"}`
- Access control: `{"resource": "billing", "expr": "user.Plan in ['pro']"}`
- Cron conditions: `{"expr": "hour(now()) >= 9 && hour(now()) < 17"}`

**Why expr**: Type-safe compilation catches errors before runtime. Bytecode
VM is fast for repeated evaluation. No arbitrary code execution risk.

## Open Questions

1. **Which libraries to include?** Start with 7, add based on demand
2. **How to handle version conflicts?** Lock to Fazt version
3. **User-supplied libraries?** Maybe v0.12 with `api/node_modules/`
