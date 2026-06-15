# Set-My-Ghostty (`smg`) — Design

**Date:** 2026-06-15
**Status:** Approved (pending implementation plan)

## Summary

A cross-platform (Linux + macOS) desktop GUI, written in Go with the Fyne
toolkit, for editing a [ghostty](https://ghostty.org) terminal configuration
file. The binary is named `smg`. It edits the config **in place while
preserving comments and ordering**, backs up before saving, and validates saves
against the installed `ghostty` binary when available.

## Goals

- Friendly tabbed GUI for the most-used ghostty settings.
- Never destroy a hand-written config: preserve comments, blank lines, and key
  ordering; write a `.bak` backup before every save.
- Work with or without the `ghostty` binary on PATH (graceful degradation).
- Ship as a single binary per OS; reproducible build + CI from day one.

## Non-Goals (v1)

- Live-applying config to a running ghostty instance (ghostty reloads config
  itself; we only write the file).
- Exhaustive coverage of every ghostty option in dedicated widgets — anything
  not covered by a tab is still editable via the Raw tab.
- Windows support (ghostty's primary targets are Linux/macOS).

## Architecture

Pure-Go Fyne app with a strict dependency direction: the config/schema/ghostty
packages know **nothing** about Fyne. The UI depends on them, never the reverse.
This keeps the hard logic (parsing, round-tripping) unit-testable without a GUI.

```
cmd/smg/main.go            entry point, flag parsing, path resolution
internal/config/           comment-preserving parser + serializer (no Fyne)
internal/schema/           valid keys/enums: ghostty CLI discovery + static fallback
internal/ghostty/          thin wrapper over the ghostty binary (validate/show-config/list-*)
internal/ui/               Fyne app + one file per tab
  app.go  appearance.go  font.go  keymaps.go  window.go  raw.go
```

Module path: `github.com/AzarAI-TOP/Set-My-Ghostty`. Binary: `smg`
(`go build -o smg ./cmd/smg`).

## The config model (core)

Ghostty config is line-oriented: `key = value`, `#` comments, and some
**repeatable** keys (notably `keybind` and `palette`). To preserve comments and
ordering, the file is parsed into an **ordered document of lines**:

- `CommentLine{Text}`
- `BlankLine`
- `KeyValueLine{Key, Value, TrailingComment, Raw}`

Repeatable keys are kept as ordered multiples (the document is a slice, not a
map). A lookup index maps key → positions for editing.

Editing semantics:

- Changing a single-valued key updates its existing `KeyValueLine` in place, or
  appends a new one at the end of the file if absent.
- Repeatable keys (keybind, palette) are managed as an ordered list: add, edit,
  remove individual entries.
- Everything not touched is serialized back byte-for-byte.

**Round-trip invariant (core test):** `serialize(parse(bytes)) == bytes` for any
valid input with no edits.

Backup: before writing, copy the current file to `<path>.bak`. If the backup
write fails, **abort the save**.

## Schema discovery

Valid keys, value types, and enum choices come from the installed ghostty
binary when present, with a bundled static fallback:

- `ghostty +show-config --default --docs` → parse keys + doc comments + default
  values.
- `ghostty +list-themes`, `+list-fonts` → populate pickers.
- Fallback: a hand-maintained `schema.go` with the keys the tabs need, so the
  app is usable when ghostty isn't on PATH (validation/theme-list disabled, with
  a non-blocking notice).

## Data flow

1. **Startup** → resolve config path → parse → discover schema → populate
   widgets.
   - Path resolution order: `--config` flag → `$XDG_CONFIG_HOME/ghostty/config`
     → `~/.config/ghostty/config` → `~/.config/ghostty/config.ghostty` → file
     picker in the UI. The chosen path is remembered (app preferences).
   - On macOS, `$XDG_CONFIG_HOME` defaults to `~/.config` as well (ghostty uses
     XDG paths on both platforms).
2. **Edit** → each widget writes into an in-memory pending-change set keyed by
   config key. No file change until save.
3. **Save** → apply pending changes into the document → write `.bak` → write
   file → run `ghostty +validate-config` (if available) → show OK or list errors
   in a status bar.

## Tabs

- **Appearance** — theme picker (from `+list-themes`), foreground/background
  color pickers, `palette`, `background-opacity` slider, `window-padding-*`.
- **Font** — `font-family` (from `+list-fonts`), `font-size`, `font-feature`,
  bold/italic style keys, `adjust-cell-width`/`adjust-cell-height`.
- **Keymaps** — table of `keybind = trigger=action` rows; add/edit/remove with
  an action dropdown driven by the schema.
- **Window & Behavior** — window decorations, tab/split behavior, cursor
  style/blink, mouse settings, `shell-integration`, `confirm-close-surface`.
- **Raw** — plain-text editor of the actual file; always present as the escape
  hatch. Syncs with the parsed model (editing here re-parses on apply; a parse
  error keeps you in Raw with the error shown).

## Error handling

| Situation | Behavior |
|-----------|----------|
| Config file missing | Offer to create it at the resolved path |
| Parse error on load | Open Raw tab with the error shown; never silently drop data |
| `ghostty +validate-config` fails | Warn in status bar, do **not** block save (ghostty is lenient) |
| Backup write fails | **Abort** the save, surface the error |
| ghostty binary absent | Disable validate/theme/font discovery, use static schema, non-blocking notice |

## Testing & build (TDD)

- **config**: golden-file round-trip tests + targeted edit tests (update key,
  add key, edit/remove repeatable entry, comment preservation).
- **schema**: parse fixture output of `+show-config --docs` into the schema
  model.
- **ui**: a form-model layer (schema + document → widget state and back) is
  separated from rendering so it can be unit-tested without spinning up Fyne.
- **Repo**: public `AzarAI-TOP/Set-My-Ghostty`, MIT license, conventional Go
  layout, README.
- **CI**: GitHub Actions matrix (ubuntu-latest + macos-latest) running
  `go vet ./...`, `go test ./...`, `go build ./...`. Linux job installs Fyne
  build deps (`libgl1-mesa-dev xorg-dev`). A `.goreleaser.yaml` is stubbed for
  later releases.

## Open questions / deferred

- Release automation (goreleaser, signed macOS builds) — deferred past v1.
- Per-key inline documentation tooltips from `+show-config --docs` — nice to
  have, can land incrementally.
