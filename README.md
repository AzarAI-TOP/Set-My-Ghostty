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

## Install
Prebuilt binaries for Linux and macOS (amd64/arm64) are attached to each
[GitHub release](https://github.com/AzarAI-TOP/Set-My-Ghostty/releases). The latest
release is **[v1.0.1](https://github.com/AzarAI-TOP/Set-My-Ghostty/releases/tag/v1.0.1)**.
Download the archive for your platform, extract, and put `smg` on your `PATH`:
```
tar -xzf smg_v1.0.1_linux_amd64.tar.gz
sudo install smg /usr/local/bin/
```

## Usage
```
smg                      # auto-detect config
smg --config /path/file  # explicit file
smg --version            # print version
```
Tabs: Appearance, Font, Keymaps, Window & Behavior, and a Raw editor. Saving
writes a `.bak` backup first and validates with `ghostty` when available.

> Tabs do not live-sync; the Raw tab is authoritative when you click "Apply raw text".

## License
MIT
