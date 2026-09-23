// Package teaprogram is the single constructor every TUI should use, so
// terminal conventions like ctrl+z backgrounding apply everywhere instead of
// being reimplemented (or forgotten) per model.
package teaprogram

import (
	tea "charm.land/bubbletea/v2"
)

// NewProgram wraps tea.NewProgram with the CLI's shared conventions. Use this
// instead of calling tea.NewProgram directly.
func NewProgram(model tea.Model, opts ...tea.ProgramOption) *tea.Program {
	opts = append([]tea.ProgramOption{tea.WithFilter(filterKeys)}, opts...)
	return tea.NewProgram(model, opts...)
}

// filterKeys normalizes keys models can't be relied on to bind themselves:
// ctrl+z always backgrounds (Bubble Tea handles Suspend natively — real
// SIGTSTP, restore on SIGCONT), and ctrl+d is remapped to ctrl+c so any model
// that already quits on ctrl+c gets ctrl+d for free.
func filterKeys(_ tea.Model, msg tea.Msg) tea.Msg {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return msg
	}
	switch key.String() {
	case "ctrl+z":
		return tea.Suspend()
	case "ctrl+d":
		return tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	}
	return msg
}
