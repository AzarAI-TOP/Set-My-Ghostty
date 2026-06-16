package schema

import (
	"testing"
)

func TestDefaultKeymapActionsIsNotEmpty(t *testing.T) {
	actions := DefaultKeymapActions()
	if len(actions) == 0 {
		t.Fatal("DefaultKeymapActions returned empty list")
	}
}

func TestDefaultKeymapActionsNoEmptyFields(t *testing.T) {
	for _, a := range DefaultKeymapActions() {
		if a.Action == "" {
			t.Error("found KeymapAction with empty Action")
		}
		if a.Label == "" {
			t.Errorf("KeymapAction %q has empty Label", a.Action)
		}
		if a.Category == "" {
			t.Errorf("KeymapAction %q has empty Category", a.Action)
		}
	}
}

func TestDefaultKeymapActionsNoDuplicateActions(t *testing.T) {
	seen := make(map[string]int, len(DefaultKeymapActions()))
	for _, a := range DefaultKeymapActions() {
		seen[a.Action]++
	}
	for action, count := range seen {
		if count > 1 {
			t.Errorf("duplicate action %q appears %d times", action, count)
		}
	}
}

func TestDefaultKeymapActionsKnownCategories(t *testing.T) {
	valid := map[ActionGroup]bool{
		GroupClipboard: true, GroupSplits: true, GroupTabs: true,
		GroupNavigation: true, GroupTerminal: true, GroupView: true,
		GroupFont: true, GroupInspector: true, GroupOther: true,
	}
	for _, a := range DefaultKeymapActions() {
		if !valid[a.Category] {
			t.Errorf("action %q has unknown category %q", a.Action, a.Category)
		}
	}
}

func TestActionLabels(t *testing.T) {
	labels := ActionLabels()
	for _, a := range DefaultKeymapActions() {
		label, ok := labels[a.Action]
		if !ok {
			t.Errorf("ActionLabels missing %q", a.Action)
		}
		if label != a.Label {
			t.Errorf("ActionLabels[%q] = %q, want %q", a.Action, label, a.Label)
		}
	}
	// Should not contain extra keys.
	if len(labels) != len(DefaultKeymapActions()) {
		t.Errorf("ActionLabels has %d entries, want %d", len(labels), len(DefaultKeymapActions()))
	}
}

func TestActionCategories(t *testing.T) {
	cats := ActionCategories()
	for _, a := range DefaultKeymapActions() {
		cat, ok := cats[a.Action]
		if !ok {
			t.Errorf("ActionCategories missing %q", a.Action)
		}
		if cat != a.Category {
			t.Errorf("ActionCategories[%q] = %q, want %q", a.Action, cat, a.Category)
		}
	}
	if len(cats) != len(DefaultKeymapActions()) {
		t.Errorf("ActionCategories has %d entries, want %d", len(cats), len(DefaultKeymapActions()))
	}
}
