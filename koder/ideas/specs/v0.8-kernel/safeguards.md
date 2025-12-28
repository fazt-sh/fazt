# Safeguards

## Summary

Fazt must survive on a $6/month VPS (1GB RAM, 1 vCPU). This document defines
the resource constraints and circuit breakers that prevent the system from
crashing under load.

## Rationale

### The Problem

Most platforms crash under load:
- Memory exhaustion kills the process
- Disk fills up, corrupts the database
- Network floods exhaust file descriptors

### The Philosophy

**Survival over Service**: When resources are scarce, Fazt prioritizes staying
alive over serving requests. A degraded system is better than a dead system.

## Resource Limits

### RAM Limits

| Resource | Limit | Rationale |
|----------|-------|-----------|
| VFS Cache | 128MB | LRU eviction |
| Large Files | 5MB | Stream from disk, don't cache |
| JS Runtime | 64MB per invocation | Terminate if exceeded |
| Total Heap | 512MB | Leave room for OS |

### Disk Limits

| Resource | Limit | Rationale |
|----------|-------|-----------|
| DB Size | 80% of partition | Reject writes beyond |
| Single Upload | 100MB | Prevent disk flood |
| VFS Total | Configurable | Owner sets limit |

### Network Limits

| Resource | Limit | Rationale |
|----------|-------|-----------|
| Concurrent Connections | 1000 | Prevent FD exhaustion |
| Request Body | 10MB | Prevent memory flood |
| Response Timeout | 30s | Prevent hung connections |

## Circuit Breakers

### The Cockpit Rule

Admin Dashboard and Auth must **always work**. They are hydrated at boot and
served from pinned memory:

```go
type CockpitCache struct {
    AdminHTML  []byte  // Pre-loaded at startup
    LoginHTML  []byte  // Pre-loaded at startup
    // Never evicted, never stale
}
```

If VFS cache is full, user sites degrade. Cockpit never degrades.

### Memory Circuit Breaker

```go
func (k *Kernel) MemoryGuard() {
    if runtime.MemStats.HeapAlloc > k.Config.MaxHeap {
        // 1. Flush analytics buffer to disk
        k.Analytics.FlushNow()

        // 2. Clear VFS cache
        k.FS.ClearCache()

        // 3. Force GC
        runtime.GC()

        // 4. If still over, reject new requests
        if stillOver {
            k.Net.SetMode(MODE_DEGRADED)
        }
    }
}
```

### Disk Circuit Breaker

```go
func (k *Kernel) DiskGuard() {
    usage := k.Storage.DiskUsage()
    if usage > 0.80 {
        // Reject all writes except system
        k.Storage.SetMode(MODE_READ_ONLY)
        k.Notify("Disk at 80%, writes disabled")
    }
}
```

### Request Circuit Breaker

```go
func (k *Kernel) RequestGuard(r *http.Request) bool {
    app := k.Apps.FromRequest(r)

    // Check egress quota
    if app.EgressUsed > app.EgressLimit {
        return false  // 429 Too Many Requests
    }

    // Check concurrent connections
    if k.Net.ActiveConnections() > MAX_CONNECTIONS {
        return false  // 503 Service Unavailable
    }

    return true
}
```

## Degradation Modes

| Mode | Behavior |
|------|----------|
| **Normal** | Full functionality |
| **Degraded** | VFS cache disabled, stream from disk |
| **Read-Only** | No writes, analytics dropped |
| **Cockpit-Only** | Only admin + auth work |
| **Emergency** | Graceful shutdown initiated |

## Monitoring Endpoints

```
GET /api/system/health    # Overall status
GET /api/system/limits    # Current thresholds
GET /api/system/cache     # VFS cache stats
GET /api/system/db        # Database stats
```

## Configuration

```json
{
  "limits": {
    "max_heap_mb": 512,
    "max_vfs_cache_mb": 128,
    "max_upload_mb": 100,
    "max_connections": 1000,
    "disk_threshold": 0.80
  }
}
```

## Implementation Priority

1. **Memory Guard** - Most common failure mode
2. **Disk Guard** - Protects database integrity
3. **Cockpit Pinning** - Ensures recoverability
4. **Request Guard** - Prevents runaway apps
