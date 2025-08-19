package ui

import (
	"errors"
	"fmt"
	"os"

	"github.com/cockroachdb/errors/withstack"
	"github.com/nuonco/nuon-go"
	"github.com/pterm/pterm"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/config/parse"
	"github.com/powertoolsdev/mono/pkg/config/sync"
	"github.com/powertoolsdev/mono/pkg/errs"
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

func PrintError(err error) error {
	if os.Getenv(debugEnvVar) != "" {
		pterm.Error.Println(fmt.Sprintf("DEBUG: %v", err))
	}

	// Construct a stack trace if this error doesn't already have one
	if !errs.HasNuonStackTrace(err) {
		err = withstack.WithStackDepth(err, 1)
	}

	cliUserErr := &CLIUserError{}
	if errors.As(err, &cliUserErr) {
		pterm.Error.Println(err.Error())
		return err
	}

	apiUserErr, ok := nuon.ToUserError(err)
	if ok {
		pterm.Error.Println(apiUserErr.Description)
		return err
	}

	if nuon.IsServerError(err) {
		pterm.Error.Println(defaultServerErrorMessage)
		return err
	}

	var cfgErr config.ErrConfig
	if errors.As(err, &cfgErr) {
		msg := fmt.Sprintf("%s %s", cfgErr.Description, cfgErr.Error())
		if cfgErr.Warning {
			pterm.Warning.Println(msg)
			return cfgErr
		}

		pterm.Error.Println(msg)
		return cfgErr
	}

	var syncErr sync.SyncErr
	if errors.As(err, &syncErr) {
		pterm.Error.Println(syncErr.Error())
		return syncErr
	}

	var syncAPIErr sync.SyncAPIErr
	if errors.As(err, &syncAPIErr) {
		pterm.Error.Println(syncAPIErr.Error())
		return syncAPIErr
	}

	var parseErr parse.ParseErr
	if errors.As(err, &parseErr) {
		pterm.Error.Println(parseErr.Description)
		if parseErr.Err != nil {
			pterm.Error.Println(parseErr.Err.Error())
		}

		return parseErr
	}

	pterm.Error.Println(err.Error())
	return err
}

func PrintLn(msg string) {
	pterm.Info.Println(msg)
}

func PrintWarning(msg string) {
	pterm.Warning.Println(msg)
}

func PrintSuccess(msg string) {
	pterm.Success.Println(msg)
}
