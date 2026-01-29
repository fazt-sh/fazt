// Package listener provides TCP-level connection management.
// This operates BEFORE net/http sees connections - critical for slowloris protection.
package listener

import (
	"log"
	"net"
	"sync"
	"sync/atomic"
)

// ConnLimiter wraps a net.Listener to enforce per-IP connection limits.
// This operates at the TCP Accept level, rejecting connections before
// they consume goroutines or enter HTTP parsing.
type ConnLimiter struct {
	net.Listener
	maxPerIP  int
	maxTotal  int64
	total     int64 // atomic
	mu        sync.Mutex
	counts    map[string]int
	onReject  func(ip string, reason string) // optional callback
}

// ConnLimiterConfig configures the connection limiter.
type ConnLimiterConfig struct {
	MaxConnsPerIP int   // Max concurrent connections per IP (default: 50)
	MaxTotalConns int64 // Max total concurrent connections (default: 10000)
	OnReject      func(ip string, reason string)
}

// NewConnLimiter creates a connection limiter wrapping the given listener.
func NewConnLimiter(l net.Listener, cfg ConnLimiterConfig) *ConnLimiter {
	if cfg.MaxConnsPerIP <= 0 {
		cfg.MaxConnsPerIP = 50
	}
	if cfg.MaxTotalConns <= 0 {
		cfg.MaxTotalConns = 10000
	}
	return &ConnLimiter{
		Listener: l,
		maxPerIP: cfg.MaxConnsPerIP,
		maxTotal: cfg.MaxTotalConns,
		counts:   make(map[string]int),
		onReject: cfg.OnReject,
	}
}

// Accept implements net.Listener.Accept with connection limiting.
func (l *ConnLimiter) Accept() (net.Conn, error) {
	for {
		conn, err := l.Listener.Accept()
		if err != nil {
			return nil, err
		}

		// Check total connection limit
		if atomic.AddInt64(&l.total, 1) > l.maxTotal {
			atomic.AddInt64(&l.total, -1)
			conn.Close()
			if l.onReject != nil {
				l.onReject("", "total_limit")
			}
			continue
		}

		// Extract IP from RemoteAddr
		ip := extractIP(conn.RemoteAddr())
		if ip == "" {
			atomic.AddInt64(&l.total, -1)
			conn.Close()
			continue
		}

		// Check per-IP limit
		l.mu.Lock()
		count := l.counts[ip]
		if count >= l.maxPerIP {
			l.mu.Unlock()
			atomic.AddInt64(&l.total, -1)
			conn.Close()
			if l.onReject != nil {
				l.onReject(ip, "per_ip_limit")
			}
			continue
		}
		l.counts[ip]++
		l.mu.Unlock()

		return &trackedConn{
			Conn: conn,
			ip:   ip,
			l:    l,
		}, nil
	}
}

// Stats returns current connection statistics.
func (l *ConnLimiter) Stats() (total int64, uniqueIPs int) {
	l.mu.Lock()
	uniqueIPs = len(l.counts)
	l.mu.Unlock()
	return atomic.LoadInt64(&l.total), uniqueIPs
}

// extractIP gets the IP string from a net.Addr.
func extractIP(addr net.Addr) string {
	switch v := addr.(type) {
	case *net.TCPAddr:
		return v.IP.String()
	case *net.UDPAddr:
		return v.IP.String()
	default:
		host, _, err := net.SplitHostPort(addr.String())
		if err != nil {
			return ""
		}
		return host
	}
}

// trackedConn wraps net.Conn to decrement counters on Close.
type trackedConn struct {
	net.Conn
	ip     string
	l      *ConnLimiter
	closed int32 // atomic flag to prevent double-decrement
}

func (c *trackedConn) Close() error {
	// Ensure we only decrement once, even if Close() is called multiple times
	if atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		c.l.mu.Lock()
		c.l.counts[c.ip]--
		if c.l.counts[c.ip] <= 0 {
			delete(c.l.counts, c.ip)
		}
		c.l.mu.Unlock()
		atomic.AddInt64(&c.l.total, -1)
	}
	return c.Conn.Close()
}

// DefaultOnReject is a no-op to avoid log noise.
// Set a custom OnReject callback for debugging or metrics.
func DefaultOnReject(ip string, reason string) {
	// No-op by default - set custom callback for logging/metrics
}

// LoggingOnReject logs rejected connections (use for debugging).
func LoggingOnReject(ip string, reason string) {
	if ip != "" {
		log.Printf("Connection rejected: ip=%s reason=%s", ip, reason)
	} else {
		log.Printf("Connection rejected: reason=%s", reason)
	}
}
