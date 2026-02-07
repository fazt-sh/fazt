# Issue 01: Custom URL Upgrade & fazt-releases CDN

**Created**: 2026-02-08
**Status**: Broken - Auth not working
**Version**: v0.28.0

## Context

Implemented custom URL upgrade support to allow `fazt upgrade <url>` instead of only GitHub releases. Built fazt-releases.zyt.app as a self-hosted binary CDN using BFBB pattern.

## Motivation

1. **Keep repo private**: Don't want to expose code in public GitHub, but still need binary distribution
2. **Custom CDN**: Self-hosted release management with auto-cleanup (keep last 3 versions)
3. **Dogfooding**: Use Fazt to host Fazt releases - BFBB paradigm showcase

## Implementation

### 1. CLI Changes (Custom URL Upgrade)

Modified fazt binary to accept optional URL parameter:

```bash
# Old behavior (GitHub only)
fazt upgrade

# New behavior (custom URL support)
fazt upgrade https://fazt-releases.zyt.app/api/releases/latest/download
fazt @zyt upgrade https://custom-cdn.com/releases/latest
```

**Modified Files:**

#### `cmd/server/main.go`
```go
func handleUpgradeCommand() {
  var customURL string
  if len(os.Args) > 2 {
    customURL = os.Args[2]
  }
  if err := provision.Upgrade(config.Version, customURL); err != nil {
    // ...
  }
}
```

#### `internal/provision/upgrade.go`
```go
func Upgrade(currentVersion string, customURL string) error {
  if customURL != "" {
    fmt.Printf("Upgrading from custom URL: %s\n", customURL)
    assetURL = customURL
    assetName = filepath.Base(customURL)
  } else {
    // Standard GitHub release flow
    release, err := getLatestRelease()
    // ...
  }
  // Download, verify, replace binary
}
```

#### `internal/handlers/upgrade_handler.go`
```go
func handleUpgrade(w http.ResponseWriter, r *http.Request) {
  customURL := r.URL.Query().Get("url")

  if customURL != "" {
    downloadURL = customURL
    latestVersion = "custom"
  } else {
    // Get latest release from GitHub
    release, err := getLatestGitHubRelease()
    // ...
  }
}
```

#### `internal/remote/client.go`
```go
func (c *Client) UpgradeWithURL(checkOnly bool, customURL string) (*UpgradeResponse, error) {
  path := "/api/upgrade"
  params := []string{}

  if customURL != "" {
    params = append(params, "url="+url.QueryEscape(customURL))
  }
  if checkOnly {
    params = append(params, "check=true")
  }

  if len(params) > 0 {
    path += "?" + strings.Join(params, "&")
  }
  // ...
}
```

### 2. fazt-releases App (BFBB)

Built as proper BFBB app: single HTML, Vue 3 from CDN, serverless API, no build step.

**Structure:**
```
servers/zyt/fazt-releases/
├── manifest.json          # App metadata
├── index.html             # BFBB app with granular components
├── api/main.js            # Serverless API (ES5/Goja)
├── favicon.png
├── apple-touch-icon.png
└── README.md
```

#### API Endpoints (`api/main.js`)

**Public endpoints:**
- `GET /api/releases` - List all releases (sorted by version, newest first)
- `GET /api/releases/latest/download` - Download latest release (auto-increments download counter)
- `GET /api/releases/:id/download` - Download specific release

**Admin-only endpoints:**
- `POST /api/releases` - Upload new release (requires `user.isAdmin`)
  - Accepts: `file` (FormData), `version`, `description`
  - Auto-cleanup: Keeps only last 3 versions
- `DELETE /api/releases/:id` - Delete release (requires `user.isAdmin`)

**Auth endpoints:**
- `GET /api/me` - Get current user (returns `{user: {...}}` or `{user: null}`)
- `GET /api/login` - Redirect to OAuth login
- `GET /api/logout` - Redirect to OAuth logout

**Storage:**
```javascript
var ds = fazt.storage.ds  // Document store (metadata)
var s3 = fazt.storage.s3  // Blob storage (binaries)

// Version comparison
function compareVersions(a, b) {
  var va = parseVersion(a)  // {major, minor, patch, raw}
  var vb = parseVersion(b)
  // Returns -1, 0, 1
}

// Auto-cleanup (keeps last 3)
function cleanupOldReleases() {
  var releases = ds.find('releases', {})
  releases.sort(function(a, b) {
    return compareVersions(b.version, a.version)
  })
  if (releases.length > 3) {
    for (var i = 3; i < releases.length; i++) {
      s3.delete('releases/' + release.id)
      ds.delete('releases', { id: release.id })
    }
  }
}
```

#### Frontend (`index.html`)

**BFBB Pattern:**
- Single HTML file (~800 lines)
- Vue 3 from CDN: `https://unpkg.com/vue@3/dist/vue.global.prod.js`
- Lucide icons: `https://unpkg.com/lucide@latest/dist/umd/lucide.js`
- Zero build step, zero dependencies, zero package.json

**Granular Components:**

```javascript
// Icon refresh pattern (Vue reactivity)
let iconsPending = false;
function refreshIcons() {
  if (iconsPending) return;
  iconsPending = true;
  setTimeout(() => {
    iconsPending = false;
    lucide.createIcons();
  }, 0);
}

// ReleaseCard component
const ReleaseCard = {
  props: ['release', 'isAdmin'],
  emits: ['delete', 'copy'],
  template: `
    <div class="release-card">
      <div class="release-header">
        <span class="release-version">{{ release.version }}</span>
        <span v-if="release.isLatest" class="release-badge">Latest</span>
      </div>
      <p class="release-description">{{ release.description }}</p>
      <div class="release-actions">
        <a :href="'/api/releases/' + release.id + '/download'" class="release-btn">
          <i data-lucide="download"></i> Download
        </a>
        <button @click="$emit('copy', release.id)" class="release-btn-secondary">
          <i data-lucide="clipboard"></i> Copy Command
        </button>
        <button v-if="isAdmin" @click="$emit('delete', release.id)" class="release-btn-danger">
          <i data-lucide="trash-2"></i> Delete
        </button>
      </div>
    </div>
  `,
  setup() {
    onMounted(refreshIcons);
    onUpdated(refreshIcons);
  }
};

// UploadModal component
const UploadModal = {
  props: ['show'],
  emits: ['close', 'upload'],
  // File upload with drag-and-drop, progress tracking
};

// Main app
createApp({
  components: { ReleaseCard, UploadModal },
  setup() {
    const { ref, computed, onMounted, onUpdated } = Vue;

    const user = ref(null);
    const authLoading = ref(true);
    const isAdmin = computed(() => user.value && user.value.isAdmin);

    async function loadAuth() {
      try {
        const res = await fetch('/api/me');
        const data = await res.json();
        user.value = data.user;
      } finally {
        authLoading.value = false;
      }
    }

    function login() {
      window.location.href = '/api/login';
    }

    function logout() {
      window.location.href = '/api/logout';
    }

    onMounted(() => {
      loadAuth();
      loadReleases();
      refreshIcons();
    });

    return { user, isAdmin, login, logout, /* ... */ };
  }
});
```

**Auth UI:**
- Non-authenticated: Shows "Sign In" button
- Authenticated (non-admin): Shows email, "Sign Out" button, NO upload/delete
- Authenticated (admin): Shows email, "Sign Out", upload FAB, delete buttons

### 3. Helper Script

Created `scripts/upload-release.sh`:
```bash
#!/bin/bash
VERSION=$1
TOKEN=$(cat ~/.fazt/peers/zyt.json | jq -r .token)

curl -X POST https://fazt-releases.zyt.app/api/releases \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@fazt-${VERSION}-linux-amd64.tar.gz" \
  -F "version=${VERSION}" \
  -F "description=Release ${VERSION}"
```

## The Problem

**Auth is completely broken**. The user can't sign in or sign out.

### Attempted Fixes

1. **First attempt**: Built with Vite/npm/Tailwind - WRONG, not BFBB
2. **Second attempt**: Rebuilt as BFBB with syntax errors (backtick escaping)
3. **Third attempt**: Fixed syntax, forgot granular components
4. **Fourth attempt**: Added granular components, worked locally
5. **Fifth attempt**: Added lucide icons and auth endpoints
   - API expected POST for /api/login and /api/logout
   - Frontend was doing GET (window.location.href)
   - Fixed API to accept GET
   - Added logout button with log-out icon
6. **Sixth attempt (current)**: Redeployed with GET endpoints
   - **Still broken** - auth not working

### What's Wrong

Need to investigate:
- Is OAuth configured for fazt-releases app?
- Does `fazt.auth.getUser()` work in serverless context?
- Are login/logout URLs being generated correctly?
- Is the OAuth callback URL configured?
- Does the session persist after OAuth redirect?

### Related Changes

All modified files in v0.28.0:
- `cmd/server/main.go` - Parse custom URL from CLI args
- `internal/provision/upgrade.go` - Accept customURL parameter
- `internal/handlers/upgrade_handler.go` - Accept url query param
- `internal/remote/client.go` - Add UpgradeWithURL method
- `servers/zyt/fazt-releases/*` - BFBB app (not tracked in git)
- `LICENSE` - Changed to restrictive source-available license

## Testing

**Local test (worked):**
```bash
cd servers/zyt/fazt-releases
fazt @local app deploy .
# Visit http://192.168.64.3:3000/fazt-releases
```

**Production deploy (broken):**
```bash
fazt @zyt app deploy .
# Visit https://fazt-releases.zyt.app
# Sign In button doesn't work
```

## Next Steps

1. **Debug OAuth**: Check if fazt-releases has OAuth configured
   - Check database: `SELECT * FROM oauth_configs WHERE app_id = 'fazt-releases'`
   - Check if `fazt.auth.getLoginURL()` returns valid URL
2. **Test serverless auth**: Verify `fazt.auth.getUser()` works in api/main.js
3. **Check session**: Verify session cookie persists after OAuth redirect
4. **Simplify**: Maybe remove auth entirely for now, make it public upload?
5. **Fallback**: Use API tokens instead of OAuth for admin operations?

## Lessons Learned

1. **BFBB is hard**: Easy to slip into build-tool thinking
2. **Backticks in non-template context**: Use string concatenation instead
3. **GET vs POST**: Login redirects should be GET, not POST
4. **Test before deploy**: Should have tested auth flow locally first
5. **OAuth is complex**: Maybe too complex for this use case?

## Open Questions

1. How do you configure OAuth for a fazt app?
2. Does every app need separate OAuth config or inherit from instance?
3. Should fazt-releases use simpler auth (API tokens, basic auth)?
4. Is the session cookie domain-scoped correctly for zyt.app?
5. Should we just make uploads public and rely on unlisted app visibility?
