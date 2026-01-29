package listener

import (
	"net"
	"runtime"

	"github.com/valyala/tcplisten"
)

// ListenTCP creates a TCP listener with platform-specific optimizations.
//
// On Linux, enables:
//   - TCP_DEFER_ACCEPT: Kernel only wakes Go when client sends data
//     (filters slowloris connections that connect but never send)
//   - TCP_FASTOPEN: Reduces latency on repeat connections
//
// On other platforms, falls back to standard net.Listen.
func ListenTCP(network, addr string) (net.Listener, error) {
	// Normalize network to tcp4 for tcplisten compatibility
	if network == "tcp" {
		network = "tcp4"
	}

	if runtime.GOOS == "linux" {
		cfg := tcplisten.Config{
			DeferAccept: true, // Key slowloris defense at kernel level
			FastOpen:    true, // Performance optimization
			// ReusePort: true, // Uncomment for multi-process scaling
		}
		return cfg.NewListener(network, addr)
	}

	// Fallback for non-Linux (macOS, Windows, etc.)
	return net.Listen(network, addr)
}

// ListenConfig holds configuration for creating optimized listeners.
type ListenConfig struct {
	// DeferAccept enables TCP_DEFER_ACCEPT on Linux.
	// The kernel won't wake Go until the client sends data.
	DeferAccept bool

	// FastOpen enables TCP_FASTOPEN for reduced latency.
	FastOpen bool

	// ReusePort enables SO_REUSEPORT for multi-process scaling.
	ReusePort bool
}

// Listen creates a TCP listener with the given configuration.
func Listen(network, addr string, cfg ListenConfig) (net.Listener, error) {
	if runtime.GOOS == "linux" {
		tcpCfg := tcplisten.Config{
			DeferAccept: cfg.DeferAccept,
			FastOpen:    cfg.FastOpen,
			ReusePort:   cfg.ReusePort,
		}
		return tcpCfg.NewListener(network, addr)
	}

	// Fallback for non-Linux
	return net.Listen(network, addr)
}
