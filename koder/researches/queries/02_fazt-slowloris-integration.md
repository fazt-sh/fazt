# Research Query: Integrate Slowloris Protection into Fazt

## Context

You have access to the fazt source code. This query assumes you've found a
solution from the first research query (01_go-slowloris-protection.md) and now
need to integrate it.

## Relevant Source Files

### Server Startup
**`cmd/server/main.go`** (lines 2929-3005)
- HTTP server creation at line 2930
- CertMagic HTTPS at line 3005: `certmagic.HTTPS([]string{cfgDomain}, handler)`
- Standard HTTP at line 3000: `srv.ListenAndServe()`

```go
// Current server creation (line 2929-2937)
srv := &http.Server{
    Addr:              ":" + cfg.Server.Port,
    Handler:           handler,
    ReadHeaderTimeout: 5 * time.Second,
    ReadTimeout:       10 * time.Second,
    WriteTimeout:      30 * time.Second,
    IdleTimeout:       60 * time.Second,
}
```

### Middleware Chain
**`cmd/server/main.go`** (lines 2913-2927)
```go
// Current middleware order
handler := connLimiter.Middleware(
    globalRateLimiter.Middleware(
        middleware.RequestTracing(
            loggingMiddleware(
                middleware.BodySizeLimit(...)(
                    middleware.SecurityHeaders(
                        corsMiddleware(
                            recoveryMiddleware(rootHandler),
                        ),
                    ),
                ),
            ),
        ),
    ),
)
```

### Existing Rate Limiting
**`internal/middleware/ratelimit.go`**
- `RateLimiter` - per-IP token bucket (runs after headers parsed)
- `ConnectionLimiter` - per-IP connection count (runs after headers parsed)
- Both are **too late** for slowloris protection

### CertMagic Integration
**`internal/database/certmagic.go`**
- SQLite-backed certificate storage
- Used for automatic HTTPS

## Integration Questions

1. **Where should the `net.Listener` wrapper be inserted?**
   - Before `srv.Serve(listener)`?
   - Before `certmagic.HTTPS()`?
   - How does it interact with TLS?

2. **CertMagic complication**
   - `certmagic.HTTPS()` creates its own listeners internally
   - How do we wrap a listener we don't control?
   - Can we use `certmagic.Listen()` instead and wrap that?

3. **Dual-mode operation**
   - HTTP mode: `srv.ListenAndServe()` (line 3000)
   - HTTPS mode: `certmagic.HTTPS()` (line 3005)
   - Solution must work for both

4. **Configuration**
   - Where should limits be configured? (`internal/config/config.go`)
   - Should limits be in the database? (like other settings)
   - Sensible defaults?

## Desired Implementation

Show me how to modify `cmd/server/main.go` to add TCP-level protection:

```go
// Pseudocode of what we need
if cfg.HTTPS.Enabled {
    // How do we wrap certmagic's internal listener?
    certmagic.HTTPS(...)  // <-- need protection here
} else {
    listener, _ := net.Listen("tcp", ":"+cfg.Server.Port)
    protected := ???.Wrap(listener, ???.Config{
        MaxConnsPerIP: 50,
    })
    srv.Serve(protected)  // instead of srv.ListenAndServe()
}
```

## Constraints

- Must work with CertMagic's automatic certificate management
- Must preserve existing middleware chain
- Must work with host-based routing (subdomains)
- No breaking changes to existing functionality
- Pure Go (cross-compilation support)

## Success Criteria

After integration:
- `TestSecurity_ServiceDuringSlowloris` shows >80% legitimate requests succeed
- No performance regression on `TestBaseline_Throughput`
- HTTPS with automatic certs still works
- All existing tests pass
