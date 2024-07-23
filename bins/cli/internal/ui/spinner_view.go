package ui

import (
	"github.com/cockroachdb/errors"
	"github.com/mattn/go-runewidth"
	"github.com/nuonco/nuon-go"
	"github.com/pterm/pterm"
)

const (
	// in order to prevent tearing in the CLI, we ensure that the length of the spinner text is _always_ consistent
	// this prevents the spinner from partially updating, where the text being a different size distorts it.
	defaultSpinnerWidth int = 30
)

type SpinnerView struct {
	json     bool
	spinner  *pterm.SpinnerPrinter
	prevText string
}

func NewSpinnerView(json bool) *SpinnerView {
	return &SpinnerView{
		json,
		nil,
		"",
	}
}

func (v *SpinnerView) formatText(text string) string {
	updatedText := runewidth.FillRight(text, len(v.prevText))
	v.prevText = text
	return updatedText
}

func (v *SpinnerView) Start(text string) {
	if v.json {
		return
	}

	spinner, err := pterm.DefaultSpinner.Start(v.formatText(text))
	if err != nil {
		printDebugErr(err)
		return
	}
	v.spinner = spinner
}

func (v *SpinnerView) Update(text string) {
	if v.json {
		return
	}

	v.spinner.UpdateText(v.formatText(text))
}

func (v *SpinnerView) Fail(err error) {
	if v.json {
		PrintJSONError(err)
		return
	}

	if hints := errors.FlattenHints(err); hints != "" {
		v.spinner.Fail(v.formatText(hints))
		return
	}

	userErr, ok := nuon.ToUserError(err)
	if ok {
		v.spinner.Fail(v.formatText(userErr.Description))
		return
	}

	if nuon.IsServerError(err) {
		v.spinner.Fail(v.formatText(defaultServerErrorMessage))
		return
	}

	v.spinner.Fail(v.formatText(err.Error()))
}

func (v *SpinnerView) Success(text string) {
	if v.json {
		PrintJSON(text)
		return
	}

	v.spinner.Success(v.formatText(text))
}
