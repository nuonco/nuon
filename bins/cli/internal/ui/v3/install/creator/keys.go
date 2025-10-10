package creator

import (
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Quit     key.Binding
	Help     key.Binding
	Browser  key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("up/k", "scroll up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("down/j", "scroll down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next field"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev field"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "esc"),
		key.WithHelp("ctrl+c/esc", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Browser: key.NewBinding(
		key.WithKeys("B"),
		key.WithHelp("B", "open in browser"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.ShiftTab, k.Enter, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Tab, k.ShiftTab, k.Up, k.Down},
		{k.Enter, k.Browser, k.Quit, k.Help},
	}
}
