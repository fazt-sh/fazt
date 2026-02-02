# Fazt CLI: @peer-Primary Design

**Version**: 0.18.0
**Status**: IMPLEMENTED
**Date**: 2026-02-02
**Updated**: 2026-02-02

---

## Executive Summary

This document describes the **implemented** `@peer`-primary CLI design in Fazt v0.18.0. The `@peer` prefix is now the **primary modality** for specifying remote operations. The core mental model: **all operations are local by default unless `@peer` is added**.

```bash
fazt app list              # local
fazt @zyt app list         # remote (zyt)

fazt app deploy ./app      # local
fazt @zyt app deploy ./app # remote (zyt)
```

This design fundamentally shifts Fazt's CLI philosophy toward **local-first computing** with explicit remote targeting.

---

## Part 1: Mental Model

### Core Principle

**Local-first, explicit remote.**

| Pattern | Meaning |
|---------|---------|
| `fazt <command>` | Operates on LOCAL fazt instance |
| `fazt @peer <command>` | Operates on REMOTE peer named `peer` |

This is not just syntax - it's a **philosophical stance**:

1. **Your machine is the primary context** - Fazt runs locally
2. **Remote peers are explicit extensions** - You consciously choose when to operate remotely
3. **Consistency over convenience** - Same commands work locally and remotely

### Why This is Superior

**The flag-based model (current):**
```bash
fazt app deploy ./app --to zyt    # "deploy to" sounds like pushing
fazt app remove myapp --from zyt  # "remove from" is confusing (is zyt source or target?)
fazt app list zyt                 # positional peer (inconsistent with above)
fazt app info myapp --on zyt      # yet another flag
```

**The @peer model (proposed):**
```bash
fazt @zyt app deploy ./app        # deploy ON zyt
fazt @zyt app remove myapp        # remove ON zyt
fazt @zyt app list                # list ON zyt
fazt @zyt app info myapp          # info ON zyt
```

The `@peer` prefix answers one question unambiguously: **WHERE is this operation happening?**

### Analogy: SSH

The `@peer` syntax mirrors SSH's well-understood pattern:

```bash
ssh user@host command             # run command on host
scp file user@host:path           # copy to host

fazt @zyt app list                # list apps on zyt
fazt @zyt server info             # get server info from zyt
```

Users already understand `@` as "at this remote location."

---

## Part 2: Command Categories

### Category A: Single-Peer Operations (Query/Modify)

These commands operate on ONE peer - local by default, remote with `@peer`.

#### Query Operations

```bash
# Local (default)
fazt app list                     # list local apps
fazt app info myapp               # get local app info
fazt app logs myapp               # view local app logs
fazt server status                # local server status

# Remote (explicit)
fazt @zyt app list                # list apps on zyt
fazt @zyt app info myapp          # get app info from zyt
fazt @zyt app logs myapp -f       # follow logs on zyt
fazt @zyt server info             # get server info from zyt
```

#### Modification Operations

```bash
# Local
fazt app remove myapp             # remove from local
fazt app link api myapp           # create local alias
fazt app unlink api               # remove local alias
fazt app reserve admin            # reserve local subdomain

# Remote
fazt @zyt app remove myapp        # remove from zyt
fazt @zyt app link api myapp      # create alias on zyt
fazt @zyt app unlink api          # remove alias on zyt
fazt @zyt app reserve admin       # reserve subdomain on zyt
```

### Category B: Deployment Operations (Source to Target)

Deployment always has a SOURCE (local directory) and TARGET (where to deploy).

```bash
# Deploy to local fazt
fazt app deploy ./my-app                    # deploy ./my-app to local

# Deploy to remote peer
fazt @zyt app deploy ./my-app               # deploy ./my-app to zyt
fazt @zyt app deploy ./my-app --name blog   # deploy with custom name

# Install from URL
fazt app install github:user/repo           # install to local
fazt @zyt app install github:user/repo      # install to zyt
```

**The `@peer` specifies the TARGET, not the source.** The source is always the positional argument.

### Category C: Data Transfer (Two Peers Involved)

Some operations involve TWO locations (source and destination). Here we need additional syntax.

#### Pull: Remote to Local

Pull brings data FROM a remote peer TO local filesystem.

```bash
# Pull from remote to current directory
fazt @zyt app pull myapp                    # pulls from zyt to ./myapp

# Pull to specific directory
fazt @zyt app pull myapp --to ./local-copy  # pulls from zyt to ./local-copy
```

Here `@peer` specifies the SOURCE (where to pull FROM). The `--to` flag is optional local destination.

#### Fork: Within Same Peer or Across Peers

Fork creates a copy of an app.

```bash
# Fork within same peer
fazt app fork myapp --as myapp-v2           # fork local app locally
fazt @zyt app fork myapp --as myapp-v2      # fork app on zyt, stays on zyt

# Fork across peers (explicit second peer)
fazt @zyt app fork myapp --to prod          # fork from zyt TO prod
fazt @local app fork myapp --to zyt         # fork from local TO zyt
```

Here `@peer` is the SOURCE peer, `--to` is the optional DESTINATION peer.

### Category D: Local-Only Operations

Some commands are inherently local - no peer concept.

```bash
fazt app create myapp                       # create local template
fazt app create myapp --template vue        # create with template
fazt app validate ./my-app                  # validate locally

fazt server init                            # initialize local server
fazt server start                           # start local server
fazt server stop                            # stop local server

fazt service install                        # install systemd service
fazt service start                          # start systemd service
```

Using `@peer` with these commands is an error:
```bash
fazt @zyt server init                       # ERROR: cannot init remote server
fazt @zyt app create myapp                  # ERROR: templates are local only
```

### Category E: Peer Management

Peer management commands configure the client's knowledge of remote peers.

```bash
fazt peer add zyt --url https://admin.zyt.app --token xxx
fazt peer list
fazt peer list --format json                # JSON output
fazt peer remove zyt
fazt peer default zyt                       # set default for when @peer is required but omitted
```

These are always local operations (managing local config).

---

## Part 3: Edge Cases

### 3.1 Multiple Peers / All Peers

**Question:** How to list apps across all configured peers?

**Option A: Explicit loop (recommended)**
```bash
fazt app list                               # local only
fazt @zyt app list                          # zyt only
fazt @prod app list                         # prod only
```

**Option B: Special syntax for "all"**
```bash
fazt @all app list                          # list from all peers
fazt @* app list                            # alternative
```

**Recommendation:** Start with Option A (explicit). Users can script aggregation:
```bash
for peer in local zyt prod; do
  echo "=== $peer ==="
  fazt @$peer app list
done
```

The `@all` syntax could be added later if demand exists.

### 3.2 Default Peer Concept

**Question:** Should there be a "default peer" for when `@peer` is omitted?

**Recommendation:** No. Omitting `@peer` means LOCAL. This is the core principle.

However, we can have a **default REMOTE peer** for scripts that explicitly want remote-by-default:

```bash
export FAZT_PEER=zyt                        # environment variable

fazt app list                               # still LOCAL (explicit override)
fazt @_ app list                            # uses $FAZT_PEER (zyt)
fazt @zyt app list                          # explicit zyt
```

The `@_` syntax could mean "use default remote peer from environment."

**Alternative:** Keep it simple. No magic. Local is local, remote requires `@peer`.

### 3.3 Local Peer Alias

For symmetry, `@local` should explicitly mean the local fazt instance:

```bash
fazt app list                               # local (implicit)
fazt @local app list                        # local (explicit) - same result

# Useful in scripts for clarity
fazt @$PEER app list                        # where PEER might be "local" or "zyt"
```

### 3.4 Server Operations

Most server commands are local-only. But some read-only operations make sense remotely:

```bash
# Local only (mutating)
fazt server init                            # local only
fazt server start                           # local only
fazt server set-config --domain x           # local only

# Can be remote (read-only)
fazt server status                          # local status
fazt @zyt server info                       # remote server info
fazt @zyt server version                    # remote version
```

Rule: Only **read-only informational commands** work with `@peer` for server operations.

### 3.5 What About `--to` and `--from` Flags?

With `@peer` as primary, do we still need directional flags?

**Keep for data transfer operations only:**

```bash
# Pull: @peer is source, --to is local destination
fazt @zyt app pull myapp --to ./backup

# Fork across peers: @peer is source, --to is remote destination
fazt @staging app fork myapp --to production
```

**Retire for single-peer operations:**

```bash
# OLD (flag-based)
fazt app deploy ./app --to zyt
fazt app remove myapp --from zyt

# NEW (@peer-based)
fazt @zyt app deploy ./app
fazt @zyt app remove myapp
```

### 3.6 Peer Chaining

**Question:** Can you reference multiple peers in one command?

**Answer:** Only for operations that naturally involve two peers (fork across peers):

```bash
fazt @zyt app fork myapp --to prod          # from zyt to prod
```

You cannot do:
```bash
fazt @zyt @prod app list                    # ERROR: only one @peer allowed
```

---

## Part 4: Complete Command Reference

### App Commands

| Command | Local | Remote | Notes |
|---------|-------|--------|-------|
| `app list` | `fazt app list` | `fazt @peer app list` | Query |
| `app info <app>` | `fazt app info myapp` | `fazt @peer app info myapp` | Query |
| `app logs <app>` | `fazt app logs myapp` | `fazt @peer app logs myapp` | Query, supports `-f` |
| `app deploy <dir>` | `fazt app deploy ./dir` | `fazt @peer app deploy ./dir` | Deploy to target |
| `app install <url>` | `fazt app install url` | `fazt @peer app install url` | Install to target |
| `app remove <app>` | `fazt app remove myapp` | `fazt @peer app remove myapp` | Modify |
| `app create <name>` | `fazt app create myapp` | N/A | Local only |
| `app validate <dir>` | `fazt app validate ./dir` | N/A | Local only |
| `app pull <app>` | N/A | `fazt @peer app pull myapp` | Pull FROM peer |
| `app fork <app>` | `fazt app fork myapp` | `fazt @peer app fork myapp [--to peer2]` | Copy app |
| `app upgrade <app>` | `fazt app upgrade myapp` | `fazt @peer app upgrade myapp` | Upgrade git-sourced |
| `app lineage <app>` | `fazt app lineage myapp` | `fazt @peer app lineage myapp` | Show fork tree |

### Alias Commands

| Command | Local | Remote | Notes |
|---------|-------|--------|-------|
| `app link <sub> <app>` | `fazt app link api myapp` | `fazt @peer app link api myapp` | Create alias |
| `app unlink <sub>` | `fazt app unlink api` | `fazt @peer app unlink api` | Remove alias |
| `app reserve <sub>` | `fazt app reserve admin` | `fazt @peer app reserve admin` | Reserve subdomain |
| `app swap <a> <b>` | `fazt app swap a b` | `fazt @peer app swap a b` | Swap aliases |
| `app split <sub>` | `fazt app split api --ids x:50,y:50` | `fazt @peer app split api --ids x:50,y:50` | Traffic split |

### Server Commands

| Command | Local | Remote | Notes |
|---------|-------|--------|-------|
| `server init` | `fazt server init` | N/A | Local only |
| `server start` | `fazt server start` | N/A | Local only |
| `server stop` | `fazt server stop` | N/A | Local only |
| `server status` | `fazt server status` | N/A | Local only |
| `server info` | `fazt server info` | `fazt @peer server info` | Read-only remote OK |
| `server set-config` | `fazt server set-config` | N/A | Local only |
| `server set-credentials` | `fazt server set-credentials` | N/A | Local only |
| `server create-key` | `fazt server create-key` | N/A | Local only |

### Peer Management

| Command | Description |
|---------|-------------|
| `peer add <name> --url <url> --token <token>` | Register new peer |
| `peer list` | List configured peers |
| `peer list --format json` | List peers in JSON format |
| `peer remove <name>` | Unregister peer |
| `peer status [name]` | Check peer connectivity |
| `peer upgrade [name]` | Check for peer updates |

### Service Commands (Local Only)

| Command | Description |
|---------|-------------|
| `service install` | Install systemd service |
| `service start` | Start systemd service |
| `service stop` | Stop systemd service |
| `service status` | Show systemd service status |
| `service logs` | Follow systemd logs |

---

## Part 5: Benefits Analysis

### 5.1 Visual Clarity

**Flag-based (scattered):**
```bash
fazt app deploy ./my-app --to zyt --name blog --spa
#                        ^^^^^^^^ peer buried in flags
```

**@peer-based (prefix):**
```bash
fazt @zyt app deploy ./my-app --name blog --spa
#    ^^^^ peer is immediately visible
```

The `@peer` prefix is **always in the same position** - right after `fazt`.

### 5.2 Consistent Mental Model

One rule to remember: **`@peer` = "do this operation THERE instead of HERE"**

No need to remember:
- Is it `--to` or `--from` or `--on`?
- Is peer positional or a flag?
- Which commands support which peer flags?

### 5.3 Easier Script Scanning

```bash
#!/bin/bash
# Deploy pipeline

fazt app validate ./app
fazt @staging app deploy ./app
fazt @staging app logs app -f &
sleep 30
fazt @staging app info app

# Easy to see: first command is local, rest are on staging
```

Versus:
```bash
#!/bin/bash
fazt app validate ./app
fazt app deploy ./app --to staging
fazt app logs app --on staging -f &
sleep 30
fazt app info app --on staging

# Peer is scattered across different flag positions
```

### 5.4 Natural Language Alignment

The `@peer` syntax reads naturally:

- `fazt @zyt app list` = "fazt AT zyt, app list"
- `fazt @prod server info` = "fazt AT prod, server info"

This mirrors how we talk about remote operations:
- "Run this at production"
- "Check the logs at staging"

### 5.5 SSH Familiarity

Developers already use `@` syntax daily:

```bash
ssh admin@server.com
scp file.txt admin@server.com:~/
git@github.com:user/repo.git
```

The `@peer` pattern leverages existing mental models.

### 5.6 Reduced Cognitive Load

**Questions users must answer with flag-based:**
1. What's the peer flag for this command?
2. Is it required or optional?
3. Does it go before or after other args?
4. Is there a default peer?

**Questions with @peer-based:**
1. Am I operating locally or remotely?
   - Locally: just run the command
   - Remotely: add `@peer` prefix

---

## Part 6: Comparison to Other CLIs

### 6.1 SSH: user@host

```bash
ssh user@host                     # connect to host
ssh user@host ls -la              # run command on host
```

**Similarity:** The `@` clearly indicates "at this remote location."

**Difference:** SSH puts user before host; fazt puts peer after `@`.

### 6.2 kubectl: --context flag

```bash
kubectl --context=prod get pods
kubectl get pods                  # uses current context
```

**Similarity:** Both specify which cluster/peer to target.

**Difference:** kubectl uses a flag (position-independent), fazt uses prefix (always first).

**Advantage of @peer:** More visible, consistent position.

### 6.3 Docker: -H flag and contexts

```bash
docker -H ssh://user@host ps      # connect to remote docker
docker context use production     # switch default context
docker ps                         # uses current context
```

**Similarity:** Both have per-command targeting and default contexts.

**Difference:** Docker's `-H` is buried in flags; fazt's `@peer` is prominent.

### 6.4 Heroku: git remote naming

```bash
heroku logs --app myapp
heroku logs --remote staging
git push heroku main
```

**Similarity:** Both target specific environments/remotes.

**Difference:** Heroku mixes `--app` and `--remote` flags; fazt uses uniform `@peer`.

### Summary Comparison

| CLI | Remote Targeting | Position | Visibility |
|-----|-----------------|----------|------------|
| SSH | `user@host` | Prefix | High |
| kubectl | `--context=name` | Flag (any position) | Medium |
| Docker | `-H host` or context | Flag or implicit | Low |
| Heroku | `--app` or `--remote` | Flag | Medium |
| **Fazt** | `@peer` | **Prefix (fixed)** | **High** |

Fazt's `@peer` provides the **highest visibility** and **most consistent positioning**.

---

## Part 7: Migration Strategy (COMPLETED)

### Implementation Status

**v0.18.0 (Current)**: Full implementation complete

All `--to`, `--from`, `--on` flags have been removed from app commands. The `@peer` syntax is now the ONLY way to specify remote operations.

**Breaking changes**:
```bash
# Old (REMOVED - no longer works)
fazt app deploy ./app --to zyt
fazt app list zyt
fazt app info myapp --on zyt
fazt app remove myapp --from zyt

# New (CURRENT - only way)
fazt @zyt app deploy ./app
fazt @zyt app list
fazt @zyt app info myapp
fazt @zyt app remove myapp
```

### Command Renames

**`fazt remote` → `fazt peer`**

The `remote` command has been renamed to `peer` for consistency:

```bash
# Old
fazt remote list
fazt remote add zyt --url ... --token ...

# New
fazt peer list
fazt peer add zyt --url ... --token ...
```

### New Features

**SQL query command**:
```bash
fazt sql "SELECT * FROM apps"                 # Local
fazt @zyt sql "SELECT * FROM apps"            # Remote
fazt sql "SELECT * FROM apps" --format json   # JSON output
```

**Format flag for peer list**:
```bash
fazt peer list --format markdown              # Default
fazt peer list --format json                  # JSON output
```

---

## Part 8: Code Implications

### 8.1 Parser Changes

Current `handlePeerCommand` already exists. Needs enhancement:

```go
func main() {
    // ... existing code ...

    // @peer prefix handling (already exists)
    if strings.HasPrefix(command, "@") {
        handlePeerCommand(command[1:], os.Args[2:])
        return
    }

    // Local command handling (no @peer)
    switch command {
    case "app":
        handleAppCommandLocal(os.Args[2:])  // local-only operations
    // ...
    }
}
```

### 8.2 Unified Command Router

Create unified handler that works for both local and remote:

```go
type CommandContext struct {
    Peer      *remote.Peer  // nil for local
    IsLocal   bool
    Args      []string
}

func handleAppCommand(ctx CommandContext) {
    switch ctx.Args[0] {
    case "list":
        if ctx.IsLocal {
            handleLocalAppList(ctx.Args[1:])
        } else {
            handleRemoteAppList(ctx.Peer, ctx.Args[1:])
        }
    // ...
    }
}
```

### 8.3 Command Gateway Enhancement

The existing `/api/cmd` endpoint already supports remote execution. Ensure parity:

```go
// Remote execution via HTTP POST /api/cmd
type CmdRequest struct {
    Command string   `json:"command"`
    Args    []string `json:"args"`
}

// All "app" subcommands should be routable through this gateway
```

### 8.4 Flag Removal

Remove peer-targeting flags from individual commands:

```go
// Before
flags.String("to", "", "Target peer name")
flags.String("from", "", "Source peer name")
flags.String("on", "", "Context peer")

// After
// These flags no longer exist - peer comes from @peer prefix
```

### 8.5 Help Text Updates

Update all help text to show `@peer` as primary:

```go
func printAppHelp() {
    fmt.Println(`Fazt CLI - App Management

USAGE:
  fazt app <command> [args]          Local operations
  fazt @<peer> app <command> [args]  Remote operations

LOCAL COMMANDS:
  create <name>      Create app template
  validate <dir>     Validate app structure

LOCAL OR REMOTE COMMANDS:
  list               List apps
  info <app>         Show app details
  deploy <dir>       Deploy directory
  remove <app>       Remove app
  logs <app>         View logs (-f to follow)

EXAMPLES:
  fazt app list                      # List local apps
  fazt @zyt app list                 # List apps on zyt
  fazt @zyt app deploy ./my-app      # Deploy to zyt
  fazt @zyt app logs myapp -f        # Follow logs on zyt
`)
}
```

---

## Part 9: Is This Actually Superior?

### Arguments FOR @peer-Primary

1. **Conceptual clarity** - One syntax, one rule, one position
2. **Visual scanning** - Peer is always visible at the start
3. **SSH familiarity** - Developers already know `@` means "at"
4. **Local-first philosophy** - Aligns with sovereign compute vision
5. **Reduced flag proliferation** - No more `--to`/`--from`/`--on` confusion

### Arguments AGAINST @peer-Primary

1. **More typing for remote** - `fazt @zyt app list` vs `fazt app list zyt`
2. **Breaking change** - Existing scripts need updates
3. **Discoverability** - Users might not know about `@peer` syntax
4. **Loss of positional shortcuts** - Can't do `fazt app list zyt` anymore

### Counter-Arguments

1. **More typing:** Only 2 extra characters (`@` and space). Clarity > brevity.
2. **Breaking change:** Phased migration over multiple releases.
3. **Discoverability:** Put it in every help message, every error message.
4. **Positional shortcuts:** Consistency > shortcuts.

### Verdict: Yes, This is Superior

The `@peer` model is superior because it **reduces cognitive load** for the primary use case (local operations) while making remote operations **explicitly visible**.

The flag-based model requires users to remember different flags (`--to`, `--from`, `--on`) for different commands. The `@peer` model has ONE rule: prefix with `@peer` for remote.

This aligns perfectly with Fazt's **sovereign compute** philosophy: your local instance is primary, remote peers are conscious extensions.

---

## Part 10: Implementation Checklist

### Status: COMPLETED ✓

All phases have been completed in v0.18.0:

### Phase 1: Foundation ✓
- [x] Audit all commands for local vs remote capability
- [x] Ensure `/api/cmd` gateway handles all app subcommands
- [x] Add `@local` as explicit alias for local operations
- [x] Update `handlePeerCommand` to route all supported commands

### Phase 2: Documentation ✓
- [x] Update all help text to show `@peer` syntax first
- [x] Create migration guide in knowledge-base
- [x] Update CLAUDE.md with new patterns
- [x] Add examples to error messages

### Phase 3: Deprecation ✓
- [x] Removed all flag-based peer targeting (no deprecation phase)
- [x] Clean break in v0.18.0

### Phase 4: Cleanup ✓
- [x] Remove `--to`, `--from`, `--on` flags
- [x] Simplify command handlers
- [x] Update tests
- [x] Rename `remote` command to `peer`
- [x] Add `sql` command with `--format` flag
- [x] Add `--format` flag to `peer list`

---

## Appendix A: Quick Reference Card

```
LOCAL OPERATIONS (no prefix)
============================
fazt app list                 # list apps
fazt app info myapp           # app details
fazt app deploy ./dir         # deploy
fazt app remove myapp         # remove
fazt app create myapp         # create template
fazt server start             # start server

REMOTE OPERATIONS (@peer prefix)
================================
fazt @zyt app list            # list apps on zyt
fazt @zyt app info myapp      # app details from zyt
fazt @zyt app deploy ./dir    # deploy to zyt
fazt @zyt app remove myapp    # remove from zyt
fazt @zyt server info         # server info from zyt

DATA TRANSFER (source in @peer, dest in --to)
=============================================
fazt @zyt app pull myapp      # pull from zyt to ./myapp
fazt @zyt app pull myapp --to ./local
fazt @zyt app fork myapp --to prod  # fork from zyt to prod

LOCAL-ONLY (no remote equivalent)
=================================
fazt server init              # initialize
fazt server start/stop        # control
fazt service install          # systemd
fazt app create               # templates
fazt app validate             # validation

PEER MANAGEMENT
===============
fazt peer add zyt --url ... --token ...
fazt peer list
fazt peer list --format json
fazt peer remove zyt

SQL QUERIES
===========
fazt sql "SELECT * FROM apps"
fazt @zyt sql "SELECT * FROM apps"
fazt sql "SELECT * FROM apps" --format json
```

---

## Appendix B: Error Messages

### Missing @peer for Remote Operation

```
$ fazt app deploy ./app
Error: No local fazt server running.

To deploy to a remote peer:
  fazt @<peer> app deploy ./app

Available peers:
  - zyt (https://admin.zyt.app)
  - local (http://192.168.64.3:8080)

To start local server:
  fazt server start
```

### Unknown Peer

```
$ fazt @unknown app list
Error: Peer 'unknown' not found.

Configured peers:
  - zyt
  - local

To add a peer:
  fazt remote add <name> --url <url> --token <token>
```

### Command Not Supported Remotely

```
$ fazt @zyt server init
Error: 'server init' cannot be run remotely.

This command must be run on the server itself.
```

### Deprecated Syntax

```
$ fazt app deploy ./app --to zyt
Deploying to zyt...

[DEPRECATED] The --to flag is deprecated.
Use @peer syntax instead:
  fazt @zyt app deploy ./app

This flag will be removed in v0.20.0.

Done.
```

---

## Conclusion

The `@peer`-primary design transforms Fazt's CLI from a hodgepodge of flags to a clean, consistent interface. By making local operations the default and requiring explicit `@peer` for remote operations, we:

1. Align with the **sovereign compute** philosophy
2. Eliminate **flag confusion** (`--to` vs `--from` vs `--on`)
3. Provide **visual clarity** with fixed prefix position
4. Leverage **existing mental models** from SSH
5. Create a **memorable, discoverable** pattern

The migration path is gradual: support both syntaxes, show deprecation notices, eventually remove the flag-based peer targeting.

This is not just a syntax change - it's a statement about Fazt's identity as a **local-first platform** with conscious remote extensions.
