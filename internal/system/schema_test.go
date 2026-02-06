package system

import (
	"encoding/json"
	"testing"
)

func TestSchemaExtractsAllGroups(t *testing.T) {
	raw := GetSchemaJSON()
	if len(raw) == 0 {
		t.Fatal("schema JSON is empty")
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("failed to unmarshal schema: %v", err)
	}

	// All top-level groups should be present
	groups := []string{"hardware", "storage", "runtime", "capacity", "net"}
	for _, g := range groups {
		if _, ok := schema[g]; !ok {
			t.Errorf("missing group %q in schema", g)
		}
	}
}

func TestSchemaHardwareFieldsReadOnly(t *testing.T) {
	raw := GetSchemaJSON()
	var schema map[string]json.RawMessage
	json.Unmarshal(raw, &schema)

	var hw map[string]FieldSchema
	if err := json.Unmarshal(schema["hardware"], &hw); err != nil {
		t.Fatalf("failed to unmarshal hardware: %v", err)
	}

	// All hardware fields should be read-only
	for name, fs := range hw {
		if !fs.ReadOnly {
			t.Errorf("hardware.%s should be read_only", name)
		}
	}

	// Check specific fields exist
	if _, ok := hw["total_ram"]; !ok {
		t.Error("missing hardware.total_ram")
	}
	if _, ok := hw["cpu_cores"]; !ok {
		t.Error("missing hardware.cpu_cores")
	}
}

func TestSchemaRangeTagsParse(t *testing.T) {
	raw := GetSchemaJSON()
	var schema map[string]json.RawMessage
	json.Unmarshal(raw, &schema)

	var net map[string]FieldSchema
	if err := json.Unmarshal(schema["net"], &net); err != nil {
		t.Fatalf("failed to unmarshal net: %v", err)
	}

	// max_calls should have range 1,20
	mc, ok := net["max_calls"]
	if !ok {
		t.Fatal("missing net.max_calls")
	}
	if mc.Min == nil || *mc.Min != 1 {
		t.Errorf("net.max_calls min: got %v, want 1", mc.Min)
	}
	if mc.Max == nil || *mc.Max != 20 {
		t.Errorf("net.max_calls max: got %v, want 20", mc.Max)
	}
	if mc.ReadOnly {
		t.Error("net.max_calls should not be read_only")
	}
}

func TestSchemaLabelsAndDescs(t *testing.T) {
	raw := GetSchemaJSON()
	var schema map[string]json.RawMessage
	json.Unmarshal(raw, &schema)

	var storage map[string]FieldSchema
	json.Unmarshal(schema["storage"], &storage)

	vfs, ok := storage["max_vfs"]
	if !ok {
		t.Fatal("missing storage.max_vfs")
	}
	if vfs.Label != "VFS Cache" {
		t.Errorf("label: got %q, want %q", vfs.Label, "VFS Cache")
	}
	if vfs.Unit != "bytes" {
		t.Errorf("unit: got %q, want %q", vfs.Unit, "bytes")
	}
}

func TestSchemaNetPhase2Phase3Fields(t *testing.T) {
	raw := GetSchemaJSON()
	var schema map[string]json.RawMessage
	json.Unmarshal(raw, &schema)

	var net map[string]FieldSchema
	json.Unmarshal(schema["net"], &net)

	// Phase 2 fields
	if _, ok := net["rate_limit"]; !ok {
		t.Error("missing net.rate_limit")
	}
	if _, ok := net["rate_burst"]; !ok {
		t.Error("missing net.rate_burst")
	}

	// Phase 3 fields
	if _, ok := net["log_buffer"]; !ok {
		t.Error("missing net.log_buffer")
	}
	if _, ok := net["cache_max_items"]; !ok {
		t.Error("missing net.cache_max_items")
	}
}
