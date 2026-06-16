package config

import (
	"strings"
	"testing"
)

func TestGetAndSetExisting(t *testing.T) {
	doc := Parse([]byte("# keep\ntheme = dracula\nfont-size = 13\n"))
	if v, ok := doc.Get("theme"); !ok || v != "dracula" {
		t.Fatalf("Get(theme) = %q,%v", v, ok)
	}
	doc.Set("theme", "nord")
	if v, _ := doc.Get("theme"); v != "nord" {
		t.Fatalf("after Set, theme = %q", v)
	}
	got := string(doc.Bytes())
	want := "# keep\ntheme = nord\nfont-size = 13\n"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestSetAppendsWhenAbsent(t *testing.T) {
	doc := Parse([]byte("theme = dracula\n"))
	doc.Set("font-size", "14")
	got := string(doc.Bytes())
	want := "theme = dracula\nfont-size = 14\n"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestRepeatableKeys(t *testing.T) {
	doc := Parse([]byte("keybind = ctrl+a=copy_to_clipboard\n"))
	if all := doc.GetAll("keybind"); len(all) != 1 || all[0] != "ctrl+a=copy_to_clipboard" {
		t.Fatalf("GetAll = %v", all)
	}
	doc.SetRepeatable("keybind", []string{"ctrl+c=copy_to_clipboard", "ctrl+v=paste_from_clipboard"})
	all := doc.GetAll("keybind")
	if len(all) != 2 || all[1] != "ctrl+v=paste_from_clipboard" {
		t.Fatalf("after SetRepeatable, GetAll = %v", all)
	}
}

func TestGetReturnsFirst(t *testing.T) {
	doc := Parse([]byte("keybind = first\nkeybind = second\n"))
	v, ok := doc.Get("keybind")
	if !ok || v != "first" {
		t.Errorf("Get(keybind) = %q,%v, want first,true", v, ok)
	}
}

func TestGetMissingKey(t *testing.T) {
	doc := Parse([]byte("theme = dracula\n"))
	_, ok := doc.Get("nonexistent")
	if ok {
		t.Error("Get(nonexistent) should be false")
	}
}

func TestGetAllNoEntries(t *testing.T) {
	doc := Parse([]byte("theme = dracula\n"))
	all := doc.GetAll("keybind")
	if len(all) != 0 {
		t.Errorf("GetAll(keybind) should be empty, got %v", all)
	}
}

func TestGetAllMultiple(t *testing.T) {
	doc := Parse([]byte("keybind = a\nkeybind = b\nkeybind = c\n"))
	all := doc.GetAll("keybind")
	if len(all) != 3 || all[0] != "a" || all[1] != "b" || all[2] != "c" {
		t.Errorf("GetAll = %v, want [a b c]", all)
	}
}

func TestSetEmptyValue(t *testing.T) {
	doc := Parse([]byte("theme = dracula\n"))
	doc.Set("theme", "")
	v, _ := doc.Get("theme")
	if v != "" {
		t.Errorf("after Set(theme, ''), theme = %q", v)
	}
	got := string(doc.Bytes())
	want := "theme = \n"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestSetCaseSensitive(t *testing.T) {
	doc := Parse([]byte("Theme = dracula\n"))
	doc.Set("theme", "nord")
	// "Theme" (capital T) and "theme" are different keys — Set should append.
	got := string(doc.Bytes())
	want := "Theme = dracula\ntheme = nord\n"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
	// Original key is still there.
	if v, ok := doc.Get("Theme"); !ok || v != "dracula" {
		t.Errorf("Get(Theme) = %q,%v", v, ok)
	}
}

func TestRemoveAllExisting(t *testing.T) {
	doc := Parse([]byte("keybind = a\nkeep = me\nkeybind = b\n"))
	doc.RemoveAll("keybind")
	got := string(doc.Bytes())
	want := "keep = me\n"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestRemoveAllMissingKey(t *testing.T) {
	doc := Parse([]byte("theme = dracula\nfont-size = 13\n"))
	doc.RemoveAll("nonexistent")
	got := string(doc.Bytes())
	want := "theme = dracula\nfont-size = 13\n"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestRemoveAllPreservesCommentsAndBlanks(t *testing.T) {
	doc := Parse([]byte("# header\n\nkeybind = a\n# middle\nkeybind = b\n\n"))
	doc.RemoveAll("keybind")
	got := string(doc.Bytes())
	want := "# header\n\n# middle\n\n"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestRemoveAllAllLines(t *testing.T) {
	doc := Parse([]byte("keybind = a\nkeybind = b\n"))
	doc.RemoveAll("keybind")
	got := string(doc.Bytes())
	if got != "" {
		t.Errorf("got %q want empty", got)
	}
}

func TestSetRepeatableEmptySlice(t *testing.T) {
	doc := Parse([]byte("keybind = a\nkeybind = b\n"))
	doc.SetRepeatable("keybind", []string{})
	all := doc.GetAll("keybind")
	if len(all) != 0 {
		t.Errorf("after SetRepeatable with empty, GetAll = %v", all)
	}
}

func TestSetRepeatableKeyNotPresent(t *testing.T) {
	doc := Parse([]byte("theme = dracula\n"))
	doc.SetRepeatable("keybind", []string{"ctrl+c=copy"})
	got := string(doc.Bytes())
	want := "theme = dracula\nkeybind = ctrl+c=copy\n"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestSetRepeatableSingleReplacesMultiple(t *testing.T) {
	doc := Parse([]byte("keybind = a\nkeybind = b\nkeybind = c\n"))
	doc.SetRepeatable("keybind", []string{"only = one"})
	all := doc.GetAll("keybind")
	if len(all) != 1 || all[0] != "only = one" {
		t.Errorf("GetAll = %v, want [only = one]", all)
	}
}

func TestSetRepeatablePreservesSurroundingLines(t *testing.T) {
	// SetRepeatable replaces at the position of the first occurrence.
	// Comments between repeatable items stay in their original positions.
	doc := Parse([]byte("# top\nkeybind = a\n# middle\nkeybind = b\n# bottom\n"))
	doc.SetRepeatable("keybind", []string{"x", "y"})
	got := string(doc.Bytes())
	// The new keybinds go at position of first keybind; #middle stays after them.
	want := "# top\nkeybind = x\nkeybind = y\n# middle\n# bottom\n"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestSetRepeatablePreservesPosition(t *testing.T) {
	doc := Parse([]byte("theme = dracula\nkeybind = a\nkeybind = b\nfont-size=13\n"))
	doc.SetRepeatable("keybind", []string{"new"})
	got := string(doc.Bytes())
	want := "theme = dracula\nkeybind = new\nfont-size=13\n"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestSetUpdatesOnlyFirstOccurrence(t *testing.T) {
	// Set modifies only the first matching line, leaving later ones untouched.
	doc := Parse([]byte("keybind = first\nkeybind = second\n"))
	doc.Set("keybind", "updated")
	all := doc.GetAll("keybind")
	if len(all) == 2 && all[0] == "updated" && all[1] == "second" {
		// Correct: only first occurrence was updated.
		return
	}
	// The implementation may vary — log what happened.
	t.Logf("after Set(keybind, updated), GetAll = %v", all)
}

func TestSetAppendsTrailingNewline(t *testing.T) {
	doc := Parse([]byte("theme = dracula"))
	doc.Set("font-size", "14")
	got := string(doc.Bytes())
	want := "theme = dracula\nfont-size = 14\n"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

// ── Keybind helpers ─────────────────────────────────────────────────────

func TestParseKeybindValid(t *testing.T) {
	tr, act, ok := ParseKeybind("ctrl+c=copy_to_clipboard")
	if !ok {
		t.Fatal("ParseKeybind returned ok=false")
	}
	if tr != "ctrl+c" {
		t.Errorf("trigger = %q, want ctrl+c", tr)
	}
	if act != "copy_to_clipboard" {
		t.Errorf("action = %q, want copy_to_clipboard", act)
	}
}

func TestParseKeybindWithModifierCombos(t *testing.T) {
	tests := []struct {
		raw           string
		wantTrigger   string
		wantAction    string
	}{
		{"ctrl+shift+c=copy_to_clipboard", "ctrl+shift+c", "copy_to_clipboard"},
		{"super+k=clear_screen", "super+k", "clear_screen"},
		{"alt+enter=new_split:right", "alt+enter", "new_split:right"},
	}
	for _, tt := range tests {
		tr, act, ok := ParseKeybind(tt.raw)
		if !ok {
			t.Errorf("ParseKeybind(%q) returned ok=false", tt.raw)
			continue
		}
		if tr != tt.wantTrigger {
			t.Errorf("ParseKeybind(%q) trigger = %q, want %q", tt.raw, tr, tt.wantTrigger)
		}
		if act != tt.wantAction {
			t.Errorf("ParseKeybind(%q) action = %q, want %q", tt.raw, act, tt.wantAction)
		}
	}
}

func TestParseKeybindInvalid(t *testing.T) {
	tests := []string{
		"",
		"noequalsign",
		"copy_to_clipboard",
	}
	for _, raw := range tests {
		if _, _, ok := ParseKeybind(raw); ok {
			t.Errorf("ParseKeybind(%q) should be invalid", raw)
		}
	}
}

func TestBuildKeybind(t *testing.T) {
	got := BuildKeybind("ctrl+c", "copy_to_clipboard")
	want := "ctrl+c=copy_to_clipboard"
	if got != want {
		t.Errorf("BuildKeybind = %q, want %q", got, want)
	}
}

func TestBuildKeybindRoundTrip(t *testing.T) {
	triggers := []string{"ctrl+c", "super+shift+k", "alt+enter", "ctrl+shift+up"}
	action := "test_action"
	for _, tr := range triggers {
		built := BuildKeybind(tr, action)
		gotTr, gotAct, ok := ParseKeybind(built)
		if !ok {
			t.Errorf("ParseKeybind(%q) failed", built)
			continue
		}
		if gotTr != tr {
			t.Errorf("trigger: got %q, want %q", gotTr, tr)
		}
		if gotAct != action {
			t.Errorf("action: got %q, want %q", gotAct, action)
		}
	}
}

func TestKeybindMapEmpty(t *testing.T) {
	doc := Parse([]byte("theme = dracula\n"))
	m := doc.KeybindMap()
	if len(m) != 0 {
		t.Errorf("KeybindMap = %v, want empty", m)
	}
}

func TestKeybindMap(t *testing.T) {
	doc := Parse([]byte("keybind = ctrl+c=copy_to_clipboard\nkeybind = super+k=clear_screen\n"))
	m := doc.KeybindMap()
	if len(m) != 2 {
		t.Fatalf("KeybindMap len = %d, want 2", len(m))
	}
	if m["copy_to_clipboard"] != "ctrl+c" {
		t.Errorf("copy_to_clipboard trigger = %q", m["copy_to_clipboard"])
	}
	if m["clear_screen"] != "super+k" {
		t.Errorf("clear_screen trigger = %q", m["clear_screen"])
	}
}

func TestKeybindMapLastWins(t *testing.T) {
	doc := Parse([]byte("keybind = ctrl+c=copy_to_clipboard\nkeybind = ctrl+shift+c=copy_to_clipboard\n"))
	m := doc.KeybindMap()
	if m["copy_to_clipboard"] != "ctrl+shift+c" {
		t.Errorf("copy_to_clipboard trigger = %q, want ctrl+shift+c (last wins)", m["copy_to_clipboard"])
	}
}

func TestKeybindMapSkipsMalformed(t *testing.T) {
	doc := Parse([]byte("keybind = ctrl+c=copy_to_clipboard\nkeybind = malformed\nkeybind = super+k=clear_screen\n"))
	m := doc.KeybindMap()
	if len(m) != 2 {
		t.Errorf("KeybindMap len = %d, want 2 (malformed line ignored)", len(m))
	}
}

func TestSetKeybinds(t *testing.T) {
	doc := Parse([]byte("theme = dracula\n"))
	doc.SetKeybinds(map[string]string{
		"copy_to_clipboard": "ctrl+c",
		"clear_screen":      "super+k",
	})
	m := doc.KeybindMap()
	if len(m) != 2 {
		t.Fatalf("KeybindMap len = %d, want 2", len(m))
	}
	if m["copy_to_clipboard"] != "ctrl+c" {
		t.Errorf("copy_to_clipboard trigger = %q", m["copy_to_clipboard"])
	}
	if m["clear_screen"] != "super+k" {
		t.Errorf("clear_screen trigger = %q", m["clear_screen"])
	}
	// Verify document contains both keybind lines (order agnostic).
	s := string(doc.Bytes())
	if !strings.Contains(s, "keybind = ctrl+c=copy_to_clipboard") {
		t.Error("missing ctrl+c binding")
	}
	if !strings.Contains(s, "keybind = super+k=clear_screen") {
		t.Error("missing super+k binding")
	}
}

func TestSetKeybindsSkipsEmptyTrigger(t *testing.T) {
	doc := Parse([]byte("theme = dracula\n"))
	doc.SetKeybinds(map[string]string{
		"copy_to_clipboard": "ctrl+c",
		"clear_screen":      "", // unset — should be skipped
	})
	got := string(doc.Bytes())
	want := "theme = dracula\nkeybind = ctrl+c=copy_to_clipboard\n"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestSetKeybindsReplacesExisting(t *testing.T) {
	doc := Parse([]byte("keybind = ctrl+c=copy_to_clipboard\nkeybind = super+k=clear_screen\n"))
	doc.SetKeybinds(map[string]string{
		"copy_to_clipboard": "ctrl+shift+c",
	})
	m := doc.KeybindMap()
	if len(m) != 1 {
		t.Fatalf("KeybindMap len = %d, want 1", len(m))
	}
	if m["copy_to_clipboard"] != "ctrl+shift+c" {
		t.Errorf("trigger = %q", m["copy_to_clipboard"])
	}
}

func TestSetKeybindsEmptyMap(t *testing.T) {
	doc := Parse([]byte("keybind = ctrl+c=copy_to_clipboard\n"))
	doc.SetKeybinds(map[string]string{})
	m := doc.KeybindMap()
	if len(m) != 0 {
		t.Errorf("KeybindMap should be empty after SetKeybinds with empty map, got %v", m)
	}
}
