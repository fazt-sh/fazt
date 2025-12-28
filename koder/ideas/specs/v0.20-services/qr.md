# QR Service

## Summary

Generate QR codes from text or URLs. Returns PNG images. Simple, focused,
one job.

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

## HTTP Endpoint

```
GET /_services/qr?data=https://example.com
GET /_services/qr?data=https://example.com&size=512
```

Returns PNG image directly.

**Parameters:**

| Param | Description | Default |
|-------|-------------|---------|
| `data` | Content to encode (required) | - |
| `size` | Image size in pixels | 256 |

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

## JS API

```javascript
fazt.services.qr.generate(data, options?)
// options: { size }
// Returns: string (path to PNG)

fazt.services.qr.dataUrl(data, options?)
// Returns: string (base64 data URL)
// Useful for embedding directly: <img src="{{dataUrl}}">
```

## Go Library

```go
import "github.com/skip2/go-qrcode"
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

| Limit | Default |
|-------|---------|
| `maxDataLength` | 2048 bytes |
| `maxSize` | 1024 px |
| `minSize` | 64 px |

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
