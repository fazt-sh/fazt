# Test-First Migration Strategy: @peer-Primary CLI

**Version**: 0.18.0
**Status**: Design Document
**Date**: 2026-02-02

---

## Executive Summary

This document outlines a comprehensive test-first migration strategy for transitioning Fazt's CLI to the @peer-primary pattern. The key innovation is using `knowledge-base/cli/` as a **single source of truth** for both:

1. **Human documentation** (rendered as web docs)
2. **Binary --help output** (embedded at build time)

This ensures documentation never drifts from implementation.

---

## Part 1: Documentation Structure Design

### Directory Structure

```
knowledge-base/cli/
├── _meta.yaml                    # CLI metadata, version, global config
├── fazt.md                       # Top-level overview (fazt --help)
├── app/
│   ├── _index.md                 # App command overview (fazt app --help)
│   ├── list.md                   # fazt app list
│   ├── deploy.md                 # fazt app deploy
│   ├── info.md                   # fazt app info
│   ├── remove.md                 # fazt app remove
│   ├── create.md                 # fazt app create
│   ├── validate.md               # fazt app validate
│   ├── logs.md                   # fazt app logs
│   ├── install.md                # fazt app install
│   ├── upgrade.md                # fazt app upgrade
│   ├── pull.md                   # fazt app pull
│   ├── link.md                   # fazt app link
│   ├── unlink.md                 # fazt app unlink
│   ├── reserve.md                # fazt app reserve
│   ├── fork.md                   # fazt app fork
│   ├── swap.md                   # fazt app swap
│   ├── split.md                  # fazt app split
│   └── lineage.md                # fazt app lineage
├── remote/
│   ├── _index.md                 # Remote command overview
│   ├── add.md                    # fazt remote add
│   ├── list.md                   # fazt remote list
│   ├── remove.md                 # fazt remote remove
│   ├── default.md                # fazt remote default
│   └── status.md                 # fazt remote status
├── server/
│   ├── _index.md                 # Server command overview
│   ├── init.md                   # fazt server init
│   ├── start.md                  # fazt server start
│   ├── status.md                 # fazt server status
│   ├── set-config.md             # fazt server set-config
│   ├── set-credentials.md        # fazt server set-credentials
│   └── create-key.md             # fazt server create-key
├── service/
│   ├── _index.md                 # Service command overview
│   ├── install.md                # fazt service install
│   ├── start.md                  # fazt service start
│   ├── stop.md                   # fazt service stop
│   ├── status.md                 # fazt service status
│   └── logs.md                   # fazt service logs
└── topics/
    ├── peer-syntax.md            # @peer pattern explanation
    ├── local-first.md            # Local-first philosophy
    └── migration.md              # Migration from flags to @peer
```

### Frontmatter Schema

Each command document uses this YAML frontmatter:

```yaml
---
# Required
command: "app deploy"           # Full command path
version: "0.18.0"               # When this was last updated
category: "deployment"          # Category: query, modify, deploy, transfer, local-only, management

# Command Definition
syntax: "fazt [@peer] app deploy <directory> [flags]"
description: "Deploy a directory to a fazt instance"

# Arguments
arguments:
  - name: "directory"
    type: "path"
    required: true
    description: "Path to the directory to deploy"

# Flags
flags:
  - name: "--name"
    short: "-n"
    type: "string"
    default: "directory name"
    description: "App name (defaults to directory name)"
  - name: "--spa"
    type: "bool"
    default: false
    description: "Enable SPA routing (clean URLs)"
  - name: "--no-build"
    type: "bool"
    default: false
    description: "Skip build step"
  - name: "--include-private"
    type: "bool"
    default: false
    description: "Include gitignored private/ directory"

# Peer Support
peer:
  supported: true               # Can use @peer prefix
  local: true                   # Works locally
  remote: true                  # Works remotely
  deprecated_flags:             # Old flags being retired
    - "--to"

# Examples (machine-parseable)
examples:
  - title: "Deploy to local"
    command: "fazt app deploy ./my-app"
    description: "Deploy directory to local fazt instance"
  - title: "Deploy to remote peer"
    command: "fazt @zyt app deploy ./my-app"
    description: "Deploy directory to remote peer 'zyt'"
  - title: "Deploy with custom name"
    command: "fazt @zyt app deploy ./my-app --name blog"
    description: "Deploy with explicit app name"
  - title: "Deploy SPA"
    command: "fazt @zyt app deploy ./my-spa --spa"
    description: "Deploy with SPA routing enabled"

# Related Commands
related:
  - "app list"
  - "app remove"
  - "app validate"

# Errors
errors:
  - code: "ENOENT"
    message: "Directory does not exist"
    solution: "Verify the path exists and is accessible"
  - code: "ENOPEER"
    message: "No local fazt server running"
    solution: "Start local server with 'fazt server start' or target remote peer with @peer"
---
```

### Markdown Body Convention

After frontmatter, the markdown body follows this structure:

```markdown
# fazt app deploy

Deploy a local directory to a fazt instance.

## Synopsis

```
fazt app deploy <directory> [--name <name>] [--spa] [--no-build]
fazt @<peer> app deploy <directory> [--name <name>] [--spa] [--no-build]
```

## Description

The `deploy` command uploads a directory to a fazt instance. By default,
it deploys to the local fazt server. Use the `@peer` prefix to deploy
to a remote peer.

The command automatically detects and runs build tools (npm, pnpm, yarn, bun)
if a `package.json` is present. Use `--no-build` to skip this step.

## Arguments

- `<directory>` - Path to deploy (required)

## Flags

- `--name, -n <name>` - Override app name (default: directory name)
- `--spa` - Enable SPA routing for clean URLs
- `--no-build` - Skip automatic build step
- `--include-private` - Include gitignored `private/` directory

## Examples

### Deploy to local fazt

```bash
fazt app deploy ./my-app
```

### Deploy to remote peer

```bash
fazt @zyt app deploy ./my-app
```

### Deploy with SPA routing

```bash
fazt @zyt app deploy ./my-spa --spa
```

## Common Errors

### No local server running

```
Error: No local fazt server running.

To deploy to a remote peer:
  fazt @<peer> app deploy ./app

To start local server:
  fazt server start
```

## See Also

- `fazt app list` - List deployed apps
- `fazt app remove` - Remove an app
- `fazt app validate` - Validate before deploy
```

### Conventions

1. **File naming**: kebab-case matching command name
2. **Frontmatter required**: All command docs must have complete frontmatter
3. **Examples are testable**: Every example should be runnable
4. **Errors are actionable**: Include solution for each error
5. **Related commands linked**: Help users discover related functionality

---

## Part 2: Binary Help System Design

### Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Build Time                                │
├─────────────────────────────────────────────────────────────────┤
│  knowledge-base/cli/**/*.md                                      │
│           │                                                      │
│           ▼                                                      │
│  go:embed → internal/help/docs.go                               │
│           │                                                      │
│           ▼                                                      │
│  fazt binary (docs embedded)                                    │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                        Runtime                                   │
├─────────────────────────────────────────────────────────────────┤
│  fazt app deploy --help                                         │
│           │                                                      │
│           ▼                                                      │
│  Load embedded app/deploy.md                                    │
│           │                                                      │
│           ▼                                                      │
│  Parse YAML frontmatter                                         │
│           │                                                      │
│           ▼                                                      │
│  Render markdown → terminal (glamour)                           │
│           │                                                      │
│           ▼                                                      │
│  Display with colors/formatting                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Implementation Approach

#### 1. Embedding Docs (internal/help/embed.go)

```go
package help

import (
    "embed"
)

//go:embed docs/*
var DocsFS embed.FS
```

Build copies `knowledge-base/cli/` to `internal/help/docs/` before compilation.

#### 2. Help Parser (internal/help/parser.go)

```go
package help

import (
    "bytes"
    "gopkg.in/yaml.v3"
)

type CommandDoc struct {
    // Frontmatter
    Command     string     `yaml:"command"`
    Version     string     `yaml:"version"`
    Category    string     `yaml:"category"`
    Syntax      string     `yaml:"syntax"`
    Description string     `yaml:"description"`
    Arguments   []Argument `yaml:"arguments"`
    Flags       []Flag     `yaml:"flags"`
    Peer        PeerConfig `yaml:"peer"`
    Examples    []Example  `yaml:"examples"`
    Related     []string   `yaml:"related"`
    Errors      []Error    `yaml:"errors"`

    // Markdown body
    Body string
}

type Argument struct {
    Name        string `yaml:"name"`
    Type        string `yaml:"type"`
    Required    bool   `yaml:"required"`
    Description string `yaml:"description"`
}

type Flag struct {
    Name        string `yaml:"name"`
    Short       string `yaml:"short"`
    Type        string `yaml:"type"`
    Default     string `yaml:"default"`
    Description string `yaml:"description"`
}

type PeerConfig struct {
    Supported       bool     `yaml:"supported"`
    Local           bool     `yaml:"local"`
    Remote          bool     `yaml:"remote"`
    DeprecatedFlags []string `yaml:"deprecated_flags"`
}

type Example struct {
    Title       string `yaml:"title"`
    Command     string `yaml:"command"`
    Description string `yaml:"description"`
}

type Error struct {
    Code     string `yaml:"code"`
    Message  string `yaml:"message"`
    Solution string `yaml:"solution"`
}

func ParseDoc(content []byte) (*CommandDoc, error) {
    // Split frontmatter and body
    parts := bytes.SplitN(content, []byte("---"), 3)
    if len(parts) < 3 {
        return nil, fmt.Errorf("invalid doc format")
    }

    var doc CommandDoc
    if err := yaml.Unmarshal(parts[1], &doc); err != nil {
        return nil, err
    }

    doc.Body = string(parts[2])
    return &doc, nil
}
```

#### 3. Terminal Renderer (internal/help/render.go)

```go
package help

import (
    "github.com/charmbracelet/glamour"
    "github.com/fatih/color"
)

type Renderer struct {
    termWidth int
    colorized bool
}

func NewRenderer() *Renderer {
    // Detect terminal width
    width := 80
    if w, _, err := term.GetSize(0); err == nil {
        width = w
    }

    return &Renderer{
        termWidth: width,
        colorized: term.IsTerminal(0),
    }
}

func (r *Renderer) RenderDoc(doc *CommandDoc) string {
    var out strings.Builder

    // Render synopsis
    out.WriteString(r.renderSynopsis(doc))

    // Render description
    out.WriteString(r.renderDescription(doc))

    // Render arguments
    if len(doc.Arguments) > 0 {
        out.WriteString(r.renderArguments(doc))
    }

    // Render flags
    if len(doc.Flags) > 0 {
        out.WriteString(r.renderFlags(doc))
    }

    // Render examples
    if len(doc.Examples) > 0 {
        out.WriteString(r.renderExamples(doc))
    }

    return out.String()
}

func (r *Renderer) renderSynopsis(doc *CommandDoc) string {
    bold := color.New(color.Bold)
    return bold.Sprintf("USAGE:\n  %s\n\n", doc.Syntax)
}

// ... additional render methods
```

#### 4. Help Command Handler (cmd/server/help.go)

```go
func handleHelp(args []string) {
    loader := help.NewLoader(help.DocsFS)
    renderer := help.NewRenderer()

    // Determine which doc to load
    docPath := "fazt.md"
    if len(args) > 0 {
        docPath = strings.Join(args, "/") + ".md"
    }

    doc, err := loader.Load(docPath)
    if err != nil {
        // Fall back to _index.md if exists
        doc, err = loader.Load(strings.Join(args, "/") + "/_index.md")
        if err != nil {
            fmt.Printf("No help available for: %s\n", strings.Join(args, " "))
            os.Exit(1)
        }
    }

    output := renderer.RenderDoc(doc)

    // Use pager for long output
    if len(output) > renderer.termWidth * 24 {
        pager.Display(output)
    } else {
        fmt.Print(output)
    }
}
```

### Commands to Support

```bash
fazt --help                     # From fazt.md
fazt app --help                 # From app/_index.md
fazt app deploy --help          # From app/deploy.md
fazt help @peer                 # From topics/peer-syntax.md
fazt help migration             # From topics/migration.md
```

### Technical Stack

| Component | Library | Purpose |
|-----------|---------|---------|
| Embedding | `go:embed` | Bundle docs into binary |
| YAML parsing | `gopkg.in/yaml.v3` | Parse frontmatter |
| Markdown parsing | `github.com/yuin/goldmark` | Parse markdown body |
| Terminal rendering | `github.com/charmbracelet/glamour` | Pretty markdown output |
| Colors | `github.com/fatih/color` | Terminal colors |
| Terminal detection | `golang.org/x/term` | Detect width, TTY |
| Paging | `github.com/charmbracelet/bubbletea` (optional) | Long doc paging |

### Build Integration

```makefile
# Makefile
.PHONY: build

# Copy docs before building
docs:
    mkdir -p internal/help/docs
    cp -r knowledge-base/cli/* internal/help/docs/

build: docs
    go build -o fazt ./cmd/server

# Verify docs match binary version
check-docs:
    @VERSION=$$(go run ./cmd/server --version | cut -d' ' -f2)
    @DOC_VERSION=$$(yq '.version' knowledge-base/cli/_meta.yaml)
    @if [ "$$VERSION" != "$$DOC_VERSION" ]; then \
        echo "Version mismatch: binary=$$VERSION docs=$$DOC_VERSION"; \
        exit 1; \
    fi
```

---

## Part 3: Test Strategy Design

### Test Categories

#### Category A: Unit Tests (Command Parsing)

Test individual command parsing without execution.

```go
// cmd/server/cli_test.go

func TestPeerPrefixParsing(t *testing.T) {
    tests := []struct {
        name     string
        args     []string
        wantPeer string
        wantCmd  string
        wantArgs []string
    }{
        {
            name:     "no peer prefix",
            args:     []string{"app", "list"},
            wantPeer: "",
            wantCmd:  "app",
            wantArgs: []string{"list"},
        },
        {
            name:     "with peer prefix",
            args:     []string{"@zyt", "app", "list"},
            wantPeer: "zyt",
            wantCmd:  "app",
            wantArgs: []string{"list"},
        },
        {
            name:     "local peer explicit",
            args:     []string{"@local", "app", "list"},
            wantPeer: "local",
            wantCmd:  "app",
            wantArgs: []string{"list"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            peer, cmd, args := parseCLIArgs(tt.args)
            if peer != tt.wantPeer {
                t.Errorf("peer = %q, want %q", peer, tt.wantPeer)
            }
            if cmd != tt.wantCmd {
                t.Errorf("cmd = %q, want %q", cmd, tt.wantCmd)
            }
            if !reflect.DeepEqual(args, tt.wantArgs) {
                t.Errorf("args = %v, want %v", args, tt.wantArgs)
            }
        })
    }
}

func TestFlagParsing(t *testing.T) {
    tests := []struct {
        name      string
        args      []string
        wantName  string
        wantSPA   bool
        wantBuild bool
    }{
        {
            name:      "basic deploy",
            args:      []string{"./app"},
            wantName:  "app",
            wantSPA:   false,
            wantBuild: true,
        },
        {
            name:      "with flags",
            args:      []string{"./app", "--name", "blog", "--spa", "--no-build"},
            wantName:  "blog",
            wantSPA:   true,
            wantBuild: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            opts := parseDeployFlags(tt.args)
            if opts.Name != tt.wantName {
                t.Errorf("Name = %q, want %q", opts.Name, tt.wantName)
            }
            if opts.SPA != tt.wantSPA {
                t.Errorf("SPA = %v, want %v", opts.SPA, tt.wantSPA)
            }
            if opts.Build != tt.wantBuild {
                t.Errorf("Build = %v, want %v", opts.Build, !tt.wantBuild)
            }
        })
    }
}
```

#### Category B: Integration Tests (End-to-End)

Test complete command execution with mock/test database.

```go
// cmd/server/integration_test.go

func TestAppDeployIntegration(t *testing.T) {
    // Setup test environment
    tmpDir := t.TempDir()
    dbPath := filepath.Join(tmpDir, "test.db")

    // Initialize test database
    initTestDB(t, dbPath)

    // Create test app directory
    appDir := filepath.Join(tmpDir, "my-app")
    os.Mkdir(appDir, 0755)
    os.WriteFile(filepath.Join(appDir, "index.html"), []byte("<h1>Test</h1>"), 0644)
    os.WriteFile(filepath.Join(appDir, "manifest.json"), []byte(`{"name":"my-app"}`), 0644)

    // Execute deploy command
    ctx := &CommandContext{
        DBPath: dbPath,
        Peer:   nil, // local
        Args:   []string{appDir},
    }

    result, err := executeAppDeploy(ctx)
    if err != nil {
        t.Fatalf("deploy failed: %v", err)
    }

    // Verify deployment
    if result.AppName != "my-app" {
        t.Errorf("AppName = %q, want %q", result.AppName, "my-app")
    }
    if result.FileCount != 2 {
        t.Errorf("FileCount = %d, want %d", result.FileCount, 2)
    }

    // Verify app is in database
    apps := listApps(dbPath)
    if len(apps) != 1 || apps[0].Name != "my-app" {
        t.Errorf("app not found in database")
    }
}

func TestRemoteAppDeployIntegration(t *testing.T) {
    // Setup test peer
    peer := setupTestPeer(t)
    defer peer.Close()

    // Create test app
    appDir := createTestApp(t)

    // Execute remote deploy
    ctx := &CommandContext{
        Peer: peer,
        Args: []string{appDir},
    }

    result, err := executeAppDeploy(ctx)
    if err != nil {
        t.Fatalf("remote deploy failed: %v", err)
    }

    // Verify via remote API
    apps, _ := peer.ListApps()
    if len(apps) != 1 {
        t.Errorf("app not deployed to remote peer")
    }
}
```

#### Category C: Documentation Tests (Examples Work)

Extract and test examples from documentation.

```go
// internal/help/doctest_test.go

func TestDocumentationExamples(t *testing.T) {
    // Load all docs
    docs := loadAllDocs(t)

    for _, doc := range docs {
        for _, example := range doc.Examples {
            t.Run(doc.Command+"/"+example.Title, func(t *testing.T) {
                // Skip examples that need real peers
                if strings.Contains(example.Command, "@zyt") {
                    t.Skip("requires real peer")
                }

                // Parse the example command
                args := parseCommand(example.Command)

                // Execute in test environment
                result := executeInTestEnv(t, args)

                // Verify no error (unless example shows error)
                if result.ExitCode != 0 && !example.ExpectsError {
                    t.Errorf("example failed: %s", result.Stderr)
                }
            })
        }
    }
}

func TestDocsFrontmatterValid(t *testing.T) {
    docs := loadAllDocs(t)

    for _, doc := range docs {
        t.Run(doc.Command, func(t *testing.T) {
            // Required fields
            if doc.Command == "" {
                t.Error("missing command field")
            }
            if doc.Syntax == "" {
                t.Error("missing syntax field")
            }
            if doc.Description == "" {
                t.Error("missing description field")
            }

            // Validate peer config
            if doc.Peer.Supported {
                if !strings.Contains(doc.Syntax, "[@peer]") &&
                   !strings.Contains(doc.Syntax, "@<peer>") {
                    t.Error("peer supported but syntax doesn't show @peer")
                }
            }

            // Validate examples have required fields
            for i, ex := range doc.Examples {
                if ex.Command == "" {
                    t.Errorf("example %d missing command", i)
                }
            }
        })
    }
}
```

#### Category D: Regression Tests (Backward Compatibility)

Ensure old patterns still work during migration.

```go
// cmd/server/regression_test.go

func TestDeprecatedFlagsStillWork(t *testing.T) {
    // These should still work but show deprecation notice
    tests := []struct {
        name string
        args []string
        want string
    }{
        {
            name: "deploy with --to flag",
            args: []string{"app", "deploy", "./app", "--to", "local"},
            want: "Deployed", // Should succeed
        },
        {
            name: "remove with --from flag",
            args: []string{"app", "remove", "myapp", "--from", "local"},
            want: "Removed",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := executeInTestEnv(t, tt.args)

            if !strings.Contains(result.Stdout, tt.want) {
                t.Errorf("command failed or wrong output")
            }

            // Should show deprecation notice
            if !strings.Contains(result.Stderr, "DEPRECATED") {
                t.Errorf("expected deprecation notice")
            }
        })
    }
}

func TestOldBehaviorPreserved(t *testing.T) {
    // Test that all current behaviors work
    tests := []struct {
        name string
        args []string
    }{
        {"list with positional peer", []string{"app", "list", "local"}},
        {"info with --on flag", []string{"app", "info", "myapp", "--on", "local"}},
        {"deploy with --to flag", []string{"app", "deploy", "./app", "--to", "local"}},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := executeInTestEnv(t, tt.args)
            if result.ExitCode != 0 {
                t.Errorf("regression: %s no longer works", tt.name)
            }
        })
    }
}
```

### Test File Structure

```
cmd/server/
├── main.go
├── main_test.go                  # Existing tests
├── cli_test.go                   # Unit tests for CLI parsing
├── integration_test.go           # Integration tests
├── regression_test.go            # Backward compatibility tests
├── testdata/
│   ├── fixtures/
│   │   ├── test-app/
│   │   │   ├── index.html
│   │   │   └── manifest.json
│   │   └── test-db.db
│   └── golden/
│       ├── help-output.txt       # Expected help outputs
│       └── deploy-output.txt

internal/help/
├── embed.go
├── parser.go
├── parser_test.go                # Frontmatter parsing tests
├── render.go
├── render_test.go                # Rendering tests
├── doctest_test.go               # Documentation example tests
└── docs/                         # Embedded at build time
```

### Mock/Fixture Strategy

#### Database Mocking

```go
// testutil/db.go

func NewTestDB(t *testing.T) *TestDB {
    tmpDir := t.TempDir()
    dbPath := filepath.Join(tmpDir, "test.db")

    database.Init(dbPath)

    return &TestDB{
        Path: dbPath,
        DB:   database.GetDB(),
    }
}

func (tdb *TestDB) AddPeer(name, url, token string) {
    remote.AddPeer(tdb.DB, name, url, token, "")
}

func (tdb *TestDB) DeployApp(name string, files map[string][]byte) {
    // Add app to VFS
}
```

#### HTTP Mocking for Remote Calls

```go
// testutil/mockpeer.go

type MockPeer struct {
    *httptest.Server
    apps    []AppInfo
    aliases []AliasInfo
}

func NewMockPeer(t *testing.T) *MockPeer {
    mp := &MockPeer{}

    mux := http.NewServeMux()

    // Mock /api/cmd endpoint
    mux.HandleFunc("/api/cmd", func(w http.ResponseWriter, r *http.Request) {
        var req struct {
            Command string   `json:"command"`
            Args    []string `json:"args"`
        }
        json.NewDecoder(r.Body).Decode(&req)

        // Route to mock handlers
        result := mp.handleCommand(req.Command, req.Args)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "data": map[string]interface{}{
                "success": true,
                "data":    result,
            },
        })
    })

    mp.Server = httptest.NewServer(mux)
    return mp
}
```

---

## Part 4: Implementation Phases

### Phase 0: Preparation (1-2 days)

**Criteria**: Documentation complete, tests written but failing

**Tasks**:
- [ ] Document ALL current CLI behavior exhaustively
- [ ] Identify every command and its current peer handling
- [ ] Create test fixtures (apps, databases, mock peers)
- [ ] Set up test infrastructure (helpers, mocks)

**Deliverables**:
- `knowledge-base/cli/commands.md` - Complete command reference
- `cmd/server/testdata/` - Test fixtures
- `testutil/` - Test utilities

### Phase 1: Tests First (2-3 days)

**Criteria**: All tests written, majority failing (TDD)

**Tasks**:
- [ ] Write unit tests for @peer parsing
- [ ] Write integration tests for local operations
- [ ] Write integration tests for remote operations
- [ ] Write documentation tests (examples extraction)
- [ ] Write regression tests for deprecated patterns

**Test Coverage Targets**:
- Unit tests: 100% of parsing logic
- Integration tests: All command paths
- Regression tests: All current behaviors

### Phase 2: Documentation (2-3 days)

**Criteria**: All docs written, frontmatter complete

**Tasks**:
- [ ] Create `knowledge-base/cli/` structure
- [ ] Write `_meta.yaml` with global config
- [ ] Write `fazt.md` (top-level help)
- [ ] Write all `app/*.md` docs
- [ ] Write all `remote/*.md` docs
- [ ] Write all `server/*.md` docs
- [ ] Write all `service/*.md` docs
- [ ] Write `topics/*.md` guides
- [ ] Validate frontmatter with tests

### Phase 3: Help System (2-3 days)

**Criteria**: `--help` shows docs from knowledge-base

**Tasks**:
- [ ] Implement `internal/help/embed.go` (go:embed)
- [ ] Implement `internal/help/parser.go` (YAML + markdown)
- [ ] Implement `internal/help/render.go` (terminal output)
- [ ] Wire help command to doc loader
- [ ] Add build step to copy docs
- [ ] Test help output matches docs

### Phase 4: CLI Refactor (3-5 days)

**Criteria**: Tests pass, @peer is primary

**Tasks**:
- [ ] Refactor `handlePeerCommand` for all commands
- [ ] Add `@local` as explicit local alias
- [ ] Add deprecation warnings for old flags
- [ ] Update error messages with @peer hints
- [ ] Remove duplicate code paths
- [ ] Run full test suite

### Phase 5: Polish (1-2 days)

**Criteria**: UX is excellent, docs are perfect

**Tasks**:
- [ ] Review all error messages
- [ ] Add helpful hints in errors
- [ ] Test edge cases (no peers, bad networks)
- [ ] Final documentation review
- [ ] Update CLAUDE.md with new patterns

### Total Timeline: 11-18 days

---

## Part 5: Example Documentation

See `knowledge-base/cli/app/deploy.md` for complete example.

---

## Part 6: Migration Checklist

### Commands to Migrate

#### App Commands

- [ ] `app list` - Add @peer support, deprecate positional peer
- [ ] `app info` - Add @peer support, deprecate --on flag
- [ ] `app deploy` - Add @peer support, deprecate --to flag
- [ ] `app remove` - Add @peer support, deprecate --from flag
- [ ] `app logs` - Add @peer support, deprecate --on flag
- [ ] `app install` - Add @peer support, deprecate --to flag
- [ ] `app create` - Local only (no peer support)
- [ ] `app validate` - Local only (no peer support)
- [ ] `app upgrade` - Add @peer support, deprecate --from flag
- [ ] `app pull` - Add @peer support, deprecate --from flag
- [ ] `app link` - Add @peer support, deprecate --to flag
- [ ] `app unlink` - Add @peer support, deprecate --from flag
- [ ] `app reserve` - Add @peer support, deprecate --on flag
- [ ] `app fork` - Add @peer support, deprecate --to flag
- [ ] `app swap` - Add @peer support, deprecate --on flag
- [ ] `app split` - Add @peer support, deprecate --on flag
- [ ] `app lineage` - Add @peer support, deprecate --on flag

#### Server Commands

- [ ] `server info` - Add @peer support (read-only remote)
- [ ] `server init` - Local only
- [ ] `server start` - Local only
- [ ] `server status` - Local only
- [ ] `server set-config` - Local only
- [ ] `server set-credentials` - Local only
- [ ] `server create-key` - Local only

#### Remote Commands

- [ ] No changes (peer management is local)

#### Service Commands

- [ ] No changes (systemd is local)

### Tests to Write

- [ ] `TestPeerPrefixParsing` - Unit test for @peer parsing
- [ ] `TestLocalCommandExecution` - Local operations
- [ ] `TestRemoteCommandExecution` - Remote operations
- [ ] `TestDeprecationWarnings` - Old flags show warnings
- [ ] `TestErrorMessages` - Helpful error messages
- [ ] `TestDocExamples` - Examples in docs work
- [ ] `TestHelpOutput` - Help matches docs
- [ ] `TestLocalAlias` - @local works

### Documentation to Create

- [ ] `_meta.yaml` - CLI metadata
- [ ] `fazt.md` - Top-level help
- [ ] `app/_index.md` - App command group
- [ ] `app/deploy.md` - Deploy command
- [ ] `app/list.md` - List command
- [ ] (... all 17 app subcommands)
- [ ] `remote/_index.md` - Remote command group
- [ ] (... all 5 remote subcommands)
- [ ] `server/_index.md` - Server command group
- [ ] (... all 7 server subcommands)
- [ ] `service/_index.md` - Service command group
- [ ] (... all 5 service subcommands)
- [ ] `topics/peer-syntax.md` - @peer guide
- [ ] `topics/local-first.md` - Philosophy
- [ ] `topics/migration.md` - Migration guide

### Code Changes

- [ ] Add `internal/help/` package
- [ ] Add `go:embed` for docs
- [ ] Refactor `handlePeerCommand`
- [ ] Add `@local` handling
- [ ] Add deprecation warning function
- [ ] Update error formatting
- [ ] Add help command routing

---

## Part 7: Technical Stack

### Core Libraries

| Library | Version | Purpose |
|---------|---------|---------|
| `go:embed` | builtin | Embed docs in binary |
| `gopkg.in/yaml.v3` | v3.0.1 | Parse YAML frontmatter |
| `github.com/yuin/goldmark` | v1.5.4 | Parse markdown |
| `github.com/charmbracelet/glamour` | v0.6.0 | Render markdown to terminal |
| `github.com/fatih/color` | v1.15.0 | Terminal colors |
| `golang.org/x/term` | v0.13.0 | Terminal detection |

### Development Tools

| Tool | Purpose |
|------|---------|
| `go test` | Run tests |
| `go build` | Build binary |
| `yq` | YAML processing (CI) |
| `make` | Build orchestration |

### CI Integration

```yaml
# .github/workflows/cli.yaml
name: CLI Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Verify doc versions
        run: |
          BINARY_VERSION=$(go run ./cmd/server --version | cut -d' ' -f2)
          DOC_VERSION=$(yq '.version' knowledge-base/cli/_meta.yaml)
          if [ "$BINARY_VERSION" != "$DOC_VERSION" ]; then
            echo "Version mismatch!"
            exit 1
          fi

      - name: Run tests
        run: go test -v ./...

      - name: Test doc examples
        run: go test -v -run TestDocumentationExamples ./internal/help/...
```

---

## Part 8: Success Criteria

### Functional

- [ ] `fazt @peer app <cmd>` works for all app commands
- [ ] `fazt app <cmd>` defaults to local
- [ ] `@local` explicitly targets local
- [ ] Old flags show deprecation warnings
- [ ] `--help` shows content from knowledge-base/cli/
- [ ] All tests pass

### Documentation

- [ ] Every command has a doc file
- [ ] Every doc has valid frontmatter
- [ ] Every example is testable
- [ ] Docs render beautifully in terminal
- [ ] Docs render beautifully as web pages

### UX

- [ ] Error messages are helpful
- [ ] Deprecation notices include migration path
- [ ] Help is discoverable (`fazt help @peer`)
- [ ] Completion is straightforward

### Quality

- [ ] 90%+ test coverage on CLI parsing
- [ ] All regression tests pass
- [ ] CI validates doc-binary version sync
- [ ] No regressions in existing functionality

---

## Appendix A: Quick Reference

```
DIRECTORY STRUCTURE
==================
knowledge-base/cli/
├── _meta.yaml          # Version, global config
├── fazt.md            # fazt --help
├── app/               # fazt app --help, subcommands
├── remote/            # fazt remote --help, subcommands
├── server/            # fazt server --help, subcommands
├── service/           # fazt service --help, subcommands
└── topics/            # Guides (peer-syntax, migration)

FRONTMATTER KEYS
================
command, version, category, syntax, description
arguments[], flags[], peer{}, examples[], related[], errors[]

TEST CATEGORIES
===============
A. Unit tests      - Parsing logic
B. Integration     - Full command execution
C. Documentation   - Examples work
D. Regression      - Old patterns work

IMPLEMENTATION ORDER
====================
0. Prepare       → Document current state
1. Tests First   → Write failing tests (TDD)
2. Documentation → Create all doc files
3. Help System   → Wire --help to docs
4. CLI Refactor  → Make tests pass
5. Polish        → UX refinements
```

---

## Appendix B: Version History

| Version | Date | Changes |
|---------|------|---------|
| 0.18.0 | 2026-02-02 | Initial design document |
