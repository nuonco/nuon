package ui

import (
	"fmt"
)

type DeleteView struct {
	SpinnerView
	model string
	id    string
}

func NewDeleteView(model, id string) *DeleteView {
	return &DeleteView{
		*NewSpinnerView(),
		model,
		id,
	}
}

func (v *DeleteView) Start() {
	v.SpinnerView.Start(fmt.Sprintf("deleting %s %s", v.model, v.id))
}

func (v *DeleteView) Success() {
	v.SpinnerView.Success(fmt.Sprintf("successfully deleted %s %s", v.model, v.id))
}
