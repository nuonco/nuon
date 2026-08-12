package ui

import (
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
)

type SpinnerView struct {
	bubblesSpinner *bubbles.SpinnerView
}

func NewSpinnerView(json, interactive bool) *SpinnerView {
	return &SpinnerView{
		bubblesSpinner: bubbles.NewSpinnerView(json, interactive),
	}
}

func (v *SpinnerView) Start(text string) {
	v.bubblesSpinner.Start(text)
}

func (v *SpinnerView) Update(text string) {
	v.bubblesSpinner.Update(text)
}

// Fail renders the error on the spinner line and returns it marked as
// rendered (except in agent mode, where the spinner writes to stderr and the
// command boundary still owes the stdout envelope).
func (v *SpinnerView) Fail(err error) error {
	v.bubblesSpinner.Fail(err)
	return spinnerRendered(err)
}

func (v *SpinnerView) Success(text string) {
	v.bubblesSpinner.Success(text)
}
