package ui

import (
	"errors"
	"os"

	"github.com/nuonco/nuon-go"
	"github.com/pterm/pterm"

	"github.com/powertoolsdev/mono/pkg/config"
)

const (
	defaultServerErrorMessage  string = "Oops, we have experienced a server error. Please try again in a few minutes."
	defaultUnknownErrorMessage string = "Oops, we have experienced an unexpected error. Please let us know about this."
	debugEnvVar                string = "NUON_DEBUG"
)

type CLIUserError struct {
	Msg string
}

func (u *CLIUserError) Error() string {
	return u.Msg
}

func PrintError(err error) {
	if os.Getenv(debugEnvVar) != "" {
		pterm.Error.Println(err.Error())
	}

	cliUserErr := &CLIUserError{}
	if errors.As(err, &cliUserErr) {
		pterm.Error.Println(err.Error())
		os.Exit(1)
		return
	}

	apiUserErr, ok := nuon.ToUserError(err)
	if ok {
		pterm.Error.Println(apiUserErr.Description)
		os.Exit(1)
		return
	}

	if nuon.IsServerError(err) {
		pterm.Error.Println(defaultServerErrorMessage)
		os.Exit(1)
		return
	}

	var cfgErr config.ErrConfig
	if errors.As(err, &cfgErr) {
		pterm.Error.Println(cfgErr.Description)
		return
	}

	pterm.Error.Println(defaultUnknownErrorMessage)
	os.Exit(1)
}

func PrintLn(msg string) {
	pterm.Info.Println(msg)
}
