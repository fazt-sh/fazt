# PDF Service

## Summary

Generate PDFs from HTML and CSS. Invoices, receipts, reports, certificates—
anything expressible as HTML. Uses WASM internally (v0.10 primitive) for
rendering. No JavaScript execution—static HTML/CSS only.

## Capabilities

| Operation  | Description                      |
| ---------- | -------------------------------- |
| `fromHtml` | Render HTML string to PDF        |
| `fromFile` | Render HTML file from VFS to PDF |
| `fromUrl`  | Render internal app URL to PDF   |
| `merge`    | Combine multiple PDFs into one   |
| `info`     | Get PDF metadata (pages, size)   |

## What It Can Render

**Supported:**
- HTML5 semantic elements
- CSS3 (flexbox, grid, positioning)
- Inline styles and `<style>` blocks
- Images from VFS (referenced by path)
- Web fonts (embedded or from VFS)
- Tables, lists, headers
- Page breaks via CSS (`page-break-before`, `page-break-after`)
- Headers/footers via CSS (`@page` rules)

**Not Supported:**
- JavaScript execution
- External URLs (images, fonts must be local)
- Interactive elements (forms, buttons)
- CSS animations/transitions

## Usage

### From HTML String

```javascript
const html = `
  <html>
    <head>
      <style>
        body { font-family: sans-serif; margin: 2cm; }
        .header { border-bottom: 1px solid #ccc; }
        .amount { font-size: 24px; font-weight: bold; }
      </style>
    </head>
    <body>
      <div class="header">
        <h1>Invoice #12345</h1>
      </div>
      <p>Amount due: <span class="amount">$150.00</span></p>
    </body>
  </html>
`;

const pdfPath = await fazt.services.pdf.fromHtml(html, {
  pageSize: 'A4',
  margin: '1cm'
});

// Returns: "_pdf/abc123.pdf"
```

### From VFS File

```javascript
// Render an HTML template stored in VFS
const pdfPath = await fazt.services.pdf.fromFile('templates/invoice.html', {
  pageSize: 'letter',
  orientation: 'portrait'
});
```

### From Internal URL

```javascript
// Render what an internal endpoint returns
// Useful for dynamic content
const pdfPath = await fazt.services.pdf.fromUrl('/api/invoice/12345', {
  pageSize: 'A4'
});

// The URL is fetched internally (no external HTTP)
// Response HTML is rendered to PDF
```

### With Data Binding

```javascript
// Load template, inject data, render
const template = await fazt.fs.read('templates/receipt.html');

const html = template
  .replace('{{name}}', order.customerName)
  .replace('{{total}}', order.total.toFixed(2))
  .replace('{{date}}', new Date().toLocaleDateString());

const pdfPath = await fazt.services.pdf.fromHtml(html);
```

### Return Bytes Instead of Path

```javascript
// Get PDF bytes directly (for streaming response)
const pdfBytes = await fazt.services.pdf.fromHtml(html, {
  output: 'bytes'
});

return {
  body: pdfBytes,
  headers: {
    'Content-Type': 'application/pdf',
    'Content-Disposition': 'attachment; filename="invoice.pdf"'
  }
};
```

### Merge Multiple PDFs

```javascript
const merged = await fazt.services.pdf.merge([
  '_pdf/cover.pdf',
  '_pdf/report.pdf',
  '_pdf/appendix.pdf'
]);

// Returns: "_pdf/merged_abc123.pdf"
```

## Page Options

### Page Sizes

| Size     | Dimensions                         |
| -------- | ---------------------------------- |
| `A4`     | 210mm x 297mm (default)            |
| `letter` | 8.5in x 11in                       |
| `legal`  | 8.5in x 14in                       |
| `A3`     | 297mm x 420mm                      |
| `A5`     | 148mm x 210mm                      |
| Custom   | `{ width: '8in', height: '10in' }` |

### Margins

```javascript
// Uniform margin
{ margin: '1cm' }

// Per-side margins
{ margin: { top: '2cm', right: '1cm', bottom: '2cm', left: '1cm' } }

// CSS shorthand
{ margin: '2cm 1cm' }  // top/bottom, left/right
```

### Orientation

```javascript
{ orientation: 'portrait' }   // default
{ orientation: 'landscape' }
```

## Referencing Assets

Images and fonts must be in the app's VFS:

```html
<!-- Images: use absolute paths from VFS root -->
<img src="/images/logo.png">
<img src="/assets/signature.jpg">

<!-- Fonts: embed via @font-face -->
<style>
  @font-face {
    font-family: 'CustomFont';
    src: url('/fonts/custom.woff2') format('woff2');
  }
  body { font-family: 'CustomFont', sans-serif; }
</style>
```

The service resolves `/path` to the app's VFS automatically.

## CSS Print Features

```css
/* Page size and margins */
@page {
  size: A4;
  margin: 2cm;
}

/* First page different margins */
@page :first {
  margin-top: 3cm;
}

/* Force page breaks */
.chapter { page-break-before: always; }
.keep-together { page-break-inside: avoid; }

/* Headers and footers */
@page {
  @top-center { content: "Company Name"; }
  @bottom-right { content: "Page " counter(page) " of " counter(pages); }
}
```

## HTTP Endpoint

```
POST /_services/pdf/render
Content-Type: application/json

{
  "html": "<html>...</html>",
  "pageSize": "A4",
  "margin": "1cm"
}

Response: application/pdf (binary)
```

**Query params for GET (from VFS file):**

```
GET /_services/pdf/render?file=templates/invoice.html&pageSize=A4
```

## Caching

PDFs are stored in VFS under `_pdf/`:

```
VFS:
├── templates/
│   └── invoice.html
└── _pdf/                    # Generated PDFs
    ├── abc123.pdf
    └── def456.pdf
```

- Generated PDFs persist until explicitly deleted
- No automatic cache invalidation (PDFs are often final documents)
- Use `fazt.services.pdf.delete(path)` to clean up

## JS API

```javascript
fazt.services.pdf.fromHtml(html, options?)
// options: { pageSize, margin, orientation, output }
// output: 'path' (default) | 'bytes'
// Returns: string (path) or Uint8Array (bytes)

fazt.services.pdf.fromFile(path, options?)
// Same options as fromHtml
// Returns: string (path) or Uint8Array (bytes)

fazt.services.pdf.fromUrl(url, options?)
// url: internal app URL (e.g., '/api/invoice/123')
// Returns: string (path) or Uint8Array (bytes)

fazt.services.pdf.merge(paths, options?)
// paths: array of PDF paths in VFS
// options: { output }
// Returns: string (path) or Uint8Array (bytes)

fazt.services.pdf.info(path)
// Returns: { pages, width, height, sizeBytes }

fazt.services.pdf.delete(path)
// Remove generated PDF from VFS
```

## Limits

| Limit            | Default       |
| ---------------- | ------------- |
| `maxHtmlSizeKB`  | 500           |
| `maxPdfSizeMB`   | 50            |
| `maxPages`       | 200           |
| `timeoutSeconds` | 30            |
| `maxStorageMB`   | 500 (per app) |

## CLI

```bash
# Render HTML file to PDF
fazt services pdf render templates/invoice.html --output invoice.pdf

# Render with options
fazt services pdf render report.html --page-size letter --orientation landscape

# Get PDF info
fazt services pdf info invoice.pdf

# List generated PDFs
fazt services pdf list --app myapp

# Clean up old PDFs
fazt services pdf purge --app myapp --older-than 30d
```

## Implementation Notes

- Uses WASM primitive (v0.10) with embedded `libpdf` module
- HTML parsing via Go HTML parser
- CSS parsing via pure Go CSS parser
- Layout engine in WASM (Rust-based typesetting)
- PDF generation via WASM
- No external network calls during rendering

## Example: Invoice Generation

```javascript
// api/invoice/[id].js
module.exports = async (req) => {
  const orderId = req.params.id;
  const order = await fazt.storage.ds.findOne('orders', { id: orderId });

  if (!order) {
    return { status: 404 };
  }

  const html = `
    <!DOCTYPE html>
    <html>
    <head>
      <style>
        body { font-family: system-ui; margin: 0; padding: 2cm; }
        .header { display: flex; justify-content: space-between; }
        .logo { height: 50px; }
        table { width: 100%; border-collapse: collapse; margin: 2em 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background: #f5f5f5; }
        .total { font-size: 1.5em; text-align: right; margin-top: 1em; }
      </style>
    </head>
    <body>
      <div class="header">
        <img src="/images/logo.png" class="logo">
        <div>
          <h1>Invoice #${order.id}</h1>
          <p>${new Date(order.createdAt).toLocaleDateString()}</p>
        </div>
      </div>

      <h2>Bill To</h2>
      <p>${order.customer.name}<br>${order.customer.address}</p>

      <table>
        <thead>
          <tr><th>Item</th><th>Qty</th><th>Price</th><th>Total</th></tr>
        </thead>
        <tbody>
          ${order.items.map(item => `
            <tr>
              <td>${item.name}</td>
              <td>${item.qty}</td>
              <td>$${item.price.toFixed(2)}</td>
              <td>$${(item.qty * item.price).toFixed(2)}</td>
            </tr>
          `).join('')}
        </tbody>
      </table>

      <div class="total">
        <strong>Total: $${order.total.toFixed(2)}</strong>
      </div>
    </body>
    </html>
  `;

  const pdfBytes = await fazt.services.pdf.fromHtml(html, {
    pageSize: 'A4',
    output: 'bytes'
  });

  return {
    body: pdfBytes,
    headers: {
      'Content-Type': 'application/pdf',
      'Content-Disposition': `attachment; filename="invoice-${order.id}.pdf"`
    }
  };
};
```

## Example: Multi-Page Report

```javascript
// Generate report with cover page and content
const coverHtml = `
  <html>
    <body style="display: flex; align-items: center; justify-content: center; height: 100vh;">
      <div style="text-align: center;">
        <h1>Annual Report 2024</h1>
        <p>Confidential</p>
      </div>
    </body>
  </html>
`;

const contentHtml = `
  <html>
    <head>
      <style>
        @page { @bottom-right { content: "Page " counter(page); } }
        .chapter { page-break-before: always; }
      </style>
    </head>
    <body>
      <div class="chapter">
        <h2>Chapter 1: Overview</h2>
        <p>...</p>
      </div>
      <div class="chapter">
        <h2>Chapter 2: Financial Summary</h2>
        <p>...</p>
      </div>
    </body>
  </html>
`;

// Render separately, then merge
const cover = await fazt.services.pdf.fromHtml(coverHtml);
const content = await fazt.services.pdf.fromHtml(contentHtml);
const report = await fazt.services.pdf.merge([cover, content]);
```
