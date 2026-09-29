package teaprogram

import (
	tea "charm.land/bubbletea/v2"
)

func NewProgram(model tea.Model, opts ...tea.ProgramOption) *tea.Program {
	opts = append([]tea.ProgramOption{tea.WithFilter(filterKeys)}, opts...)
	return tea.NewProgram(model, opts...)
}

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
