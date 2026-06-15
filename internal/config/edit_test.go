package config

import "testing"

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
