package egress

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"


	"github.com/fazt-sh/fazt/internal/system"
)

// blockedNets contains IP ranges that serverless code must never reach.
var blockedNets []net.IPNet

func init() {
	cidrs := []string{
		"127.0.0.0/8",     // Loopback
		"10.0.0.0/8",      // Private (A)
		"172.16.0.0/12",   // Private (B)
		"192.168.0.0/16",  // Private (C)
		"169.254.0.0/16",  // Link-local / cloud metadata
		"100.64.0.0/10",   // CGNAT
		"0.0.0.0/8",       // "This network"
		"::1/128",         // IPv6 loopback
		"fc00::/7",        // IPv6 unique-local
		"fe80::/10",       // IPv6 link-local
	}
	for _, cidr := range cidrs {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Sprintf("egress: bad CIDR %q: %v", cidr, err))
		}
		blockedNets = append(blockedNets, *ipnet)
	}
}

// isBlockedIP returns true if the IP falls in a blocked range.
func isBlockedIP(ip net.IP) bool {
	for _, n := range blockedNets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// isIPLiteral returns true if the host is a raw IP address (not a domain).
func isIPLiteral(host string) bool {
	// Strip brackets for IPv6 literals like [::1]
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}
	return net.ParseIP(host) != nil
}

// canonicalizeHost normalizes a hostname for allowlist comparison.
func canonicalizeHost(raw string) string {
	host := strings.ToLower(raw)
	host = strings.TrimSuffix(host, ".")
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	host = strings.TrimSuffix(host, ".")
	return host
}

// unsafeHeaders are stripped from user-provided options before dispatch.
var unsafeHeaders = map[string]bool{
	"host":                true,
	"connection":          true,
	"proxy-authorization": true,
	"proxy-connection":    true,
	"transfer-encoding":   true,
	"accept-encoding":     true,
}

// FetchOptions describes the options for a fetch call.
type FetchOptions struct {
	Method  string
	Headers map[string]string
	Body    string
	Timeout time.Duration
	Auth    string // Secret name for Phase 2 injection
}

// FetchResponse is the response returned from a fetch call.
type FetchResponse struct {
	Status  int
	OK      bool
	Headers map[string]string
	body    []byte
}

// Text returns the response body as a string.
func (r *FetchResponse) Text() string {
	return string(r.body)
}

// Body returns the raw response body bytes.
func (r *FetchResponse) Body() []byte {
	return r.body
}

// EgressProxy owns the hardened http.Client and enforces all security rules.
type EgressProxy struct {
	client       *http.Client
	allowlist    *Allowlist
	secrets      *SecretsStore
	rateLimiter  *RateLimiter
	logger       *NetLogger
	cache        *NetCache
	callLimit    int
	maxReqBody   int64
	maxRespBody  int64
	maxRedirects int
	perAppLimit  int32
	globalLimit  int32
	appConns     sync.Map   // map[string]*int32
	globalConns  int32
}

// NewEgressProxy creates a new EgressProxy with settings from system.Limits.Net.
func NewEgressProxy(allowlist *Allowlist) *EgressProxy {
	netLimits := system.GetLimits().Net

	proxy := &EgressProxy{
		allowlist:    allowlist,
		rateLimiter:  NewRateLimiter(netLimits.RateLimit, netLimits.RateBurst),
		callLimit:    netLimits.MaxCalls,
		maxReqBody:   netLimits.MaxRequestBody,
		maxRespBody:  netLimits.MaxResponse,
		maxRedirects: netLimits.MaxRedirects,
		perAppLimit:  int32(netLimits.AppConcurrency),
		globalLimit:  int32(netLimits.Concurrency),
	}

	// Safe dialer: validates resolved IPs before connecting
	safeDialer := &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 10 * time.Second,
	}

	transport := &http.Transport{
		Proxy: nil, // CRITICAL: ignore HTTP_PROXY/HTTPS_PROXY env
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, errBlocked(fmt.Sprintf("invalid address: %s", addr))
			}

			// Resolve DNS
			ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil {
				return nil, errNet(fmt.Sprintf("DNS resolution failed: %v", err))
			}

			// Check every resolved IP
			for _, ipAddr := range ips {
				if isBlockedIP(ipAddr.IP) {
					return nil, errBlocked(fmt.Sprintf("blocked IP %s for host %s", ipAddr.IP, host))
				}
			}

			// Connect to the first valid IP
			return safeDialer.DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
		},
		DisableCompression:     true, // Raw bodies so LimitReader is accurate
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxResponseHeaderBytes: 1 << 20, // 1MB header limit
		MaxIdleConns:          20,
		MaxIdleConnsPerHost:   2,
		IdleConnTimeout:       10 * time.Second,
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12},
	}

	proxy.client = &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return proxy.checkRedirect(req, via)
		},
		Timeout: 0, // We control timeout via context
	}

	return proxy
}

// checkRedirect validates each redirect hop.
func (p *EgressProxy) checkRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= p.maxRedirects {
		return errBlocked(fmt.Sprintf("too many redirects (%d)", len(via)))
	}

	host := canonicalizeHost(req.URL.Hostname())

	// Re-check allowlist at each hop
	if p.allowlist != nil {
		// Get app ID from the first request's context
		appID := ""
		if len(via) > 0 {
			if id, ok := via[0].Context().Value(ctxKeyAppID).(string); ok {
				appID = id
			}
		}
		if !p.allowlist.IsAllowed(host, appID) {
			return errBlocked(fmt.Sprintf("redirect to non-allowed domain: %s", host))
		}
	}

	// Block scheme downgrade
	if req.URL.Scheme != "https" {
		// Check if domain explicitly allows HTTP
		if p.allowlist != nil {
			appID := ""
			if len(via) > 0 {
				if id, ok := via[0].Context().Value(ctxKeyAppID).(string); ok {
					appID = id
				}
			}
			entry := p.allowlist.entryFor(host, appID)
			if entry == nil || entry.HTTPSOnly {
				return errBlocked(fmt.Sprintf("redirect to non-HTTPS URL: %s", req.URL))
			}
		} else {
			return errBlocked(fmt.Sprintf("redirect to non-HTTPS URL: %s", req.URL))
		}
	}

	return nil
}

type contextKey string

const ctxKeyAppID contextKey = "egress_app_id"

// Fetch performs a validated outbound HTTP request.
func (p *EgressProxy) Fetch(ctx context.Context, appID string, rawURL string, opts FetchOptions) (*FetchResponse, error) {
	// Parse and validate URL
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, errBlocked(fmt.Sprintf("invalid URL: %v", err))
	}

	host := canonicalizeHost(parsed.Hostname())

	// Block IP literals — allowlist operates on domain names
	if isIPLiteral(parsed.Hostname()) {
		return nil, errBlocked(fmt.Sprintf("IP literal URLs not allowed: %s", parsed.Hostname()))
	}

	// Check scheme
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return nil, errBlocked(fmt.Sprintf("unsupported scheme: %s", parsed.Scheme))
	}

	// HTTPS enforcement
	if parsed.Scheme != "https" {
		if p.allowlist != nil {
			entry := p.allowlist.entryFor(host, appID)
			if entry == nil || entry.HTTPSOnly {
				return nil, errBlocked("HTTPS required (domain does not allow HTTP)")
			}
		} else {
			return nil, errBlocked("HTTPS required")
		}
	}

	// Check allowlist
	if p.allowlist != nil && !p.allowlist.IsAllowed(host, appID) {
		return nil, errBlocked(fmt.Sprintf("domain not in allowlist: %s", host))
	}

	// Get per-domain config for rate limiting and response size
	var domainRate, domainBurst int
	var domainMaxResp int64
	if p.allowlist != nil {
		if entry := p.allowlist.entryFor(host, appID); entry != nil {
			domainRate = entry.RateLimit
			domainBurst = entry.RateBurst
			if entry.MaxResponse > 0 {
				domainMaxResp = entry.MaxResponse
			}
		}
	}

	// Rate limiting
	if p.rateLimiter != nil && !p.rateLimiter.Allow(host, domainRate, domainBurst) {
		return nil, errRate(fmt.Sprintf("rate limit exceeded for %s", host))
	}

	// Check request body size
	if int64(len(opts.Body)) > p.maxReqBody {
		return nil, errSize(fmt.Sprintf("request body too large: %d > %d bytes", len(opts.Body), p.maxReqBody))
	}

	// Concurrency control — global
	if atomic.LoadInt32(&p.globalConns) >= p.globalLimit {
		return nil, errLimit("global concurrent outbound limit reached")
	}

	// Concurrency control — per-app
	appCount := p.getAppCounter(appID)
	if atomic.LoadInt32(appCount) >= p.perAppLimit {
		return nil, errLimit(fmt.Sprintf("per-app concurrent outbound limit reached for %s", appID))
	}

	// Check cache before acquiring concurrency slots
	if p.cache != nil && p.cache.Enabled() {
		cacheKey, cacheable := CacheKey(opts.Method, rawURL, opts.Auth != "")
		if cacheable {
			// Check per-domain cache TTL
			cacheTTL := 0
			if p.allowlist != nil {
				if entry := p.allowlist.entryFor(host, appID); entry != nil {
					cacheTTL = entry.CacheTTL
				}
			}
			if cacheTTL > 0 {
				if cached, ok := p.cache.Get(cacheKey); ok {
					return cached, nil
				}
			}
		}
	}

	// Acquire concurrency slots
	atomic.AddInt32(&p.globalConns, 1)
	defer atomic.AddInt32(&p.globalConns, -1)
	atomic.AddInt32(appCount, 1)
	defer atomic.AddInt32(appCount, -1)

	fetchStart := time.Now()

	// Build request
	method := opts.Method
	if method == "" {
		method = "GET"
	}

	var bodyReader io.Reader
	if opts.Body != "" {
		bodyReader = strings.NewReader(opts.Body)
	}

	// Store app ID in context for redirect checking
	reqCtx := context.WithValue(ctx, ctxKeyAppID, appID)

	req, err := http.NewRequestWithContext(reqCtx, method, rawURL, bodyReader)
	if err != nil {
		return nil, errNet(fmt.Sprintf("failed to create request: %v", err))
	}

	// Set user headers (sanitized)
	for k, v := range opts.Headers {
		if !unsafeHeaders[strings.ToLower(k)] {
			req.Header.Set(k, v)
		}
	}

	// Force identity encoding — prevents gzip bombs
	req.Header.Set("Accept-Encoding", "identity")

	// Inject auth secret if requested
	if opts.Auth != "" {
		if err := InjectSecretIntoRequest(p.secrets, req, opts.Auth, appID, host); err != nil {
			return nil, err
		}
	}

	// Use per-domain response limit if configured
	respLimit := p.maxRespBody
	if domainMaxResp > 0 {
		respLimit = domainMaxResp
	}

	// Execute request
	resp, err := p.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, errTimeout("request timed out")
		}
		// Check if it's one of our errors (from DialContext or CheckRedirect)
		if ee, ok := err.(*EgressError); ok {
			return nil, ee
		}
		// Unwrap url.Error to find our EgressError
		if ue, ok := err.(*url.Error); ok {
			if ee, ok2 := ue.Err.(*EgressError); ok2 {
				return nil, ee
			}
		}
		return nil, errNet(fmt.Sprintf("request failed: %v", err))
	}
	defer resp.Body.Close()

	// Read response body with size limit
	body, err := io.ReadAll(io.LimitReader(resp.Body, respLimit+1))
	if err != nil {
		return nil, errNet(fmt.Sprintf("failed to read response: %v", err))
	}
	if int64(len(body)) > respLimit {
		return nil, errSize(fmt.Sprintf("response body too large: exceeds %d bytes", respLimit))
	}

	// Build response headers — lowercase keys, first value only
	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[strings.ToLower(k)] = v[0]
		}
	}

	fetchResp := &FetchResponse{
		Status:  resp.StatusCode,
		OK:      resp.StatusCode >= 200 && resp.StatusCode < 300,
		Headers: headers,
		body:    body,
	}

	fetchDuration := time.Since(fetchStart)

	// Log the request
	if p.logger != nil {
		p.logger.LogFromFetch(appID, rawURL, method, fetchResp, nil, fetchDuration, int64(len(opts.Body)))
	}

	// Store in cache if applicable
	if p.cache != nil && p.cache.Enabled() {
		cacheKey, cacheable := CacheKey(method, rawURL, opts.Auth != "")
		if cacheable {
			cacheTTL := 0
			if p.allowlist != nil {
				if entry := p.allowlist.entryFor(host, appID); entry != nil {
					cacheTTL = entry.CacheTTL
				}
			}
			if cacheTTL > 0 && fetchResp.OK {
				p.cache.Put(cacheKey, fetchResp, time.Duration(cacheTTL)*time.Second)
			}
		}
	}

	return fetchResp, nil
}

// getAppCounter returns the atomic counter for a given app.
func (p *EgressProxy) getAppCounter(appID string) *int32 {
	val, _ := p.appConns.LoadOrStore(appID, new(int32))
	return val.(*int32)
}

// SetSecrets sets the secrets store for auth injection.
func (p *EgressProxy) SetSecrets(secrets *SecretsStore) {
	p.secrets = secrets
}

// SetLogger sets the net logger for request logging.
func (p *EgressProxy) SetLogger(logger *NetLogger) {
	p.logger = logger
}

// SetCache sets the response cache.
func (p *EgressProxy) SetCache(cache *NetCache) {
	p.cache = cache
}

// GlobalConnections returns the current global connection count (for testing).
func (p *EgressProxy) GlobalConnections() int32 {
	return atomic.LoadInt32(&p.globalConns)
}
