package detail

import (
	"charm.land/bubbles/v2/key"
)

const spacebar = " "

type keyMap struct {
	Help key.Binding
	Quit key.Binding

	Esc key.Binding

	Enter key.Binding

	Up       key.Binding
	Down     key.Binding
	PageDown key.Binding
	PageUp   key.Binding

	Left  key.Binding
	Right key.Binding
	Tab   key.Binding

	Copy    key.Binding
	Slash   key.Binding
	Browser key.Binding
	Execute key.Binding
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Help, k.Quit},
		{k.Up, k.Down},
		{k.Tab},
		{k.Enter},
		{k.Copy, k.Browser},
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Tab, k.Quit, k.Esc, k.Browser}
}

func (k *keyMap) updateNavigationKeys(viewMode ViewMode) {
	if viewMode == ExecuteView {
		k.Up = key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "scroll up"),
		)
		k.Down = key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "scroll down"),
		)
	} else {
		k.Up = key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		)
		k.Down = key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		)
	}
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
		key.WithHelp("↳", "select run"),
	),
	Copy: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "Copy Id"),
	),
	Slash: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "Search Runs"),
	),
	Browser: key.NewBinding(
		key.WithKeys("B"),
		key.WithHelp("B", "Browser"),
	),
	Execute: key.NewBinding(
		key.WithKeys("E"),
		key.WithHelp("E", "Execute"),
		key.WithDisabled(),
	),
}
