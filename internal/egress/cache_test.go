package egress

import (
	"testing"
	"time"

	"github.com/fazt-sh/fazt/internal/system"
)

func testCache() *NetCache {
	system.ResetCachedLimits()
	// Override defaults for testing
	return &NetCache{
		items:    make(map[string]*cacheEntry),
		maxItems: 10,
		maxBytes: 10 * 1024, // 10KB
	}
}

func TestCacheGetMiss(t *testing.T) {
	c := testCache()

	_, ok := c.Get("nonexistent")
	if ok {
		t.Error("expected miss for nonexistent key")
	}
}

func TestCachePutAndGet(t *testing.T) {
	c := testCache()

	resp := &FetchResponse{
		Status:  200,
		OK:      true,
		Headers: map[string]string{"content-type": "text/plain"},
		body:    []byte("hello"),
	}

	c.Put("GET:https://api.com/data", resp, 5*time.Second)

	got, ok := c.Get("GET:https://api.com/data")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.Status != 200 {
		t.Errorf("Status: got %d, want 200", got.Status)
	}
	if got.Text() != "hello" {
		t.Errorf("Text(): got %q, want %q", got.Text(), "hello")
	}
}

func TestCacheExpiration(t *testing.T) {
	c := testCache()

	resp := &FetchResponse{body: []byte("data")}
	c.Put("key", resp, 1*time.Millisecond)

	time.Sleep(5 * time.Millisecond)

	_, ok := c.Get("key")
	if ok {
		t.Error("expected miss for expired entry")
	}
}

func TestCacheEviction(t *testing.T) {
	c := &NetCache{
		items:    make(map[string]*cacheEntry),
		maxItems: 2,
		maxBytes: 100 * 1024,
	}

	c.Put("key1", &FetchResponse{body: []byte("a")}, time.Minute)
	c.Put("key2", &FetchResponse{body: []byte("b")}, time.Minute)
	c.Put("key3", &FetchResponse{body: []byte("c")}, time.Minute) // Should evict key1

	_, ok := c.Get("key1")
	if ok {
		t.Error("key1 should have been evicted")
	}
	_, ok = c.Get("key3")
	if !ok {
		t.Error("key3 should be in cache")
	}
}

func TestCacheLRUOrder(t *testing.T) {
	c := &NetCache{
		items:    make(map[string]*cacheEntry),
		maxItems: 2,
		maxBytes: 100 * 1024,
	}

	c.Put("key1", &FetchResponse{body: []byte("a")}, time.Minute)
	c.Put("key2", &FetchResponse{body: []byte("b")}, time.Minute)

	// Access key1 to make it recently used
	c.Get("key1")

	// Adding key3 should evict key2 (least recently used), not key1
	c.Put("key3", &FetchResponse{body: []byte("c")}, time.Minute)

	_, ok := c.Get("key1")
	if !ok {
		t.Error("key1 should still be in cache (recently accessed)")
	}
	_, ok = c.Get("key2")
	if ok {
		t.Error("key2 should have been evicted (LRU)")
	}
}

func TestCacheClear(t *testing.T) {
	c := testCache()
	c.Put("key1", &FetchResponse{body: []byte("a")}, time.Minute)
	c.Put("key2", &FetchResponse{body: []byte("b")}, time.Minute)

	c.Clear()

	stats := c.Stats()
	if stats.Items != 0 {
		t.Errorf("Items: got %d, want 0", stats.Items)
	}
}

func TestCacheStats(t *testing.T) {
	c := testCache()
	c.Put("key1", &FetchResponse{body: []byte("hello")}, time.Minute)

	c.Get("key1") // hit
	c.Get("key2") // miss

	stats := c.Stats()
	if stats.Items != 1 {
		t.Errorf("Items: got %d, want 1", stats.Items)
	}
	if stats.Hits != 1 {
		t.Errorf("Hits: got %d, want 1", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Misses: got %d, want 1", stats.Misses)
	}
}

func TestCacheKeyRules(t *testing.T) {
	tests := []struct {
		method    string
		url       string
		hasAuth   bool
		cacheable bool
	}{
		{"GET", "https://api.com/data", false, true},
		{"", "https://api.com/data", false, true}, // empty method = GET
		{"POST", "https://api.com/data", false, false},
		{"GET", "https://api.com/data", true, false}, // auth = not cacheable
		{"PUT", "https://api.com/data", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.method+"_"+tt.url, func(t *testing.T) {
			_, cacheable := CacheKey(tt.method, tt.url, tt.hasAuth)
			if cacheable != tt.cacheable {
				t.Errorf("cacheable: got %v, want %v", cacheable, tt.cacheable)
			}
		})
	}
}

func TestCacheDisabledByDefault(t *testing.T) {
	system.ResetCachedLimits()
	c := NewNetCache() // Uses system defaults (0, 0)
	if c.Enabled() {
		t.Error("cache should be disabled by default (maxItems=0)")
	}
}
