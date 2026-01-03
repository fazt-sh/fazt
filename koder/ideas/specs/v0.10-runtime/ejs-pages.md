# EJS Pages

## Summary

PHP/ERB-style pages with embedded JavaScript for quick prototyping. Files with
`.ejs` extension mix HTML with server-side code. Uses standard EJS syntax for
full editor support (syntax highlighting, snippets, etc.).

## Why

The `api/main.js` pattern is great for APIs but awkward for simple pages:

```javascript
// api/main.js - verbose for a simple page
module.exports = async function(request) {
  const items = await fazt.storage.ds.find('items');
  return {
    headers: { 'Content-Type': 'text/html' },
    body: `<html><body><ul>${items.map(i =>
      `<li>${i.name}</li>`).join('')}</ul></body></html>`
  };
};
```

With `.ejs` files:

```html
<!-- items.ejs - clean and readable -->
<html>
<body>
  <ul>
  <% const items = fazt.storage.ds.find('items'); %>
  <% for (const item of items) { %>
    <li><%= item.name %></li>
  <% } %>
  </ul>
</body>
</html>
```

**Philosophy alignment:**
- Single binary: EJS parser is ~200 lines, uses existing Goja runtime
- Zero-build: No compilation step, edit and refresh
- Events as spine: Page renders emit `http.request` events like normal

## Frontmatter

Optional YAML frontmatter between `---` markers:

```ejs
---
title: Dashboard
description: User dashboard with stats
layout: _layout.ejs
auth: required
cache: 5m
---
<h1><%= title %></h1>
```

### Available Fields

| Field         | Type     | Description                          |
|---------------|----------|--------------------------------------|
| `title`       | string   | Page title (available as variable)   |
| `description` | string   | Meta description                     |
| `layout`      | string   | Wrap page in layout partial          |
| `auth`        | string   | `required` or `optional`             |
| `cache`       | duration | Cache rendered output (e.g., `5m`)   |
| `contentType` | string   | Override response type               |
| `*`           | any      | Custom fields available as variables |

### Layout Wrapping

When `layout` is specified, the page content becomes `content` variable:

```ejs
<!-- _layout.ejs -->
<!DOCTYPE html>
<html>
<head><title><%= title %></title></head>
<body>
  <%- include('_nav.ejs') %>
  <main><%- content %></main>
</body>
</html>
```

```ejs
<!-- dashboard.ejs -->
---
title: Dashboard
layout: _layout.ejs
---
<h1>Welcome</h1>
<p>This becomes the "content" variable in layout</p>
```

### Auth Shorthand

```ejs
---
auth: required
---
<%# No need to check auth manually - 401 returned if not authenticated %>
<h1>Protected Page</h1>
```

Equivalent to:
```ejs
<%
  if (!request.user) {
    return { status: 401, redirect: '/login' };
  }
%>
```

### Response Caching

```ejs
---
cache: 5m
---
<%# Rendered HTML cached for 5 minutes %>
<%# Cache key includes: path + query + user (if auth) %>
```

Useful for pages with expensive queries but infrequent data changes.

## Syntax

### Code Blocks

| Delimiter | Purpose                | Example                 |
|-----------|------------------------|-------------------------|
| `<% %>`   | Execute code           | `<% const x = 1; %>`    |
| `<%= %>`  | Output (HTML escaped)  | `<%= user.name %>`      |
| `<%- %>`  | Output (raw, unescaped)| `<%- htmlContent %>`    |
| `<%# %>`  | Comment (not rendered) | `<%# TODO: fix this %>` |

### Examples

**Variables:**
```html
<% const user = fazt.storage.kv.get('user:123'); %>
<h1>Hello, <%= user.name %></h1>
```

**Conditionals:**
```html
<% if (user.isPremium) { %>
  <span class="badge">Premium</span>
<% } else { %>
  <a href="/upgrade">Upgrade now</a>
<% } %>
```

**Loops:**
```html
<ul>
<% for (const item of items) { %>
  <li><%= item.name %> - $<%= item.price %></li>
<% } %>
</ul>
```

**Raw output (careful with XSS):**
```html
<%- markdownToHtml(post.content) %>
```

## Routing

`.ejs` files are served by their path, just like static files:

```
my-app/
├── index.ejs         → /
├── about.ejs         → /about
├── dashboard.ejs     → /dashboard
└── blog/
    ├── index.ejs     → /blog
    └── post.ejs      → /blog/post
```

Query parameters available via `request`:

```html
<!-- /blog/post?id=123 -->
<% const post = fazt.storage.ds.get('posts', request.query.id); %>
<h1><%= post.title %></h1>
```

## Request Object

Same as serverless handlers:

```javascript
request = {
  method: 'GET',
  url: {
    pathname: '/dashboard',
    search: '?tab=stats',
    query: { tab: 'stats' }
  },
  headers: { ... },
  body: '...',
  json: { ... }
}
```

## Partials

Use `include()` for reusable fragments:

```html
<!-- layout.ejs -->
<!DOCTYPE html>
<html>
<head>
  <title><%= title %></title>
  <%- include('_head.ejs') %>
</head>
<body>
  <%- include('_nav.ejs', { user: currentUser }) %>
  <%- content %>
  <%- include('_footer.ejs') %>
</body>
</html>
```

```html
<!-- _nav.ejs (partial, starts with _) -->
<nav>
  <a href="/">Home</a>
  <% if (user) { %>
    <span>Welcome, <%= user.name %></span>
  <% } %>
</nav>
```

### Include Resolution

1. Relative to current file: `include('./partials/_header.ejs')`
2. Within app boundary only
3. Max depth: 10 (prevent infinite recursion)
4. Partials conventionally prefixed with `_`

### Passing Data to Partials

```html
<%- include('_card.ejs', { title: 'Hello', body: item.desc }) %>
```

Variables passed become available in the partial's scope.

## Full API Access

`.ejs` files have access to the complete `fazt.*` namespace:

```html
<%
  // Storage
  const user = fazt.storage.kv.get('user:' + request.query.id);
  const posts = fazt.storage.ds.find('posts', { author: user.id });

  // Rate limiting
  const { allowed } = fazt.limits.consume('page:' + request.ip, {
    limit: 100, window: '1m'
  });

  // Templates (for dynamic content)
  const emailBody = fazt.lib.template.render(emailTpl, { user });
%>
```

## Execution Model

**Synchronous, PHP-style.** All `fazt.*` calls block until complete:

```html
<%
  // These run sequentially, each blocks until done
  const user = fazt.storage.kv.get('user:123');
  const orders = fazt.storage.ds.find('orders', { userId: user.id });
  const stats = fazt.storage.ds.aggregate('orders', {
    sum: 'total', where: { userId: user.id }
  });
%>
```

No async/await needed (or supported). This is intentional:
- Simpler mental model (like PHP, ERB, JSP)
- Go handles async internally
- Pages are short-lived (30s timeout)

## Forms and POST

```html
<!-- login.ejs -->
<%
  let error = null;

  if (request.method === 'POST') {
    const { email, password } = request.json || {};
    const user = fazt.auth.verify(email, password);

    if (user) {
      // Set session and redirect
      return { redirect: '/dashboard', cookies: { session: user.token } };
    }
    error = 'Invalid credentials';
  }
%>

<form method="POST">
  <% if (error) { %>
    <div class="error"><%= error %></div>
  <% } %>
  <input name="email" type="email" required>
  <input name="password" type="password" required>
  <button type="submit">Login</button>
</form>
```

### Response Control

Return a response object to override default HTML output:

```html
<%
  if (!isAuthenticated) {
    return { redirect: '/login' };
  }

  if (request.query.format === 'json') {
    return { json: { data: items } };
  }
%>
```

## Error Handling

```html
<%
  try {
    const data = fazt.storage.ds.get('items', request.query.id);
    if (!data) {
      return { status: 404, body: 'Not found' };
    }
  } catch (e) {
    console.error('Failed to load:', e);
    return { status: 500, body: 'Server error' };
  }
%>
```

Unhandled exceptions return 500 with error message (dev mode) or generic
message (production).

## Resource Limits

Same as serverless handlers:

| Limit          | Value | Rationale          |
|----------------|-------|--------------------|
| Execution time | 30s   | Prevent hung pages |
| Memory         | 64MB  | Protect system     |
| Include depth  | 10    | Prevent recursion  |

## When to Use What

| Use Case          | Approach                 |
|-------------------|--------------------------|
| JSON API          | `api/main.js` handlers   |
| Data-driven pages | `.ejs` files             |
| Complex SPA       | Static HTML + JS + API   |
| Quick prototype   | `.ejs` files             |
| Background jobs   | Cron + handlers          |

## Implementation

### Parser

Simple state machine, ~200 lines:

```go
func Parse(source string) (Template, error) {
    // Returns parsed template with code blocks and static segments
}

func (t *Template) Render(vm *goja.Runtime, data map[string]any) (string, error) {
    // Execute code blocks, interpolate results into output
}
```

### VFS Integration

`.ejs` files stored in VFS like any other file. On request:

1. Load file from VFS
2. Parse into template (cacheable)
3. Create Goja VM with `fazt.*` bindings
4. Set `request` object
5. Execute template
6. Return HTML response

### Caching (Optional)

If template parsing proves slow (unlikely for small files):

```go
type CompiledCache struct {
    mu    sync.RWMutex
    cache map[string]*CachedTemplate
}

type CachedTemplate struct {
    Template *Template
    ModTime  time.Time
}
```

ModTime-based invalidation, same pattern as VFS file cache.

## Not Included

- **Layouts/inheritance**: Use `include()` for composition
- **Filters/helpers**: Use plain JS functions
- **Asset pipeline**: Static files served as-is
- **Hot reload**: Edit file, refresh browser

Keep it simple. For complex apps, use proper frontend frameworks.
