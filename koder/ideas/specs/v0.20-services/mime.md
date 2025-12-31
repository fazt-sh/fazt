# Mime Library

## Summary

Pure mimetype detection functions. Extracted from the media service because
mimetype detection is a stateless lookup, not image processing.

## Usage

```javascript
// Detect from file path (reads magic bytes)
fazt.lib.mime.detect('/uploads/photo.jpg')
// 'image/jpeg'

// Detect from raw bytes
const buffer = await fazt.fs.read('/uploads/mystery-file')
fazt.lib.mime.fromBytes(buffer)
// 'application/pdf'

// Convert between mime and extension
fazt.lib.mime.toExt('image/png')     // 'png'
fazt.lib.mime.fromExt('mp4')         // 'video/mp4'

// Type checks
fazt.lib.mime.isImage('/uploads/photo.jpg')  // true
fazt.lib.mime.is('/uploads/doc.pdf', 'application/pdf')  // true
```

## JS API

```javascript
// Detection
fazt.lib.mime.detect(path)           // Detect from file path
fazt.lib.mime.fromBytes(buffer)      // Detect from raw bytes

// Conversion
fazt.lib.mime.toExt(mime)            // 'image/png' → 'png'
fazt.lib.mime.fromExt(ext)           // 'png' → 'image/png'

// Checks
fazt.lib.mime.isImage(path)          // Is this an image?
fazt.lib.mime.is(path, mime)         // Does path match mime type?
```

## HTTP Endpoint

None. Pure functions aren't exposed via HTTP.

## Why Extract from Media?

The media service is about image transformation (resize, crop, optimize).
Mimetype detection is a pure lookup that:
- Has no state
- Does no I/O beyond reading bytes
- Returns the same output for the same input
- Is useful beyond image processing

Moving it to `fazt.lib.mime` aligns with the services/lib distinction:
- Services: stateful, have lifecycle, do I/O transforms
- Lib: pure functions, no state

## Go Implementation

Uses `net/http.DetectContentType` plus extended magic byte detection:

```go
func Detect(path string) (string, error) {
    f, err := os.Open(path)
    if err != nil {
        return "", err
    }
    defer f.Close()

    // Read first 512 bytes for detection
    buffer := make([]byte, 512)
    n, _ := f.Read(buffer)

    return http.DetectContentType(buffer[:n]), nil
}
```

Extended detection covers formats not in stdlib:
- WebP, AVIF, HEIC (images)
- WASM, SQLite, ZIP variants (applications)

## Supported Types

Common types with reliable detection:

| Category | Types |
|----------|-------|
| Images | jpeg, png, gif, webp, svg, avif, heic, ico, bmp |
| Documents | pdf, docx, xlsx, pptx |
| Audio | mp3, ogg, wav, flac, aac |
| Video | mp4, webm, avi, mov |
| Archives | zip, gzip, tar, rar, 7z |
| Code | wasm, sqlite |

## Implementation Notes

- ~5KB binary addition
- Pure Go (no CGO)
- Magic byte database embedded
- Falls back to extension-based detection
