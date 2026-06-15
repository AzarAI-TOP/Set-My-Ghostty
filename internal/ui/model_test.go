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
