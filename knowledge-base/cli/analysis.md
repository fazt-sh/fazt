# Fazt CLI Consistency Analysis

**Version**: 0.18.0
**Date**: 2026-02-01
**Updated**: 2026-02-02
**Status**: IMPLEMENTED in v0.18.0

---

## Executive Summary

The Fazt CLI has evolved organically, resulting in **five distinct patterns for peer targeting**, **three different identifier passing styles**, and **inconsistent verb usage**. These inconsistencies create cognitive load for users who must remember special rules for each command.

### Key Findings

1. **Peer Targeting Chaos**: Commands use `--to`, `--from`, `--on`, positional args, and `@peer` prefix interchangeably
2. **Semantic Confusion**: `--from` means "source peer" in `pull` but "target peer" in `remove`
3. **Missing Defaults**: `deploy` requires `--to` even for single-peer setups; no local-first workflow
4. **Identifier Ambiguity**: Some commands accept positional identifiers, others require `--alias`/`--id` flags
5. **Verb Inconsistency**: `remove` vs `unlink` vs `delete`; `add` vs `create` vs `install`

### Recommended Changes

1. Standardize on **semantic peer flags**: `--to` (destination), `--from` (source), peer as positional for context
2. **Default to default peer** when not specified (with single-peer auto-default)
3. **Unify identifier passing**: positional-first with optional disambiguation flags
4. **Consistent verbs**: `create`, `remove`, `list`, `show`

---

## Current State Analysis

### Command Inventory

```
fazt app list [peer]
fazt app info [identifier] [--alias|--id] [--on peer]
fazt app deploy <dir> --to <peer> [--name]
fazt app install <url> [--to peer]
fazt app remove [identifier] [--alias|--id] [--from peer] [--with-forks]
fazt app link <subdomain> --id <app_id> [--to peer]
fazt app unlink <subdomain> [--from peer]
fazt app reserve <subdomain> [--on peer]
fazt app fork [--alias|--id] [--as name] [--to peer]
fazt app lineage [--alias|--id] [--on peer]
fazt app logs <app> [--on peer] [-f]
fazt app upgrade <app> [--from peer]
fazt app pull <app> [--from peer] [--to dir]
fazt app swap <a1> <a2> [--on peer]
fazt app split <subdomain> --ids <pairs> [--on peer]
fazt app create <name> [--template]
fazt app validate <dir>

fazt peer add <name> --url <url> --token <token>
fazt peer list
fazt peer remove <name>
fazt peer default <name>
fazt peer status [name]
fazt peer upgrade [name]

fazt server init [--username] [--password] [--domain] [--db]
fazt server start [--domain] [--port] [--db]
fazt server status
fazt server set-credentials [--username] [--password]
fazt server set-config [--domain] [--port] [--env]
fazt server create-key [--name]
fazt server reset-admin

fazt service install [--domain] [--email] [--https]
fazt service start
fazt service stop
fazt service status
fazt service logs

fazt @<peer> app <command>
fazt @<peer> server info
```

### Peer Targeting Patterns (Current)

| Pattern | Commands Using It | Example |
|---------|-------------------|---------|
| Positional arg | `list`, `info` | `fazt app list zyt` |
| `--to` flag | `deploy`, `install`, `link`, `fork` | `fazt app deploy . --to zyt` |
| `--from` flag | `remove`, `unlink`, `pull`, `upgrade` | `fazt app remove x --from zyt` |
| `--on` flag | `info`, `reserve`, `logs`, `swap`, `split`, `lineage` | `fazt app logs x --on zyt` |
| `@peer` prefix | All remote commands | `fazt @zyt app list` |

**Problem**: User must remember which flag to use for each command. The semantic distinction is unclear.

### Identifier Passing Patterns (Current)

| Pattern | Commands Using It | Example |
|---------|-------------------|---------|
| Positional only | `list`, `logs`, `upgrade`, `pull` | `fazt app logs myapp` |
| Positional OR flags | `info`, `remove` | `fazt app info myapp` OR `--alias myapp` |
| Flags only | `fork`, `lineage` | `fazt app fork --alias myapp` |
| Mixed (subdomain + id flag) | `link`, `split` | `fazt app link sub --id app_123` |

**Problem**: No consistent rule for when to use positional vs flags.

### Verb Usage Analysis

| Action | Current Verbs | Ideal Verb |
|--------|--------------|------------|
| View list | `list` | `list` |
| View details | `info` | `show` or `info` |
| Create new | `create` (local), `install` (remote), `deploy` (to peer), `add` (remote peer) | `create` |
| Delete | `remove` (app/peer), `unlink` (alias) | `remove` |
| Modify | `link`, `reserve`, `swap`, `split`, `set-credentials`, `set-config` | context-specific |

---

## Problems Identified

### Problem 1: Peer Flag Semantic Ambiguity

**Current State**: `--from` is used inconsistently:
- `app pull` uses `--from` as SOURCE peer (makes sense)
- `app remove` uses `--from` as TARGET peer (confusing - you remove FROM the peer)
- `app upgrade` uses `--from` as TARGET peer (confusing - you upgrade ON the peer)
- `app unlink` uses `--from` as TARGET peer (removal direction)

**User Mental Model**: Users expect:
- `--from` = where data comes FROM (source)
- `--to` = where data goes TO (destination)
- Operations on a peer = just specify the peer

**Example Confusion**:
```bash
# These read identically but mean different things:
fazt app pull myapp --from zyt     # Pull FROM zyt (source)
fazt app remove myapp --from zyt   # Remove FROM zyt (target?? reads like source)
```

**Git/Docker Comparison**:
```bash
# Git uses clear semantics
git push origin main              # Push TO origin
git pull origin main              # Pull FROM origin

# Docker targets are always destinations
docker push registry/image        # Push TO registry
docker pull registry/image        # Pull FROM registry
```

### Problem 2: Deploy Requires --to Even for Single Peer

**Current State**:
```bash
fazt app deploy ./my-app --to zyt
```

With only one configured peer, user still must specify `--to`.

**Expected Behavior**:
```bash
# If only one peer exists, this should work:
fazt app deploy ./my-app

# Or if a default peer is set:
fazt app deploy ./my-app
```

**Git Comparison**:
```bash
# Git defaults to origin/main
git push                 # Works when upstream is set
git push origin main     # Explicit is optional
```

### Problem 3: Identifier Ambiguity

**Current State**: Some commands require `--alias`/`--id`, others accept positional.

```bash
# These are equivalent:
fazt app info myapp --on zyt
fazt app info --alias myapp --on zyt
fazt app info --id app_123 --on zyt

# But fork REQUIRES flags:
fazt app fork myapp            # ERROR: --alias or --id required
fazt app fork --alias myapp    # Works
```

**Why This Matters**: Users build mental models from early commands. If `info myapp` works, they expect `fork myapp` to work.

**Kubectl Comparison**:
```bash
# kubectl is consistently positional-first:
kubectl get pod mypod
kubectl delete pod mypod
kubectl describe pod mypod
```

### Problem 4: Multiple Ways to Express Same Intent

**Current State**: To list apps on a peer:
```bash
fazt app list zyt
fazt @zyt app list
```

Both work. Which is preferred? Documentation doesn't clarify.

**Problem**: Two syntaxes = two things to learn, maintain, document.

**Recommendation**: `@peer` prefix should be the ADVANCED syntax for remote execution of local commands. Normal commands should use positional/flags.

### Problem 5: Inconsistent Required vs Optional Flags

| Command | Peer Flag | Required? | Behavior if Omitted |
|---------|-----------|-----------|---------------------|
| `deploy` | `--to` | Effectively yes | Error or uses default |
| `install` | `--to` | No | Uses default |
| `list` | positional | No | Uses default |
| `link` | `--to` | No | Uses default |
| `remove` | `--from` | No | Uses default |

**Problem**: Inconsistent requirement levels confuse users about what's mandatory.

### Problem 6: Verb Consistency Issues

**Create/Add/Install**:
```bash
fazt app create myapp          # Creates LOCAL template
fazt app install github:...    # Installs FROM URL to peer
fazt app deploy ./dir          # Deploys FROM local TO peer
fazt peer add zyt            # Adds peer config
```

Four different verbs for "bring something into existence."

**Remove/Delete/Unlink**:
```bash
fazt app remove myapp          # Removes app
fazt app unlink subdomain      # Removes alias
fazt peer remove zyt         # Removes peer
```

`remove` vs `unlink` - why different verbs for similar actions?

---

## Design Principles

Based on analysis of git, docker, kubectl, and POLS, here are the principles for a consistent CLI:

### Principle 1: Semantic Clarity

**Peer flags should be directional**:
- `--to` = destination (deploy to, fork to, push to)
- `--from` = source (pull from, clone from)
- Positional peer OR omitted = context/target (list, info, logs, status)

**Rule**: If data flows IN a direction, use `--to`/`--from`. If you're just specifying WHERE, use positional or let it default.

### Principle 2: Progressive Disclosure

**Simple cases should be simple**:
```bash
fazt app list                  # Works with default peer
fazt app deploy ./app          # Works with default peer
fazt app info myapp            # Works with default peer
```

**Complex cases remain possible**:
```bash
fazt app list zyt
fazt app deploy ./app --to production
fazt app info --id app_123 --on staging
```

### Principle 3: Consistent Identifier Handling

**Pattern**: `<command> [identifier] [--qualifier] [--peer-flag peer]`

- Positional identifier FIRST (when single identifier)
- `--alias`/`--id` flags for disambiguation when needed
- Never REQUIRE flags when positional is unambiguous

**Examples**:
```bash
fazt app info myapp             # Resolved as alias first, then id
fazt app info --id app_123      # Explicit: this is an ID
fazt app info --alias myapp     # Explicit: this is an alias
```

### Principle 4: Verb Consistency

| Action | Verb | Used For |
|--------|------|----------|
| Create something new | `create` | `app create`, `remote create` (alias: `add`) |
| Remove something | `remove` | `app remove`, `alias remove` (replaces `unlink`) |
| View list | `list` | `app list`, `remote list`, `alias list` |
| View details | `show` or `info` | `app show` (or keep `info`) |
| Modify | context-specific | `app link`, `app reserve` |

### Principle 5: @peer as Power Feature

`@peer` prefix is for remote execution - running local commands on remote peer:
```bash
fazt @production server info    # Get server info from production
fazt @staging app list          # List apps on staging
```

It should NOT be the primary way to target peers. Standard commands should use:
```bash
fazt app list production        # Positional peer
fazt app deploy ./app --to production   # Directional flag
```

---

## Proposed Design

### Peer Targeting Rules

| Operation Type | Pattern | Examples |
|----------------|---------|----------|
| **Data transfer (outbound)** | `--to <peer>` | `deploy`, `fork`, `push` |
| **Data transfer (inbound)** | `--from <peer>` | `pull`, `clone` |
| **Query/Status/Modify** | `<peer>` positional or omit for default | `list`, `info`, `logs`, `remove`, `link` |

**Rationale**:
- `--to` and `--from` communicate DATA DIRECTION
- Context commands (query, status, modify) just need to know WHICH peer

### Revised Command Structure

```
# Apps - Query/Status (peer as optional positional)
fazt app list [peer]
fazt app show <identifier> [peer] [--alias|--id]    # renamed from info
fazt app logs <identifier> [peer] [-f]

# Apps - Data Transfer (directional flags)
fazt app deploy <dir> [--to peer] [--name name]
fazt app install <url> [--to peer] [--name name]
fazt app pull <identifier> [--from peer] [--to dir]
fazt app fork <identifier> [--to peer] [--as name]

# Apps - Modify (peer as optional positional)
fazt app remove <identifier> [peer] [--with-forks]
fazt app create <name> [--template type]           # local only

# Aliases (peer as optional positional after subdomain)
fazt alias list [peer]                              # was: app list --aliases
fazt alias link <subdomain> <identifier> [peer]    # was: app link
fazt alias remove <subdomain> [peer]               # was: app unlink
fazt alias reserve <subdomain> [peer]
fazt alias swap <a1> <a2> [peer]
fazt alias split <subdomain> --targets <pairs> [peer]

# Lineage (peer as optional positional)
fazt app lineage <identifier> [peer]               # was: --alias/--id required

# Remote Peers
fazt peer list
fazt peer add <name> --url <url> --token <token>
fazt peer remove <name>
fazt peer default <name>
fazt peer status [name]
fazt peer upgrade [name]

# Local Server
fazt server init [--username] [--password] [--domain]
fazt server start [--port] [--domain]
fazt server status
fazt server config [--key] [--value]               # replaces set-config/set-credentials

# Remote Execution (power feature)
fazt @<peer> <any-command>                         # executes locally-valid command on remote
```

### Default Peer Behavior

```
1. If peer specified → use it
2. If single peer configured → auto-use it
3. If default peer set → use it
4. If multiple peers, no default → error with helpful message
```

**Implementation**:
```go
// ResolvePeer enhanced logic
func ResolvePeer(db *sql.DB, name string) (*Peer, error) {
    if name != "" {
        return GetPeer(db, name)
    }

    peers, _ := ListPeers(db)
    if len(peers) == 0 {
        return nil, ErrNoPeers
    }
    if len(peers) == 1 {
        return &peers[0], nil  // Auto-use single peer
    }

    defaultPeer := getDefaultPeer(db)
    if defaultPeer != "" {
        return GetPeer(db, defaultPeer)
    }

    return nil, ErrNoDefaultPeer
}
```

### Identifier Resolution

**Pattern**: `[identifier]` resolves as:
1. Try as alias first
2. Try as app_id if starts with `app_`
3. Fail with helpful error

**Disambiguation**: Use `--alias` or `--id` when you need explicit behavior:
```bash
fazt app show app_foo           # Ambiguous? Might be alias "app_foo"
fazt app show --id app_foo      # Definitely the ID
```

---

## Migration Strategy

### Phase 1: Add Aliases (Non-Breaking)

Add new command forms as aliases:

| New Command | Alias For | Notes |
|-------------|-----------|-------|
| `fazt app show` | `fazt app info` | Preferred new name |
| `fazt alias list` | `fazt app list --aliases` | New namespace |
| `fazt alias link` | `fazt app link` | Move to alias namespace |
| `fazt alias remove` | `fazt app unlink` | Consistent verb |

### Phase 2: Normalize Flags (Potentially Breaking)

Change flag semantics:

| Old | New | Breaking? |
|-----|-----|-----------|
| `app remove --from` | `app remove [peer]` | Soft break - old works with deprecation warning |
| `app upgrade --from` | `app upgrade [peer]` | Soft break |
| `app fork --alias x` | `app fork x` | Soft break - old works |
| `app lineage --alias x` | `app lineage x` | Soft break |

### Phase 3: Default Peer Enhancement (Non-Breaking)

Enhance default peer behavior to auto-select single peer.

### Phase 4: Deprecation Warnings

Add warnings for deprecated patterns:
```
Warning: --from is deprecated for 'app remove'. Use positional peer instead.
  Old: fazt app remove myapp --from zyt
  New: fazt app remove myapp zyt
```

### Phase 5: Remove Deprecated Patterns

After sufficient deprecation period (2-3 releases), remove old patterns.

---

## Examples: Before and After

### Deploy App

**Before**:
```bash
# Must specify --to even with single peer
fazt app deploy ./my-app --to local
```

**After**:
```bash
# Auto-uses single/default peer
fazt app deploy ./my-app

# Or explicit when needed
fazt app deploy ./my-app --to production
```

### Get App Info

**Before**:
```bash
fazt app info myapp --on zyt
fazt app info --alias myapp --on zyt
fazt app info --id app_123 --on zyt
```

**After**:
```bash
fazt app show myapp zyt              # Positional peer
fazt app show myapp                  # Uses default peer
fazt app show --id app_123 zyt       # Explicit ID when needed
```

### Fork an App

**Before**:
```bash
fazt app fork --alias myapp --as myapp-v2 --to zyt
```

**After**:
```bash
fazt app fork myapp --as myapp-v2 --to zyt
fazt app fork myapp --as myapp-v2            # Uses default peer
```

### Remove an App

**Before**:
```bash
fazt app remove myapp --from zyt
```

**After**:
```bash
fazt app remove myapp zyt
fazt app remove myapp                        # Uses default peer
```

### Manage Aliases

**Before**:
```bash
fazt app list zyt --aliases
fazt app link subdomain --id app_123 --to zyt
fazt app unlink subdomain --from zyt
```

**After**:
```bash
fazt alias list zyt
fazt alias link subdomain app_123 zyt        # Both identifiers positional
fazt alias remove subdomain zyt
```

### Pull App Files

**Before**:
```bash
fazt app pull myapp --from zyt --to ./local
```

**After** (unchanged - directional makes sense here):
```bash
fazt app pull myapp --from zyt --to ./local
fazt app pull myapp --to ./local              # Uses default peer as source
```

---

## Implementation Checklist

- [ ] Add `fazt app show` as alias for `fazt app info`
- [ ] Create `fazt alias` namespace with `list`, `link`, `remove`, `reserve`, `swap`, `split`
- [ ] Make positional identifier work for `fork` and `lineage`
- [ ] Change `app remove` and `app upgrade` to use positional peer
- [ ] Enhance `ResolvePeer` to auto-select single peer
- [ ] Add deprecation warnings for old flag patterns
- [ ] Update help text to prefer new patterns
- [ ] Update documentation

---

## Appendix: CLI Design References

### Git
- Clear directional verbs: push, pull, fetch
- Remote name as positional: `git push origin main`
- Sensible defaults: `git push` pushes to upstream

### Docker
- Resource as positional: `docker rm container_id`
- Registry as part of image name: `docker push registry/image`
- Subcommands for namespaces: `docker container ls`, `docker image ls`

### Kubectl
- Resource-first pattern: `kubectl get pod mypod`
- Namespace as flag: `-n kube-system`
- Consistent verbs: `get`, `create`, `delete`, `describe`

### Fly.io
- App context from directory: runs in app dir, uses fly.toml
- Region as flag: `--region lax`
- Deploy is implicit destination: `fly deploy`

---

## Conclusion

The Fazt CLI needs a deliberate refactoring to achieve consistency. The key changes are:

1. **Semantic peer flags**: `--to` (destination), `--from` (source), positional (context)
2. **Smart defaults**: Auto-select single/default peer
3. **Positional-first identifiers**: With `--alias`/`--id` for disambiguation
4. **Alias namespace**: Move subdomain operations to `fazt alias`
5. **Consistent verbs**: `show` instead of `info`, `remove` consistently

These changes will reduce cognitive load, make the CLI more intuitive, and align with user expectations from other well-designed CLIs.
