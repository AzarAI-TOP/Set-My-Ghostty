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
