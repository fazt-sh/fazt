# Research Query: Go net/http Slowloris Protection Without Reverse Proxy

## The Problem

Go's `net/http.Server` is vulnerable to slowloris attacks. The server spawns one
goroutine per connection, and each goroutine blocks waiting for HTTP headers:

```go
// Inside net/http - this blocks until headers complete
req, err := c.readRequest()
```

An attacker sends partial HTTP headers slowly, never completing them. The
goroutine stays blocked. With enough connections, all server resources are
consumed and legitimate users can't connect.

**`ReadHeaderTimeout` is insufficient** - attacker can send data just under the
timeout threshold, or use many IPs to rotate connections.

**Middleware can't help** - it runs AFTER headers are parsed. Slowloris attacks
happen BEFORE that point.

## What We Need

A Go-native solution that limits or rejects connections **at the TCP accept
level**, before `net/http` starts parsing headers.

```
┌─────────────────────────────────────────────────────────┐
│                    Attack Surface                        │
│                                                          │
│  TCP Accept    →    Header Parse    →    Request Handle  │
│  (net.Listener)     (net/http)           (http.Handler)  │
│                                                          │
│       ↑                  ↑                    ↑          │
│  NEED PROTECTION    ReadHeaderTimeout    Middleware      │
│  HERE               (weak)               (too late)      │
└─────────────────────────────────────────────────────────┘
```

## Specific Questions

1. **Is there a `net.Listener` wrapper that tracks connections per-IP and rejects
   when a threshold is exceeded?**

2. **Does `fasthttp` or another Go HTTP server handle slowloris better than
   `net/http`? How?**

3. **Can `netutil.LimitListener` from `golang.org/x/net` help? What are its
   limitations?**

4. **What socket options (`SO_*`, `TCP_*`) can Go set to mitigate at kernel
   level?**

5. **How do production Go services (Cloudflare, etc.) handle this without
   putting nginx/caddy in front?**

## Solution Requirements

- Pure Go (no CGO, must cross-compile)
- Works with standard `http.Handler`
- Works with TLS (`tls.Listener`)
- Single binary (no sidecar nginx/haproxy)
- Per-IP connection limiting at TCP level
- Minimal latency overhead

## Ideal Solution Shape

```go
// Something like this:
listener, _ := net.Listen("tcp", ":443")
protected := slowloris.Wrap(listener, slowloris.Config{
    MaxConnsPerIP:      50,
    MaxTotalConns:      10000,
    ConnectionTimeout:  10 * time.Second,
})
srv := &http.Server{Handler: myHandler}
srv.Serve(protected)
```

Or built into an alternative server:

```go
// Or a hardened server that handles this internally
srv := hardened.Server{
    Handler:       myHandler,
    MaxConnsPerIP: 50,
}
srv.ListenAndServe()
```

## Search Terms

- go slowloris protection net.Listener
- golang per-ip connection limit tcp
- go http server connection tracking
- fasthttp slowloris protection
- golang.org/x/net/netutil LimitListener per-ip
- go tcp accept rate limiting
- production go http without nginx
