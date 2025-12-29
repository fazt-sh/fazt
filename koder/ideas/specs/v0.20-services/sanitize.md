# Sanitize Service

## Summary

HTML and text sanitization primitives. Strip dangerous content from user input
to prevent XSS attacks. Security-critical - implemented in Go for reliability.

## Why a Service

Sanitization is:
- **Security-critical**: Get it wrong → XSS vulnerabilities
- **Subtle**: Many bypass techniques, edge cases
- **Universal**: Every app accepting user HTML needs this
- **Better in Go**: Battle-tested libraries, not JS regex

## Capabilities

| Operation | Description |
|-----------|-------------|
| `html` | Sanitize HTML, keep safe tags/attributes |
| `text` | Strip all HTML, return plain text |
| `markdown` | Sanitize markdown output |
| `url` | Validate and sanitize URLs |

## Usage

### Sanitize HTML

```javascript
const userInput = '<script>alert("xss")</script><p onclick="evil()">Hello <b>world</b></p>';

const safe = fazt.services.sanitize.html(userInput);
// '<p>Hello <b>world</b></p>'
// - Script tag removed
// - onclick handler removed
// - Safe tags preserved
```

### Custom Policies

```javascript
// Strict - text formatting only
const strict = fazt.services.sanitize.html(input, {
  policy: 'strict'
});
// Allows: b, i, em, strong, p, br

// Basic - common content tags
const basic = fazt.services.sanitize.html(input, {
  policy: 'basic'
});
// Allows: strict + a, ul, ol, li, blockquote, code, pre

// Rich - user-generated content
const rich = fazt.services.sanitize.html(input, {
  policy: 'rich'
});
// Allows: basic + img, h1-h6, table, hr

// Custom allowlist
const custom = fazt.services.sanitize.html(input, {
  allow: ['p', 'a', 'img'],
  allowAttrs: {
    'a': ['href', 'title'],
    'img': ['src', 'alt']
  }
});
```

### Strip All HTML

```javascript
const html = '<p>Hello <b>world</b>!</p>';

const text = fazt.services.sanitize.text(html);
// 'Hello world!'
```

### Sanitize Markdown Output

```javascript
// After rendering markdown, sanitize the HTML output
const markdown = '# Hello\n\n<script>bad</script>';
const html = fazt.services.markdown.render(markdown);
const safe = fazt.services.sanitize.markdown(html);
```

### Sanitize URLs

```javascript
fazt.services.sanitize.url('https://example.com/path')
// 'https://example.com/path' (valid)

fazt.services.sanitize.url('javascript:alert(1)')
// null (dangerous protocol)

fazt.services.sanitize.url('//evil.com', { requireAbsolute: true })
// null (protocol-relative not allowed)
```

## Policies

### Strict Policy

Minimal formatting, suitable for comments:

```javascript
{
  tags: ['b', 'i', 'em', 'strong', 'p', 'br', 'span'],
  attrs: {}  // No attributes allowed
}
```

### Basic Policy

Standard content, suitable for blog posts:

```javascript
{
  tags: ['b', 'i', 'em', 'strong', 'p', 'br', 'span', 'a', 'ul', 'ol', 'li',
         'blockquote', 'code', 'pre'],
  attrs: {
    'a': ['href', 'title', 'rel'],
    'code': ['class']  // For syntax highlighting
  },
  urlSchemes: ['http', 'https', 'mailto']
}
```

### Rich Policy

Full user content, suitable for CMS:

```javascript
{
  tags: ['b', 'i', 'em', 'strong', 'p', 'br', 'span', 'a', 'ul', 'ol', 'li',
         'blockquote', 'code', 'pre', 'img', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
         'table', 'thead', 'tbody', 'tr', 'th', 'td', 'hr', 'div', 'figure',
         'figcaption', 'video', 'audio', 'source'],
  attrs: {
    'a': ['href', 'title', 'rel', 'target'],
    'img': ['src', 'alt', 'title', 'width', 'height'],
    'video': ['src', 'controls', 'width', 'height'],
    'audio': ['src', 'controls'],
    'source': ['src', 'type'],
    '*': ['class', 'id']  // Allow class/id on all
  },
  urlSchemes: ['http', 'https', 'mailto', 'data']
}
```

## JS API

```javascript
fazt.services.sanitize.html(input, options?)
// options: { policy: 'strict'|'basic'|'rich', allow: [], allowAttrs: {} }
// Returns: string (sanitized HTML)

fazt.services.sanitize.text(input)
// Returns: string (plain text, all HTML removed)

fazt.services.sanitize.markdown(input, options?)
// Same as html() but tuned for markdown output
// Returns: string (sanitized HTML)

fazt.services.sanitize.url(input, options?)
// options: { requireAbsolute: boolean, allowedSchemes: [] }
// Returns: string | null (null if invalid/dangerous)
```

## HTTP Endpoint

Not exposed via HTTP. Sanitization is a JS-side operation, not an endpoint.

## Go Library

Uses `bluemonday` - battle-tested HTML sanitizer:

```go
import "github.com/microcosm-cc/bluemonday"

// Pre-built policies
var (
    strictPolicy = bluemonday.StrictPolicy()
    basicPolicy  = bluemonday.UGCPolicy()
    richPolicy   = bluemonday.NewPolicy()
)
```

## Security Considerations

- **Always sanitize before rendering**: Never trust user input
- **Sanitize on output, not input**: Store original, sanitize on display
- **Use appropriate policy**: Stricter is safer
- **URLs need separate handling**: `href` can contain `javascript:`
- **CSS can be dangerous**: Style attributes can exfiltrate data

## Common Patterns

### Comment System

```javascript
// User submits comment
const comment = request.body.comment;

// Store original
await fazt.storage.ds.insert('comments', {
  raw: comment,
  html: fazt.services.sanitize.html(comment, { policy: 'strict' }),
  createdAt: Date.now()
});

// Display sanitized version
```

### Rich Text Editor

```javascript
// User submits blog post
const content = request.body.content;

// Sanitize with rich policy
const safe = fazt.services.sanitize.html(content, { policy: 'rich' });

// Store both
await fazt.storage.ds.insert('posts', {
  raw: content,
  html: safe
});
```

### External Links

```javascript
// Sanitize user-provided URLs
const userUrl = request.body.website;
const safeUrl = fazt.services.sanitize.url(userUrl);

if (safeUrl) {
  await fazt.storage.ds.update('profiles', { id: userId }, {
    website: safeUrl
  });
}
```

## Limits

| Limit | Default |
|-------|---------|
| `maxInputSizeKB` | 512 |
| `maxTagDepth` | 100 |

## Implementation Notes

- ~50KB binary addition
- Pure Go (bluemonday has no CGO)
- Policies are compiled once, reused
- Streaming not supported (full document required)
