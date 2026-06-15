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
