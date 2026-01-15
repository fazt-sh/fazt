package runtime

import (
	"io/fs"
	"sync"

	"github.com/fazt-sh/fazt/internal/assets"
)

var (
	stdlibCache map[string]string
	stdlibOnce  sync.Once
)

// stdlibModules lists all available stdlib modules
var stdlibModules = []string{
	"lodash",
	"cheerio",
	"uuid",
	"zod",
	"marked",
	"dayjs",
	"validator",
}

// loadStdlib initializes the stdlib cache from embedded files.
func loadStdlib() map[string]string {
	stdlibOnce.Do(func() {
		stdlibCache = make(map[string]string)
		for _, name := range stdlibModules {
			path := "stdlib/" + name + ".min.js"
			data, err := fs.ReadFile(assets.StdlibFS, path)
			if err != nil {
				// Skip libraries that failed to load
				continue
			}
			stdlibCache[name] = string(data)
		}
	})
	return stdlibCache
}

// GetStdlibModule returns a stdlib module source, or empty string if not found.
func GetStdlibModule(name string) (string, bool) {
	stdlib := loadStdlib()
	source, ok := stdlib[name]
	return source, ok
}

// IsStdlibModule checks if a name is a stdlib module.
func IsStdlibModule(name string) bool {
	_, ok := GetStdlibModule(name)
	return ok
}

// ListStdlibModules returns the list of available stdlib modules.
func ListStdlibModules() []string {
	return stdlibModules
}
