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
```

**Parameters:**

| Param | Description |
|-------|-------------|
| `w` | Width |
| `h` | Height |
| `fit` | Fit mode: cover, contain, fill |
| `thumb` | Square thumbnail size |
| `q` | Quality (1-100) |
| `fmt` | Output format |

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
```

## Go Libraries

Pure Go, no CGO:

```go
import (
    "image"
    "image/jpeg"
    "image/png"
    "golang.org/x/image/draw"
    "golang.org/x/image/webp"
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
