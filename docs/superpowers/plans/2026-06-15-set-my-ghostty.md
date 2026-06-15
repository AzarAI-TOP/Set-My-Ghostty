# Set-My-Ghostty Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build `smg`, a Go + Fyne desktop GUI that edits a ghostty config file in place while preserving comments, with backup, validation, and four feature tabs plus a raw editor.

**Architecture:** Core packages (`config`, `schema`, `ghostty`) are pure Go with no Fyne dependency and are TDD-tested. A `ui` package depends on them: a testable form-model bridges schema+document to widget state, and Fyne tabs render it. Config is parsed into an ordered line document so unedited lines round-trip byte-for-byte.

**Tech Stack:** Go 1.26, Fyne v2, the `ghostty` CLI (optional, for validation/discovery), GitHub Actions CI.

**Spec:** `docs/superpowers/specs/2026-06-15-set-my-ghostty-design.md`

---

## File Structure

```
go.mod                                  module github.com/AzarAI-TOP/Set-My-Ghostty
cmd/smg/main.go                         entry point: flags, path resolution, launch UI
internal/config/document.go             Line/Document model + Parse + Bytes (serialize)
internal/config/document_test.go        round-trip + edit tests
internal/config/edit.go                 Get/Set/AddRepeatable/RemoveAll/SetRepeatable
internal/config/edit_test.go            edit-operation tests
internal/config/save.go                 Save (backup .bak then write) + ResolvePath
internal/config/save_test.go            backup + path-resolution tests
internal/schema/schema.go               Option/Schema types + Static() fallback
internal/schema/schema_test.go          static schema sanity tests
internal/schema/discover.go             ParseShowConfigDocs
internal/schema/discover_test.go        parse fixture of +show-config --docs
internal/ghostty/cli.go                 Detect + ValidateConfig/ListThemes/ListFonts/ShowConfigDocs
internal/ghostty/cli_test.go            Detect + parsing tests (no real binary needed)
internal/ui/model.go                    Model: schema+document -> values, pending edits, Apply
internal/ui/model_test.go               model logic tests (no Fyne)
internal/ui/app.go                      Fyne app, window, tab container, load/save, status bar
internal/ui/appearance.go               Appearance tab
internal/ui/font.go                     Font tab
internal/ui/keymaps.go                  Keymaps tab
internal/ui/window.go                   Window & Behavior tab
internal/ui/raw.go                      Raw text-editor tab
README.md  LICENSE  .gitignore
.github/workflows/ci.yml                vet/test/build matrix (ubuntu+macos)
.goreleaser.yaml                        stub for later releases
```

Dependency direction: `config`, `schema`, `ghostty` import nothing from `ui`. `ui` imports all three. `cmd/smg` imports `config` + `ui`.

---

## Task 1: Project scaffolding

**Files:**
- Create: `go.mod`, `cmd/smg/main.go`, `README.md`, `LICENSE`

- [ ] **Step 1: Initialize the module**

Run:
```bash
cd ~/Workspace/Set-My-Ghostty
go mod init github.com/AzarAI-TOP/Set-My-Ghostty
```
Expected: creates `go.mod` with `go 1.26`.

- [ ] **Step 2: Minimal compilable entry point**

Create `cmd/smg/main.go`:
```go
// Command smg is a GUI editor for ghostty configuration files.
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	configPath := flag.String("config", "", "path to ghostty config file (default: auto-detect)")
	flag.Parse()

	// UI wiring is added in a later task; for now just prove the binary builds.
	if *configPath != "" {
		fmt.Fprintf(os.Stdout, "smg: using config %s\n", *configPath)
	}
}
```

- [ ] **Step 3: Verify it builds**

Run: `go build -o smg ./cmd/smg && ./smg --help`
Expected: builds; `--help` prints the `-config` flag usage.

- [ ] **Step 4: Add LICENSE (MIT) and README**

Create `LICENSE` with the standard MIT text, copyright `2026 AzarAI-TOP`.
Create `README.md`:
```markdown
# Set-My-Ghostty

A GUI to manage your [ghostty](https://ghostty.org) configuration. The binary is `smg`.

## Status
Work in progress. See `docs/superpowers/specs/` for the design.

## Build
```
go build -o smg ./cmd/smg
```

## License
MIT
```

- [ ] **Step 5: Commit**

```bash
git add go.mod cmd/smg/main.go README.md LICENSE
git commit -m "feat: scaffold smg module and entry point"
```

---

## Task 2: Config document model + parser (round-trip)

**Files:**
- Create: `internal/config/document.go`
- Test: `internal/config/document_test.go`

- [ ] **Step 1: Write the failing round-trip test**

Create `internal/config/document_test.go`:
```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/config/ -run TestParse -v`
Expected: FAIL — `Parse`, `Document`, `KindComment` undefined.

- [ ] **Step 3: Implement the model and parser**

Create `internal/config/document.go`:
```go
// Package config parses, edits, and serializes ghostty config files while
// preserving comments, blank lines, and the ordering of unedited lines.
package config

import (
	"bytes"
	"strings"
)

// LineKind classifies a line in a ghostty config file.
type LineKind int

const (
	KindBlank LineKind = iota
	KindComment
	KindKeyValue
)

// Line is one physical line of the config file.
type Line struct {
	Kind  LineKind
	Key   string // KindKeyValue only
	Value string // KindKeyValue only
	Raw   string // original text; used verbatim unless dirty is set
	dirty bool   // when true, Bytes() regenerates from Key/Value
}

// Document is an ordered list of lines plus trailing-newline state.
type Document struct {
	Lines           []Line
	trailingNewline bool
}

// Parse reads config bytes into a Document.
func Parse(b []byte) *Document {
	d := &Document{}
	s := string(b)
	if s == "" {
		return d
	}
	d.trailingNewline = strings.HasSuffix(s, "\n")
	body := s
	if d.trailingNewline {
		body = strings.TrimSuffix(s, "\n")
	}
	for _, raw := range strings.Split(body, "\n") {
		d.Lines = append(d.Lines, classify(raw))
	}
	return d
}

func classify(raw string) Line {
	trimmed := strings.TrimSpace(raw)
	switch {
	case trimmed == "":
		return Line{Kind: KindBlank, Raw: raw}
	case strings.HasPrefix(trimmed, "#"):
		return Line{Kind: KindComment, Raw: raw}
	case strings.Contains(raw, "="):
		parts := strings.SplitN(raw, "=", 2)
		return Line{
			Kind:  KindKeyValue,
			Key:   strings.TrimSpace(parts[0]),
			Value: strings.TrimSpace(parts[1]),
			Raw:   raw,
		}
	default:
		// A line with no '=' and not a comment: keep verbatim, treat as comment-like.
		return Line{Kind: KindComment, Raw: raw}
	}
}

func (l Line) serialize() string {
	if l.Kind == KindKeyValue && l.dirty {
		return l.Key + " = " + l.Value
	}
	return l.Raw
}

// Bytes serializes the document back to config bytes.
func (d *Document) Bytes() []byte {
	if len(d.Lines) == 0 {
		return nil
	}
	var buf bytes.Buffer
	for i, l := range d.Lines {
		if i > 0 {
			buf.WriteByte('\n')
		}
		buf.WriteString(l.serialize())
	}
	if d.trailingNewline {
		buf.WriteByte('\n')
	}
	return buf.Bytes()
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/config/ -run TestParse -v`
Expected: PASS (both tests).

- [ ] **Step 5: Commit**

```bash
git add internal/config/document.go internal/config/document_test.go
git commit -m "feat(config): comment-preserving parser with round-trip serialization"
```

---

## Task 3: Config edit operations

**Files:**
- Create: `internal/config/edit.go`
- Test: `internal/config/edit_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/config/edit_test.go`:
```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/config/ -run "TestGet|TestSet|TestRepeat" -v`
Expected: FAIL — `Get`, `Set`, `GetAll`, `SetRepeatable` undefined.

- [ ] **Step 3: Implement edit operations**

Create `internal/config/edit.go`:
```go
package config

// Get returns the value of the first KeyValue line with the given key.
func (d *Document) Get(key string) (string, bool) {
	for _, l := range d.Lines {
		if l.Kind == KindKeyValue && l.Key == key {
			return l.Value, true
		}
	}
	return "", false
}

// GetAll returns the values of every KeyValue line with the given key, in order.
func (d *Document) GetAll(key string) []string {
	var out []string
	for _, l := range d.Lines {
		if l.Kind == KindKeyValue && l.Key == key {
			out = append(out, l.Value)
		}
	}
	return out
}

// Set updates the first existing line with key, or appends a new line.
func (d *Document) Set(key, value string) {
	for i := range d.Lines {
		if d.Lines[i].Kind == KindKeyValue && d.Lines[i].Key == key {
			d.Lines[i].Value = value
			d.Lines[i].dirty = true
			return
		}
	}
	d.append(key, value)
}

// RemoveAll deletes every KeyValue line with the given key.
func (d *Document) RemoveAll(key string) {
	kept := d.Lines[:0]
	for _, l := range d.Lines {
		if l.Kind == KindKeyValue && l.Key == key {
			continue
		}
		kept = append(kept, l)
	}
	d.Lines = kept
}

// SetRepeatable replaces all lines for key with one line per value, appended at
// the position of the former first occurrence (or end of file if none existed).
func (d *Document) SetRepeatable(key string, values []string) {
	insertAt := -1
	kept := make([]Line, 0, len(d.Lines))
	for _, l := range d.Lines {
		if l.Kind == KindKeyValue && l.Key == key {
			if insertAt == -1 {
				insertAt = len(kept)
			}
			continue
		}
		kept = append(kept, l)
	}
	newLines := make([]Line, 0, len(values))
	for _, v := range values {
		newLines = append(newLines, Line{Kind: KindKeyValue, Key: key, Value: v, dirty: true})
	}
	if insertAt == -1 {
		insertAt = len(kept)
	}
	if d.trailingNewline || len(d.Lines) == 0 {
		d.trailingNewline = true
	}
	d.Lines = append(kept[:insertAt:insertAt], append(newLines, kept[insertAt:]...)...)
}

func (d *Document) append(key, value string) {
	d.Lines = append(d.Lines, Line{Kind: KindKeyValue, Key: key, Value: value, dirty: true})
	d.trailingNewline = true
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/config/ -v`
Expected: PASS (all config tests, including Task 2's).

- [ ] **Step 5: Commit**

```bash
git add internal/config/edit.go internal/config/edit_test.go
git commit -m "feat(config): get/set/remove and repeatable-key edit operations"
```

---

## Task 4: Save with backup + config path resolution

**Files:**
- Create: `internal/config/save.go`
- Test: `internal/config/save_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/config/save_test.go`:
```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/config/ -run "TestSave|TestResolve" -v`
Expected: FAIL — `Save`, `ResolvePath` undefined.

- [ ] **Step 3: Implement save + resolve**

Create `internal/config/save.go`:
```go
package config

import (
	"os"
	"path/filepath"
)

// Save writes a .bak backup of the current file (if it exists) and then writes
// the document. If the backup cannot be written, the save is aborted and the
// original file is left untouched.
func Save(path string, d *Document) error {
	if existing, err := os.ReadFile(path); err == nil {
		if err := os.WriteFile(path+".bak", existing, 0o644); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	return os.WriteFile(path, d.Bytes(), 0o644)
}

// ResolvePath picks the config file to edit. Order:
//  1. flagPath, if non-empty.
//  2. $XDG_CONFIG_HOME/ghostty/config (XDG_CONFIG_HOME defaults to ~/.config).
//  3. ~/.config/ghostty/config.ghostty (this user's existing file).
//  4. The canonical default path even if it does not yet exist.
func ResolvePath(flagPath string) (string, error) {
	if flagPath != "" {
		return flagPath, nil
	}
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	canonical := filepath.Join(base, "ghostty", "config")
	candidates := []string{
		canonical,
		filepath.Join(base, "ghostty", "config.ghostty"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	return canonical, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/config/ -v`
Expected: PASS (all config tests).

- [ ] **Step 5: Commit**

```bash
git add internal/config/save.go internal/config/save_test.go
git commit -m "feat(config): save with .bak backup and config path resolution"
```

---

## Task 5: Schema types + static fallback

**Files:**
- Create: `internal/schema/schema.go`
- Test: `internal/schema/schema_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/schema/schema_test.go`:
```go
package schema

import "testing"

func TestStaticHasCoreKeys(t *testing.T) {
	s := Static()
	for _, key := range []string{"theme", "font-family", "font-size", "background-opacity", "keybind"} {
		opt, ok := s.Options[key]
		if !ok {
			t.Errorf("static schema missing %q", key)
			continue
		}
		if opt.Key != key {
			t.Errorf("option key mismatch: %q != %q", opt.Key, key)
		}
	}
}

func TestStaticTypesAndRepeatable(t *testing.T) {
	s := Static()
	if s.Options["background-opacity"].Type != TypeFloat {
		t.Errorf("background-opacity should be float")
	}
	if !s.Options["keybind"].Repeatable {
		t.Errorf("keybind should be repeatable")
	}
	if s.Options["font-size"].Type != TypeInt {
		t.Errorf("font-size should be int")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/schema/ -run TestStatic -v`
Expected: FAIL — `Static`, `TypeFloat` undefined.

- [ ] **Step 3: Implement schema types + static set**

Create `internal/schema/schema.go`:
```go
// Package schema describes the ghostty config options that smg understands:
// their value types, enum choices, and whether they are repeatable.
package schema

// Type is the value type of a config option.
type Type int

const (
	TypeString Type = iota
	TypeBool
	TypeInt
	TypeFloat
	TypeEnum
	TypeColor
)

// Option describes a single ghostty config key.
type Option struct {
	Key        string
	Type       Type
	Enum       []string // populated for TypeEnum
	Default    string
	Doc        string
	Repeatable bool // true for keys like keybind/palette that may appear many times
}

// Schema is a set of options keyed by config key.
type Schema struct {
	Options map[string]Option
}

// Static returns the bundled fallback schema used when the ghostty binary is
// unavailable. It covers the keys the GUI tabs need.
func Static() *Schema {
	opts := []Option{
		{Key: "theme", Type: TypeString, Doc: "Color theme name."},
		{Key: "background", Type: TypeColor, Doc: "Background color."},
		{Key: "foreground", Type: TypeColor, Doc: "Foreground (text) color."},
		{Key: "background-opacity", Type: TypeFloat, Default: "1", Doc: "Window background opacity (0-1)."},
		{Key: "background-blur", Type: TypeBool, Doc: "Blur content behind a translucent window."},
		{Key: "palette", Type: TypeString, Repeatable: true, Doc: "Palette color override, e.g. 0=#1d1f21."},
		{Key: "window-padding-x", Type: TypeInt, Doc: "Horizontal window padding."},
		{Key: "window-padding-y", Type: TypeInt, Doc: "Vertical window padding."},

		{Key: "font-family", Type: TypeString, Doc: "Font family name."},
		{Key: "font-size", Type: TypeInt, Default: "13", Doc: "Font size in points."},
		{Key: "font-feature", Type: TypeString, Repeatable: true, Doc: "OpenType font feature, e.g. -liga."},
		{Key: "font-style", Type: TypeString, Doc: "Preferred font style name."},
		{Key: "adjust-cell-width", Type: TypeString, Doc: "Adjust cell width, e.g. 5% or -2px."},
		{Key: "adjust-cell-height", Type: TypeString, Doc: "Adjust cell height, e.g. 5% or -2px."},

		{Key: "keybind", Type: TypeString, Repeatable: true, Doc: "Key binding: trigger=action."},

		{Key: "window-decoration", Type: TypeEnum, Enum: []string{"auto", "none", "client", "server"}, Doc: "Window decorations."},
		{Key: "confirm-close-surface", Type: TypeBool, Default: "true", Doc: "Confirm before closing a surface."},
		{Key: "cursor-style", Type: TypeEnum, Enum: []string{"block", "bar", "underline", "block_hollow"}, Doc: "Cursor style."},
		{Key: "cursor-style-blink", Type: TypeBool, Doc: "Blink the cursor."},
		{Key: "mouse-hide-while-typing", Type: TypeBool, Doc: "Hide the mouse while typing."},
		{Key: "shell-integration", Type: TypeEnum, Enum: []string{"none", "detect", "bash", "elvish", "fish", "zsh"}, Doc: "Shell integration mode."},
	}
	s := &Schema{Options: make(map[string]Option, len(opts))}
	for _, o := range opts {
		s.Options[o.Key] = o
	}
	return s
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/schema/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/schema/schema.go internal/schema/schema_test.go
git commit -m "feat(schema): option types and bundled static fallback schema"
```

---

## Task 6: Schema discovery from `+show-config --docs`

`ghostty +show-config --default --docs` prints each option as doc-comment lines
(starting with `# `) followed by a `key = value` line. This task parses that
output and merges it over the static schema so discovered defaults/docs win but
the curated types/enums/`Repeatable` flags are preserved.

**Files:**
- Create: `internal/schema/discover.go`
- Test: `internal/schema/discover_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/schema/discover_test.go`:
```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/schema/ -run "TestParseShow|TestMerge" -v`
Expected: FAIL — `ParseShowConfigDocs`, `MergeFrom` undefined.

- [ ] **Step 3: Implement discovery + merge**

Create `internal/schema/discover.go`:
```go
package schema

import (
	"bufio"
	"io"
	"strings"
)

// ParseShowConfigDocs parses the output of `ghostty +show-config --default
// --docs` into a Schema. Doc-comment lines (starting with '#') accumulate until
// the next `key = value` line, which they describe.
func ParseShowConfigDocs(r io.Reader) (*Schema, error) {
	s := &Schema{Options: map[string]Option{}}
	var docLines []string
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "":
			docLines = nil
		case strings.HasPrefix(trimmed, "#"):
			docLines = append(docLines, strings.TrimSpace(strings.TrimPrefix(trimmed, "#")))
		case strings.Contains(trimmed, "="):
			parts := strings.SplitN(trimmed, "=", 2)
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			s.Options[key] = Option{
				Key:     key,
				Type:    TypeString,
				Default: val,
				Doc:     strings.Join(docLines, " "),
			}
			docLines = nil
		}
	}
	return s, sc.Err()
}

// MergeFrom returns a copy of s with docs and defaults taken from other when
// present, while preserving s's curated Type, Enum, and Repeatable fields. Keys
// only in other are added with their discovered (string) type.
func (s *Schema) MergeFrom(other *Schema) *Schema {
	out := &Schema{Options: make(map[string]Option, len(s.Options))}
	for k, o := range s.Options {
		out.Options[k] = o
	}
	for k, disc := range other.Options {
		if cur, ok := out.Options[k]; ok {
			if disc.Doc != "" {
				cur.Doc = disc.Doc
			}
			if disc.Default != "" {
				cur.Default = disc.Default
			}
			out.Options[k] = cur
		} else {
			out.Options[k] = disc
		}
	}
	return out
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/schema/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/schema/discover.go internal/schema/discover_test.go
git commit -m "feat(schema): discover options from ghostty +show-config --docs"
```

---

## Task 7: ghostty CLI wrapper

Wraps the optional `ghostty` binary. Detection and output parsing are tested
without needing a real ghostty install; the parsing helpers are exported and
unit-tested against fixture strings.

**Files:**
- Create: `internal/ghostty/cli.go`
- Test: `internal/ghostty/cli_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/ghostty/cli_test.go`:
```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/ghostty/ -v`
Expected: FAIL — `CLI`, `parseList` undefined.

- [ ] **Step 3: Implement the wrapper**

Create `internal/ghostty/cli.go`:
```go
// Package ghostty is a thin wrapper over the optional `ghostty` binary, used to
// validate configs and discover themes/fonts/options. All methods degrade
// gracefully when the binary is not available.
package ghostty

import (
	"bytes"
	"os/exec"
	"strings"
)

// CLI represents a located ghostty binary. Path is empty if none was found.
type CLI struct {
	Path string
}

// Detect locates the ghostty binary on PATH.
func Detect() *CLI {
	path, err := exec.LookPath("ghostty")
	if err != nil {
		return &CLI{}
	}
	return &CLI{Path: path}
}

// Available reports whether a ghostty binary was found.
func (c *CLI) Available() bool { return c.Path != "" }

// ValidateConfig runs `ghostty +validate-config --config-file=<path>` and
// reports whether it succeeded along with combined output.
func (c *CLI) ValidateConfig(path string) (ok bool, output string, err error) {
	if !c.Available() {
		return false, "", errUnavailable
	}
	out, runErr := c.run("+validate-config", "--config-file="+path)
	return runErr == nil, out, nil
}

// ListThemes returns available theme names.
func (c *CLI) ListThemes() ([]string, error) {
	out, err := c.run("+list-themes", "--plain")
	if err != nil {
		return nil, err
	}
	return parseList(out), nil
}

// ListFonts returns available font family names.
func (c *CLI) ListFonts() ([]string, error) {
	out, err := c.run("+list-fonts")
	if err != nil {
		return nil, err
	}
	return parseList(out), nil
}

// ShowConfigDocs returns the output of `+show-config --default --docs`.
func (c *CLI) ShowConfigDocs() (string, error) {
	return c.run("+show-config", "--default", "--docs")
}

func (c *CLI) run(args ...string) (string, error) {
	if !c.Available() {
		return "", errUnavailable
	}
	var buf bytes.Buffer
	cmd := exec.Command(c.Path, args...)
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.String(), err
}

// parseList splits command output into trimmed, non-empty lines.
func parseList(out string) []string {
	var res []string
	for _, line := range strings.Split(out, "\n") {
		if t := strings.TrimSpace(line); t != "" {
			res = append(res, t)
		}
	}
	return res
}

type cliError string

func (e cliError) Error() string { return string(e) }

const errUnavailable = cliError("ghostty binary not available")
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/ghostty/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/ghostty/cli.go internal/ghostty/cli_test.go
git commit -m "feat(ghostty): CLI wrapper for validate/list/show-config"
```

---

## Task 8: UI form-model (testable bridge, no Fyne)

The Model holds the parsed document + schema and tracks pending edits so tabs
stay dumb. This is where the testable UI logic lives — no Fyne import.

**Files:**
- Create: `internal/ui/model.go`
- Test: `internal/ui/model_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/ui/model_test.go`:
```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/ui/ -run "TestValue|TestApply|TestSet" -v`
Expected: FAIL — `Model`, `NewModel` undefined.

- [ ] **Step 3: Implement the model**

Create `internal/ui/model.go`:
```go
package ui

import (
	"github.com/AzarAI-TOP/Set-My-Ghostty/internal/config"
	"github.com/AzarAI-TOP/Set-My-Ghostty/internal/schema"
)

// Model bridges a parsed config Document and a Schema to the widgets. Edits are
// staged in pending maps and flushed into the Document by Apply.
type Model struct {
	Doc    *config.Document
	Schema *schema.Schema

	pendingScalar map[string]string
	pendingList   map[string][]string
}

// NewModel builds a model over a document and schema.
func NewModel(doc *config.Document, s *schema.Schema) *Model {
	return &Model{
		Doc:           doc,
		Schema:        s,
		pendingScalar: map[string]string{},
		pendingList:   map[string][]string{},
	}
}

// Value returns the current value for a scalar key: pending edit, else the
// document value, else the schema default, else "".
func (m *Model) Value(key string) string {
	if v, ok := m.pendingScalar[key]; ok {
		return v
	}
	if v, ok := m.Doc.Get(key); ok {
		return v
	}
	if opt, ok := m.Schema.Options[key]; ok {
		return opt.Default
	}
	return ""
}

// SetValue stages a scalar edit. If the new value equals the document value the
// pending edit is cleared (no-op edits do not make the model dirty).
func (m *Model) SetValue(key, value string) {
	if cur, ok := m.Doc.Get(key); ok && cur == value {
		delete(m.pendingScalar, key)
		return
	}
	m.pendingScalar[key] = value
}

// List returns the current values for a repeatable key (pending else document).
func (m *Model) List(key string) []string {
	if v, ok := m.pendingList[key]; ok {
		return v
	}
	return m.Doc.GetAll(key)
}

// SetList stages an edit to a repeatable key.
func (m *Model) SetList(key string, values []string) {
	m.pendingList[key] = values
}

// Dirty reports whether there are unsaved edits.
func (m *Model) Dirty() bool {
	return len(m.pendingScalar) > 0 || len(m.pendingList) > 0
}

// Apply flushes all pending edits into the document and clears them.
func (m *Model) Apply() {
	for k, v := range m.pendingScalar {
		m.Doc.Set(k, v)
	}
	for k, vs := range m.pendingList {
		m.Doc.SetRepeatable(k, vs)
	}
	m.pendingScalar = map[string]string{}
	m.pendingList = map[string][]string{}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/ui/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/ui/model.go internal/ui/model_test.go
git commit -m "feat(ui): testable form-model bridging schema and document"
```

---

## Task 9: Fyne app shell (window, tabs, load/save, status bar)

This task introduces the Fyne dependency and the `App` struct that owns the
window and shared widget helpers. Tab files (Tasks 10–14) define methods on
`*App`. UI rendering is verified manually (Step 5), not by unit test.

**Files:**
- Create: `internal/ui/app.go`
- Modify: `cmd/smg/main.go`

- [ ] **Step 1: Add the Fyne dependency**

Run:
```bash
go get fyne.io/fyne/v2@latest
```
Expected: `go.mod`/`go.sum` updated with fyne v2. (On Linux this needs the
system packages `libgl1-mesa-dev xorg-dev`; install via `sudo dnf install
mesa-libGL-devel libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel`
on Fedora if the build complains.)

- [ ] **Step 2: Implement the app shell**

Create `internal/ui/app.go`:
```go
package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/AzarAI-TOP/Set-My-Ghostty/internal/config"
	"github.com/AzarAI-TOP/Set-My-Ghostty/internal/ghostty"
	"github.com/AzarAI-TOP/Set-My-Ghostty/internal/schema"
)

// App owns the window, the form-model, and discovered data shared across tabs.
type App struct {
	model  *Model
	path   string
	cli    *ghostty.CLI
	themes []string
	fonts  []string

	win    fyne.Window
	status *widget.Label
}

// Run loads the config at path, builds the schema and window, and blocks until
// the window is closed.
func Run(path string) error {
	doc, err := load(path)
	if err != nil {
		return err
	}

	cli := ghostty.Detect()
	sch := schema.Static()
	var themes, fonts []string
	if cli.Available() {
		if out, err := cli.ShowConfigDocs(); err == nil {
			if disc, derr := schema.ParseShowConfigDocs(stringsReader(out)); derr == nil {
				sch = sch.MergeFrom(disc)
			}
		}
		themes, _ = cli.ListThemes()
		fonts, _ = cli.ListFonts()
	}

	a := &App{
		model:  NewModel(doc, sch),
		path:   path,
		cli:    cli,
		themes: themes,
		fonts:  fonts,
		status: widget.NewLabel(""),
	}

	fa := fyneapp.NewWithID("top.azarai.smg")
	a.win = fa.NewWindow("Set-My-Ghostty")
	a.status.SetText("Editing " + path)
	if !cli.Available() {
		a.status.SetText("ghostty not found — validation and theme/font lists disabled")
	}

	tabs := container.NewAppTabs(
		a.appearanceTab(),
		a.fontTab(),
		a.keymapsTab(),
		a.windowTab(),
		a.rawTab(),
	)

	saveBtn := widget.NewButton("Save", a.save)
	bottom := container.NewBorder(nil, nil, nil, saveBtn, a.status)
	a.win.SetContent(container.NewBorder(nil, bottom, nil, nil, tabs))
	a.win.Resize(fyne.NewSize(720, 560))
	a.win.ShowAndRun()
	return nil
}

// save flushes edits, writes the file (with backup), and validates.
func (a *App) save() {
	a.model.Apply()
	if err := config.Save(a.path, a.model.Doc); err != nil {
		a.status.SetText("Save failed: " + err.Error())
		return
	}
	if a.cli.Available() {
		ok, out, _ := a.cli.ValidateConfig(a.path)
		if !ok {
			a.status.SetText("Saved with warnings: " + firstLine(out))
			return
		}
	}
	a.status.SetText("Saved to " + a.path)
}

func load(path string) (*config.Document, error) {
	b, err := readFileAllowMissing(path)
	if err != nil {
		return nil, err
	}
	return config.Parse(b), nil
}

func firstLine(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			return s[:i]
		}
	}
	if s == "" {
		return "(no detail)"
	}
	return s
}

func (a *App) infof(format string, args ...any) {
	a.status.SetText(fmt.Sprintf(format, args...))
}
```

Also create `internal/ui/io.go` for the small filesystem/string helpers (kept out
of app.go to keep it focused):
```go
package ui

import (
	"io"
	"os"
	"strings"
)

func readFileAllowMissing(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	return b, err
}

func stringsReader(s string) io.Reader { return strings.NewReader(s) }
```

- [ ] **Step 3: Wire main.go to the UI**

Replace `cmd/smg/main.go` with:
```go
// Command smg is a GUI editor for ghostty configuration files.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/AzarAI-TOP/Set-My-Ghostty/internal/config"
	"github.com/AzarAI-TOP/Set-My-Ghostty/internal/ui"
)

func main() {
	configPath := flag.String("config", "", "path to ghostty config file (default: auto-detect)")
	flag.Parse()

	path, err := config.ResolvePath(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "smg:", err)
		os.Exit(1)
	}
	if err := ui.Run(path); err != nil {
		fmt.Fprintln(os.Stderr, "smg:", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 4: Verify it builds**

Run: `go build -o smg ./cmd/smg`
Expected: builds. (It won't fully run until the tab methods exist in Tasks 10–14; this step only confirms the shell compiles once those stubs are added. If building now, temporarily add stub tab methods returning `container.NewTabItem("", widget.NewLabel(""))` — but prefer to do Step 4's full verify after Task 14.)

- [ ] **Step 5: Commit**

```bash
git add internal/ui/app.go internal/ui/io.go cmd/smg/main.go go.mod go.sum
git commit -m "feat(ui): Fyne app shell with load/save and status bar"
```

> Note: build/run verification of the window happens at the end of Task 14 once
> all five tab methods exist.

---

## Task 10: Shared widget helpers + Appearance tab

Adds `internal/ui/widgets.go` (shared row helpers used by every tab) and the
Appearance tab. These are pure wiring on top of the tested Model, so they are
verified by building/running, not unit tests.

**Files:**
- Create: `internal/ui/widgets.go`, `internal/ui/appearance.go`

- [ ] **Step 1: Implement shared helpers**

Create `internal/ui/widgets.go`:
```go
package ui

import (
	"fyne.io/fyne/v2/widget"
)

func boolToStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// entryItem builds a labeled text entry bound to a scalar key.
func (a *App) entryItem(label, key string) *widget.FormItem {
	e := widget.NewEntry()
	e.SetText(a.model.Value(key))
	e.OnChanged = func(s string) { a.model.SetValue(key, s) }
	return widget.NewFormItem(label, e)
}

// checkItem builds a labeled checkbox bound to a bool key ("true"/"false").
func (a *App) checkItem(label, key string) *widget.FormItem {
	c := widget.NewCheck("", func(b bool) { a.model.SetValue(key, boolToStr(b)) })
	c.SetChecked(a.model.Value(key) == "true")
	return widget.NewFormItem(label, c)
}

// selectItem builds a labeled dropdown bound to a scalar key.
func (a *App) selectItem(label, key string, options []string) *widget.FormItem {
	sel := widget.NewSelect(options, func(s string) { a.model.SetValue(key, s) })
	if v := a.model.Value(key); v != "" {
		sel.SetSelected(v)
	}
	return widget.NewFormItem(label, sel)
}
```

- [ ] **Step 2: Implement the Appearance tab**

Create `internal/ui/appearance.go`:
```go
package ui

import (
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (a *App) appearanceTab() *container.TabItem {
	form := widget.NewForm()
	if len(a.themes) > 0 {
		form.AppendItem(a.selectItem("Theme", "theme", a.themes))
	} else {
		form.AppendItem(a.entryItem("Theme", "theme"))
	}
	form.AppendItem(a.entryItem("Background color", "background"))
	form.AppendItem(a.entryItem("Foreground color", "foreground"))
	form.AppendItem(a.entryItem("Background opacity (0-1)", "background-opacity"))
	form.AppendItem(a.checkItem("Background blur", "background-blur"))
	form.AppendItem(a.entryItem("Window padding X", "window-padding-x"))
	form.AppendItem(a.entryItem("Window padding Y", "window-padding-y"))
	return container.NewTabItem("Appearance", container.NewVScroll(form))
}
```

- [ ] **Step 3: Verify it compiles**

Run: `go build ./internal/ui/`
Expected: fails only on the not-yet-defined `fontTab`, `keymapsTab`, `windowTab`,
`rawTab` (defined in Tasks 11–14). The Appearance + helpers code itself should
have no errors. (Optional: `go vet ./internal/ui/ 2>&1 | grep appearance` should
be empty.)

- [ ] **Step 4: Commit**

```bash
git add internal/ui/widgets.go internal/ui/appearance.go
git commit -m "feat(ui): shared widget helpers and Appearance tab"
```

---

## Task 11: Font tab

**Files:**
- Create: `internal/ui/font.go`

- [ ] **Step 1: Implement the Font tab**

Create `internal/ui/font.go`:
```go
package ui

import (
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (a *App) fontTab() *container.TabItem {
	form := widget.NewForm()
	if len(a.fonts) > 0 {
		form.AppendItem(a.selectItem("Font family", "font-family", a.fonts))
	} else {
		form.AppendItem(a.entryItem("Font family", "font-family"))
	}
	form.AppendItem(a.entryItem("Font size", "font-size"))
	form.AppendItem(a.entryItem("Font style", "font-style"))
	form.AppendItem(a.entryItem("Adjust cell width", "adjust-cell-width"))
	form.AppendItem(a.entryItem("Adjust cell height", "adjust-cell-height"))
	return container.NewTabItem("Font", container.NewVScroll(form))
}
```

> Note: `font-feature` is repeatable; it is editable via the Raw tab in v1. A
> dedicated list editor can be added later, reusing the keymaps list pattern.

- [ ] **Step 2: Commit**

```bash
git add internal/ui/font.go
git commit -m "feat(ui): Font tab"
```

---

## Task 12: Keymaps tab (repeatable list editor)

**Files:**
- Create: `internal/ui/keymaps.go`

- [ ] **Step 1: Implement the Keymaps tab**

Create `internal/ui/keymaps.go`:
```go
package ui

import (
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (a *App) keymapsTab() *container.TabItem {
	binds := append([]string{}, a.model.List("keybind")...)
	list := container.NewVBox()

	sync := func() { a.model.SetList("keybind", binds) }

	var rebuild func()
	rebuild = func() {
		list.RemoveAll()
		for i := range binds {
			idx := i
			entry := widget.NewEntry()
			entry.SetText(binds[idx])
			entry.OnChanged = func(s string) {
				binds[idx] = s
				sync()
			}
			del := widget.NewButton("Remove", func() {
				binds = append(binds[:idx], binds[idx+1:]...)
				sync()
				rebuild()
			})
			list.Add(container.NewBorder(nil, nil, nil, del, entry))
		}
		list.Refresh()
	}
	rebuild()

	add := widget.NewButton("Add keybind", func() {
		binds = append(binds, "")
		sync()
		rebuild()
	})
	help := widget.NewLabel("Format: trigger=action  (e.g. ctrl+c=copy_to_clipboard)")

	content := container.NewBorder(help, add, nil, nil, container.NewVScroll(list))
	return container.NewTabItem("Keymaps", content)
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/ui/keymaps.go
git commit -m "feat(ui): Keymaps tab with add/edit/remove keybind rows"
```

---

## Task 13: Window & Behavior tab

**Files:**
- Create: `internal/ui/window.go`

- [ ] **Step 1: Implement the tab + enum helper**

Create `internal/ui/window.go`:
```go
package ui

import (
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// enum returns the schema enum choices for a key (nil if not an enum).
func (a *App) enum(key string) []string {
	if opt, ok := a.model.Schema.Options[key]; ok {
		return opt.Enum
	}
	return nil
}

func (a *App) windowTab() *container.TabItem {
	form := widget.NewForm()
	form.AppendItem(a.selectItem("Window decoration", "window-decoration", a.enum("window-decoration")))
	form.AppendItem(a.checkItem("Confirm close surface", "confirm-close-surface"))
	form.AppendItem(a.selectItem("Cursor style", "cursor-style", a.enum("cursor-style")))
	form.AppendItem(a.checkItem("Cursor blink", "cursor-style-blink"))
	form.AppendItem(a.checkItem("Hide mouse while typing", "mouse-hide-while-typing"))
	form.AppendItem(a.selectItem("Shell integration", "shell-integration", a.enum("shell-integration")))
	return container.NewTabItem("Window & Behavior", container.NewVScroll(form))
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/ui/window.go
git commit -m "feat(ui): Window & Behavior tab"
```

---

## Task 14: Raw editor tab

**Files:**
- Create: `internal/ui/raw.go`

- [ ] **Step 1: Implement the Raw tab**

Create `internal/ui/raw.go`:
```go
package ui

import (
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/AzarAI-TOP/Set-My-Ghostty/internal/config"
)

func (a *App) rawTab() *container.TabItem {
	entry := widget.NewMultiLineEntry()
	entry.SetText(string(a.model.Doc.Bytes()))

	applyRaw := widget.NewButton("Apply raw text", func() {
		a.model.Doc = config.Parse([]byte(entry.Text))
		a.model.pendingScalar = map[string]string{}
		a.model.pendingList = map[string][]string{}
		a.infof("Raw text applied in memory — press Save to write the file")
	})
	reload := widget.NewButton("Refresh from form edits", func() {
		a.model.Apply()
		entry.SetText(string(a.model.Doc.Bytes()))
	})

	top := container.NewHBox(applyRaw, reload)
	return container.NewTabItem("Raw", container.NewBorder(top, nil, nil, nil, entry))
}
```

> v1 limitation (documented in README): tabs do not live-sync with each other.
> The Raw tab is the source of truth when you press "Apply raw text"; "Refresh
> from form edits" pulls current form state back into the text box.

- [ ] **Step 2: Build the whole app**

Run: `go build -o smg ./cmd/smg`
Expected: builds cleanly — all five tab methods now exist.

- [ ] **Step 3: Manual run verification**

Run:
```bash
cp ~/.config/ghostty/config.ghostty /tmp/smg-test-config 2>/dev/null || printf 'theme = dracula\nfont-size = 13\n' > /tmp/smg-test-config
./smg --config /tmp/smg-test-config
```
Expected: a window opens with five tabs (Appearance, Font, Keymaps, Window &
Behavior, Raw). Change font size, add a keybind, click Save. Confirm:
```bash
cat /tmp/smg-test-config       # shows your edits
cat /tmp/smg-test-config.bak   # shows the pre-save content
```

- [ ] **Step 4: Commit**

```bash
git add internal/ui/raw.go
git commit -m "feat(ui): Raw editor tab; complete five-tab window"
```

---

## Task 15: GitHub repo, CI, and release stub

**Files:**
- Create: `.github/workflows/ci.yml`, `.goreleaser.yaml`
- Modify: `README.md`

- [ ] **Step 1: CI workflow**

Create `.github/workflows/ci.yml`:
```yaml
name: CI
on:
  push:
    branches: [ main ]
  pull_request:
jobs:
  test:
    strategy:
      fail-fast: false
      matrix:
        os: [ ubuntu-latest, macos-latest ]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.26'
      - name: Install Fyne build deps (Linux)
        if: runner.os == 'Linux'
        run: sudo apt-get update && sudo apt-get install -y libgl1-mesa-dev xorg-dev
      - name: Vet
        run: go vet ./...
      - name: Test
        run: go test ./...
      - name: Build
        run: go build -o smg ./cmd/smg
```

- [ ] **Step 2: goreleaser stub (for later)**

Create `.goreleaser.yaml`:
```yaml
# Stub config for future releases. Run `goreleaser release` once tags are ready.
version: 2
project_name: Set-My-Ghostty
builds:
  - id: smg
    main: ./cmd/smg
    binary: smg
    env: [ CGO_ENABLED=1 ]
    goos: [ linux, darwin ]
    goarch: [ amd64, arm64 ]
archives:
  - formats: [ tar.gz ]
    name_template: "smg_{{ .Os }}_{{ .Arch }}"
release:
  draft: true
```

- [ ] **Step 3: Update README with usage**

Append to `README.md`:
```markdown
## Usage
```
smg                      # auto-detect config
smg --config /path/file  # explicit file
```
Tabs: Appearance, Font, Keymaps, Window & Behavior, and a Raw editor. Saving
writes a `.bak` backup first and validates with `ghostty` when available.

> Tabs do not live-sync; the Raw tab is authoritative when you click "Apply raw text".
```

- [ ] **Step 4: Verify the full suite locally**

Run: `go vet ./... && go test ./... && go build -o smg ./cmd/smg`
Expected: all pass, binary builds.

- [ ] **Step 5: Commit CI/release/docs**

```bash
git add .github/workflows/ci.yml .goreleaser.yaml README.md
git commit -m "ci: add GitHub Actions matrix and goreleaser stub"
```

- [ ] **Step 6: Create the GitHub repo and push**

Run:
```bash
git branch -M main
gh repo create AzarAI-TOP/Set-My-Ghostty --public --source=. --remote=origin \
  --description "GUI to manage your ghostty configuration (smg)" --push
```
Expected: repo created under AzarAI-TOP, `main` pushed, CI starts.

- [ ] **Step 7: Confirm CI is green**

Run: `gh run watch` (or `gh run list`)
Expected: the CI workflow completes successfully on both ubuntu and macos.

---

## Self-Review

**Spec coverage check:**

| Spec requirement | Task |
|---|---|
| Comment/order-preserving model | Task 2 |
| Edit ops incl. repeatable keys | Task 3 |
| Save with `.bak` + abort-on-backup-failure | Task 4 |
| Config path resolution order | Task 4 |
| Static fallback schema | Task 5 |
| Schema discovery from `+show-config --docs` | Task 6 |
| ghostty CLI wrapper (validate/list/show) | Task 7 |
| Testable form-model | Task 8 |
| App shell, load/save, status bar, validation | Task 9 |
| Appearance tab | Task 10 |
| Font tab | Task 11 |
| Keymaps tab | Task 12 |
| Window & Behavior tab | Task 13 |
| Raw editor tab | Task 14 |
| Public MIT repo + CI matrix + goreleaser stub | Tasks 1 & 15 |

**Error handling coverage:** file missing → `readFileAllowMissing` returns nil → empty doc (Task 9); parse never errors (lenient classifier, Task 2); validate failure → status warning, non-blocking (Task 9 `save`); backup failure → `Save` aborts before writing (Task 4); ghostty absent → static schema + disabled lists + notice (Task 9 `Run`).

**Type consistency:** `Model` methods (`Value/SetValue/List/SetList/Apply/Dirty`) used identically in Tasks 8–14. `App` tab methods (`appearanceTab/fontTab/keymapsTab/windowTab/rawTab`) match the `container.NewAppTabs` call in Task 9. Helper names (`entryItem/checkItem/selectItem/enum`) consistent across Tasks 10–13. `config` API (`Parse/Bytes/Get/GetAll/Set/SetRepeatable/Save/ResolvePath`) consistent across Tasks 2–9.

**Placeholder scan:** none — every code step contains complete code.
