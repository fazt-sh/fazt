# OS Nomenclature

## Summary

Adopt 1:1 naming conventions with Operating System kernels. This eliminates
mental translation between "app logic" and "system logic" and leverages
50 years of systems engineering vocabulary.

## Rationale

### Why OS Naming?

1. **Universal Mental Model**: Developers already understand OS concepts
2. **AI Compatibility**: LLMs are trained on OS documentation
3. **Consistency**: File paths match CLI commands match mental models
4. **Clarity**: No ambiguity about where code belongs

### Example Command

```bash
# Clear and intuitive
fazt proc start
fazt net route add blog.example.com my-blog
fazt fs ls /apps/my-blog/
fazt storage backup
```

## Component Mapping

### pkg/kernel/ Structure

| Directory   | OS Concept     | Responsibility                      |
| ----------- | -------------- | ----------------------------------- |
| `proc/`     | Process        | Binary lifecycle, systemd, upgrades |
| `fs/`       | File System    | VFS, blob storage, file serving     |
| `net/`      | Network        | SSL, routing, domain management     |
| `storage/`  | Block Storage  | SQLite engine, EDD migrations       |
| `security/` | Access Control | JWT, auth, permissions              |
| `driver/`   | Device Drivers | Litestream, external integrations   |
| `syscall/`  | System Calls   | API bridge between apps and kernel  |

### Syscall Pattern

Applications never touch kernel packages directly. They issue "syscalls"
through a unified interface:

```go
// Bad: App imports kernel directly
import "fazt/pkg/kernel/fs"
vfs.Get(path)

// Good: App uses syscall interface
import "fazt/pkg/kernel/syscall"
kernel.FS.Read(path)
```

### Why Syscall Matters

1. **Isolation**: Apps can't bypass kernel controls
2. **Versioning**: Kernel internals can change without breaking apps
3. **Auditing**: All app actions go through a single layer
4. **Testing**: Mock the syscall layer, test the app

## CLI Restructure

### Command Groups

| Group      | Commands                              | Kernel Module  |
| ---------- | ------------------------------------- | -------------- |
| `proc`     | start, stop, restart, upgrade, status | `proc/`        |
| `fs`       | ls, mount, unmount, cat, rm           | `fs/`          |
| `net`      | route add/rm, domain map, vpn         | `net/`         |
| `storage`  | migrate, backup, restore, gc          | `storage/`     |
| `security` | init, sign, verify, totp              | `security/`    |
| `app`      | deploy, install, remove, list         | (uses syscall) |

### Migration from v0.7

| v0.7 Command         | v0.8 Command          |
| -------------------- | --------------------- |
| `fazt server start`  | `fazt proc start`     |
| `fazt server init`   | `fazt proc init`      |
| `fazt server status` | `fazt proc status`    |
| `fazt deploy`        | `fazt app deploy`     |
| `fazt backup create` | `fazt storage backup` |

## Internal API

### The Kernel Object

```go
type Kernel struct {
    Proc     *proc.Manager
    FS       *fs.VFS
    Net      *net.Router
    Storage  *storage.Engine
    Security *security.Guard
}

// Usage in handlers
func (h *Handler) ServeHTTP(w, r *http.Request) {
    file, err := h.kernel.FS.Read(r.URL.Path)
    if err != nil {
        h.kernel.Security.AuditLog("file_not_found", r)
    }
}
```

### Benefits for AI Agents

Agents respond better to rigid, standard system domains:

```
Human: "Add a field to storage following EDD rules"
Agent: Opens pkg/kernel/storage/, follows EDD patterns