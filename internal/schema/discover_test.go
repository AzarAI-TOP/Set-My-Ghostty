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
