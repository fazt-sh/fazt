package gaps

import (
	"testing"
)

func TestTracker_Record(t *testing.T) {
	tracker := NewTracker()

	id := tracker.Record(Gap{
		Category:    CategorySecurity,
		Severity:    SeverityCritical,
		Description: "Test gap",
	})

	if id != "GAP-001" {
		t.Errorf("expected GAP-001, got %s", id)
	}

	gap, ok := tracker.Get(id)
	if !ok {
		t.Error("gap not found")
	}
	if gap.Description != "Test gap" {
		t.Errorf("wrong description: %s", gap.Description)
	}
	if gap.DiscoveredAt.IsZero() {
		t.Error("discovered_at not set")
	}
}

func TestTracker_Severity(t *testing.T) {
	tracker := NewTracker()

	tracker.Record(Gap{Severity: SeverityCritical, Description: "Critical 1"})
	tracker.Record(Gap{Severity: SeverityHigh, Description: "High 1"})
	tracker.Record(Gap{Severity: SeverityCritical, Description: "Critical 2"})
	tracker.Record(Gap{Severity: SeverityLow, Description: "Low 1"})

	critical := tracker.BySeverity(SeverityCritical)
	if len(critical) != 2 {
		t.Errorf("expected 2 critical gaps, got %d", len(critical))
	}

	if !tracker.HasBlockers() {
		t.Error("expected HasBlockers to return true")
	}
}

func TestTracker_Resolve(t *testing.T) {
	tracker := NewTracker()

	id := tracker.Record(Gap{
		Category:    CategorySecurity,
		Severity:    SeverityCritical,
		Description: "Test gap",
	})

	if !tracker.HasBlockers() {
		t.Error("expected blockers before resolve")
	}

	tracker.Resolve(id)

	if tracker.HasBlockers() {
		t.Error("expected no blockers after resolve")
	}

	gap, _ := tracker.Get(id)
	if !gap.Resolved {
		t.Error("gap should be marked resolved")
	}
	if gap.ResolvedAt == nil {
		t.Error("resolved_at should be set")
	}
}

func TestTracker_Count(t *testing.T) {
	tracker := NewTracker()

	tracker.Record(Gap{Severity: SeverityCritical})
	tracker.Record(Gap{Severity: SeverityHigh})
	tracker.Record(Gap{Severity: SeverityHigh})
	tracker.Record(Gap{Severity: SeverityMedium})

	counts := tracker.Count()

	if counts[SeverityCritical] != 1 {
		t.Errorf("expected 1 critical, got %d", counts[SeverityCritical])
	}
	if counts[SeverityHigh] != 2 {
		t.Errorf("expected 2 high, got %d", counts[SeverityHigh])
	}
}

func TestTracker_ToMarkdown(t *testing.T) {
	tracker := NewTracker()

	tracker.Record(Gap{
		Category:    CategorySecurity,
		Severity:    SeverityCritical,
		Description: "Security issue",
		Remediation: "Fix the thing",
	})
	tracker.Record(Gap{
		Category: CategoryPerformance,
		Severity: SeverityMedium,
		Description: "Performance issue",
	})

	md := tracker.ToMarkdown()

	if md == "" {
		t.Error("markdown should not be empty")
	}
	if !contains(md, "Critical") {
		t.Error("markdown should contain Critical section")
	}
	if !contains(md, "Security issue") {
		t.Error("markdown should contain gap description")
	}
	if !contains(md, "Fix the thing") {
		t.Error("markdown should contain remediation")
	}
}

func TestTracker_JSON(t *testing.T) {
	tracker := NewTracker()

	tracker.Record(Gap{
		Category:    CategorySecurity,
		Severity:    SeverityCritical,
		Description: "Test gap",
	})

	json, err := tracker.ToJSON()
	if err != nil {
		t.Errorf("ToJSON failed: %v", err)
	}

	// Load into new tracker
	tracker2 := NewTracker()
	if err := tracker2.LoadJSON(json); err != nil {
		t.Errorf("LoadJSON failed: %v", err)
	}

	gaps := tracker2.All()
	if len(gaps) != 1 {
		t.Errorf("expected 1 gap, got %d", len(gaps))
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(len(s) >= len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
