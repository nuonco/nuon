package ui

import (
	"fmt"
)

type DeleteView struct {
	SpinnerView
	model string
	id    string
}

func NewDeleteView(model string, id string) *DeleteView {
	return &DeleteView{
		*NewSpinnerView(false),
		id,
		model,
	}
}

func (v *DeleteView) Start() {
	v.SpinnerView.Start(fmt.Sprintf("deleting %s %s", v.model, v.id))
}

func (v *DeleteView) Success() {
	v.SpinnerView.Success(fmt.Sprintf("successfully deleted %s %s", v.model, v.id))
}

func (v *DeleteView) SuccessQueued() {
	v.SpinnerView.Success(fmt.Sprintf("successfully queued %s to be deleted %s", v.id, v.model))
}

func (v *DeleteView) Fail(err error) {
	v.SpinnerView.Fail(fmt.Errorf("failed to create %s: %w", v.model, err))
}
