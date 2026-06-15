package ui

import (
	"fyne.io/fyne/v2/widget"
)

func boolToStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// entryItem builds a labeled text entry bound to a scalar key.
func (a *App) entryItem(label, key string) *widget.FormItem {
	e := widget.NewEntry()
	e.SetText(a.model.Value(key))
	e.OnChanged = func(s string) { a.model.SetValue(key, s) }
	return widget.NewFormItem(label, e)
}

// checkItem builds a labeled checkbox bound to a bool key ("true"/"false").
func (a *App) checkItem(label, key string) *widget.FormItem {
	c := widget.NewCheck("", func(b bool) { a.model.SetValue(key, boolToStr(b)) })
	c.SetChecked(a.model.Value(key) == "true")
	return widget.NewFormItem(label, c)
}

// selectItem builds a labeled dropdown bound to a scalar key.
func (a *App) selectItem(label, key string, options []string) *widget.FormItem {
	sel := widget.NewSelect(options, func(s string) { a.model.SetValue(key, s) })
	if v := a.model.Value(key); v != "" {
		sel.SetSelected(v)
	}
	return widget.NewFormItem(label, sel)
}
