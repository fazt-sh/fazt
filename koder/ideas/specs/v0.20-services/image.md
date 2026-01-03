# Image Service

## Summary

Image processing primitives. Resize, crop, optimize, and generate thumbnails.
All operations use pure Go libraries (no CGO, no external dependencies).

**Note:** Renamed from `media` for clarity. This service handles images only.
QR/barcode generation moved to `fazt.services.qr`. Mimetype detection moved
to `fazt.lib.mime`.

## Capabilities

| Operation   | Description                    |
| ----------- | ------------------------------ |
| `resize`    | Scale image to dimensions      |
| `crop`      | Extract region from image      |
| `thumbnail` | Generate square thumbnail      |
| `optimize`  | Compress without resize        |
| `convert`   | Change format (jpg, png, webp) |
| `blurhash`  | Generate blur placeholder hash |

## Usage

### Resize

```javascript
// Resize to width, maintain aspect ratio
const resized = await fazt.services.image.resize(path, {
  width: 800
});

// Resize to exact dimensions
const exact = await fazt.services.image.resize(path, {
  width: 800,
  height: 600,
  fit: 'cover'      // 'cover' | 'contain' | 'fill'
});

// Returns path to processed image in storage
// e.g., "_image/abc123_800x600.jpg"
```

### Fit Modes

| Mode      | Behavior                                     |
| --------- | -------------------------------------------- |
| `contain` | Fit within bounds, preserve aspect (default) |
| `cover`   | Fill bounds, crop excess                     |
| `fill`    | Stretch to exact dimensions                  |

### Thumbnail

```javascript
// Square thumbnail (center crop)
const thumb = await fazt.services.image.thumbnail(path, 200);
// Returns path: "_image/abc123_thumb_200.jpg"
```

### Crop

```javascript
const cropped = await fazt.services.image.crop(path, {
  x: 100,
  y: 100,
  width: 400,
  height: 300
});
```

### Optimize

```javascript
// Compress without resizing
const optimized = await fazt.services.image.optimize(path, {
  quality: 80       // 1-100, default 85
});
```

### Convert

```javascript
// Change format
const webp = await fazt.services.image.convert(path, 'webp');
const jpg = await fazt.services.image.convert(path, 'jpg');
```

### Blurhash

Generate compact blur placeholders for progressive image loading:

```javascript
// Generate blurhash from image
const hash = await fazt.services.image.blurhash(path);
// "LEHV6nWB2yk8pyo0adR*.7kCMdnj"

// With custom component count (higher = more detail, longer hash)
const detailed = await fazt.services.image.blurhash(path, {
  xComponents: 6,
  yComponents: 4
});

// Decode to data URL for immediate display
const placeholder = fazt.services.image.blurhashDataUrl(hash, {
  width: 32,
  height: 32
});
// "data:image/png;base64,..."

// Use case: Store hash with image, render placeholder while loading
await fazt.storage.ds.insert('images', {
  path: '/photos/sunset.jpg',
  blurhash: await fazt.services.image.blurhash('/photos/sunset.jpg'),
  width: 1920,
  height: 1080
});
```

## Supported Formats

| Format | Read | Write             |
| ------ | ---- | ----------------- |
| JPEG   | Yes  | Yes               |
| PNG    | Yes  | Yes               |
| GIF    | Yes  | Yes (first frame) |
| WebP   | Yes  | Yes               |
| BMP    | Yes  | No                |

## HTTP Endpoint

On-the-fly processing via URL parameters:

```
GET /_services/image/{path}?w=800&h=600&fit=cover
GET /_services/image/{path}?thumb=200
GET /_services/image/{path}?q=80
GET /_services/image/{path}?fmt=webp
GET /_services/image/{path}?blurhash=1
```

**Parameters:**

| Param      | Description                             |
| ---------- | --------------------------------------- |
| `w`        | Width                                   |
| `h`        | Height                                  |
| `fit`      | Fit mode: cover, contain, fill          |
| `thumb`    | Square thumbnail size                   |
| `q`        | Quality (1-100)                         |
| `fmt`      | Output format                           |
| `blurhash` | Return blurhash string instead of image |

**Example:**

```html
<!-- Original -->
<img src="/images/photo.jpg">

<!-- Resized on-the-fly -->
<img src="/_services/image/images/photo.jpg?w=400">

<!-- Thumbnail -->
<img src="/_services/image/images/photo.jpg?thumb=100">

<!-- WebP with quality -->
<img src="/_services/image/images/photo.jpg?fmt=webp&q=75">
```

## Caching

Processed images are cached:

```
VFS:
├── images/
│   └── photo.jpg              # Original
└── _image/                    # Cache (auto-managed)
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
fazt.services.image.resize(path, options)
// options: { width, height, fit, quality }
// Returns: string (path to processed image)

fazt.services.image.thumbnail(path, size)
// Returns: string (path to thumbnail)

fazt.services.image.crop(path, options)
// options: { x, y, width, height }
// Returns: string (path to cropped image)

fazt.services.image.optimize(path, options)
// options: { quality }
// Returns: string (path to optimized image)

fazt.services.image.convert(path, format)
// format: 'jpg' | 'png' | 'webp'
// Returns: string (path to converted image)

fazt.services.image.info(path)
// Returns: { width, height, format, size }

// Blurhash
fazt.services.image.blurhash(path, options?)
// options: { xComponents: 4, yComponents: 3 }
// Returns: string (blurhash)

fazt.services.image.blurhashDataUrl(hash, options?)
// options: { width, height }
// Returns: string (data URL)
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
)
```

## Limits

| Limit                | Default       |
| -------------------- | ------------- |
| `maxSourceSizeMB`    | 20            |
| `maxDimension`       | 8000 px       |
| `maxCacheSizeMB`     | 500 (per app) |
| `cacheRetentionDays` | 30            |

## CLI

```bash
# Process image
fazt services image resize images/photo.jpg --width 800

# View cache stats
fazt services image cache --app myapp

# Clear cache
fazt services image cache clear --app myapp
```

## Example: Responsive Images

```html
<picture>
  <source
    srcset="/_services/image/hero.jpg?w=1200&fmt=webp 1200w,
            /_services/image/hero.jpg?w=800&fmt=webp 800w,
            /_services/image/hero.jpg?w=400&fmt=webp 400w"
    type="image/webp">
  <source
    srcset="/_services/image/hero.jpg?w=1200 1200w,
            /_services/image/hero.jpg?w=800 800w,
            /_services/image/hero.jpg?w=400 400w"
    type="image/jpeg">
  <img src="/_services/image/hero.jpg?w=800" alt="Hero image">
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
  const thumb = await fazt.services.image.thumbnail(user.avatar, 100);

  return { redirect: thumb };
};
```

## Migration from media

```javascript
// Deprecated (warns)
fazt.services.media.resize(...)
fazt.services.media.thumbnail(...)

// New
fazt.services.image.resize(...)
fazt.services.image.thumbnail(...)

// Moved to fazt.services.qr
fazt.services.media.qr(...)        // → fazt.services.qr.generate(...)
fazt.services.media.barcode(...)   // → fazt.services.qr.barcode(...)

// Moved to fazt.lib.mime
fazt.services.media.mimetype(...)  // → fazt.lib.mime.detect(...)
```
