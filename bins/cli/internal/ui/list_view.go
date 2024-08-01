package ui

import (
	"github.com/cockroachdb/errors/withstack"
	"github.com/powertoolsdev/mono/pkg/errs"
	"github.com/pterm/pterm"
)

type ListView struct {
}

func NewListView() *ListView {
	return &ListView{}
}

func (v *ListView) Render(data [][]string) {
	if len(data) <= 1 {
		pterm.DefaultBasicText.Println("No items found")
		return
	}

	pterm.DefaultTable.
		WithData(data).
		WithHasHeader().
		WithHeaderRowSeparator("-").
		Render()
}

func (v *ListView) Error(err error) error {
	if !errs.HasNuonStackTrace(err) {
		err = withstack.WithStackDepth(err, 1)
	}
	return PrintError(err)
}

func (v *ListView) Print(msg string) {
	pterm.DefaultBasicText.Println(msg)
}
