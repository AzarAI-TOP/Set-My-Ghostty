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
		a.model.DiscardPending()
		a.infof("Raw text applied in memory — press Save to write the file")
	})
	reload := widget.NewButton("Refresh from form edits", func() {
		a.model.Apply()
		entry.SetText(string(a.model.Doc.Bytes()))
	})

	top := container.NewHBox(applyRaw, reload)
	return container.NewTabItem("Raw", container.NewBorder(top, nil, nil, nil, entry))
}
