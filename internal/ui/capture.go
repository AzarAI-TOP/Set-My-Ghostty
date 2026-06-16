package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// showKeyCaptureDialog opens a modal popup that listens for a key combination
// and calls onSet with the captured trigger string (or "" if cleared).
func showKeyCaptureDialog(actionLabel, action, currentTrigger string, onSet func(string), parent fyne.Window) {
	captureLabel := widget.NewLabelWithStyle("Press a key combination...", fyne.TextAlignCenter, fyne.TextStyle{Monospace: true})
	captureLabel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}

	instruction := widget.NewLabel("Press the desired key combination,\nthen click Set.")
	instruction.Alignment = fyne.TextAlignCenter
	instruction.Wrapping = fyne.TextWrapWord

	var trigger string
	var modCtrl, modAlt, modShift, modSuper bool
	var mainKey string

	// Hidden entry to hold canvas focus so key events reach us.
	focusEntry := widget.NewEntry()
	focusEntry.Hidden = true
	focusEntry.Disable()

	updateDisplay := func() {
		if mainKey == "" && !modCtrl && !modAlt && !modShift && !modSuper {
			captureLabel.SetText("Press a key combination...")
			return
		}
		var parts []string
		if modCtrl {
			parts = append(parts, "ctrl")
		}
		if modAlt {
			parts = append(parts, "alt")
		}
		if modShift {
			parts = append(parts, "shift")
		}
		if modSuper {
			parts = append(parts, "super")
		}
		if mainKey != "" {
			parts = append(parts, mainKey)
		}
		captureLabel.SetText(strings.Join(parts, "+"))
		trigger = strings.Join(parts, "+")
	}

	resetModifiers := func() {
		modCtrl = false
		modAlt = false
		modShift = false
		modSuper = false
		mainKey = ""
		trigger = ""
	}

	// ── Install desktop key handlers ────────────────────────────────────
	deskCanvas, canDesk := parent.Canvas().(desktop.Canvas)
	if !canDesk {
		// Non-desktop environment — just show the dialog without capture.
		dialog.NewInformation("Not supported", "Key capture requires a desktop environment.\nPlease use the Raw tab instead.", parent)
		return
	}

	oldDown := deskCanvas.OnKeyDown()
	oldUp := deskCanvas.OnKeyUp()

	keyDown := func(ev *fyne.KeyEvent) {
		name := string(ev.Name)
		switch {
		case hasPrefix(name, "Control"):
			modCtrl = true
		case hasPrefix(name, "Alt"):
			modAlt = true
		case hasPrefix(name, "Shift"):
			modShift = true
		case hasPrefix(name, "Super"), hasPrefix(name, "Meta"):
			modSuper = true
		case name == "Escape":
			resetModifiers()
		default:
			// Regular key pressed — record it once.
			if mainKey == "" {
				mainKey = cleanKeyName(name)
			}
		}
		updateDisplay()
	}

	keyUp := func(ev *fyne.KeyEvent) {
		name := string(ev.Name)
		switch {
		case hasPrefix(name, "Control"):
			modCtrl = false
		case hasPrefix(name, "Alt"):
			modAlt = false
		case hasPrefix(name, "Shift"):
			modShift = false
		case hasPrefix(name, "Super"), hasPrefix(name, "Meta"):
			modSuper = false
		default:
			// Don't clear mainKey here — we want it to stay visible.
		}
		updateDisplay()
	}

	deskCanvas.SetOnKeyDown(keyDown)
	deskCanvas.SetOnKeyUp(keyUp)

	restoreHandlers := func() {
		deskCanvas.SetOnKeyDown(oldDown)
		deskCanvas.SetOnKeyUp(oldUp)
	}

	// ── Build dialog content ────────────────────────────────────────────
	actionNameLabel := widget.NewLabelWithStyle(actionLabel, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	actionNameLabel.Alignment = fyne.TextAlignCenter

	content := container.NewVBox(
		actionNameLabel,
		container.NewPadded(captureLabel),
		instruction,
		focusEntry,
	)

	// d is referenced by the button handlers below, so declare it first and
	// assign once the content (which embeds the buttons) is built.
	var d *dialog.CustomDialog

	setBtn := widget.NewButtonWithIcon("Set", theme.ConfirmIcon(), func() {
		restoreHandlers()
		// Require a real key, not a modifier-only combo (e.g. just "ctrl"),
		// which ghostty would reject as an invalid trigger.
		if mainKey != "" {
			onSet(trigger)
		}
		d.Hide()
	})
	clearBtn := widget.NewButton("Clear", func() {
		restoreHandlers()
		onSet("")
		d.Hide()
	})
	cancelBtn := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		restoreHandlers()
		d.Hide()
	})

	buttons := container.NewHBox(
		clearBtn,
		container.NewPadded(),
		setBtn,
		cancelBtn,
	)

	d = dialog.NewCustomWithoutButtons("Set shortcut", container.NewBorder(
		content, buttons, nil, nil, nil,
	), parent)
	d.Resize(fyne.NewSize(380, 280))
	d.Show()
}

// showAddCustomActionDialog opens a dialog to create a new custom keybinding.
func showAddCustomActionDialog(knownActions []string, onAdd func(action, trigger string), parent fyne.Window) {
	actionEntry := widget.NewEntry()
	actionEntry.SetPlaceHolder("action_name (e.g. toggle_mouse)")
	actionEntry.Validator = validation.NewRegexp(`^[a-z_][a-z0-9_:+-]*$`, "Invalid action name")

	triggerLabel := widget.NewLabelWithStyle("Not set — click Capture", fyne.TextAlignCenter, fyne.TextStyle{Monospace: true})
	var capturedTrigger string

	captureBtn := widget.NewButton("Capture key", func() {
		showKeyCaptureDialog("New action", "", "", func(trigger string) {
			capturedTrigger = trigger
			if trigger == "" {
				triggerLabel.SetText("Not set")
			} else {
				triggerLabel.SetText(trigger)
			}
		}, parent)
	})

	form := container.NewVBox(
		widget.NewLabel("Action name:"),
		actionEntry,
		widget.NewLabel("Shortcut:"),
		container.NewBorder(nil, nil, nil, captureBtn, triggerLabel),
	)

	d := dialog.NewCustomConfirm("Add custom shortcut", "Add", "Cancel", form,
		func(ok bool) {
			if ok && actionEntry.Text != "" && capturedTrigger != "" {
				onAdd(actionEntry.Text, capturedTrigger)
			}
		}, parent)
	d.Resize(fyne.NewSize(360, 220))
	d.Show()
}

// ── Helpers ─────────────────────────────────────────────────────────────

func hasPrefix(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return strings.EqualFold(s[:len(prefix)], prefix)
}

// cleanKeyName converts a Fyne KeyName string to a ghostty-compatible
// lowercase key name.
func cleanKeyName(name string) string {
	// Handle special key names.
	m := map[string]string{
		"Escape":     "escape",
		"Tab":        "tab",
		"BackSpace":  "backspace",
		"Return":     "enter",
		"KP_Enter":   "enter",
		"Space":      "space",
		"Up":         "up",
		"Down":       "down",
		"Left":       "left",
		"Right":      "right",
		"Home":       "home",
		"End":        "end",
		"Prior":      "page_up",
		"Next":       "page_down",
		"Delete":     "delete",
		"Insert":     "insert",
	}
	if v, ok := m[name]; ok {
		return v
	}
	return strings.ToLower(name)
}

