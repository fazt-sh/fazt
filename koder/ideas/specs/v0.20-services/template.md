# Template

## Summary

Mustache-style templating library for dynamic content generation. Pure function,
no state. Designed for email templates, notifications, PDF generation, and
simple server-rendered pages.

## Why

Apps need templates for:
- Transactional emails ("Your order {{orderId}} shipped")
- Notification messages (SMS, push)
- PDF invoices/receipts
- Simple server-rendered pages

Without a primitive, apps either:
- Use string interpolation (unsafe, no escaping)
- Bundle a JS template library (bloat)
- Write custom logic (repeated effort)

**Philosophy alignment:**
- Pure Go: Uses `text/template` internally, ~150 lines wrapper
- JSON everywhere: Data is JSON, output is string
- No state: Pure function, no side effects

## API

### Basic Rendering

```javascript
// Simple variable substitution
const result = fazt.lib.template.render(
  'Hello {{name}}!',
  { name: 'Alice' }
);
// "Hello Alice!"

// Nested objects
const result = fazt.lib.template.render(
  'Order #{{order.id}} for {{customer.name}}',
  {
    order: { id: '12345' },
    customer: { name: 'Alice' }
  }
);
// "Order #12345 for Alice"
```

### Loops

```javascript
const template = `
Items:
{{#items}}
- {{name}}: ${{price}}
{{/items}}
`;

const result = fazt.lib.template.render(template, {
  items: [
    { name: 'Widget', price: '9.99' },
    { name: 'Gadget', price: '19.99' }
  ]
});
// "Items:\n- Widget: $9.99\n- Gadget: $19.99\n"
```

### Conditionals

```javascript
const template = `
{{#isPremium}}
Welcome, Premium member!
{{/isPremium}}
{{^isPremium}}
Upgrade to Premium for more features.
{{/isPremium}}
`;

const result = fazt.lib.template.render(template, { isPremium: true });
// "Welcome, Premium member!"
```

### Compiled Templates

For repeated rendering, compile once:

```javascript
// Compile template (validates syntax, returns handle)
const compiled = fazt.lib.template.compile(`
  <h1>Hello {{name}}</h1>
  <p>Your balance: ${{balance}}</p>
`);

// Render multiple times with different data
const html1 = fazt.lib.template.renderCompiled(compiled, { name: 'Alice', balance: '100.00' });
const html2 = fazt.lib.template.renderCompiled(compiled, { name: 'Bob', balance: '250.00' });
```

### Escaping

```javascript
// HTML escaping (default for {{var}})
const result = fazt.lib.template.render(
  '<p>{{content}}</p>',
  { content: '<script>alert("xss")</script>' }
);
// "<p>&lt;script&gt;alert(&quot;xss&quot;)&lt;/script&gt;</p>"

// Raw output (triple braces, use carefully)
const result = fazt.lib.template.render(
  '<div>{{{htmlContent}}}</div>',
  { htmlContent: '<strong>Bold</strong>' }
);
// "<div><strong>Bold</strong></div>"

// Manual escaping utility
const safe = fazt.lib.template.escape('<script>', 'html');
// "&lt;script&gt;"

const urlSafe = fazt.lib.template.escape('hello world', 'url');
// "hello%20world"
```

## Syntax Reference

| Syntax | Description | Example |
|--------|-------------|---------|
| `{{var}}` | Variable (HTML escaped) | `{{name}}` → `Alice` |
| `{{obj.prop}}` | Nested property | `{{user.email}}` |
| `{{{var}}}` | Raw output (no escaping) | `{{{html}}}` |
| `{{#list}}...{{/list}}` | Loop over array | Iterate items |
| `{{#bool}}...{{/bool}}` | Conditional (truthy) | Show if true |
| `{{^bool}}...{{/bool}}` | Inverted (falsy) | Show if false |
| `{{! comment }}` | Comment (not rendered) | `{{! TODO }}` |

## Common Patterns

### Email Template

```javascript
const emailTemplate = `
Hi {{customer.name}},

Your order #{{order.id}} has shipped!

Items:
{{#order.items}}
  - {{name}} (x{{qty}}): ${{subtotal}}
{{/order.items}}

Total: ${{order.total}}

Track your package: {{tracking.url}}

Thanks,
{{company.name}}
`;

const emailBody = fazt.lib.template.render(emailTemplate, {
  customer: { name: 'Alice' },
  order: {
    id: 'ORD-123',
    items: [...],
    total: '59.99'
  },
  tracking: { url: 'https://...' },
  company: { name: 'Acme Inc' }
});

await fazt.services.notify.send({
  to: customer.id,
  channels: ['email'],
  title: 'Your order has shipped!',
  body: emailBody
});
```

### Markdown + Template + CSS Pipeline

```javascript
// 1. Markdown template with variables
const invoiceTemplate = `
# Invoice {{invoice.number}}

**Customer:** {{customer.name}}
**Date:** {{invoice.date}}

| Item | Qty | Price |
|------|-----|-------|
{{#items}}
| {{name}} | {{qty}} | ${{price}} |
{{/items}}

**Total: ${{invoice.total}}**
`;

// 2. Fill template with data
const markdown = fazt.lib.template.render(invoiceTemplate, invoiceData);

// 3. Render markdown to HTML with CSS
const html = fazt.services.markdown.render(markdown, {
  css: '/styles/invoice.css'
});

// 4. Optional: Generate PDF
const pdf = await fazt.services.pdf.fromHtml(html);
```

### Notification Templates

```javascript
// Store templates in KV or define inline
const templates = {
  'order.shipped': {
    title: 'Order Shipped',
    body: 'Your order #{{orderId}} is on its way!'
  },
  'payment.received': {
    title: 'Payment Received',
    body: 'We received ${{amount}} for invoice #{{invoiceId}}'
  }
};

async function notify(userId, templateKey, data) {
  const template = templates[templateKey];
  await fazt.services.notify.send({
    to: userId,
    title: fazt.lib.template.render(template.title, data),
    body: fazt.lib.template.render(template.body, data)
  });
}
```

### Simple Landing Page

```javascript
// api/landing.js
const pageTemplate = `
<!DOCTYPE html>
<html>
<head>
  <title>{{title}}</title>
  <style>{{css}}</style>
</head>
<body>
  <h1>{{headline}}</h1>
  <p>{{description}}</p>
  {{#features}}
  <div class="feature">
    <h3>{{name}}</h3>
    <p>{{desc}}</p>
  </div>
  {{/features}}
</body>
</html>
`;

module.exports = async function(request) {
  const html = fazt.lib.template.render(pageTemplate, {
    title: 'My Product',
    css: await fazt.fs.read('/styles/landing.css'),
    headline: 'Build faster',
    description: '...',
    features: [...]
  });

  return {
    headers: { 'Content-Type': 'text/html' },
    body: html
  };
};
```

## Error Handling

```javascript
// Invalid template syntax throws
try {
  fazt.lib.template.compile('Hello {{name}');  // Missing closing braces
} catch (e) {
  // { code: 'TEMPLATE_SYNTAX', message: 'Unclosed tag at position 6' }
}

// Missing variables render as empty string (no error)
const result = fazt.lib.template.render('Hello {{name}}', {});
// "Hello "

// Strict mode throws on missing variables
const result = fazt.lib.template.render('Hello {{name}}', {}, { strict: true });
// throws { code: 'MISSING_VAR', variable: 'name' }
```

## Implementation Notes

Uses Go's `text/template` with Mustache-compatible delimiters:

```go
func Render(templateStr string, data map[string]any) (string, error) {
    tmpl, err := template.New("").
        Delims("{{", "}}").
        Funcs(defaultFuncs).
        Parse(templateStr)
    if err != nil {
        return "", err
    }

    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, data); err != nil {
        return "", err
    }
    return buf.String(), nil
}
```

**Binary impact:** ~150 lines wrapper around stdlib. No external deps.

## What This Is NOT

- **Not a frontend framework** - Use React/Vue/Svelte for complex UIs
- **Not a full template engine** - No inheritance, partials, macros
- **Not for untrusted templates** - Templates come from app code, not users

Keep it simple. For complex rendering, use a proper frontend framework.
