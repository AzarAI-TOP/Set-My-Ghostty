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
