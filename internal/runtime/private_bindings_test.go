package runtime

import (
	"bytes"
	"database/sql"
	"testing"

	"github.com/dop251/goja"
	_ "modernc.org/sqlite"
)

// setupPrivateTestDB creates a temporary in-memory database for testing
func setupPrivateTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}

	schema := `
	CREATE TABLE files (
		site_id TEXT NOT NULL,
		path TEXT NOT NULL,
		content BLOB,
		size_bytes INTEGER NOT NULL,
		mime_type TEXT,
		hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (site_id, path)
	);
	CREATE TABLE apps (
		id TEXT PRIMARY KEY,
		title TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

// createPrivateFile creates a file in the test database
func createPrivateFile(t *testing.T, db *sql.DB, siteID, path, content string) {
	_, err := db.Exec(`
		INSERT INTO files (site_id, path, content, size_bytes, mime_type, hash)
		VALUES (?, ?, ?, ?, 'application/json', 'test-hash')
	`, siteID, path, content, len(content))
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
}

func TestPrivateFileLoader_Read(t *testing.T) {
	db := setupPrivateTestDB(t)
	defer db.Close()

	createPrivateFile(t, db, "test-app", "private/config.json", `{"key": "value"}`)

	loader := NewPrivateFileLoader(db, "test-app")

	t.Run("read existing file", func(t *testing.T) {
		content, err := loader.Read("config.json")
		if err != nil {
			t.Fatalf("Read failed: %v", err)
		}
		if content != `{"key": "value"}` {
			t.Errorf("expected {\"key\": \"value\"}, got %s", content)
		}
	})

	t.Run("read with leading slash", func(t *testing.T) {
		content, err := loader.Read("/config.json")
		if err != nil {
			t.Fatalf("Read failed: %v", err)
		}
		if content != `{"key": "value"}` {
			t.Errorf("expected {\"key\": \"value\"}, got %s", content)
		}
	})

	t.Run("read non-existent file", func(t *testing.T) {
		_, err := loader.Read("missing.json")
		if err == nil {
			t.Error("expected error for missing file")
		}
	})

	t.Run("block path traversal", func(t *testing.T) {
		_, err := loader.Read("../api/main.js")
		if err == nil {
			t.Error("expected error for path traversal")
		}
	})
}

func TestPrivateFileLoader_Exists(t *testing.T) {
	db := setupPrivateTestDB(t)
	defer db.Close()

	createPrivateFile(t, db, "test-app", "private/data.json", `[]`)

	loader := NewPrivateFileLoader(db, "test-app")

	if !loader.Exists("data.json") {
		t.Error("expected data.json to exist")
	}

	if loader.Exists("missing.json") {
		t.Error("expected missing.json to not exist")
	}

	if loader.Exists("../api/main.js") {
		t.Error("expected path traversal to return false")
	}
}

func TestPrivateFileLoader_List(t *testing.T) {
	db := setupPrivateTestDB(t)
	defer db.Close()

	createPrivateFile(t, db, "test-app", "private/config.json", `{}`)
	createPrivateFile(t, db, "test-app", "private/data/users.json", `[]`)
	createPrivateFile(t, db, "test-app", "public/index.html", `<html></html>`)

	loader := NewPrivateFileLoader(db, "test-app")
	files := loader.List()

	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d: %v", len(files), files)
	}

	// Check that paths don't have "private/" prefix
	found := make(map[string]bool)
	for _, f := range files {
		found[f] = true
	}

	if !found["config.json"] {
		t.Error("missing config.json in list")
	}
	if !found["data/users.json"] {
		t.Error("missing data/users.json in list")
	}
}

func TestInjectPrivateNamespace(t *testing.T) {
	db := setupPrivateTestDB(t)
	defer db.Close()

	createPrivateFile(t, db, "test-app", "private/config.json", `{"debug": true}`)
	createPrivateFile(t, db, "test-app", "private/users.json", `[{"id": 1, "name": "Alice"}]`)

	loader := NewPrivateFileLoader(db, "test-app")

	vm := goja.New()
	// Set up fazt namespace first
	vm.Set("fazt", vm.NewObject())

	if err := InjectPrivateNamespace(vm, loader); err != nil {
		t.Fatalf("InjectPrivateNamespace failed: %v", err)
	}

	t.Run("read returns string", func(t *testing.T) {
		val, err := vm.RunString(`fazt.private.read('config.json')`)
		if err != nil {
			t.Fatalf("read failed: %v", err)
		}
		if val.String() != `{"debug": true}` {
			t.Errorf("expected {\"debug\": true}, got %s", val.String())
		}
	})

	t.Run("readJSON returns object", func(t *testing.T) {
		val, err := vm.RunString(`fazt.private.readJSON('config.json').debug`)
		if err != nil {
			t.Fatalf("readJSON failed: %v", err)
		}
		if val.ToBoolean() != true {
			t.Errorf("expected true, got %v", val.ToBoolean())
		}
	})

	t.Run("readJSON returns array", func(t *testing.T) {
		val, err := vm.RunString(`fazt.private.readJSON('users.json')[0].name`)
		if err != nil {
			t.Fatalf("readJSON failed: %v", err)
		}
		if val.String() != "Alice" {
			t.Errorf("expected Alice, got %s", val.String())
		}
	})

	t.Run("exists returns boolean", func(t *testing.T) {
		val, err := vm.RunString(`fazt.private.exists('config.json')`)
		if err != nil {
			t.Fatalf("exists failed: %v", err)
		}
		if val.ToBoolean() != true {
			t.Errorf("expected true")
		}

		val, err = vm.RunString(`fazt.private.exists('missing.json')`)
		if err != nil {
			t.Fatalf("exists failed: %v", err)
		}
		if val.ToBoolean() != false {
			t.Errorf("expected false")
		}
	})

	t.Run("list returns array", func(t *testing.T) {
		val, err := vm.RunString(`fazt.private.list().length`)
		if err != nil {
			t.Fatalf("list failed: %v", err)
		}
		if val.ToInteger() != 2 {
			t.Errorf("expected 2, got %d", val.ToInteger())
		}
	})

	t.Run("read missing returns undefined", func(t *testing.T) {
		val, err := vm.RunString(`fazt.private.read('missing.json') === undefined`)
		if err != nil {
			t.Fatalf("read failed: %v", err)
		}
		if val.ToBoolean() != true {
			t.Errorf("expected undefined for missing file")
		}
	})

	t.Run("readJSON missing returns null", func(t *testing.T) {
		val, err := vm.RunString(`fazt.private.readJSON('missing.json') === null`)
		if err != nil {
			t.Fatalf("readJSON failed: %v", err)
		}
		if val.ToBoolean() != true {
			t.Errorf("expected null for missing file")
		}
	})
}

func TestPrivateFileLoader_EdgeCases(t *testing.T) {
	db := setupPrivateTestDB(t)
	defer db.Close()

	// Empty file
	createPrivateFile(t, db, "test-app", "private/empty.json", ``)

	// Large-ish JSON
	largeJSON := `{"items": [`
	for i := 0; i < 100; i++ {
		if i > 0 {
			largeJSON += ","
		}
		largeJSON += `{"id":` + string(rune('0'+i%10)) + `}`
	}
	largeJSON += `]}`
	createPrivateFile(t, db, "test-app", "private/large.json", largeJSON)

	// Nested path
	createPrivateFile(t, db, "test-app", "private/data/nested/deep.json", `{"deep": true}`)

	// Non-JSON file
	createPrivateFile(t, db, "test-app", "private/readme.txt", `This is plain text`)

	// File with special characters in content
	createPrivateFile(t, db, "test-app", "private/special.json", `{"emoji": "🚀", "quotes": "\"hello\""}`)

	loader := NewPrivateFileLoader(db, "test-app")

	t.Run("empty file read", func(t *testing.T) {
		content, err := loader.Read("empty.json")
		if err != nil {
			t.Fatalf("Read empty failed: %v", err)
		}
		if content != "" {
			t.Errorf("expected empty string, got %q", content)
		}
	})

	t.Run("empty file readJSON returns null", func(t *testing.T) {
		vm := goja.New()
		vm.Set("fazt", vm.NewObject())
		InjectPrivateNamespace(vm, loader)

		val, err := vm.RunString(`fazt.private.readJSON('empty.json')`)
		if err != nil {
			t.Fatalf("readJSON failed: %v", err)
		}
		// Empty string is invalid JSON, should return null
		if !goja.IsNull(val) {
			t.Errorf("expected null for empty file, got %v", val)
		}
	})

	t.Run("nested path read", func(t *testing.T) {
		content, err := loader.Read("data/nested/deep.json")
		if err != nil {
			t.Fatalf("Read nested failed: %v", err)
		}
		if content != `{"deep": true}` {
			t.Errorf("expected {\"deep\": true}, got %s", content)
		}
	})

	t.Run("non-JSON file read", func(t *testing.T) {
		content, err := loader.Read("readme.txt")
		if err != nil {
			t.Fatalf("Read txt failed: %v", err)
		}
		if content != "This is plain text" {
			t.Errorf("expected plain text, got %s", content)
		}
	})

	t.Run("non-JSON file readJSON returns null", func(t *testing.T) {
		vm := goja.New()
		vm.Set("fazt", vm.NewObject())
		InjectPrivateNamespace(vm, loader)

		val, err := vm.RunString(`fazt.private.readJSON('readme.txt')`)
		if err != nil {
			t.Fatalf("readJSON failed: %v", err)
		}
		if !goja.IsNull(val) {
			t.Errorf("expected null for non-JSON, got %v", val)
		}
	})

	t.Run("special characters in JSON", func(t *testing.T) {
		vm := goja.New()
		vm.Set("fazt", vm.NewObject())
		InjectPrivateNamespace(vm, loader)

		val, err := vm.RunString(`fazt.private.readJSON('special.json').emoji`)
		if err != nil {
			t.Fatalf("readJSON failed: %v", err)
		}
		if val.String() != "🚀" {
			t.Errorf("expected emoji, got %s", val.String())
		}
	})

	t.Run("list includes nested files", func(t *testing.T) {
		files := loader.List()
		found := make(map[string]bool)
		for _, f := range files {
			found[f] = true
		}
		if !found["data/nested/deep.json"] {
			t.Error("nested file not in list")
		}
	})
}

func TestPrivateFileLoader_PathTraversalVariants(t *testing.T) {
	db := setupPrivateTestDB(t)
	defer db.Close()

	// Create a file outside private that we should NOT be able to access
	createPrivateFile(t, db, "test-app", "api/main.js", `function handler() {}`)
	createPrivateFile(t, db, "test-app", "private/legit.json", `{"ok": true}`)

	loader := NewPrivateFileLoader(db, "test-app")

	traversalAttempts := []string{
		"../api/main.js",
		"..%2fapi/main.js",
		"....//api/main.js",
		"./../api/main.js",
		".//../api/main.js",
		"private/../api/main.js",
		"/../../api/main.js",
		"..\\api\\main.js",
	}

	for _, attempt := range traversalAttempts {
		t.Run("traversal_"+attempt, func(t *testing.T) {
			_, err := loader.Read(attempt)
			if err == nil {
				t.Errorf("path traversal succeeded for %q", attempt)
			}
		})

		t.Run("exists_"+attempt, func(t *testing.T) {
			if loader.Exists(attempt) {
				t.Errorf("exists returned true for traversal %q", attempt)
			}
		})
	}
}

func TestPrivateFileLoader_AppIsolation(t *testing.T) {
	db := setupPrivateTestDB(t)
	defer db.Close()

	// Create files in two different apps
	createPrivateFile(t, db, "app-a", "private/secret.json", `{"app": "A"}`)
	createPrivateFile(t, db, "app-b", "private/secret.json", `{"app": "B"}`)

	loaderA := NewPrivateFileLoader(db, "app-a")
	loaderB := NewPrivateFileLoader(db, "app-b")

	t.Run("app A reads its own file", func(t *testing.T) {
		content, err := loaderA.Read("secret.json")
		if err != nil {
			t.Fatalf("Read failed: %v", err)
		}
		if content != `{"app": "A"}` {
			t.Errorf("wrong content: %s", content)
		}
	})

	t.Run("app B reads its own file", func(t *testing.T) {
		content, err := loaderB.Read("secret.json")
		if err != nil {
			t.Fatalf("Read failed: %v", err)
		}
		if content != `{"app": "B"}` {
			t.Errorf("wrong content: %s", content)
		}
	})

	t.Run("app A cannot see app B files in list", func(t *testing.T) {
		files := loaderA.List()
		if len(files) != 1 {
			t.Errorf("expected 1 file, got %d", len(files))
		}
	})
}

func TestPrivateFileLoader_NoArgumentHandling(t *testing.T) {
	db := setupPrivateTestDB(t)
	defer db.Close()

	loader := NewPrivateFileLoader(db, "test-app")

	vm := goja.New()
	vm.Set("fazt", vm.NewObject())
	InjectPrivateNamespace(vm, loader)

	t.Run("read with no args returns undefined", func(t *testing.T) {
		val, err := vm.RunString(`fazt.private.read() === undefined`)
		if err != nil {
			t.Fatalf("failed: %v", err)
		}
		if !val.ToBoolean() {
			t.Error("expected undefined")
		}
	})

	t.Run("readJSON with no args returns null", func(t *testing.T) {
		val, err := vm.RunString(`fazt.private.readJSON() === null`)
		if err != nil {
			t.Fatalf("failed: %v", err)
		}
		if !val.ToBoolean() {
			t.Error("expected null")
		}
	})

	t.Run("exists with no args returns false", func(t *testing.T) {
		val, err := vm.RunString(`fazt.private.exists()`)
		if err != nil {
			t.Fatalf("failed: %v", err)
		}
		if val.ToBoolean() {
			t.Error("expected false")
		}
	})
}

// Suppress unused import warning
var _ = bytes.NewReader
