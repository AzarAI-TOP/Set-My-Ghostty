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
