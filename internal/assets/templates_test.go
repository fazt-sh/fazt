package assets

import (
	"io/fs"
	"testing"
)

func TestListTemplates(t *testing.T) {
	templates := ListTemplates()

	if len(templates) == 0 {
		t.Fatal("expected at least one template")
	}

	// Check for expected templates
	found := make(map[string]bool)
	for _, name := range templates {
		found[name] = true
	}

	expected := []string{"minimal", "vite", "static", "vue", "vue-api"}
	for _, name := range expected {
		if !found[name] {
			t.Errorf("expected '%s' template", name)
		}
	}
}

func TestGetTemplate(t *testing.T) {
	t.Run("minimal template has required files", func(t *testing.T) {
		tmplFS, err := GetTemplate("minimal")
		if err != nil {
			t.Fatalf("GetTemplate(minimal) failed: %v", err)
		}

		// Check manifest.json exists
		if _, err := fs.Stat(tmplFS, "manifest.json"); err != nil {
			t.Error("minimal template missing manifest.json")
		}

		// Check index.html exists
		if _, err := fs.Stat(tmplFS, "index.html"); err != nil {
			t.Error("minimal template missing index.html")
		}
	})

	t.Run("vite template has required files", func(t *testing.T) {
		tmplFS, err := GetTemplate("vite")
		if err != nil {
			t.Fatalf("GetTemplate(vite) failed: %v", err)
		}

		required := []string{
			"manifest.json",
			"index.html",
			"package.json",
			"vite.config.js",
			"src/main.js",
			"api/hello.js",
		}

		for _, f := range required {
			if _, err := fs.Stat(tmplFS, f); err != nil {
				t.Errorf("vite template missing %s", f)
			}
		}
	})

	t.Run("unknown template returns error", func(t *testing.T) {
		_, err := GetTemplate("nonexistent")
		if err == nil {
			t.Error("expected error for nonexistent template")
		}
	})
}

func TestTemplateExists(t *testing.T) {
	if !TemplateExists("minimal") {
		t.Error("TemplateExists(minimal) should return true")
	}
	if !TemplateExists("vite") {
		t.Error("TemplateExists(vite) should return true")
	}
	if !TemplateExists("static") {
		t.Error("TemplateExists(static) should return true")
	}
	if !TemplateExists("vue") {
		t.Error("TemplateExists(vue) should return true")
	}
	if !TemplateExists("vue-api") {
		t.Error("TemplateExists(vue-api) should return true")
	}
	if TemplateExists("nonexistent") {
		t.Error("TemplateExists(nonexistent) should return false")
	}
}

func TestNewTemplates(t *testing.T) {
	t.Run("static template has required files", func(t *testing.T) {
		tmplFS, err := GetTemplate("static")
		if err != nil {
			t.Fatalf("GetTemplate(static) failed: %v", err)
		}

		required := []string{"manifest.json", "index.html"}
		for _, f := range required {
			if _, err := fs.Stat(tmplFS, f); err != nil {
				t.Errorf("static template missing %s", f)
			}
		}
	})

	t.Run("vue template has required files", func(t *testing.T) {
		tmplFS, err := GetTemplate("vue")
		if err != nil {
			t.Fatalf("GetTemplate(vue) failed: %v", err)
		}

		required := []string{
			"manifest.json",
			"index.html",
			"package.json",
			"vite.config.js",
			"src/main.js",
		}

		for _, f := range required {
			if _, err := fs.Stat(tmplFS, f); err != nil {
				t.Errorf("vue template missing %s", f)
			}
		}
	})

	t.Run("vue-api template has required files", func(t *testing.T) {
		tmplFS, err := GetTemplate("vue-api")
		if err != nil {
			t.Fatalf("GetTemplate(vue-api) failed: %v", err)
		}

		required := []string{
			"manifest.json",
			"index.html",
			"package.json",
			"vite.config.js",
			"src/main.js",
			"src/lib/api.js",
			"src/lib/session.js",
			"src/lib/settings.js",
			"api/main.js",
		}

		for _, f := range required {
			if _, err := fs.Stat(tmplFS, f); err != nil {
				t.Errorf("vue-api template missing %s", f)
			}
		}
	})
}
