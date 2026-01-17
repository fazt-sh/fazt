# Plan 18: App Ecosystem & CLI Restructure

**Goal**: Unified app management with POLS-compliant CLI, git-based
distribution, and `/fazt-app` skill for Claude-driven development.

**Principle**: `fazt app` is the single namespace for all app operations.
Users learn one pattern, capabilities grow monotonically.

## Part 1: CLI Restructure

### Current State (Broken POLS)

```
fazt remote deploy <dir>     # App op mixed with peer management
fazt remote apps             # App op mixed with peer management
fazt deploy                  # Legacy, unclear target
```

### Target State

```
fazt app <verb>              # All app operations
fazt remote <verb>           # Peer management only
fazt server <verb>           # Server management (unchanged)
```

### Command Matrix

| Command | Description | Replaces |
|---------|-------------|----------|
| `fazt app list [peer]` | List apps | `fazt remote apps` |
| `fazt app deploy <dir> [--to peer]` | Deploy folder | `fazt remote deploy` |
| `fazt app install <url> [--to peer]` | Install from git | NEW |
| `fazt app pull <app> [--from peer] <dir>` | Download to folder | NEW |
| `fazt app upgrade <app>` | Upgrade git-sourced app | NEW |
| `fazt app remove <app> [--from peer]` | Remove app | NEW |
| `fazt app info <app> [peer]` | Show app details | NEW |

### Remote Commands (Slimmed)

```
fazt remote add <name>       # Add peer (unchanged)
fazt remote list             # List peers (unchanged)
fazt remote remove <name>    # Remove peer (unchanged)
fazt remote default <name>   # Set default (unchanged)
fazt remote status [name]    # Check health (unchanged)
fazt remote upgrade [name]   # Upgrade fazt binary (keep here - it's infra)
```

**Removed from remote:**
- `fazt remote deploy` → `fazt app deploy --to`
- `fazt remote apps` → `fazt app list`

### Deprecation Strategy

Old commands print warning but still work for one release:

```go
// In cmd/server/remote.go
case "deploy":
    fmt.Fprintln(os.Stderr,
        "DEPRECATED: Use 'fazt app deploy <dir> --to <peer>' instead")
    return runAppDeploy(args) // Delegate to new command
case "apps":
    fmt.Fprintln(os.Stderr,
        "DEPRECATED: Use 'fazt app list [peer]' instead")
    return runAppList(args)
```

Remove deprecated shims in v0.10.0.

---

## Part 2: Database Schema Changes

### Migration 011: App Source Tracking

```sql
-- Extend apps table for source tracking
ALTER TABLE apps ADD COLUMN source_type TEXT DEFAULT 'personal';
ALTER TABLE apps ADD COLUMN source_url TEXT;
ALTER TABLE apps ADD COLUMN source_ref TEXT;
ALTER TABLE apps ADD COLUMN source_commit TEXT;
ALTER TABLE apps ADD COLUMN installed_at TEXT;
ALTER TABLE apps ADD COLUMN updated_at TEXT;

-- Index for git-sourced apps (upgrade checking)
CREATE INDEX IF NOT EXISTS idx_apps_source
    ON apps(source_type) WHERE source_type = 'git';
```

### Source Types

| Type | Description | Example |
|------|-------------|---------|
| `personal` | Deployed from local folder | `fazt app deploy ./my-site` |
| `git` | Installed from git repo | `fazt app install github:user/repo` |

### Source Fields

| Field | Purpose | Example |
|-------|---------|---------|
| `source_url` | Git repo + path | `github.com/user/repo/apps/blog` |
| `source_ref` | What user specified | `v1.2.0`, `main`, `abc1234` |
| `source_commit` | Resolved commit SHA | `abc1234def5678...` |
| `installed_at` | When installed | ISO timestamp |
| `updated_at` | Last upgrade | ISO timestamp |

---

## Part 3: Git Integration (go-git)

### Package: `internal/git/`

```go
// git.go - Core git operations
package git

import (
    "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing"
)

// CloneOptions configures a clone operation
type CloneOptions struct {
    URL       string // github.com/user/repo
    Path      string // subfolder within repo (optional)
    Ref       string // tag, branch, or commit (optional)
    TargetDir string // where to clone
}

// Clone fetches a repo (or subfolder) to local directory
func Clone(opts CloneOptions) (*CloneResult, error)

// CloneResult contains clone metadata
type CloneResult struct {
    CommitSHA   string
    CommitTime  time.Time
    RefResolved string // What ref resolved to
    Files       int    // Number of files cloned
}

// ResolveRef converts ref (tag/branch/commit) to full commit SHA
func ResolveRef(repoURL, ref string) (string, error)

// GetLatestCommit returns HEAD commit for a repo+path
func GetLatestCommit(repoURL, path string) (string, error)
```

### URL Parsing

Support multiple formats:

```go
// parser.go
type RepoRef struct {
    Host   string // github.com
    Owner  string // user
    Repo   string // repo
    Path   string // apps/blog (optional)
    Ref    string // v1.0.0 (optional)
}

// ParseURL handles various formats
func ParseURL(input string) (*RepoRef, error)

// Supported formats:
// github.com/user/repo
// github.com/user/repo/path/to/app
// github.com/user/repo/path/to/app@v1.0.0
// github.com/user/repo@main
// github:user/repo/app (shorthand)
```

### Sparse Checkout (Optimization)

For repos with many apps, only fetch the needed subfolder:

```go
func (c *CloneOptions) WithSparseCheckout() *CloneOptions {
    // Only fetch files under c.Path
    // Reduces bandwidth significantly for monorepos
}
```

---

## Part 4: CLI Implementation

### File: `cmd/server/app.go`

```go
package main

func init() {
    registerCommand("app", runApp, "App management")
}

func runApp(args []string) error {
    if len(args) == 0 {
        return runAppList(nil) // Default: list local apps
    }

    switch args[0] {
    case "list":
        return runAppList(args[1:])
    case "deploy":
        return runAppDeploy(args[1:])
    case "install":
        return runAppInstall(args[1:])
    case "pull":
        return runAppPull(args[1:])
    case "upgrade":
        return runAppUpgrade(args[1:])
    case "remove":
        return runAppRemove(args[1:])
    case "info":
        return runAppInfo(args[1:])
    default:
        return fmt.Errorf("unknown command: app %s", args[0])
    }
}
```

### Command: `fazt app install`

```go
func runAppInstall(args []string) error {
    // Parse flags
    url := args[0]
    peer := flagString("to", "")      // --to peer (optional)
    name := flagString("name", "")    // --name override (optional)

    // Parse git URL
    ref, err := git.ParseURL(url)
    if err != nil {
        return fmt.Errorf("invalid URL: %w", err)
    }

    // Clone to temp directory
    tmpDir, err := os.MkdirTemp("", "fazt-install-*")
    defer os.RemoveAll(tmpDir)

    result, err := git.Clone(git.CloneOptions{
        URL:       ref.FullURL(),
        Path:      ref.Path,
        Ref:       ref.Ref,
        TargetDir: tmpDir,
    })
    if err != nil {
        return fmt.Errorf("clone failed: %w", err)
    }

    // Read manifest to get app name
    manifest, err := readManifest(tmpDir)
    if err != nil {
        return fmt.Errorf("no manifest.json found")
    }
    appName := manifest.Name
    if name != "" {
        appName = name // Override if specified
    }

    // Deploy to target
    if peer == "" {
        // Local deployment
        return deployLocal(tmpDir, appName, &SourceInfo{
            Type:   "git",
            URL:    url,
            Ref:    ref.Ref,
            Commit: result.CommitSHA,
        })
    }

    // Remote deployment
    return deployRemote(tmpDir, appName, peer, &SourceInfo{
        Type:   "git",
        URL:    url,
        Ref:    ref.Ref,
        Commit: result.CommitSHA,
    })
}
```

### Command: `fazt app upgrade`

```go
func runAppUpgrade(args []string) error {
    appName := args[0]

    // Get app from DB
    app, err := db.GetApp(appName)
    if err != nil {
        return err
    }

    // Check if git-sourced
    if app.SourceType != "git" {
        return fmt.Errorf("%s is not installed from git", appName)
    }

    // Get latest commit
    latest, err := git.GetLatestCommit(app.SourceURL, "")
    if err != nil {
        return err
    }

    // Compare
    if latest == app.SourceCommit {
        fmt.Println("Already up to date")
        return nil
    }

    // Reinstall with same URL, new commit
    fmt.Printf("Upgrading %s: %s → %s\n",
        appName, app.SourceCommit[:7], latest[:7])

    return runAppInstall([]string{app.SourceURL})
}
```

### Command: `fazt app pull`

```go
func runAppPull(args []string) error {
    appName := args[0]
    targetDir := flagString("to", "./"+appName)
    peer := flagString("from", "")

    // Get app files from VFS
    var files []VFSFile
    if peer == "" {
        files, err = getLocalAppFiles(appName)
    } else {
        files, err = getRemoteAppFiles(appName, peer)
    }
    if err != nil {
        return err
    }

    // Write to local directory
    for _, f := range files {
        path := filepath.Join(targetDir, f.Path)
        os.MkdirAll(filepath.Dir(path), 0755)
        os.WriteFile(path, f.Content, 0644)
    }

    fmt.Printf("Pulled %d files to %s\n", len(files), targetDir)
    return nil
}
```

---

## Part 5: API Endpoints

### New Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/apps/{name}/files` | GET | List all files in app |
| `/api/apps/{name}/files/{path}` | GET | Get file content |
| `/api/apps/{name}/source` | GET | Get source metadata |

### Handler: Get App Files

```go
// handlers/apps.go
func (h *Handler) GetAppFiles(w http.ResponseWriter, r *http.Request) {
    appName := chi.URLParam(r, "name")

    files, err := h.hosting.ListFiles(appName)
    if err != nil {
        api.NotFound(w, "app not found")
        return
    }

    api.Success(w, http.StatusOK, files)
}

func (h *Handler) GetAppFile(w http.ResponseWriter, r *http.Request) {
    appName := chi.URLParam(r, "name")
    filePath := chi.URLParam(r, "*") // Wildcard for nested paths

    content, err := h.hosting.ReadFile(appName, filePath)
    if err != nil {
        api.NotFound(w, "file not found")
        return
    }

    // Return raw content with correct MIME type
    w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(filePath)))
    w.Write(content)
}
```

---

## Part 6: /fazt-app Skill

### File: `.claude/commands/fazt-app.md`

````markdown
# /fazt-app - Build Fazt Apps

Build and deploy apps to fazt instances with Claude.

## Usage

```
/fazt-app <description>
/fazt-app "build a pomodoro tracker with task persistence"
```

## Context

You are building an app for fazt.sh - a single-binary personal cloud.

### App Structure

```
my-app/
├── manifest.json      # Required: {"name": "my-app"}
├── index.html         # Entry point
├── main.js            # ES6 module entry
├── components/        # Vue components (plain JS, not SFC)
└── api/               # Serverless functions
    └── data.js        # → GET/POST /api/data
```

### Frontend Stack (Zero Build)

```html
<script type="importmap">
{
  "imports": {
    "vue": "https://unpkg.com/vue@3/dist/vue.esm-browser.js"
  }
}
</script>
<script src="https://cdn.tailwindcss.com"></script>
<script type="module" src="main.js"></script>
```

### Vue Components (Plain JS, Not SFC)

```javascript
// components/Timer.js
import { ref, computed } from 'vue'

export default {
  props: ['duration'],
  setup(props) {
    const remaining = ref(props.duration)
    const formatted = computed(() => {
      const m = Math.floor(remaining.value / 60)
      const s = remaining.value % 60
      return `${m}:${s.toString().padStart(2, '0')}`
    })
    return { remaining, formatted }
  },
  template: `
    <div class="text-4xl font-mono">{{ formatted }}</div>
  `
}
```

### Serverless Functions

```javascript
// api/tasks.js
function handler(req) {
  const db = fazt.storage.kv

  if (req.method === 'GET') {
    const tasks = db.get('tasks') || []
    return { status: 200, body: JSON.stringify(tasks) }
  }

  if (req.method === 'POST') {
    const tasks = db.get('tasks') || []
    const newTask = JSON.parse(req.body)
    tasks.push({ ...newTask, id: Date.now() })
    db.set('tasks', tasks)
    return { status: 201, body: JSON.stringify(newTask) }
  }

  return { status: 405, body: 'Method not allowed' }
}
```

### User Identification Pattern

For apps that need user-specific data without auth:

```javascript
// Get or create user ID from query string
function getUserId() {
  const params = new URLSearchParams(location.search)
  let id = params.get('u')
  if (!id) {
    id = crypto.randomUUID().split('-')[0]
    history.replaceState(null, '', `?u=${id}`)
  }
  return id
}

// Use in API calls
fetch(`/api/tasks?user=${getUserId()}`)
```

## Location Behavior

Where to create app files depends on context:

| Scenario | Location |
|----------|----------|
| In fazt repo, no flag | `servers/zyt/{app}/` |
| Not in fazt repo, no flag | `/tmp/fazt-{app}-{hash}/` |
| `--in <dir>` specified | `<dir>/{app}/` |
| `--tmp` flag | `/tmp/fazt-{app}-{hash}/` |

### Detection Logic

```
1. Check if --in or --tmp flag provided → use that
2. Check if cwd is fazt repo (has CLAUDE.md with "fazt")
   → Yes: use servers/zyt/{app}/
   → No: use /tmp/fazt-{app}-{hash}/
```

### Examples

```bash
# In fazt repo (persisted, can commit to git)
/fazt-app "pomodoro tracker"
→ Creates: servers/zyt/pomodoro/

# Explicit directory
/fazt-app "game" --in ~/projects/games
→ Creates: ~/projects/games/game/

# Force temp (throwaway)
/fazt-app "quick test" --tmp
→ Creates: /tmp/fazt-quick-test-a1b2c3/
```

## Workflow

1. Determine location (see Location Behavior above)
2. Create app folder with manifest.json
3. Write all files (html, js, api)
4. Deploy: `fazt app deploy <folder> --to local`
5. Report URL: `http://{name}.192.168.64.3:8080`
6. For edits: modify files in same folder, redeploy

## Deployment

```bash
# To local instance
fazt app deploy servers/zyt/myapp/ --to local

# To production
fazt app deploy servers/zyt/myapp/ --to zyt
```

## Design Guidelines

- Modern, clean UI with Tailwind
- Dark mode support (respect prefers-color-scheme)
- Mobile-first responsive design
- Minimal dependencies (prefer native APIs)
- No build step required
````

---

## Part 7: Tests

### Test: CLI Commands

```go
// cmd/server/app_test.go
func TestAppCommands(t *testing.T) {
    t.Run("app list shows apps", func(t *testing.T) {
        // Setup test DB with apps
        // Run command, capture output
        // Assert output contains expected apps
    })

    t.Run("app deploy creates app", func(t *testing.T) {
        // Create temp dir with manifest
        // Run deploy command
        // Assert app exists in DB
    })

    t.Run("app install from github", func(t *testing.T) {
        // Mock git clone
        // Run install command
        // Assert app created with source_type = 'git'
    })

    t.Run("app pull downloads files", func(t *testing.T) {
        // Setup app with files
        // Run pull command
        // Assert files exist in target dir
    })

    t.Run("app upgrade checks git source", func(t *testing.T) {
        // Create git-sourced app
        // Mock newer commit available
        // Run upgrade, assert new commit installed
    })
}
```

### Test: Git Package

```go
// internal/git/git_test.go
func TestParseURL(t *testing.T) {
    cases := []struct {
        input string
        want  RepoRef
    }{
        {
            "github.com/user/repo",
            RepoRef{Host: "github.com", Owner: "user", Repo: "repo"},
        },
        {
            "github.com/user/repo/apps/blog",
            RepoRef{Host: "github.com", Owner: "user", Repo: "repo",
                Path: "apps/blog"},
        },
        {
            "github.com/user/repo@v1.0.0",
            RepoRef{Host: "github.com", Owner: "user", Repo: "repo",
                Ref: "v1.0.0"},
        },
        {
            "github.com/user/repo/app@main",
            RepoRef{Host: "github.com", Owner: "user", Repo: "repo",
                Path: "app", Ref: "main"},
        },
    }

    for _, tc := range cases {
        got, err := ParseURL(tc.input)
        assert.NoError(t, err)
        assert.Equal(t, tc.want, *got)
    }
}

func TestClone(t *testing.T) {
    // Integration test with real GitHub repo
    // Use a known small public repo
    t.Run("clones public repo", func(t *testing.T) {
        tmpDir := t.TempDir()
        result, err := Clone(CloneOptions{
            URL:       "github.com/fazt-sh/example-app",
            TargetDir: tmpDir,
        })
        assert.NoError(t, err)
        assert.NotEmpty(t, result.CommitSHA)
        assert.FileExists(t, filepath.Join(tmpDir, "manifest.json"))
    })
}
```

### Test: API Endpoints

```go
// internal/handlers/apps_test.go
func TestGetAppFiles(t *testing.T) {
    h := setupTestHandler(t)

    // Create test app with files
    h.hosting.Deploy("test-app", map[string][]byte{
        "index.html": []byte("<h1>Test</h1>"),
        "api/data.js": []byte("function handler(){}"),
    })

    // Request file list
    req := httptest.NewRequest("GET", "/api/apps/test-app/files", nil)
    w := httptest.NewRecorder()
    h.GetAppFiles(w, req)

    assert.Equal(t, 200, w.Code)
    var files []string
    json.Unmarshal(w.Body.Bytes(), &files)
    assert.Contains(t, files, "index.html")
    assert.Contains(t, files, "api/data.js")
}
```

---

## Part 8: Implementation Order

### Phase 1: CLI Restructure (Breaking Change)

1. Create `cmd/server/app.go` with new command structure
2. Move deploy logic from `remote.go` to `app.go`
3. Update `remote.go` to remove app commands (with deprecation)
4. Update help text and examples
5. Add tests for new CLI structure

### Phase 2: Database & Source Tracking

1. Create migration 011 for source columns
2. Update deploy handlers to set source_type='personal'
3. Add source info to app list/info output
4. Add tests for source tracking

### Phase 3: Git Integration

1. Add go-git dependency: `go get github.com/go-git/go-git/v5`
2. Create `internal/git/` package
3. Implement URL parser with tests
4. Implement Clone with sparse checkout
5. Integration tests with real repos

### Phase 4: Install & Upgrade Commands

1. Implement `fazt app install`
2. Implement `fazt app upgrade`
3. Implement `fazt app info` (show source details)
4. Add tests

### Phase 5: Pull Command & API

1. Add `/api/apps/{name}/files` endpoint
2. Implement `fazt app pull`
3. Add tests

### Phase 6: Skill

1. Create `.claude/commands/fazt-app.md`
2. Test skill with sample app builds
3. Document in CLAUDE.md

---

## Files to Create/Modify

| File | Action | Phase |
|------|--------|-------|
| `cmd/server/app.go` | Create | 1 |
| `cmd/server/app_test.go` | Create | 1 |
| `cmd/server/remote.go` | Modify (slim down) | 1 |
| `internal/database/migrations/011_source.sql` | Create | 2 |
| `internal/database/db.go` | Add migration 11 | 2 |
| `internal/git/git.go` | Create | 3 |
| `internal/git/parser.go` | Create | 3 |
| `internal/git/git_test.go` | Create | 3 |
| `internal/handlers/apps.go` | Modify (add endpoints) | 5 |
| `internal/handlers/apps_test.go` | Modify (add tests) | 5 |
| `.claude/commands/fazt-app.md` | Create | 6 |
| `CLAUDE.md` | Update CLI docs | 6 |
| `go.mod` | Add go-git | 3 |

---

## Version

This would be **v0.10.0** - major CLI restructure + new capabilities.

### Changelog Entry

```markdown
## v0.10.0 - App Ecosystem

### Breaking Changes
- `fazt remote deploy` → `fazt app deploy --to <peer>`
- `fazt remote apps` → `fazt app list [peer]`

### New Features
- `fazt app install <url>` - Install apps from GitHub
- `fazt app upgrade <app>` - Upgrade git-sourced apps
- `fazt app pull <app>` - Download app to local folder
- `fazt app info <app>` - Show app details and source
- Source tracking for installed apps
- `/fazt-app` Claude skill for app development

### Dependencies
- Added go-git for git operations (+5MB binary size)
```

---

## Future Considerations (Not in This Plan)

These are explicitly deferred:

| Feature | Trigger to Implement |
|---------|---------------------|
| Private repo auth | User requests private repo support |
| SSH key management | User prefers SSH over HTTPS |
| Commit signing | Marketplace trust requirements |
| VFS versioning (fazt.git) | Agent self-modification needs |
| Push to GitHub | Agent publishing needs |

The schema and API are designed to accommodate these without changes.
