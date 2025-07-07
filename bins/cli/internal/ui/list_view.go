package ui

import (
	"github.com/cockroachdb/errors/withstack"
	"github.com/powertoolsdev/mono/pkg/errs"
	"github.com/pterm/pterm"
)

type ListView struct{}

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

func (v *ListView) RenderPaging(data [][]string, offset, limit int, hasMore bool) {
	if len(data) <= 1 {
		pterm.DefaultBasicText.Println("No items found")
		return
	}

	pterm.DefaultTable.
		WithData(data).
		WithHasHeader().
		WithHeaderRowSeparator("-").
		Render()

	pterm.DefaultBasicText.Printf("offset %d, limit %d, ", offset, limit)
	if hasMore {
		pterm.DefaultBasicText.Println("more items available")
	} else {
		pterm.DefaultBasicText.Println("no more items available")
	}
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
