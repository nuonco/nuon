package ui

import (
	"fmt"
)

type CreateView struct {
	SpinnerView
	model string
}

func NewCreateView(model string, json bool) *CreateView {
	return &CreateView{
		*NewSpinnerView(json),
		model,
	}
}

func (v *CreateView) Start() {
	v.SpinnerView.Start(fmt.Sprintf("creating %s", v.model))
}

func (v *CreateView) Success(id string) {
	v.SpinnerView.Success(fmt.Sprintf("successfully created %s %s", v.model, id))
}

func (v *CreateView) Fail(err error) {
	v.SpinnerView.Fail(fmt.Errorf("failed to create %s: %w", v.model, err))
}
