# Set-My-Ghostty — Build & Install
#
# Build:
#   make          → build smg binary
#   make install  → install binary + desktop file + icon (DESTDIR support)
#   make clean    → remove build artifacts
#   make test     → run tests

BINARY   := smg
PREFIX   ?= /usr/local
BINDIR   ?= $(PREFIX)/bin
DATADIR  ?= $(PREFIX)/share
DESTDIR  ?=

GO       ?= go
GOFLAGS  ?= -ldflags "-X main.version=$(shell git describe --tags --always --dirty 2>/dev/null || echo dev)"
VETARGS  ?= ./...
TESTARGS ?= -v -count=1 ./...

.PHONY: all build install uninstall clean test vet

all: build

# --- Build ----------------------------------------------------------------

build: $(BINARY)

$(BINARY): cmd/smg/main.go
	$(GO) build $(GOFLAGS) -o $@ ./cmd/smg

# --- Install / Uninstall ---------------------------------------------------

install: build
	install -d $(DESTDIR)$(BINDIR)
	install -m 755 $(BINARY) $(DESTDIR)$(BINDIR)/$(BINARY)
	install -d $(DESTDIR)$(DATADIR)/applications
	install -m 644 smg.desktop $(DESTDIR)$(DATADIR)/applications/smg.desktop
	install -d $(DESTDIR)$(DATADIR)/icons/hicolor/scalable/apps
	install -m 644 internal/ui/smg.png $(DESTDIR)$(DATADIR)/icons/hicolor/256x256/apps/smg.png
	install -m 644 smg.svg $(DESTDIR)$(DATADIR)/icons/hicolor/scalable/apps/smg.svg
	install -d $(DESTDIR)$(DATADIR)/licenses/smg
	install -m 644 LICENSE $(DESTDIR)$(DATADIR)/licenses/smg/LICENSE
	@echo "Installed smg to $(DESTDIR)$(BINDIR)/$(BINARY)"
	@echo "Icon cache update (may need root):"
	@echo "  gtk-update-icon-cache -f -t $(DESTDIR)$(DATADIR)/icons/hicolor 2>/dev/null || true"

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/$(BINARY)
	rm -f $(DESTDIR)$(DATADIR)/applications/smg.desktop
	rm -f $(DESTDIR)$(DATADIR)/icons/hicolor/scalable/apps/smg.svg
	rm -rf $(DESTDIR)$(DATADIR)/licenses/smg
	@echo "Uninstalled smg"

# --- Test / Vet -----------------------------------------------------------

test:
	$(GO) test $(TESTARGS)

vet:
	$(GO) vet $(VETARGS)

# --- Clean ----------------------------------------------------------------

clean:
	rm -f $(BINARY) $(BINARY).test *.out *.test
	rm -rf dist/
