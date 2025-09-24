package workflow

import (
	"github.com/charmbracelet/bubbles/key"
)

const spacebar = " "

type keyMap struct {
	Help key.Binding
	Quit key.Binding

	// hybrid key
	Esc key.Binding

	Enter key.Binding

	// nav controls (we override or handle directly)
	Up   key.Binding
	Down key.Binding
	// nav for step detail viewport
	PageDown key.Binding
	PageUp   key.Binding

	// shift focus
	Left  key.Binding
	Right key.Binding
	Tab   key.Binding

	// display controls
	ToggleJson       key.Binding
	OpenQuickLink    key.Binding
	OpenTemplateLink key.Binding

	// actions
	Copy           key.Binding
	Slash          key.Binding
	Browser        key.Binding
	ApproveStep    key.Binding
	CancelWorkflow key.Binding
	ApproveAll     key.Binding
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Help, k.Quit},
		{k.Up, k.Down},
		{k.Enter, k.ApproveStep},
		{k.Copy, k.Browser},
		{k.ToggleJson},
		{k.OpenQuickLink, k.OpenTemplateLink},
		{k.ApproveAll, k.CancelWorkflow},
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit, k.Enter, k.Esc, k.ToggleJson, k.ApproveStep, k.Browser, k.OpenQuickLink, k.OpenTemplateLink}
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
		key.WithHelp("esc", "back"), // TODO: dynamic cancel, back, or quit
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
		key.WithHelp("←/h", "next step"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "next step"),
	),

	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("↳", "select"),
	),
	Copy: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "Copy Id"),
	),
	Slash: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "Search Steps"),
	),
	// step detail actions
	Browser: key.NewBinding(
		key.WithKeys("B"),
		key.WithHelp("B", "Browser"),
	),
	OpenQuickLink: key.NewBinding(
		key.WithKeys("L"),
		key.WithHelp("L", "quick link"),
		key.WithDisabled(), // disabled by default
	),
	OpenTemplateLink: key.NewBinding(
		key.WithKeys("T"),
		key.WithHelp("T", "template"),
		key.WithDisabled(), // disabled by default
	),
	// step options
	ApproveStep: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "approve step "),
		key.WithDisabled(), // disabled by default
	),
	// workflow actions
	CancelWorkflow: key.NewBinding(
		key.WithKeys("C"),
		key.WithHelp("C", "cancel workflow"),
	),

	ApproveAll: key.NewBinding(
		key.WithKeys("A"),
		key.WithHelp("A", "approve all"),
	),
	// step detail display options
	ToggleJson: key.NewBinding(
		key.WithKeys("J"),
		key.WithHelp("J", "toggle json"),
		key.WithDisabled(), // disabled by default
	),
}
