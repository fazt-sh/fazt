package assets

import (
	"fmt"
	"io/fs"
)

// GetTemplate returns a filesystem for the named template.
// Template files may contain {{.Name}} placeholders for substitution.
func GetTemplate(name string) (fs.FS, error) {
	sub, err := fs.Sub(TemplatesFS, "templates/"+name)
	if err != nil {
		return nil, fmt.Errorf("template not found: %s", name)
	}

	// Verify the template directory exists by trying to read it
	entries, err := fs.ReadDir(sub, ".")
	if err != nil || len(entries) == 0 {
		return nil, fmt.Errorf("template not found: %s", name)
	}

	return sub, nil
}

// ListTemplates returns the names of all available templates.
func ListTemplates() []string {
	entries, err := fs.ReadDir(TemplatesFS, "templates")
	if err != nil {
		return nil
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names
}

// TemplateExists checks if a template with the given name exists.
func TemplateExists(name string) bool {
	_, err := GetTemplate(name)
	return err == nil
}
