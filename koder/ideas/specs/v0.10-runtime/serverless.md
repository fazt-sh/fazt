# Serverless v2

## Summary

Serverless v2 moves function code to a dedicated `api/` folder and adds a
`require()` shim for local imports. This prevents collisions with frontend
code and enables modular backend development.

## Rationale

### The Problem (v0.7)

Entry point was `main.js` in the root:
- Collides with frontend `js/main.js`
- No code organization
- Single file becomes unwieldy

### The Solution (v0.10)

Dedicated `api/` folder:

```
my-app/
├── index.html
├── styles.css
├── js/
│   └── main.js         # Frontend code
└── api/
    ├── main.js         # Serverless entry point
    ├── db.js           # Database helpers
    └── utils.js        # Shared utilities
```

## Request Handling

### Entry Point

All requests to `/api/*` route to `api/main.js`:

```javascript
// api/main.js
module.exports = async function(request) {
    const path = request.url.pathname;

    if (path === '/api/users') {
        return handleUsers(request);
    }

    return { status: 404, body: 'Not found' };
};
```

### Request Object

```javascript
{
    method: 'POST',
    url: {
        pathname: '/api/users',
        search: '?page=1',
        query: { page: '1' }
    },
    headers: {
        'content-type': 'application/json',
        'authorization': 'Bearer ...'
    },
    body: '{"name": "Alice"}',
    json: { name: 'Alice' }  // Parsed if Content-Type is JSON
}
```

### Response Object

```javascript
// Simple response
return { body: 'Hello' };

// Full response
return {
    status: 201,
    headers: { 'x-custom': 'value' },
    body: JSON.stringify({ id: 123 })
};

// JSON shorthand
return { json: { id: 123 } };

// Redirect
return { redirect: '/api/success' };
```

## Local Imports

### The `require()` Shim

```javascript
// api/main.js
const db = require('./db.js');
const { validate } = require('./utils.js');

module.exports = async function(request) {
    const user = request.json;
    validate(user);
    await db.insert('users', user);
    return { json: { success: true } };
};
```

### Resolution Rules

1. Relative paths only: `./file.js` or `../file.js`
2. Must be within `api/` folder
3. No traversal outside app boundary
4. File must exist in VFS

### Security

```javascript
// Allowed
require('./db.js');
require('./helpers/utils.js');

// Blocked (outside api/)
require('../index.html');
require('/etc/passwd');
require('../../other-app/api/secrets.js');
```

## Module System

### Exports Pattern

```javascript
// api/db.js
async function insert(table, data) {
    return fazt.storage.ds.insert(table, data);
}

async function find(table, query) {
    return fazt.storage.ds.find(table, query);
}

module.exports = { insert, find };
```

### Single Export

```javascript
// api/validator.js
module.exports = function validate(data) {
    if (!data.email) throw new Error('Email required');
};
```

## Error Handling

### Thrown Errors

```javascript
module.exports = async function(request) {
    if (!request.json.email) {
        throw new Error('Email required');
    }
    // ...
};

// Results in 500 response with error message
```

### Custom Errors

```javascript
module.exports = async function(request) {
    if (!request.json.email) {
        return {
            status: 400,
            json: { error: 'Email required' }
        };
    }
    // ...
};
```

## Resource Limits

| Limit | Value | Rationale |
|-------|-------|-----------|
| Execution time | 30s | Prevent hung requests |
| Memory | 64MB | Protect overall system |
| File size | 1MB | Prevent giant scripts |
| Require depth | 10 | Prevent circular imports |

## Invocation Modes

### HTTP Request

```
POST /api/users
Content-Type: application/json
{"name": "Alice"}
```

### CLI Trigger

```bash
fazt app run app_x9z2k --input '{"name":"Alice"}'
```

### Cron Trigger

See `cron.md` for scheduled invocations.
