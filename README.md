# Set-My-Ghostty

A GUI to manage your [ghostty](https://ghostty.org) configuration. The binary is `smg`.

## Status
Work in progress. See `docs/superpowers/specs/` for the design.

## Build
```
go build -o smg ./cmd/smg
```
On Linux you need Fyne's build deps, e.g. on Fedora:
`sudo dnf install mesa-libGL-devel libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel libXxf86vm-devel`

## Usage
```
smg                      # auto-detect config
smg --config /path/file  # explicit file
```
Tabs: Appearance, Font, Keymaps, Window & Behavior, and a Raw editor. Saving
writes a `.bak` backup first and validates with `ghostty` when available.

> Tabs do not live-sync; the Raw tab is authoritative when you click "Apply raw text".

## License
MIT
