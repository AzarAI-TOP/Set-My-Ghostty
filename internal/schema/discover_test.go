package schema

import (
	"strings"
	"testing"
)

const sampleDocs = `# The color theme.
# Accepts a built-in name.
theme =

# Font size in points.
font-size = 13

keybind =
`

func TestParseShowConfigDocs(t *testing.T) {
	s, err := ParseShowConfigDocs(strings.NewReader(sampleDocs))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	theme, ok := s.Options["theme"]
	if !ok {
		t.Fatal("missing theme")
	}
	if !strings.Contains(theme.Doc, "color theme") {
		t.Errorf("theme doc not captured: %q", theme.Doc)
	}
	if s.Options["font-size"].Default != "13" {
		t.Errorf("font-size default = %q", s.Options["font-size"].Default)
	}
}

func TestMergeKeepsCuratedFlags(t *testing.T) {
	discovered, _ := ParseShowConfigDocs(strings.NewReader(sampleDocs))
	merged := Static().MergeFrom(discovered)
	// keybind stays repeatable (curated) but gains nothing harmful from discovery.
	if !merged.Options["keybind"].Repeatable {
		t.Errorf("merge lost keybind repeatable flag")
	}
	// discovered doc for theme overrides the static one.
	if !strings.Contains(merged.Options["theme"].Doc, "color theme") {
		t.Errorf("merge did not take discovered doc")
	}
}

func TestParseShowConfigDocsEmpty(t *testing.T) {
	s, err := ParseShowConfigDocs(strings.NewReader(""))
	if err != nil {
		t.Fatalf("parse empty: %v", err)
	}
	if len(s.Options) != 0 {
		t.Errorf("expected empty schema, got %d options", len(s.Options))
	}
}

func TestParseShowConfigDocsOnlyComments(t *testing.T) {
	input := `# Just a comment
# Another comment
`
	s, err := ParseShowConfigDocs(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(s.Options) != 0 {
		t.Errorf("expected no options from comments-only, got %d", len(s.Options))
	}
}

func TestParseShowConfigDocsNoDocComment(t *testing.T) {
	input := "theme = dracula\n"
	s, err := ParseShowConfigDocs(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	opt, ok := s.Options["theme"]
	if !ok {
		t.Fatal("missing theme")
	}
	if opt.Doc != "" {
		t.Errorf("expected empty doc, got %q", opt.Doc)
	}
	if opt.Default != "dracula" {
		t.Errorf("expected default dracula, got %q", opt.Default)
	}
}

func TestParseShowConfigDocsMultiLineDoc(t *testing.T) {
	input := `# First line of doc.
# Second line of doc.
# Third line of doc.
key = value
`
	s, err := ParseShowConfigDocs(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	opt := s.Options["key"]
	if !strings.Contains(opt.Doc, "First line") {
		t.Errorf("doc missing first line: %q", opt.Doc)
	}
	if !strings.Contains(opt.Doc, "Second line") {
		t.Errorf("doc missing second line: %q", opt.Doc)
	}
	if !strings.Contains(opt.Doc, "Third line") {
		t.Errorf("doc missing third line: %q", opt.Doc)
	}
}

func TestParseShowConfigDocsParagraphBreak(t *testing.T) {
	input := `# First paragraph.
#
# Second paragraph.
key = value
`
	s, err := ParseShowConfigDocs(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	opt := s.Options["key"]
	if !strings.Contains(opt.Doc, "First paragraph.") {
		t.Errorf("doc missing first paragraph: %q", opt.Doc)
	}
	if !strings.Contains(opt.Doc, "Second paragraph.") {
		t.Errorf("doc missing second paragraph: %q", opt.Doc)
	}
}

func TestParseShowConfigDocsMultipleKeys(t *testing.T) {
	input := `# First key
a = 1

# Second key
b = 2

# Third key
c = 3
`
	s, err := ParseShowConfigDocs(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(s.Options) != 3 {
		t.Errorf("expected 3 options, got %d", len(s.Options))
	}
	for _, k := range []string{"a", "b", "c"} {
		if _, ok := s.Options[k]; !ok {
			t.Errorf("missing key %q", k)
		}
	}
}

func TestParseShowConfigDocsKeyWithEqualsInValue(t *testing.T) {
	input := "keybind = ctrl+a=copy\n"
	s, err := ParseShowConfigDocs(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	// The parser splits on first '=' for the key=value in the show-config output.
	// But the ghostty output format for keybinds is "keybind = trigger=action",
	// so the "value" from SplitN(raw, "=", 2) captures everything after first '='.
	opt, ok := s.Options["keybind"]
	if !ok {
		t.Fatal("missing keybind")
	}
	if opt.Default != "ctrl+a=copy" {
		t.Errorf("default = %q, want ctrl+a=copy", opt.Default)
	}
}

func TestMergeFromEmptyDiscovered(t *testing.T) {
	curated := Static()
	discovered := &Schema{Options: map[string]Option{}}
	merged := curated.MergeFrom(discovered)
	if len(merged.Options) != len(curated.Options) {
		t.Errorf("merge with empty discovered changed option count: %d vs %d",
			len(merged.Options), len(curated.Options))
	}
	// All curated keys should still be present.
	for k, v := range curated.Options {
		mv, ok := merged.Options[k]
		if !ok {
			t.Errorf("merged missing key %q", k)
			continue
		}
		if mv.Type != v.Type {
			t.Errorf("key %q: type changed from %v to %v", k, v.Type, mv.Type)
		}
		if mv.Repeatable != v.Repeatable {
			t.Errorf("key %q: repeatable changed from %v to %v", k, v.Repeatable, mv.Repeatable)
		}
	}
}

func TestMergeFromOnlyDiscovered(t *testing.T) {
	curated := &Schema{Options: map[string]Option{}}
	discovered := &Schema{Options: map[string]Option{
		"newkey": {Key: "newkey", Default: "val"},
	}}
	merged := curated.MergeFrom(discovered)
	if len(merged.Options) != 1 {
		t.Errorf("expected 1 option, got %d", len(merged.Options))
	}
	opt, ok := merged.Options["newkey"]
	if !ok {
		t.Fatal("missing newkey")
	}
	if opt.Default != "val" {
		t.Errorf("default = %q", opt.Default)
	}
}

func TestMergeFromOverridesDefaultAndDoc(t *testing.T) {
	curated := &Schema{Options: map[string]Option{
		"key": {Key: "key", Default: "old", Doc: "old doc", Type: TypeInt},
	}}
	discovered := &Schema{Options: map[string]Option{
		"key": {Key: "key", Default: "new", Doc: "new doc"},
	}}
	merged := curated.MergeFrom(discovered)
	opt := merged.Options["key"]
	if opt.Default != "new" {
		t.Errorf("default = %q, want new", opt.Default)
	}
	if opt.Doc != "new doc" {
		t.Errorf("doc = %q, want new doc", opt.Doc)
	}
	if opt.Type != TypeInt {
		t.Errorf("type should remain TypeInt, got %v", opt.Type)
	}
}

func TestStaticHasAllTabKeys(t *testing.T) {
	s := Static()
	required := []struct {
		key        string
		wantType   Type
		repeatable bool
	}{
		{"theme", TypeString, false},
		{"background", TypeColor, false},
		{"foreground", TypeColor, false},
		{"background-opacity", TypeFloat, false},
		{"background-blur", TypeBool, false},
		{"palette", TypeString, true},
		{"window-padding-x", TypeInt, false},
		{"window-padding-y", TypeInt, false},
		{"font-family", TypeString, false},
		{"font-size", TypeInt, false},
		{"font-feature", TypeString, true},
		{"font-style", TypeString, false},
		{"adjust-cell-width", TypeString, false},
		{"adjust-cell-height", TypeString, false},
		{"keybind", TypeString, true},
		{"window-decoration", TypeEnum, false},
		{"confirm-close-surface", TypeBool, false},
		{"cursor-style", TypeEnum, false},
		{"cursor-style-blink", TypeBool, false},
		{"mouse-hide-while-typing", TypeBool, false},
		{"shell-integration", TypeEnum, false},
	}
	for _, r := range required {
		opt, ok := s.Options[r.key]
		if !ok {
			t.Errorf("static schema missing required key %q", r.key)
			continue
		}
		if opt.Type != r.wantType {
			t.Errorf("key %q: type = %v, want %v", r.key, opt.Type, r.wantType)
		}
		if opt.Repeatable != r.repeatable {
			t.Errorf("key %q: repeatable = %v, want %v", r.key, opt.Repeatable, r.repeatable)
		}
	}
}

func TestStaticEnumValues(t *testing.T) {
	s := Static()
	tests := []struct {
		key    string
		expect []string
	}{
		{"window-decoration", []string{"auto", "none", "client", "server"}},
		{"cursor-style", []string{"block", "bar", "underline", "block_hollow"}},
		{"shell-integration", []string{"none", "detect", "bash", "elvish", "fish", "zsh"}},
	}
	for _, tt := range tests {
		opt, ok := s.Options[tt.key]
		if !ok {
			t.Errorf("missing key %q", tt.key)
			continue
		}
		if len(opt.Enum) != len(tt.expect) {
			t.Errorf("key %q: enum length = %d, want %d", tt.key, len(opt.Enum), len(tt.expect))
			continue
		}
		for i, v := range tt.expect {
			if opt.Enum[i] != v {
				t.Errorf("key %q: enum[%d] = %q, want %q", tt.key, i, opt.Enum[i], v)
			}
		}
	}
}
