package config

import "testing"

func TestParseBytesRoundTrip(t *testing.T) {
	inputs := []string{
		"",
		"theme = dracula\n",
		"# a comment\n\nfont-size = 13\ntheme = dracula\n",
		"font-size = 13", // no trailing newline
		"keybind = ctrl+a=copy_to_clipboard\nkeybind = ctrl+v=paste_from_clipboard\n",
		"  spaced-key   =   spaced value  \n", // odd spacing preserved on unedited lines
	}
	for _, in := range inputs {
		doc := Parse([]byte(in))
		got := string(doc.Bytes())
		if got != in {
			t.Errorf("round-trip mismatch\n in: %q\nout: %q", in, got)
		}
	}
}

func TestParseClassifiesLines(t *testing.T) {
	doc := Parse([]byte("# c\n\ntheme = dracula\n"))
	if len(doc.Lines) != 3 {
		t.Fatalf("want 3 lines, got %d", len(doc.Lines))
	}
	if doc.Lines[0].Kind != KindComment {
		t.Errorf("line 0: want comment, got %v", doc.Lines[0].Kind)
	}
	if doc.Lines[1].Kind != KindBlank {
		t.Errorf("line 1: want blank, got %v", doc.Lines[1].Kind)
	}
	if doc.Lines[2].Kind != KindKeyValue || doc.Lines[2].Key != "theme" || doc.Lines[2].Value != "dracula" {
		t.Errorf("line 2: want theme=dracula, got %+v", doc.Lines[2])
	}
}
