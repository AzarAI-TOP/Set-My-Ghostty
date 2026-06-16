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

func TestParseValueContainsEquals(t *testing.T) {
	// Value with embedded '=' — only the first '=' is the separator.
	doc := Parse([]byte("keybind = ctrl+a=copy_to_clipboard\n"))
	if len(doc.Lines) != 1 {
		t.Fatalf("want 1 line, got %d", len(doc.Lines))
	}
	line := doc.Lines[0]
	if line.Key != "keybind" {
		t.Errorf("key = %q, want keybind", line.Key)
	}
	if line.Value != "ctrl+a=copy_to_clipboard" {
		t.Errorf("value = %q, want ctrl+a=copy_to_clipboard", line.Value)
	}
}

func TestParseEmptyKey(t *testing.T) {
	doc := Parse([]byte("= value\n"))
	if len(doc.Lines) != 1 {
		t.Fatalf("want 1 line, got %d", len(doc.Lines))
	}
	if doc.Lines[0].Kind != KindKeyValue {
		t.Fatalf("want KindKeyValue, got %v", doc.Lines[0].Kind)
	}
	if doc.Lines[0].Key != "" {
		t.Errorf("key = %q, want empty", doc.Lines[0].Key)
	}
	if doc.Lines[0].Value != "value" {
		t.Errorf("value = %q, want value", doc.Lines[0].Value)
	}
}

func TestParseEmptyValue(t *testing.T) {
	doc := Parse([]byte("key = \n"))
	if len(doc.Lines) != 1 {
		t.Fatalf("want 1 line, got %d", len(doc.Lines))
	}
	if doc.Lines[0].Kind != KindKeyValue {
		t.Fatalf("want KindKeyValue, got %v", doc.Lines[0].Kind)
	}
	if doc.Lines[0].Key != "key" {
		t.Errorf("key = %q, want key", doc.Lines[0].Key)
	}
	if doc.Lines[0].Value != "" {
		t.Errorf("value = %q, want empty", doc.Lines[0].Value)
	}
}

func TestParseBlankLinesVariants(t *testing.T) {
	inputs := []string{
		"  \n",   // spaces only
		"\t\n",   // tab only
		"  \t  \n", // mixed whitespace
	}
	for _, in := range inputs {
		doc := Parse([]byte(in))
		if len(doc.Lines) != 1 {
			t.Errorf("input %q: want 1 line, got %d", in, len(doc.Lines))
			continue
		}
		if doc.Lines[0].Kind != KindBlank {
			t.Errorf("input %q: want KindBlank, got %v", in, doc.Lines[0].Kind)
		}
	}
}

func TestParseCommentVariants(t *testing.T) {
	inputs := []string{
		"# simple\n",
		"## double hash\n",
		"#  \n", // comment with trailing spaces
		"#key=value\n", // no space after # — still a comment
	}
	for _, in := range inputs {
		doc := Parse([]byte(in))
		if len(doc.Lines) != 1 {
			t.Errorf("input %q: want 1 line, got %d", in, len(doc.Lines))
			continue
		}
		if doc.Lines[0].Kind != KindComment {
			t.Errorf("input %q: want KindComment, got %v", in, doc.Lines[0].Kind)
		}
	}
}

func TestParseLineWithoutEquals(t *testing.T) {
	// Lines without '=' that aren't comments or blank — treated as comment-like.
	doc := Parse([]byte("some bare text\n"))
	if len(doc.Lines) != 1 {
		t.Fatalf("want 1 line, got %d", len(doc.Lines))
	}
	if doc.Lines[0].Kind != KindComment {
		t.Errorf("bare text line: want KindComment, got %v", doc.Lines[0].Kind)
	}
}

func TestParseUnicode(t *testing.T) {
	input := "# 颜色主题\ntheme = 暗色\n"
	doc := Parse([]byte(input))
	got := string(doc.Bytes())
	if got != input {
		t.Errorf("Unicode round-trip\n in: %q\nout: %q", input, got)
	}
}

func TestParseTrailingNewlinePreservation(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"has trailing newline", "theme = dracula\n"},
		{"no trailing newline", "theme = dracula"},
		{"empty with newline", "\n"},
		{"empty no newline", ""},
		{"multiple lines with newline", "a = 1\nb = 2\n"},
		{"multiple lines no newline", "a = 1\nb = 2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := Parse([]byte(tt.input))
			got := string(doc.Bytes())
			if got != tt.input {
				t.Errorf("round-trip\n in: %q\nout: %q", tt.input, got)
			}
		})
	}
}

func TestParseLargeDocument(t *testing.T) {
	var input string
	for i := 0; i < 1000; i++ {
		input += "key" + itoa(i) + " = value" + itoa(i) + "\n"
	}
	doc := Parse([]byte(input))
	if len(doc.Lines) != 1000 {
		t.Fatalf("want 1000 lines, got %d", len(doc.Lines))
	}
	got := string(doc.Bytes())
	if got != input {
		t.Error("large document round-trip mismatch")
	}
}

// itoa is a small helper to avoid importing strconv in tests.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	n := len(buf)
	neg := i < 0
	if neg {
		i = -i
	}
	for i > 0 {
		n--
		buf[n] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		n--
		buf[n] = '-'
	}
	return string(buf[n:])
}

func TestDocumentBytesEmpty(t *testing.T) {
	doc := Parse(nil)
	if b := doc.Bytes(); b != nil {
		t.Errorf("Bytes() = %q, want nil", b)
	}

	doc = Parse([]byte(""))
	if b := doc.Bytes(); b != nil {
		t.Errorf("Bytes() for empty string = %q, want nil", b)
	}
}

func TestLineSerializeDirty(t *testing.T) {
	line := Line{Kind: KindKeyValue, Key: "theme", Value: "nord", dirty: true, Raw: "theme = old\n"}
	got := line.serialize()
	want := "theme = nord"
	if got != want {
		t.Errorf("dirty serialize: got %q, want %q", got, want)
	}
}

func TestLineSerializeClean(t *testing.T) {
	line := Line{Kind: KindKeyValue, Key: "theme", Value: "nord", Raw: "theme = nord"}
	got := line.serialize()
	want := "theme = nord"
	if got != want {
		t.Errorf("clean serialize: got %q, want %q", got, want)
	}

	line2 := Line{Kind: KindComment, Raw: "# comment"}
	if s := line2.serialize(); s != "# comment" {
		t.Errorf("comment serialize: got %q", s)
	}

	line3 := Line{Kind: KindBlank, Raw: ""}
	if s := line3.serialize(); s != "" {
		t.Errorf("blank serialize: got %q", s)
	}
}
