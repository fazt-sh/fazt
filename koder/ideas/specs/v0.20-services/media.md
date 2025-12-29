# Media Service

## Summary

Image processing primitives. Resize, crop, optimize, and generate thumbnails.
All operations use pure Go libraries (no CGO, no external dependencies).

## Capabilities

| Operation | Description |
|-----------|-------------|
| `resize` | Scale image to dimensions |
| `crop` | Extract region from image |
| `thumbnail` | Generate square thumbnail |
| `optimize` | Compress without resize |
| `convert` | Change format (jpg, png, webp) |
| `blurhash` | Generate blur placeholder hash |
| `qr` | Generate QR codes |
| `barcode` | Generate 1D barcodes |
| `mimetype` | Detect file MIME type |

## Usage

### Resize

```javascript
// Resize to width, maintain aspect ratio
const resized = await fazt.services.media.resize(path, {
  width: 800
});

// Resize to exact dimensions
const exact = await fazt.services.media.resize(path, {
  width: 800,
  height: 600,
  fit: 'cover'      // 'cover' | 'contain' | 'fill'
});

// Returns path to processed image in storage
// e.g., "_media/abc123_800x600.jpg"
```

### Fit Modes

| Mode | Behavior |
|------|----------|
| `contain` | Fit within bounds, preserve aspect (default) |
| `cover` | Fill bounds, crop excess |
| `fill` | Stretch to exact dimensions |

### Thumbnail

```javascript
// Square thumbnail (center crop)
const thumb = await fazt.services.media.thumbnail(path, 200);
// Returns path: "_media/abc123_thumb_200.jpg"
```

### Crop

```javascript
const cropped = await fazt.services.media.crop(path, {
  x: 100,
  y: 100,
  width: 400,
  height: 300
});
```

### Optimize

```javascript
// Compress without resizing
const optimized = await fazt.services.media.optimize(path, {
  quality: 80       // 1-100, default 85
});
```

### Convert

```javascript
// Change format
const webp = await fazt.services.media.convert(path, 'webp');
const jpg = await fazt.services.media.convert(path, 'jpg');
```

### Blurhash

Generate compact blur placeholders for progressive image loading:

```javascript
// Generate blurhash from image
const hash = await fazt.services.media.blurhash(path);
// "LEHV6nWB2yk8pyo0adR*.7kCMdnj"

// With custom component count (higher = more detail, longer hash)
const detailed = await fazt.services.media.blurhash(path, {
  xComponents: 6,
  yComponents: 4
});

// Decode to data URL for immediate display
const placeholder = fazt.services.media.blurhashDataUrl(hash, {
  width: 32,
  height: 32
});
// "data:image/png;base64,..."

// Use case: Store hash with image, render placeholder while loading
await fazt.storage.ds.insert('images', {
  path: '/photos/sunset.jpg',
  blurhash: await fazt.services.media.blurhash('/photos/sunset.jpg'),
  width: 1920,
  height: 1080
});
```

### QR Code

Generate QR codes from text, URLs, or data:

```javascript
// Generate QR code image
const qrPath = await fazt.services.media.qr('https://example.com', {
  size: 256          // Pixels (default: 256)
});
// Returns path: "_media/qr_abc123.png"

// With options
const qr = await fazt.services.media.qr('Hello World', {
  size: 400,
  level: 'H',        // Error correction: L, M, Q, H (default: M)
  format: 'png'      // png, svg (default: png)
});

// As data URL (for inline embedding)
const dataUrl = await fazt.services.media.qrDataUrl('https://example.com');
// "data:image/png;base64,..."

// As SVG string
const svg = await fazt.services.media.qrSvg('https://example.com');
// "<svg xmlns=..."
```

### Barcode

Generate 1D barcodes for products, inventory, labels:

```javascript
// Generate barcode image
const barcodePath = await fazt.services.media.barcode('123456789012', {
  format: 'ean13'    // Barcode format
});
// Returns path: "_media/barcode_abc123.png"

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
const barcode = await fazt.services.media.barcode('5901234123457', {
  format: 'ean13',
  width: 200,
  height: 80,
  includeText: true  // Show number below barcode (default: true)
});

// As data URL
const dataUrl = await fazt.services.media.barcodeDataUrl('123456789012', {
  format: 'ean13'
});
```

### MIME Type Detection

Detect file type from content (not extension):

```javascript
// Detect from file path
const mime = await fazt.services.media.mimetype('/uploads/document');
// "application/pdf"

// Detect from buffer/bytes
const mime = fazt.services.media.mimetypeFromBytes(buffer);
// "image/png"

// Get file extension for MIME type
const ext = fazt.services.media.extFromMime('image/jpeg');
// "jpg"

// Get MIME type for extension
const mime = fazt.services.media.mimeFromExt('pdf');
// "application/pdf"

// Check if file is image
const isImage = await fazt.services.media.isImage('/uploads/file');
// true/false

// Check if file is specific type
const isPdf = await fazt.services.media.is('/uploads/file', 'application/pdf');
// true/false
```

## Supported Formats

| Format | Read | Write |
|--------|------|-------|
| JPEG | Yes | Yes |
| PNG | Yes | Yes |
| GIF | Yes | Yes (first frame) |
| WebP | Yes | Yes |
| BMP | Yes | No |

## HTTP Endpoint

On-the-fly processing via URL parameters:

```
GET /_services/media/{path}?w=800&h=600&fit=cover
GET /_services/media/{path}?thumb=200
GET /_services/media/{path}?q=80
GET /_services/media/{path}?fmt=webp
GET /_services/media/{path}?blurhash=1
GET /_services/qr?data=https://example.com&size=256
GET /_services/barcode?data=123456789012&format=ean13
```

**Image Processing Parameters:**

| Param | Description |
|-------|-------------|
| `w` | Width |
| `h` | Height |
| `fit` | Fit mode: cover, contain, fill |
| `thumb` | Square thumbnail size |
| `q` | Quality (1-100) |
| `fmt` | Output format |
| `blurhash` | Return blurhash string instead of image |

**QR Code Parameters:**

| Param | Description |
|-------|-------------|
| `data` | Content to encode (required) |
| `size` | Image size in pixels (default: 256) |
| `level` | Error correction: L, M, Q, H (default: M) |
| `fmt` | Output format: png, svg (default: png) |

**Barcode Parameters:**

| Param | Description |
|-------|-------------|
| `data` | Content to encode (required) |
| `format` | Barcode format: ean13, code128, etc. (required) |
| `width` | Image width (default: 200) |
| `height` | Image height (default: 80) |
| `text` | Show text below barcode: 1/0 (default: 1) |

**Example:**

```html
<!-- Original -->
<img src="/images/photo.jpg">

<!-- Resized on-the-fly -->
<img src="/_services/media/images/photo.jpg?w=400">

<!-- Thumbnail -->
<img src="/_services/media/images/photo.jpg?thumb=100">

<!-- WebP with quality -->
<img src="/_services/media/images/photo.jpg?fmt=webp&q=75">

<!-- QR Code -->
<img src="/_services/qr?data=https://example.com">

<!-- Barcode -->
<img src="/_services/barcode?data=5901234123457&format=ean13">
```

## Caching

Processed images are cached:

```
VFS:
├── images/
│   └── photo.jpg              # Original
└── _media/                    # Cache (auto-managed)
    ├── photo_w800.jpg
    ├── photo_thumb_100.jpg
    └── photo_w400_webp_q75.webp
```

- Cache key derived from: source path + params
- Cache invalidated when source file changes
- Auto-cleanup of stale cache entries

## JS API

```javascript
// Image Processing
fazt.services.media.resize(path, options)
// options: { width, height, fit, quality }
// Returns: string (path to processed image)

fazt.services.media.thumbnail(path, size)
// Returns: string (path to thumbnail)

fazt.services.media.crop(path, options)
// options: { x, y, width, height }
// Returns: string (path to cropped image)

fazt.services.media.optimize(path, options)
// options: { quality }
// Returns: string (path to optimized image)

fazt.services.media.convert(path, format)
// format: 'jpg' | 'png' | 'webp'
// Returns: string (path to converted image)

fazt.services.media.info(path)
// Returns: { width, height, format, size }

// Blurhash
fazt.services.media.blurhash(path, options?)
// options: { xComponents: 4, yComponents: 3 }
// Returns: string (blurhash)

fazt.services.media.blurhashDataUrl(hash, options?)
// options: { width, height }
// Returns: string (data URL)

// QR Code
fazt.services.media.qr(data, options?)
// options: { size, level, format }
// Returns: string (path to QR image)

fazt.services.media.qrDataUrl(data, options?)
fazt.services.media.qrSvg(data, options?)

// Barcode
fazt.services.media.barcode(data, options)
// options: { format, width, height, includeText }
// Returns: string (path to barcode image)

fazt.services.media.barcodeDataUrl(data, options)

// MIME Type
fazt.services.media.mimetype(path)
// Returns: string (MIME type)

fazt.services.media.mimetypeFromBytes(buffer)
fazt.services.media.extFromMime(mime)
fazt.services.media.mimeFromExt(ext)
fazt.services.media.isImage(path)
fazt.services.media.is(path, mime)
```

## Go Libraries

Pure Go, no CGO:

```go
import (
    // Image processing
    "image"
    "image/jpeg"
    "image/png"
    "golang.org/x/image/draw"
    "golang.org/x/image/webp"

    // Blurhash
    "github.com/buckket/go-blurhash"  // ~5KB

    // QR Code
    "github.com/skip2/go-qrcode"       // ~15KB

    // Barcode
    "github.com/boombuler/barcode"     // ~20KB

    // MIME Type detection
    "github.com/gabriel-vasile/mimetype"  // ~15KB
)
```

## Limits

| Limit | Default |
|-------|---------|
| `maxSourceSizeMB` | 20 |
| `maxDimension` | 8000 px |
| `maxCacheSizeMB` | 500 (per app) |
| `cacheRetentionDays` | 30 |

## CLI

```bash
# Process image
fazt services media resize images/photo.jpg --width 800

# View cache stats
fazt services media cache --app myapp

# Clear cache
fazt services media cache clear --app myapp
```

## Example: Responsive Images

```html
<picture>
  <source
    srcset="/_services/media/hero.jpg?w=1200&fmt=webp 1200w,
            /_services/media/hero.jpg?w=800&fmt=webp 800w,
            /_services/media/hero.jpg?w=400&fmt=webp 400w"
    type="image/webp">
  <source
    srcset="/_services/media/hero.jpg?w=1200 1200w,
            /_services/media/hero.jpg?w=800 800w,
            /_services/media/hero.jpg?w=400 400w"
    type="image/jpeg">
  <img src="/_services/media/hero.jpg?w=800" alt="Hero image">
</picture>
```

## Example: Avatar Thumbnails

```javascript
// api/avatar.js
module.exports = async (req) => {
  const userId = req.params.id;
  const user = await fazt.storage.ds.findOne('users', { id: userId });

  if (!user.avatar) {
    return { redirect: '/default-avatar.png' };
  }

  // Generate thumbnail if needed, return path
  const thumb = await fazt.services.media.thumbnail(user.avatar, 100);

  return { redirect: thumb };
};
```
