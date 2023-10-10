package ui

import (
	"github.com/nuonco/nuon-go"
	"github.com/pterm/pterm"
)

const (
	defaultServerErrorMessage  string = "Oops, we have experienced a server error. Please try again in a few minutes."
	defaultUnknownErrorMessage string = "Oops, we have experienced an unexpected error. Please let us know about this."
)

func PrintError(err error) {
	userErr, ok := nuon.ToUserError(err)
	if ok {
		pterm.Error.Println(userErr.Description)
		return
	}

	if nuon.IsServerError(err) {
		pterm.Error.Println(defaultServerErrorMessage)
		return
	}

	pterm.Error.Println(defaultUnknownErrorMessage)
}
