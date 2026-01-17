package build

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHasBuildScript(t *testing.T) {
	t.Run("returns true when build script exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		pkgPath := filepath.Join(tmpDir, "package.json")
		os.WriteFile(pkgPath, []byte(`{"scripts":{"build":"vite build"}}`), 0644)

		if !hasBuildScript(pkgPath) {
			t.Error("expected hasBuildScript to return true")
		}
	})

	t.Run("returns false when no build script", func(t *testing.T) {
		tmpDir := t.TempDir()
		pkgPath := filepath.Join(tmpDir, "package.json")
		os.WriteFile(pkgPath, []byte(`{"scripts":{"start":"node index.js"}}`), 0644)

		if hasBuildScript(pkgPath) {
			t.Error("expected hasBuildScript to return false")
		}
	})

	t.Run("returns false when no package.json", func(t *testing.T) {
		tmpDir := t.TempDir()
		pkgPath := filepath.Join(tmpDir, "package.json")

		if hasBuildScript(pkgPath) {
			t.Error("expected hasBuildScript to return false for missing file")
		}
	})
}

func TestBuild(t *testing.T) {
	t.Run("simple app returns source dir", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.WriteFile(filepath.Join(tmpDir, "index.html"), []byte("<h1>Hi</h1>"), 0644)
		os.WriteFile(filepath.Join(tmpDir, "manifest.json"), []byte(`{"name":"test"}`), 0644)

		result, err := Build(tmpDir, nil)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}

		if result.OutputDir != tmpDir {
			t.Errorf("expected OutputDir=%s, got %s", tmpDir, result.OutputDir)
		}
		if result.Method != "source" {
			t.Errorf("expected Method=source, got %s", result.Method)
		}
	})

	t.Run("app with existing dist uses dist", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create package.json with build script
		pkg := `{"scripts":{"build":"echo build"}}`
		os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkg), 0644)

		// Create dist/ with index.html
		distDir := filepath.Join(tmpDir, "dist")
		os.MkdirAll(distDir, 0755)
		os.WriteFile(filepath.Join(distDir, "index.html"), []byte("<h1>Built</h1>"), 0644)

		// Temporarily modify PATH to hide package managers
		origPath := os.Getenv("PATH")
		os.Setenv("PATH", "/nonexistent")
		defer os.Setenv("PATH", origPath)

		result, err := Build(tmpDir, nil)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}

		if result.OutputDir != distDir {
			t.Errorf("expected OutputDir=%s, got %s", distDir, result.OutputDir)
		}
		if result.Method != "existing" {
			t.Errorf("expected Method=existing, got %s", result.Method)
		}
	})

	t.Run("app requiring build without pkg mgr fails", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create package.json with build script but no dist/
		pkg := `{"scripts":{"build":"vite build"}}`
		os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkg), 0644)
		os.WriteFile(filepath.Join(tmpDir, "index.html"), []byte("<h1>Source</h1>"), 0644)

		// Temporarily modify PATH to hide package managers
		origPath := os.Getenv("PATH")
		os.Setenv("PATH", "/nonexistent")
		defer os.Setenv("PATH", origPath)

		_, err := Build(tmpDir, nil)
		if err != ErrBuildRequired {
			t.Errorf("expected ErrBuildRequired, got %v", err)
		}
	})
}

func TestFindBuildOutput(t *testing.T) {
	t.Run("finds dist with index.html", func(t *testing.T) {
		tmpDir := t.TempDir()
		distDir := filepath.Join(tmpDir, "dist")
		os.MkdirAll(distDir, 0755)
		os.WriteFile(filepath.Join(distDir, "index.html"), []byte("<h1>Hi</h1>"), 0644)

		result := findBuildOutput(tmpDir)
		if result != distDir {
			t.Errorf("expected %s, got %s", distDir, result)
		}
	})

	t.Run("finds build with manifest.json", func(t *testing.T) {
		tmpDir := t.TempDir()
		buildDir := filepath.Join(tmpDir, "build")
		os.MkdirAll(buildDir, 0755)
		os.WriteFile(filepath.Join(buildDir, "manifest.json"), []byte(`{}`), 0644)

		result := findBuildOutput(tmpDir)
		if result != buildDir {
			t.Errorf("expected %s, got %s", buildDir, result)
		}
	})

	t.Run("returns empty for no build output", func(t *testing.T) {
		tmpDir := t.TempDir()

		result := findBuildOutput(tmpDir)
		if result != "" {
			t.Errorf("expected empty, got %s", result)
		}
	})
}

func TestNeedsBuild(t *testing.T) {
	t.Run("returns true when has build script", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{"scripts":{"build":"x"}}`), 0644)

		if !NeedsBuild(tmpDir) {
			t.Error("expected NeedsBuild to return true")
		}
	})

	t.Run("returns false when no build script", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{}`), 0644)

		if NeedsBuild(tmpDir) {
			t.Error("expected NeedsBuild to return false")
		}
	})
}
