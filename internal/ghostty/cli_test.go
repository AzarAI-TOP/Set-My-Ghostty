package ghostty

import "testing"

func TestUnavailableWhenNoPath(t *testing.T) {
	c := &CLI{Path: ""}
	if c.Available() {
		t.Error("CLI with empty path should be unavailable")
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
