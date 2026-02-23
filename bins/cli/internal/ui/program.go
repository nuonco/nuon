package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// NewProgram creates a tea.Program with appropriate options for the current
// environment. When interactive is false, the renderer and input are disabled
// so that bubbletea does not attempt to access a TTY.
func NewProgram(model tea.Model, interactive bool, opts ...tea.ProgramOption) *tea.Program {
	if !interactive {
		opts = append(opts, tea.WithInput(nil), tea.WithoutRenderer())
	}
	return tea.NewProgram(model, opts...)
}
