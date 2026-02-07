# Plan 41: Video Support — Probe + Graceful Transcoding

## Context

Media processing is a platform concern. Images are done: on-demand resize, user-scoped cache, in-memory LRU, step-snapped widths. This plan adds video support to the platform.

Constraint: $6 VPS (1 vCPU, 512MB). 1 worker ≈ 1900 transcodes/day > 1000 videos/day demand. Optimize for daily throughput, not latency. No ffmpeg = no queue, no error, serve original.

## Phase 1: Video Probe (Pure Go, Zero Deps)

Parse MP4 box structure to detect codec. No ffmpeg, no CGo.

```js
var info = fazt.app.media.probe(file.data)
// { container: "mp4", videoCodec: "hevc", audioCodec: "aac",
//   width: 1920, height: 1080, duration: 12.5, compatible: false }
```

- `internal/services/media/probe.go` — ~150-200 lines byte parsing
- Parse `ftyp` → container, walk `moov/trak/mdia/minf/stbl/stsd` → codec fourcc
- `compatible = videoCodec == "h264" && audioCodec in (aac, "")`
- Expose as `fazt.app.media.probe(data)` binding

## Phase 2: system.Limits.Video

```go
type Video struct {
    FFmpegAvailable bool  // exec.LookPath("ffmpeg") at startup
    Concurrency     int   // 1 on $6 VPS, 2 on $40+
    MaxDurationSec  int   // 120s on $6, 300s on $20, 600s on $40+
    MaxInputMB      int   // 100 / 200 / 500
    OutputMaxHeight int   // 720p / 1080p / 1080p
}
```

## Phase 3: Worker-Based Transcoding

Uses existing `fazt.worker.spawn()`. Graceful degradation:

```
Upload → store original → probe → compatible? done
  → NOT compatible + ffmpeg available → queue worker (nice +19, -threads 1)
  → NOT compatible + NO ffmpeg → done, serve original. No queue, no error.
```

Worker: `ffmpeg -i input -c:v libx264 -preset medium -crf 23 -c:a aac -movflags +faststart output.mp4`

Serving: prefer H.264 variant if exists, fall back to original.

## Phase 4 (Future): Distributed Processing

- Beefy peer downloads, transcodes, re-uploads
- Worker configured to call external API (Cloudflare Stream, etc.)
- Both fit existing worker model — just different execution backends

## Throughput Analysis

### Target scenario: "Alt TikTok on $6 VPS"
- 100 users, 10 videos/day each = 1000 videos/day
- Each video ~10MB, ~30s, non-compatible (HEVC from iPhone)
- Peak simultaneous uploads: 10

### Capacity ($6 VPS, 1 vCPU, 512MB)
- 1 transcode of 10MB/30s clip ≈ 45 seconds (nice +19, -threads 1)
- 1 worker = ~80 transcodes/hour = ~1920/day
- 1000 demand/day < 1920 capacity/day → queue drains in ~12 hours
- Backlog never grows unboundedly

### Resource usage during transcode
- CPU: ~100% of one core (nice +19 so server stays responsive)
- RAM: ~150-250MB per transcode (tight on 512MB, fine on 1GB)
- Disk: ~100MB temp (input + output)

## Verification Tests

### Probe tests
- Detect H.264 codec from real MP4 file → `videoCodec: "h264", compatible: true`
- Detect HEVC codec from MOV file → `videoCodec: "hevc", compatible: false`
- Detect VP9 from WebM → `videoCodec: "vp9", compatible: false`
- Non-video file → graceful error, not a crash
- Extract correct width, height, duration from known test files

### Graceful degradation tests
- No ffmpeg on system → `FFmpegAvailable: false`, no workers queued
- Upload non-compatible video without ffmpeg → stored, served as-is, no error
- Upload compatible video (H.264) → no transcode queued even with ffmpeg

### Transcoding tests (require ffmpeg)
- HEVC input → H.264 output, valid MP4 with `+faststart`
- Output respects `OutputMaxHeight` (e.g., 1080p → 720p on $6 VPS)
- Worker runs at nice +19, -threads 1
- Worker timeout respected (30min max from worker system)
- Concurrent limit enforced: only N transcodes at once

### Throughput/load tests
- Queue 10 transcode jobs → processed sequentially (Concurrency=1)
- Server stays responsive during transcoding (nice +19 verified)
- media.serve() returns original while transcode is pending
- media.serve() returns H.264 variant after transcode completes
- Transcode of 10MB/30s clip completes in < 90 seconds on test hardware

### Integration tests
- Upload → probe → queue → transcode → serve H.264 (full lifecycle)
- Upload → delete original → cache + variant cleaned up
- Delete user → all originals + variants + transcoded files removed

## Implementation Order

1. Video probe + binding (pure Go)
2. system.Limits.Video struct + defaults
3. Worker transcoding (if ffmpeg available)
4. media.serve() video awareness — prefer H.264 variant
5. Verification tests (all sections above)

## Already Done

- `internal/services/media/` — transform, cache, memcache, ProcessAndCache
- `fazt.app.user.media.serve(path)` + `fazt.app.media.serve(path)`
- In-memory LRU, step-snapped widths, concurrency limiter
- Cache invalidation on s3.put/delete (DB + memory)
