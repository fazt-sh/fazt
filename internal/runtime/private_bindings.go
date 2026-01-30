package runtime

import (
	"database/sql"
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
)

// PrivateFileLoader loads files from the private/ directory
type PrivateFileLoader struct {
	db    *sql.DB
	appID string
}

// NewPrivateFileLoader creates a new private file loader
func NewPrivateFileLoader(db *sql.DB, appID string) *PrivateFileLoader {
	return &PrivateFileLoader{db: db, appID: appID}
}

// InjectPrivateNamespace adds fazt.private.* to the VM
func InjectPrivateNamespace(vm *goja.Runtime, loader *PrivateFileLoader) error {
	// Get existing fazt object
	faztVal := vm.Get("fazt")
	if faztVal == nil || faztVal == goja.Undefined() {
		return nil // fazt namespace not set up yet
	}
	fazt, ok := faztVal.(*goja.Object)
	if !ok {
		return nil
	}

	private := vm.NewObject()

	// fazt.private.read(path) -> string
	private.Set("read", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return goja.Undefined()
		}
		path := call.Argument(0).String()
		content, err := loader.Read(path)
		if err != nil {
			return goja.Undefined()
		}
		return vm.ToValue(content)
	})

	// fazt.private.readJSON(path) -> object
	private.Set("readJSON", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return goja.Null()
		}
		path := call.Argument(0).String()
		content, err := loader.Read(path)
		if err != nil {
			return goja.Null()
		}
		var data interface{}
		if err := json.Unmarshal([]byte(content), &data); err != nil {
			return goja.Null()
		}
		return vm.ToValue(data)
	})

	// fazt.private.exists(path) -> bool
	private.Set("exists", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue(false)
		}
		path := call.Argument(0).String()
		exists := loader.Exists(path)
		return vm.ToValue(exists)
	})

	// fazt.private.list() -> []string
	private.Set("list", func(call goja.FunctionCall) goja.Value {
		files := loader.List()
		return vm.ToValue(files)
	})

	fazt.Set("private", private)
	return nil
}

// Read loads a file from private/ directory
func (l *PrivateFileLoader) Read(path string) (string, error) {
	// Sanitize path - prevent traversal
	path = filepath.Clean(path)
	if strings.Contains(path, "..") {
		return "", sql.ErrNoRows
	}

	// Build full path (files stored as "private/filename.json")
	fullPath := "private/" + strings.TrimPrefix(path, "/")

	var content string
	err := l.db.QueryRow(`
		SELECT content FROM files
		WHERE site_id = ? AND path = ?
	`, l.appID, fullPath).Scan(&content)
	return content, err
}

// Exists checks if a private file exists
func (l *PrivateFileLoader) Exists(path string) bool {
	path = filepath.Clean(path)
	if strings.Contains(path, "..") {
		return false
	}
	fullPath := "private/" + strings.TrimPrefix(path, "/")

	var exists bool
	l.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM files WHERE site_id = ? AND path = ?)
	`, l.appID, fullPath).Scan(&exists)
	return exists
}

// List returns all files in the private/ directory
func (l *PrivateFileLoader) List() []string {
	rows, err := l.db.Query(`
		SELECT path FROM files
		WHERE site_id = ? AND path LIKE 'private/%'
	`, l.appID)
	if err != nil {
		return []string{}
	}
	defer rows.Close()

	var files []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err == nil {
			// Strip "private/" prefix for user-facing paths
			files = append(files, strings.TrimPrefix(path, "private/"))
		}
	}
	return files
}
