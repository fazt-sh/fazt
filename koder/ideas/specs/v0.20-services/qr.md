# QR Service

## Summary

Generate QR codes and barcodes from text or URLs. Returns PNG/SVG images.
Consolidated from separate QR and barcode functions.

**Note:** Barcode generation moved here from `fazt.services.media` for cleaner
organization - both QR and barcodes are code generation, not image processing.

## Usage

### Generate QR Code

```javascript
const png = await fazt.services.qr.generate('https://example.com', {
  size: 256          // Pixels (default: 256)
});

// Returns path to generated image: "_qr/abc123.png"
```

### Serve Directly

```javascript
// api/qr.js
module.exports = async (req) => {
  const data = req.query.data;
  const size = parseInt(req.query.size) || 256;

  const path = await fazt.services.qr.generate(data, { size });

  return { redirect: path };
};
```

## In HTML

```html
<img src="/_services/qr?data=https://example.com" alt="QR Code">

<img src="/_services/qr?data=https://example.com/contact&size=200" alt="Contact QR">
```

## In Markdown (Shortcode)

```markdown
Scan to visit:

{{qr data="https://example.com" size="200"}}
```

### Generate QR as SVG

```javascript
const svg = await fazt.services.qr.svg('https://example.com', {
  size: 256
});
// Returns: string (SVG markup)
```

### Generate Barcode

```javascript
// Generate barcode image
const barcodePath = await fazt.services.qr.barcode('123456789012', {
  format: 'ean13'    // Barcode format (required)
});
// Returns path: "_qr/barcode_abc123.png"

// Supported formats
const barcodes = {
  'codabar': '123456',
  'code39': 'HELLO',
  'code93': 'HELLO123',
  'code128': 'Hello World!',
  'ean8': '12345670',
  'ean13': '5901234123457',
  'itf': '123456789012',
  'upca': '012345678905',
  'upce': '01234565'
};

// With options
const barcode = await fazt.services.qr.barcode('5901234123457', {
  format: 'ean13',
  width: 200,
  height: 80,
  includeText: true  // Show number below barcode (default: true)
});

// As data URL
const dataUrl = await fazt.services.qr.barcodeDataUrl('123456789012', {
  format: 'ean13'
});
```

## HTTP Endpoints

### QR Code
```
GET /_services/qr?data=https://example.com
GET /_services/qr?data=https://example.com&size=512
GET /_services/qr?data=https://example.com&fmt=svg
```

### Barcode
```
GET /_services/barcode?data=5901234123457&format=ean13
GET /_services/barcode?data=HELLO&format=code39&width=300
```

**QR Parameters:**

| Param   | Description                  | Default |
| ------- | ---------------------------- | ------- |
| `data`  | Content to encode (required) | ------- |
| `size`  | Image size in pixels         | 256     |
| `level` | Error correction: L, M, Q, H | M       |
| `fmt`   | Output format: png, svg      | png     |

**Barcode Parameters:**

| Param    | Description                  | Default |
| -------- | ---------------------------- | ------- |
| `data`   | Content to encode (required) | ------- |
| `format` | Barcode format (required)    | ------- |
| `width`  | Image width                  | 200     |
| `height` | Image height                 | 80      |
| `text`   | Show text below: 1/0         | 1       |

## JS API

```javascript
// QR Code
fazt.services.qr.generate(data, options?)
// options: { size, level }
// Returns: string (path to PNG)

fazt.services.qr.dataUrl(data, options?)
// Returns: string (base64 data URL)

fazt.services.qr.svg(data, options?)
// Returns: string (SVG markup)

// Barcode
fazt.services.qr.barcode(data, options)
// options: { format, width, height, includeText }
// Returns: string (path to PNG)

fazt.services.qr.barcodeDataUrl(data, options)
// Returns: string (base64 data URL)
```

## Go Libraries

```go
import (
    "github.com/skip2/go-qrcode"   // ~15KB, QR generation
    "github.com/boombuler/barcode" // ~20KB, barcode generation
)
```

Pure Go, no CGO.

## Caching

Generated QR codes are cached:

```
VFS:
└── _qr/
    ├── abc123.png      # hash of data+size
    └── def456.png
```

- Cache key: hash of (data + size)
- Same input = same cached output
- Auto-cleanup after 30 days unused

## Limits

| Limit           | Default    |
| --------------- | ---------- |
| `maxDataLength` | 2048 bytes |
| `maxSize`       | 1024 px    |
| `minSize`       | 64 px      |

## Use Cases

### VPN Peer Setup (v0.13)

```javascript
// api/vpn-qr.js
module.exports = async (req) => {
  const config = await fazt.net.vpn.peerConfig(req.params.peerId);
  const qr = await fazt.services.qr.generate(config);
  return { redirect: qr };
};
```

### Share Link

```html
<a href="/share/abc123">
  <img src="/_services/qr?data=https://myapp.example.com/share/abc123">
  Scan to open on phone
</a>
```

### Contact Card (vCard)

```javascript
const vcard = `BEGIN:VCARD
VERSION:3.0
N:Doe;John
EMAIL:john@example.com
END:VCARD`;

const qr = await fazt.services.qr.generate(vcard);
```

### WiFi Credentials

```javascript
const wifi = 'WIFI:T:WPA;S:MyNetwork;P:MyPassword;;';
const qr = await fazt.services.qr.generate(wifi);
```
