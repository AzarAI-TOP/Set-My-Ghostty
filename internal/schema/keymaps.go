// Package schema describes the ghostty config options and known keybinding
// actions that smg understands.
package schema

// ActionGroup is a category label for grouping keybinding actions.
type ActionGroup string

const (
	GroupClipboard  ActionGroup = "Clipboard"
	GroupSplits     ActionGroup = "Splits"
	GroupTabs       ActionGroup = "Tabs"
	GroupNavigation ActionGroup = "Navigation"
	GroupTerminal   ActionGroup = "Terminal"
	GroupView       ActionGroup = "View"
	GroupFont       ActionGroup = "Font Size"
	GroupInspector  ActionGroup = "Inspector"
	GroupOther      ActionGroup = "Other"
)

// KeymapAction describes a single ghostty keybinding action.
type KeymapAction struct {
	Action   string      // ghostty action name, e.g. "copy_to_clipboard"
	Label    string      // human-readable label, e.g. "Copy to clipboard"
	Category ActionGroup // group for UI organisation
}

// DefaultKeymapActions returns the curated list of known ghostty keybinding
// actions. This list is used to populate the keymaps tab with a Gnome-like
// interface where each action shows its current shortcut or "Not Set".
func DefaultKeymapActions() []KeymapAction {
	return []KeymapAction{
		// ── Clipboard ──────────────────────────────────────────────────
		{Action: "copy_to_clipboard", Label: "Copy to clipboard", Category: GroupClipboard},
		{Action: "paste_from_clipboard", Label: "Paste from clipboard", Category: GroupClipboard},
		{Action: "paste_from_selection", Label: "Paste from selection", Category: GroupClipboard},

		// ── Splits ─────────────────────────────────────────────────────
		{Action: "new_split:right", Label: "New split right", Category: GroupSplits},
		{Action: "new_split:down", Label: "New split down", Category: GroupSplits},
		{Action: "new_split:left", Label: "New split left", Category: GroupSplits},
		{Action: "new_split:up", Label: "New split up", Category: GroupSplits},
		{Action: "toggle_split_zoom", Label: "Toggle split zoom", Category: GroupSplits},
		{Action: "equalize_splits", Label: "Equalize splits", Category: GroupSplits},

		// ── Tabs ───────────────────────────────────────────────────────
		{Action: "new_tab", Label: "New tab", Category: GroupTabs},
		{Action: "close_tab", Label: "Close tab", Category: GroupTabs},
		{Action: "next_tab", Label: "Next tab", Category: GroupTabs},
		{Action: "previous_tab", Label: "Previous tab", Category: GroupTabs},
		{Action: "goto_tab:1", Label: "Go to tab 1", Category: GroupTabs},
		{Action: "goto_tab:2", Label: "Go to tab 2", Category: GroupTabs},
		{Action: "goto_tab:3", Label: "Go to tab 3", Category: GroupTabs},
		{Action: "goto_tab:4", Label: "Go to tab 4", Category: GroupTabs},
		{Action: "goto_tab:5", Label: "Go to tab 5", Category: GroupTabs},
		{Action: "goto_tab:6", Label: "Go to tab 6", Category: GroupTabs},
		{Action: "goto_tab:7", Label: "Go to tab 7", Category: GroupTabs},
		{Action: "goto_tab:8", Label: "Go to tab 8", Category: GroupTabs},
		{Action: "goto_tab:9", Label: "Go to tab 9", Category: GroupTabs},
		{Action: "toggle_tab_overview", Label: "Toggle tab overview", Category: GroupTabs},

		// ── Navigation ─────────────────────────────────────────────────
		{Action: "scroll_to_top", Label: "Scroll to top", Category: GroupNavigation},
		{Action: "scroll_to_bottom", Label: "Scroll to bottom", Category: GroupNavigation},
		{Action: "page_scroll_up", Label: "Page scroll up", Category: GroupNavigation},
		{Action: "page_scroll_down", Label: "Page scroll down", Category: GroupNavigation},
		{Action: "line_scroll_up", Label: "Line scroll up", Category: GroupNavigation},
		{Action: "line_scroll_down", Label: "Line scroll down", Category: GroupNavigation},

		// ── Terminal ───────────────────────────────────────────────────
		{Action: "clear_screen", Label: "Clear screen", Category: GroupTerminal},
		{Action: "clear_scrollback", Label: "Clear scrollback", Category: GroupTerminal},
		{Action: "close_surface", Label: "Close surface", Category: GroupTerminal},
		{Action: "reload_config", Label: "Reload config", Category: GroupTerminal},

		// ── View ───────────────────────────────────────────────────────
		{Action: "toggle_fullscreen", Label: "Toggle fullscreen", Category: GroupView},
		{Action: "toggle_quick_terminal", Label: "Toggle quick terminal", Category: GroupView},

		// ── Font Size ──────────────────────────────────────────────────
		{Action: "change_font_size:+1", Label: "Increase font size", Category: GroupFont},
		{Action: "change_font_size:-1", Label: "Decrease font size", Category: GroupFont},
		{Action: "change_font_size:reset", Label: "Reset font size", Category: GroupFont},

		// ── Inspector ──────────────────────────────────────────────────
		{Action: "inspector:terminal", Label: "Terminal inspector", Category: GroupInspector},
		{Action: "inspector:webview", Label: "Webview inspector", Category: GroupInspector},
	}
}

// ActionLabels returns a map of action → label for quick lookup.
func ActionLabels() map[string]string {
	m := make(map[string]string, len(DefaultKeymapActions()))
	for _, a := range DefaultKeymapActions() {
		m[a.Action] = a.Label
	}
	return m
}

// ActionCategories returns a map of action → category for quick lookup.
func ActionCategories() map[string]ActionGroup {
	m := make(map[string]ActionGroup, len(DefaultKeymapActions()))
	for _, a := range DefaultKeymapActions() {
		m[a.Action] = a.Category
	}
	return m
}
