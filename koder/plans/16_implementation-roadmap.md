# Plan 16: Implementation Roadmap

**Date**: January 13, 2026
**Status**: Active
**Depends On**: Plan 15 (Kernel API Spec)
**Goal**: Ship working software incrementally, with tests

---

## 1. Philosophy

This plan is different. It's not a spec—it's a **build order**.

**Constraints**:
1. Each phase results in a deployable, testable release
2. Every feature ships with tests
3. Built by LLM agents (plan for focused work sessions)
4. Live system at zyt.app must stay stable

**Priorities** (from user):
1. Deploy apps to server (verify current works)
2. MCP server for Claude Code control
3. Serverless runtime (Goja + require shim)
4. Visual analytics (as a Fazt app)

---

## 1b. Multi-Instance Architecture

### The Problem

Current design assumes:
- One local CLI talks to ONE server
- Config stored in local `data.db` (single token + single URL)

Reality:
- Users have MULTIPLE servers (zyt.app, home.local, work.xyz)
- Each server has MULTIPLE apps
- MCP needs to understand this topology

### The Model

```
┌─────────────────────────────────────────────────────────────┐
│                     YOUR MACHINE                             │
│                                                              │
│  ┌──────────────┐     ┌─────────────────────────────────┐   │
│  │ Claude Code  │────▶│  MCP Server (local fazt binary) │   │
│  └──────────────┘     └─────────────────────────────────┘   │
│                                    │                         │
│                                    │ reads                   │
│                                    ▼                         │
│                       ┌─────────────────────────┐           │
│                       │  ~/.fazt/config.json    │           │
│                       │  - servers (name→creds) │           │
│                       │  - default_server       │           │
│                       └─────────────────────────┘           │
└────────────────────────────────────┬────────────────────────┘
                                     │
                    deploys to / queries
                                     │
        ┌────────────────────────────┼────────────────────────┐
        ▼                            ▼                        ▼
┌──────────────┐           ┌──────────────┐          ┌──────────────┐
│   zyt.app    │           │  home.local  │          │  work.xyz    │
│   (server)   │           │   (server)   │          │   (server)   │
├──────────────┤           ├──────────────┤          ├──────────────┤
│ - blog       │           │ - homelab    │          │ - dashboard  │
│ - portfolio  │           │ - cameras    │          │ - api        │
│ - analytics  │           │ - iot        │          │ - docs       │
└──────────────┘           └──────────────┘          └──────────────┘
```

### Config File Design

Move from `data.db` to `~/.fazt/config.json` for client mode:

```json
{
  "version": 1,
  "default_server": "zyt",
  "servers": {
    "zyt": {
      "url": "https://zyt.app",
      "token": "fzt_abc123def456..."
    },
    "home": {
      "url": "https://home.local:8080",
      "token": "fzt_xyz789..."
    }
  }
}
```

**Why JSON file instead of SQLite?**
- Client config is simple key-value, doesn't need SQL
- Easy to read/edit manually
- No DB initialization needed for client-only mode
- Cleaner separation: server uses `data.db`, client uses `config.json`

### UX: Smart Defaults (Don't Over-Engineer)

99% of users have ONE server. Don't force `--server` when it's obvious:

| Scenario | `fazt deploy ./blog --name blog` |
|----------|----------------------------------|
| 0 servers configured | Error: "Run `fazt servers add`" |
| 1 server configured | **Just works** (auto-selects) |
| Multiple servers, default set | Uses default |
| Multiple servers, no default | Error: "Specify `--to` or set default" |

**The simple case is simple:**
```bash
# First time setup (once)
fazt servers add prod --url https://zyt.app --token fzt_abc...

# Forever after (no --server needed)
fazt deploy ./blog --name blog
fazt deploy ./portfolio --name portfolio
fazt apps list
```

**Only gets complex when YOU have multiple servers:**
```bash
fazt servers add home --url https://home.local --token fzt_xyz...
fazt deploy ./blog --name blog --to prod
fazt deploy ./iot --name sensors --to home
```

### CLI Changes

**Server Management:**
```bash
# Add a server
fazt servers add zyt --url https://zyt.app --token fzt_abc...
fazt servers add home --url https://home.local:8080 --token fzt_xyz...

# List configured servers
fazt servers list
# Output:
#   NAME    URL                      DEFAULT
#   zyt     https://zyt.app          *
#   home    https://home.local:8080

# Set default
fazt servers default home

# Remove
fazt servers remove home

# Test connection
fazt servers ping zyt
```

**Deploy with Target:**
```bash
# Deploy to default server
fazt deploy ./my-site --name blog

# Deploy to specific server
fazt deploy ./my-site --name blog --to home

# Deploy to specific server (alternative syntax)
fazt deploy ./my-site --name blog --server home
```

**Other Commands:**
```bash
# List apps (default server)
fazt apps list

# List apps on specific server
fazt apps list --server home

# Logs from specific app on specific server
fazt logs blog --server zyt
```

### MCP Tools (Multi-Server Aware)

```
fazt_servers_list
  params: {}
  returns: {
    servers: [
      { name: "zyt", url: "https://zyt.app", is_default: true },
      { name: "home", url: "https://home.local:8080", is_default: false }
    ]
  }

fazt_apps_list
  params: { server?: string }  // omit = default server
  returns: {
    server: "zyt",
    apps: [
      { name: "blog", domain: "blog.zyt.app", ... },
      { name: "portfolio", domain: "portfolio.zyt.app", ... }
    ]
  }

fazt_deploy
  params: {
    server?: string,           // target server (or default)
    app_name: string,          // name for the app
    files: { [path]: content } // files to deploy
  }
  returns: { server, app_name, url, file_count, size_bytes }

fazt_app_logs
  params: { server?: string, app_name: string, limit?: number }
  returns: { server, app_name, logs: [...] }

fazt_app_delete
  params: { server?: string, app_name: string }
  returns: { server, app_name, deleted: true }

fazt_system_status
  params: { server?: string }
  returns: { server, version, uptime, health, app_count }
```

### Migration Path

1. **Phase 0.5** (new): Implement config file + multi-server CLI
2. **Phase 1**: Build MCP on top of multi-server foundation
3. Deprecate `data.db` for client config (keep for server mode)

Backwards compat:
- If `~/.fazt/config.json` doesn't exist, check `data.db` for old config
- Migrate automatically on first use

### Server-Side: No Changes

Each Fazt server is independent:
- Has its own `data.db`
- Has its own `api_keys` table
- Manages its own apps

The multi-server awareness is **client-side only**. Servers don't know about each other (until Mesh in v0.16).

### API Key Flow (Updated)

```
1. SSH to server (e.g., zyt.app)
   $ fazt server create-key --name "my-laptop"
   → Token: fzt_abc123...

2. On your laptop:
   $ fazt servers add zyt --url https://zyt.app --token fzt_abc123...

3. Repeat for other servers:
   $ fazt servers add home --url https://home.local --token fzt_xyz...

4. Deploy:
   $ fazt deploy ./blog --name blog --to zyt
   $ fazt deploy ./homelab --name dashboard --to home

5. Or via MCP (Claude Code):
   → fazt_deploy { server: "zyt", app_name: "blog", files: {...} }
```

### Upgrade Flow (Unchanged)

`fazt upgrade` works the same—it upgrades the local binary.

Each server is upgraded independently by SSH-ing to it and running `fazt upgrade` there.

---

## 2. Current State Assessment

### What Works Today (v0.7.x)

| Capability | Status | Notes |
|------------|--------|-------|
| `fazt deploy` | Works | ZIP upload to VFS |
| Static hosting | Works | Subdomain routing |
| Analytics | Works | Buffered events, `/api/stats` |
| Admin dashboard | Works | Embedded React SPA |
| Install script | Works | systemd + CLI modes |
| Self-upgrade | Works | `fazt upgrade` |
| HTTPS/TLS | Works | CertMagic |

### What's Missing

| Capability | Status | Blocking |
|------------|--------|----------|
| Serverless runtime | Partial | `main.js` detected but not executed properly |
| MCP server | Not started | Can't control from Claude Code |
| Apps model | Not started | Still "sites" not "apps" |
| `require()` shim | Not started | Can't split serverless code |

---

## 2b. Execution Checklists

Each phase has a checklist of files to create/modify and tests to run.
Claude should use these as the source of truth for execution.

### Phase 0.5 Checklist: Multi-Server Config

**Files to Create:**
```
internal/clientconfig/
├── config.go           # Config struct, Load(), Save(), GetServer()
└── config_test.go      # Tests for all methods
```

**Files to Modify:**
```
cmd/server/main.go      # Add 'servers' command group
                        # Update handleDeployCommand to use clientconfig
```

**Tests to Run:**
```bash
go test -v -cover ./internal/clientconfig/...
```

**Definition of Done:**
- [ ] `fazt servers add` works
- [ ] `fazt servers list` works
- [ ] `fazt servers default` works
- [ ] `fazt servers remove` works
- [ ] `fazt deploy` uses new config
- [ ] Migration from old data.db works
- [ ] All tests pass with >80% coverage

---

### Phase 1 Checklist: MCP Server

**Files to Create:**
```
internal/mcp/
├── server.go           # MCP server, transport layer
├── tools.go            # Tool definitions
├── handler.go          # HTTP handlers for /mcp/*
└── server_test.go      # Tests
```

**Files to Modify:**
```
cmd/server/main.go      # Add 'fazt server create-key' command
                        # Add 'fazt mcp' command (optional)
internal/handlers/routes.go  # Register /mcp/* routes (or main.go)
```

**Tests to Run:**
```bash
go test -v -cover ./internal/mcp/...
go test -v -cover ./internal/clientconfig/...
```

**Definition of Done:**
- [ ] `POST /mcp/initialize` works
- [ ] `POST /mcp/tools/list` returns tool list
- [ ] `POST /mcp/tools/call` executes tools
- [ ] `fazt_servers_list` tool works
- [ ] `fazt_apps_list` tool works
- [ ] `fazt_deploy` tool works
- [ ] `fazt_app_delete` tool works
- [ ] `fazt_app_logs` tool works
- [ ] `fazt_system_status` tool works
- [ ] `fazt server create-key` works
- [ ] All tests pass with >80% coverage

---

### Phase 2 Checklist: Serverless Runtime

**Files to Create:**
```
internal/runtime/
├── runtime.go          # VM pool, Execute()
├── inject.go           # Request/response injection
├── require.go          # require() shim
├── fazt.go             # fazt.* namespace
├── runtime_test.go     # Core tests
├── require_test.go     # require() tests
└── fazt_test.go        # Namespace tests
```

**Files to Modify:**
```
internal/handlers/hosting.go  # Route /api/* to runtime
                              # Or create serverless.go
```

**Dependencies:**
```bash
go get github.com/dop251/goja
```

**Tests to Run:**
```bash
go test -v -cover ./internal/runtime/...
```

**Definition of Done:**
- [ ] `api/main.js` executes on `/api/*` requests
- [ ] `request` object injected correctly
- [ ] `respond()` helper works
- [ ] `console.log()` captured
- [ ] `require('./file.js')` works
- [ ] `require()` can't escape `api/` folder
- [ ] Circular dependencies handled
- [ ] Timeout enforced (100ms)
- [ ] `fazt.app.id`, `fazt.app.name` work
- [ ] `fazt.env.get()` works
- [ ] All tests pass with >80% coverage

---

### Phase 5 Checklist: Release

**Pre-release:**
```bash
# All tests pass
go test -v -cover ./...

# Build succeeds
go build -o fazt ./cmd/server

# Admin SPA builds
cd admin && npm run build
```

**Release Steps:**
```bash
# Update version in cmd/server/main.go
# Update CHANGELOG.md

# Commit
git add -A
git commit -m "chore: prepare v0.8.0 release"

# Tag
git tag v0.8.0

# Push
git push origin master
git push origin v0.8.0

# Wait for GitHub Actions to build
# Check: https://github.com/fazt-sh/fazt/releases
```

---

### Phase 6 Checklist: Deploy to Production

**User Actions (provide these steps to user):**
```bash
# SSH to server
ssh user@zyt.app

# Upgrade fazt
sudo fazt upgrade

# Verify
fazt version
```

---

### Phase 7 Checklist: Local Setup

**User Actions:**
```bash
# On server: create API key
fazt server create-key --name "laptop"
# → Note the token

# On laptop: configure
fazt servers add zyt --url https://zyt.app --token <TOKEN>

# Test
fazt apps list
```

---

### Phase 8 Checklist: MCP Setup

**Claude Code Config:**
```json
// Add to Claude Code MCP settings
{
  "fazt": {
    "command": "fazt",
    "args": ["mcp", "serve"]
  }
}
```

**Test:**
```
Use fazt_servers_list tool
Use fazt_apps_list tool
Use fazt_deploy tool
```

---

## 3. Release Phases

### Phase 0: Verification (1 session)

**Goal**: Confirm current deployment works on live server.

**Tasks**:
```
[ ] Create test static site (index.html + CSS)
[ ] Deploy to zyt.app: fazt deploy ./test-site --name hello
[ ] Verify https://hello.zyt.app loads
[ ] Check analytics captures the visit
[ ] Document any issues found
```

**Deliverable**: Confidence that v0.7 works, or list of bugs to fix.

**Test**: Manual verification + screenshot.

---

### Phase 0.5: Multi-Server Config (2-3 sessions)

**Goal**: Implement `~/.fazt/config.json` and multi-server CLI before MCP.

This is **prerequisite** for MCP—can't build MCP on single-server foundation.

#### 0.5.1 Config Package

```go
// internal/clientconfig/config.go
package clientconfig

import (
    "encoding/json"
    "os"
    "path/filepath"
)

type Config struct {
    Version       int               `json:"version"`
    DefaultServer string            `json:"default_server"`
    Servers       map[string]Server `json:"servers"`
}

type Server struct {
    URL   string `json:"url"`
    Token string `json:"token"`
}

func ConfigPath() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".fazt", "config.json")
}

func Load() (*Config, error) {
    data, err := os.ReadFile(ConfigPath())
    if os.IsNotExist(err) {
        return &Config{Version: 1, Servers: make(map[string]Server)}, nil
    }
    if err != nil {
        return nil, err
    }
    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}

func (c *Config) Save() error {
    dir := filepath.Dir(ConfigPath())
    if err := os.MkdirAll(dir, 0700); err != nil {
        return err
    }
    data, err := json.MarshalIndent(c, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(ConfigPath(), data, 0600)
}

func (c *Config) GetServer(name string) (*Server, error) {
    // Explicit server specified
    if name != "" {
        srv, ok := c.Servers[name]
        if !ok {
            return nil, fmt.Errorf("server '%s' not found", name)
        }
        return &srv, nil
    }

    // No server specified - try smart defaults

    // Case 1: Default is set → use it
    if c.DefaultServer != "" {
        srv, ok := c.Servers[c.DefaultServer]
        if ok {
            return &srv, nil
        }
    }

    // Case 2: Only ONE server configured → use it (no need for --server)
    if len(c.Servers) == 1 {
        for _, srv := range c.Servers {
            return &srv, nil
        }
    }

    // Case 3: Zero servers
    if len(c.Servers) == 0 {
        return nil, fmt.Errorf("no servers configured\nRun: fazt servers add <name> --url <url> --token <token>")
    }

    // Case 4: Multiple servers, no default
    return nil, fmt.Errorf("multiple servers configured, specify --to <server> or set default:\n  fazt servers default <name>")
}
```

#### 0.5.2 CLI Commands

```bash
# New command group
fazt servers add <name> --url <url> --token <token>
fazt servers list
fazt servers default <name>
fazt servers remove <name>
fazt servers ping <name>
```

```go
// cmd/server/main.go
case "servers":
    handleServersCommand(os.Args[2:])

func handleServersCommand(args []string) {
    if len(args) == 0 {
        printServersHelp()
        return
    }
    switch args[0] {
    case "add":
        handleServersAdd(args[1:])
    case "list":
        handleServersList()
    case "default":
        handleServersDefault(args[1:])
    case "remove":
        handleServersRemove(args[1:])
    case "ping":
        handleServersPing(args[1:])
    }
}
```

#### 0.5.3 Update Deploy Command

```go
func handleDeployCommand() {
    flags := flag.NewFlagSet("deploy", flag.ExitOnError)
    path := flags.String("path", ".", "Path to deploy")
    name := flags.String("name", "", "App name (required)")
    server := flags.String("to", "", "Target server (default: default server)")
    // Also accept --server as alias
    flags.StringVar(server, "server", "", "Target server (alias for --to)")

    flags.Parse(os.Args[2:])

    // Load config
    cfg, err := clientconfig.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Get target server
    srv, err := cfg.GetServer(*server)
    if err != nil {
        log.Fatalf("Error: %v\nRun 'fazt servers add' to configure a server.", err)
    }

    // Deploy to srv.URL with srv.Token
    // ... rest of deploy logic
}
```

#### 0.5.4 Migration from Old Config

```go
func migrateOldConfig() {
    cfg, _ := clientconfig.Load()
    if len(cfg.Servers) > 0 {
        return // Already migrated
    }

    // Check old data.db
    dbPath := "./data.db"
    if envPath := os.Getenv("FAZT_DB_PATH"); envPath != "" {
        dbPath = envPath
    }

    if _, err := os.Stat(dbPath); os.IsNotExist(err) {
        return // No old config
    }

    // Extract old config
    db, err := sql.Open("sqlite", dbPath)
    if err != nil {
        return
    }
    defer db.Close()

    var token, serverURL string
    db.QueryRow("SELECT value FROM config WHERE key = 'api_key.token'").Scan(&token)
    db.QueryRow("SELECT value FROM config WHERE key = 'client.server_url'").Scan(&serverURL)

    if token != "" && serverURL != "" {
        cfg.Servers["default"] = clientconfig.Server{
            URL:   serverURL,
            Token: token,
        }
        cfg.DefaultServer = "default"
        cfg.Save()
        fmt.Println("Migrated config from data.db to ~/.fazt/config.json")
    }
}
```

**Tests**:
```go
// internal/clientconfig/config_test.go
func TestConfig_LoadEmpty(t *testing.T) { ... }
func TestConfig_SaveAndLoad(t *testing.T) { ... }
func TestConfig_GetServer_Default(t *testing.T) { ... }
func TestConfig_GetServer_Named(t *testing.T) { ... }
func TestConfig_GetServer_NotFound(t *testing.T) { ... }

// cmd/server/servers_test.go
func TestServersAdd(t *testing.T) { ... }
func TestServersList(t *testing.T) { ... }
func TestServersDefault(t *testing.T) { ... }
func TestServersRemove(t *testing.T) { ... }
func TestServersPing(t *testing.T) { ... }
```

**Deliverable**: Multi-server CLI works, MCP can be built on top.

---

### Phase 1: MCP Server (3-4 sessions)

**Goal**: Control Fazt from Claude Code locally.

**Why first**: Enables faster development—you can deploy/debug via Claude Code
instead of SSH + CLI.

#### 1.1 MCP Transport Layer

Implement MCP-over-HTTP on the existing server (no separate port).

```
Endpoints:
  POST /mcp/initialize      → Handshake
  POST /mcp/tools/list      → List available tools
  POST /mcp/tools/call      → Execute a tool
```

**Implementation**:
```go
// internal/mcp/server.go
type MCPServer struct {
    db      *sql.DB
    tools   map[string]Tool
}

type Tool struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{} `json:"inputSchema"`
    Handler     func(params map[string]interface{}) (interface{}, error)
}
```

**Files to create**:
- `internal/mcp/server.go` - Core MCP logic
- `internal/mcp/tools.go` - Tool definitions
- `internal/mcp/handler.go` - HTTP handlers
- `internal/mcp/server_test.go` - Tests

#### 1.2 Core Tools (Multi-Server Aware)

All tools read from `~/.fazt/config.json` and support optional `server` param:

```go
// internal/mcp/tools.go

type MCPTools struct {
    config *clientconfig.Config
}

func (t *MCPTools) Register() []Tool {
    return []Tool{
        {
            Name: "fazt_servers_list",
            Description: "List configured Fazt servers",
            Handler: t.serversList,
        },
        {
            Name: "fazt_apps_list",
            Description: "List apps on a server",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "server": {"type": "string", "description": "Server name (omit for default)"},
                },
            },
            Handler: t.appsList,
        },
        {
            Name: "fazt_deploy",
            Description: "Deploy files to an app",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "server":   {"type": "string", "description": "Server name (omit for default)"},
                    "app_name": {"type": "string", "description": "App name"},
                    "files":    {"type": "object", "description": "Map of path -> content"},
                },
                "required": []string{"app_name", "files"},
            },
            Handler: t.deploy,
        },
        {
            Name: "fazt_app_delete",
            Description: "Delete an app",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "server":   {"type": "string"},
                    "app_name": {"type": "string"},
                },
                "required": []string{"app_name"},
            },
            Handler: t.appDelete,
        },
        {
            Name: "fazt_app_logs",
            Description: "Get logs for an app",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "server":   {"type": "string"},
                    "app_name": {"type": "string"},
                    "limit":    {"type": "integer", "default": 100},
                },
                "required": []string{"app_name"},
            },
            Handler: t.appLogs,
        },
        {
            Name: "fazt_system_status",
            Description: "Get server status",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "server": {"type": "string"},
                },
            },
            Handler: t.systemStatus,
        },
    }
}

func (t *MCPTools) appsList(params map[string]interface{}) (interface{}, error) {
    serverName, _ := params["server"].(string)
    srv, err := t.config.GetServer(serverName)
    if err != nil {
        return nil, err
    }

    // Call server's /api/sites
    resp, err := t.httpGet(srv, "/api/sites")
    if err != nil {
        return nil, err
    }

    return map[string]interface{}{
        "server": serverName,
        "apps":   resp["data"],
    }, nil
}

func (t *MCPTools) deploy(params map[string]interface{}) (interface{}, error) {
    serverName, _ := params["server"].(string)
    appName := params["app_name"].(string)
    files := params["files"].(map[string]interface{})

    srv, err := t.config.GetServer(serverName)
    if err != nil {
        return nil, err
    }

    // Create ZIP in memory from files map
    zipBuffer := createZipFromFiles(files)

    // POST to /api/deploy
    resp, err := t.httpPostMultipart(srv, "/api/deploy", appName, zipBuffer)
    if err != nil {
        return nil, err
    }

    return map[string]interface{}{
        "server":     serverName,
        "app_name":   appName,
        "url":        fmt.Sprintf("https://%s.%s", appName, extractDomain(srv.URL)),
        "file_count": resp["file_count"],
        "size_bytes": resp["size_bytes"],
    }, nil
}
```

#### 1.3 Authentication

MCP requests authenticated via:
- Bearer token (same as CLI deploy)
- Only works from localhost OR with valid token

```go
func (s *MCPServer) authenticate(r *http.Request) bool {
    // Option 1: Request from localhost
    if isLocalhost(r.RemoteAddr) {
        return true
    }
    // Option 2: Valid Bearer token
    token := extractBearerToken(r)
    return s.validateToken(token)
}
```

#### 1.4 Headless Key Creation (CLI)

Add `fazt server create-key` for headless environments:

```go
// cmd/server/main.go - add to handleServerCommand()
case "create-key":
    handleCreateKeyCommand()

func handleCreateKeyCommand() {
    flags := flag.NewFlagSet("create-key", flag.ExitOnError)
    name := flags.String("name", "", "Key name (required)")

    flags.Parse(os.Args[3:])

    if *name == "" {
        fmt.Println("Error: --name is required")
        os.Exit(1)
    }

    // Initialize DB
    db := database.GetDB()
    token, err := hosting.CreateAPIKey(db, *name, "deploy")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }

    fmt.Println("API Key created successfully!")
    fmt.Println()
    fmt.Printf("Token: %s\n", token)
    fmt.Println()
    fmt.Println("⚠️  Save this token - it won't be shown again!")
    fmt.Println()
    fmt.Println("To configure your client:")
    fmt.Printf("  fazt client set-auth-token --token %s --server <YOUR_SERVER_URL>\n", token)
}
```

**Tests**:
```go
func TestCLI_CreateKey(t *testing.T) { ... }
func TestCLI_CreateKey_RequiresName(t *testing.T) { ... }
```

#### 1.5 Claude Code Integration

Create MCP config for local Claude Code:

```json
// ~/.claude/mcp-servers/fazt.json
{
  "fazt": {
    "type": "http",
    "url": "https://zyt.app/mcp",
    "headers": {
      "Authorization": "Bearer ${FAZT_TOKEN}"
    }
  }
}
```

**Tests**:
```go
// internal/mcp/server_test.go
func TestMCPInitialize(t *testing.T) { ... }
func TestMCPToolsList(t *testing.T) { ... }
func TestMCPToolCall_SystemStatus(t *testing.T) { ... }
func TestMCPToolCall_AppList(t *testing.T) { ... }
func TestMCPToolCall_AppDeploy(t *testing.T) { ... }
func TestMCPToolCall_AppDelete(t *testing.T) { ... }
func TestMCPToolCall_LogsRead(t *testing.T) { ... }
func TestMCPAuth_Localhost(t *testing.T) { ... }
func TestMCPAuth_BearerToken(t *testing.T) { ... }
func TestMCPAuth_Rejected(t *testing.T) { ... }
```

**Deliverable**: `fazt_*` tools work from Claude Code.

---

### Phase 2: Serverless Runtime (3-4 sessions)

**Goal**: Execute JavaScript in `api/main.js` on HTTP requests.

#### 2.1 Goja Runtime Foundation

```go
// internal/runtime/runtime.go
type Runtime struct {
    db     *sql.DB
    pool   *VMPool  // Reuse VMs for performance
}

type ExecuteResult struct {
    Status  int
    Headers map[string]string
    Body    interface{}
    Logs    []LogEntry
}

func (r *Runtime) Execute(ctx context.Context, appID string, req Request) (*ExecuteResult, error)
```

**Implementation notes**:
- Use `github.com/dop251/goja` for JS execution
- VM pool with max 10 concurrent VMs
- 100ms timeout per execution (configurable)
- 64MB memory limit per VM

#### 2.2 Request/Response Injection

```javascript
// Injected globals in api/main.js

// Request object (read-only)
request = {
    method: "POST",
    path: "/api/users",
    query: { id: "123" },
    headers: { "content-type": "application/json" },
    body: { name: "Alice" }
}

// Response helper
function respond(statusOrBody, body) {
    if (typeof statusOrBody === 'number') {
        return { status: statusOrBody, body: body };
    }
    return { status: 200, body: statusOrBody };
}

// Console (captured)
console.log("debug info")  // → stored in logs
console.error("oops")      // → stored in logs
```

#### 2.3 Routing Logic

```
Request: GET /api/users?id=123
         ↓
1. Check if app has api/main.js in VFS
   - No  → 404 "No serverless handler"
   - Yes → Continue
         ↓
2. Load api/main.js from VFS
         ↓
3. Create Goja VM, inject request object
         ↓
4. Execute script (with timeout)
         ↓
5. Extract return value as response
         ↓
6. Return HTTP response
```

**Routing table**:
```
/api/*           → Serverless (if api/main.js exists)
/api/*           → 404 (if no api/main.js)
/*               → Static files from VFS
```

#### 2.4 `require()` Shim

```go
// internal/runtime/require.go
func (r *Runtime) shimRequire(vm *goja.Runtime, appID string, basePath string) {
    vm.Set("require", func(call goja.FunctionCall) goja.Value {
        path := call.Argument(0).String()

        // Security: Only allow relative paths within api/
        if !isRelativePath(path) {
            panic(vm.NewGoError(fmt.Errorf("require() only accepts relative paths")))
        }

        resolved := resolvePath(basePath, path)
        if !strings.HasPrefix(resolved, "api/") {
            panic(vm.NewGoError(fmt.Errorf("cannot require files outside api/")))
        }

        // Load from VFS
        content, err := r.loadFile(appID, resolved)
        if err != nil {
            panic(vm.NewGoError(err))
        }

        // Check cache (prevent circular deps)
        if cached, ok := r.moduleCache[resolved]; ok {
            return cached
        }

        // Execute module
        module := r.executeModule(vm, resolved, content)
        r.moduleCache[resolved] = module

        return module
    })
}
```

**Example**:
```javascript
// api/main.js
const db = require('./db.js');
const { validate } = require('./utils.js');

if (request.method === 'POST') {
    const user = request.body;
    if (!validate(user)) {
        return respond(400, { error: "Invalid user" });
    }
    db.createUser(user);
    return respond(201, { id: user.id });
}

return respond(200, db.listUsers());
```

```javascript
// api/db.js
const users = [];
module.exports = {
    createUser: (u) => users.push(u),
    listUsers: () => users
};
```

```javascript
// api/utils.js
module.exports = {
    validate: (user) => user.name && user.email
};
```

#### 2.5 Basic `fazt.*` Namespace

Start minimal, expand later:

```javascript
fazt.app.id      // Current app UUID
fazt.app.name    // Current app name
fazt.env.get(key) // Environment variable (from env_vars table)
fazt.log.info(msg)
fazt.log.warn(msg)
fazt.log.error(msg)
```

**Tests**:
```go
// internal/runtime/runtime_test.go
func TestRuntime_SimpleHandler(t *testing.T) { ... }
func TestRuntime_RequestInjection(t *testing.T) { ... }
func TestRuntime_Respond(t *testing.T) { ... }
func TestRuntime_Console(t *testing.T) { ... }
func TestRuntime_Timeout(t *testing.T) { ... }
func TestRuntime_MemoryLimit(t *testing.T) { ... }

// internal/runtime/require_test.go
func TestRequire_LocalFile(t *testing.T) { ... }
func TestRequire_NestedDependency(t *testing.T) { ... }
func TestRequire_CircularDependency(t *testing.T) { ... }
func TestRequire_SecurityEscapeAttempt(t *testing.T) { ... }
func TestRequire_NotFound(t *testing.T) { ... }

// internal/runtime/fazt_test.go
func TestFazt_AppInfo(t *testing.T) { ... }
func TestFazt_EnvGet(t *testing.T) { ... }
func TestFazt_Logging(t *testing.T) { ... }
```

**Deliverable**: `api/main.js` executes on HTTP requests to `/api/*`.

---

### Phase 3: Analytics Dashboard App (2 sessions)

**Goal**: Visual analytics like Plausible, built as a Fazt app.

This dogfoods the serverless runtime—analytics dashboard is itself a Fazt app
that queries the analytics API.

#### 3.1 App Structure

```
analytics-app/
├── index.html        # SPA shell
├── styles.css        # Styling
├── app.js            # Frontend JS (vanilla, no framework)
├── api/
│   └── main.js       # Backend: proxy to /api/stats, /api/events
└── app.json          # Manifest
```

#### 3.2 Frontend (Minimal, Vanilla JS)

No React/Vue—keep it simple. Use:
- [Chart.js](https://www.chartjs.org/) for charts (CDN)
- Vanilla JS for DOM manipulation
- CSS Grid for layout

**Features**:
- Page views over time (line chart)
- Top pages (bar chart)
- Top referrers (bar chart)
- Visitor map (optional, if geo data exists)
- Real-time visitor count
- Date range picker (7d, 30d, 90d)

#### 3.3 Backend (`api/main.js`)

```javascript
// api/main.js
const { fetchStats, fetchEvents } = require('./data.js');

if (request.path === '/api/stats') {
    const range = request.query.range || '7d';
    return respond(fetchStats(range));
}

if (request.path === '/api/events') {
    const { limit, offset } = request.query;
    return respond(fetchEvents(limit, offset));
}

return respond(404, { error: 'Not found' });
```

```javascript
// api/data.js
// For now, use fazt.fetch() to call the main API
// Later, could use fazt.storage for direct DB access

module.exports = {
    fetchStats: async (range) => {
        // Proxy to main analytics API
        const res = await fazt.net.fetch('/api/stats?range=' + range);
        return res.json();
    },
    fetchEvents: async (limit, offset) => {
        const res = await fazt.net.fetch(`/api/events?limit=${limit}&offset=${offset}`);
        return res.json();
    }
};
```

**Note**: This requires `fazt.net.fetch()` to be implemented. Add to Phase 2.

#### 3.4 Deployment

```bash
fazt deploy ./analytics-app --name analytics
# Access at: https://analytics.zyt.app
```

**Tests**:
```go
// Integration test: deploy analytics app, verify it renders
func TestAnalyticsApp_Deploy(t *testing.T) { ... }
func TestAnalyticsApp_StatsEndpoint(t *testing.T) { ... }
func TestAnalyticsApp_FrontendLoads(t *testing.T) { ... }
```

**Deliverable**: Plausible-like dashboard at `analytics.zyt.app`.

---

### Phase 4: Sites → Apps Migration (2 sessions)

**Goal**: Rename "sites" to "apps", add UUIDs, prepare for v0.8.

This is the schema/API migration from Plan 15. Do it after serverless works
so we don't break things mid-development.

#### 4.1 Database Migration

```sql
-- migrations/004_sites_to_apps.sql

-- Add new columns
ALTER TABLE sites ADD COLUMN source TEXT DEFAULT 'personal';
ALTER TABLE sites ADD COLUMN manifest TEXT;

-- Generate stable app IDs (if not present)
UPDATE sites SET id = 'app_' || hex(randomblob(4))
WHERE id NOT LIKE 'app_%';

-- Rename table
ALTER TABLE sites RENAME TO apps;

-- Update files table
ALTER TABLE files RENAME COLUMN site_id TO app_id;

-- Create domains table (decouple from apps)
CREATE TABLE IF NOT EXISTS domains (
    id TEXT PRIMARY KEY,
    domain TEXT NOT NULL UNIQUE,
    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    is_primary INTEGER DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Migrate existing domains
INSERT INTO domains (id, domain, app_id, is_primary, created_at)
SELECT 'dom_' || hex(randomblob(4)), name, id, 1, created_at
FROM apps
WHERE name IS NOT NULL;
```

#### 4.2 API Aliases (Backwards Compat)

```go
// internal/handlers/compat.go

// Old endpoints redirect/alias to new
func RegisterCompatRoutes(r chi.Router) {
    // /api/sites → /api/apps
    r.Get("/api/sites", withDeprecationWarning(appsHandler.List))
    r.Get("/api/sites/{id}", withDeprecationWarning(appsHandler.Get))

    // Old deploy → new deploy
    r.Post("/api/deploy", withDeprecationWarning(appsHandler.DeployByName))
}

func withDeprecationWarning(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Deprecation-Warning", "This endpoint is deprecated. Use /api/apps instead.")
        next(w, r)
    }
}
```

#### 4.3 CLI Updates

```bash
# New commands
fazt apps list
fazt apps create <name>
fazt apps delete <name>
fazt apps info <name>

# Old command still works
fazt deploy ./my-app --name blog  # Still works, uses new backend
```

**Tests**:
```go
func TestMigration_SitesToApps(t *testing.T) { ... }
func TestMigration_DomainsTable(t *testing.T) { ... }
func TestAPI_AppsEndpoints(t *testing.T) { ... }
func TestAPI_SitesBackwardsCompat(t *testing.T) { ... }
func TestCLI_AppsCommands(t *testing.T) { ... }
```

**Deliverable**: Clean "apps" model, old CLI still works.

---

## 4. Implementation Order

```
Week 1:  Phase 0 (verify) + Phase 1.1-1.2 (MCP transport + tools)
Week 2:  Phase 1.3-1.4 (MCP auth + Claude integration)
Week 3:  Phase 2.1-2.2 (Goja runtime + request/response)
Week 4:  Phase 2.3-2.4 (routing + require shim)
Week 5:  Phase 2.5 + Phase 3 (fazt namespace + analytics app)
Week 6:  Phase 4 (sites → apps migration)
```

**Notes**:
- "Week" is approximate—depends on session velocity
- Each phase can ship independently
- Later phases can start before earlier ones finish (with care)

---

## 5. Session Planning for LLM Agents

Each "session" is one focused Claude Code work session (~30-60 min).

### Session Template

```
Session: Phase X.Y - [Name]
─────────────────────────────
Goal: [One sentence]

Context files to read:
- koder/plans/16_implementation-roadmap.md (this file, relevant section)
- [specific code files]

Tasks:
1. [ ] Create/modify files
2. [ ] Write tests
3. [ ] Run tests
4. [ ] Commit with message format: "feat(scope): description"

Definition of Done:
- [ ] All tests pass
- [ ] Code compiles
- [ ] Manual smoke test (if applicable)
```

### Session Breakdown

| Session | Phase | Focus | Files |
|---------|-------|-------|-------|
| 1 | 0 | Verify deploy works | Manual test on zyt.app |
| 2 | 0.5.1 | Config package | internal/clientconfig/config.go |
| 3 | 0.5.2 | `fazt servers` CLI | cmd/server/main.go (servers cmds) |
| 4 | 0.5.3 | Update deploy + migrate | cmd/server/main.go (deploy) |
| 5 | 1.1 | MCP transport | internal/mcp/server.go, handler.go |
| 6 | 1.2 | MCP tools (multi-server) | internal/mcp/tools.go |
| 7 | 1.3 | Headless key creation | cmd/server/main.go |
| 8 | 1.4 | Claude integration | docs, manual test |
| 9 | 2.1 | Goja foundation | internal/runtime/runtime.go |
| 10 | 2.2 | Request/response | internal/runtime/inject.go |
| 11 | 2.3 | Routing | internal/handlers/serverless.go |
| 12 | 2.4 | require() shim | internal/runtime/require.go |
| 13 | 2.5 | fazt namespace | internal/runtime/fazt.go |
| 14 | 3.1-3.2 | Analytics frontend | analytics-app/ |
| 15 | 3.3-3.4 | Analytics backend + deploy | analytics-app/api/ |
| 16 | 4.1 | DB migration | migrations/, internal/db/ |
| 17 | 4.2-4.3 | API + CLI updates | internal/handlers/, cmd/ |

---

## 6. Testing Strategy

### Unit Tests

Every Go package has `*_test.go`:
```
internal/mcp/server_test.go
internal/runtime/runtime_test.go
internal/runtime/require_test.go
internal/runtime/fazt_test.go
```

Run: `go test ./internal/...`

### Integration Tests

Test full request flows:
```go
// internal/integration/deploy_test.go
func TestDeploy_StaticSite(t *testing.T) { ... }
func TestDeploy_ServerlessApp(t *testing.T) { ... }
func TestMCP_FullFlow(t *testing.T) { ... }
```

Run: `go test ./internal/integration/...`

### E2E Tests (Manual Checklist)

Before each release:
```
[ ] Deploy static site to zyt.app
[ ] Deploy serverless app to zyt.app
[ ] Check analytics captures visits
[ ] Test MCP from Claude Code
[ ] Run fazt upgrade on server
[ ] Verify no downtime during upgrade
```

---

## 7. Release Process

### Version Numbering

```
v0.7.x  - Current stable (bug fixes only)
v0.8.0  - MCP + Serverless + Apps model
```

### Release Checklist

```
[ ] All tests pass locally
[ ] Push to master
[ ] Tag release: git tag v0.8.0
[ ] Push tag: git push origin v0.8.0
[ ] GitHub Actions builds + publishes
[ ] SSH to server: curl -sL https://raw.githubusercontent.com/.../install.sh | bash
[ ] Verify upgrade successful
[ ] Announce (if applicable)
```

---

## 8. Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Serverless breaks static hosting | Route serverless ONLY to /api/*, static unchanged |
| MCP exposes security hole | Localhost-only by default, token required for remote |
| require() escapes sandbox | Strict path validation, only api/ folder |
| Migration breaks existing sites | Run migration on test DB first, backup before prod |
| VM memory exhaustion | Pool with hard limit (10 VMs), per-VM memory cap |

---

## 9. Success Criteria

At the end of Phase 4:

- [ ] `fazt deploy` works for static AND serverless apps
- [ ] Claude Code can list/deploy/delete apps via MCP
- [ ] Analytics dashboard runs as a Fazt app
- [ ] All tests pass
- [ ] Zero downtime on zyt.app during development
- [ ] Clean "apps" model in database

---

## 10. What's NOT in This Plan

| Feature | When | Why |
|---------|------|-----|
| Scheduler/Jobs | Phase 5+ | Need solid runtime first |
| Events system | Phase 5+ | Need apps model first |
| Marketplace/git install | Phase 6+ | Need manifest support first |
| Multi-user | v0.9+ | Single-admin is fine for now |
| Cron | Phase 5+ | Scheduler prerequisite |

---

## Appendix A: Go Libraries

| Purpose | Library | Notes |
|---------|---------|-------|
| JS runtime | `github.com/dop251/goja` | Pure Go, no CGO |
| HTTP router | `github.com/go-chi/chi/v5` | Already in use |
| SQLite | `modernc.org/sqlite` | Pure Go, no CGO |
| JSON Schema | `github.com/santhosh-tekuri/jsonschema` | For MCP validation |
| UUID | `github.com/google/uuid` | For app IDs |

---

## Appendix B: MCP Protocol Reference

Based on [Model Context Protocol](https://modelcontextprotocol.io/):

```
POST /mcp/initialize
Request:  { "protocolVersion": "1.0", "capabilities": {} }
Response: { "protocolVersion": "1.0", "serverInfo": { "name": "fazt", "version": "0.8.0" } }

POST /mcp/tools/list
Request:  {}
Response: { "tools": [{ "name": "fazt_system_status", "description": "...", "inputSchema": {...} }] }

POST /mcp/tools/call
Request:  { "name": "fazt_system_status", "arguments": {} }
Response: { "content": [{ "type": "text", "text": "{\"uptime\": 3600, ...}" }] }
```

---

**Plan Status**: Ready for implementation
**Next Action**: Start Phase 0 (verify current deployment works)
