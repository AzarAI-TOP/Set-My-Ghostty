package ui

import (
	"testing"

	"github.com/AzarAI-TOP/Set-My-Ghostty/internal/config"
	"github.com/AzarAI-TOP/Set-My-Ghostty/internal/schema"
)

func newModel(t *testing.T, src string) *Model {
	t.Helper()
	return NewModel(config.Parse([]byte(src)), schema.Static())
}

func TestValueFallsBackToDefault(t *testing.T) {
	m := newModel(t, "")
	if got := m.Value("font-size"); got != "13" { // schema default
		t.Errorf("Value(font-size) = %q, want default 13", got)
	}
}

func TestValuePrefersDocThenPending(t *testing.T) {
	m := newModel(t, "font-size = 15\n")
	if got := m.Value("font-size"); got != "15" {
		t.Errorf("Value = %q want 15", got)
	}
	m.SetValue("font-size", "20")
	if got := m.Value("font-size"); got != "20" {
		t.Errorf("after SetValue = %q want 20", got)
	}
	if !m.Dirty() {
		t.Error("model should be dirty after SetValue")
	}
}

func TestApplyFlushesScalarsAndLists(t *testing.T) {
	m := newModel(t, "font-size = 15\n")
	m.SetValue("font-size", "20")
	m.SetList("keybind", []string{"ctrl+c=copy_to_clipboard"})
	m.Apply()
	out := string(m.Doc.Bytes())
	if out != "font-size = 20\nkeybind = ctrl+c=copy_to_clipboard\n" {
		t.Errorf("applied doc = %q", out)
	}
	if m.Dirty() {
		t.Error("model should be clean after Apply")
	}
}

func TestSetValueEqualToCurrentClearsPending(t *testing.T) {
	m := newModel(t, "font-size = 15\n")
	m.SetValue("font-size", "15") // same as doc value -> not a real change
	if m.Dirty() {
		t.Error("setting the same value should not mark dirty")
	}
}

func TestValueFallsThroughDocWhenNoDefault(t *testing.T) {
	m := newModel(t, "theme = dracula\n")
	got := m.Value("theme")
	if got != "dracula" {
		t.Errorf("Value(theme) = %q, want dracula", got)
	}
}

func TestValueUnknownKeyReturnsEmpty(t *testing.T) {
	m := newModel(t, "")
	got := m.Value("nonexistent_key")
	if got != "" {
		t.Errorf("Value(nonexistent) = %q, want empty", got)
	}
}

func TestListReturnsDocValues(t *testing.T) {
	m := newModel(t, "keybind = a\nkeybind = b\n")
	got := m.List("keybind")
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("List = %v, want [a b]", got)
	}
}

func TestListReturnsPendingWhenSet(t *testing.T) {
	m := newModel(t, "keybind = a\nkeybind = b\n")
	m.SetList("keybind", []string{"x", "y", "z"})
	got := m.List("keybind")
	if len(got) != 3 || got[0] != "x" || got[2] != "z" {
		t.Errorf("List = %v, want [x y z]", got)
	}
}

func TestListNoEntriesReturnsNil(t *testing.T) {
	m := newModel(t, "")
	got := m.List("keybind")
	if got == nil {
		// nil is acceptable — just check length 0
		return
	}
	if len(got) != 0 {
		t.Errorf("List = %v, want empty", got)
	}
}

func TestDirtyScalarAndList(t *testing.T) {
	m := newModel(t, "")
	if m.Dirty() {
		t.Error("fresh model should be clean")
	}
	m.SetValue("theme", "nord")
	if !m.Dirty() {
		t.Error("model should be dirty after scalar set")
	}
	m.Apply()
	if m.Dirty() {
		t.Error("model should be clean after apply")
	}
	m.SetList("keybind", []string{"a"})
	if !m.Dirty() {
		t.Error("model should be dirty after list set")
	}
}

func TestApplyWithNoPendingEdits(t *testing.T) {
	m := newModel(t, "theme = nord\n")
	m.Apply() // no-op
	if m.Dirty() {
		t.Error("model should be clean after no-op apply")
	}
	out := string(m.Doc.Bytes())
	if out != "theme = nord\n" {
		t.Errorf("doc = %q, should be unchanged", out)
	}
}

func TestApplyClearsPendingMaps(t *testing.T) {
	m := newModel(t, "")
	m.SetValue("a", "1")
	m.SetList("b", []string{"2"})
	m.Apply()
	// After Apply, pending maps should be fresh.
	if len(m.pendingScalar) != 0 {
		t.Errorf("pendingScalar len = %d, want 0", len(m.pendingScalar))
	}
	if len(m.pendingList) != 0 {
		t.Errorf("pendingList len = %d, want 0", len(m.pendingList))
	}
}

func TestApplyScalarReplacesExisting(t *testing.T) {
	m := newModel(t, "theme = old\n")
	m.SetValue("theme", "new")
	m.Apply()
	v, _ := m.Doc.Get("theme")
	if v != "new" {
		t.Errorf("theme = %q, want new", v)
	}
}

func TestApplyScalarAppendsNewKey(t *testing.T) {
	m := newModel(t, "theme = nord\n")
	m.SetValue("font-size", "14")
	m.Apply()
	out := string(m.Doc.Bytes())
	if out != "theme = nord\nfont-size = 14\n" {
		t.Errorf("doc = %q", out)
	}
}

func TestApplyListReplacesAll(t *testing.T) {
	m := newModel(t, "keybind = a\nkeybind = b\n")
	m.SetList("keybind", []string{"c"})
	m.Apply()
	all := m.Doc.GetAll("keybind")
	if len(all) != 1 || all[0] != "c" {
		t.Errorf("keybind = %v, want [c]", all)
	}
}

func TestApplyListEmptyRemovesAll(t *testing.T) {
	m := newModel(t, "keybind = a\nkeybind = b\n")
	m.SetList("keybind", []string{})
	m.Apply()
	all := m.Doc.GetAll("keybind")
	if len(all) != 0 {
		t.Errorf("keybind = %v, want empty", all)
	}
}

func TestModelInteractionScalarThenListSameKey(t *testing.T) {
	// Setting a scalar then a list for the same key — list should win.
	m := newModel(t, "")
	m.SetValue("keybind", "scalar_val")
	m.SetList("keybind", []string{"list_val"})
	m.Apply()
	all := m.Doc.GetAll("keybind")
	if len(all) != 1 || all[0] != "list_val" {
		t.Errorf("keybind = %v, want [list_val]", all)
	}
	// Scalar should not have been set since keybind is repeatable — but
	// model doesn't know about schema, both maps are independent.
	// SetValue writes to pendingScalar, SetList writes to pendingList.
	// Apply() applies scalars first, then lists — so list overwrites.
	// The scalar set is still there but the list overwrites it.
	scalarV := m.pendingScalar["keybind"]
	if scalarV != "" {
		// Check if scalar was written — it was, but Apply applied list after.
		t.Logf("pendingScalar[keybind] = %q (applied but overwritten by list)", scalarV)
	}
}

// ── Keybind Model tests ─────────────────────────────────────────────────

func TestKeybindFromDoc(t *testing.T) {
	m := newModel(t, "keybind = ctrl+c=copy_to_clipboard\n")
	got := m.Keybind("copy_to_clipboard")
	if got != "ctrl+c" {
		t.Errorf("Keybind = %q, want ctrl+c", got)
	}
}

func TestKeybindFromDocMissing(t *testing.T) {
	m := newModel(t, "keybind = ctrl+c=copy_to_clipboard\n")
	got := m.Keybind("nonexistent_action")
	if got != "" {
		t.Errorf("Keybind(nonexistent) = %q, want empty", got)
	}
}

func TestKeybindPrefersPending(t *testing.T) {
	m := newModel(t, "keybind = ctrl+c=copy_to_clipboard\n")
	m.SetKeybind("copy_to_clipboard", "ctrl+shift+c")
	got := m.Keybind("copy_to_clipboard")
	if got != "ctrl+shift+c" {
		t.Errorf("Keybind = %q, want ctrl+shift+c", got)
	}
}

func TestSetKeybindMakesDirty(t *testing.T) {
	m := newModel(t, "")
	if m.Dirty() {
		t.Error("fresh model should be clean")
	}
	m.SetKeybind("copy_to_clipboard", "ctrl+c")
	if !m.Dirty() {
		t.Error("model should be dirty after SetKeybind")
	}
}

func TestClearKeybindRemovesPending(t *testing.T) {
	m := newModel(t, "")
	m.SetKeybind("copy_to_clipboard", "ctrl+c")
	m.ClearKeybind("copy_to_clipboard")
	if m.Dirty() {
		t.Error("model should be clean after ClearKeybind of only pending")
	}
	got := m.Keybind("copy_to_clipboard")
	if got != "" {
		t.Errorf("Keybind = %q after Clear, want empty", got)
	}
}

func TestKeybindApplyFlushesToDoc(t *testing.T) {
	m := newModel(t, "")
	m.SetKeybind("copy_to_clipboard", "ctrl+c")
	m.SetKeybind("clear_screen", "super+k")
	m.Apply()
	m2 := m.Doc.KeybindMap()
	if m2["copy_to_clipboard"] != "ctrl+c" {
		t.Errorf("copy_to_clipboard trigger = %q", m2["copy_to_clipboard"])
	}
	if m2["clear_screen"] != "super+k" {
		t.Errorf("clear_screen trigger = %q", m2["clear_screen"])
	}
}

func TestKeybindApplyOverwritesExisting(t *testing.T) {
	m := newModel(t, "keybind = ctrl+c=old_action\n")
	m.SetKeybind("new_action", "ctrl+c")
	m.Apply()
	m2 := m.Doc.KeybindMap()
	// ctrl+c should now point to new_action, not old_action
	if m2["new_action"] != "ctrl+c" {
		t.Errorf("new_action trigger = %q", m2["new_action"])
	}
	if _, ok := m2["old_action"]; ok {
		t.Error("old_action should have been removed")
	}
}

func TestKeybindApplyEmptyTriggerUnsets(t *testing.T) {
	m := newModel(t, "keybind = ctrl+c=copy_to_clipboard\n")
	m.SetKeybind("copy_to_clipboard", "") // empty trigger = unset
	m.Apply()
	m2 := m.Doc.KeybindMap()
	if _, ok := m2["copy_to_clipboard"]; ok {
		t.Error("copy_to_clipboard should have been removed after setting empty trigger")
	}
}

func TestKeybindApplyCleansPending(t *testing.T) {
	m := newModel(t, "")
	m.SetKeybind("copy_to_clipboard", "ctrl+c")
	m.Apply()
	if m.Dirty() {
		t.Error("model should be clean after Apply")
	}
	if len(m.pendingKeybinds) != 0 {
		t.Errorf("pendingKeybinds len = %d, want 0", len(m.pendingKeybinds))
	}
}
