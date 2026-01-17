# Plan 19: Vite Dev Enhancement

**Goal**: Unified build/deploy model with progressive enhancement. Vite-capable
machines get optimized builds; minimal machines get working apps via Go fallback.

**Principle**: Deploy always packages "build output". The build step is abstract
—Vite for capable machines, Go file-copy for minimal machines. Same output
structure, same deploy process.

## The Problem

Current app development has friction:

| Issue | Impact |
|-------|--------|
| No HMR during development | Slow iteration, manual refresh |
| No error overlay | Debugging requires console |
| Large single-file apps | Hard to maintain (tetris = 54KB HTML) |
| CDN dependencies | Latency, offline fragility |
| No tree-shaking | Larger bundles than necessary |

## The Solution

**Progressive enhancement at build-time, not runtime.**

```
Developer machine (has pkg mgr)     Minimal machine (no pkg mgr)
        │                                  │
        ▼                                  ▼
   npm/pnpm/bun build              Check for dist/ or pre-built branch
        │                                  │
        ▼                                  ▼
   dist/ folder                    Use existing build OR fail clearly
        │                                  │
        ▼                                  ▼
   Deploy dist/                    Deploy pre-built OR error
```

**Result**: Server receives same structure regardless of build method.

**Key constraint**: Complex apps that require building MUST have either:
- A package manager available, OR
- A pre-built dist/ folder, OR
- A pre-built branch (fazt-dist)

If none available → **fail with clear error** (not broken deployment).

---

## Part 1: Embedded Templates

### Template Storage

```
internal/assets/
├── system/
│   └── admin/           # existing embedded admin SPA
└── templates/
    ├── minimal/         # bare minimum app
    │   ├── manifest.json
    │   └── index.html
    └── vite/            # vite-ready app
        ├── manifest.json
        ├── index.html
        ├── src/
        │   └── main.js
        ├── api/
        │   └── hello.js
        ├── package.json
        └── vite.config.js
```

### Embedding with go:embed

```go
// internal/assets/templates.go
package assets

import "embed"

//go:embed templates/*
var Templates embed.FS

// GetTemplate returns template files for a given template name
func GetTemplate(name string) (fs.FS, error) {
    return fs.Sub(Templates, "templates/"+name)
}

// ListTemplates returns available template names
func ListTemplates() []string {
    entries, _ := fs.ReadDir(Templates, "templates")
    var names []string
    for _, e := range entries {
        if e.IsDir() {
            names = append(names, e.Name())
        }
    }
    return names
}
```

### Minimal Template

```
templates/minimal/
├── manifest.json    →  {"name": "{{.Name}}"}
└── index.html       →  <!DOCTYPE html>...basic HTML
```

### Vite Template

```
templates/vite/
├── manifest.json
├── index.html
├── src/
│   └── main.js
├── api/
│   └── hello.js
├── package.json
└── vite.config.js
```

#### manifest.json
```json
{
  "name": "{{.Name}}",
  "version": "1.0.0"
}
```

#### index.html
```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{.Name}}</title>
  <script type="importmap">
  {
    "imports": {
      "three": "https://cdn.jsdelivr.net/npm/three@0.160.0/build/three.module.js",
      "vue": "https://unpkg.com/vue@3/dist/vue.esm-browser.js"
    }
  }
  </script>
  <script src="https://cdn.tailwindcss.com"></script>
</head>
<body>
  <div id="app"></div>
  <script type="module" src="./src/main.js"></script>
</body>
</html>
```

#### src/main.js
```javascript
// Main entry point
// Works with AND without Vite

console.log('{{.Name}} loaded')

// Example: Vue app
// import { createApp, ref } from 'vue'
// createApp({ setup() { return { count: ref(0) } } }).mount('#app')
```

#### api/hello.js
```javascript
// Serverless function - runs on fazt server via Goja
// Access: GET /api/hello

function handler(req) {
  return {
    status: 200,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      message: "Hello from {{.Name}}",
      time: Date.now()
    })
  }
}
```

#### package.json
```json
{
  "name": "{{.Name}}",
  "version": "1.0.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview"
  },
  "devDependencies": {
    "vite": "^5.0.0",
    "rollup-plugin-copy": "^3.5.0"
  }
}
```

#### vite.config.js
```javascript
import { defineConfig } from 'vite'
import copy from 'rollup-plugin-copy'

export default defineConfig({
  base: './',
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    sourcemap: false,
    minify: 'esbuild',
  },
  plugins: [
    copy({
      targets: [
        { src: 'api/*', dest: 'dist/api' },
        { src: 'manifest.json', dest: 'dist' }
      ],
      hook: 'writeBundle'
    })
  ],
  server: {
    port: 7100,
    host: true,
  },
})
```

**Key insight**: The vite.config.js uses rollup-plugin-copy to include api/ and
manifest.json in the dist/ output. This makes dist/ a complete deployable unit.

---

## Part 2: CLI Command - `fazt app create`

### Command Syntax

```bash
fazt app create <name>                    # Uses 'minimal' template
fazt app create <name> --template vite    # Uses 'vite' template
fazt app create <name> --template <name>  # Uses named template
fazt app create --list-templates          # Show available templates
```

### Implementation

```go
// cmd/server/app.go - add to runApp switch

case "create":
    return runAppCreate(args[1:])
```

```go
// cmd/server/app_create.go
package main

import (
    "bytes"
    "fmt"
    "io/fs"
    "os"
    "path/filepath"
    "strings"
    "text/template"

    "fazt/internal/assets"
)

func runAppCreate(args []string) error {
    // Parse flags
    fs := flag.NewFlagSet("create", flag.ExitOnError)
    templateName := fs.String("template", "minimal", "Template to use")
    listTemplates := fs.Bool("list-templates", false, "List available templates")
    fs.Parse(args)

    // List templates mode
    if *listTemplates {
        templates := assets.ListTemplates()
        fmt.Println("Available templates:")
        for _, t := range templates {
            fmt.Printf("  - %s\n", t)
        }
        return nil
    }

    // Get app name
    if fs.NArg() < 1 {
        return fmt.Errorf("usage: fazt app create <name> [--template <template>]")
    }
    appName := fs.Arg(0)

    // Validate app name
    if !isValidAppName(appName) {
        return fmt.Errorf("invalid app name: %s (use lowercase letters, numbers, hyphens)", appName)
    }

    // Get template
    tmplFS, err := assets.GetTemplate(*templateName)
    if err != nil {
        return fmt.Errorf("template not found: %s", *templateName)
    }

    // Determine output directory
    outputDir := appName
    if _, err := os.Stat(outputDir); err == nil {
        return fmt.Errorf("directory already exists: %s", outputDir)
    }

    // Template data
    data := map[string]string{
        "Name": appName,
    }

    // Copy template files with substitution
    err = fs.WalkDir(tmplFS, ".", func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }

        destPath := filepath.Join(outputDir, path)

        if d.IsDir() {
            return os.MkdirAll(destPath, 0755)
        }

        content, err := fs.ReadFile(tmplFS, path)
        if err != nil {
            return err
        }

        // Apply template substitution
        tmpl, err := template.New(path).Parse(string(content))
        if err != nil {
            // Not a template, copy as-is
            return os.WriteFile(destPath, content, 0644)
        }

        var buf bytes.Buffer
        if err := tmpl.Execute(&buf, data); err != nil {
            return os.WriteFile(destPath, content, 0644)
        }

        return os.WriteFile(destPath, buf.Bytes(), 0644)
    })

    if err != nil {
        return fmt.Errorf("failed to create app: %w", err)
    }

    // Success message
    fmt.Printf("Created %s from '%s' template\n\n", appName, *templateName)

    if *templateName == "vite" {
        fmt.Println("Next steps:")
        fmt.Printf("  cd %s\n", appName)
        fmt.Println("  npm install        # Install dev dependencies")
        fmt.Println("  npm run dev        # Start dev server with HMR")
        fmt.Println("  npm run build      # Build for production")
        fmt.Println("")
        fmt.Println("Or deploy directly (works without npm):")
        fmt.Printf("  fazt app deploy %s --to zyt\n", appName)
    } else {
        fmt.Println("Next steps:")
        fmt.Printf("  fazt app deploy %s --to zyt\n", appName)
    }

    return nil
}

func isValidAppName(name string) bool {
    if len(name) == 0 || len(name) > 63 {
        return false
    }
    for _, c := range name {
        if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
            return false
        }
    }
    return name[0] != '-' && name[len(name)-1] != '-'
}
```

---

## Part 3: Unified Build Model

### The Build Abstraction

"Build" produces a deployable folder. Implementation varies:

| Scenario | Build Method | Output |
|----------|--------------|--------|
| Has pkg mgr + build script | `npm/pnpm/bun run build` | dist/ |
| Has build script, no pkg mgr, has dist/ | Use existing | dist/ |
| Has build script, no pkg mgr, no dist/ | **ERROR** | - |
| No build script | Go copies source | source as-is |

### Package Manager Detection

Support multiple package managers in priority order:

```go
// internal/build/pkgmgr.go
package build

import "os/exec"

type PackageManager struct {
    Name       string
    Binary     string
    InstallCmd []string
    BuildCmd   []string
    LockFile   string
}

var PackageManagers = []PackageManager{
    {
        Name:       "bun",
        Binary:     "bun",
        InstallCmd: []string{"install"},
        BuildCmd:   []string{"run", "build"},
        LockFile:   "bun.lockb",
    },
    {
        Name:       "pnpm",
        Binary:     "pnpm",
        InstallCmd: []string{"install"},
        BuildCmd:   []string{"run", "build"},
        LockFile:   "pnpm-lock.yaml",
    },
    {
        Name:       "yarn",
        Binary:     "yarn",
        InstallCmd: []string{"install"},
        BuildCmd:   []string{"run", "build"},
        LockFile:   "yarn.lock",
    },
    {
        Name:       "npm",
        Binary:     "npm",
        InstallCmd: []string{"install"},
        BuildCmd:   []string{"run", "build"},
        LockFile:   "package-lock.json",
    },
}

// DetectPackageManager finds available package manager
// Priority: lockfile match > first available
func DetectPackageManager(srcDir string) *PackageManager {
    // First, check for lockfiles (indicates project preference)
    for _, pm := range PackageManagers {
        lockPath := filepath.Join(srcDir, pm.LockFile)
        if _, err := os.Stat(lockPath); err == nil {
            // Lockfile exists, check if binary available
            if _, err := exec.LookPath(pm.Binary); err == nil {
                return &pm
            }
        }
    }

    // No lockfile match, return first available
    for _, pm := range PackageManagers {
        if _, err := exec.LookPath(pm.Binary); err == nil {
            return &pm
        }
    }

    return nil // No package manager available
}
```

### Build Detection Logic

```go
// internal/build/build.go
package build

import (
    "encoding/json"
    "errors"
    "os"
    "os/exec"
    "path/filepath"
)

var ErrBuildRequired = errors.New("app requires building but no package manager available and no pre-built dist/ exists")

type BuildResult struct {
    OutputDir string // Absolute path to deployable folder
    Method    string // "bun", "pnpm", "npm", "existing", "copy"
    PkgMgr    string // Which package manager was used (if any)
    Files     int    // Number of files
}

// Build prepares an app directory for deployment
func Build(srcDir string) (*BuildResult, error) {
    // Check for package.json with build script
    pkgPath := filepath.Join(srcDir, "package.json")
    if hasBuildScript(pkgPath) {
        return buildWithPackageManager(srcDir)
    }

    // No build script - use source directly
    return copySource(srcDir)
}

func hasBuildScript(pkgPath string) bool {
    data, err := os.ReadFile(pkgPath)
    if err != nil {
        return false
    }

    var pkg struct {
        Scripts map[string]string `json:"scripts"`
    }
    if err := json.Unmarshal(data, &pkg); err != nil {
        return false
    }

    _, has := pkg.Scripts["build"]
    return has
}

func buildWithPackageManager(srcDir string) (*BuildResult, error) {
    pm := DetectPackageManager(srcDir)
    if pm == nil {
        // No package manager - check for existing dist/
        return useExistingBuild(srcDir)
    }

    fmt.Printf("Using %s to build...\n", pm.Name)

    // Run install if node_modules missing
    nodeModules := filepath.Join(srcDir, "node_modules")
    if _, err := os.Stat(nodeModules); os.IsNotExist(err) {
        cmd := exec.Command(pm.Binary, pm.InstallCmd...)
        cmd.Dir = srcDir
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        if err := cmd.Run(); err != nil {
            return nil, fmt.Errorf("%s install failed: %w", pm.Name, err)
        }
    }

    // Run build
    cmd := exec.Command(pm.Binary, pm.BuildCmd...)
    cmd.Dir = srcDir
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
        return nil, fmt.Errorf("%s run build failed: %w", pm.Name, err)
    }

    // Find output directory (dist/ or build/)
    outputDir := findBuildOutput(srcDir)
    if outputDir == "" {
        return nil, fmt.Errorf("build succeeded but no output directory found")
    }

    return &BuildResult{
        OutputDir: outputDir,
        Method:    pm.Name,
        PkgMgr:    pm.Name,
        Files:     countFiles(outputDir),
    }, nil
}

func useExistingBuild(srcDir string) (*BuildResult, error) {
    outputDir := findBuildOutput(srcDir)
    if outputDir == "" {
        // No existing build - this is an error, not a fallback
        return nil, ErrBuildRequired
    }

    fmt.Println("Using existing build output...")
    return &BuildResult{
        OutputDir: outputDir,
        Method:    "existing",
        Files:     countFiles(outputDir),
    }, nil
}

func copySource(srcDir string) (*BuildResult, error) {
    // For simple apps, the source IS the deployable output
    // No actual copy needed - just return the source dir
    return &BuildResult{
        OutputDir: srcDir,
        Method:    "copy",
        Files:     countFiles(srcDir),
    }, nil
}

func findBuildOutput(srcDir string) string {
    // Check common build output directories in priority order
    candidates := []string{"dist", "build", "out"}
    for _, dir := range candidates {
        path := filepath.Join(srcDir, dir)
        indexPath := filepath.Join(path, "index.html")
        if _, err := os.Stat(indexPath); err == nil {
            return path
        }
    }
    return ""
}

func countFiles(dir string) int {
    count := 0
    filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
        if err == nil && !d.IsDir() {
            count++
        }
        return nil
    })
    return count
}
```

### Updated Deploy Command

```go
// cmd/server/app_deploy.go (modified)

func runAppDeploy(args []string) error {
    fs := flag.NewFlagSet("deploy", flag.ExitOnError)
    peer := fs.String("to", "", "Target peer")
    noBuild := fs.Bool("no-build", false, "Skip build step")
    fs.Parse(args)

    if fs.NArg() < 1 {
        return fmt.Errorf("usage: fazt app deploy <dir> --to <peer>")
    }

    srcDir := fs.Arg(0)

    // Build step (unless skipped)
    var deployDir string
    if *noBuild {
        deployDir = srcDir
        fmt.Println("Skipping build (--no-build)")
    } else {
        result, err := build.Build(srcDir)
        if err != nil {
            return fmt.Errorf("build failed: %w", err)
        }
        deployDir = result.OutputDir
        fmt.Printf("Build: %s (%d files via %s)\n",
            deployDir, result.Files, result.Method)
    }

    // Read manifest
    manifestPath := filepath.Join(deployDir, "manifest.json")
    manifest, err := readManifest(manifestPath)
    if err != nil {
        // Try source dir manifest (for builds that didn't copy it)
        manifestPath = filepath.Join(srcDir, "manifest.json")
        manifest, err = readManifest(manifestPath)
        if err != nil {
            return fmt.Errorf("no manifest.json found")
        }
    }

    // Package and deploy
    if *peer == "" {
        return deployLocal(deployDir, manifest.Name, nil)
    }
    return deployRemote(deployDir, manifest.Name, *peer, nil)
}
```

---

## Part 4: Git Install Enhancements

### Subfolder Support

Extend URL parser from Plan 18:

```go
// internal/git/parser.go (extended)

// ParseURL handles:
// github.com/user/repo
// github.com/user/repo/path/to/app
// github.com/user/repo@branch
// github.com/user/repo/path/to/app@v1.0.0
// github:user/repo/app (shorthand)

func ParseURL(input string) (*RepoRef, error) {
    // ... existing parsing logic ...

    // Extract path within repo (after repo name, before @)
    // e.g., github.com/user/repo/apps/calculator@main
    //       → Path: "apps/calculator", Ref: "main"
}
```

### Pre-built Branch Detection

For complex apps without npm on the installing machine:

```go
// internal/git/branches.go

// PrebuiltBranches to check (in order)
var PrebuiltBranches = []string{
    "fazt-dist",  // Recommended convention
    "dist",
    "release",
    "gh-pages",
}

// FindPrebuiltBranch checks if repo has a pre-built branch
func FindPrebuiltBranch(repoURL string) (string, error) {
    for _, branch := range PrebuiltBranches {
        if BranchExists(repoURL, branch) {
            return branch, nil
        }
    }
    return "", nil // No pre-built branch found
}
```

### Updated Install Command

```go
// cmd/server/app_install.go (extended)

func runAppInstall(args []string) error {
    // ... parse URL and flags ...

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

    // Build step
    buildResult, err := build.Build(tmpDir)
    if err != nil {
        // Build failed - check for pre-built branch
        if ref.Ref == "" { // Only if user didn't specify a ref
            prebuilt, _ := git.FindPrebuiltBranch(ref.FullURL())
            if prebuilt != "" {
                fmt.Printf("Build failed. Trying pre-built branch '%s'...\n", prebuilt)
                ref.Ref = prebuilt
                // Re-clone from pre-built branch
                result, err = git.Clone(git.CloneOptions{
                    URL:       ref.FullURL(),
                    Path:      ref.Path,
                    Ref:       prebuilt,
                    TargetDir: tmpDir,
                })
                if err != nil {
                    return fmt.Errorf("clone of pre-built branch failed: %w", err)
                }
                // Pre-built branch is already "built" - use as-is
                buildResult = &build.BuildResult{
                    OutputDir: tmpDir,
                    Method:    "prebuilt",
                }
            }
        }
        if buildResult == nil {
            return fmt.Errorf("build failed and no pre-built branch found: %w", err)
        }
    }

    // Deploy
    // ... rest of deploy logic ...
}
```

---

## Part 5: API & GUI Install Support

All install/deploy operations work via API, enabling:
- Admin dashboard (GUI)
- LLM harness (internal AI)
- MCP tools
- External automation

### API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `POST /api/apps/install` | POST | Install from GitHub URL |
| `POST /api/apps/deploy` | POST | Deploy from uploaded ZIP |
| `POST /api/apps/create` | POST | Create from template |
| `GET /api/apps` | GET | List apps |
| `DELETE /api/apps/{name}` | DELETE | Remove app |

### Install Endpoint

```go
// internal/handlers/apps.go

type InstallRequest struct {
    URL  string `json:"url"`  // GitHub URL
    Name string `json:"name"` // Optional name override
}

func (h *Handler) InstallApp(w http.ResponseWriter, r *http.Request) {
    var req InstallRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        api.BadRequest(w, "invalid request")
        return
    }

    // Parse URL
    ref, err := git.ParseURL(req.URL)
    if err != nil {
        api.BadRequest(w, "invalid GitHub URL")
        return
    }

    // Clone to temp
    tmpDir, err := os.MkdirTemp("", "fazt-install-*")
    if err != nil {
        api.InternalError(w, err)
        return
    }
    defer os.RemoveAll(tmpDir)

    result, err := git.Clone(git.CloneOptions{
        URL:       ref.FullURL(),
        Path:      ref.Path,
        Ref:       ref.Ref,
        TargetDir: tmpDir,
    })
    if err != nil {
        api.BadRequest(w, fmt.Sprintf("clone failed: %v", err))
        return
    }

    // Build (server likely has no npm - will use existing or copy)
    buildResult, err := build.Build(tmpDir)
    if err != nil {
        // Try pre-built branch
        prebuilt, _ := git.FindPrebuiltBranch(ref.FullURL())
        if prebuilt != "" {
            ref.Ref = prebuilt
            result, err = git.Clone(git.CloneOptions{
                URL:       ref.FullURL(),
                Path:      ref.Path,
                Ref:       prebuilt,
                TargetDir: tmpDir,
            })
            if err != nil {
                api.BadRequest(w, "app requires building; no pre-built branch found")
                return
            }
            buildResult = &build.BuildResult{OutputDir: tmpDir, Method: "prebuilt"}
        } else {
            api.BadRequest(w, "app requires building; install from a machine with npm")
            return
        }
    }

    // Deploy locally
    manifest, _ := readManifest(filepath.Join(buildResult.OutputDir, "manifest.json"))
    appName := manifest.Name
    if req.Name != "" {
        appName = req.Name
    }

    if err := h.hosting.Deploy(appName, buildResult.OutputDir, &hosting.SourceInfo{
        Type:   "git",
        URL:    req.URL,
        Ref:    ref.Ref,
        Commit: result.CommitSHA,
    }); err != nil {
        api.InternalError(w, err)
        return
    }

    api.Success(w, http.StatusCreated, map[string]string{
        "name":   appName,
        "url":    fmt.Sprintf("https://%s.%s", appName, h.config.Domain),
        "source": req.URL,
    })
}
```

### Create Endpoint (for LLM harness)

```go
type CreateRequest struct {
    Name     string `json:"name"`
    Template string `json:"template"` // "minimal" or "vite"
}

func (h *Handler) CreateApp(w http.ResponseWriter, r *http.Request) {
    var req CreateRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        api.BadRequest(w, "invalid request")
        return
    }

    if req.Template == "" {
        req.Template = "minimal"
    }

    // Get template
    tmplFS, err := assets.GetTemplate(req.Template)
    if err != nil {
        api.BadRequest(w, "unknown template: "+req.Template)
        return
    }

    // Create in temp directory
    outputDir := filepath.Join(os.TempDir(), "fazt-create-"+req.Name)
    os.RemoveAll(outputDir) // Clean if exists

    // Copy template with substitution
    if err := copyTemplate(tmplFS, outputDir, req.Name); err != nil {
        api.InternalError(w, err)
        return
    }

    api.Success(w, http.StatusCreated, map[string]string{
        "name":     req.Name,
        "template": req.Template,
        "path":     outputDir,
    })
}
```

### Admin Dashboard UI

Add to Apps page:

```jsx
// admin/src/pages/Apps.jsx (addition)

function InstallFromGitHub() {
  const [url, setUrl] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  const install = async () => {
    setLoading(true)
    setError(null)
    try {
      const res = await fetch('/api/apps/install', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url })
      })
      if (!res.ok) {
        const data = await res.json()
        throw new Error(data.message || 'Install failed')
      }
      const data = await res.json()
      // Refresh app list, show success
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="border rounded p-4">
      <h3>Install from GitHub</h3>
      <input
        type="text"
        placeholder="github.com/user/repo or github.com/user/repo/app"
        value={url}
        onChange={e => setUrl(e.target.value)}
        className="w-full border rounded px-3 py-2"
      />
      {error && <p className="text-red-500 text-sm mt-2">{error}</p>}
      <button
        onClick={install}
        disabled={loading || !url}
        className="mt-2 px-4 py-2 bg-blue-500 text-white rounded"
      >
        {loading ? 'Installing...' : 'Install'}
      </button>
    </div>
  )
}
```

---

## Part 6: Documentation

### GitHub Actions Template

Create `koder/docs/github-actions-fazt.md`:

```markdown
# Building Apps for Fazt with GitHub Actions

For apps that require a build step (Vite, React, etc.), set up GitHub Actions
to automatically build and publish a `fazt-dist` branch.

## Setup

Create `.github/workflows/fazt.yml`:

\`\`\`yaml
name: Build for Fazt

on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Node
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Build
        run: npm run build

      - name: Deploy to fazt-dist branch
        uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: \${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./dist
          publish_branch: fazt-dist
\`\`\`

## How It Works

1. Push code to `main` branch
2. GitHub Actions builds automatically (~1-2 minutes)
3. Built output pushed to `fazt-dist` branch
4. Users install via: `fazt app install github:you/repo`
5. Fazt detects no npm, finds `fazt-dist` branch, uses that

## Benefits

- Source stays clean (no dist/ in main)
- Apps installable on any fazt instance (no npm required)
- GUI install works (admin dashboard)
- Automatic rebuilds on every push
```

### Update CLAUDE.md

Add to App Development section:

```markdown
### App Templates

Create apps from templates:

\`\`\`bash
fazt app create myapp                    # Minimal template
fazt app create myapp --template vite    # Vite-ready with HMR
fazt app create --list-templates         # Show available templates
\`\`\`

### Vite Development Workflow

For apps created with `--template vite`:

\`\`\`bash
cd myapp
npm install          # One-time setup
npm run dev          # Dev server with HMR at http://192.168.64.3:7100

# When ready to deploy:
npm run build        # Optional - creates optimized dist/
fazt app deploy . --to zyt
\`\`\`

**No npm? No problem.** Deploy works without building:

\`\`\`bash
fazt app deploy myapp --to zyt   # Deploys source directly
\`\`\`

The app will work (using CDN imports) but won't be optimized.

### Build Behavior

| Has npm | Has build script | Result |
|---------|-----------------|--------|
| Yes | Yes | Runs npm build, deploys dist/ |
| Yes | No | Deploys source |
| No | Yes + dist/ exists | Deploys existing dist/ |
| No | Yes + no dist/ | Deploys source (with warning) |
| No | No | Deploys source |
```

---

## Part 7: Tests

### Template Tests

```go
// internal/assets/templates_test.go

func TestListTemplates(t *testing.T) {
    templates := ListTemplates()
    assert.Contains(t, templates, "minimal")
    assert.Contains(t, templates, "vite")
}

func TestGetTemplate(t *testing.T) {
    t.Run("minimal template has required files", func(t *testing.T) {
        fs, err := GetTemplate("minimal")
        assert.NoError(t, err)

        // Check manifest.json exists
        _, err = fs.Open("manifest.json")
        assert.NoError(t, err)

        // Check index.html exists
        _, err = fs.Open("index.html")
        assert.NoError(t, err)
    })

    t.Run("vite template has required files", func(t *testing.T) {
        fs, err := GetTemplate("vite")
        assert.NoError(t, err)

        required := []string{
            "manifest.json",
            "index.html",
            "package.json",
            "vite.config.js",
            "src/main.js",
            "api/hello.js",
        }
        for _, f := range required {
            _, err = fs.Open(f)
            assert.NoError(t, err, "missing: %s", f)
        }
    })
}
```

### Build Tests

```go
// internal/build/build_test.go

func TestBuild(t *testing.T) {
    t.Run("simple app returns source dir", func(t *testing.T) {
        // Create temp dir with just index.html
        tmpDir := t.TempDir()
        os.WriteFile(filepath.Join(tmpDir, "index.html"), []byte("<h1>Hi</h1>"), 0644)
        os.WriteFile(filepath.Join(tmpDir, "manifest.json"), []byte(`{"name":"test"}`), 0644)

        result, err := Build(tmpDir)
        assert.NoError(t, err)
        assert.Equal(t, tmpDir, result.OutputDir)
        assert.Equal(t, "copy", result.Method)
    })

    t.Run("app with existing dist uses dist", func(t *testing.T) {
        tmpDir := t.TempDir()

        // Create package.json with build script
        pkg := `{"scripts":{"build":"echo build"}}`
        os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkg), 0644)

        // Create dist/ with index.html
        distDir := filepath.Join(tmpDir, "dist")
        os.MkdirAll(distDir, 0755)
        os.WriteFile(filepath.Join(distDir, "index.html"), []byte("<h1>Built</h1>"), 0644)

        // Mock: npm not available
        origPath := os.Getenv("PATH")
        os.Setenv("PATH", "")
        defer os.Setenv("PATH", origPath)

        result, err := Build(tmpDir)
        assert.NoError(t, err)
        assert.Equal(t, distDir, result.OutputDir)
        assert.Equal(t, "existing", result.Method)
    })
}
```

### CLI Tests

```go
// cmd/server/app_create_test.go

func TestAppCreate(t *testing.T) {
    t.Run("creates minimal app", func(t *testing.T) {
        tmpDir := t.TempDir()
        os.Chdir(tmpDir)

        err := runAppCreate([]string{"test-app"})
        assert.NoError(t, err)

        // Check files created
        assert.FileExists(t, "test-app/manifest.json")
        assert.FileExists(t, "test-app/index.html")

        // Check manifest has correct name
        data, _ := os.ReadFile("test-app/manifest.json")
        assert.Contains(t, string(data), `"name": "test-app"`)
    })

    t.Run("creates vite app", func(t *testing.T) {
        tmpDir := t.TempDir()
        os.Chdir(tmpDir)

        err := runAppCreate([]string{"--template", "vite", "vite-app"})
        assert.NoError(t, err)

        // Check vite-specific files
        assert.FileExists(t, "vite-app/package.json")
        assert.FileExists(t, "vite-app/vite.config.js")
        assert.FileExists(t, "vite-app/src/main.js")
        assert.FileExists(t, "vite-app/api/hello.js")
    })

    t.Run("rejects existing directory", func(t *testing.T) {
        tmpDir := t.TempDir()
        os.Chdir(tmpDir)
        os.MkdirAll("existing", 0755)

        err := runAppCreate([]string{"existing"})
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "already exists")
    })
}
```

---

## Part 8: Implementation Order

### Phase 1: Template Infrastructure

1. Create `internal/assets/templates/` directory structure
2. Create minimal template files
3. Create vite template files
4. Add `templates.go` with go:embed
5. Add tests for template loading

### Phase 2: CLI Create Command

1. Add `create` subcommand to `cmd/server/app.go`
2. Implement `runAppCreate` with template substitution
3. Add `--template` and `--list-templates` flags
4. Add tests

### Phase 3: Build Package

1. Create `internal/build/` package
2. Implement package manager detection (bun, pnpm, yarn, npm)
3. Implement `Build()` with package manager support
4. Implement fallback to existing dist/
5. Implement error for build-required apps without capability
6. Add tests

### Phase 4: Update Deploy

1. Modify `runAppDeploy` to use build package
2. Add `--no-build` flag
3. Update output messages
4. Add tests

### Phase 5: Git Enhancements

1. Extend URL parser for subfolders
2. Add pre-built branch detection
3. Update `runAppInstall` with build step
4. Update `runAppInstall` with pre-built fallback
5. Add tests

### Phase 6: GUI Install

1. Add `/api/apps/install` endpoint
2. Add handler with build + pre-built fallback
3. Update admin dashboard with install UI
4. Test end-to-end

### Phase 7: Documentation

1. Create GitHub Actions guide
2. Update CLAUDE.md with template docs
3. Update CLAUDE.md with build behavior

---

## Files to Create/Modify

| File | Action | Phase |
|------|--------|-------|
| `internal/assets/templates/minimal/*` | Create | 1 |
| `internal/assets/templates/vite/*` | Create | 1 |
| `internal/assets/templates.go` | Create | 1 |
| `internal/assets/templates_test.go` | Create | 1 |
| `cmd/server/app.go` | Modify (add create) | 2 |
| `cmd/server/app_create.go` | Create | 2 |
| `cmd/server/app_create_test.go` | Create | 2 |
| `internal/build/build.go` | Create | 3 |
| `internal/build/pkgmgr.go` | Create | 3 |
| `internal/build/build_test.go` | Create | 3 |
| `cmd/server/app_deploy.go` | Modify (use build) | 4 |
| `internal/git/parser.go` | Modify (subfolders) | 5 |
| `internal/git/branches.go` | Create | 5 |
| `internal/handlers/apps.go` | Modify (install endpoint) | 6 |
| `admin/src/pages/Apps.jsx` | Modify (install UI) | 6 |
| `koder/docs/github-actions-fazt.md` | Create | 7 |
| `CLAUDE.md` | Update | 7 |

---

## Version

This would be **v0.9.24** - enhancement to existing app ecosystem.

### Changelog Entry

```markdown
## v0.9.24 - Vite Dev Enhancement

### New Features
- `fazt app create` - Scaffold apps from templates
- `fazt app create --template vite` - Vite-ready app with HMR support
- Unified build model - automatic npm build or graceful fallback
- Pre-built branch detection for git installs
- GUI install from GitHub URLs

### Templates
- `minimal` - Basic HTML app
- `vite` - Vite project with HMR, Tailwind, importmaps

### Build Behavior
- Apps with package.json + build script: runs build via npm/pnpm/bun/yarn
- Falls back to existing dist/ if no package manager available
- **Fails with clear error** if build required but impossible
- Git install checks for fazt-dist branch when build unavailable

### Package Manager Support
- Detects: bun, pnpm, yarn, npm (in priority order)
- Respects lockfiles (uses matching package manager)

### Documentation
- GitHub Actions guide for automated builds
- Updated CLAUDE.md with template and build docs
```

---

## Part 9: LLM Harness Integration

The internal LLM harness (and MCP tools) can use this system for app deployment.

### Via CLI (Subprocess)

```go
// LLM harness calls fazt CLI
exec.Command("fazt", "app", "create", "my-app", "--template", "vite")
exec.Command("fazt", "app", "deploy", "my-app", "--to", "zyt")
```

### Via API (HTTP)

```go
// Create app from template
POST /api/apps/create
{
  "name": "my-app",
  "template": "vite"
}

// Deploy (for apps already on server filesystem)
POST /api/apps/deploy
{
  "name": "my-app",
  "source_dir": "/path/to/my-app"
}

// Install from GitHub
POST /api/apps/install
{
  "url": "github.com/user/repo/app",
  "name": "my-app"  // optional override
}
```

### Workflow Example

LLM harness building an app:

```
1. LLM generates app files to temp directory
2. LLM calls: POST /api/apps/deploy { source_dir: "/tmp/my-app" }
3. fazt builds (if needed) and deploys
4. LLM receives: { url: "https://my-app.zyt.app" }
```

Or via CLI:

```
1. LLM generates files to /tmp/my-app/
2. LLM calls: fazt app deploy /tmp/my-app --to zyt
3. fazt builds (if needed) and deploys
4. LLM parses stdout for URL
```

### MCP Tool Extensions

The existing MCP server can expose these as tools:

```json
{
  "name": "app_create",
  "description": "Create new app from template",
  "parameters": {
    "name": { "type": "string" },
    "template": { "type": "string", "enum": ["minimal", "vite"] }
  }
}

{
  "name": "app_deploy",
  "description": "Deploy app to fazt instance",
  "parameters": {
    "source_dir": { "type": "string" },
    "target": { "type": "string" }
  }
}
```

---

## Future Considerations (Not in This Plan)

| Feature | Trigger to Implement |
|---------|---------------------|
| React template | User demand for React-specific setup |
| Vue template | User demand for Vue-specific setup |
| Svelte template | User demand for Svelte-specific setup |
| `fazt app eject` | Convert minimal to vite mid-project |
| Custom template repos | Power users want their own templates |
| Build caching | Repeated deploys of same source |

---

## Design Decisions Captured

1. **Why embedded templates?** Single binary philosophy. No network fetch,
   works offline, version-locked to fazt version.

2. **Why rollup-plugin-copy?** Vite doesn't copy arbitrary folders by default.
   This is the standard solution. (Yes, it's ridiculous.)

3. **Why check multiple branch names?** Different communities use different
   conventions (dist, release, gh-pages). Being flexible helps adoption.

4. **Why fail instead of deploying broken source?** Deploying non-functional
   code is worse than a clear error. User knows exactly what to fix.

5. **Why importmaps in template?** Makes source files browser-runnable without
   build. This is the key to graceful degradation for SIMPLE apps.

6. **Why support multiple package managers?** Developers have preferences.
   bun is fastest, pnpm is efficient, yarn has workspaces. Respect the lockfile.
