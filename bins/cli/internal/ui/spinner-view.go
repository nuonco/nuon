package ui

import "github.com/pterm/pterm"

type SpinnerView struct {
}

func NewSpinnerView() *SpinnerView {
	pterm.DefaultSpinner.Start()
	return &SpinnerView{}
}

func (v *SpinnerView) Start(text string) {
	pterm.DefaultSpinner.Start(text)
}

func (v *SpinnerView) Update(text string) {
	pterm.DefaultSpinner.UpdateText(text)
}

func (v *SpinnerView) Fail(err error) {
	pterm.DefaultSpinner.Fail(err.Error() + "\n")
}

func (v *SpinnerView) Success(text string) {
	pterm.DefaultSpinner.Success(text + "\n")
}
