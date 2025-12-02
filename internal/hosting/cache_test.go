package hosting

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestVFSCacheBehavior(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	fs := NewSQLFileSystem(db)
	site := "cache-test"
	path := "index.html"
	content1 := []byte("Version 1")
	content2 := []byte("Version 2")

	// 1. Write File (Version 1)
	if err := fs.WriteFile(site, path, bytes.NewReader(content1), int64(len(content1)), "text/plain"); err != nil {
		t.Fatalf("Initial write failed: %v", err)
	}

	// 2. Read File (Should populate cache)
	file, err := fs.ReadFile(site, path)
	if err != nil {
		t.Fatalf("Initial read failed: %v", err)
	}
	defer file.Content.Close()
	
	// 3. Sneaky Update directly to DB (Bypassing WriteFile and Cache Invalidation)
	// This simulates the "stale cache" scenario to prove we are reading from cache.
	_, err = db.Exec("UPDATE files SET content = ? WHERE site_id = ? AND path = ?", content2, site, path)
	if err != nil {
		t.Fatalf("Direct DB update failed: %v", err)
	}

	// 4. Read File again (Should be Version 1 from Cache)
	fileCached, err := fs.ReadFile(site, path)
	if err != nil {
		t.Fatalf("Cached read failed: %v", err)
	}
	defer fileCached.Content.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(fileCached.Content)
	if buf.String() != string(content1) {
		t.Errorf("Cache miss! Expected %q (Version 1), got %q", string(content1), buf.String())
	} else {
		t.Log("Cache hit confirmed: Got old content despite DB update")
	}

	// 5. Proper Write (Version 2) -> Should invalidate cache
	if err := fs.WriteFile(site, path, bytes.NewReader(content2), int64(len(content2)), "text/plain"); err != nil {
		t.Fatalf("Proper write failed: %v", err)
	}

	// 6. Read File again (Should be Version 2 from DB)
	fileNew, err := fs.ReadFile(site, path)
	if err != nil {
		t.Fatalf("Post-write read failed: %v", err)
	}
	defer fileNew.Content.Close()

	bufNew := new(bytes.Buffer)
	bufNew.ReadFrom(fileNew.Content)
	if bufNew.String() != string(content2) {
		t.Errorf("Cache invalidation failed! Expected %q (Version 2), got %q", string(content2), bufNew.String())
	}
}

func TestVFSCacheEviction(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	fs := NewSQLFileSystem(db)
	site := "evict-test"

	// Fill cache with enough items to trigger eviction
	// We need > 1000 items. Logic is: check > 1000, then clear.
	// If we insert 1000 items, size is 1000.
	// Insert 1001st: size becomes 1001.
	// Insert 1002nd: size is 1001 (>1000), so it clears.
	for i := 0; i <= 1002; i++ {
		name := fmt.Sprintf("file%d", i)
		fs.WriteFile(site, name, strings.NewReader("data"), 4, "text/plain")
		// Read to populate cache
		f, _ := fs.ReadFile(site, name)
		f.Content.Close()
	}

	// Check if cache was cleared (size should be 1 now, for the last item read)
	// We can't check size directly.
	// But we can check if the *first* item is still cached by doing a sneaky update.
	
	// Update file0 in DB
	db.Exec("UPDATE files SET content = ? WHERE site_id = ? AND path = ?", []byte("updated"), site, "file0")

	// Read file0. If evicted, we get "updated". If cached, we get "data".
	f, _ := fs.ReadFile(site, "file0")
	defer f.Content.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(f.Content)

	if buf.String() == "updated" {
		t.Log("Eviction confirmed: file0 was re-fetched from DB")
	} else {
		t.Error("Eviction failed: file0 is still in cache")
	}
}
