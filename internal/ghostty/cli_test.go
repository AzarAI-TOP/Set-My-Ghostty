package ghostty

import (
	"testing"
)

func TestUnavailableWhenNoPath(t *testing.T) {
	c := &CLI{Path: ""}
	if c.Available() {
		t.Error("CLI with empty path should be unavailable")
	}
}

func TestAvailableWhenPathSet(t *testing.T) {
	c := &CLI{Path: "/usr/bin/ghostty"}
	if !c.Available() {
		t.Error("CLI with path should be available")
	}
}

func TestValidateConfigUnavailable(t *testing.T) {
	c := &CLI{}
	ok, out, err := c.ValidateConfig("/any/path")
	if ok {
		t.Error("should not be ok when unavailable")
	}
	if out != "" {
		t.Errorf("output should be empty, got %q", out)
	}
	if err != errUnavailable {
		t.Errorf("err = %v, want %v", err, errUnavailable)
	}
}

func TestListThemesUnavailable(t *testing.T) {
	c := &CLI{}
	themes, err := c.ListThemes()
	if themes != nil {
		t.Errorf("themes should be nil, got %v", themes)
	}
	if err != errUnavailable {
		t.Errorf("err = %v, want %v", err, errUnavailable)
	}
}

func TestListFontsUnavailable(t *testing.T) {
	c := &CLI{}
	fonts, err := c.ListFonts()
	if fonts != nil {
		t.Errorf("fonts should be nil, got %v", fonts)
	}
	if err != errUnavailable {
		t.Errorf("err = %v, want %v", err, errUnavailable)
	}
}

func TestShowConfigDocsUnavailable(t *testing.T) {
	c := &CLI{}
	out, err := c.ShowConfigDocs()
	if out != "" {
		t.Errorf("output should be empty, got %q", out)
	}
	if err != errUnavailable {
		t.Errorf("err = %v, want %v", err, errUnavailable)
	}
}

func TestParseList(t *testing.T) {
	out := "Dracula\nNord\n\n  Solarized Dark  \n"
	got := parseList(out)
	want := []string{"Dracula", "Nord", "Solarized Dark"}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("idx %d: got %q want %q", i, got[i], want[i])
		}
	}
}

func TestParseListEmpty(t *testing.T) {
	got := parseList("")
	if len(got) != 0 {
		t.Errorf("empty input: got %v, want empty", got)
	}
}

func TestParseListOnlyWhitespace(t *testing.T) {
	got := parseList("  \n  \n  ")
	if len(got) != 0 {
		t.Errorf("whitespace-only: got %v, want empty", got)
	}
}

func TestParseListSingleLine(t *testing.T) {
	got := parseList("Dracula")
	if len(got) != 1 || got[0] != "Dracula" {
		t.Errorf("single line: got %v, want [Dracula]", got)
	}
}

func TestParseListPreservesSpacesInside(t *testing.T) {
	got := parseList("  Dracula Theme  \n")
	if len(got) != 1 || got[0] != "Dracula Theme" {
		t.Errorf("got %v, want [Dracula Theme]", got)
	}
}

func TestErrUnavailableMessage(t *testing.T) {
	err := errUnavailable
	if err.Error() != "ghostty binary not available" {
		t.Errorf("error message = %q", err.Error())
	}
}
