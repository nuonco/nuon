package ui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/nuonco/nuon/bins/cli/internal/ui/teaprogram"
)

func NewProgram(model tea.Model, interactive bool, opts ...tea.ProgramOption) *tea.Program {
	if !interactive {
		opts = append(opts, tea.WithInput(nil), tea.WithoutRenderer())
	}
	return teaprogram.NewProgram(model, opts...)
}
