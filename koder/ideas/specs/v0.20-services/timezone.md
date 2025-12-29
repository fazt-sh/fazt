# Timezone Service

## Summary

Native timezone handling with embedded IANA timezone database. Convert, format,
and reason about time across the world's 400+ timezones without external deps.

## The Problem

```javascript
// JavaScript Date is timezone-hostile
new Date().toLocaleString('en-US', { timeZone: 'America/New_York' })
// Works in browsers, but...
// - Node.js depends on system tzdata
// - Edge runtimes may lack tzdata
// - Embedded systems often have stale/missing data

// Real-world needs:
// - "What time is 3pm Tokyo in New York?"
// - "Is DST active in Berlin right now?"
// - "Schedule this for 9am user's local time"
// - "Show all times in the user's timezone"
```

## The Solution

Fazt embeds the full IANA timezone database (~500KB). Every Fazt instance,
regardless of host OS, has correct, consistent timezone handling.

```javascript
fazt.services.timezone.convert('2024-03-15T15:00:00', 'Asia/Tokyo', 'America/New_York')
// "2024-03-15T02:00:00" (1am due to DST)

fazt.services.timezone.now('Europe/London')
// { time: "2024-12-29T14:30:00", offset: "+00:00", isDST: false }
```

## Usage

### Get Current Time

```javascript
// Current time in a timezone
fazt.services.timezone.now('America/New_York')
// {
//   time: "2024-12-29T09:30:45",
//   offset: "-05:00",
//   offsetSeconds: -18000,
//   isDST: false,
//   abbreviation: "EST",
//   timezone: "America/New_York"
// }

fazt.services.timezone.now('Europe/Paris')
// {
//   time: "2024-12-29T15:30:45",
//   offset: "+01:00",
//   offsetSeconds: 3600,
//   isDST: false,
//   abbreviation: "CET",
//   timezone: "Europe/Paris"
// }
```

### Convert Between Timezones

```javascript
// Convert a specific time
fazt.services.timezone.convert(
  '2024-07-15T14:00:00',  // time
  'America/Los_Angeles',  // from
  'Asia/Tokyo'            // to
)
// "2024-07-16T06:00:00"

// With full details
fazt.services.timezone.convert(
  '2024-07-15T14:00:00',
  'America/Los_Angeles',
  'Asia/Tokyo',
  { details: true }
)
// {
//   from: { time: "2024-07-15T14:00:00", timezone: "America/Los_Angeles", offset: "-07:00" },
//   to: { time: "2024-07-16T06:00:00", timezone: "Asia/Tokyo", offset: "+09:00" },
//   differenceHours: 16
// }
```

### Parse with Timezone

```javascript
// Parse time with timezone context
fazt.services.timezone.parse('2024-03-10T02:30:00', 'America/New_York')
// null - this time doesn't exist (DST skip)

fazt.services.timezone.parse('2024-11-03T01:30:00', 'America/New_York')
// Ambiguous - returns first occurrence by default
// {
//   time: "2024-11-03T01:30:00",
//   offset: "-04:00",  // EDT (before fall back)
//   ambiguous: true
// }

fazt.services.timezone.parse('2024-11-03T01:30:00', 'America/New_York', { prefer: 'later' })
// {
//   time: "2024-11-03T01:30:00",
//   offset: "-05:00",  // EST (after fall back)
//   ambiguous: true
// }
```

### Format in Timezone

```javascript
// Format a UTC timestamp in a specific timezone
fazt.services.timezone.format(1703854245000, 'Asia/Tokyo')
// "2024-12-29T23:30:45"

fazt.services.timezone.format(1703854245000, 'Asia/Tokyo', {
  format: 'full'
})
// "Sunday, December 29, 2024 at 11:30:45 PM JST"

fazt.services.timezone.format(1703854245000, 'Asia/Tokyo', {
  format: 'date'
})
// "2024-12-29"

fazt.services.timezone.format(1703854245000, 'Asia/Tokyo', {
  format: 'time'
})
// "23:30:45"

// Custom format
fazt.services.timezone.format(1703854245000, 'Asia/Tokyo', {
  pattern: 'YYYY-MM-DD HH:mm'
})
// "2024-12-29 23:30"
```

### DST Information

```javascript
// Check if DST is active
fazt.services.timezone.isDST('America/New_York')
// false (in December)

fazt.services.timezone.isDST('America/New_York', '2024-07-15T12:00:00')
// true (in July)

// Get DST transitions
fazt.services.timezone.transitions('America/New_York', 2024)
// [
//   { type: 'start', time: '2024-03-10T02:00:00', offset: '-04:00' },
//   { type: 'end', time: '2024-11-03T02:00:00', offset: '-05:00' }
// ]

fazt.services.timezone.transitions('Asia/Tokyo', 2024)
// [] - Japan doesn't observe DST
```

### Timezone Information

```javascript
// Get timezone details
fazt.services.timezone.info('America/New_York')
// {
//   name: "America/New_York",
//   currentOffset: "-05:00",
//   currentAbbr: "EST",
//   isDST: false,
//   standardOffset: "-05:00",
//   dstOffset: "-04:00",
//   country: "US"
// }

// List all timezones
fazt.services.timezone.list()
// ["Africa/Abidjan", "Africa/Accra", ... "Pacific/Wallis"]

// List by region
fazt.services.timezone.list({ region: 'America' })
// ["America/Adak", "America/Anchorage", ...]

// List by country
fazt.services.timezone.list({ country: 'JP' })
// ["Asia/Tokyo"]

// Search timezones
fazt.services.timezone.search('york')
// ["America/New_York"]

fazt.services.timezone.search('pacific')
// ["America/Los_Angeles", "Pacific/Auckland", "Pacific/Fiji", ...]
```

### Offset Calculations

```javascript
// Get offset between timezones
fazt.services.timezone.offset('America/New_York', 'Asia/Tokyo')
// 14 (hours, can vary with DST)

fazt.services.timezone.offset('America/New_York', 'Asia/Tokyo', '2024-07-15T12:00:00')
// 13 (during EDT)

// Offset from UTC
fazt.services.timezone.offsetFromUTC('America/New_York')
// -5 (or -4 during DST)
```

### Scheduling Helpers

```javascript
// "Next 9am in New York"
fazt.services.timezone.next('09:00', 'America/New_York')
// Returns UTC timestamp for next 9am NY time

// "9am every day in user's timezone"
fazt.services.timezone.scheduleDaily('09:00', 'Europe/London')
// {
//   nextRun: 1703930400000,  // UTC timestamp
//   cronUTC: "0 9 * * *",    // Only valid when UK is on GMT
//   note: "Shifts with DST"
// }

// Check if a time is within business hours
fazt.services.timezone.isWithin(
  Date.now(),
  'America/New_York',
  { start: '09:00', end: '17:00', weekdays: true }
)
// true/false
```

## JS API

```javascript
// Current time
fazt.services.timezone.now(tz)

// Conversion
fazt.services.timezone.convert(time, fromTz, toTz, options?)

// Parsing
fazt.services.timezone.parse(time, tz, options?)
// options: { prefer: 'earlier' | 'later' }

// Formatting
fazt.services.timezone.format(timestamp, tz, options?)
// options: { format: 'iso' | 'full' | 'date' | 'time', pattern: string }

// DST
fazt.services.timezone.isDST(tz, time?)
fazt.services.timezone.transitions(tz, year)

// Info
fazt.services.timezone.info(tz)
fazt.services.timezone.list(options?)
fazt.services.timezone.search(query)

// Offsets
fazt.services.timezone.offset(fromTz, toTz, time?)
fazt.services.timezone.offsetFromUTC(tz, time?)

// Scheduling
fazt.services.timezone.next(time, tz)
fazt.services.timezone.scheduleDaily(time, tz)
fazt.services.timezone.isWithin(timestamp, tz, range)
```

## HTTP Endpoint

Not exposed via HTTP. Timezone operations are JS-side calculations.

## Go Implementation

Uses Go's `time` package with embedded tzdata:

```go
import (
    "time"
    _ "time/tzdata" // Embed IANA database
)

func Now(tz string) (*TimeInfo, error) {
    loc, err := time.LoadLocation(tz)
    if err != nil {
        return nil, err
    }
    t := time.Now().In(loc)
    name, offset := t.Zone()
    return &TimeInfo{
        Time:          t.Format(time.RFC3339),
        Offset:        formatOffset(offset),
        OffsetSeconds: offset,
        IsDST:         t.IsDST(),
        Abbreviation:  name,
        Timezone:      tz,
    }, nil
}

func Convert(timeStr, fromTz, toTz string) (string, error) {
    from, _ := time.LoadLocation(fromTz)
    to, _ := time.LoadLocation(toTz)
    t, _ := time.ParseInLocation(time.RFC3339, timeStr, from)
    return t.In(to).Format(time.RFC3339), nil
}
```

## Common Patterns

### User Timezone Storage

```javascript
// Store user's timezone preference
await fazt.storage.ds.update('users', { id: userId }, {
  timezone: 'America/New_York'
});

// Display times in user's timezone
async function formatForUser(userId, timestamp) {
  const user = await fazt.storage.ds.findOne('users', { id: userId });
  return fazt.services.timezone.format(timestamp, user.timezone || 'UTC');
}
```

### Global Event Scheduling

```javascript
// Event in organizer's timezone, shown in viewer's timezone
async function getEventTime(eventId, viewerTz) {
  const event = await fazt.storage.ds.findOne('events', { id: eventId });

  return {
    original: `${event.startTime} (${event.timezone})`,
    local: fazt.services.timezone.convert(
      event.startTime,
      event.timezone,
      viewerTz
    )
  };
}
```

### Daily Digest

```javascript
// Send digest at 8am user's local time
async function scheduleDailyDigest(userId) {
  const user = await fazt.storage.ds.findOne('users', { id: userId });
  const nextRun = fazt.services.timezone.next('08:00', user.timezone);

  await fazt.storage.ds.insert('scheduled_jobs', {
    userId,
    type: 'daily_digest',
    runAt: nextRun
  });
}
```

### Business Hours Check

```javascript
// Route support to available team
async function routeSupport(ticket) {
  const teams = [
    { name: 'US', tz: 'America/New_York', hours: { start: '09:00', end: '17:00' } },
    { name: 'EU', tz: 'Europe/London', hours: { start: '09:00', end: '17:00' } },
    { name: 'APAC', tz: 'Asia/Tokyo', hours: { start: '09:00', end: '17:00' } }
  ];

  const available = teams.find(team =>
    fazt.services.timezone.isWithin(Date.now(), team.tz, {
      ...team.hours,
      weekdays: true
    })
  );

  return available || teams[0]; // Default to US if none available
}
```

### Timezone Picker UI

```javascript
// Grouped timezone list for UI
function getTimezoneOptions() {
  const regions = ['America', 'Europe', 'Asia', 'Pacific', 'Africa', 'Australia'];

  return regions.map(region => ({
    label: region,
    options: fazt.services.timezone.list({ region }).map(tz => {
      const info = fazt.services.timezone.info(tz);
      return {
        value: tz,
        label: `${tz.replace('_', ' ')} (${info.currentOffset})`
      };
    })
  }));
}
```

## IANA Database

The embedded database includes all 400+ IANA timezones:

| Region | Count | Examples |
|--------|-------|----------|
| Africa | 50+ | Africa/Cairo, Africa/Johannesburg |
| America | 140+ | America/New_York, America/Los_Angeles |
| Asia | 80+ | Asia/Tokyo, Asia/Shanghai |
| Australia | 15+ | Australia/Sydney, Australia/Perth |
| Europe | 60+ | Europe/London, Europe/Paris |
| Pacific | 40+ | Pacific/Auckland, Pacific/Honolulu |

Updated with each Go release (IANA updates ~10x/year).

## Limits

| Limit | Default |
|-------|---------|
| `listMax` | 500 (all timezones fit) |
| `yearRange` | 1970-2100 for transitions |

## Implementation Notes

- ~500KB binary addition (embedded tzdata)
- Pure Go (no CGO, no external deps)
- Timezone data survives without internet
- Deterministic across all Fazt instances
- Go 1.15+ embeds tzdata via `time/tzdata` import

