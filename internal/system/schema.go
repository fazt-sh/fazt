package system

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// FieldSchema describes one leaf field's metadata for the admin UI / API docs.
type FieldSchema struct {
	Label    string `json:"label"`
	Desc     string `json:"desc"`
	Unit     string `json:"unit,omitempty"`
	Min      *int64 `json:"min,omitempty"`
	Max      *int64 `json:"max,omitempty"`
	ReadOnly bool   `json:"read_only"`
}

var (
	schemaOnce sync.Once
	schemaJSON []byte
)

// GetSchemaJSON returns the cached JSON schema for Limits.
// Built once via sync.Once, subsequent calls return the pre-built bytes.
func GetSchemaJSON() []byte {
	schemaOnce.Do(func() {
		schema := buildSchema(reflect.TypeOf(Limits{}))
		schemaJSON, _ = json.Marshal(schema)
	})
	return schemaJSON
}

// buildSchema walks a struct type and extracts tag metadata into nested maps.
func buildSchema(t reflect.Type) map[string]interface{} {
	result := make(map[string]interface{})

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Determine JSON key
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}
		key := jsonTag
		if key == "" {
			key = field.Name
		}
		// Strip options like ",omitempty"
		if idx := strings.Index(key, ","); idx != -1 {
			key = key[:idx]
		}

		// If the field is a struct, recurse
		if field.Type.Kind() == reflect.Struct {
			result[key] = buildSchema(field.Type)
			continue
		}

		// Leaf field — extract tag metadata
		fs := FieldSchema{
			Label:    field.Tag.Get("label"),
			Desc:     field.Tag.Get("desc"),
			Unit:     field.Tag.Get("unit"),
			ReadOnly: field.Tag.Get("readonly") == "true",
		}

		// Parse range tag "min,max"
		if rangeTag := field.Tag.Get("range"); rangeTag != "" {
			parts := strings.SplitN(rangeTag, ",", 2)
			if len(parts) == 2 {
				if minVal, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
					fs.Min = &minVal
				}
				if maxVal, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
					fs.Max = &maxVal
				}
			}
		}

		result[key] = fs
	}

	return result
}
