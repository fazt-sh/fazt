# Services/Lib Split

## Summary

Refactor `fazt.services.*` by moving pure utility functions to a new
`fazt.lib.*` namespace. Services remain for stateful/effectful operations.
Libraries are pure functions with no state or side effects.

## The Problem

The current `fazt.services.*` namespace conflates two distinct concepts:

**Actual services** (stateful, lifecycle, I/O):
- forms, comments, shorturl, search, hooks, captcha
- image processing, PDF generation, markdown rendering

**Pure utilities** (stateless, no I/O):
- money, humanize, timezone, sanitize, password, geo

This violates the OS metaphor. In operating systems:
- **Services** are daemons that manage state and handle requests
- **Libraries** are linked code providing functions

Calling `fazt.services.money.add(100, 200)` feels wrong because it's not a
service—it's a function call.

## The Solution

Split into two namespaces:

```
fazt.services.*    Stateful services (stores data, manages lifecycle, does I/O)
fazt.lib.*         Pure functions (no state, no side effects)
```

## What Stays in Services

| Service    | Reason                                                      |
| ---------- | ----------------------------------------------------------- |
| `forms`    | Stores form submissions                                     |
| `comments` | Stores threaded comments                                    |
| `shorturl` | Stores URLs, tracks clicks                                  |
| `search`   | Maintains full-text index                                   |
| `hooks`    | Stores events, manages delivery                             |
| `captcha`  | Tracks challenge state                                      |
| `image`    | Transforms images, uses cache (renamed from `media`)        |
| `pdf`      | Generates PDFs, WASM runtime                                |
| `qr`       | Generates QR/barcodes (consolidated)                        |
| `markdown` | Renders markdown (borderline, shortcodes have side effects) |

## What Moves to Lib

| Library    | Reason                                         |
| ---------- | ---------------------------------------------- |
| `money`    | Pure arithmetic                                |
| `humanize` | Pure formatting                                |
| `timezone` | Pure time math                                 |
| `sanitize` | Pure string operations                         |
| `password` | Pure crypto (hash/verify)                      |
| `geo`      | Pure math + embedded data lookup               |
| `mime`     | Pure mimetype detection (extracted from media) |

## Additional Cleanups

### Rename: media → image

The `media` service only handles images. Rename to `image` for clarity:

```javascript
// Before
fazt.services.media.resize(path, options)
fazt.services.media.thumbnail(path, size)

// After
fazt.services.image.resize(path, options)
fazt.services.image.thumbnail(path, size)
```

### Extract: mime from media

Mimetype detection is a pure lookup, not image processing:

```javascript
// Before
fazt.services.media.mimetype(path)
fazt.services.media.mimetypeFromBytes(buffer)
fazt.services.media.extFromMime(mime)
fazt.services.media.mimeFromExt(ext)
fazt.services.media.isImage(path)
fazt.services.media.is(path, mime)

// After
fazt.lib.mime.detect(path)
fazt.lib.mime.fromBytes(buffer)
fazt.lib.mime.toExt(mime)
fazt.lib.mime.fromExt(ext)
fazt.lib.mime.isImage(path)
fazt.lib.mime.is(path, mime)
```

### Consolidate: QR/barcode

Remove QR from media, keep only in `fazt.services.qr`:

```javascript
// Remove these duplicates
fazt.services.media.qr()
fazt.services.media.qrDataUrl()
fazt.services.media.qrSvg()
fazt.services.media.barcode()
fazt.services.media.barcodeDataUrl()

// Keep only
fazt.services.qr.generate()
fazt.services.qr.dataUrl()
fazt.services.qr.svg()
fazt.services.qr.barcode()
fazt.services.qr.barcodeDataUrl()
```

### Extract: blurhash from media

Blurhash is placeholder generation, not image transformation:

```javascript
// Before
fazt.services.media.blurhash(path, options)
fazt.services.media.blurhashDataUrl(hash, options)

// After (stays in image, it operates on images)
fazt.services.image.blurhash(path, options)
fazt.services.image.blurhashDataUrl(hash, options)
```

## New Namespace: fazt.lib.*

```javascript
fazt.lib
├── money
│   ├── add(), subtract(), multiply(), divide()
│   ├── percent(), addPercent(), subtractPercent()
│   ├── format(), parse(), compare(), min(), max()
│   ├── split(), allocate(), currency(), currencies()
├── humanize
│   ├── bytes(), time(), duration(), number(), compact()
│   ├── ordinal(), plural(), truncate(), list()
├── timezone
│   ├── now(), convert(), parse(), format()
│   ├── isDST(), transitions(), info(), list(), search()
│   ├── offset(), offsetFromUTC()
│   ├── next(), scheduleDaily(), isWithin()
├── sanitize
│   ├── html(), text(), markdown(), url()
├── password
│   ├── hash(), verify(), needsRehash(), config()
├── geo
│   ├── distance(), fromIP(), countryFromIP()
│   ├── contains(), inBounds(), bounds()
│   ├── timezone(), countryAt(), nearby()
└── mime
    ├── detect(), fromBytes()
    ├── toExt(), fromExt()
    ├── isImage(), is()
```

## Cleaned Services Namespace

```javascript
fazt.services
├── forms
│   ├── list(), get(), delete(), count(), clear()
├── comments
│   ├── add(), list(), get(), update(), delete()
│   ├── hide(), show(), approve(), count()
├── shorturl
│   ├── create(), get(), update(), delete(), list()
│   ├── stats(), clicks()
├── search
│   ├── index(), indexFiles(), query(), reindex(), dropIndex(), indexes()
├── hooks
│   ├── events(), event(), replay(), replayFailed(), stats()
│   ├── register(), list(), update(), delete(), emit()
│   ├── deliveries(), retryDelivery()
├── captcha
│   ├── create(), verify()
├── image                    # renamed from media
│   ├── resize(), thumbnail(), crop(), optimize(), convert(), info()
│   ├── blurhash(), blurhashDataUrl()
├── pdf
│   ├── fromHtml(), fromFile(), fromUrl(), merge(), info(), delete()
├── qr                       # consolidated
│   ├── generate(), dataUrl(), svg()
│   ├── barcode(), barcodeDataUrl()
└── markdown
    ├── render(), renderFile(), extract()
```

## Migration

Deprecation aliases for backward compatibility:

```javascript
// Old paths work but emit deprecation warning
fazt.services.money.add(100, 200)
// → "Deprecated: use fazt.lib.money.add()"

fazt.services.media.resize(path, options)
// → "Deprecated: use fazt.services.image.resize()"

fazt.services.media.mimetype(path)
// → "Deprecated: use fazt.lib.mime.detect()"
```

Timeline:
- v0.20: Both paths work, old paths warn
- v0.21: Old paths removed

## Why fazt.lib?

Considered alternatives:

| Option        | Pros                      | Cons                            |
| ------------- | ------------------------- | ------------------------------- |
| `fazt.lib.*`  | Short, mirrors OS concept | New namespace                   |
| `fazt.util.*` | JS convention             | Generic, overused               |
| `fazt.std.*`  | "stdlib" feel             | Unfamiliar                      |
| Root level    | No nesting                | Crowds root (already 25+ items) |

`fazt.lib.*` wins:
- Short (4 chars)
- Clear meaning (library functions)
- OS-aligned (services vs libraries)
- Doesn't crowd root namespace

## CLI Impact

No CLI changes. These are JS-only APIs.

## HTTP API Impact

No HTTP changes. Pure functions aren't exposed via HTTP (nothing to expose).

## Implementation Notes

- Pure refactoring, no new functionality
- Aliases via Go shims in JS runtime
- Deprecation warnings via structured logging
- ~0 binary size change (just reorganization)

## Affected Specs

Update these specs to reference new paths:

- `money.md` → Update examples to `fazt.lib.money.*`
- `humanize.md` → Update examples to `fazt.lib.humanize.*`
- `timezone.md` → Update examples to `fazt.lib.timezone.*`
- `sanitize.md` → Update examples to `fazt.lib.sanitize.*`
- `password.md` → Update examples to `fazt.lib.password.*`
- `geo.md` → Update examples to `fazt.lib.geo.*`
- `media.md` → Rename to `image.md`, remove QR/barcode/mimetype
- `qr.md` → Add barcode methods

## Checklist

- [ ] Create `fazt.lib` namespace in JS runtime
- [ ] Move money, humanize, timezone, sanitize, password, geo
- [ ] Extract mime from media
- [ ] Rename media → image
- [ ] Consolidate QR into qr service
- [ ] Add deprecation aliases
- [ ] Update SURFACE.md
- [ ] Update affected spec files
