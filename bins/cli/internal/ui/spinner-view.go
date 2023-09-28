package ui

import (
	"github.com/pterm/pterm"
)

type SpinnerView struct {
	json    bool
	spinner *pterm.SpinnerPrinter
}

func NewSpinnerView(json bool) *SpinnerView {
	return &SpinnerView{
		json,
		nil,
	}
}

func (v *SpinnerView) Start(text string) {
	if v.json {
		return
	}

	spinner, _ := pterm.DefaultSpinner.Start(text)
	v.spinner = spinner
}

func (v *SpinnerView) Update(text string) {
	if v.json {
		return
	}

	// force clearing the line
	// TODO: this is a work-around for a pterm bug that we should be able to remove in the future
	// we think it's related to this: https://github.com/pterm/pterm/pull/447
	v.spinner.UpdateText("                                                                                                   ")

	v.spinner.UpdateText(text)
}

func (v *SpinnerView) Fail(err error) {
	if v.json {
		printJSONError(err)
		return
	}

	v.spinner.Fail(err.Error())
}

func (v *SpinnerView) Success(text string) {
	if v.json {
		printJSON(text)
		return
	}

	v.spinner.Success(text)
}
