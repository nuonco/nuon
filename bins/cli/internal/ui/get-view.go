package ui

import (
	"github.com/pterm/pterm"
)

type GetView struct {
}

func NewGetView() *GetView {
	return &GetView{}
}

func (v *GetView) Render(data [][]string) {
	pterm.DefaultTable.
		WithData(data).
		Render()
}

func (v *GetView) Error(err error) {
	PrintError(err)
}
