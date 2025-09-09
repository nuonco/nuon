package ui

import (
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/bubbles"
)

type SpinnerView struct {
	bubblesSpinner *bubbles.SpinnerView
}

func NewSpinnerView(json bool) *SpinnerView {
	return &SpinnerView{
		bubblesSpinner: bubbles.NewSpinnerView(json),
	}
}

func (v *SpinnerView) Start(text string) {
	v.bubblesSpinner.Start(text)
}

func (v *SpinnerView) Update(text string) {
	v.bubblesSpinner.Update(text)
}

func (v *SpinnerView) Fail(err error) {
	v.bubblesSpinner.Fail(err)
}

func (v *SpinnerView) Success(text string) {
	v.bubblesSpinner.Success(text)
}
