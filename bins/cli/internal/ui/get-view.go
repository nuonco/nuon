package ui

import (
	"github.com/nuonco/nuon-go"
	"github.com/pterm/pterm"
)

const (
	defaultServerErrorMessage  string = "Oops, we have experienced a server error. Please try again in a few minutes."
	defaultUnknownErrorMessage string = "Oops, we have experienced an unexpected error. Please let us know about this."
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
