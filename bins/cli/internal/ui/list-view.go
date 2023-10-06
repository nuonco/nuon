package ui

import (
	"github.com/nuonco/nuon-go"
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

func (v *ListView) Error(err error) {
	userErr, ok := nuon.ToUserError(err)
	if ok {
		pterm.Error.Println(userErr.Description)
		return
	}

	if nuon.IsServerError(err) {
		pterm.Error.Println(defaultServerErrorMessage)
		return
	}

	pterm.Error.Println(err)
}
