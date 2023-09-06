package ui

import "github.com/pterm/pterm"

type UpdateView struct {
}

func NewUpdateView() *UpdateView {
	pterm.DefaultSpinner.Start()
	return &UpdateView{}
}

func (v *UpdateView) Update(text string) {
	pterm.DefaultSpinner.UpdateText(pterm.DefaultBasicText.Sprintln(text))
}

func (v *UpdateView) Fail(text string) {
	pterm.DefaultSpinner.Fail(pterm.DefaultBasicText.Sprintln(text))
}

func (v *UpdateView) Success(text string) {
	pterm.DefaultSpinner.Success(pterm.DefaultBasicText.Sprintln(text))
}
