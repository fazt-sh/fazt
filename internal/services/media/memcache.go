package media

import (
	"container/list"
	"strings"
	"sync"

	"github.com/fazt-sh/fazt/internal/system"
)

// memEntry is a single cached variant in the LRU.
type memEntry struct {
	key  string // "appID:path"
	data []byte
	mime string
}

// memCache is a byte-size-bounded LRU for processed media variants.
// It sits in front of the SQLite cache to avoid DB hits on repeated requests.
type memCache struct {
	mu      sync.Mutex
	items   map[string]*list.Element
	order   *list.List // front = most recent
	size    int64
	maxSize int64
}

var (
	globalMemCache *memCache
	memCacheOnce   sync.Once
)

func getMemCache() *memCache {
	memCacheOnce.Do(func() {
		mb := system.GetLimits().Media.CacheMemoryMB
		if mb <= 0 {
			mb = 16
		}
		globalMemCache = &memCache{
			items:   make(map[string]*list.Element),
			order:   list.New(),
			maxSize: int64(mb) * 1024 * 1024,
		}
	})
	return globalMemCache
}

func memCacheKey(appID, path string) string {
	return appID + ":" + path
}

// get returns cached data or nil on miss. Promotes on hit.
func (c *memCache) get(appID, path string) ([]byte, string) {
	key := memCacheKey(appID, path)
	c.mu.Lock()
	defer c.mu.Unlock()

	el, ok := c.items[key]
	if !ok {
		return nil, ""
	}

	c.order.MoveToFront(el)
	entry := el.Value.(*memEntry)
	return entry.data, entry.mime
}

// put adds or updates an entry, evicting LRU entries if over budget.
func (c *memCache) put(appID, path string, data []byte, mime string) {
	key := memCacheKey(appID, path)
	entrySize := int64(len(data))

	// Don't cache entries larger than 25% of max — one huge image shouldn't evict everything
	if entrySize > c.maxSize/4 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Update existing
	if el, ok := c.items[key]; ok {
		old := el.Value.(*memEntry)
		c.size -= int64(len(old.data))
		old.data = data
		old.mime = mime
		c.size += entrySize
		c.order.MoveToFront(el)
		c.evict()
		return
	}

	// Add new
	entry := &memEntry{key: key, data: data, mime: mime}
	el := c.order.PushFront(entry)
	c.items[key] = el
	c.size += entrySize
	c.evict()
}

// evict removes LRU entries until size is within budget.
func (c *memCache) evict() {
	for c.size > c.maxSize && c.order.Len() > 0 {
		el := c.order.Back()
		if el == nil {
			break
		}
		entry := el.Value.(*memEntry)
		c.size -= int64(len(entry.data))
		delete(c.items, entry.key)
		c.order.Remove(el)
	}
}

// invalidatePrefix removes all entries whose path starts with prefix.
// Used when an original blob is overwritten or deleted.
func (c *memCache) invalidatePrefix(appID, pathPrefix string) {
	prefix := memCacheKey(appID, pathPrefix)
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, el := range c.items {
		if strings.HasPrefix(key, prefix) {
			entry := el.Value.(*memEntry)
			c.size -= int64(len(entry.data))
			c.order.Remove(el)
			delete(c.items, key)
		}
	}
}
