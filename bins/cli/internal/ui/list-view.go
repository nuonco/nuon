package ui

import (
	"github.com/pterm/pterm"
)

type ListView struct {
	headerRow []string
}

func NewListView(headerRow []string) *ListView {
	return &ListView{
		headerRow: headerRow,
	}
}

func (v *ListView) Render(rows [][]string) {
	if len(rows) == 0 {
		pterm.DefaultBasicText.Println("No items found")
		return
	}

	data := append([][]string{v.headerRow}, rows...)
	pterm.DefaultTable.
		WithData(data).
		WithHasHeader().
		WithHeaderRowSeparator("-").
		Render()
}

func (v *ListView) Error(err error) {
	pterm.Error.Println(err)
}
