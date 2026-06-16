package ui

import (
	"sort"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/AzarAI-TOP/Set-My-Ghostty/internal/schema"
)

// tappableRow is a clickable row that shows an action name and its shortcut.
type tappableRow struct {
	widget.BaseWidget
	name    string
	trigger string
	onTap   func()
}

func newTappableRow(name, trigger string, onTap func()) *tappableRow {
	r := &tappableRow{name: name, trigger: trigger, onTap: onTap}
	r.ExtendBaseWidget(r)
	return r
}

func (r *tappableRow) Tapped(_ *fyne.PointEvent) {
	if r.onTap != nil {
		r.onTap()
	}
}

func (r *tappableRow) CreateRenderer() fyne.WidgetRenderer {
	var triggerLabel *widget.Label
	if r.trigger != "" {
		triggerLabel = widget.NewLabelWithStyle(r.trigger, fyne.TextAlignCenter, fyne.TextStyle{Monospace: true, Bold: true})
	} else {
		triggerLabel = widget.NewLabelWithStyle("Not Set", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})
	}
	triggerLabel.Alignment = fyne.TextAlignCenter

	nameLabel := widget.NewLabel(r.name)
	nameLabel.Wrapping = fyne.TextTruncate

	content := container.NewBorder(nil, nil, nameLabel, nil, triggerLabel)
	return widget.NewSimpleRenderer(content)
}

type visibleAction struct {
	schema.KeymapAction
	trigger string
}

func (a *App) keymapsTab() *container.TabItem {
	allActions := schema.DefaultKeymapActions()
	actionLabels := schema.ActionLabels()

	var mu sync.Mutex
	binds := a.model.Doc.KeybindMap()
	if a.model.Dirty() {
		for action, trigger := range a.model.pendingKeybinds {
			binds[action] = trigger
		}
	}

	// ── Widgets ─────────────────────────────────────────────────────────
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search actions...")

	listContainer := container.NewVBox()
	scroll := container.NewVScroll(listContainer)

	// rebuild populates the list based on a search filter.
	var rebuild func(filter string)
	rebuild = func(filter string) {
		mu.Lock()
		defer mu.Unlock()

		listContainer.RemoveAll()

		// Determine which default actions match the filter.
		var visible []visibleAction
		for _, act := range allActions {
			trigger := binds[act.Action]
			if filter == "" || matchesFilter(act.Label, act.Action, trigger, filter) {
				visible = append(visible, visibleAction{KeymapAction: act, trigger: trigger})
			}
		}

		// Custom actions (in doc but not in default list).
		var custom []string
		for action := range binds {
			if _, known := actionLabels[action]; !known {
				if filter == "" || strings.Contains(strings.ToLower(action), strings.ToLower(filter)) {
					custom = append(custom, action)
				}
			}
		}
		sort.Strings(custom)

		// Render groups.
		groups := groupByAction(visible)
		for _, g := range groups {
			// Category header.
			header := widget.NewLabelWithStyle(g.Category, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			listContainer.Add(container.NewPadded(header))

			for _, va := range g.Items {
				va := va // capture loop variable
				row := newTappableRow(va.Label, va.trigger, func() {
					showKeyCaptureDialog(va.Label, va.Action, va.trigger, func(newTrigger string) {
						mu.Lock()
						if newTrigger == "" {
							delete(binds, va.Action)
						} else {
							binds[va.Action] = newTrigger
						}
						a.model.SetKeybind(va.Action, newTrigger)
						mu.Unlock()
						rebuild(searchEntry.Text)
					}, a.win)
				})
				listContainer.Add(row)
			}
		}

		// Custom actions section.
		if len(custom) > 0 {
			header := widget.NewLabelWithStyle("Custom", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			listContainer.Add(container.NewPadded(header))
			for _, action := range custom {
				trigger := binds[action]
				row := newTappableRow(action, trigger, func() {
					showKeyCaptureDialog(action, action, trigger, func(newTrigger string) {
						mu.Lock()
						if newTrigger == "" {
							delete(binds, action)
						} else {
							binds[action] = newTrigger
						}
						a.model.SetKeybind(action, newTrigger)
						mu.Unlock()
						rebuild(searchEntry.Text)
					}, a.win)
				})
				listContainer.Add(row)
			}
		}

		// Add custom shortcut button.
		addBtn := widget.NewButtonWithIcon("Add custom shortcut", theme.ContentAddIcon(), func() {
			showAddCustomActionDialog(nil, func(action, trigger string) {
				mu.Lock()
				binds[action] = trigger
				a.model.SetKeybind(action, trigger)
				mu.Unlock()
				rebuild(searchEntry.Text)
			}, a.win)
		})
		listContainer.Add(container.NewPadded(addBtn))

		scroll.Refresh()
	}

	searchEntry.OnChanged = rebuild
	rebuild("")

	content := container.NewBorder(searchEntry, nil, nil, nil, scroll)
	return container.NewTabItem("Keymaps", content)
}

func matchesFilter(label, action, trigger, filter string) bool {
	f := strings.ToLower(filter)
	return strings.Contains(strings.ToLower(label), f) ||
		strings.Contains(strings.ToLower(action), f) ||
		strings.Contains(strings.ToLower(trigger), f)
}

type groupedActions struct {
	Category string
	Items    []visibleAction
}

func groupByAction(items []visibleAction) []groupedActions {
	catOrder := []schema.ActionGroup{
		schema.GroupClipboard, schema.GroupSplits, schema.GroupTabs,
		schema.GroupNavigation, schema.GroupTerminal, schema.GroupView,
		schema.GroupFont, schema.GroupInspector,
	}
	groups := make(map[schema.ActionGroup][]visibleAction)
	for _, it := range items {
		groups[it.Category] = append(groups[it.Category], it)
	}
	var result []groupedActions
	for _, c := range catOrder {
		if acts, ok := groups[c]; ok {
			result = append(result, groupedActions{Category: string(c), Items: acts})
		}
	}
	return result
}

var _ = sortedActionKeys // keep import alive
