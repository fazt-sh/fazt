# Plan 44: Drop — File & Folder Hosting via Fazt

## Status: IDEA

## Problem

fazt-sdk needs to be consumable outside the monorepo. npm is overhead for a
single-user ecosystem. We need a way to host built artifacts (like `fazt-sdk.mjs`)
at stable URLs. More broadly, fazt lacks a simple "upload and get a link" feature.

## Idea

A fazt-app called "drop" — personal Dropbox-style file hosting.

```
fazt upload ./flowers.zip     → https://drop.zyt.app/a1b2/flowers.zip
fazt upload ./photos/         → https://drop.zyt.app/c3d4/photos/
fazt upload ./fazt-sdk.mjs    → https://drop.zyt.app/e5f6/fazt-sdk.mjs
```

### URL Format

```
drop.<domain>/<short-hash>/<original-name>
```

- **Short hash** — content-addressable, deduplicates
- **Original name** — preserved for human readability and browser downloads
- Folders get an index: `drop.zyt.app/<hash>/` lists contents

### Why a fazt-app (not core)

- Uses existing blob storage, auth, HTTP serving
- No changes to the binary
- `fazt upload` is just a CLI alias for `fazt @<peer> app exec drop upload <file>`
- Folder upload is basically a lightweight `app deploy` without a manifest

### Features (MVP)

- Single file upload → permanent URL
- Folder upload → directory listing + file serving
- Content-type detection (serves `.mjs` as `application/javascript`)
- Auth-gated uploads, public reads

### Features (Later)

- Expiring links
- Private/auth-gated downloads
- Upload history / management page
- Bandwidth/storage limits

## Immediate Use Case

Host `fazt-sdk.mjs` so external app repos can import it:
```js
import { createAppClient } from 'https://drop.zyt.app/e5f6/fazt-sdk.mjs'
```

## SDK Distribution (Current State)

For now, the SDK works via relative paths within the monorepo. Bundle is built
with `packages/fazt-sdk/build.sh` → `dist/fazt-sdk.mjs` (32KB unminified).

When Drop exists, the release process adds one step:
```bash
./packages/fazt-sdk/build.sh
fazt upload ./packages/fazt-sdk/dist/fazt-sdk.mjs
```
