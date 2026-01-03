# v0.20 - Services

**Theme**: Common patterns as platform primitives.

## Summary

v0.20 introduces a services layer - Go libraries that provide common patterns
to apps without requiring each app to reinvent them. Services sit between
the kernel (low-level primitives) and apps (user code).

## Philosophy

- **Go, not JS**: Services are compiled Go code, not runtime scripts
- **Primitives, not opinions**: Minimal API, maximum flexibility
- **App-scoped**: All service calls isolated to calling app
- **Zero config**: Works out of the box with sensible defaults

## Architecture

```
┌─────────────────────────────────────────────┐
│                   Apps                       │
│         (user code, JS runtime)              │
├─────────────────────────────────────────────┤
│                 Services                     │  ← Stateful
│   forms | image | pdf | markdown | search    │
│   qr | comments | shorturl | captcha | hooks │
│   rag                                         │
├─────────────────────────────────────────────┤
│                   Lib                        │  ← Pure functions
│   money | humanize | timezone | sanitize     │
│   password | geo | mime                      │
├─────────────────────────────────────────────┤
│                  Kernel                      │
│   proc | fs | net | storage | security       │
└─────────────────────────────────────────────┘
```

### Services vs Lib

**Services** (`fazt.services.*`) - Stateful, have lifecycle, do I/O:
- Store data (forms, comments, shorturl, search, hooks, captcha)
- Transform with caching (image, pdf, qr, markdown)

**Lib** (`fazt.lib.*`) - Pure functions, no state, no side effects:
- Compute values (money, humanize, timezone, sanitize, password, geo, mime)

## Future Repo Structure

Designed for eventual separation:

```
github.com/fazt-sh/
├── kernel/      # Core primitives
├── services/    # Stateful services
├── lib/         # Pure utility functions
└── fazt/        # Binary, CLI, admin
```

Current structure (separation-ready):

```
fazt/
├── pkg/
│   ├── kernel/
│   ├── services/
│   │   ├── forms/
│   │   ├── image/       # Renamed from media
│   │   ├── pdf/
│   │   ├── markdown/
│   │   ├── search/
│   │   ├── qr/          # Consolidated (includes barcode)
│   │   ├── comments/
│   │   ├── shorturl/
│   │   ├── captcha/
│   │   └── hooks/
│   └── lib/
│       ├── money/
│       ├── humanize/
│       ├── timezone/
│       ├── sanitize/
│       ├── password/
│       ├── geo/
│       └── mime/        # Extracted from media
└── ...
```

## Services

| Service | Purpose | Key Capability |
|---------|---------|----------------|
| `forms` | Collect form submissions | POST endpoint, zero config |
| `image` | Image processing | Resize, thumbnail, blurhash |
| `pdf` | Generate PDFs | HTML/CSS to PDF via WASM |
| `markdown` | Compile markdown | Goldmark, shortcodes, CSS |
| `search` | Full-text search | Bleve indexing |
| `qr` | QR & barcode generation | PNG/SVG from text/URL |
| `comments` | User feedback on entities | Threading, moderation |
| `shorturl` | Shareable short links | Click tracking, expiration |
| `captcha` | Spam protection | Math/text challenges |
| `hooks` | Bidirectional webhooks | Inbound verification, outbound delivery |
| `rag` | RAG pipelines | Ingest, retrieve, ask with grounded answers |

## Lib

| Library | Purpose | Key Capability |
|---------|---------|----------------|
| `money` | Decimal arithmetic | Integer cents, no float errors |
| `humanize` | Human-readable formatting | Bytes, time, numbers, ordinals |
| `timezone` | IANA timezone handling | Embedded tzdata, DST-aware |
| `sanitize` | HTML/text sanitization | XSS protection, policy-based |
| `password` | Secure credential hashing | Argon2id, timing-safe verify |
| `geo` | Geographic primitives | Distance, IP geolocation, geofencing |
| `mime` | Mimetype detection | File/buffer analysis |

## Common Properties

**Services** share:

1. **App Scoping**: Data isolated by `app_uuid`
2. **Origin Check**: HTTP endpoints only accept same-origin requests
3. **Rate Limiting**: Per-IP limits via kernel limits system
4. **JS API**: Accessible via `fazt.services.*` namespace
5. **HTTP Endpoints**: Available at `/_services/{service}/...`

**Lib** functions:

1. **Pure**: No side effects, same input → same output
2. **Stateless**: No data stored, no lifecycle
3. **JS API**: Accessible via `fazt.lib.*` namespace
4. **No HTTP**: Not exposed as endpoints (nothing to expose)

## Documents

### Services
- `forms.md` - Form submission collection
- `image.md` - Image processing (resize, blurhash)
- `pdf.md` - PDF generation from HTML/CSS
- `markdown.md` - Markdown compilation
- `search.md` - Full-text search
- `qr.md` - QR and barcode generation
- `comments.md` - User comments/feedback
- `shorturl.md` - Short URL generation
- `captcha.md` - Spam protection challenges
- `hooks.md` - Bidirectional webhook handling
- `rag.md` - RAG pipeline service

### Lib
- `money.md` - Decimal currency arithmetic
- `humanize.md` - Human-readable formatting
- `timezone.md` - IANA timezone handling
- `sanitize.md` - HTML/text sanitization
- `password.md` - Secure Argon2id password hashing
- `geo.md` - Geographic primitives and IP geolocation
- `mime.md` - Mimetype detection (extracted from media)
- `lib-split.md` - Migration spec for services/lib separation

## Dependencies

- v0.10 (Runtime): JS API bridge
- v0.9 (Storage): Data persistence
- v0.8 (Kernel): Limits, scoping

## Non-Goals

- **Not an app framework**: Services are primitives, not Rails
- **Not extensible**: First-party services only, no plugins
- **Not JS-based**: Go code, compiled into binary
