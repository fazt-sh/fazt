package egress

import (
	"sync"
	"time"

	"github.com/fazt-sh/fazt/internal/system"
)

// cacheEntry holds a cached response with its expiration and size.
type cacheEntry struct {
	response  *FetchResponse
	expiresAt time.Time
	size      int64
}

// NetCache is an in-memory LRU cache for fetch responses.
type NetCache struct {
	items    map[string]*cacheEntry
	order    []string // LRU order (oldest first)
	mu       sync.RWMutex
	maxItems int
	maxBytes int64
	curBytes int64
	hits     int64
	misses   int64
}

// NewNetCache creates a NetCache with settings from system.Limits.Net.
func NewNetCache() *NetCache {
	netLimits := system.GetLimits().Net
	return &NetCache{
		items:    make(map[string]*cacheEntry),
		maxItems: netLimits.CacheMaxItems,
		maxBytes: netLimits.CacheMaxBytes,
	}
}

// Enabled returns true if the cache is configured.
func (c *NetCache) Enabled() bool {
	return c.maxItems > 0 && c.maxBytes > 0
}

// Get returns a cached response if present and not expired.
func (c *NetCache) Get(key string) (*FetchResponse, bool) {
	c.mu.RLock()
	entry, ok := c.items[key]
	c.mu.RUnlock()

	if !ok {
		c.mu.Lock()
		c.misses++
		c.mu.Unlock()
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		// Expired — remove
		c.mu.Lock()
		c.removeEntry(key)
		c.misses++
		c.mu.Unlock()
		return nil, false
	}

	// Move to end of LRU order (most recently used)
	c.mu.Lock()
	c.hits++
	c.touchLRU(key)
	c.mu.Unlock()

	return entry.response, true
}

// Put adds a response to the cache with the given TTL.
func (c *NetCache) Put(key string, resp *FetchResponse, ttl time.Duration) {
	if !c.Enabled() || ttl <= 0 {
		return
	}

	size := int64(len(resp.body)) + int64(len(key)) + 200 // rough overhead

	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove existing entry if present
	if _, exists := c.items[key]; exists {
		c.removeEntry(key)
	}

	// Evict until under limits
	for (len(c.items) >= c.maxItems || c.curBytes+size > c.maxBytes) && len(c.order) > 0 {
		c.removeEntry(c.order[0])
	}

	// Store
	c.items[key] = &cacheEntry{
		response:  resp,
		expiresAt: time.Now().Add(ttl),
		size:      size,
	}
	c.order = append(c.order, key)
	c.curBytes += size
}

// Clear removes all cached entries.
func (c *NetCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*cacheEntry)
	c.order = nil
	c.curBytes = 0
}

// Stats returns cache statistics.
type CacheStats struct {
	Items    int   `json:"items"`
	Bytes    int64 `json:"bytes"`
	MaxItems int   `json:"max_items"`
	MaxBytes int64 `json:"max_bytes"`
	Hits     int64 `json:"hits"`
	Misses   int64 `json:"misses"`
}

func (c *NetCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return CacheStats{
		Items:    len(c.items),
		Bytes:    c.curBytes,
		MaxItems: c.maxItems,
		MaxBytes: c.maxBytes,
		Hits:     c.hits,
		Misses:   c.misses,
	}
}

// removeEntry removes an entry from the cache (caller must hold write lock).
func (c *NetCache) removeEntry(key string) {
	entry, ok := c.items[key]
	if !ok {
		return
	}
	c.curBytes -= entry.size
	delete(c.items, key)

	// Remove from order
	for i, k := range c.order {
		if k == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			break
		}
	}
}

// touchLRU moves a key to the end of the order (most recently used).
func (c *NetCache) touchLRU(key string) {
	for i, k := range c.order {
		if k == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			c.order = append(c.order, key)
			break
		}
	}
}

// CacheKey builds the cache key for a fetch request.
// Only GET requests without auth are cacheable.
func CacheKey(method, rawURL string, hasAuth bool) (string, bool) {
	if method != "" && method != "GET" {
		return "", false
	}
	if hasAuth {
		return "", false
	}
	return "GET:" + rawURL, true
}
