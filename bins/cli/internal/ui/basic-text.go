package ui

import "github.com/pterm/pterm"

type BasicText struct {
	pterm.BasicTextPrinter
}

func NewBasicText() *BasicText {
	return &BasicText{
		pterm.DefaultBasicText,
	}
}
