package watch

import (
	"github.com/charmbracelet/bubbles/key"
)

const spacebar = " "

type keyMap struct {
	Help key.Binding
	Quit key.Binding

	Esc key.Binding

	Enter key.Binding

	Up   key.Binding
	Down key.Binding

	PageDown key.Binding
	PageUp   key.Binding

	Left  key.Binding
	Right key.Binding
	Tab   key.Binding

	OpenWorkflow key.Binding
	Refresh      key.Binding
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Help, k.Quit},
		{k.Up, k.Down},
		{k.Tab},
		{k.Enter, k.OpenWorkflow},
		{k.Refresh},
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Tab, k.Quit, k.Enter, k.Esc, k.OpenWorkflow, k.Refresh}
}

var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Esc: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),

	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch focus"),
	),

	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),

	PageDown: key.NewBinding(
		key.WithKeys("pgdown", spacebar, "f"),
		key.WithHelp("f/pgdn", "page down"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup", "b"),
		key.WithHelp("b/pgup", "page up"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "right"),
	),

	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("↳", "select"),
	),

	OpenWorkflow: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "open workflow"),
		key.WithDisabled(),
	),

	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
}
