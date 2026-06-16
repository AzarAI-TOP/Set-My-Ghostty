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

func TestSaveAbortsOnBackupFailure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte("original\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Make the directory read-only to cause backup write failure.
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(dir, 0o755)

	doc := Parse([]byte("replacement\n"))
	if err := Save(path, doc); err == nil {
		t.Error("Save should fail when backup cannot be written")
	}

	// Verify original file is untouched.
	got, _ := os.ReadFile(path)
	if string(got) != "original\n" {
		t.Errorf("original file modified: %q", got)
	}
}

func TestSaveToExistingDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	doc := Parse([]byte("theme = nord\n"))
	if err := Save(path, doc); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "theme = nord\n" {
		t.Errorf("file = %q", got)
	}
}

func TestSaveLargeDocument(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")

	var original string
	for i := 0; i < 500; i++ {
		original += "key" + itoa(i) + " = value" + itoa(i) + "\n"
	}
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	doc := Parse([]byte(original))
	doc.Set("key0", "changed")

	if err := Save(path, doc); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Verify backup.
	bak, err := os.ReadFile(path + ".bak")
	if err != nil {
		t.Fatalf("backup missing: %v", err)
	}
	if string(bak) != original {
		t.Error("backup content mismatch")
	}

	// Verify new file differs.
	got, _ := os.ReadFile(path)
	if string(got) == original {
		t.Error("file should differ from original after edit")
	}
}

func TestSaveEmptyDocument(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte("original\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	doc := Parse(nil)
	if err := Save(path, doc); err != nil {
		t.Fatalf("Save empty doc: %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "" {
		t.Errorf("file should be empty, got %q", got)
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

func TestResolvePathFindsConfigGhostty(t *testing.T) {
	dir := t.TempDir()
	gh := filepath.Join(dir, "ghostty")
	os.MkdirAll(gh, 0o755)
	// Only create config.ghostty — ResolvePath should prefer it over canonical.
	want := filepath.Join(gh, "config.ghostty")
	os.WriteFile(want, nil, 0o644)
	t.Setenv("XDG_CONFIG_HOME", dir)
	got, err := ResolvePath("")
	if err != nil || got != want {
		t.Fatalf("ResolvePath = %q,%v want %q", got, err, want)
	}
}

func TestResolvePathFallsBackToCanonical(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	got, err := ResolvePath("")
	if err != nil {
		t.Fatalf("ResolvePath: %v", err)
	}
	want := filepath.Join(dir, "ghostty", "config")
	if got != want {
		t.Errorf("ResolvePath = %q, want %q", got, want)
	}
}

func TestResolvePathDefaultsToDotConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // Windows fallback

	got, err := ResolvePath("")
	if err != nil {
		t.Fatalf("ResolvePath: %v", err)
	}
	want := filepath.Join(home, ".config", "ghostty", "config")
	if got != want {
		t.Errorf("ResolvePath = %q, want %q", got, want)
	}
}
