// Package schema describes the ghostty config options that smg understands:
// their value types, enum choices, and whether they are repeatable.
package schema

// Type is the value type of a config option.
type Type int

const (
	TypeString Type = iota
	TypeBool
	TypeInt
	TypeFloat
	TypeEnum
	TypeColor
)

// Option describes a single ghostty config key.
type Option struct {
	Key        string
	Type       Type
	Enum       []string // populated for TypeEnum
	Default    string
	Doc        string
	Repeatable bool // true for keys like keybind/palette that may appear many times
}

// Schema is a set of options keyed by config key.
type Schema struct {
	Options map[string]Option
}

// Static returns the bundled fallback schema used when the ghostty binary is
// unavailable. It covers the keys the GUI tabs need.
func Static() *Schema {
	opts := []Option{
		{Key: "theme", Type: TypeString, Doc: "Color theme name."},
		{Key: "background", Type: TypeColor, Doc: "Background color."},
		{Key: "foreground", Type: TypeColor, Doc: "Foreground (text) color."},
		{Key: "background-opacity", Type: TypeFloat, Default: "1", Doc: "Window background opacity (0-1)."},
		{Key: "background-blur", Type: TypeBool, Doc: "Blur content behind a translucent window."},
		{Key: "palette", Type: TypeString, Repeatable: true, Doc: "Palette color override, e.g. 0=#1d1f21."},
		{Key: "window-padding-x", Type: TypeInt, Doc: "Horizontal window padding."},
		{Key: "window-padding-y", Type: TypeInt, Doc: "Vertical window padding."},

		{Key: "font-family", Type: TypeString, Doc: "Font family name."},
		{Key: "font-size", Type: TypeInt, Default: "13", Doc: "Font size in points."},
		{Key: "font-feature", Type: TypeString, Repeatable: true, Doc: "OpenType font feature, e.g. -liga."},
		{Key: "font-style", Type: TypeString, Doc: "Preferred font style name."},
		{Key: "adjust-cell-width", Type: TypeString, Doc: "Adjust cell width, e.g. 5% or -2px."},
		{Key: "adjust-cell-height", Type: TypeString, Doc: "Adjust cell height, e.g. 5% or -2px."},

		{Key: "keybind", Type: TypeString, Repeatable: true, Doc: "Key binding: trigger=action."},

		{Key: "window-decoration", Type: TypeEnum, Enum: []string{"auto", "none", "client", "server"}, Doc: "Window decorations."},
		{Key: "confirm-close-surface", Type: TypeBool, Default: "true", Doc: "Confirm before closing a surface."},
		{Key: "cursor-style", Type: TypeEnum, Enum: []string{"block", "bar", "underline", "block_hollow"}, Doc: "Cursor style."},
		{Key: "cursor-style-blink", Type: TypeBool, Doc: "Blink the cursor."},
		{Key: "mouse-hide-while-typing", Type: TypeBool, Doc: "Hide the mouse while typing."},
		{Key: "shell-integration", Type: TypeEnum, Enum: []string{"none", "detect", "bash", "elvish", "fish", "zsh"}, Doc: "Shell integration mode."},
	}
	s := &Schema{Options: make(map[string]Option, len(opts))}
	for _, o := range opts {
		s.Options[o.Key] = o
	}
	return s
}
