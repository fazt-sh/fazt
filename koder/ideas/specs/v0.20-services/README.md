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
│                 Services                     │  ← NEW
│   forms | media | pdf | markdown | search    │
│   qr | comments | shorturl | captcha | hooks │
│   sanitize | money | humanize | timezone     │
├─────────────────────────────────────────────┤
│                  Kernel                      │
│   proc | fs | net | storage | security       │
└─────────────────────────────────────────────┘
```

## Future Repo Structure

Designed for eventual separation:

```
github.com/fazt-sh/
├── kernel/      # Core primitives
├── services/    # Common patterns (this)
└── fazt/        # Binary, CLI, admin
```

Current structure (separation-ready):

```
fazt/
├── pkg/
│   ├── kernel/
│   └── services/
│       ├── forms/
│       ├── media/
│       ├── pdf/
│       ├── markdown/
│       ├── search/
│       ├── qr/
│       ├── comments/
│       ├── shorturl/
│       ├── captcha/
│       ├── hooks/
│       ├── sanitize/
│       ├── money/
│       ├── humanize/
│       └── timezone/
└── ...
```

## Services

| Service | Purpose | Key Capability |
|---------|---------|----------------|
| `forms` | Collect form submissions | POST endpoint, zero config |
| `media` | Image processing | Resize, blurhash, QR, barcode, mimetype |
| `pdf` | Generate PDFs | HTML/CSS to PDF via WASM |
| `markdown` | Compile markdown | Goldmark, shortcodes, CSS |
| `search` | Full-text search | Bleve indexing |
| `qr` | Generate QR codes | PNG from text/URL |
| `comments` | User feedback on entities | Threading, moderation |
| `shorturl` | Shareable short links | Click tracking, expiration |
| `captcha` | Spam protection | Math/text challenges |
| `hooks` | Bidirectional webhooks | Inbound verification, outbound delivery |
| `sanitize` | HTML/text sanitization | XSS protection, policy-based |
| `money` | Decimal arithmetic | Integer cents, no float errors |
| `humanize` | Human-readable formatting | Bytes, time, numbers, ordinals |
| `timezone` | IANA timezone handling | Embedded tzdata, DST-aware |

## Common Properties

All services share:

1. **App Scoping**: Data isolated by `app_uuid`
2. **Origin Check**: HTTP endpoints only accept same-origin requests
3. **Rate Limiting**: Per-IP limits via kernel limits system
4. **JS API**: Accessible via `fazt.services.*` namespace
5. **HTTP Endpoints**: Available at `/_services/{service}/...`

## Documents

- `forms.md` - Form submission collection
- `media.md` - Image processing (resize, blurhash, QR, barcode, mimetype)
- `pdf.md` - PDF generation from HTML/CSS
- `markdown.md` - Markdown compilation
- `search.md` - Full-text search
- `qr.md` - QR code generation
- `comments.md` - User comments/feedback
- `shorturl.md` - Short URL generation
- `captcha.md` - Spam protection challenges
- `hooks.md` - Bidirectional webhook handling
- `sanitize.md` - HTML/text sanitization
- `money.md` - Decimal currency arithmetic
- `humanize.md` - Human-readable formatting
- `timezone.md` - IANA timezone handling

## Dependencies

- v0.10 (Runtime): JS API bridge
- v0.9 (Storage): Data persistence
- v0.8 (Kernel): Limits, scoping

## Non-Goals

- **Not an app framework**: Services are primitives, not Rails
- **Not extensible**: First-party services only, no plugins
- **Not JS-based**: Go code, compiled into binary
