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
