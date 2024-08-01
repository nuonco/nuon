package ui

import (
	"github.com/cockroachdb/errors/withstack"
	"github.com/powertoolsdev/mono/pkg/errs"
	"github.com/pterm/pterm"
)

type GetView struct {
}

func NewGetView() *GetView {
	return &GetView{}
}

func (v *GetView) Print(msg string) {
	pterm.DefaultBasicText.Println(msg)
}

func (v *GetView) Render(data [][]string) {
	pterm.DefaultTable.
		WithData(data).
		Render()
}

func (v *GetView) Error(err error) error {
	if !errs.HasNuonStackTrace(err) {
		err = withstack.WithStackDepth(err, 1)
	}
	PrintError(err)
	return err
}
