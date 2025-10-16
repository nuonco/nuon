package action

import (
	"github.com/charmbracelet/bubbles/key"
)

const spacebar = " "

type keyMap struct {
	Help key.Binding
	Quit key.Binding

	// hybrid key
	Esc key.Binding

	// Enter key.Binding

	// nav controls (we override or handle directly)
	Up   key.Binding
	Down key.Binding
	// nav for steps detail viewport
	PageDown key.Binding
	PageUp   key.Binding

	// shift focus
	Left  key.Binding
	Right key.Binding
	Tab   key.Binding

	// actions
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
		// {k.Enter},
		{k.Copy, k.Browser},
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Tab, k.Quit, k.Esc, k.Browser}
}

// updateNavigationKeys updates the Up/Down key bindings based on view mode
func (k *keyMap) updateNavigationKeys(viewMode string) {
	if viewMode == "execute" {
		// Execute mode: Only arrow keys for scrolling
		k.Up = key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "scroll up"),
		)
		k.Down = key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "scroll down"),
		)
	} else {
		// Runs mode: Both arrow keys and j/k for navigation
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

	// Enter: key.NewBinding(
	// 	key.WithKeys("enter"),
	// 	key.WithHelp("↳", "select"),
	// ),
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
		key.WithDisabled(), // disabled by default. only enabled IFF the action has a manual trigger.
	),
}
