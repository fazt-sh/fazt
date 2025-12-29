# Humanize Service

## Summary

Human-readable formatting for numbers, bytes, durations, and time. Transform
raw data into text humans can scan and understand instantly.

## The Problem

```javascript
// Raw data is hard for humans to parse
file.size        // 1073741824
elapsed          // 3723000
count            // 1847293
position         // 3

// Users expect:
// "1.0 GB"
// "1 hour ago"
// "1,847,293"
// "3rd"
```

## The Solution

```javascript
fazt.services.humanize.bytes(1073741824)      // "1.0 GB"
fazt.services.humanize.time(Date.now() - 3723000)  // "1 hour ago"
fazt.services.humanize.number(1847293)        // "1,847,293"
fazt.services.humanize.ordinal(3)             // "3rd"
```

## Usage

### Bytes

```javascript
// Binary (1024-based, default)
fazt.services.humanize.bytes(0)                // "0 B"
fazt.services.humanize.bytes(1023)             // "1023 B"
fazt.services.humanize.bytes(1024)             // "1.0 KB"
fazt.services.humanize.bytes(1048576)          // "1.0 MB"
fazt.services.humanize.bytes(1073741824)       // "1.0 GB"
fazt.services.humanize.bytes(1099511627776)    // "1.0 TB"

// SI (1000-based)
fazt.services.humanize.bytes(1000, { si: true })     // "1.0 kB"
fazt.services.humanize.bytes(1000000, { si: true })  // "1.0 MB"

// Precision control
fazt.services.humanize.bytes(1536, { precision: 2 })  // "1.50 KB"
fazt.services.humanize.bytes(1536, { precision: 0 })  // "2 KB"
```

### Time (Relative)

```javascript
const now = Date.now();

// Past
fazt.services.humanize.time(now - 5000)        // "5 seconds ago"
fazt.services.humanize.time(now - 120000)      // "2 minutes ago"
fazt.services.humanize.time(now - 3600000)     // "1 hour ago"
fazt.services.humanize.time(now - 86400000)    // "1 day ago"
fazt.services.humanize.time(now - 604800000)   // "1 week ago"
fazt.services.humanize.time(now - 2592000000)  // "1 month ago"
fazt.services.humanize.time(now - 31536000000) // "1 year ago"

// Future
fazt.services.humanize.time(now + 3600000)     // "in 1 hour"
fazt.services.humanize.time(now + 86400000)    // "in 1 day"

// Just now threshold
fazt.services.humanize.time(now - 500)         // "just now"
fazt.services.humanize.time(now - 3000)        // "just now" (default: 10s)

// Custom threshold
fazt.services.humanize.time(now - 3000, { justNow: 5000 })  // "3 seconds ago"

// Works with timestamps and Date objects
fazt.services.humanize.time(1703980800000)     // "2 months ago" (example)
fazt.services.humanize.time(new Date('2024-01-01'))
```

### Duration

```javascript
// Milliseconds to human duration
fazt.services.humanize.duration(0)             // "0ms"
fazt.services.humanize.duration(500)           // "500ms"
fazt.services.humanize.duration(1500)          // "1s"
fazt.services.humanize.duration(65000)         // "1m 5s"
fazt.services.humanize.duration(3723000)       // "1h 2m"
fazt.services.humanize.duration(90061000)      // "1d 1h"

// Verbose mode
fazt.services.humanize.duration(3723000, { verbose: true })
// "1 hour, 2 minutes, 3 seconds"

// Precision control
fazt.services.humanize.duration(3723456, { parts: 3 })  // "1h 2m 3s"
fazt.services.humanize.duration(3723456, { parts: 1 })  // "1h"
```

### Numbers

```javascript
// Comma formatting (locale-aware)
fazt.services.humanize.number(1234)            // "1,234"
fazt.services.humanize.number(1234567)         // "1,234,567"
fazt.services.humanize.number(1234.56)         // "1,234.56"

// Locale support
fazt.services.humanize.number(1234567, { locale: 'de-DE' })  // "1.234.567"
fazt.services.humanize.number(1234567, { locale: 'fr-FR' })  // "1 234 567"

// Compact notation
fazt.services.humanize.compact(1234)           // "1.2K"
fazt.services.humanize.compact(1234567)        // "1.2M"
fazt.services.humanize.compact(1234567890)     // "1.2B"
fazt.services.humanize.compact(1234567890000)  // "1.2T"

// Compact with precision
fazt.services.humanize.compact(1567, { precision: 2 })  // "1.57K"
```

### Ordinals

```javascript
fazt.services.humanize.ordinal(1)   // "1st"
fazt.services.humanize.ordinal(2)   // "2nd"
fazt.services.humanize.ordinal(3)   // "3rd"
fazt.services.humanize.ordinal(4)   // "4th"
fazt.services.humanize.ordinal(11)  // "11th"
fazt.services.humanize.ordinal(12)  // "12th"
fazt.services.humanize.ordinal(13)  // "13th"
fazt.services.humanize.ordinal(21)  // "21st"
fazt.services.humanize.ordinal(22)  // "22nd"
fazt.services.humanize.ordinal(23)  // "23rd"
fazt.services.humanize.ordinal(100) // "100th"
fazt.services.humanize.ordinal(101) // "101st"
```

### Pluralization

```javascript
fazt.services.humanize.plural(0, 'item')   // "0 items"
fazt.services.humanize.plural(1, 'item')   // "1 item"
fazt.services.humanize.plural(5, 'item')   // "5 items"

// Irregular plurals
fazt.services.humanize.plural(1, 'person', 'people')  // "1 person"
fazt.services.humanize.plural(5, 'person', 'people')  // "5 people"

// Without count
fazt.services.humanize.plural(5, 'item', null, { count: false })  // "items"
```

### Truncate

```javascript
// Truncate text with ellipsis
fazt.services.humanize.truncate('Hello World', 8)     // "Hello..."
fazt.services.humanize.truncate('Hello World', 20)    // "Hello World"

// Custom suffix
fazt.services.humanize.truncate('Hello World', 8, { suffix: '…' })  // "Hello…"

// Word-aware (don't break mid-word)
fazt.services.humanize.truncate('Hello beautiful World', 12, { word: true })
// "Hello..."
```

### List

```javascript
// Oxford comma formatting
fazt.services.humanize.list(['apple'])                    // "apple"
fazt.services.humanize.list(['apple', 'banana'])          // "apple and banana"
fazt.services.humanize.list(['apple', 'banana', 'cherry'])
// "apple, banana, and cherry"

// Custom conjunction
fazt.services.humanize.list(['red', 'blue'], { conjunction: 'or' })
// "red or blue"

// No oxford comma
fazt.services.humanize.list(['a', 'b', 'c'], { oxford: false })
// "a, b and c"
```

## JS API

```javascript
// Bytes
fazt.services.humanize.bytes(bytes, options?)
// options: { si: boolean, precision: number }

// Time
fazt.services.humanize.time(timestamp, options?)
// options: { justNow: number }
// Returns relative time string

// Duration
fazt.services.humanize.duration(ms, options?)
// options: { verbose: boolean, parts: number }

// Numbers
fazt.services.humanize.number(n, options?)
// options: { locale: string }

fazt.services.humanize.compact(n, options?)
// options: { precision: number }

// Text
fazt.services.humanize.ordinal(n)
fazt.services.humanize.plural(count, singular, plural?, options?)
fazt.services.humanize.truncate(text, length, options?)
fazt.services.humanize.list(items, options?)
```

## HTTP Endpoint

Not exposed via HTTP. Humanization is a JS-side formatting operation.

## Go Library

Uses `dustin/go-humanize` with custom extensions:

```go
import "github.com/dustin/go-humanize"

func Bytes(b uint64, si bool) string {
    if si {
        return humanize.SI(float64(b), "B")
    }
    return humanize.IBytes(b)
}

func Time(t time.Time) string {
    return humanize.Time(t)
}

func Ordinal(n int) string {
    return humanize.Ordinal(n)
}

func Comma(n int64) string {
    return humanize.Comma(n)
}
```

## Common Patterns

### File Upload UI

```javascript
async function handleUpload(file) {
  const sizeStr = fazt.services.humanize.bytes(file.size);
  console.log(`Uploading ${file.name} (${sizeStr})...`);

  const start = Date.now();
  await upload(file);

  const elapsed = fazt.services.humanize.duration(Date.now() - start);
  console.log(`Completed in ${elapsed}`);
}
// "Uploading photo.jpg (2.4 MB)..."
// "Completed in 3s"
```

### Activity Feed

```javascript
async function getActivityFeed(userId) {
  const activities = await fazt.storage.ds.find('activities',
    { userId },
    { orderBy: { createdAt: 'desc' }, limit: 20 }
  );

  return activities.map(a => ({
    ...a,
    when: fazt.services.humanize.time(a.createdAt)
  }));
}
// [{ action: 'commented', when: '5 minutes ago' }, ...]
```

### Dashboard Stats

```javascript
function formatStats(stats) {
  return {
    users: fazt.services.humanize.compact(stats.totalUsers),
    storage: fazt.services.humanize.bytes(stats.storageUsed),
    requests: fazt.services.humanize.number(stats.requestCount),
    uptime: fazt.services.humanize.duration(stats.uptimeMs)
  };
}
// { users: "12.4K", storage: "4.2 GB", requests: "1,234,567", uptime: "45d 3h" }
```

### Leaderboard

```javascript
function formatLeaderboard(entries) {
  return entries.map((entry, i) => ({
    rank: fazt.services.humanize.ordinal(i + 1),
    name: entry.name,
    score: fazt.services.humanize.number(entry.score)
  }));
}
// [{ rank: "1st", name: "Alice", score: "1,234,567" }, ...]
```

### Comment Meta

```javascript
function formatComment(comment) {
  return {
    text: comment.text,
    author: comment.authorName,
    time: fazt.services.humanize.time(comment.createdAt),
    likes: fazt.services.humanize.plural(comment.likeCount, 'like')
  };
}
// { text: "Great post!", author: "Bob", time: "2 hours ago", likes: "5 likes" }
```

## Limits

| Limit | Default |
|-------|---------|
| `maxNumber` | Number.MAX_SAFE_INTEGER |
| `maxTextLength` | 10,000 chars |

## Implementation Notes

- ~20KB binary addition
- Pure Go (go-humanize has no CGO)
- All formatting is deterministic
- Locale support via Go's `x/text` package

