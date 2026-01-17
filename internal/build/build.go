package build

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

// ErrBuildRequired is returned when an app requires building but cannot be built
var ErrBuildRequired = errors.New("app requires building but no package manager available and no pre-built dist/ exists")

// Result contains information about a build operation
type Result struct {
	OutputDir string // Absolute path to deployable folder
	Method    string // How it was built: "bun", "pnpm", "npm", "yarn", "existing", "source"
	PkgMgr    string // Which package manager was used (if any)
	Files     int    // Number of files in output
}

// Options configures the build process
type Options struct {
	Verbose bool // Print build output
}

// Build prepares an app directory for deployment.
// Returns the path to the deployable directory.
func Build(srcDir string, opts *Options) (*Result, error) {
	if opts == nil {
		opts = &Options{}
	}

	srcDir, err := filepath.Abs(srcDir)
	if err != nil {
		return nil, fmt.Errorf("invalid source directory: %w", err)
	}

	// Check for package.json with build script
	pkgPath := filepath.Join(srcDir, "package.json")
	if hasBuildScript(pkgPath) {
		return buildWithPackageManager(srcDir, opts)
	}

	// No build script - use source directly
	return useSource(srcDir)
}

// hasBuildScript checks if package.json exists and has a "build" script
func hasBuildScript(pkgPath string) bool {
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return false
	}

	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return false
	}

	_, has := pkg.Scripts["build"]
	return has
}

// buildWithPackageManager runs the build using an available package manager
func buildWithPackageManager(srcDir string, opts *Options) (*Result, error) {
	pm := DetectPackageManager(srcDir)
	if pm == nil {
		// No package manager - check for existing dist/
		return useExistingBuild(srcDir)
	}

	if opts.Verbose {
		fmt.Printf("Using %s to build...\n", pm.Name)
	}

	// Run install if node_modules missing
	nodeModules := filepath.Join(srcDir, "node_modules")
	if _, err := os.Stat(nodeModules); os.IsNotExist(err) {
		if opts.Verbose {
			fmt.Printf("Installing dependencies with %s...\n", pm.Name)
		}
		cmd := exec.Command(pm.Binary, pm.InstallCmd...)
		cmd.Dir = srcDir
		if opts.Verbose {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("%s install failed: %w", pm.Name, err)
		}
	}

	// Run build
	if opts.Verbose {
		fmt.Printf("Building with %s...\n", pm.Name)
	}
	cmd := exec.Command(pm.Binary, pm.BuildCmd...)
	cmd.Dir = srcDir
	if opts.Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s run build failed: %w", pm.Name, err)
	}

	// Find output directory (dist/ or build/)
	outputDir := findBuildOutput(srcDir)
	if outputDir == "" {
		return nil, fmt.Errorf("build succeeded but no output directory found (expected dist/ or build/)")
	}

	return &Result{
		OutputDir: outputDir,
		Method:    pm.Name,
		PkgMgr:    pm.Name,
		Files:     countFiles(outputDir),
	}, nil
}

// useExistingBuild uses an existing dist/ or build/ directory
func useExistingBuild(srcDir string) (*Result, error) {
	outputDir := findBuildOutput(srcDir)
	if outputDir == "" {
		// No existing build - this is an error for apps that need building
		return nil, ErrBuildRequired
	}

	return &Result{
		OutputDir: outputDir,
		Method:    "existing",
		Files:     countFiles(outputDir),
	}, nil
}

// useSource returns the source directory as the deployable output
// Used for simple apps without a build step
func useSource(srcDir string) (*Result, error) {
	return &Result{
		OutputDir: srcDir,
		Method:    "source",
		Files:     countFiles(srcDir),
	}, nil
}

// findBuildOutput locates the build output directory
func findBuildOutput(srcDir string) string {
	// Check common build output directories in priority order
	candidates := []string{"dist", "build", "out", ".output"}
	for _, dir := range candidates {
		path := filepath.Join(srcDir, dir)
		// Check if directory exists and has index.html or manifest.json
		if _, err := os.Stat(path); err == nil {
			indexPath := filepath.Join(path, "index.html")
			manifestPath := filepath.Join(path, "manifest.json")
			if _, err := os.Stat(indexPath); err == nil {
				return path
			}
			if _, err := os.Stat(manifestPath); err == nil {
				return path
			}
		}
	}
	return ""
}

// countFiles counts the number of files in a directory
func countFiles(dir string) int {
	count := 0
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			// Skip node_modules and hidden files
			rel, _ := filepath.Rel(dir, path)
			if len(rel) > 0 && rel[0] != '.' && !contains(rel, "node_modules") {
				count++
			}
		}
		return nil
	})
	return count
}

// contains checks if a path contains a directory name
func contains(path, dir string) bool {
	for _, part := range filepath.SplitList(path) {
		if part == dir {
			return true
		}
	}
	// Also check with separator
	return filepath.Base(path) == dir || filepath.Dir(path) == dir ||
		len(path) > len(dir) && path[:len(dir)+1] == dir+string(filepath.Separator)
}

// NeedsBuild checks if a directory needs a build step
func NeedsBuild(srcDir string) bool {
	pkgPath := filepath.Join(srcDir, "package.json")
	return hasBuildScript(pkgPath)
}
