package ui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// findLabel recursively searches a rendered object tree for a Label with the
// given text and returns it (nil if not found).
func findLabel(o fyne.CanvasObject, text string) *widget.Label {
	switch v := o.(type) {
	case *widget.Label:
		if v.Text == text {
			return v
		}
	case *fyne.Container:
		for _, c := range v.Objects {
			if l := findLabel(c, text); l != nil {
				return l
			}
		}
	}
	return nil
}

// TestTappableRowNameIsVisible guards against the regression where the action
// name label collapsed to ~0 width (showing only the trigger) because a
// truncating label was placed in a Border left slot.
func TestTappableRowNameIsVisible(t *testing.T) {
	test.NewApp()
	const name = "Copy to clipboard"
	row := newTappableRow(name, "", nil)

	w := test.NewWindow(row)
	defer w.Close()
	w.Resize(fyne.NewSize(600, 48))

	var nameLabel *widget.Label
	for _, o := range test.WidgetRenderer(row).Objects() {
		if l := findLabel(o, name); l != nil {
			nameLabel = l
			break
		}
	}
	if nameLabel == nil {
		t.Fatal("name label not present in rendered row")
	}
	if got := nameLabel.Size().Width; got < 200 {
		t.Errorf("name label width = %v in a 600px row; it collapsed and the action name is hidden", got)
	}
}
