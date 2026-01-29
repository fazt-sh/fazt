// Package gaps provides gap discovery and tracking for the test harness.
package gaps

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"
)

// Severity levels for gaps.
type Severity string

const (
	SeverityCritical Severity = "critical" // Block release
	SeverityHigh     Severity = "high"     // Fix soon
	SeverityMedium   Severity = "medium"   // Backlog
	SeverityLow      Severity = "low"      // Nice to have
)

// Category types for gaps.
type Category string

const (
	CategorySecurity    Category = "security"
	CategoryPerformance Category = "performance"
	CategoryBehavior    Category = "behavior"
	CategorySpec        Category = "spec"
)

// Gap represents a discovered gap in fazt's implementation.
type Gap struct {
	ID           string    `json:"id"`
	Category     Category  `json:"category"`
	Severity     Severity  `json:"severity"`
	Description  string    `json:"description"`
	DiscoveredBy string    `json:"discovered_by"` // Test name that found it
	SpecRef      string    `json:"spec_ref"`      // Link to relevant spec
	Remediation  string    `json:"remediation"`   // Suggested fix
	DiscoveredAt time.Time `json:"discovered_at"`
	Resolved     bool      `json:"resolved"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty"`
}

// Tracker manages discovered gaps.
type Tracker struct {
	gaps    map[string]*Gap
	counter int
	mu      sync.RWMutex
}

// NewTracker creates a new gap tracker.
func NewTracker() *Tracker {
	return &Tracker{
		gaps: make(map[string]*Gap),
	}
}

// Record adds a new gap to the tracker.
func (t *Tracker) Record(g Gap) string {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.counter++
	g.ID = fmt.Sprintf("GAP-%03d", t.counter)
	g.DiscoveredAt = time.Now()
	t.gaps[g.ID] = &g
	return g.ID
}

// RecordFromTest records a gap discovered during a test.
func (t *Tracker) RecordFromTest(testName string, category Category, severity Severity, description string) string {
	return t.Record(Gap{
		Category:     category,
		Severity:     severity,
		Description:  description,
		DiscoveredBy: testName,
	})
}

// Get returns a gap by ID.
func (t *Tracker) Get(id string) (*Gap, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	g, ok := t.gaps[id]
	return g, ok
}

// Resolve marks a gap as resolved.
func (t *Tracker) Resolve(id string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	g, ok := t.gaps[id]
	if !ok {
		return false
	}
	g.Resolved = true
	now := time.Now()
	g.ResolvedAt = &now
	return true
}

// All returns all gaps.
func (t *Tracker) All() []*Gap {
	t.mu.RLock()
	defer t.mu.RUnlock()

	gaps := make([]*Gap, 0, len(t.gaps))
	for _, g := range t.gaps {
		gaps = append(gaps, g)
	}

	// Sort by severity (critical first), then ID
	sort.Slice(gaps, func(i, j int) bool {
		if gaps[i].Severity != gaps[j].Severity {
			return severityOrder(gaps[i].Severity) < severityOrder(gaps[j].Severity)
		}
		return gaps[i].ID < gaps[j].ID
	})

	return gaps
}

// Open returns all unresolved gaps.
func (t *Tracker) Open() []*Gap {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var gaps []*Gap
	for _, g := range t.gaps {
		if !g.Resolved {
			gaps = append(gaps, g)
		}
	}

	sort.Slice(gaps, func(i, j int) bool {
		if gaps[i].Severity != gaps[j].Severity {
			return severityOrder(gaps[i].Severity) < severityOrder(gaps[j].Severity)
		}
		return gaps[i].ID < gaps[j].ID
	})

	return gaps
}

// BySeverity returns gaps filtered by severity.
func (t *Tracker) BySeverity(severity Severity) []*Gap {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var gaps []*Gap
	for _, g := range t.gaps {
		if g.Severity == severity && !g.Resolved {
			gaps = append(gaps, g)
		}
	}

	sort.Slice(gaps, func(i, j int) bool {
		return gaps[i].ID < gaps[j].ID
	})

	return gaps
}

// ByCategory returns gaps filtered by category.
func (t *Tracker) ByCategory(category Category) []*Gap {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var gaps []*Gap
	for _, g := range t.gaps {
		if g.Category == category && !g.Resolved {
			gaps = append(gaps, g)
		}
	}

	sort.Slice(gaps, func(i, j int) bool {
		if gaps[i].Severity != gaps[j].Severity {
			return severityOrder(gaps[i].Severity) < severityOrder(gaps[j].Severity)
		}
		return gaps[i].ID < gaps[j].ID
	})

	return gaps
}

// Count returns gap counts by severity.
func (t *Tracker) Count() map[Severity]int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	counts := make(map[Severity]int)
	for _, g := range t.gaps {
		if !g.Resolved {
			counts[g.Severity]++
		}
	}
	return counts
}

// HasBlockers returns true if there are critical gaps.
func (t *Tracker) HasBlockers() bool {
	return len(t.BySeverity(SeverityCritical)) > 0
}

// ToJSON serializes all gaps to JSON.
func (t *Tracker) ToJSON() ([]byte, error) {
	return json.MarshalIndent(t.All(), "", "  ")
}

// ToMarkdown generates a markdown TODO list.
func (t *Tracker) ToMarkdown() string {
	var result string
	result = "# Fazt Hardening TODO\n\n"

	sections := []struct {
		severity Severity
		title    string
		desc     string
	}{
		{SeverityCritical, "Critical (Block Release)", "Must fix before release"},
		{SeverityHigh, "High (Fix Soon)", "Should fix in current cycle"},
		{SeverityMedium, "Medium (Backlog)", "Add to backlog"},
		{SeverityLow, "Low (Nice to Have)", "Consider if time permits"},
	}

	for _, section := range sections {
		gaps := t.BySeverity(section.severity)
		if len(gaps) == 0 {
			continue
		}

		result += fmt.Sprintf("## %s\n\n", section.title)
		for _, g := range gaps {
			checkbox := "[ ]"
			if g.Resolved {
				checkbox = "[x]"
			}
			result += fmt.Sprintf("- %s %s: %s\n", checkbox, g.ID, g.Description)
			if g.Remediation != "" {
				result += fmt.Sprintf("  - Fix: %s\n", g.Remediation)
			}
			if g.SpecRef != "" {
				result += fmt.Sprintf("  - Spec: %s\n", g.SpecRef)
			}
		}
		result += "\n"
	}

	return result
}

// LoadJSON loads gaps from JSON.
func (t *Tracker) LoadJSON(data []byte) error {
	var gaps []*Gap
	if err := json.Unmarshal(data, &gaps); err != nil {
		return err
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	for _, g := range gaps {
		t.gaps[g.ID] = g
		// Update counter to avoid ID collisions
		var num int
		if _, err := fmt.Sscanf(g.ID, "GAP-%d", &num); err == nil {
			if num > t.counter {
				t.counter = num
			}
		}
	}

	return nil
}

func severityOrder(s Severity) int {
	switch s {
	case SeverityCritical:
		return 0
	case SeverityHigh:
		return 1
	case SeverityMedium:
		return 2
	case SeverityLow:
		return 3
	default:
		return 4
	}
}
