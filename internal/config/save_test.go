package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveWritesBackupThenFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte("theme = dracula\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	doc := Parse([]byte("theme = nord\n"))
	if err := Save(path, doc); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "theme = nord\n" {
		t.Errorf("file = %q", got)
	}
	bak, err := os.ReadFile(path + ".bak")
	if err != nil {
		t.Fatalf("backup missing: %v", err)
	}
	if string(bak) != "theme = dracula\n" {
		t.Errorf("backup = %q (should hold the pre-save content)", bak)
	}
}

func TestSaveNoBackupWhenFileAbsent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	doc := Parse([]byte("theme = nord\n"))
	if err := Save(path, doc); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(path + ".bak"); !os.IsNotExist(err) {
		t.Errorf("expected no .bak for a brand-new file")
	}
}

func TestResolvePathPrefersFlag(t *testing.T) {
	got, err := ResolvePath("/explicit/path")
	if err != nil || got != "/explicit/path" {
		t.Fatalf("ResolvePath(flag) = %q,%v", got, err)
	}
}

func TestResolvePathFindsExisting(t *testing.T) {
	dir := t.TempDir()
	gh := filepath.Join(dir, "ghostty")
	os.MkdirAll(gh, 0o755)
	want := filepath.Join(gh, "config")
	os.WriteFile(want, nil, 0o644)
	t.Setenv("XDG_CONFIG_HOME", dir)
	got, err := ResolvePath("")
	if err != nil || got != want {
		t.Fatalf("ResolvePath = %q,%v want %q", got, err, want)
	}
}
