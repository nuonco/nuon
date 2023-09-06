package ui

import "github.com/pterm/pterm"

type Spinner struct {
	pterm.SpinnerPrinter
}

func NewSpinner() *Spinner {
	return &Spinner{
		pterm.DefaultSpinner,
	}
}
