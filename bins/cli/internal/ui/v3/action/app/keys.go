package app

import (
	"github.com/charmbracelet/bubbles/key"
)

const spacebar = " "

type keyMap struct {
	Help key.Binding
	// high level
	Quit key.Binding
	Esc  key.Binding
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Help, k.Quit, k.Esc},
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Help, k.Quit, k.Esc,
	}
}

// updateNavigationKeys updates the Up/Down key bindings based on view mode
var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Esc: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}
