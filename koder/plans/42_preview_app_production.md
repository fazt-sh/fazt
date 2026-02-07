# Plan 42: Preview App — Production Quality

## Context

Preview is the reference media app for fazt. Currently a functional demo — needs to become a polished, production-quality app that showcases what the platform can do. Depends on Plan 41 (video probe) for video features.

## Current Issues

- Theme/color switching doesn't work
- No video support
- No filtering or organization
- No visual overlays or indicators
- Generic look, not production-ready

## Requirements

### Video Support
- Upload accepts images AND videos (`accept="image/*,video/mp4"`)
- Probe uploaded videos via `fazt.app.media.probe()`, store codec info in metadata
- Video thumbnails with play button overlay (center) + duration badge (bottom-right)
- Incompatible codec: warning badge, still upload
- Transcoded variant available: swap seamlessly

### Filters (Time-Based)
- Today, Yesterday, Last 7 days, Last 30 days, All
- Filter chips at top of gallery
- Server-side filtering via ds.find with date range

### Overlays & Indicators
- Videos: duration badge, play button, file size
- Images: dimensions on hover
- Watched/seen icon (eye, tracked via `fazt.app.user.kv`)
- Transcoding in progress: spinner overlay

### Icons
- Lucide (bundled or CDN)

### Theming
- Multiple color themes like admin (not just light/dark)
- Colors: slate, blue, emerald, violet, amber, rose
- Persist choice in `fazt.app.user.kv`
- Theme picker in settings that actually works

### Polish
- Grid layout with proper aspect ratios
- Loading skeletons, smooth transitions
- Empty states with helpful prompts
- Error states with retry
- Mobile-responsive
- Keyboard navigation (arrows, escape to close)
- Lightbox for full-size viewing with swipe

## Implementation Order

1. Theme system (fix what's broken, add color options)
2. Video upload + probe integration
3. Gallery grid with overlays (duration, play, seen)
4. Time-based filters
5. Lightbox with keyboard/swipe
6. Polish pass (skeletons, transitions, empty states, mobile)
7. Transcoding status UI (after Plan 41 Phase 3)
